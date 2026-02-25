package scanner

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/vigilum/backend/internal/domain"
	"github.com/vigilum/backend/internal/ml"
)

// MLScanner performs vulnerability detection using machine learning
type MLScanner struct {
	logger           *slog.Logger
	inferenceClient  *ml.InferenceService
	supportedChains  []domain.ChainID
	featureExtractor FeatureExtractor
}

// FeatureExtractor extracts ML features from contract bytecode
type FeatureExtractor interface {
	ExtractFeatures(bytecode []byte, sourceCode string) ([]float64, error)
}

// NewMLScanner creates a new ML-based scanner
func NewMLScanner(logger *slog.Logger, inferenceClient *ml.InferenceService, extractor FeatureExtractor) *MLScanner {
	return &MLScanner{
		logger:          logger.With("scanner", "ml"),
		inferenceClient: inferenceClient,
		supportedChains: []domain.ChainID{
			1,      // Ethereum mainnet
			137,    // Polygon
			43114,  // Avalanche
			56,     // BSC
		},
		featureExtractor: extractor,
	}
}

// Name returns the scanner identifier
func (s *MLScanner) Name() string {
	return "ml"
}

// ScanType returns the type of analysis
func (s *MLScanner) ScanType() domain.ScanType {
	return domain.ScanTypeML
}

// Scan performs ML-based vulnerability detection
func (s *MLScanner) Scan(ctx context.Context, contract *domain.Contract) (*ScanResult, error) {
	if s.inferenceClient == nil {
		return &ScanResult{
			Vulnerabilities: []domain.Vulnerability{},
			RiskScore:       0.0,
			ThreatLevel:     domain.ThreatLevelNone,
			Metrics:         domain.ScanMetrics{},
		}, nil
	}

	s.logger.Info("Starting ML scan",
		"contract_id", contract.ID,
		"address", contract.Address,
	)

	// Extract features from bytecode or source code
	features, err := s.featureExtractor.ExtractFeatures(contract.Bytecode, contract.SourceCode)
	if err != nil {
		s.logger.Warn("Feature extraction failed",
			"contract_id", contract.ID,
			"error", err,
		)
		return &ScanResult{
			Vulnerabilities: []domain.Vulnerability{},
			RiskScore:       0.0,
			ThreatLevel:     domain.ThreatLevelNone,
			Metrics:         domain.ScanMetrics{},
		}, nil
	}

	// Run inference
	req := &ml.PredictionRequest{
		Features:   features,
		ContractID: string(contract.ID),
		Priority:   "normal",
	}

	resp, err := s.inferenceClient.Predict(ctx, req)
	if err != nil {
		s.logger.Warn("Inference failed",
			"contract_id", contract.ID,
			"error", err,
		)
		return &ScanResult{
			Vulnerabilities: []domain.Vulnerability{},
			RiskScore:       0.0,
			ThreatLevel:     domain.ThreatLevelNone,
			Metrics:         domain.ScanMetrics{},
		}, nil
	}

	// Convert ML output to vulnerabilities
	vulnerabilities := s.interpretPrediction(contract, resp, features)

	// Calculate threat level
	threatLevel := s.determineThreatLevel(resp.RiskScore)

	result := &ScanResult{
		Vulnerabilities: vulnerabilities,
		RiskScore:       resp.RiskScore / 100.0 * 10.0, // Normalize to 0-10
		ThreatLevel:     threatLevel,
		Metrics: domain.ScanMetrics{
			TotalIssues:  len(vulnerabilities),
			HighCount:    len(vulnerabilities), // Conservative: treat ML findings as high confidence
		},
		Metadata: map[string]any{
			"model_version":    resp.ModelVersion,
			"confidence_score": resp.ConfidenceScore,
			"inference_time":   resp.InferenceTime,
			"feature_importance": resp.Features,
		},
	}

	s.logger.Info("ML scan completed",
		"contract_id", contract.ID,
		"vulnerabilities", len(vulnerabilities),
		"risk_score", result.RiskScore,
		"threat_level", threatLevel,
	)

	return result, nil
}

