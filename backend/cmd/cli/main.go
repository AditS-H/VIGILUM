// Package main provides the VIGILUM CLI for contract security scanning and analysis.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	version     = "0.1.0"
	defaultAPI  = "http://localhost:8080"
	defaultTimeout = 30 * time.Second
)

// Config holds CLI configuration.
type Config struct {
	APIEndpoint string
	APIKey      string
	Timeout     time.Duration
	OutputJSON  bool
	Verbose     bool
}

// CLI is the main command-line interface.
type CLI struct {
	config Config
	client *http.Client
	stdout io.Writer
	stderr io.Writer
}

// NewCLI creates a new CLI instance.
func NewCLI(config Config) *CLI {
	return &CLI{
		config: config,
		client: &http.Client{
			Timeout: config.Timeout,
		},
		stdout: os.Stdout,
		stderr: os.Stderr,
	}
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	// Parse global flags
	flags := flag.NewFlagSet("vigilum", flag.ContinueOnError)
	
	var (
		apiEndpoint = flags.String("api", getEnvOrDefault("VIGILUM_API", defaultAPI), "API endpoint")
		apiKey      = flags.String("key", os.Getenv("VIGILUM_API_KEY"), "API key")
		timeout     = flags.Duration("timeout", defaultTimeout, "Request timeout")
		jsonOutput  = flags.Bool("json", false, "Output JSON format")
		verbose     = flags.Bool("verbose", false, "Verbose output")
		showVersion = flags.Bool("version", false, "Show version")
		showHelp    = flags.Bool("help", false, "Show help")
	)

	// Parse flags up to the subcommand
	if err := flags.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return nil
		}
		return err
	}

	if *showVersion {
		fmt.Printf("vigilum version %s\n", version)
		return nil
	}

	if *showHelp || flags.NArg() == 0 {
		printUsage()
		return nil
	}

	config := Config{
		APIEndpoint: *apiEndpoint,
		APIKey:      *apiKey,
		Timeout:     *timeout,
		OutputJSON:  *jsonOutput,
		Verbose:     *verbose,
	}

	cli := NewCLI(config)
	
	// Execute subcommand
	subCmd := flags.Arg(0)
	subArgs := flags.Args()[1:]

	switch subCmd {
	case "scan":
		return cli.runScan(subArgs)
	case "info":
		return cli.runInfo(subArgs)
	case "risk":
		return cli.runRisk(subArgs)
	case "alerts":
		return cli.runAlerts(subArgs)
	case "health":
		return cli.runHealth()
	case "version":
		fmt.Printf("vigilum version %s\n", version)
		return nil
	case "help":
		printUsage()
		return nil
	default:
		return fmt.Errorf("unknown command: %s", subCmd)
	}
}

func printUsage() {
	fmt.Print(`
VIGILUM CLI - Decentralized Blockchain Security Scanner

USAGE:
    vigilum [OPTIONS] <COMMAND> [ARGS]

OPTIONS:
    -api <url>      API endpoint (default: http://localhost:8080, env: VIGILUM_API)
    -key <key>      API key for authentication (env: VIGILUM_API_KEY)
    -timeout <dur>  Request timeout (default: 30s)
    -json           Output in JSON format
    -verbose        Enable verbose output
    -version        Show version information
    -help           Show this help message

COMMANDS:
    scan <address>      Scan a contract for vulnerabilities
    info <address>      Get security information about a contract
    risk <address>      Get risk score for an address
    alerts [address]    List security alerts
    health              Check API health status
    version             Show version information
    help                Show this help message

EXAMPLES:
    # Scan a contract on Ethereum mainnet
    vigilum scan 0x1234... -chain 1

    # Get risk score with JSON output
    vigilum -json risk 0x1234...

    # Check API health
    vigilum health

    # Scan with custom API endpoint
    vigilum -api https://api.vigilum.network scan 0x1234...

ENVIRONMENT:
    VIGILUM_API         API endpoint URL
    VIGILUM_API_KEY     API key for authentication

`)
}

