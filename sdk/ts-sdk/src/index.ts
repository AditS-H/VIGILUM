/**
 * VIGILUM SDK
 * ===========
 * TypeScript SDK for smart contract security analysis.
 * 
 * @example
 * // Basic usage
 * import { VigilumClient, ContractAnalyzer } from '@vigilum/sdk';
 * 
 * const client = new VigilumClient({ apiKey: 'your-api-key' });
 * const result = await client.scan({ address: '0x...', chainId: 1 });
 * 
 * @example
 * // Transaction guard
 * import { TransactionGuard } from '@vigilum/sdk';
 * 
 * const guard = new TransactionGuard(publicClient);
 * const { safe, warnings } = await guard.check(transaction);
 * 
 * @example
 * // React hooks
 * import { createVigilumHooks } from '@vigilum/sdk';
 * import * as React from 'react';
 * 
 * const { useVigilum, useScan } = createVigilumHooks(React);
 */

// Core client
export { VigilumClient, VigilumApiError } from './client';

// Local analysis
export { ContractAnalyzer, type AnalysisResult } from './analyzer';

// Transaction guard
export { 
  TransactionGuard, 
  type GuardResult, 
  type GuardWarning, 
  type GuardOptions,
  type SimulationResult,
  type StateChange,
  type SimulatedLog,
} from './guard';

// Real-time monitoring
export { 
  ContractMonitor, 
  Watchlist,
  type MonitoringConfig,
  type MonitorAlert,
  type MonitorStats,
  type PendingTransaction,
  type WatchlistEntry,
} from './monitor';

// React hooks factory
export { createVigilumHooks, type VigilumHooks } from './hooks';

// Types
export * from './types';

// Constants
export * from './constants';