// SupportedChains returns supported blockchain networks
func (s *MLScanner) SupportedChains() []domain.ChainID {
	return s.supportedChains
}

// IsHealthy checks if the ML service is operational
func (s *MLScanner) IsHealthy(ctx context.Context) bool {
	if s.inferenceClient == nil {
		return false
	}
	// Check if model can perform inference
	testFeatures := make([]float64, 50)
	for i := range testFeatures {
		testFeatures[i] = 0.5
	}
	req := &ml.PredictionRequest{
		Features:   testFeatures,
		ContractID: "health_check",
	}
	_, err := s.inferenceClient.Predict(ctx, req)
	return err == nil
}

// interpretPrediction converts ML prediction to vulnerability findings
func (s *MLScanner) interpretPrediction(contract *domain.Contract, resp *ml.PredictionResponse, features []float64) []domain.Vulnerability {
	vulnerabilities := make([]domain.Vulnerability, 0)

	// High confidence (>75%) and high risk score trigger vulnerability reports
	if resp.RiskScore >= 75 && resp.ConfidenceScore >= 0.8 {
		// Primary vulnerability from ML detection
		vuln := domain.Vulnerability{
			Type:        domain.VulnLogicError,
			Severity:    domain.ThreatLevelCritical,
			Title:       "ML-Detected Vulnerability Pattern",
			Description: fmt.Sprintf("Machine learning model detected critical vulnerability pattern (risk: %.1f%%, confidence: %.1f%%)", resp.RiskScore, resp.ConfidenceScore*100),
			Confidence:  resp.ConfidenceScore,
			DetectedBy:  "ml",
			Location: domain.CodeLocation{
				File:      "bytecode",
				StartLine: 1,
			},
		}
		vulnerabilities = append(vulnerabilities, vuln)
	}

	// Medium-high risk (50-75%) with good confidence suggests potential issues
	if resp.RiskScore >= 50 && resp.RiskScore < 75 && resp.ConfidenceScore >= 0.75 {
		vuln := domain.Vulnerability{
			Type:        domain.VulnLogicError,
			Severity:    domain.ThreatLevelHigh,
			Title:       "ML-Detected Potential Vulnerability",
			Description: fmt.Sprintf("Model detected potential vulnerability pattern (risk: %.1f%%, confidence: %.1f%%)", resp.RiskScore, resp.ConfidenceScore*100),
			Confidence:  resp.ConfidenceScore,
			DetectedBy:  "ml",
			Location: domain.CodeLocation{
				File:      "bytecode",
				StartLine: 1,
			},
		}
		vulnerabilities = append(vulnerabilities, vuln)
	}

	// Medium risk (25-50%) as informational
	if resp.RiskScore >= 25 && resp.RiskScore < 50 && resp.ConfidenceScore >= 0.7 {
		vuln := domain.Vulnerability{
			Type:        domain.VulnLogicError,
			Severity:    domain.ThreatLevelMedium,
			Title:       "ML-Flagged Contract Pattern",
			Description: fmt.Sprintf("Model flagged unusual pattern (risk: %.1f%%, confidence: %.1f%%)", resp.RiskScore, resp.ConfidenceScore*100),
			Confidence:  resp.ConfidenceScore,
			DetectedBy:  "ml",
			Location: domain.CodeLocation{
				File:      "bytecode",
				StartLine: 1,
			},
		}
		vulnerabilities = append(vulnerabilities, vuln)
	}

	return vulnerabilities
}

// determineThreatLevel maps risk score to threat level
func (s *MLScanner) determineThreatLevel(riskScore float64) domain.ThreatLevel {
	switch {
	case riskScore >= 75:
		return domain.ThreatLevelCritical
	case riskScore >= 50:
		return domain.ThreatLevelHigh
	case riskScore >= 25:
		return domain.ThreatLevelMedium
	case riskScore >= 10:
		return domain.ThreatLevelLow
	default:
		return domain.ThreatLevelInfo
	}
}