// ═══════════════════════════════════════════════════════════════════════════
// SCAN COMMAND
// ═══════════════════════════════════════════════════════════════════════════

func (c *CLI) runScan(args []string) error {
	flags := flag.NewFlagSet("scan", flag.ContinueOnError)
	chainID := flags.Int("chain", 1, "Chain ID (1=Ethereum, 137=Polygon, etc.)")
	deep := flags.Bool("deep", false, "Enable deep analysis")
	wait := flags.Bool("wait", true, "Wait for scan completion")

	if err := flags.Parse(args); err != nil {
		return err
	}

	if flags.NArg() < 1 {
		return fmt.Errorf("address required: vigilum scan <address>")
	}

	address := flags.Arg(0)
	if !isValidAddress(address) {
		return fmt.Errorf("invalid address: %s", address)
	}

	if c.config.Verbose {
		fmt.Fprintf(c.stderr, "Scanning contract %s on chain %d...\n", address, *chainID)
	}

	// Submit scan request
	reqBody := map[string]interface{}{
		"address": address,
		"chainId": *chainID,
		"options": map[string]interface{}{
			"deep": *deep,
		},
	}

	resp, err := c.post("/api/v1/scan", reqBody)
	if err != nil {
		return fmt.Errorf("failed to submit scan: %w", err)
	}

	var scanResult ScanResult
	if err := json.Unmarshal(resp, &scanResult); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if !scanResult.Success {
		return fmt.Errorf("scan failed: %s", scanResult.Error.Message)
	}

	// If not waiting, just return the scan ID
	if !*wait {
		if c.config.OutputJSON {
			return c.outputJSON(scanResult)
		}
		fmt.Fprintf(c.stdout, "Scan submitted: %s\n", scanResult.Data.ID)
		fmt.Fprintf(c.stdout, "Status: %s\n", scanResult.Data.Status)
		return nil
	}

	// Poll for completion
	scanID := scanResult.Data.ID
	for scanResult.Data.Status == "pending" || scanResult.Data.Status == "running" {
		if c.config.Verbose {
			fmt.Fprintf(c.stderr, "Status: %s...\n", scanResult.Data.Status)
		}
		time.Sleep(2 * time.Second)

		resp, err = c.get(fmt.Sprintf("/api/v1/scan/%s", scanID))
		if err != nil {
			return fmt.Errorf("failed to get scan status: %w", err)
		}

		if err := json.Unmarshal(resp, &scanResult); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}
	}

	return c.outputScanResult(scanResult)
}

func (c *CLI) outputScanResult(result ScanResult) error {
	if c.config.OutputJSON {
		return c.outputJSON(result)
	}

	data := result.Data
	fmt.Fprintf(c.stdout, "\n╔══════════════════════════════════════════════════════════════╗\n")
	fmt.Fprintf(c.stdout, "║                    VIGILUM SCAN REPORT                       ║\n")
	fmt.Fprintf(c.stdout, "╚══════════════════════════════════════════════════════════════╝\n\n")

	fmt.Fprintf(c.stdout, "Contract:     %s\n", data.ContractAddress)
	fmt.Fprintf(c.stdout, "Chain ID:     %d\n", data.ChainID)
	fmt.Fprintf(c.stdout, "Status:       %s\n", data.Status)
	fmt.Fprintf(c.stdout, "Risk Score:   %d/100\n", data.RiskScore)
	fmt.Fprintf(c.stdout, "Threat Level: %s\n", formatThreatLevel(data.ThreatLevel))
	fmt.Fprintf(c.stdout, "\n")

	// Metrics
	m := data.Metrics
	fmt.Fprintf(c.stdout, "─── FINDINGS ─────────────────────────────────────────────────\n")
	fmt.Fprintf(c.stdout, "Critical: %d  │  High: %d  │  Medium: %d  │  Low: %d  │  Info: %d\n",
		m.CriticalCount, m.HighCount, m.MediumCount, m.LowCount, m.InfoCount)
	fmt.Fprintf(c.stdout, "Total Issues: %d\n\n", m.TotalIssues)

	// Vulnerabilities
	if len(data.Vulnerabilities) > 0 {
		fmt.Fprintf(c.stdout, "─── VULNERABILITIES ──────────────────────────────────────────\n")
		for i, vuln := range data.Vulnerabilities {
			fmt.Fprintf(c.stdout, "\n[%d] %s (%s)\n", i+1, vuln.Type, formatSeverity(vuln.Severity))
			fmt.Fprintf(c.stdout, "    %s\n", vuln.Description)
			if vuln.Location != "" {
				fmt.Fprintf(c.stdout, "    Location: %s\n", vuln.Location)
			}
		}
		fmt.Fprintf(c.stdout, "\n")
	}

	fmt.Fprintf(c.stdout, "──────────────────────────────────────────────────────────────\n")
	fmt.Fprintf(c.stdout, "Scan completed at: %s\n", data.CompletedAt)

	return nil
}

