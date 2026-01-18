// Package main_test provides tests for the VIGILUM CLI.
package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCLI_ValidAddress(t *testing.T) {
	tests := []struct {
		addr  string
		valid bool
	}{
		{"0x1234567890123456789012345678901234567890", true},
		{"0xdead", false},
		{"1234567890123456789012345678901234567890", false},
		{"0xGGGG567890123456789012345678901234567890", true}, // We only check length + prefix
		{"", false},
	}

	for _, tt := range tests {
		if isValidAddress(tt.addr) != tt.valid {
			t.Errorf("isValidAddress(%s) = %v, want %v", tt.addr, !tt.valid, tt.valid)
		}
	}
}

func TestCLI_FormatThreatLevel(t *testing.T) {
	tests := []struct {
		level    string
		expected string
	}{
		{"critical", "🔴 CRITICAL"},
		{"high", "🟠 HIGH"},
		{"medium", "🟡 MEDIUM"},
		{"low", "🟢 LOW"},
		{"info", "⚪ INFO"},
		{"none", "⚪ INFO"},
		{"unknown", "unknown"},
	}

	for _, tt := range tests {
		result := formatThreatLevel(tt.level)
		if result != tt.expected {
			t.Errorf("formatThreatLevel(%s) = %s, want %s", tt.level, result, tt.expected)
		}
	}
}

func TestCLI_Health(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success":   true,
				"timestamp": "2024-01-01T00:00:00Z",
				"data": map[string]interface{}{
					"status":  "healthy",
					"version": "0.1.0",
				},
			})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	config := Config{
		APIEndpoint: server.URL,
		Timeout:     defaultTimeout,
	}

	cli := NewCLI(config)
	var stdout bytes.Buffer
	cli.stdout = &stdout

	err := cli.runHealth()
	if err != nil {
		t.Fatalf("runHealth failed: %v", err)
	}

	output := stdout.String()
	if output == "" {
		t.Error("expected non-empty output")
	}
}

func TestCLI_HealthJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success":   true,
				"timestamp": "2024-01-01T00:00:00Z",
				"data": map[string]interface{}{
					"status":  "healthy",
					"version": "0.1.0",
				},
			})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	config := Config{
		APIEndpoint: server.URL,
		Timeout:     defaultTimeout,
		OutputJSON:  true,
	}

	cli := NewCLI(config)
	var stdout bytes.Buffer
	cli.stdout = &stdout

	err := cli.runHealth()
	if err != nil {
		t.Fatalf("runHealth failed: %v", err)
	}

	// Should be valid JSON
	var result map[string]interface{}
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Errorf("output is not valid JSON: %v", err)
	}
}

func TestCLI_RunVersion(t *testing.T) {
	err := run([]string{"-version"})
	if err != nil {
		t.Errorf("version command failed: %v", err)
	}
}

func TestCLI_RunHelp(t *testing.T) {
	err := run([]string{"-help"})
	if err != nil {
		t.Errorf("help command failed: %v", err)
	}
}

func TestCLI_NoCommand(t *testing.T) {
	// Should show help, not error
	err := run([]string{})
	if err != nil {
		t.Errorf("no command should show help, not error: %v", err)
	}
}

func TestCLI_UnknownCommand(t *testing.T) {
	err := run([]string{"unknowncommand"})
	if err == nil {
		t.Error("expected error for unknown command")
	}
}

func TestCLI_ScanRequiresAddress(t *testing.T) {
	err := run([]string{"scan"})
	if err == nil {
		t.Error("expected error when address is missing")
	}
}

func TestCLI_ScanInvalidAddress(t *testing.T) {
	err := run([]string{"scan", "invalid"})
	if err == nil {
		t.Error("expected error for invalid address")
	}
}

func TestCLI_InfoRequiresAddress(t *testing.T) {
	err := run([]string{"info"})
	if err == nil {
		t.Error("expected error when address is missing")
	}
}

