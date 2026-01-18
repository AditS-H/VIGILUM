// Package scanner provides pattern-based vulnerability detection.
package scanner

import (
	"context"
	"regexp"
	"strings"

	"github.com/vigilum/backend/internal/domain"
)

// VulnerabilityPattern defines a pattern-based vulnerability detector.
type VulnerabilityPattern struct {
	ID          string
	Name        string
	Description string
	Severity    domain.ThreatLevel
	VulnType    domain.VulnType
	Category    string
	Confidence  float64

	// Source patterns (regex for source code)
	SourcePatterns []string

	// Bytecode patterns (hex patterns)
	BytecodePatterns []string

	// Negation patterns (if found, vulnerability doesn't apply)
	SafePatterns []string
}

// compiledPattern is a pre-compiled vulnerability pattern.
type compiledPattern struct {
	VulnerabilityPattern
	sourceRegexps []regexp.Regexp
	safeRegexps   []regexp.Regexp
}

// PatternDetector implements pattern-based vulnerability detection.
type PatternDetector struct {
	patterns []compiledPattern
}

// NewPatternDetector creates a detector with all built-in patterns.
func NewPatternDetector() *PatternDetector {
	pd := &PatternDetector{
		patterns: make([]compiledPattern, 0),
	}
	pd.registerBuiltinPatterns()
	return pd
}

// DetectFromSource scans source code for vulnerability patterns.
func (pd *PatternDetector) DetectFromSource(ctx context.Context, source string) []domain.Vulnerability {
	vulns := make([]domain.Vulnerability, 0)

	for _, cp := range pd.patterns {
		for _, re := range cp.sourceRegexps {
			matches := re.FindAllStringIndex(source, -1)
			if len(matches) == 0 {
				continue
			}

			// Check if safe patterns negate this finding
			isSafe := false
			for _, safeRe := range cp.safeRegexps {
				if safeRe.MatchString(source) {
					isSafe = true
					break
				}
			}
			if isSafe {
				continue
			}

			// Create vulnerability for each match
			for _, match := range matches {
				line := countLines(source[:match[0]])
				vulns = append(vulns, domain.Vulnerability{
					Type:        cp.VulnType,
					Severity:    cp.Severity,
					Confidence:  cp.Confidence,
					Title:       cp.Name,
					Description: cp.Description,
					Location:    domain.CodeLocation{StartLine: line},
					Remediation: getRemediation(cp.ID),
				})
			}
		}
	}

	return vulns
}

