package ml

import (
	"math"
	"regexp"
	"strings"
)

// DefaultFeatureExtractor extracts ML features from smart contract bytecode and source
type DefaultFeatureExtractor struct{}

// ExtractFeatures extracts 50 ML features from contract bytecode and source code
func (fe *DefaultFeatureExtractor) ExtractFeatures(bytecode []byte, sourceCode string) ([]float64, error) {
	features := make([]float64, 50)

	// Bytecode-based features (0-24)
	features[0] = float64(len(bytecode)) / 10000.0 // Normalized bytecode length
	features[1] = fe.calculateBytecodeCopyRate(bytecode)
	features[2] = fe.calculateOpcodeEntropy(bytecode)
	features[3] = fe.countOpcodeType(bytecode, 0x60) // PUSH opcodes
	features[4] = fe.countOpcodeType(bytecode, 0xF1) // CALL opcodes
	features[5] = fe.countOpcodeType(bytecode, 0x54) // SLOAD opcodes
	features[6] = fe.countOpcodeType(bytecode, 0x55) // SSTORE opcodes
	features[7] = fe.countOpcodeType(bytecode, 0x57) // JUMPI opcodes
	features[8] = fe.countOpcodeType(bytecode, 0x56) // JUMP opcodes
	features[9] = fe.calculateControlFlowComplexity(bytecode)

	// Source code-based features (10-34)
	if sourceCode != "" {
		features[10] = float64(len(sourceCode)) / 10000.0 // Normalized source length
		features[11] = fe.countPattern(sourceCode, `call\s*\(`)
		features[12] = fe.countPattern(sourceCode, `delegatecall\s*\(`)
		features[13] = fe.countPattern(sourceCode, `selfdestruct\s*\(`)
		features[14] = fe.countPattern(sourceCode, `tx\.origin`)
		features[15] = fe.countPattern(sourceCode, `block\.timestamp`)
		features[16] = fe.countPattern(sourceCode, `now`)
		features[17] = fe.countPattern(sourceCode, `\.transfer\s*\(`)
		features[18] = fe.countPattern(sourceCode, `require\s*\(`)
		features[19] = fe.countPattern(sourceCode, `assert\s*\(`)
		features[20] = fe.countPattern(sourceCode, `revert\s*\(`)
		features[21] = fe.countPattern(sourceCode, `public\s+.*function`)
		features[22] = fe.countPattern(sourceCode, `payable`)
		features[23] = fe.countPattern(sourceCode, `modifier`)
		features[24] = fe.countLines(sourceCode)
	}

	// Combined/derived features (25-49)
	features[25] = fe.calculateComplexityRatio(features)
	features[26] = fe.calculateRiskIndicators(sourceCode, bytecode)
	features[27] = fe.calculateFunctionDiversity(sourceCode)
	features[28] = fe.calculateMathOperationFrequency(sourceCode)
	features[29] = fe.calculateLoopComplexity(sourceCode)
	features[30] = fe.calculateStateModificationFrequency(sourceCode)
	features[31] = fe.calculateExternalCallFrequency(sourceCode)
	features[32] = fe.calculateReturnValueHandling(sourceCode)
	features[33] = fe.calculateTimestampDependency(sourceCode)
	features[34] = fe.calculateAccessControlPatterns(sourceCode)

	// Normalized versions of top features (35-49)
	for i := 0; i < 15; i++ {
		if i+35 < len(features) {
			features[i+35] = features[i] / (1.0 + features[i])
		}
	}

	return features, nil
}

// calculateBytecodeCopyRate measures code duplication
func (fe *DefaultFeatureExtractor) calculateBytecodeCopyRate(bytecode []byte) float64 {
	if len(bytecode) < 32 {
		return 0.0
	}

	chunks := make(map[string]int)
	chunkSize := 32
	duplicates := 0.0

	for i := 0; i <= len(bytecode)-chunkSize; i++ {
		chunk := string(bytecode[i : i+chunkSize])
		if count, exists := chunks[chunk]; exists {
			duplicates += float64(count)
		}
		chunks[chunk]++
	}

	return duplicates / float64(len(bytecode)/chunkSize)
}

// calculateOpcodeEntropy measures bytecode randomness
func (fe *DefaultFeatureExtractor) calculateOpcodeEntropy(bytecode []byte) float64 {
	if len(bytecode) == 0 {
		return 0.0
	}

	freq := make(map[byte]int)
	for _, b := range bytecode {
		freq[b]++
	}

	entropy := 0.0
	length := float64(len(bytecode))

	for _, count := range freq {
		p := float64(count) / length
		if p > 0 {
			entropy -= p * math.Log2(p)
		}
	}

	return entropy / 8.0 // Normalize to 0-1
}

// countOpcodeType counts occurrences of specific opcode
func (fe *DefaultFeatureExtractor) countOpcodeType(bytecode []byte, opcode byte) float64 {
	count := 0.0
	for _, b := range bytecode {
		if b == opcode {
			count++
		}
	}
	return count / float64(len(bytecode)+1)
}

