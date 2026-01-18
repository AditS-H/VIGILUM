// Package oracle implements threat intelligence aggregation and on-chain publishing.
package oracle

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/vigilum/backend/internal/domain"
)

// ═══════════════════════════════════════════════════════════════════════════════
// FEED FETCHER INTERFACE
// ═══════════════════════════════════════════════════════════════════════════════

// FeedFetcher defines the interface for fetching threat intelligence.
type FeedFetcher interface {
	Source() FeedSource
	Fetch(ctx context.Context) ([]FeedEvent, error)
	SetLastFetchTime(t time.Time)
}

// ═══════════════════════════════════════════════════════════════════════════════
// GITHUB POC FEED
// ═══════════════════════════════════════════════════════════════════════════════

// GitHubFeed fetches exploit PoCs from GitHub.
type GitHubFeed struct {
	httpClient    *http.Client
	apiKey        string
	baseURL       string
	lastFetchTime time.Time
	config        FeedConfig
}

// NewGitHubFeed creates a GitHub feed fetcher.
func NewGitHubFeed(config FeedConfig) *GitHubFeed {
	return &GitHubFeed{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		apiKey:     config.APIKey,
		baseURL:    "https://api.github.com",
		config:     config,
	}
}

func (g *GitHubFeed) Source() FeedSource {
	return FeedSourceGitHub
}

func (g *GitHubFeed) SetLastFetchTime(t time.Time) {
	g.lastFetchTime = t
}

// Fetch searches GitHub for exploit-related repositories.
func (g *GitHubFeed) Fetch(ctx context.Context) ([]FeedEvent, error) {
	// Search queries for finding exploit repos
	queries := []string{
		"blockchain exploit in:name,description",
		"smart contract vulnerability in:name,description",
		"defi hack proof of concept",
		"solidity reentrancy exploit",
		"flash loan attack",
	}

	var events []FeedEvent
	for _, query := range queries {
		repos, err := g.searchRepos(ctx, query)
		if err != nil {
			continue // Log and continue with other queries
		}
		events = append(events, repos...)
	}

	return events, nil
}

// GitHubSearchResult represents a GitHub API search response.
type GitHubSearchResult struct {
	TotalCount int `json:"total_count"`
	Items      []struct {
		ID          int    `json:"id"`
		Name        string `json:"name"`
		FullName    string `json:"full_name"`
		Description string `json:"description"`
		HTMLURL     string `json:"html_url"`
		CreatedAt   string `json:"created_at"`
		UpdatedAt   string `json:"updated_at"`
		Topics      []string `json:"topics"`
		Language    string `json:"language"`
	} `json:"items"`
}

func (g *GitHubFeed) searchRepos(ctx context.Context, query string) ([]FeedEvent, error) {
	// Build URL with date filter if we have a last fetch time
	url := fmt.Sprintf("%s/search/repositories?q=%s&sort=updated&per_page=30",
		g.baseURL, strings.ReplaceAll(query, " ", "+"))
	
	if !g.lastFetchTime.IsZero() {
		url += fmt.Sprintf("+created:>%s", g.lastFetchTime.Format("2006-01-02"))
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	if g.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+g.apiKey)
	}

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API error: %d - %s", resp.StatusCode, body)
	}

	var result GitHubSearchResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var events []FeedEvent
	for _, item := range result.Items {
		event := g.repoToEvent(item)
		if event != nil {
			events = append(events, *event)
		}
	}

	return events, nil
}