// registerBuiltinPatterns adds all known vulnerability patterns.
func (pd *PatternDetector) registerBuiltinPatterns() {
	patterns := []VulnerabilityPattern{
		// ═══════════════════════════════════════════════════════════════════════
		// REENTRANCY
		// ═══════════════════════════════════════════════════════════════════════
		{
			ID:          "REENTRANCY-ETH",
			Name:        "Reentrancy via ETH Transfer",
			Description: "External call sends ETH before state update, enabling reentrancy attack",
			Severity:    domain.ThreatLevelCritical,
			VulnType:    domain.VulnReentrancy,
			Category:    "reentrancy",
			Confidence:  0.8,
			SourcePatterns: []string{
				`\.call\{value:\s*\w+\}\s*\(""\)`,          // .call{value: amount}("")
				`\.call\.value\(\w+\)\s*\(`,                // .call.value(amount)(
				`\.send\(\w+\)`,                            // .send(amount)
				`\.transfer\(\w+\)`,                        // .transfer(amount)
			},
			SafePatterns: []string{
				`ReentrancyGuard`,                          // Using OZ ReentrancyGuard
				`nonReentrant`,                             // Using nonReentrant modifier
				`mutex`,                                    // Custom mutex pattern
				`locked\s*=\s*true`,                        // Manual lock
			},
		},
		{
			ID:          "REENTRANCY-CALLBACK",
			Name:        "Reentrancy via Callback",
			Description: "ERC callback functions may allow reentrancy",
			Severity:    domain.ThreatLevelHigh,
			VulnType:    domain.VulnReentrancy,
			Category:    "reentrancy",
			Confidence:  0.6,
			SourcePatterns: []string{
				`onERC721Received`,
				`onERC1155Received`,
				`onERC1155BatchReceived`,
				`tokensReceived`,
			},
			SafePatterns: []string{
				`ReentrancyGuard`,
				`nonReentrant`,
			},
		},

		// ═══════════════════════════════════════════════════════════════════════
		// ACCESS CONTROL
		// ═══════════════════════════════════════════════════════════════════════
		{
			ID:          "TX-ORIGIN",
			Name:        "tx.origin Authentication",
			Description: "Using tx.origin for authentication is vulnerable to phishing attacks",
			Severity:    domain.ThreatLevelHigh,
			VulnType:    domain.VulnAccessControl,
			Category:    "access-control",
			Confidence:  0.9,
			SourcePatterns: []string{
				`tx\.origin\s*==`,
				`require\s*\(\s*tx\.origin`,
				`if\s*\(\s*tx\.origin`,
			},
			SafePatterns: []string{
				`tx\.origin\s*==\s*msg\.sender`, // This is a valid pattern for bot protection
			},
		},
		{
			ID:          "MISSING-OWNER-CHECK",
			Name:        "Missing Owner Check",
			Description: "Function may be missing access control for sensitive operation",
			Severity:    domain.ThreatLevelHigh,
			VulnType:    domain.VulnAccessControl,
			Category:    "access-control",
			Confidence:  0.5,
			SourcePatterns: []string{
				`function\s+withdraw\s*\([^)]*\)\s*(external|public)(?![^{]*only)`,
				`function\s+mint\s*\([^)]*\)\s*(external|public)(?![^{]*only)`,
				`function\s+burn\s*\([^)]*\)\s*(external|public)(?![^{]*only)`,
				`function\s+setOwner\s*\([^)]*\)\s*(external|public)(?![^{]*only)`,
			},
			SafePatterns: []string{
				`onlyOwner`,
				`onlyAdmin`,
				`onlyRole`,
			},
		},
		{
			ID:          "UNPROTECTED-SELFDESTRUCT",
			Name:        "Unprotected Self-Destruct",
			Description: "selfdestruct may be callable by attackers",
			Severity:    domain.ThreatLevelCritical,
			VulnType:    domain.VulnAccessControl,
			Category:    "access-control",
			Confidence:  0.7,
			SourcePatterns: []string{
				`selfdestruct\s*\(`,
			},
			SafePatterns: []string{
				`onlyOwner.*selfdestruct`,
				`require.*owner.*selfdestruct`,
			},
		},

		// ═══════════════════════════════════════════════════════════════════════
		// INTEGER ISSUES
		// ═══════════════════════════════════════════════════════════════════════
		{
			ID:          "OVERFLOW-UNCHECKED",
			Name:        "Unchecked Arithmetic",
			Description: "Unchecked block may cause overflow/underflow in pre-0.8.0 behavior",
			Severity:    domain.ThreatLevelMedium,
			VulnType:    domain.VulnOverflow,
			Category:    "arithmetic",
			Confidence:  0.7,
			SourcePatterns: []string{
				`unchecked\s*\{`,
			},
			SafePatterns: []string{}, // May be intentional for gas optimization
		},
		{
			ID:          "DIVISION-BEFORE-MULTIPLY",
			Name:        "Division Before Multiplication",
			Description: "Division before multiplication can cause precision loss",
			Severity:    domain.ThreatLevelMedium,
			VulnType:    domain.VulnPrecisionLoss,
			Category:    "arithmetic",
			Confidence:  0.6,
			SourcePatterns: []string{
				`\w+\s*/\s*\w+\s*\*\s*\w+`,  // x / y * z
			},
			SafePatterns: []string{},
		},

		// ═══════════════════════════════════════════════════════════════════════
		// EXTERNAL CALLS
		// ═══════════════════════════════════════════════════════════════════════
		{
			ID:          "UNCHECKED-CALL",
			Name:        "Unchecked External Call",
			Description: "Return value of external call not checked",
			Severity:    domain.ThreatLevelHigh,
			VulnType:    domain.VulnLogicError,
			Category:    "external-calls",
			Confidence:  0.7,
			SourcePatterns: []string{
				`\.call\{[^}]*\}\s*\([^)]*\)\s*;(?!\s*(require|if|bool\s+success))`,
				`\.delegatecall\s*\([^)]*\)\s*;(?!\s*(require|if|bool\s+success))`,
				`\.staticcall\s*\([^)]*\)\s*;(?!\s*(require|if|bool\s+success))`,
			},
			SafePatterns: []string{
				`\(bool\s+\w*,\s*\)\s*=\s*.*\.call`,  // (bool success, ) = ...
			},
		},
		{
			ID:          "ARBITRARY-DELEGATECALL",
			Name:        "Arbitrary Delegatecall Target",
			Description: "delegatecall target may be controllable by attacker",
			Severity:    domain.ThreatLevelCritical,
			VulnType:    domain.VulnLogicError,
			Category:    "external-calls",
			Confidence:  0.6,
			SourcePatterns: []string{
				`\.delegatecall\s*\(\s*\w+`,          // delegatecall with variable
			},
			SafePatterns: []string{
				`implementation\.delegatecall`,       // Proxy pattern
				`_implementation\(\)\.delegatecall`,  // OZ proxy pattern
			},
		},

		// ═══════════════════════════════════════════════════════════════════════
		// RANDOMNESS
		// ═══════════════════════════════════════════════════════════════════════
		{
			ID:          "WEAK-RANDOMNESS",
			Name:        "Weak Randomness Source",
			Description: "Block variables used for randomness can be manipulated by miners",
			Severity:    domain.ThreatLevelMedium,
			VulnType:    domain.VulnWeakRandomness,
			Category:    "randomness",
			Confidence:  0.9,
			SourcePatterns: []string{
				`block\.timestamp.*keccak256`,
				`block\.number.*keccak256`,
				`blockhash.*keccak256`,
				`block\.difficulty`,
				`block\.prevrandao`,
			},
			SafePatterns: []string{
				`VRFCoordinator`,  // Using Chainlink VRF
				`RandomizerAPI`,   // Using other secure randomness
			},
		},

		// ═══════════════════════════════════════════════════════════════════════
		// DENIAL OF SERVICE
		// ═══════════════════════════════════════════════════════════════════════
		{
			ID:          "UNBOUNDED-LOOP",
			Name:        "Unbounded Loop",
			Description: "Loop over unbounded array may cause out-of-gas",
			Severity:    domain.ThreatLevelMedium,
			VulnType:    domain.VulnDOS,
			Category:    "dos",
			Confidence:  0.6,
			SourcePatterns: []string{
				`for\s*\([^;]*;\s*\w+\s*<\s*\w+\.length`,  // for (i = 0; i < arr.length; ...)
			},
			SafePatterns: []string{
				`pagination`,
				`batch`,
			},
		},

		// ═══════════════════════════════════════════════════════════════════════
		// ORACLE / MEV
		// ═══════════════════════════════════════════════════════════════════════
		{
			ID:          "PRICE-MANIPULATION",
			Name:        "Potential Price Manipulation",
			Description: "Single-block price queries vulnerable to manipulation",
			Severity:    domain.ThreatLevelHigh,
			VulnType:    domain.VulnOracleManipulation,
			Category:    "oracle",
			Confidence:  0.5,
			SourcePatterns: []string{
				`getReserves\(\)`,                    // Uniswap V2 reserves
				`slot0\(\)`,                          // Uniswap V3 spot price
				`latestAnswer\(\)`,                   // Chainlink without staleness check
			},
			SafePatterns: []string{
				`TWAP`,
				`twap`,
				`observe`,                            // Uniswap V3 TWAP
				`latestRoundData`,                    // Chainlink with full data
			},
		},
		{
			ID:          "FLASHLOAN-PATTERN",
			Name:        "Flash Loan Integration",
			Description: "Contract integrates with flash loans - verify protection",
			Severity:    domain.ThreatLevelInfo,
			VulnType:    domain.VulnFlashLoan,
			Category:    "flash-loan",
			Confidence:  1.0,
			SourcePatterns: []string{
				`onFlashLoan`,
				`flashLoan`,
				`executeOperation`,                   // Aave flash loan callback
			},
			SafePatterns: []string{},
		},

		// ═══════════════════════════════════════════════════════════════════════
		// UPGRADABILITY
		// ═══════════════════════════════════════════════════════════════════════
		{
			ID:          "UNPROTECTED-UPGRADE",
			Name:        "Unprotected Upgrade Function",
			Description: "Upgrade function may lack proper access control",
			Severity:    domain.ThreatLevelCritical,
			VulnType:    domain.VulnAccessControl,
			Category:    "upgradability",
			Confidence:  0.6,
			SourcePatterns: []string{
				`function\s+upgrade\w*\s*\([^)]*\)\s*(external|public)`,
				`function\s+_setImplementation\s*\(`,
			},
			SafePatterns: []string{
				`onlyOwner`,
				`onlyAdmin`,
				`onlyProxy`,
			},
		},
		{
			ID:          "STORAGE-COLLISION",
			Name:        "Potential Storage Collision",
			Description: "Proxy pattern may have storage collision risk",
			Severity:    domain.ThreatLevelHigh,
			VulnType:    domain.VulnStorageCollision,
			Category:    "upgradability",
			Confidence:  0.4,
			SourcePatterns: []string{
				`delegatecall.*implementation`,
				`ERC1967`,
				`TransparentProxy`,
			},
			SafePatterns: []string{
				`ERC1967Upgrade`,                    // OZ implementation
				`@openzeppelin`,
			},
		},

		// ═══════════════════════════════════════════════════════════════════════
		// INFORMATIONAL
		// ═══════════════════════════════════════════════════════════════════════
		{
			ID:          "DEPRECATED-FUNCTION",
			Name:        "Deprecated Function Usage",
			Description: "Contract uses deprecated Solidity functions",
			Severity:    domain.ThreatLevelInfo,
			VulnType:    domain.VulnLogicError,
			Category:    "best-practice",
			Confidence:  1.0,
			SourcePatterns: []string{
				`sha3\(`,                            // Use keccak256
				`suicide\(`,                         // Use selfdestruct
				`block\.blockhash`,                  // Use blockhash
				`msg\.gas`,                          // Use gasleft()
			},
			SafePatterns: []string{},
		},
		{
			ID:          "FLOATING-PRAGMA",
			Name:        "Floating Pragma",
			Description: "Contract uses floating pragma version",
			Severity:    domain.ThreatLevelInfo,
			VulnType:    domain.VulnLogicError,
			Category:    "best-practice",
			Confidence:  1.0,
			SourcePatterns: []string{
				`pragma solidity\s*\^`,              // ^0.8.0
				`pragma solidity\s*>=`,              // >=0.8.0
			},
			SafePatterns: []string{},
		},
	}

	// Compile patterns
	for _, p := range patterns {
		cp := compiledPattern{
			VulnerabilityPattern: p,
			sourceRegexps:        make([]regexp.Regexp, 0, len(p.SourcePatterns)),
			safeRegexps:          make([]regexp.Regexp, 0, len(p.SafePatterns)),
		}

		for _, pattern := range p.SourcePatterns {
			if re, err := regexp.Compile(pattern); err == nil {
				cp.sourceRegexps = append(cp.sourceRegexps, *re)
			}
		}

		for _, pattern := range p.SafePatterns {
			if re, err := regexp.Compile(pattern); err == nil {
				cp.safeRegexps = append(cp.safeRegexps, *re)
			}
		}

		pd.patterns = append(pd.patterns, cp)
	}
}

