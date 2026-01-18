/**
 * React hooks for VIGILUM SDK integration.
 * 
 * Note: This file requires React and wagmi as peer dependencies.
 */

import type { Address, PublicClient } from 'viem';
import type { ScanResult, ContractInfo, Alert } from './types';
import { VigilumClient } from './client';
import { ContractAnalyzer, type AnalysisResult } from './analyzer';
import { TransactionGuard, type GuardResult, type GuardOptions } from './guard';
import { ContractMonitor, type MonitorAlert, type MonitorStats } from './monitor';

// ═══════════════════════════════════════════════════════════════════════════
// HOOK RETURN TYPES
// ═══════════════════════════════════════════════════════════════════════════

export interface UseVigilumResult {
  client: VigilumClient;
  scan: (address: Address, chainId: number) => Promise<ScanResult>;
  getContract: (address: Address, chainId: number) => Promise<ContractInfo>;
  getAlerts: (address: Address, chainId?: number) => Promise<Alert[]>;
  isLoading: boolean;
  error: Error | null;
}

export interface UseScanResult {
  data: ScanResult | null;
  isLoading: boolean;
  error: Error | null;
  refetch: () => Promise<void>;
}

export interface UseAnalysisResult {
  data: AnalysisResult | null;
  isLoading: boolean;
  error: Error | null;
  analyze: (address: Address) => Promise<AnalysisResult>;
}

export interface UseGuardResult {
  guard: TransactionGuard;
  check: (tx: unknown) => Promise<GuardResult>;
  quickCheck: (tx: unknown) => Promise<{ safe: boolean; reason: string }>;
  isLoading: boolean;
  lastResult: GuardResult | null;
}

export interface UseMonitorResult {
  monitor: ContractMonitor | null;
  alerts: MonitorAlert[];
  stats: MonitorStats | null;
  start: () => Promise<void>;
  stop: () => void;
  addAddress: (address: Address) => void;
  removeAddress: (address: Address) => void;
  isRunning: boolean;
}

// ═══════════════════════════════════════════════════════════════════════════
// HOOK FACTORIES
// ═══════════════════════════════════════════════════════════════════════════

/**
 * Creates VIGILUM hooks compatible with any React state management.
 * This factory pattern allows the SDK to work with different React setups.
 * 
 * @example
 * // Usage with React
 * const { useVigilum, useScan, useAnalysis } = createVigilumHooks(React);
 */