func (g *GitHubFeed) repoToEvent(item struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	FullName    string `json:"full_name"`
	Description string `json:"description"`
	HTMLURL     string `json:"html_url"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
	Topics      []string `json:"topics"`
	Language    string `json:"language"`
}) *FeedEvent {
	// Analyze repo name and description for threat indicators
	combined := strings.ToLower(item.Name + " " + item.Description)
	
	// Check for exploit keywords
	exploitPatterns := []string{
		"exploit", "hack", "vulnerability", "poc", "proof-of-concept",
		"reentrancy", "flash-loan", "overflow", "underflow", "access-control",
	}
	
	isExploit := false
	for _, pattern := range exploitPatterns {
		if strings.Contains(combined, pattern) {
			isExploit = true
			break
		}
	}
	
	if !isExploit {
		return nil // Not relevant
	}

	// Extract contract addresses from description
	targets := g.extractTargets(combined)
	
	// Determine severity based on keywords
	severity := domain.ThreatLevelLow
	if strings.Contains(combined, "critical") || strings.Contains(combined, "0day") {
		severity = domain.ThreatLevelCritical
	} else if strings.Contains(combined, "high") || strings.Contains(combined, "exploit") {
		severity = domain.ThreatLevelHigh
	} else if strings.Contains(combined, "medium") {
		severity = domain.ThreatLevelMedium
	}

	createdAt, _ := time.Parse(time.RFC3339, item.CreatedAt)
	
	return &FeedEvent{
		ID:          fmt.Sprintf("github-%d", item.ID),
		Source:      FeedSourceGitHub,
		Type:        ThreatEventExploit,
		Title:       item.Name,
		Description: item.Description,
		Severity:    severity,
		Targets:     targets,
		Tags:        append(item.Topics, item.Language),
		Confidence:  0.5, // GitHub repos need manual review
		FetchedAt:   time.Now(),
		PublishedAt: createdAt,
		ExternalURL: item.HTMLURL,
	}
}

// extractTargets finds Ethereum addresses in text.
func (g *GitHubFeed) extractTargets(text string) []TargetRef {
	// Match Ethereum addresses
	addrPattern := regexp.MustCompile(`0x[a-fA-F0-9]{40}`)
	matches := addrPattern.FindAllString(text, -1)
	
	var targets []TargetRef
	seen := make(map[string]bool)
	for _, addr := range matches {
		if seen[addr] {
			continue
		}
		seen[addr] = true
		targets = append(targets, TargetRef{
			Type:    TargetTypeContract,
			ChainID: 1, // Default to mainnet
			Address: domain.Address(addr),
		})
	}
	
	return targets
}

// ═══════════════════════════════════════════════════════════════════════════════
// CHAINABUSE FEED
// ═══════════════════════════════════════════════════════════════════════════════

// ChainAbuseFeed fetches scam/abuse reports from ChainAbuse.
type ChainAbuseFeed struct {
	httpClient    *http.Client
	apiKey        string
	baseURL       string
	lastFetchTime time.Time
	config        FeedConfig
}

// NewChainAbuseFeed creates a ChainAbuse feed fetcher.
func NewChainAbuseFeed(config FeedConfig) *ChainAbuseFeed {
	return &ChainAbuseFeed{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		apiKey:     config.APIKey,
		baseURL:    "https://api.chainabuse.com",
		config:     config,
	}
}

func (c *ChainAbuseFeed) Source() FeedSource {
	return FeedSourceChainAbuse
}

func (c *ChainAbuseFeed) SetLastFetchTime(t time.Time) {
	c.lastFetchTime = t
}

// Fetch retrieves recent abuse reports.
func (c *ChainAbuseFeed) Fetch(ctx context.Context) ([]FeedEvent, error) {
	if c.apiKey == "" {
		return nil, nil // Skip if no API key configured
	}

	// ChainAbuse API endpoint (simplified - actual API may differ)
	url := fmt.Sprintf("%s/v1/reports?limit=100", c.baseURL)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ChainAbuse API error: %d", resp.StatusCode)
	}

	// Parse response (structure depends on actual API)
	var reports []struct {
		ID           string `json:"id"`
		Address      string `json:"address"`
		Chain        string `json:"chain"`
		Category     string `json:"category"`
		Description  string `json:"description"`
		ReportCount  int    `json:"report_count"`
		CreatedAt    string `json:"created_at"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&reports); err != nil {
		return nil, err
	}

	var events []FeedEvent
	for _, report := range reports {
		createdAt, _ := time.Parse(time.RFC3339, report.CreatedAt)
		
		// Map category to threat type
		eventType := ThreatEventSuspicious
		severity := domain.ThreatLevelMedium
		
		switch strings.ToLower(report.Category) {
		case "rug_pull":
			eventType = ThreatEventRugPull
			severity = domain.ThreatLevelHigh
		case "phishing":
			eventType = ThreatEventPhishing
			severity = domain.ThreatLevelHigh
		case "scam":
			severity = domain.ThreatLevelMedium
		}

		// Higher confidence with more reports
		confidence := 0.3 + float64(report.ReportCount)*0.1
		if confidence > 0.95 {
			confidence = 0.95
		}

		events = append(events, FeedEvent{
			ID:          fmt.Sprintf("chainabuse-%s", report.ID),
			Source:      FeedSourceChainAbuse,
			Type:        eventType,
			Title:       fmt.Sprintf("%s report: %s", report.Category, report.Address[:10]),
			Description: report.Description,
			Severity:    severity,
			Targets: []TargetRef{{
				Type:    TargetTypeContract,
				ChainID: chainNameToID(report.Chain),
				Address: domain.Address(report.Address),
			}},
			Confidence:  confidence,
			FetchedAt:   time.Now(),
			PublishedAt: createdAt,
		})
	}

	return events, nil
}

