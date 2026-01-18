/**
 * Local contract analysis utilities.
 */

import type { Address, Hex, PublicClient } from 'viem';
import { MALICIOUS_PATTERNS, RISK_THRESHOLDS } from './constants';
import { ThreatLevel, VulnerabilityType, type Vulnerability } from './types';

export interface AnalysisResult {
  address: Address;
  riskScore: number;
  threatLevel: typeof ThreatLevel[keyof typeof ThreatLevel];
  findings: Vulnerability[];
  isProxy: boolean;
  isVerified: boolean;
  bytecodeHash: string;
}

export class ContractAnalyzer {
  private readonly client: PublicClient;

  constructor(client: PublicClient) {
    this.client = client;
  }

  /**
   * Perform local analysis on a contract.
   * For deep analysis, use VigilumClient.scan() instead.
   */
  async analyze(address: Address): Promise<AnalysisResult> {
    // Get bytecode
    const bytecode = await this.client.getBytecode({ address });
    
    if (!bytecode) {
      return this.emptyResult(address);
    }

    const findings: Vulnerability[] = [];
    
    // Quick pattern checks
    const patterns = this.checkPatterns(bytecode);
    findings.push(...patterns);

    // Check for proxy pattern
    const isProxy = this.detectProxy(bytecode);
    
    // Calculate risk score
    const riskScore = this.calculateRiskScore(findings);
    const threatLevel = this.getThreatLevel(riskScore);

    // Compute bytecode hash
    const bytecodeHash = await this.hashBytecode(bytecode);

    return {
      address,
      riskScore,
      threatLevel,
      findings,
      isProxy,
      isVerified: false, // Would need to check Etherscan API
      bytecodeHash,
    };
  }

  /**
   * Quick risk check - faster than full analysis.
   */
  async quickCheck(address: Address): Promise<{ 
    riskLevel: 'safe' | 'caution' | 'danger';
    reason: string;
  }> {
    const bytecode = await this.client.getBytecode({ address });
    
    if (!bytecode) {
      return { riskLevel: 'caution', reason: 'Contract not found or EOA' };
    }

    // Check for obvious honeypot patterns
    if (this.hasHoneypotPattern(bytecode)) {
      return { riskLevel: 'danger', reason: 'Potential honeypot detected' };
    }

    // Check for hidden mint
    if (this.hasHiddenMint(bytecode)) {
      return { riskLevel: 'danger', reason: 'Hidden mint function detected' };
    }

    // Check for selfdestruct
    if (bytecode.includes('ff')) {
      return { riskLevel: 'caution', reason: 'Contract has selfdestruct' };
    }

    return { riskLevel: 'safe', reason: 'No obvious red flags' };
  }

  /**
   * Compare two contracts for similarity.
   */
  async compareContracts(
    address1: Address,
    address2: Address
  ): Promise<{ similarity: number; isClone: boolean }> {
    const [bytecode1, bytecode2] = await Promise.all([
      this.client.getBytecode({ address: address1 }),
      this.client.getBytecode({ address: address2 }),
    ]);

    if (!bytecode1 || !bytecode2) {
      return { similarity: 0, isClone: false };
    }

    // Simple Jaccard similarity on opcodes
    const ops1 = this.extractOpcodes(bytecode1);
    const ops2 = this.extractOpcodes(bytecode2);

    const intersection = ops1.filter((op) => ops2.includes(op)).length;
    const union = new Set([...ops1, ...ops2]).size;

    const similarity = union > 0 ? intersection / union : 0;
    const isClone = similarity > 0.95;

    return { similarity, isClone };
  }

  // ═══════════════════════════════════════════════════════════════════════════
  // PRIVATE METHODS
  // ═══════════════════════════════════════════════════════════════════════════

  private emptyResult(address: Address): AnalysisResult {
    return {
      address,
      riskScore: 0,
      threatLevel: ThreatLevel.NONE,
      findings: [],
      isProxy: false,
      isVerified: false,
      bytecodeHash: '0x',
    };
  }

