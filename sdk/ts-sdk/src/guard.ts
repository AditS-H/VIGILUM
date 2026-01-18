/**
 * Transaction Guard - Pre-execution security checks
 */

import type { Address, Hex, PublicClient, TransactionRequest } from 'viem';
import { ThreatLevel } from './types';
import { ContractAnalyzer } from './analyzer';
import { RISK_THRESHOLDS } from './constants';

export interface GuardResult {
  /** Whether the transaction is safe to execute */
  safe: boolean;
  /** Risk level of the transaction */
  riskLevel: ThreatLevel;
  /** Risk score (0-100) */
  riskScore: number;
  /** Warnings and findings */
  warnings: GuardWarning[];
  /** Simulation results */
  simulation?: SimulationResult;
  /** Recommended action */
  recommendation: 'proceed' | 'caution' | 'reject';
}

export interface GuardWarning {
  type: 'contract' | 'recipient' | 'value' | 'gas' | 'data' | 'simulation';
  severity: ThreatLevel;
  message: string;
  details?: Record<string, unknown>;
}

export interface SimulationResult {
  success: boolean;
  gasUsed: bigint;
  returnData?: Hex;
  error?: string;
  stateChanges?: StateChange[];
  logs?: SimulatedLog[];
}

export interface StateChange {
  address: Address;
  slot: Hex;
  before: Hex;
  after: Hex;
}

export interface SimulatedLog {
  address: Address;
  topics: Hex[];
  data: Hex;
}

export interface GuardOptions {
  /** Simulate transaction before checking */
  simulate?: boolean;
  /** Maximum acceptable risk score */
  maxRiskScore?: number;
  /** Check recipient address */
  checkRecipient?: boolean;
  /** Known safe addresses (skip checks) */
  trustedAddresses?: Address[];
}

const DEFAULT_OPTIONS: GuardOptions = {
  simulate: true,
  maxRiskScore: 60,
  checkRecipient: true,
  trustedAddresses: [],
};

export class TransactionGuard {
  private readonly client: PublicClient;
  private readonly analyzer: ContractAnalyzer;
  private readonly options: GuardOptions;

  constructor(
    client: PublicClient,
    options?: GuardOptions
  ) {
    this.client = client;
    this.analyzer = new ContractAnalyzer(client);
    this.options = { ...DEFAULT_OPTIONS, ...options };
  }

  /**
   * Check a transaction for security risks before execution.
   */
  async check(tx: TransactionRequest): Promise<GuardResult> {
    const warnings: GuardWarning[] = [];
    let riskScore = 0;

    // Skip checks for trusted addresses
    if (tx.to && this.options.trustedAddresses?.includes(tx.to)) {
      return this.safeResult();
    }

    // Check recipient contract
    if (tx.to && this.options.checkRecipient) {
      const recipientWarnings = await this.checkRecipient(tx.to);
      warnings.push(...recipientWarnings);
      riskScore += this.calculateWarningsRisk(recipientWarnings);
    }

    // Check for suspicious value transfers
    if (tx.value && tx.value > 0n) {
      const valueWarnings = this.checkValue(tx);
      warnings.push(...valueWarnings);
      riskScore += this.calculateWarningsRisk(valueWarnings);
    }

    // Check calldata
    if (tx.data && tx.data !== '0x') {
      const dataWarnings = this.checkCalldata(tx.data);
      warnings.push(...dataWarnings);
      riskScore += this.calculateWarningsRisk(dataWarnings);
    }

    // Check gas
    if (tx.gas) {
      const gasWarnings = this.checkGas(tx);
      warnings.push(...gasWarnings);
      riskScore += this.calculateWarningsRisk(gasWarnings);
    }

    // Simulate transaction
    let simulation: SimulationResult | undefined;
    if (this.options.simulate && tx.to) {
      simulation = await this.simulateTransaction(tx);
      if (simulation) {
        const simWarnings = this.analyzeSimulation(simulation);
        warnings.push(...simWarnings);
        riskScore += this.calculateWarningsRisk(simWarnings);
      }
    }

    // Calculate final risk level
    riskScore = Math.min(100, riskScore);
    const riskLevel = this.getRiskLevel(riskScore);
    const recommendation = this.getRecommendation(riskScore);
    const safe = riskScore < (this.options.maxRiskScore ?? 60);

    return {
      safe,
      riskLevel,
      riskScore,
      warnings,
      simulation,
      recommendation,
    };
  }

