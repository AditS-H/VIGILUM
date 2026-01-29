// Package repositories implements HumanProofRepository using PostgreSQL.
package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/vigilum/backend/internal/domain"
)

// HumanProofRepository implements domain.HumanProofRepository using PostgreSQL.
type HumanProofRepository struct {
	db *sql.DB
}

// NewHumanProofRepository creates a new human proof repository.
func NewHumanProofRepository(db *sql.DB) *HumanProofRepository {
	return &HumanProofRepository{db: db}
}

// Create inserts a new human proof record.
func (r *HumanProofRepository) Create(ctx context.Context, proof *domain.HumanProof) error {
	proofDataJSON, err := json.Marshal(proof.ProofData)
	if err != nil {
		return fmt.Errorf("marshal proof data: %w", err)
	}

	query := `
		INSERT INTO human_proofs (id, user_id, proof_hash, proof_data, verified, created_at, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err = r.db.ExecContext(ctx, query, proof.ID, proof.UserID, proof.ProofHash, proofDataJSON, proof.Verified, proof.CreatedAt, proof.ExpiresAt)
	if err != nil {
		return fmt.Errorf("create human proof: %w", err)
	}
	return nil
}

// GetByID retrieves a proof by ID.
func (r *HumanProofRepository) GetByID(ctx context.Context, id string) (*domain.HumanProof, error) {
	proof := &domain.HumanProof{}
	var proofDataJSON []byte

	query := `
		SELECT id, user_id, proof_hash, proof_data, verified, created_at, verified_at, verifier_address, tx_hash, expires_at
		FROM human_proofs
		WHERE id = $1
	`
	err := r.db.QueryRowContext(ctx, query, id).
		Scan(&proof.ID, &proof.UserID, &proof.ProofHash, &proofDataJSON, &proof.Verified, &proof.CreatedAt, &proof.VerifiedAt, &proof.VerifierAddress, &proof.TxHash, &proof.ExpiresAt)
	if err == sql.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get human proof by id: %w", err)
	}

	if len(proofDataJSON) > 0 {
		proof.ProofData = &domain.ProofData{}
		if err := json.Unmarshal(proofDataJSON, proof.ProofData); err != nil {
			return nil, fmt.Errorf("unmarshal proof data: %w", err)
		}
	}
	return proof, nil
}

// GetByHash retrieves a proof by proof hash.
func (r *HumanProofRepository) GetByHash(ctx context.Context, hash []byte) (*domain.HumanProof, error) {
	proof := &domain.HumanProof{}
	var proofDataJSON []byte

	query := `
		SELECT id, user_id, proof_hash, proof_data, verified, created_at, verified_at, verifier_address, tx_hash, expires_at
		FROM human_proofs
		WHERE proof_hash = $1
		LIMIT 1
	`
	err := r.db.QueryRowContext(ctx, query, hash).
		Scan(&proof.ID, &proof.UserID, &proof.ProofHash, &proofDataJSON, &proof.Verified, &proof.CreatedAt, &proof.VerifiedAt, &proof.VerifierAddress, &proof.TxHash, &proof.ExpiresAt)
	if err == sql.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get human proof by hash: %w", err)
	}

	if len(proofDataJSON) > 0 {
		proof.ProofData = &domain.ProofData{}
		if err := json.Unmarshal(proofDataJSON, proof.ProofData); err != nil {
			return nil, fmt.Errorf("unmarshal proof data: %w", err)
		}
	}
	return proof, nil
}

// GetByUserID retrieves all proofs for a user, ordered by created_at DESC.
func (r *HumanProofRepository) GetByUserID(ctx context.Context, userID string, limit int, offset int) ([]*domain.HumanProof, error) {
	query := `
		SELECT id, user_id, proof_hash, proof_data, verified, created_at, verified_at, verifier_address, tx_hash, expires_at
		FROM human_proofs
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("get human proofs by user id: %w", err)
	}
	defer rows.Close()

	var proofs []*domain.HumanProof
	for rows.Next() {
		proof := &domain.HumanProof{}
		var proofDataJSON []byte
		err := rows.Scan(&proof.ID, &proof.UserID, &proof.ProofHash, &proofDataJSON, &proof.Verified, &proof.CreatedAt, &proof.VerifiedAt, &proof.VerifierAddress, &proof.TxHash, &proof.ExpiresAt)
		if err != nil {
			return nil, fmt.Errorf("scan human proof: %w", err)
		}

		if len(proofDataJSON) > 0 {
			proof.ProofData = &domain.ProofData{}
			if err := json.Unmarshal(proofDataJSON, proof.ProofData); err != nil {
				return nil, fmt.Errorf("unmarshal proof data: %w", err)
			}
		}
		proofs = append(proofs, proof)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return proofs, nil
}