// countLines returns line number at position.
func countLines(s string) int {
	return strings.Count(s, "\n") + 1
}

// formatLocation formats location string.
func formatLocation(line int) string {
	return "line " + string(rune('0'+line%10))
}

// getRemediation returns fix guidance for a vulnerability.
func getRemediation(vulnID string) string {
	remediations := map[string]string{
		"REENTRANCY-ETH":       "Use ReentrancyGuard from OpenZeppelin or follow checks-effects-interactions pattern",
		"REENTRANCY-CALLBACK":  "Add nonReentrant modifier to functions handling callbacks",
		"TX-ORIGIN":            "Replace tx.origin with msg.sender for authentication",
		"MISSING-OWNER-CHECK":  "Add appropriate access control modifier (onlyOwner, onlyRole, etc.)",
		"UNPROTECTED-SELFDESTRUCT": "Add onlyOwner modifier or remove selfdestruct entirely",
		"OVERFLOW-UNCHECKED":   "Review unchecked block for potential overflow/underflow issues",
		"DIVISION-BEFORE-MULTIPLY": "Reorder operations to multiply before dividing to preserve precision",
		"UNCHECKED-CALL":       "Check return value of external calls and handle failures",
		"ARBITRARY-DELEGATECALL": "Restrict delegatecall targets to trusted implementations only",
		"WEAK-RANDOMNESS":      "Use Chainlink VRF or commit-reveal scheme for secure randomness",
		"UNBOUNDED-LOOP":       "Add pagination or limit array size to prevent DoS",
		"PRICE-MANIPULATION":   "Use TWAP or multiple oracle sources for price data",
		"FLASHLOAN-PATTERN":    "Verify flash loan integration handles all edge cases",
		"UNPROTECTED-UPGRADE":  "Add access control to upgrade functions",
		"STORAGE-COLLISION":    "Use ERC1967 storage slots or OpenZeppelin's proxy implementation",
		"DEPRECATED-FUNCTION":  "Replace deprecated function with modern equivalent",
		"FLOATING-PRAGMA":      "Lock pragma to a specific compiler version",
	}

	if r, ok := remediations[vulnID]; ok {
		return r
	}
	return "Review and fix the identified vulnerability"
}