  /**
   * Quick check - faster but less thorough.
   */
  async quickCheck(tx: TransactionRequest): Promise<{
    safe: boolean;
    reason: string;
  }> {
    // Skip if trusted
    if (tx.to && this.options.trustedAddresses?.includes(tx.to)) {
      return { safe: true, reason: 'Trusted address' };
    }

    // Check if sending to contract
    if (tx.to) {
      const code = await this.client.getBytecode({ address: tx.to });
      
      if (code) {
        // Quick analysis
        const quickResult = await this.analyzer.quickCheck(tx.to);
        if (quickResult.riskLevel === 'danger') {
          return { safe: false, reason: quickResult.reason };
        }
      }
    }

    // Check for approval to suspicious contract
    if (tx.data?.startsWith('0x095ea7b3')) {
      const spender = ('0x' + tx.data.slice(34, 74)) as Address;
      const spenderCheck = await this.analyzer.quickCheck(spender);
      if (spenderCheck.riskLevel === 'danger') {
        return { safe: false, reason: 'Approving dangerous contract' };
      }
    }

    return { safe: true, reason: 'No obvious risks detected' };
  }

  // ═══════════════════════════════════════════════════════════════════════════
  // PRIVATE METHODS
  // ═══════════════════════════════════════════════════════════════════════════

  private async checkRecipient(address: Address): Promise<GuardWarning[]> {
    const warnings: GuardWarning[] = [];
    
    try {
      const analysis = await this.analyzer.analyze(address);
      
      if (analysis.riskScore >= RISK_THRESHOLDS.HIGH) {
        warnings.push({
          type: 'recipient',
          severity: ThreatLevel.HIGH,
          message: `Recipient contract has high risk score (${analysis.riskScore})`,
          details: { findings: analysis.findings.length },
        });
      } else if (analysis.riskScore >= RISK_THRESHOLDS.MEDIUM) {
        warnings.push({
          type: 'recipient',
          severity: ThreatLevel.MEDIUM,
          message: `Recipient contract has elevated risk score (${analysis.riskScore})`,
        });
      }

      if (analysis.isProxy) {
        warnings.push({
          type: 'recipient',
          severity: ThreatLevel.LOW,
          message: 'Recipient is a proxy contract - implementation may change',
        });
      }
    } catch {
      // Contract may not exist (EOA)
    }

    return warnings;
  }

  private checkValue(tx: TransactionRequest): GuardWarning[] {
    const warnings: GuardWarning[] = [];
    
    if (!tx.value) return warnings;

    const ethValue = Number(tx.value) / 1e18;

    // Warning for large ETH transfers
    if (ethValue > 10) {
      warnings.push({
        type: 'value',
        severity: ThreatLevel.MEDIUM,
        message: `Large value transfer: ${ethValue.toFixed(4)} ETH`,
      });
    }

    // Warning for dust attacks
    if (ethValue < 0.0001 && tx.data && tx.data.length > 10) {
      warnings.push({
        type: 'value',
        severity: ThreatLevel.LOW,
        message: 'Small value with calldata - potential dust attack',
      });
    }

    return warnings;
  }

