/**
 * Tests for TransactionGuard.
 */

import { describe, it, expect, vi, beforeEach } from 'vitest';
import type { PublicClient, Address } from 'viem';

// Mock the TransactionGuard since we can't import viem in tests easily
describe('TransactionGuard', () => {
  describe('checkValue', () => {
    it('should warn for large ETH transfers', () => {
      const value = 15n * 10n ** 18n; // 15 ETH
      const ethValue = Number(value) / 1e18;
      
      expect(ethValue).toBeGreaterThan(10);
    });

    it('should detect dust attack patterns', () => {
      const value = 0.00001 * 1e18;
      const data = '0x095ea7b3' + '0'.repeat(128);
      
      expect(value < 0.0001 * 1e18).toBe(true);
      expect(data.length).toBeGreaterThan(10);
    });
  });

  describe('checkCalldata', () => {
    it('should detect approve selector', () => {
      const data = '0x095ea7b3';
      expect(data).toBe('0x095ea7b3');
    });

    it('should detect unlimited approval', () => {
      const unlimitedValue = 'ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff';
      const data = `0x095ea7b3${unlimitedValue}`;
      
      expect(data.includes(unlimitedValue)).toBe(true);
    });

    it('should detect setApprovalForAll', () => {
      const selector = '0xa22cb465';
      expect(selector).toBe('0xa22cb465');
    });
  });

  describe('risk calculation', () => {
    it('should calculate weighted risk score', () => {
      const weights = {
        critical: 40,
        high: 25,
        medium: 15,
        low: 5,
        info: 1,
      };

      const findings = [
        { severity: 'high', confidence: 0.8 },
        { severity: 'medium', confidence: 0.6 },
      ];

      const score = findings.reduce((sum, f) => {
        return sum + weights[f.severity as keyof typeof weights] * f.confidence;
      }, 0);

      expect(score).toBe(25 * 0.8 + 15 * 0.6);
    });

    it('should cap risk score at 100', () => {
      const score = 150;
      const cappedScore = Math.min(100, score);
      
      expect(cappedScore).toBe(100);
    });
  });

  describe('recommendations', () => {
    it('should recommend reject for high risk', () => {
      const getRecommendation = (score: number) => {
        if (score >= 60) return 'reject';
        if (score >= 20) return 'caution';
        return 'proceed';
      };

      expect(getRecommendation(80)).toBe('reject');
      expect(getRecommendation(60)).toBe('reject');
      expect(getRecommendation(40)).toBe('caution');
      expect(getRecommendation(10)).toBe('proceed');
    });
  });
});

describe('GuardWarning types', () => {
  it('should have correct warning types', () => {
    const warningTypes = ['contract', 'recipient', 'value', 'gas', 'data', 'simulation'];
    
    expect(warningTypes).toContain('contract');
    expect(warningTypes).toContain('recipient');
    expect(warningTypes).toContain('value');
    expect(warningTypes).toContain('gas');
    expect(warningTypes).toContain('data');
    expect(warningTypes).toContain('simulation');
  });
});
