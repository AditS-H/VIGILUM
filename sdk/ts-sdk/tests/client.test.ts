/**
 * Tests for VIGILUM SDK.
 */

import { describe, it, expect, beforeEach, vi } from 'vitest';
import { VigilumClient, VigilumApiError } from '../src/client';
import { ContractAnalyzer } from '../src/analyzer';
import { ThreatLevel, VulnerabilityType, ScanStatus } from '../src/types';
import { SUPPORTED_CHAINS, RISK_THRESHOLDS, DEFAULT_CONFIG } from '../src/constants';

describe('VigilumClient', () => {
  let client: VigilumClient;

  beforeEach(() => {
    client = new VigilumClient({
      apiKey: 'test-api-key',
      baseUrl: 'https://api.test.vigilum.io',
    });
  });

  it('should create client with config', () => {
    expect(client).toBeInstanceOf(VigilumClient);
  });

  it('should use default config values', () => {
    const defaultClient = new VigilumClient({
      apiKey: 'test-key',
    });
    expect(defaultClient).toBeInstanceOf(VigilumClient);
  });

  describe('scan', () => {
    it('should submit scan request', async () => {
      // Mock fetch
      globalThis.fetch = vi.fn().mockResolvedValue({
        ok: true,
        json: () => Promise.resolve({
          id: 'scan-123',
          contractAddress: '0x1234567890123456789012345678901234567890',
          chainId: 1,
          status: 'pending',
          riskScore: 0,
          threatLevel: 'none',
          vulnerabilities: [],
          metrics: {
            totalIssues: 0,
            criticalCount: 0,
            highCount: 0,
            mediumCount: 0,
            lowCount: 0,
            infoCount: 0,
          },
          startedAt: new Date().toISOString(),
        }),
      });

      const result = await client.scan({
        address: '0x1234567890123456789012345678901234567890',
        chainId: 1,
      });

      expect(result.id).toBe('scan-123');
      expect(result.status).toBe('pending');
    });

    it('should handle API errors', async () => {
      // Create client with no retries for fast test
      const noRetryClient = new VigilumClient({
        apiKey: 'test-api-key',
        baseUrl: 'https://api.test.vigilum.io',
        retries: 0,
      });

      globalThis.fetch = vi.fn().mockResolvedValue({
        ok: false,
        status: 401,
        json: () => Promise.resolve({ message: 'Unauthorized', code: 'AUTH_ERROR' }),
      });

      await expect(noRetryClient.scan({
        address: '0x1234567890123456789012345678901234567890',
        chainId: 1,
      })).rejects.toThrow(VigilumApiError);
    });
  });

  describe('getRiskScore', () => {
    it('should return risk score', async () => {
      globalThis.fetch = vi.fn().mockResolvedValue({
        ok: true,
        json: () => Promise.resolve({
          address: '0x1234567890123456789012345678901234567890',
          chainId: 1,
          bytecodeHash: '0x123',
          isVerified: true,
          isBlacklisted: false,
          riskScore: 45,
          threatLevel: 'medium',
          labels: [],
        }),
      });

      const score = await client.getRiskScore({
        address: '0x1234567890123456789012345678901234567890',
        chainId: 1,
      });

      expect(score).toBe(45);
    });
  });
});

describe('Types', () => {
  it('should have all threat levels', () => {
    expect(ThreatLevel.CRITICAL).toBe('critical');
    expect(ThreatLevel.HIGH).toBe('high');
    expect(ThreatLevel.MEDIUM).toBe('medium');
    expect(ThreatLevel.LOW).toBe('low');
    expect(ThreatLevel.INFO).toBe('info');
    expect(ThreatLevel.NONE).toBe('none');
  });

  it('should have all vulnerability types', () => {
    expect(VulnerabilityType.REENTRANCY).toBe('reentrancy');
    expect(VulnerabilityType.INTEGER_OVERFLOW).toBe('integer_overflow');
    expect(VulnerabilityType.ACCESS_CONTROL).toBe('access_control');
  });

  it('should have all scan statuses', () => {
    expect(ScanStatus.PENDING).toBe('pending');
    expect(ScanStatus.RUNNING).toBe('running');
    expect(ScanStatus.COMPLETED).toBe('completed');
    expect(ScanStatus.FAILED).toBe('failed');
  });
});

describe('Constants', () => {
  it('should have supported chains', () => {
    expect(SUPPORTED_CHAINS.ETHEREUM).toBe(1);
    expect(SUPPORTED_CHAINS.POLYGON).toBe(137);
    expect(SUPPORTED_CHAINS.BSC).toBe(56);
    expect(SUPPORTED_CHAINS.ARBITRUM).toBe(42161);
    expect(SUPPORTED_CHAINS.BASE).toBe(8453);
  });

  it('should have risk thresholds', () => {
    expect(RISK_THRESHOLDS.CRITICAL).toBe(80);
    expect(RISK_THRESHOLDS.HIGH).toBe(60);
    expect(RISK_THRESHOLDS.MEDIUM).toBe(40);
    expect(RISK_THRESHOLDS.LOW).toBe(20);
  });

  it('should have default config', () => {
    expect(DEFAULT_CONFIG.timeout).toBe(30000);
    expect(DEFAULT_CONFIG.retries).toBe(3);
  });
});

describe('VigilumApiError', () => {
  it('should create error with status and code', () => {
    const error = new VigilumApiError(401, 'Unauthorized', 'AUTH_ERROR');

    expect(error.status).toBe(401);
    expect(error.message).toBe('Unauthorized');
    expect(error.code).toBe('AUTH_ERROR');
    expect(error.name).toBe('VigilumApiError');
  });
});