  private checkCalldata(data: Hex): GuardWarning[] {
    const warnings: GuardWarning[] = [];
    const selector = data.slice(0, 10);

    // Known dangerous selectors
    const dangerousSelectors: Record<string, string> = {
      '0x095ea7b3': 'approve - Token approval',
      '0xa22cb465': 'setApprovalForAll - NFT collection approval',
      '0x23b872dd': 'transferFrom - May drain approved tokens',
      '0x42842e0e': 'safeTransferFrom - May transfer NFTs',
      '0xf305d719': 'addLiquidityETH - DEX interaction',
      '0x18cbafe5': 'swapExactTokensForETH - DEX swap',
    };

    if (selector in dangerousSelectors) {
      warnings.push({
        type: 'data',
        severity: ThreatLevel.INFO,
        message: `Known function: ${dangerousSelectors[selector]}`,
        details: { selector },
      });
    }

    // Check for unlimited approval (max uint256)
    if (selector === '0x095ea7b3' && data.includes('ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff')) {
      warnings.push({
        type: 'data',
        severity: ThreatLevel.HIGH,
        message: 'Unlimited token approval requested',
      });
    }

    return warnings;
  }

  private checkGas(tx: TransactionRequest): GuardWarning[] {
    const warnings: GuardWarning[] = [];
    
    if (!tx.gas) return warnings;

    // Warning for extremely high gas limit
    if (tx.gas > 10_000_000n) {
      warnings.push({
        type: 'gas',
        severity: ThreatLevel.MEDIUM,
        message: 'Unusually high gas limit',
        details: { gas: tx.gas.toString() },
      });
    }

    return warnings;
  }

  private async simulateTransaction(tx: TransactionRequest): Promise<SimulationResult | undefined> {
    try {
      const result = await this.client.call({
        to: tx.to,
        data: tx.data,
        value: tx.value,
        gas: tx.gas,
        account: tx.from,
      });

      return {
        success: true,
        gasUsed: 0n, // Would need trace for accurate gas
        returnData: result.data,
      };
    } catch (error) {
      return {
        success: false,
        gasUsed: 0n,
        error: error instanceof Error ? error.message : 'Unknown error',
      };
    }
  }

  private analyzeSimulation(sim: SimulationResult): GuardWarning[] {
    const warnings: GuardWarning[] = [];

    if (!sim.success) {
      warnings.push({
        type: 'simulation',
        severity: ThreatLevel.HIGH,
        message: `Transaction would revert: ${sim.error || 'Unknown reason'}`,
      });
    }

    return warnings;
  }

  private calculateWarningsRisk(warnings: GuardWarning[]): number {
    const weights: Record<ThreatLevel, number> = {
      [ThreatLevel.CRITICAL]: 40,
      [ThreatLevel.HIGH]: 25,
      [ThreatLevel.MEDIUM]: 15,
      [ThreatLevel.LOW]: 5,
      [ThreatLevel.INFO]: 1,
      [ThreatLevel.NONE]: 0,
    };

    return warnings.reduce((sum, w) => sum + weights[w.severity], 0);
  }

  private getRiskLevel(score: number): ThreatLevel {
    if (score >= RISK_THRESHOLDS.CRITICAL) return ThreatLevel.CRITICAL;
    if (score >= RISK_THRESHOLDS.HIGH) return ThreatLevel.HIGH;
    if (score >= RISK_THRESHOLDS.MEDIUM) return ThreatLevel.MEDIUM;
    if (score >= RISK_THRESHOLDS.LOW) return ThreatLevel.LOW;
    if (score > 0) return ThreatLevel.INFO;
    return ThreatLevel.NONE;
  }

  private getRecommendation(score: number): 'proceed' | 'caution' | 'reject' {
    if (score >= RISK_THRESHOLDS.HIGH) return 'reject';
    if (score >= RISK_THRESHOLDS.LOW) return 'caution';
    return 'proceed';
  }

  private safeResult(): GuardResult {
    return {
      safe: true,
      riskLevel: ThreatLevel.NONE,
      riskScore: 0,
      warnings: [],
      recommendation: 'proceed',
    };
  }
}
