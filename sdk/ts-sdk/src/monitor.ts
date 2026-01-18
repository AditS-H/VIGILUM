/**
 * Real-time monitoring utilities for mempool and on-chain events.
 */

import type { Address, Hash, Hex, Log, PublicClient } from 'viem';
import { ThreatLevel } from './types';
import { VigilumClient } from './client';

export interface MonitoringConfig {
  /** VIGILUM client for API access */
  client?: VigilumClient;
  /** Public client for on-chain monitoring */
  publicClient: PublicClient;
  /** Addresses to monitor */
  addresses?: Address[];
  /** Chain ID */
  chainId: number;
  /** Enable mempool monitoring */
  mempoolEnabled?: boolean;
  /** Callback for alerts */
  onAlert?: (alert: MonitorAlert) => void;
  /** Callback for pending transactions */
  onPendingTx?: (tx: PendingTransaction) => void;
}

export interface MonitorAlert {
  id: string;
  type: 'suspicious_contract' | 'large_transfer' | 'unusual_activity' | 'contract_interaction' | 'approval';
  severity: ThreatLevel;
  title: string;
  description: string;
  address?: Address;
  txHash?: Hash;
  timestamp: number;
  metadata?: Record<string, unknown>;
}

export interface PendingTransaction {
  hash: Hash;
  from: Address;
  to: Address | null;
  value: bigint;
  data: Hex;
  gas: bigint;
  gasPrice?: bigint;
  maxFeePerGas?: bigint;
  maxPriorityFeePerGas?: bigint;
}

export interface MonitorStats {
  alertCount: number;
  pendingTxCount: number;
  lastAlertAt?: number;
  lastTxAt?: number;
  uptime: number;
}

// Known high-value function selectors to watch
const WATCHED_SELECTORS = {
  '0x095ea7b3': 'approve',
  '0xa22cb465': 'setApprovalForAll',
  '0x23b872dd': 'transferFrom',
  '0x42842e0e': 'safeTransferFrom',
  '0xb88d4fde': 'safeTransferFrom',
  '0xa9059cbb': 'transfer',
  '0x2e1a7d4d': 'withdraw',
  '0x18cbafe5': 'swapExactTokensForETH',
  '0x7ff36ab5': 'swapExactETHForTokens',
} as const;

export class ContractMonitor {
  private readonly config: MonitoringConfig;
  private readonly startTime: number;
  private alertCount = 0;
  private pendingTxCount = 0;
  private lastAlertAt?: number;
  private lastTxAt?: number;
  private unsubscribers: (() => void)[] = [];

  constructor(config: MonitoringConfig) {
    this.config = config;
    this.startTime = Date.now();
  }

  /**
   * Start monitoring.
   */
  async start(): Promise<void> {
    // Watch for contract events
    if (this.config.addresses?.length) {
      for (const address of this.config.addresses) {
        this.watchAddress(address);
      }
    }

    // Watch for pending transactions if enabled
    if (this.config.mempoolEnabled) {
      this.watchMempool();
    }
  }

  /**
   * Stop all monitoring.
   */
  stop(): void {
    for (const unsub of this.unsubscribers) {
      unsub();
    }
    this.unsubscribers = [];
  }

  /**
   * Add an address to monitor.
   */
  addAddress(address: Address): void {
    if (!this.config.addresses) {
      this.config.addresses = [];
    }
    if (!this.config.addresses.includes(address)) {
      this.config.addresses.push(address);
      this.watchAddress(address);
    }
  }

  /**
   * Remove an address from monitoring.
   */
  removeAddress(address: Address): void {
    if (this.config.addresses) {
      this.config.addresses = this.config.addresses.filter(a => a !== address);
    }
  }

  /**
   * Get current monitoring stats.
   */
  getStats(): MonitorStats {
    return {
      alertCount: this.alertCount,
      pendingTxCount: this.pendingTxCount,
      lastAlertAt: this.lastAlertAt,
      lastTxAt: this.lastTxAt,
      uptime: Date.now() - this.startTime,
    };
  }

  // ═══════════════════════════════════════════════════════════════════════════
  // PRIVATE METHODS
  // ═══════════════════════════════════════════════════════════════════════════

  private watchAddress(address: Address): void {
    // Watch all events from this address
    const unwatch = this.config.publicClient.watchEvent({
      address,
      onLogs: (logs) => this.handleLogs(logs, address),
    });

    this.unsubscribers.push(unwatch);
  }

  private watchMempool(): void {
    // Note: Standard JSON-RPC doesn't support pending tx subscription
    // This would require a specialized node or service
    // For now, we poll the mempool if available
    
    const pollInterval = setInterval(async () => {
      try {
        // Get pending block if supported
        const pendingBlock = await this.config.publicClient.getBlock({
          blockTag: 'pending',
        });

        if (pendingBlock && pendingBlock.transactions) {
          for (const txHash of pendingBlock.transactions) {
            if (typeof txHash === 'string') {
              await this.checkPendingTx(txHash as Hash);
            }
          }
        }
      } catch {
        // Pending block not supported by this provider
      }
    }, 5000);

    this.unsubscribers.push(() => clearInterval(pollInterval));
  }

