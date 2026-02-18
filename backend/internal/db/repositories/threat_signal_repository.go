// Package repositories implements ThreatSignalRepository using PostgreSQL.
package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/vigilum/backend/internal/domain"
)

// ThreatSignalRepository implements domain.ThreatSignalRepository using PostgreSQL.
type ThreatSignalRepository struct {
	db *sql.DB
}

// NewThreatSignalRepository creates a new threat signal repository.
func NewThreatSignalRepository(db *sql.DB) *ThreatSignalRepository {
	return &ThreatSignalRepository{db: db}
}

// Create inserts a new threat signal.
func (r *ThreatSignalRepository) Create(ctx context.Context, signal *domain.ThreatSignal) error {
	metadataJSON, err := json.Marshal(signal.Metadata)
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}

	query := `
		INSERT INTO threat_signals (id, chain_id, entity_address, signal_type, risk_score, threat_level, confidence, source, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err = r.db.ExecContext(ctx, query, signal.ID, signal.ChainID, signal.Address, signal.SignalType, signal.RiskScore, signal.ThreatLevel, signal.Confidence, signal.Source, metadataJSON, signal.CreatedAt)
	if err != nil {
		return fmt.Errorf("create threat signal: %w", err)
	}
	return nil
}

// GetByID retrieves a signal by ID.
func (r *ThreatSignalRepository) GetByID(ctx context.Context, id string) (*domain.ThreatSignal, error) {
	signal := &domain.ThreatSignal{}
	var metadataJSON []byte

	query := `
		SELECT id, chain_id, entity_address, signal_type, risk_score, threat_level, confidence, source, metadata, created_at, published_at
		FROM threat_signals
		WHERE id = $1
	`
	err := r.db.QueryRowContext(ctx, query, id).
		Scan(&signal.ID, &signal.ChainID, &signal.Address, &signal.SignalType, &signal.RiskScore, &signal.ThreatLevel, &signal.Confidence, &signal.Source, &metadataJSON, &signal.CreatedAt, &signal.PublishedAt)
	if err == sql.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get threat signal by id: %w", err)
	}

	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &signal.Metadata); err != nil {
			return nil, fmt.Errorf("unmarshal metadata: %w", err)
		}
	}
	return signal, nil
}

// GetByEntity retrieves signals for an entity on a chain, ordered by risk descending.
func (r *ThreatSignalRepository) GetByEntity(ctx context.Context, chainID domain.ChainID, address domain.Address, limit int) ([]*domain.ThreatSignal, error) {
	query := `
		SELECT id, chain_id, entity_address, signal_type, risk_score, threat_level, confidence, source, metadata, created_at, published_at
		FROM threat_signals
		WHERE chain_id = $1 AND entity_address = $2
		ORDER BY risk_score DESC, created_at DESC
		LIMIT $3
	`
	rows, err := r.db.QueryContext(ctx, query, chainID, address, limit)
	if err != nil {
		return nil, fmt.Errorf("get by entity: %w", err)
	}
	defer rows.Close()

	var signals []*domain.ThreatSignal
	for rows.Next() {
		signal := &domain.ThreatSignal{}
		var metadataJSON []byte
		err := rows.Scan(&signal.ID, &signal.ChainID, &signal.Address, &signal.SignalType, &signal.RiskScore, &signal.ThreatLevel, &signal.Confidence, &signal.Source, &metadataJSON, &signal.CreatedAt, &signal.PublishedAt)
		if err != nil {
			return nil, fmt.Errorf("scan signal: %w", err)
		}

		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &signal.Metadata); err != nil {
				return nil, fmt.Errorf("unmarshal metadata: %w", err)
			}
		}
		signals = append(signals, signal)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return signals, nil
}

// GetUnpublished retrieves signals not yet published on-chain.
func (r *ThreatSignalRepository) GetUnpublished(ctx context.Context, limit int) ([]*domain.ThreatSignal, error) {
	query := `
		SELECT id, chain_id, entity_address, signal_type, risk_score, threat_level, confidence, source, metadata, created_at, published_at
		FROM threat_signals
		WHERE published_at IS NULL
		ORDER BY created_at ASC
		LIMIT $1
	`
	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("get unpublished: %w", err)
	}
	defer rows.Close()

	var signals []*domain.ThreatSignal
	for rows.Next() {
		signal := &domain.ThreatSignal{}
		var metadataJSON []byte
		err := rows.Scan(&signal.ID, &signal.ChainID, &signal.Address, &signal.SignalType, &signal.RiskScore, &signal.ThreatLevel, &signal.Confidence, &signal.Source, &metadataJSON, &signal.CreatedAt, &signal.PublishedAt)
		if err != nil {
			return nil, fmt.Errorf("scan signal: %w", err)
		}

		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &signal.Metadata); err != nil {
				return nil, fmt.Errorf("unmarshal metadata: %w", err)
			}
		}
		signals = append(signals, signal)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return signals, nil
}

// Update modifies an existing signal.
func (r *ThreatSignalRepository) Update(ctx context.Context, id string, signal *domain.ThreatSignal) error {
	metadataJSON, err := json.Marshal(signal.Metadata)
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}

	query := `
		UPDATE threat_signals
		SET risk_score = $1, threat_level = $2, confidence = $3, metadata = $4, published_at = $5
		WHERE id = $6
	`
	result, err := r.db.ExecContext(ctx, query, signal.RiskScore, signal.ThreatLevel, signal.Confidence, metadataJSON, signal.PublishedAt, id)
	if err != nil {
		return fmt.Errorf("update threat signal: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// MarkPublished marks a signal as published on-chain.
func (r *ThreatSignalRepository) MarkPublished(ctx context.Context, id string) error {
	query := `UPDATE threat_signals SET published_at = NOW() WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("mark published: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// Delete removes a threat signal.
func (r *ThreatSignalRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM threat_signals WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete threat signal: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// GetHighRisk returns all signals above a risk threshold.
func (r *ThreatSignalRepository) GetHighRisk(ctx context.Context, threshold int, limit int) ([]*domain.ThreatSignal, error) {
	query := `
		SELECT id, chain_id, entity_address, signal_type, risk_score, threat_level, confidence, source, metadata, created_at, published_at
		FROM threat_signals
		WHERE risk_score >= $1
		ORDER BY risk_score DESC
		LIMIT $2
	`
	rows, err := r.db.QueryContext(ctx, query, threshold, limit)
	if err != nil {
		return nil, fmt.Errorf("get high risk: %w", err)
	}
	defer rows.Close()

	var signals []*domain.ThreatSignal
	for rows.Next() {
		signal := &domain.ThreatSignal{}
		var metadataJSON []byte
		err := rows.Scan(&signal.ID, &signal.ChainID, &signal.Address, &signal.SignalType, &signal.RiskScore, &signal.ThreatLevel, &signal.Confidence, &signal.Source, &metadataJSON, &signal.CreatedAt, &signal.PublishedAt)
		if err != nil {
			return nil, fmt.Errorf("scan signal: %w", err)
		}

		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &signal.Metadata); err != nil {
				return nil, fmt.Errorf("unmarshal metadata: %w", err)
			}
		}
		signals = append(signals, signal)
	}
	return signals, rows.Err()
}

