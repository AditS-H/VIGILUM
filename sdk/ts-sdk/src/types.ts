/**
 * Core type definitions for VIGILUM SDK.
 */

import { z } from 'zod';
import type { Address, Hash } from 'viem';

// ═══════════════════════════════════════════════════════════════════════════
// ENUMS
// ═══════════════════════════════════════════════════════════════════════════

export const ThreatLevel = {
  CRITICAL: 'critical',
  HIGH: 'high',
  MEDIUM: 'medium',
  LOW: 'low',
  INFO: 'info',
  NONE: 'none',
} as const;
export type ThreatLevel = (typeof ThreatLevel)[keyof typeof ThreatLevel];

export const VulnerabilityType = {
  REENTRANCY: 'reentrancy',
  INTEGER_OVERFLOW: 'integer_overflow',
  INTEGER_UNDERFLOW: 'integer_underflow',
  ACCESS_CONTROL: 'access_control',
  UNCHECKED_CALL: 'unchecked_external_call',
  TX_ORIGIN: 'tx_origin',
  TIMESTAMP_DEPENDENCY: 'timestamp_dependency',
  FRONTRUNNING: 'frontrunning',
  FLASH_LOAN: 'flash_loan_attack',
  ORACLE_MANIPULATION: 'oracle_manipulation',
  RUG_PULL: 'rug_pull_pattern',
  HONEYPOT: 'honeypot',
  PHISHING: 'phishing_signature',
} as const;
export type VulnerabilityType = (typeof VulnerabilityType)[keyof typeof VulnerabilityType];

export const ScanStatus = {
  PENDING: 'pending',
  RUNNING: 'running',
  COMPLETED: 'completed',
  FAILED: 'failed',
} as const;
export type ScanStatus = (typeof ScanStatus)[keyof typeof ScanStatus];

// ═══════════════════════════════════════════════════════════════════════════
// ZOD SCHEMAS
// ═══════════════════════════════════════════════════════════════════════════

export const VulnerabilitySchema = z.object({
  id: z.string(),
  type: z.nativeEnum(VulnerabilityType),
  severity: z.nativeEnum(ThreatLevel),
  title: z.string(),
  description: z.string(),
  location: z
    .object({
      file: z.string().optional(),
      startLine: z.number(),
      endLine: z.number(),
      snippet: z.string().optional(),
    })
    .optional(),
  remediation: z.string().optional(),
  confidence: z.number().min(0).max(1),
  cwe: z.string().optional(),
});

export const ScanResultSchema = z.object({
  id: z.string(),
  contractAddress: z.string(),
  chainId: z.number(),
  status: z.nativeEnum(ScanStatus),
  riskScore: z.number().min(0).max(100),
  threatLevel: z.nativeEnum(ThreatLevel),
  vulnerabilities: z.array(VulnerabilitySchema),
  metrics: z.object({
    totalIssues: z.number(),
    criticalCount: z.number(),
    highCount: z.number(),
    mediumCount: z.number(),
    lowCount: z.number(),
    infoCount: z.number(),
  }),
  startedAt: z.string().datetime(),
  completedAt: z.string().datetime().optional(),
  durationMs: z.number().optional(),
});

export const ContractInfoSchema = z.object({
  address: z.string(),
  chainId: z.number(),
  name: z.string().optional(),
  bytecodeHash: z.string(),
  isVerified: z.boolean(),
  isBlacklisted: z.boolean(),
  riskScore: z.number().min(0).max(100),
  threatLevel: z.nativeEnum(ThreatLevel),
  lastScanAt: z.string().datetime().optional(),
  labels: z.array(z.string()),
});

// ═══════════════════════════════════════════════════════════════════════════
// TYPES
// ═══════════════════════════════════════════════════════════════════════════

export type Vulnerability = z.infer<typeof VulnerabilitySchema>;
export type ScanResult = z.infer<typeof ScanResultSchema>;
export type ContractInfo = z.infer<typeof ContractInfoSchema>;

export interface ScanOptions {
  /** Types of scans to perform */
  scanTypes?: ('static' | 'ml' | 'symbolic' | 'fuzz')[];
  /** Timeout in seconds */
  timeout?: number;
  /** Include informational findings */
  includeInfo?: boolean;
  /** Custom rules to apply */
  customRules?: string[];
  /** Webhook URL for scan completion */
  webhookUrl?: string;
}

export interface ClientConfig {
  /** VIGILUM API base URL */
  baseUrl: string;
  /** API key for authentication */
  apiKey: string;
  /** Request timeout in milliseconds */
  timeout?: number;
  /** Number of retries on failure */
  retries?: number;
}

export interface ContractQuery {
  /** Contract address */
  address: Address;
  /** Chain ID */
  chainId: number;
}

export interface BatchScanRequest {
  contracts: ContractQuery[];
  options?: ScanOptions;
}

export interface Alert {
  id: string;
  type: 'scan_result' | 'realtime_detection' | 'mempool_threat' | 'anomaly';
  severity: ThreatLevel;
  title: string;
  description: string;
  contractAddress?: Address;
  txHash?: Hash;
  chainId: number;
  createdAt: string;
  metadata?: Record<string, unknown>;
}

export interface ReputationInfo {
  address: Address;
  score: number;
  tier: 'bronze' | 'silver' | 'gold' | 'platinum';
  totalScans: number;
  confirmedVulns: number;
  lastActive: string;
}