// calculateControlFlowComplexity estimates branching complexity
func (fe *DefaultFeatureExtractor) calculateControlFlowComplexity(bytecode []byte) float64 {
	// Count JUMPI (branching) instructions as proxy for control flow complexity
	jumpiCount := 0
	for _, b := range bytecode {
		if b == 0x57 { // JUMPI
			jumpiCount++
		}
	}
	return float64(jumpiCount) / float64(len(bytecode)+1)
}

// countPattern counts regex pattern matches
func (fe *DefaultFeatureExtractor) countPattern(source string, pattern string) float64 {
	re := regexp.MustCompile(pattern)
	matches := re.FindAllString(source, -1)
	return float64(len(matches))
}

// countLines counts number of code lines
func (fe *DefaultFeatureExtractor) countLines(source string) float64 {
	return float64(len(strings.Split(source, "\n")))
}

// calculateComplexityRatio combines complexity indicators
func (fe *DefaultFeatureExtractor) calculateComplexityRatio(features []float64) float64 {
	if len(features) < 10 {
		return 0.0
	}
	sum := features[0] + features[1] + features[2] + features[3] + features[4]
	return sum / 5.0
}

// calculateRiskIndicators scores known vulnerability patterns
func (fe *DefaultFeatureExtractor) calculateRiskIndicators(source string, bytecode []byte) float64 {
	riskScore := 0.0

	// High-risk patterns
	if strings.Contains(source, "delegatecall") {
		riskScore += 0.2
	}
	if strings.Contains(source, "tx.origin") {
		riskScore += 0.15
	}
	if strings.Contains(source, "block.timestamp") && !strings.Contains(source, "require") {
		riskScore += 0.15
	}
	if strings.Contains(source, ".transfer(") && !strings.Contains(source, "require") {
		riskScore += 0.1
	}
	if strings.Contains(source, "for (") || strings.Contains(source, "while (") {
		riskScore += 0.1
	}

	return riskScore
}

// calculateFunctionDiversity estimates function count and diversity
func (fe *DefaultFeatureExtractor) calculateFunctionDiversity(source string) float64 {
	functionCount := float64(strings.Count(source, "function "))
	// Normalize: 10+ functions = 1.0
	return functionCount / 10.0
}

// calculateMathOperationFrequency counts arithmetic operations
func (fe *DefaultFeatureExtractor) calculateMathOperationFrequency(source string) float64 {
	ops := strings.Count(source, "+") +
		strings.Count(source, "-") +
		strings.Count(source, "*") +
		strings.Count(source, "/")
	return float64(ops) / float64(len(source)/100+1)
}

// calculateLoopComplexity estimates loop nesting
func (fe *DefaultFeatureExtractor) calculateLoopComplexity(source string) float64 {
	loopCount := float64(strings.Count(source, "for (") + strings.Count(source, "while ("))
	return loopCount / float64(len(source)/100+1)
}

// calculateStateModificationFrequency counts state writes
func (fe *DefaultFeatureExtractor) calculateStateModificationFrequency(source string) float64 {
	stateWrites := strings.Count(source, "=") - strings.Count(source, "==")
	return float64(stateWrites) / float64(len(source)/100+1)
}

// calculateExternalCallFrequency counts external calls
func (fe *DefaultFeatureExtractor) calculateExternalCallFrequency(source string) float64 {
	calls := strings.Count(source, ".call") +
		strings.Count(source, ".delegatecall") +
		strings.Count(source, ".staticcall")
	return float64(calls) / float64(len(source)/100+1)
}

// calculateReturnValueHandling checks if calls check return values
func (fe *DefaultFeatureExtractor) calculateReturnValueHandling(source string) float64 {
	if len(source) == 0 {
		return 0.0
	}
	// Count `require()` statements that validate external calls
	requires := strings.Count(source, "require(")
	calls := strings.Count(source, ".call")
	if calls == 0 {
		return 1.0
	}
	return float64(requires) / float64(calls)
}

// calculateTimestampDependency checks timestamp usage
func (fe *DefaultFeatureExtractor) calculateTimestampDependency(source string) float64 {
	timestampUsage := strings.Count(source, "block.timestamp") + strings.Count(source, "now")
	return float64(timestampUsage) / float64(len(source)/100+1)
}

// calculateAccessControlPatterns detects access control mechanisms
func (fe *DefaultFeatureExtractor) calculateAccessControlPatterns(source string) float64 {
	patterns := strings.Count(source, "onlyOwner") +
		strings.Count(source, "require(msg.sender") +
		strings.Count(source, "modifier")
	return float64(patterns) / float64(len(source)/100+1)
}

// NewDefaultFeatureExtractor creates a feature extractor
func NewDefaultFeatureExtractor() *DefaultFeatureExtractor {
	return &DefaultFeatureExtractor{}
}
