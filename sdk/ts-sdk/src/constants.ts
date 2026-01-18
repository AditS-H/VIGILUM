/**
 * SDK constants and configuration.
 */

// Supported chain IDs
export const SUPPORTED_CHAINS = {
  ETHEREUM: 1,
  POLYGON: 137,
  BSC: 56,
  ARBITRUM: 42161,
  OPTIMISM: 10,
  BASE: 8453,
  AVALANCHE: 43114,
  FANTOM: 250,
} as const;

// API endpoints
export const API_ENDPOINTS = {
  SCAN: '/api/v1/scan',
  CONTRACT: '/api/v1/contract',
  ALERTS: '/api/v1/alerts',
  REPUTATION: '/api/v1/reputation',
  REGISTRY: '/api/v1/registry',
} as const;

// Default configuration
export const DEFAULT_CONFIG = {
  timeout: 30000,
  retries: 3,
  baseUrl: 'https://api.vigilum.io',
} as const;

// Risk score thresholds
export const RISK_THRESHOLDS = {
  CRITICAL: 80,
  HIGH: 60,
  MEDIUM: 40,
  LOW: 20,
} as const;

// Known malicious bytecode patterns
export const MALICIOUS_PATTERNS = {
  HONEYPOT_TRANSFER_BLOCK: '0x7f',
  HIDDEN_MINT: '0x40c10f19',
  BLACKLIST_CHECK: '0x404e5811',
} as const;
