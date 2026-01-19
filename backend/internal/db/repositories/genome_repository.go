// Package repositories implements GenomeRepository using PostgreSQL.
package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/vigilum/backend/internal/domain"
)

// GenomeRepository implements domain.GenomeRepository using PostgreSQL.
type GenomeRepository struct {
	db *sql.DB
}

// NewGenomeRepository creates a new genome repository.
func NewGenomeRepository(db *sql.DB) *GenomeRepository {
	return &GenomeRepository{db: db}
}

// Create inserts a new genome.
func (r *GenomeRepository) Create(ctx context.Context, genome *domain.Genome) error {
	featuresJSON, err := json.Marshal(genome.Features)
	if err != nil {
		return fmt.Errorf("marshal features: %w", err)
	}

	query := `
		INSERT INTO genomes (id, genome_hash, ipfs_hash, contract_address, chain_id, label, bytecode_size, opcode_count, function_count, complexity_score, features, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	_, err = r.db.ExecContext(ctx, query, genome.ID, genome.GenomeHash, genome.IPFSHash, genome.ContractAddress, genome.ChainID, genome.Label, genome.BytecodeSize, genome.OpcodeCount, genome.FunctionCount, genome.ComplexityScore, featuresJSON, genome.CreatedAt)
	if err != nil {
		return fmt.Errorf("create genome: %w", err)
	}
	return nil
}

// GetByID retrieves a genome by ID.
func (r *GenomeRepository) GetByID(ctx context.Context, id string) (*domain.Genome, error) {
	genome := &domain.Genome{}
	var featuresJSON []byte

	query := `
		SELECT id, genome_hash, ipfs_hash, contract_address, chain_id, label, bytecode_size, opcode_count, function_count, complexity_score, features, created_at
		FROM genomes
		WHERE id = $1
	`
	err := r.db.QueryRowContext(ctx, query, id).
		Scan(&genome.ID, &genome.GenomeHash, &genome.IPFSHash, &genome.ContractAddress, &genome.ChainID, &genome.Label, &genome.BytecodeSize, &genome.OpcodeCount, &genome.FunctionCount, &genome.ComplexityScore, &featuresJSON, &genome.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get genome by id: %w", err)
	}

	if len(featuresJSON) > 0 {
		if err := json.Unmarshal(featuresJSON, &genome.Features); err != nil {
			return nil, fmt.Errorf("unmarshal features: %w", err)
		}
	}
	return genome, nil
}

// GetByHash retrieves a genome by genome hash.
func (r *GenomeRepository) GetByHash(ctx context.Context, hash []byte) (*domain.Genome, error) {
	genome := &domain.Genome{}
	var featuresJSON []byte

	query := `
		SELECT id, genome_hash, ipfs_hash, contract_address, chain_id, label, bytecode_size, opcode_count, function_count, complexity_score, features, created_at
		FROM genomes
		WHERE genome_hash = $1
		LIMIT 1
	`
	err := r.db.QueryRowContext(ctx, query, hash).
		Scan(&genome.ID, &genome.GenomeHash, &genome.IPFSHash, &genome.ContractAddress, &genome.ChainID, &genome.Label, &genome.BytecodeSize, &genome.OpcodeCount, &genome.FunctionCount, &genome.ComplexityScore, &featuresJSON, &genome.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get genome by hash: %w", err)
	}

	if len(featuresJSON) > 0 {
		if err := json.Unmarshal(featuresJSON, &genome.Features); err != nil {
			return nil, fmt.Errorf("unmarshal features: %w", err)
		}
	}
	return genome, nil
}

// GetByContractAddress retrieves genomes for a contract.
func (r *GenomeRepository) GetByContractAddress(ctx context.Context, chainID domain.ChainID, address domain.Address) (*domain.Genome, error) {
	genome := &domain.Genome{}
	var featuresJSON []byte

	query := `
		SELECT id, genome_hash, ipfs_hash, contract_address, chain_id, label, bytecode_size, opcode_count, function_count, complexity_score, features, created_at
		FROM genomes
		WHERE chain_id = $1 AND contract_address = $2
		LIMIT 1
	`
	err := r.db.QueryRowContext(ctx, query, chainID, address).
		Scan(&genome.ID, &genome.GenomeHash, &genome.IPFSHash, &genome.ContractAddress, &genome.ChainID, &genome.Label, &genome.BytecodeSize, &genome.OpcodeCount, &genome.FunctionCount, &genome.ComplexityScore, &featuresJSON, &genome.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get genome by contract: %w", err)
	}

	if len(featuresJSON) > 0 {
		if err := json.Unmarshal(featuresJSON, &genome.Features); err != nil {
			return nil, fmt.Errorf("unmarshal features: %w", err)
		}
	}
	return genome, nil
}

// Update modifies an existing genome.
func (r *GenomeRepository) Update(ctx context.Context, id string, genome *domain.Genome) error {
	featuresJSON, err := json.Marshal(genome.Features)
	if err != nil {
		return fmt.Errorf("marshal features: %w", err)
	}

	query := `
		UPDATE genomes
		SET label = $1, complexity_score = $2, features = $3
		WHERE id = $4
	`
	result, err := r.db.ExecContext(ctx, query, genome.Label, genome.ComplexityScore, featuresJSON, id)
	if err != nil {
		return fmt.Errorf("update genome: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// Delete removes a genome.
func (r *GenomeRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM genomes WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete genome: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// ListByLabel retrieves genomes with a specific label.
func (r *GenomeRepository) ListByLabel(ctx context.Context, label string, limit int, offset int) ([]*domain.Genome, error) {
	query := `
		SELECT id, genome_hash, ipfs_hash, contract_address, chain_id, label, bytecode_size, opcode_count, function_count, complexity_score, features, created_at
		FROM genomes
		WHERE label = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.QueryContext(ctx, query, label, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list by label: %w", err)
	}
	defer rows.Close()

	var genomes []*domain.Genome
	for rows.Next() {
		genome := &domain.Genome{}
		var featuresJSON []byte
		err := rows.Scan(&genome.ID, &genome.GenomeHash, &genome.IPFSHash, &genome.ContractAddress, &genome.ChainID, &genome.Label, &genome.BytecodeSize, &genome.OpcodeCount, &genome.FunctionCount, &genome.ComplexityScore, &featuresJSON, &genome.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan genome: %w", err)
		}

		if len(featuresJSON) > 0 {
			if err := json.Unmarshal(featuresJSON, &genome.Features); err != nil {
				return nil, fmt.Errorf("unmarshal features: %w", err)
			}
		}
		genomes = append(genomes, genome)
	}
	return genomes, rows.Err()
}

