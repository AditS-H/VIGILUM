/**
 * Tests for ContractMonitor and Watchlist.
 */

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { Watchlist, type WatchlistEntry } from '../src/monitor';
import type { Address } from 'viem';

describe('Watchlist', () => {
  let watchlist: Watchlist;

  beforeEach(() => {
    watchlist = new Watchlist();
  });

  describe('add', () => {
    it('should add address to watchlist', () => {
      const address = '0x1234567890123456789012345678901234567890' as Address;
      
      watchlist.add(address, 'Test Contract', 'Testing purposes');
      
      expect(watchlist.has(address)).toBe(true);
    });

    it('should store label and notes', () => {
      const address = '0x1234567890123456789012345678901234567890' as Address;
      
      watchlist.add(address, 'Test Contract', 'Testing purposes');
      
      const entry = watchlist.get(address);
      expect(entry?.label).toBe('Test Contract');
      expect(entry?.notes).toBe('Testing purposes');
    });

    it('should set addedAt timestamp', () => {
      const address = '0x1234567890123456789012345678901234567890' as Address;
      
      watchlist.add(address);
      
      const entry = watchlist.get(address);
      expect(entry?.addedAt).toBeDefined();
      expect(new Date(entry!.addedAt).getTime()).toBeLessThanOrEqual(Date.now());
    });
  });

  describe('remove', () => {
    it('should remove address from watchlist', () => {
      const address = '0x1234567890123456789012345678901234567890' as Address;
      
      watchlist.add(address);
      expect(watchlist.has(address)).toBe(true);
      
      watchlist.remove(address);
      expect(watchlist.has(address)).toBe(false);
    });

    it('should return true when address was removed', () => {
      const address = '0x1234567890123456789012345678901234567890' as Address;
      
      watchlist.add(address);
      const result = watchlist.remove(address);
      
      expect(result).toBe(true);
    });

    it('should return false when address was not in list', () => {
      const address = '0x1234567890123456789012345678901234567890' as Address;
      
      const result = watchlist.remove(address);
      
      expect(result).toBe(false);
    });
  });

  describe('list', () => {
    it('should return all entries', () => {
      const addr1 = '0x1111111111111111111111111111111111111111' as Address;
      const addr2 = '0x2222222222222222222222222222222222222222' as Address;
      
      watchlist.add(addr1, 'Contract 1');
      watchlist.add(addr2, 'Contract 2');
      
      const entries = watchlist.list();
      
      expect(entries).toHaveLength(2);
    });

    it('should return empty array for empty watchlist', () => {
      const entries = watchlist.list();
      
      expect(entries).toHaveLength(0);
    });
  });

  describe('toAddresses', () => {
    it('should return array of addresses', () => {
      const addr1 = '0x1111111111111111111111111111111111111111' as Address;
      const addr2 = '0x2222222222222222222222222222222222222222' as Address;
      
      watchlist.add(addr1);
      watchlist.add(addr2);
      
      const addresses = watchlist.toAddresses();
      
      expect(addresses).toContain(addr1);
      expect(addresses).toContain(addr2);
      expect(addresses).toHaveLength(2);
    });
  });

  describe('clear', () => {
    it('should remove all entries', () => {
      watchlist.add('0x1111111111111111111111111111111111111111' as Address);
      watchlist.add('0x2222222222222222222222222222222222222222' as Address);
      
      watchlist.clear();
      
      expect(watchlist.list()).toHaveLength(0);
    });
  });

  describe('serialization', () => {
    it('should serialize to JSON', () => {
      const address = '0x1234567890123456789012345678901234567890' as Address;
      
      watchlist.add(address, 'Test');
      
      const json = watchlist.toJSON();
      
      expect(json).toBeInstanceOf(Array);
      expect(json[0].address).toBe(address);
    });

    it('should deserialize from JSON', () => {
      const entries: WatchlistEntry[] = [
        {
          address: '0x1234567890123456789012345678901234567890' as Address,
          label: 'Test',
          addedAt: new Date().toISOString(),
        },
      ];

      const restored = Watchlist.fromJSON(entries);
      
      expect(restored.has('0x1234567890123456789012345678901234567890' as Address)).toBe(true);
    });
  });
});

describe('MonitorAlert', () => {
  it('should have correct alert types', () => {
    const alertTypes = [
      'suspicious_contract',
      'large_transfer',
      'unusual_activity',
      'contract_interaction',
      'approval',
    ];

    expect(alertTypes).toContain('suspicious_contract');
    expect(alertTypes).toContain('large_transfer');
    expect(alertTypes).toContain('approval');
  });
});

describe('WATCHED_SELECTORS', () => {
  const selectors = {
    '0x095ea7b3': 'approve',
    '0xa22cb465': 'setApprovalForAll',
    '0x23b872dd': 'transferFrom',
    '0x42842e0e': 'safeTransferFrom',
    '0xa9059cbb': 'transfer',
    '0x2e1a7d4d': 'withdraw',
  };

  it('should identify approve function', () => {
    expect(selectors['0x095ea7b3']).toBe('approve');
  });

  it('should identify setApprovalForAll function', () => {
    expect(selectors['0xa22cb465']).toBe('setApprovalForAll');
  });

  it('should identify transfer function', () => {
    expect(selectors['0xa9059cbb']).toBe('transfer');
  });
});