  private checkPatterns(bytecode: Hex): Vulnerability[] {
    const findings: Vulnerability[] = [];

    // Check for DELEGATECALL without proper validation
    if (bytecode.includes('f4') && !this.hasSafeDelegate(bytecode)) {
      findings.push({
        id: crypto.randomUUID(),
        type: VulnerabilityType.ACCESS_CONTROL,
        severity: ThreatLevel.HIGH,
        title: 'Unprotected DELEGATECALL',
        description: 'Contract uses DELEGATECALL which may be vulnerable to delegation attacks',
        confidence: 0.6,
      });
    }

    // Check for tx.origin usage
    if (bytecode.includes('32')) {
      findings.push({
        id: crypto.randomUUID(),
        type: VulnerabilityType.TX_ORIGIN,
        severity: ThreatLevel.MEDIUM,
        title: 'tx.origin Usage',
        description: 'Contract uses tx.origin which can be exploited in phishing attacks',
        confidence: 0.8,
      });
    }

    // Check for block.timestamp usage
    if (bytecode.includes('42')) {
      findings.push({
        id: crypto.randomUUID(),
        type: VulnerabilityType.TIMESTAMP_DEPENDENCY,
        severity: ThreatLevel.LOW,
        title: 'Block Timestamp Dependency',
        description: 'Contract depends on block.timestamp which can be manipulated by miners',
        confidence: 0.5,
      });
    }

    return findings;
  }

  private detectProxy(bytecode: Hex): boolean {
    // EIP-1167 minimal proxy pattern
    if (bytecode.startsWith('0x363d3d373d3d3d363d73')) {
      return true;
    }
    
    // DELEGATECALL with small bytecode suggests proxy
    if (bytecode.includes('f4') && bytecode.length < 500) {
      return true;
    }

    return false;
  }

  private hasHoneypotPattern(bytecode: Hex): boolean {
    // Check for patterns that block transfers
    return bytecode.includes(MALICIOUS_PATTERNS.HONEYPOT_TRANSFER_BLOCK);
  }

  private hasHiddenMint(bytecode: Hex): boolean {
    // Check for mint function selector
    return bytecode.includes(MALICIOUS_PATTERNS.HIDDEN_MINT.slice(2));
  }

  private hasSafeDelegate(bytecode: Hex): boolean {
    // Simplified check - real implementation would be more thorough
    // Check if there's access control before DELEGATECALL
    const delegateIndex = bytecode.indexOf('f4');
    const callerCheck = bytecode.indexOf('33'); // CALLER opcode
    return callerCheck < delegateIndex && callerCheck !== -1;
  }

  private calculateRiskScore(findings: Vulnerability[]): number {
    if (findings.length === 0) return 0;

    const weights: Record<typeof ThreatLevel[keyof typeof ThreatLevel], number> = {
      [ThreatLevel.CRITICAL]: 40,
      [ThreatLevel.HIGH]: 25,
      [ThreatLevel.MEDIUM]: 15,
      [ThreatLevel.LOW]: 5,
      [ThreatLevel.INFO]: 1,
      [ThreatLevel.NONE]: 0,
    };

    let score = 0;
    for (const finding of findings) {
      score += weights[finding.severity] * finding.confidence;
    }

    return Math.min(100, score);
  }

  private getThreatLevel(score: number): typeof ThreatLevel[keyof typeof ThreatLevel] {
    if (score >= RISK_THRESHOLDS.CRITICAL) return ThreatLevel.CRITICAL;
    if (score >= RISK_THRESHOLDS.HIGH) return ThreatLevel.HIGH;
    if (score >= RISK_THRESHOLDS.MEDIUM) return ThreatLevel.MEDIUM;
    if (score >= RISK_THRESHOLDS.LOW) return ThreatLevel.LOW;
    if (score > 0) return ThreatLevel.INFO;
    return ThreatLevel.NONE;
  }

  private extractOpcodes(bytecode: Hex): string[] {
    // Simplified opcode extraction
    const ops: string[] = [];
    const hex = bytecode.slice(2);
    
    for (let i = 0; i < hex.length; i += 2) {
      ops.push(hex.slice(i, i + 2));
    }
    
    return ops;
  }

  private async hashBytecode(bytecode: Hex): Promise<string> {
    const encoder = new TextEncoder();
    const data = encoder.encode(bytecode);
    const hashBuffer = await crypto.subtle.digest('SHA-256', data);
    const hashArray = Array.from(new Uint8Array(hashBuffer));
    return '0x' + hashArray.map((b) => b.toString(16).padStart(2, '0')).join('');
  }
}
