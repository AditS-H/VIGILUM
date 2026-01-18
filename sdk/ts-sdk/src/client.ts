/**
 * VIGILUM API Client
 */

import type { Address } from 'viem';
import { API_ENDPOINTS, DEFAULT_CONFIG } from './constants';
import type {
  Alert,
  ClientConfig,
  ContractInfo,
  ContractQuery,
  ReputationInfo,
  ScanOptions,
  ScanResult,
} from './types';
import { ContractInfoSchema, ScanResultSchema } from './types';

export class VigilumClient {
  private readonly config: Required<ClientConfig>;

  constructor(config: ClientConfig) {
    this.config = {
      baseUrl: config.baseUrl || DEFAULT_CONFIG.baseUrl,
      apiKey: config.apiKey,
      timeout: config.timeout ?? DEFAULT_CONFIG.timeout,
      retries: config.retries ?? DEFAULT_CONFIG.retries,
    };
  }

  // ═══════════════════════════════════════════════════════════════════════════
  // SCANNING
  // ═══════════════════════════════════════════════════════════════════════════

  /**
   * Submit a contract for security scanning.
   */
  async scan(query: ContractQuery, options?: ScanOptions): Promise<ScanResult> {
    const response = await this.request<ScanResult>('POST', API_ENDPOINTS.SCAN, {
      body: {
        address: query.address,
        chainId: query.chainId,
        ...options,
      },
    });
    return ScanResultSchema.parse(response);
  }

  /**
   * Get scan result by ID.
   */
  async getScanResult(scanId: string): Promise<ScanResult> {
    const response = await this.request<ScanResult>(
      'GET',
      `${API_ENDPOINTS.SCAN}/${scanId}`
    );
    return ScanResultSchema.parse(response);
  }

  /**
   * Poll for scan completion.
   */
  async waitForScan(
    scanId: string,
    options?: { pollInterval?: number; timeout?: number }
  ): Promise<ScanResult> {
    const pollInterval = options?.pollInterval ?? 2000;
    const timeout = options?.timeout ?? 300000;
    const startTime = Date.now();

    while (Date.now() - startTime < timeout) {
      const result = await this.getScanResult(scanId);
      if (result.status === 'completed' || result.status === 'failed') {
        return result;
      }
      await this.sleep(pollInterval);
    }

    throw new Error(`Scan timeout after ${timeout}ms`);
  }

  // ═══════════════════════════════════════════════════════════════════════════
  // CONTRACT INFO
  // ═══════════════════════════════════════════════════════════════════════════

  /**
   * Get contract security information.
   */
  async getContractInfo(query: ContractQuery): Promise<ContractInfo> {
    const response = await this.request<ContractInfo>(
      'GET',
      `${API_ENDPOINTS.CONTRACT}/${query.chainId}/${query.address}`
    );
    return ContractInfoSchema.parse(response);
  }

  /**
   * Check if a contract is blacklisted.
   */
  async isBlacklisted(query: ContractQuery): Promise<boolean> {
    const info = await this.getContractInfo(query);
    return info.isBlacklisted;
  }

  /**
   * Get risk score for a contract.
   */
  async getRiskScore(query: ContractQuery): Promise<number> {
    const info = await this.getContractInfo(query);
    return info.riskScore;
  }

  // ═══════════════════════════════════════════════════════════════════════════
  // ALERTS
  // ═══════════════════════════════════════════════════════════════════════════

  /**
   * Get alerts for an address.
   */
  async getAlerts(address: Address, chainId?: number): Promise<Alert[]> {
    const params = new URLSearchParams({ address });
    if (chainId) params.set('chainId', chainId.toString());
    
    return this.request<Alert[]>('GET', `${API_ENDPOINTS.ALERTS}?${params}`);
  }

  /**
   * Subscribe to alerts via webhook.
   */
  async subscribeAlerts(webhookUrl: string, filters?: { 
    addresses?: Address[];
    minSeverity?: string;
    chainIds?: number[];
  }): Promise<{ subscriptionId: string }> {
    return this.request('POST', `${API_ENDPOINTS.ALERTS}/subscribe`, {
      body: { webhookUrl, filters },
    });
  }

  // ═══════════════════════════════════════════════════════════════════════════
  // REPUTATION
  // ═══════════════════════════════════════════════════════════════════════════

  /**
   * Get reputation info for an address.
   */
  async getReputation(address: Address): Promise<ReputationInfo> {
    return this.request<ReputationInfo>(
      'GET',
      `${API_ENDPOINTS.REPUTATION}/${address}`
    );
  }

  // ═══════════════════════════════════════════════════════════════════════════
  // PRIVATE METHODS
  // ═══════════════════════════════════════════════════════════════════════════

  private async request<T>(
    method: 'GET' | 'POST' | 'PUT' | 'DELETE',
    path: string,
    options?: { body?: unknown }
  ): Promise<T> {
    const url = `${this.config.baseUrl}${path}`;
    
    let lastError: Error | undefined;
    
    for (let attempt = 0; attempt <= this.config.retries; attempt++) {
      try {
        const controller = new AbortController();
        const timeoutId = setTimeout(
          () => controller.abort(),
          this.config.timeout
        );

        const response = await fetch(url, {
          method,
          headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${this.config.apiKey}`,
            'X-Client-Version': '0.1.0',
          },
          body: options?.body ? JSON.stringify(options.body) : undefined,
          signal: controller.signal,
        });

        clearTimeout(timeoutId);

        if (!response.ok) {
          const error = await response.json().catch(() => ({})) as { message?: string; code?: string };
          throw new VigilumApiError(
            response.status,
            error.message || `HTTP ${response.status}`,
            error.code
          );
        }

        return response.json() as Promise<T>;
      } catch (error) {
        lastError = error as Error;
        if (attempt < this.config.retries) {
          await this.sleep(Math.pow(2, attempt) * 1000);
        }
      }
    }

    throw lastError;
  }

  private sleep(ms: number): Promise<void> {
    return new Promise((resolve) => setTimeout(resolve, ms));
  }
}

export class VigilumApiError extends Error {
  constructor(
    public readonly status: number,
    message: string,
    public readonly code?: string
  ) {
    super(message);
    this.name = 'VigilumApiError';
  }
}