// ═══════════════════════════════════════════════════════════════════════════
// INFO COMMAND
// ═══════════════════════════════════════════════════════════════════════════

func (c *CLI) runInfo(args []string) error {
	flags := flag.NewFlagSet("info", flag.ContinueOnError)
	chainID := flags.Int("chain", 1, "Chain ID")

	if err := flags.Parse(args); err != nil {
		return err
	}

	if flags.NArg() < 1 {
		return fmt.Errorf("address required: vigilum info <address>")
	}

	address := flags.Arg(0)
	if !isValidAddress(address) {
		return fmt.Errorf("invalid address: %s", address)
	}

	resp, err := c.get(fmt.Sprintf("/api/v1/contracts/%d/%s", *chainID, address))
	if err != nil {
		return fmt.Errorf("failed to get contract info: %w", err)
	}

	var result ContractInfoResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if !result.Success {
		return fmt.Errorf("request failed: %s", result.Error.Message)
	}

	if c.config.OutputJSON {
		return c.outputJSON(result)
	}

	data := result.Data
	fmt.Fprintf(c.stdout, "\n─── CONTRACT INFO ────────────────────────────────────────────\n")
	fmt.Fprintf(c.stdout, "Address:       %s\n", data.Address)
	fmt.Fprintf(c.stdout, "Chain ID:      %d\n", data.ChainID)
	fmt.Fprintf(c.stdout, "Bytecode Hash: %s\n", data.BytecodeHash)
	fmt.Fprintf(c.stdout, "Verified:      %t\n", data.IsVerified)
	fmt.Fprintf(c.stdout, "Blacklisted:   %t\n", data.IsBlacklisted)
	fmt.Fprintf(c.stdout, "Risk Score:    %d/100\n", data.RiskScore)
	fmt.Fprintf(c.stdout, "Threat Level:  %s\n", formatThreatLevel(data.ThreatLevel))
	if len(data.Labels) > 0 {
		fmt.Fprintf(c.stdout, "Labels:        %s\n", strings.Join(data.Labels, ", "))
	}
	fmt.Fprintf(c.stdout, "──────────────────────────────────────────────────────────────\n")

	return nil
}

// ═══════════════════════════════════════════════════════════════════════════
// RISK COMMAND
// ═══════════════════════════════════════════════════════════════════════════