  private async handleLogs(logs: Log[], address: Address): Promise<void> {
    for (const log of logs) {
      // Check for Transfer events (ERC20/ERC721)
      if (log.topics[0] === '0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef') {
        await this.handleTransfer(log, address);
      }

      // Check for Approval events
      if (log.topics[0] === '0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925') {
        await this.handleApproval(log, address);
      }
    }
  }

  private async handleTransfer(log: Log, watchedAddress: Address): Promise<void> {
    // Large value transfers generate alerts
    const value = log.data !== '0x' ? BigInt(log.data) : 0n;
    
    // Threshold: 10 ETH equivalent (adjust based on token)
    if (value > 10n * 10n ** 18n) {
      this.emitAlert({
        id: crypto.randomUUID(),
        type: 'large_transfer',
        severity: ThreatLevel.MEDIUM,
        title: 'Large Token Transfer',
        description: `Large transfer detected on monitored contract`,
        address: watchedAddress,
        txHash: log.transactionHash ?? undefined,
        timestamp: Date.now(),
        metadata: { value: value.toString() },
      });
    }
  }

  private async handleApproval(log: Log, watchedAddress: Address): Promise<void> {
    // Unlimited approvals are noteworthy
    const value = log.data !== '0x' ? BigInt(log.data) : 0n;
    const MAX_UINT256 = 2n ** 256n - 1n;

    if (value === MAX_UINT256) {
      this.emitAlert({
        id: crypto.randomUUID(),
        type: 'approval',
        severity: ThreatLevel.INFO,
        title: 'Unlimited Token Approval',
        description: 'Unlimited approval granted to spender',
        address: watchedAddress,
        txHash: log.transactionHash ?? undefined,
        timestamp: Date.now(),
      });
    }
  }

  private async checkPendingTx(hash: Hash): Promise<void> {
    try {
      const tx = await this.config.publicClient.getTransaction({ hash });
      
      if (!tx) return;

      this.pendingTxCount++;
      this.lastTxAt = Date.now();

      // Check if TX interacts with monitored addresses
      if (tx.to && this.config.addresses?.includes(tx.to)) {
        const pending: PendingTransaction = {
          hash: tx.hash,
          from: tx.from,
          to: tx.to,
          value: tx.value,
          data: tx.input,
          gas: tx.gas,
          gasPrice: tx.gasPrice,
          maxFeePerGas: tx.maxFeePerGas,
          maxPriorityFeePerGas: tx.maxPriorityFeePerGas,
        };

        this.config.onPendingTx?.(pending);

        // Check for suspicious patterns
        if (tx.input && tx.input.length > 2) {
          const selector = tx.input.slice(0, 10);
          if (selector in WATCHED_SELECTORS) {
            this.emitAlert({
              id: crypto.randomUUID(),
              type: 'contract_interaction',
              severity: ThreatLevel.INFO,
              title: `Pending: ${WATCHED_SELECTORS[selector as keyof typeof WATCHED_SELECTORS]}`,
              description: `Pending transaction to monitored contract`,
              address: tx.to,
              txHash: hash,
              timestamp: Date.now(),
              metadata: { function: WATCHED_SELECTORS[selector as keyof typeof WATCHED_SELECTORS] },
            });
          }
        }
      }
    } catch {
      // Transaction may have been mined already
    }
  }

  private emitAlert(alert: MonitorAlert): void {
    this.alertCount++;
    this.lastAlertAt = Date.now();
    this.config.onAlert?.(alert);
  }
}

/**
 * Simple address watchlist manager.
 */
export class Watchlist {
  private addresses: Map<Address, WatchlistEntry> = new Map();

  add(address: Address, label?: string, notes?: string): void {
    this.addresses.set(address, {
      address,
      label,
      notes,
      addedAt: new Date().toISOString(),
    });
  }

  remove(address: Address): boolean {
    return this.addresses.delete(address);
  }

  has(address: Address): boolean {
    return this.addresses.has(address);
  }

  get(address: Address): WatchlistEntry | undefined {
    return this.addresses.get(address);
  }

  list(): WatchlistEntry[] {
    return Array.from(this.addresses.values());
  }

  toAddresses(): Address[] {
    return Array.from(this.addresses.keys());
  }

  clear(): void {
    this.addresses.clear();
  }

  toJSON(): WatchlistEntry[] {
    return this.list();
  }

  static fromJSON(entries: WatchlistEntry[]): Watchlist {
    const watchlist = new Watchlist();
    for (const entry of entries) {
      watchlist.addresses.set(entry.address, entry);
    }
    return watchlist;
  }
}

export interface WatchlistEntry {
  address: Address;
  label?: string;
  notes?: string;
  addedAt: string;
}