export function createVigilumHooks(React: {
  useState: <T>(initial: T) => [T, (v: T | ((prev: T) => T)) => void];
  useEffect: (effect: () => void | (() => void), deps?: unknown[]) => void;
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  useCallback: <T extends (...args: any[]) => any>(cb: T, deps: unknown[]) => T;
  useMemo: <T>(factory: () => T, deps: unknown[]) => T;
  useRef: <T>(initial: T) => { current: T };
}) {
  const { useState, useEffect, useCallback, useMemo, useRef } = React;

  /**
   * Hook for VIGILUM API client.
   */
  function useVigilum(config: { apiKey: string; baseUrl?: string }): UseVigilumResult {
    const [isLoading, setIsLoading] = useState(false);
    const [error, setError] = useState<Error | null>(null);

    const client = useMemo(() => new VigilumClient({
      apiKey: config.apiKey,
      baseUrl: config.baseUrl || 'https://api.vigilum.io',
    }), [config.apiKey, config.baseUrl]);

    const scan = useCallback(async (address: Address, chainId: number) => {
      setIsLoading(true);
      setError(null);
      try {
        const result = await client.scan({ address, chainId });
        return result;
      } catch (err) {
        setError(err instanceof Error ? err : new Error('Unknown error'));
        throw err;
      } finally {
        setIsLoading(false);
      }
    }, [client]);

    const getContract = useCallback(async (address: Address, chainId: number) => {
      setIsLoading(true);
      setError(null);
      try {
        const result = await client.getContractInfo({ address, chainId });
        return result;
      } catch (err) {
        setError(err instanceof Error ? err : new Error('Unknown error'));
        throw err;
      } finally {
        setIsLoading(false);
      }
    }, [client]);

    const getAlerts = useCallback(async (address: Address, chainId?: number) => {
      setIsLoading(true);
      setError(null);
      try {
        const result = await client.getAlerts(address, chainId);
        return result;
      } catch (err) {
        setError(err instanceof Error ? err : new Error('Unknown error'));
        throw err;
      } finally {
        setIsLoading(false);
      }
    }, [client]);

    return {
      client,
      scan,
      getContract,
      getAlerts,
      isLoading,
      error,
    };
  }

  /**
   * Hook for scanning a specific contract.
   */
  function useScan(
    client: VigilumClient,
    address: Address | undefined,
    chainId: number,
    options?: { enabled?: boolean }
  ): UseScanResult {
    const [data, setData] = useState<ScanResult | null>(null);
    const [isLoading, setIsLoading] = useState(false);
    const [error, setError] = useState<Error | null>(null);

    const fetch = useCallback(async () => {
      if (!address) return;
      
      setIsLoading(true);
      setError(null);
      try {
        const result = await client.scan({ address, chainId });
        // Poll for completion
        const finalResult = await client.waitForScan(result.id);
        setData(finalResult);
      } catch (err) {
        setError(err instanceof Error ? err : new Error('Unknown error'));
      } finally {
        setIsLoading(false);
      }
    }, [client, address, chainId]);

    useEffect(() => {
      if (options?.enabled !== false && address) {
        fetch();
      }
    }, [fetch, options?.enabled, address]);

    return {
      data,
      isLoading,
      error,
      refetch: fetch,
    };
  }

  /**
   * Hook for local contract analysis.
   */
  function useAnalysis(publicClient: PublicClient): UseAnalysisResult {
    const [data, setData] = useState<AnalysisResult | null>(null);
    const [isLoading, setIsLoading] = useState(false);
    const [error, setError] = useState<Error | null>(null);

    const analyzer = useMemo(() => new ContractAnalyzer(publicClient), [publicClient]);

    const analyze = useCallback(async (address: Address) => {
      setIsLoading(true);
      setError(null);
      try {
        const result = await analyzer.analyze(address);
        setData(result);
        return result;
      } catch (err) {
        const e = err instanceof Error ? err : new Error('Unknown error');
        setError(e);
        throw e;
      } finally {
        setIsLoading(false);
      }
    }, [analyzer]);

    return {
      data,
      isLoading,
      error,
      analyze,
    };
  }

  /**
   * Hook for transaction guard.
   */
  function useGuard(
    publicClient: PublicClient,
    options?: GuardOptions
  ): UseGuardResult {
    const [isLoading, setIsLoading] = useState(false);
    const [lastResult, setLastResult] = useState<GuardResult | null>(null);

    const guard = useMemo(
      () => new TransactionGuard(publicClient, options),
      [publicClient, options]
    );

    const check = useCallback(async (tx: unknown) => {
      setIsLoading(true);
      try {
        const result = await guard.check(tx as any);
        setLastResult(result);
        return result;
      } finally {
        setIsLoading(false);
      }
    }, [guard]);

    const quickCheck = useCallback(async (tx: unknown) => {
      setIsLoading(true);
      try {
        return await guard.quickCheck(tx as any);
      } finally {
        setIsLoading(false);
      }
    }, [guard]);

    return {
      guard,
      check,
      quickCheck,
      isLoading,
      lastResult,
    };
  }

  /**
   * Hook for real-time monitoring.
   */
  function useMonitor(
    publicClient: PublicClient,
    addresses: Address[],
    chainId: number,
    options?: { autoStart?: boolean }
  ): UseMonitorResult {
    const [alerts, setAlerts] = useState<MonitorAlert[]>([]);
    const [stats, setStats] = useState<MonitorStats | null>(null);
    const [isRunning, setIsRunning] = useState(false);
    const monitorRef = useRef<ContractMonitor | null>(null);

    useEffect(() => {
      const monitor = new ContractMonitor({
        publicClient,
        addresses,
        chainId,
        onAlert: (alert) => {
          setAlerts((prev: MonitorAlert[]) => [alert, ...prev].slice(0, 100));
        },
      });

      monitorRef.current = monitor;

      if (options?.autoStart) {
        monitor.start();
        setIsRunning(true);
      }

      // Update stats periodically
      const interval = setInterval(() => {
        if (monitorRef.current) {
          setStats(monitorRef.current.getStats());
        }
      }, 1000);

      return () => {
        monitor.stop();
        clearInterval(interval);
      };
    }, [publicClient, chainId, JSON.stringify(addresses), options?.autoStart]);

    const start = useCallback(async () => {
      if (monitorRef.current && !isRunning) {
        await monitorRef.current.start();
        setIsRunning(true);
      }
    }, [isRunning]);

    const stop = useCallback(() => {
      if (monitorRef.current && isRunning) {
        monitorRef.current.stop();
        setIsRunning(false);
      }
    }, [isRunning]);

    const addAddress = useCallback((address: Address) => {
      monitorRef.current?.addAddress(address);
    }, []);

    const removeAddress = useCallback((address: Address) => {
      monitorRef.current?.removeAddress(address);
    }, []);

    return {
      monitor: monitorRef.current,
      alerts,
      stats,
      start,
      stop,
      addAddress,
      removeAddress,
      isRunning,
    };
  }

  /**
   * Hook for checking if an address is safe to interact with.
   */
  function useIsSafe(
    client: VigilumClient,
    address: Address | undefined,
    chainId: number,
    options?: { maxRiskScore?: number }
  ): { isSafe: boolean | null; riskScore: number | null; isLoading: boolean } {
    const [isSafe, setIsSafe] = useState<boolean | null>(null);
    const [riskScore, setRiskScore] = useState<number | null>(null);
    const [isLoading, setIsLoading] = useState(false);

    useEffect(() => {
      if (!address) {
        setIsSafe(null);
        setRiskScore(null);
        return;
      }

      setIsLoading(true);
      client.getContractInfo({ address, chainId })
        .then((info) => {
          setRiskScore(info.riskScore);
          setIsSafe(info.riskScore < (options?.maxRiskScore ?? 60));
        })
        .catch(() => {
          setIsSafe(null);
          setRiskScore(null);
        })
        .finally(() => {
          setIsLoading(false);
        });
    }, [client, address, chainId, options?.maxRiskScore]);

    return { isSafe, riskScore, isLoading };
  }

  return {
    useVigilum,
    useScan,
    useAnalysis,
    useGuard,
    useMonitor,
    useIsSafe,
  };
}

// ═══════════════════════════════════════════════════════════════════════════
// TYPE EXPORTS
// ═══════════════════════════════════════════════════════════════════════════

export type VigilumHooks = ReturnType<typeof createVigilumHooks>;