func (c *CLI) runRisk(args []string) error {
	flags := flag.NewFlagSet("risk", flag.ContinueOnError)
	chainID := flags.Int("chain", 1, "Chain ID")

	if err := flags.Parse(args); err != nil {
		return err
	}

	if flags.NArg() < 1 {
		return fmt.Errorf("address required: vigilum risk <address>")
	}

	address := flags.Arg(0)
	if !isValidAddress(address) {
		return fmt.Errorf("invalid address: %s", address)
	}

	resp, err := c.get(fmt.Sprintf("/api/v1/firewall/risk/%d/%s", *chainID, address))
	if err != nil {
		return fmt.Errorf("failed to get risk score: %w", err)
	}

	var result RiskResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if !result.Success {
		return fmt.Errorf("request failed: %s", result.Error.Message)
	}

	if c.config.OutputJSON {
		return c.outputJSON(result)
	}

	data := result.Data
	fmt.Fprintf(c.stdout, "\n─── RISK ASSESSMENT ──────────────────────────────────────────\n")
	fmt.Fprintf(c.stdout, "Address:      %s\n", data.Address)
	fmt.Fprintf(c.stdout, "Risk Score:   %d/100\n", data.RiskScore)
	fmt.Fprintf(c.stdout, "Risk Level:   %s\n", formatRiskLevel(data.RiskLevel))

	if len(data.Signals) > 0 {
		fmt.Fprintf(c.stdout, "\n  Signals:\n")
		for _, sig := range data.Signals {
			fmt.Fprintf(c.stdout, "    • %s (confidence: %.0f%%, source: %s)\n",
				sig.Type, sig.Confidence*100, sig.Source)
		}
	}

	fmt.Fprintf(c.stdout, "──────────────────────────────────────────────────────────────\n")

	return nil
}

// ═══════════════════════════════════════════════════════════════════════════
// ALERTS COMMAND
// ═══════════════════════════════════════════════════════════════════════════

func (c *CLI) runAlerts(args []string) error {
	flags := flag.NewFlagSet("alerts", flag.ContinueOnError)
	chainID := flags.Int("chain", 1, "Chain ID")
	limit := flags.Int("limit", 20, "Max alerts to show")

	if err := flags.Parse(args); err != nil {
		return err
	}

	path := fmt.Sprintf("/api/v1/alerts?chainId=%d&limit=%d", *chainID, *limit)
	if flags.NArg() > 0 {
		address := flags.Arg(0)
		if !isValidAddress(address) {
			return fmt.Errorf("invalid address: %s", address)
		}
		path += "&address=" + address
	}

	resp, err := c.get(path)
	if err != nil {
		return fmt.Errorf("failed to get alerts: %w", err)
	}

	var result AlertsResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if !result.Success {
		return fmt.Errorf("request failed: %s", result.Error.Message)
	}

	if c.config.OutputJSON {
		return c.outputJSON(result)
	}

	fmt.Fprintf(c.stdout, "\n─── SECURITY ALERTS ──────────────────────────────────────────\n")
	if len(result.Data) == 0 {
		fmt.Fprintf(c.stdout, "No alerts found.\n")
	} else {
		for i, alert := range result.Data {
			fmt.Fprintf(c.stdout, "\n[%d] %s\n", i+1, formatSeverity(alert.Severity))
			fmt.Fprintf(c.stdout, "    Address: %s\n", alert.Address)
			fmt.Fprintf(c.stdout, "    Type:    %s\n", alert.Type)
			fmt.Fprintf(c.stdout, "    Message: %s\n", alert.Message)
			fmt.Fprintf(c.stdout, "    Time:    %s\n", alert.CreatedAt)
		}
	}
	fmt.Fprintf(c.stdout, "──────────────────────────────────────────────────────────────\n")

	return nil
}

// ═══════════════════════════════════════════════════════════════════════════
// HEALTH COMMAND
// ═══════════════════════════════════════════════════════════════════════════

func (c *CLI) runHealth() error {
	resp, err := c.get("/health")
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}

	var result HealthResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if c.config.OutputJSON {
		return c.outputJSON(result)
	}

	fmt.Fprintf(c.stdout, "API Status: %s\n", result.Data.Status)
	fmt.Fprintf(c.stdout, "Version:    %s\n", result.Data.Version)
	
	return nil
}

// ═══════════════════════════════════════════════════════════════════════════
// HTTP CLIENT METHODS
// ═══════════════════════════════════════════════════════════════════════════