// GetByCriticalSignalType returns signals of critical threat types.
func (r *ThreatSignalRepository) GetByCriticalSignalType(ctx context.Context, limit int) ([]*domain.ThreatSignal, error) {
	query := `
		SELECT id, chain_id, entity_address, signal_type, risk_score, threat_level, confidence, source, metadata, created_at, published_at
		FROM threat_signals
		WHERE threat_level = 'critical'
		ORDER BY risk_score DESC
		LIMIT $1
	`
	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("get critical signals: %w", err)
	}
	defer rows.Close()

	var signals []*domain.ThreatSignal
	for rows.Next() {
		signal := &domain.ThreatSignal{}
		var metadataJSON []byte
		err := rows.Scan(&signal.ID, &signal.ChainID, &signal.Address, &signal.SignalType, &signal.RiskScore, &signal.ThreatLevel, &signal.Confidence, &signal.Source, &metadataJSON, &signal.CreatedAt, &signal.PublishedAt)
		if err != nil {
			return nil, fmt.Errorf("scan signal: %w", err)
		}

		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &signal.Metadata); err != nil {
				return nil, fmt.Errorf("unmarshal metadata: %w", err)
			}
		}
		signals = append(signals, signal)
	}
	return signals, rows.Err()
}

// Count returns the total number of signals.
func (r *ThreatSignalRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM threat_signals`
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count signals: %w", err)
	}
	return count, nil
}

// CountByEntity returns the number of signals for an entity.
func (r *ThreatSignalRepository) CountByEntity(ctx context.Context, chainID domain.ChainID, address domain.Address) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM threat_signals WHERE chain_id = $1 AND entity_address = $2`
	err := r.db.QueryRowContext(ctx, query, chainID, address).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count by entity: %w", err)
	}
	return count, nil
}