func TestCLI_RiskRequiresAddress(t *testing.T) {
	err := run([]string{"risk"})
	if err == nil {
		t.Error("expected error when address is missing")
	}
}

func TestCLI_ScanWithMockServer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" && r.URL.Path == "/api/v1/scan" {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success":   true,
				"timestamp": "2024-01-01T00:00:00Z",
				"data": map[string]interface{}{
					"id":              "scan-123",
					"contractAddress": "0x1234567890123456789012345678901234567890",
					"chainId":         1,
					"status":          "completed",
					"riskScore":       25,
					"threatLevel":     "low",
					"vulnerabilities": []map[string]interface{}{},
					"metrics": map[string]interface{}{
						"totalIssues":   0,
						"criticalCount": 0,
						"highCount":     0,
						"mediumCount":   0,
						"lowCount":      0,
						"infoCount":     0,
					},
					"startedAt":   "2024-01-01T00:00:00Z",
					"completedAt": "2024-01-01T00:00:01Z",
				},
			})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	config := Config{
		APIEndpoint: server.URL,
		Timeout:     defaultTimeout,
	}

	cli := NewCLI(config)
	var stdout bytes.Buffer
	cli.stdout = &stdout

	err := cli.runScan([]string{"0x1234567890123456789012345678901234567890"})
	if err != nil {
		t.Fatalf("runScan failed: %v", err)
	}

	output := stdout.String()
	if output == "" {
		t.Error("expected non-empty output")
	}
}

func TestCLI_OutputScanResult(t *testing.T) {
	config := Config{
		OutputJSON: false,
	}

	cli := NewCLI(config)
	var stdout bytes.Buffer
	cli.stdout = &stdout

	result := ScanResult{
		Success: true,
		Data: struct {
			ID              string          `json:"id"`
			ContractAddress string          `json:"contractAddress"`
			ChainID         int             `json:"chainId"`
			Status          string          `json:"status"`
			RiskScore       int             `json:"riskScore"`
			ThreatLevel     string          `json:"threatLevel"`
			Vulnerabilities []Vulnerability `json:"vulnerabilities"`
			Metrics         ScanMetrics     `json:"metrics"`
			StartedAt       string          `json:"startedAt"`
			CompletedAt     string          `json:"completedAt,omitempty"`
		}{
			ID:              "test-scan",
			ContractAddress: "0x1234567890123456789012345678901234567890",
			ChainID:         1,
			Status:          "completed",
			RiskScore:       75,
			ThreatLevel:     "high",
			Vulnerabilities: []Vulnerability{
				{
					Type:        "reentrancy",
					Severity:    "high",
					Description: "Potential reentrancy vulnerability",
					Location:    "withdraw()",
				},
			},
			Metrics: ScanMetrics{
				TotalIssues:   1,
				CriticalCount: 0,
				HighCount:     1,
				MediumCount:   0,
				LowCount:      0,
				InfoCount:     0,
			},
			StartedAt:   "2024-01-01T00:00:00Z",
			CompletedAt: "2024-01-01T00:00:05Z",
		},
	}

	err := cli.outputScanResult(result)
	if err != nil {
		t.Fatalf("outputScanResult failed: %v", err)
	}

	output := stdout.String()
	
	// Check key elements are present
	if !bytes.Contains([]byte(output), []byte("VIGILUM SCAN REPORT")) {
		t.Error("expected report header in output")
	}
	if !bytes.Contains([]byte(output), []byte("0x1234567890123456789012345678901234567890")) {
		t.Error("expected address in output")
	}
	if !bytes.Contains([]byte(output), []byte("75/100")) {
		t.Error("expected risk score in output")
	}
	if !bytes.Contains([]byte(output), []byte("reentrancy")) {
		t.Error("expected vulnerability type in output")
	}
}

func TestCLI_GetEnvOrDefault(t *testing.T) {
	// Test with no env var set
	result := getEnvOrDefault("NONEXISTENT_VAR_12345", "default")
	if result != "default" {
		t.Errorf("expected 'default', got '%s'", result)
	}
}