// Update modifies an existing proof.
func (r *HumanProofRepository) Update(ctx context.Context, id string, proof *domain.HumanProof) error {
	proofDataJSON, err := json.Marshal(proof.ProofData)
	if err != nil {
		return fmt.Errorf("marshal proof data: %w", err)
	}

	query := `
		UPDATE human_proofs
		SET verified = $1, verified_at = $2, verifier_address = $3, tx_hash = $4, expires_at = $5, proof_data = $6
		WHERE id = $7
	`
	result, err := r.db.ExecContext(ctx, query, proof.Verified, proof.VerifiedAt, proof.VerifierAddress, proof.TxHash, proof.ExpiresAt, proofDataJSON, id)
	if err != nil {
		return fmt.Errorf("update human proof: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// MarkVerified marks a proof as verified by a contract.
func (r *HumanProofRepository) MarkVerified(ctx context.Context, id string, verifierAddr string, txHash string) error {
	query := `
		UPDATE human_proofs
		SET verified = true, verified_at = NOW(), verifier_address = $1, tx_hash = $2
		WHERE id = $3
	`
	result, err := r.db.ExecContext(ctx, query, verifierAddr, txHash, id)
	if err != nil {
		return fmt.Errorf("mark verified: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// Delete removes a proof record.
func (r *HumanProofRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM human_proofs WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete human proof: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// DeleteExpired removes all expired proofs.
func (r *HumanProofRepository) DeleteExpired(ctx context.Context) (int64, error) {
	query := `DELETE FROM human_proofs WHERE expires_at IS NOT NULL AND expires_at < NOW()`
	result, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("delete expired proofs: %w", err)
	}
	return result.RowsAffected()
}

// CountByUserID returns the number of proofs for a user.
func (r *HumanProofRepository) CountByUserID(ctx context.Context, userID string) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM human_proofs WHERE user_id = $1`
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count by user id: %w", err)
	}
	return count, nil
}

// CountVerifiedByUserID returns the number of verified proofs for a user.
func (r *HumanProofRepository) CountVerifiedByUserID(ctx context.Context, userID string) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM human_proofs WHERE user_id = $1 AND verified = true`
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count verified by user id: %w", err)
	}
	return count, nil
}

// GetLatestVerifiedTimestamp returns the most recent verification timestamp for a user.
// Returns zero time if no verified proofs exist.
func (r *HumanProofRepository) GetLatestVerifiedTimestamp(ctx context.Context, userID string) (time.Time, error) {
	var verifiedAt sql.NullTime
	query := `
		SELECT verified_at 
		FROM human_proofs 
		WHERE user_id = $1 AND verified = true AND verified_at IS NOT NULL
		ORDER BY verified_at DESC 
		LIMIT 1
	`
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&verifiedAt)
	if err == sql.ErrNoRows {
		return time.Time{}, nil // No verified proofs
	}
	if err != nil {
		return time.Time{}, fmt.Errorf("get latest verified timestamp: %w", err)
	}
	if !verifiedAt.Valid {
		return time.Time{}, nil
	}
	return verifiedAt.Time, nil
}