// ListSimilar retrieves genomes similar to a given one.
func (r *GenomeRepository) ListSimilar(ctx context.Context, genomeID string, threshold float64, limit int) ([]*domain.Genome, error) {
	query := `
		SELECT id, genome_hash, ipfs_hash, contract_address, chain_id, label, bytecode_size, opcode_count, function_count, complexity_score, features, created_at
		FROM genomes
		WHERE id != $1 AND complexity_score >= $2
		ORDER BY complexity_score DESC
		LIMIT $3
	`
	rows, err := r.db.QueryContext(ctx, query, genomeID, threshold, limit)
	if err != nil {
		return nil, fmt.Errorf("list similar: %w", err)
	}
	defer rows.Close()

	var genomes []*domain.Genome
	for rows.Next() {
		genome := &domain.Genome{}
		var featuresJSON []byte
		err := rows.Scan(&genome.ID, &genome.GenomeHash, &genome.IPFSHash, &genome.ContractAddress, &genome.ChainID, &genome.Label, &genome.BytecodeSize, &genome.OpcodeCount, &genome.FunctionCount, &genome.ComplexityScore, &featuresJSON, &genome.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan genome: %w", err)
		}

		if len(featuresJSON) > 0 {
			if err := json.Unmarshal(featuresJSON, &genome.Features); err != nil {
				return nil, fmt.Errorf("unmarshal features: %w", err)
			}
		}
		genomes = append(genomes, genome)
	}
	return genomes, rows.Err()
}

// Count returns the total number of genomes.
func (r *GenomeRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM genomes`
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count genomes: %w", err)
	}
	return count, nil
}

// CountByLabel returns the number of genomes with a specific label.
func (r *GenomeRepository) CountByLabel(ctx context.Context, label string) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM genomes WHERE label = $1`
	err := r.db.QueryRowContext(ctx, query, label).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count by label: %w", err)
	}
	return count, nil
}

// GetDistribution returns count of genomes per label.
func (r *GenomeRepository) GetDistribution(ctx context.Context) (map[string]int64, error) {
	query := `SELECT label, COUNT(*) FROM genomes GROUP BY label`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("get distribution: %w", err)
	}
	defer rows.Close()

	distribution := make(map[string]int64)
	for rows.Next() {
		var label string
		var count int64
		if err := rows.Scan(&label, &count); err != nil {
			return nil, fmt.Errorf("scan distribution: %w", err)
		}
		distribution[label] = count
	}
	return distribution, rows.Err()
}