// chainNameToID converts chain name to ID.
func chainNameToID(name string) domain.ChainID {
	switch strings.ToLower(name) {
	case "ethereum", "eth":
		return 1
	case "bsc", "bnb":
		return 56
	case "polygon", "matic":
		return 137
	case "arbitrum":
		return 42161
	case "optimism":
		return 10
	case "avalanche", "avax":
		return 43114
	case "base":
		return 8453
	default:
		return 1 // Default to mainnet
	}
}

// ═══════════════════════════════════════════════════════════════════════════════
// INTERNAL FEED (from our scanner)
// ═══════════════════════════════════════════════════════════════════════════════

// InternalFeed converts internal scanner results to feed events.
type InternalFeed struct {
	vulns         []domain.Vulnerability
	lastFetchTime time.Time
}

// NewInternalFeed creates an internal feed from scanner results.
func NewInternalFeed() *InternalFeed {
	return &InternalFeed{}
}

func (i *InternalFeed) Source() FeedSource {
	return FeedSourceInternal
}

func (i *InternalFeed) SetLastFetchTime(t time.Time) {
	i.lastFetchTime = t
}

// AddVulnerabilities adds scanner results to the feed.
func (i *InternalFeed) AddVulnerabilities(vulns []domain.Vulnerability) {
	i.vulns = append(i.vulns, vulns...)
}

// Fetch returns accumulated vulnerabilities as feed events.
func (i *InternalFeed) Fetch(ctx context.Context) ([]FeedEvent, error) {
	var events []FeedEvent
	
	for _, vuln := range i.vulns {
		eventType := vulnTypeToEventType(vuln.Type)
		
		events = append(events, FeedEvent{
			ID:          fmt.Sprintf("internal-%s", vuln.ID),
			Source:      FeedSourceInternal,
			Type:        eventType,
			Title:       vuln.Title,
			Description: vuln.Description,
			Severity:    vuln.Severity,
			Targets: []TargetRef{{
				Type:    TargetTypeContract,
				Address: domain.Address(vuln.ContractID),
			}},
			Evidence: Evidence{
				CVE: vuln.CWE, // CWE is closest we have
			},
			Confidence:  vuln.Confidence,
			FetchedAt:   time.Now(),
			PublishedAt: vuln.DetectedAt,
		})
	}

	// Clear after fetching
	i.vulns = nil
	
	return events, nil
}

// vulnTypeToEventType maps vulnerability types to event types.
func vulnTypeToEventType(vt domain.VulnType) ThreatEventType {
	switch vt {
	case domain.VulnReentrancy:
		return ThreatEventExploit
	case domain.VulnFlashLoan:
		return ThreatEventFlashLoan
	case domain.VulnOracleManipulation:
		return ThreatEventOracleManip
	case domain.VulnRugPull:
		return ThreatEventRugPull
	case domain.VulnPhishing:
		return ThreatEventPhishing
	default:
		return ThreatEventVuln
	}
}
