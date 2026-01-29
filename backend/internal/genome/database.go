// Package genome implements the contract pattern database and similarity search
package genome

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// ContractGenome represents a contract's bytecode pattern
type ContractGenome struct {
	ID              string    `json:"id"`
	ContractAddress string    `json:"contract_address"`
	BytecodeHash    string    `json:"bytecode_hash"`
	Patterns        []string  `json:"patterns"`
	Features        []float64 `json:"features"`
	RiskScore       float64   `json:"risk_score"`
	Vulnerabilities []string  `json:"vulnerabilities"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// GenomeDatabase stores and searches contract patterns
type GenomeDatabase struct {
	logger      *slog.Logger
	mu          sync.RWMutex
	genomes     map[string]*ContractGenome
	indexByHash map[string][]*ContractGenome
	indexByVuln map[string][]*ContractGenome
}

// NewGenomeDatabase creates a new genome database
func NewGenomeDatabase(logger *slog.Logger) *GenomeDatabase {
	return &GenomeDatabase{
		logger:      logger.With("service", "genome-db"),
		genomes:     make(map[string]*ContractGenome),
		indexByHash: make(map[string][]*ContractGenome),
		indexByVuln: make(map[string][]*ContractGenome),
	}
}

// Store adds or updates a contract genome
func (db *GenomeDatabase) Store(ctx context.Context, genome *ContractGenome) error {
	if genome.ID == "" {
		genome.ID = generateID(genome.ContractAddress)
	}

	genome.CreatedAt = time.Now()
	genome.UpdatedAt = time.Now()

	db.mu.Lock()
	defer db.mu.Unlock()

	// Store main record
	db.genomes[genome.ID] = genome

	// Index by bytecode hash for fast similarity search
	db.indexByHash[genome.BytecodeHash] = append(
		db.indexByHash[genome.BytecodeHash],
		genome,
	)

	// Index by vulnerabilities for quick lookups
	for _, vuln := range genome.Vulnerabilities {
		db.indexByVuln[vuln] = append(db.indexByVuln[vuln], genome)
	}

	db.logger.Info("genome stored",
		"id", genome.ID,
		"address", genome.ContractAddress,
		"risk_score", genome.RiskScore,
	)

	return nil
}

// Get retrieves a genome by ID
func (db *GenomeDatabase) Get(ctx context.Context, id string) (*ContractGenome, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	genome, exists := db.genomes[id]
	if !exists {
		return nil, fmt.Errorf("genome not found: %s", id)
	}

	return genome, nil
}

// FindSimilar finds contracts with similar patterns
func (db *GenomeDatabase) FindSimilar(ctx context.Context, genome *ContractGenome, limit int) ([]*ContractGenome, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	similar := []*ContractGenome{}

	// Find contracts with same bytecode hash (exact duplicates)
	if candidates, exists := db.indexByHash[genome.BytecodeHash]; exists {
		similar = append(similar, candidates...)
	}

	// Find contracts with shared vulnerabilities
	vulnMap := make(map[string]bool)
	for _, vuln := range genome.Vulnerabilities {
		vulnMap[vuln] = true
	}

	for vuln := range vulnMap {
		if candidates, exists := db.indexByVuln[vuln]; exists {
			for _, candidate := range candidates {
				// Check if already in similar list
				found := false
				for _, s := range similar {
					if s.ID == candidate.ID {
						found = true
						break
					}
				}
				if !found {
					similar = append(similar, candidate)
				}
			}
		}
	}

	// Calculate similarity scores
	similarities := make([]struct {
		Genome   *ContractGenome
		Score    float64
		Exploited bool
	}, len(similar))

	for i, candidate := range similar {
		score := calculateSimilarity(genome, candidate)
		exploited := len(candidate.Vulnerabilities) > 0

		similarities[i] = struct {
			Genome   *ContractGenome
			Score    float64
			Exploited bool
		}{
			Genome:   candidate,
			Score:    score,
			Exploited: exploited,
		}
	}

	// Sort by similarity (highest first) and return top N
	// In production, use proper sorting library
	if len(similarities) > limit {
		similarities = similarities[:limit]
	}

	result := make([]*ContractGenome, len(similarities))
	for i, s := range similarities {
		result[i] = s.Genome
	}

	return result, nil
}

// FindByVulnerability finds all contracts with a specific vulnerability
func (db *GenomeDatabase) FindByVulnerability(ctx context.Context, vuln string, limit int) ([]*ContractGenome, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	candidates := db.indexByVuln[vuln]
	if len(candidates) > limit {
		candidates = candidates[:limit]
	}

	db.logger.Info("found contracts with vulnerability",
		"vulnerability", vuln,
		"count", len(candidates),
	)

	return candidates, nil
}

// GetVulnerabilityStats returns statistics on vulnerabilities
func (db *GenomeDatabase) GetVulnerabilityStats(ctx context.Context) map[string]int {
	db.mu.RLock()
	defer db.mu.RUnlock()

	stats := make(map[string]int)

	for vuln, candidates := range db.indexByVuln {
		stats[vuln] = len(candidates)
	}

	return stats
}

// GetRiskDistribution returns distribution of risk scores
func (db *GenomeDatabase) GetRiskDistribution(ctx context.Context) map[string]int {
	db.mu.RLock()
	defer db.mu.RUnlock()

	distribution := map[string]int{
		"critical": 0, // >= 80
		"high":     0, // 60-79
		"medium":   0, // 40-59
		"low":      0, // < 40
	}

	for _, genome := range db.genomes {
		if genome.RiskScore >= 80 {
			distribution["critical"]++
		} else if genome.RiskScore >= 60 {
			distribution["high"]++
		} else if genome.RiskScore >= 40 {
			distribution["medium"]++
		} else {
			distribution["low"]++
		}
	}

	return distribution
}

// GetTrendingVulnerabilities returns recently discovered vulnerabilities
func (db *GenomeDatabase) GetTrendingVulnerabilities(ctx context.Context, days int) []string {
	db.mu.RLock()
	defer db.mu.RUnlock()

	vulnCount := make(map[string]int)
	cutoffTime := time.Now().Add(time.Duration(-days) * 24 * time.Hour)

	for _, genome := range db.genomes {
		if genome.CreatedAt.After(cutoffTime) {
			for _, vuln := range genome.Vulnerabilities {
				vulnCount[vuln]++
			}
		}
	}

	// Convert to sorted list (top 10)
	trending := []string{}
	for vuln := range vulnCount {
		trending = append(trending, vuln)
		if len(trending) >= 10 {
			break
		}
	}

	return trending
}

// Delete removes a genome from the database
func (db *GenomeDatabase) Delete(ctx context.Context, id string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	genome, exists := db.genomes[id]
	if !exists {
		return fmt.Errorf("genome not found: %s", id)
	}

	// Remove from indexes
	delete(db.genomes, id)

	// Remove from hash index
	if candidates, exists := db.indexByHash[genome.BytecodeHash]; exists {
		filtered := []*ContractGenome{}
		for _, c := range candidates {
			if c.ID != id {
				filtered = append(filtered, c)
			}
		}
		db.indexByHash[genome.BytecodeHash] = filtered
	}

	// Remove from vulnerability indexes
	for _, vuln := range genome.Vulnerabilities {
		if candidates, exists := db.indexByVuln[vuln]; exists {
			filtered := []*ContractGenome{}
			for _, c := range candidates {
				if c.ID != id {
					filtered = append(filtered, c)
				}
			}
			db.indexByVuln[vuln] = filtered
		}
	}

	db.logger.Info("genome deleted", "id", id)
	return nil
}

// Helper functions

// generateID creates a unique ID for a genome
func generateID(contractAddr string) string {
	hash := sha256.Sum256([]byte(contractAddr + time.Now().String()))
	return hex.EncodeToString(hash[:8])
}

// calculateSimilarity computes similarity score between two genomes
func calculateSimilarity(g1, g2 *ContractGenome) float64 {
	// Check if exact match
	if g1.BytecodeHash == g2.BytecodeHash {
		return 1.0
	}

	// Count shared vulnerabilities
	shared := 0
	vulnMap := make(map[string]bool)
	for _, v := range g1.Vulnerabilities {
		vulnMap[v] = true
	}
	for _, v := range g2.Vulnerabilities {
		if vulnMap[v] {
			shared++
		}
	}

	if len(g1.Vulnerabilities) == 0 && len(g2.Vulnerabilities) == 0 {
		return 0.0
	}

	total := len(g1.Vulnerabilities) + len(g2.Vulnerabilities)
	if total == 0 {
		return 0.0
	}

	return float64(shared) / float64(total)
}