func (c *CLI) get(path string) ([]byte, error) {
	url := c.config.APIEndpoint + path
	
	req, err := http.NewRequestWithContext(context.Background(), "GET", url, nil)
	if err != nil {
		return nil, err
	}

	c.setHeaders(req)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func (c *CLI) post(path string, body interface{}) ([]byte, error) {
	url := c.config.APIEndpoint + path

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(context.Background(), "POST", url, strings.NewReader(string(jsonBody)))
	if err != nil {
		return nil, err
	}

	c.setHeaders(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func (c *CLI) setHeaders(req *http.Request) {
	req.Header.Set("User-Agent", "vigilum-cli/"+version)
	if c.config.APIKey != "" {
		req.Header.Set("X-API-Key", c.config.APIKey)
	}
}

func (c *CLI) outputJSON(v interface{}) error {
	enc := json.NewEncoder(c.stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// ═══════════════════════════════════════════════════════════════════════════
// RESPONSE TYPES
// ═══════════════════════════════════════════════════════════════════════════

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ScanResult struct {
	Success   bool     `json:"success"`
	Error     APIError `json:"error,omitempty"`
	Timestamp string   `json:"timestamp"`
	Data      struct {
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
	} `json:"data"`
}

type Vulnerability struct {
	Type        string `json:"type"`
	Severity    string `json:"severity"`
	Description string `json:"description"`
	Location    string `json:"location,omitempty"`
}

type ScanMetrics struct {
	TotalIssues   int `json:"totalIssues"`
	CriticalCount int `json:"criticalCount"`
	HighCount     int `json:"highCount"`
	MediumCount   int `json:"mediumCount"`
	LowCount      int `json:"lowCount"`
	InfoCount     int `json:"infoCount"`
}

type ContractInfoResponse struct {
	Success   bool     `json:"success"`
	Error     APIError `json:"error,omitempty"`
	Timestamp string   `json:"timestamp"`
	Data      struct {
		Address      string   `json:"address"`
		ChainID      int      `json:"chainId"`
		BytecodeHash string   `json:"bytecodeHash"`
		IsVerified   bool     `json:"isVerified"`
		IsBlacklisted bool    `json:"isBlacklisted"`
		RiskScore    int      `json:"riskScore"`
		ThreatLevel  string   `json:"threatLevel"`
		Labels       []string `json:"labels"`
	} `json:"data"`
}

type RiskResponse struct {
	Success   bool     `json:"success"`
	Error     APIError `json:"error,omitempty"`
	Timestamp string   `json:"timestamp"`
	Data      struct {
		Address   string `json:"address"`
		RiskScore int    `json:"riskScore"`
		RiskLevel string `json:"riskLevel"`
		Signals   []struct {
			Type       string  `json:"type"`
			Confidence float64 `json:"confidence"`
			Source     string  `json:"source"`
			Timestamp  string  `json:"timestamp"`
		} `json:"signals"`
	} `json:"data"`
}

type AlertsResponse struct {
	Success   bool     `json:"success"`
	Error     APIError `json:"error,omitempty"`
	Timestamp string   `json:"timestamp"`
	Data      []Alert  `json:"data"`
}

type Alert struct {
	ID        string `json:"id"`
	Address   string `json:"address"`
	ChainID   int    `json:"chainId"`
	Type      string `json:"type"`
	Severity  string `json:"severity"`
	Message   string `json:"message"`
	CreatedAt string `json:"createdAt"`
}

type HealthResponse struct {
	Success   bool   `json:"success"`
	Timestamp string `json:"timestamp"`
	Data      struct {
		Status  string `json:"status"`
		Version string `json:"version"`
	} `json:"data"`
}

// ═══════════════════════════════════════════════════════════════════════════
// UTILITY FUNCTIONS
// ═══════════════════════════════════════════════════════════════════════════

func isValidAddress(addr string) bool {
	if len(addr) != 42 {
		return false
	}
	if !strings.HasPrefix(addr, "0x") {
		return false
	}
	return true
}

func formatThreatLevel(level string) string {
	switch level {
	case "critical":
		return "🔴 CRITICAL"
	case "high":
		return "🟠 HIGH"
	case "medium":
		return "🟡 MEDIUM"
	case "low":
		return "🟢 LOW"
	case "info", "none":
		return "⚪ INFO"
	default:
		return level
	}
}

func formatSeverity(sev string) string {
	return formatThreatLevel(sev)
}

func formatRiskLevel(level string) string {
	return formatThreatLevel(level)
}

func getEnvOrDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
