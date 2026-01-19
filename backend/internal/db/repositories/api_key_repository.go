// Package repositories implements APIKeyRepository using PostgreSQL.
package repositories

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"

	"github.com/vigilum/backend/internal/domain"
)

// APIKeyRepository implements domain.APIKeyRepository using PostgreSQL.
type APIKeyRepository struct {
	db *sql.DB
}

// NewAPIKeyRepository creates a new API key repository.
func NewAPIKeyRepository(db *sql.DB) *APIKeyRepository {
	return &APIKeyRepository{db: db}
}

// Create inserts a new API key.
func (r *APIKeyRepository) Create(ctx context.Context, key *domain.APIKey) error {
	query := `
		INSERT INTO api_keys (id, key_hash, user_id, name, tier, rate_limit, requests_today, created_at, expires_at, revoked)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := r.db.ExecContext(ctx, query, key.ID, key.KeyHash, key.UserID, key.Name, key.Tier, key.RateLimit, 0, key.CreatedAt, key.ExpiresAt, false)
	if err != nil {
		return fmt.Errorf("create api key: %w", err)
	}
	return nil
}

// GetByHash retrieves an API key by hashed key value.
func (r *APIKeyRepository) GetByHash(ctx context.Context, keyHash []byte) (*domain.APIKey, error) {
	key := &domain.APIKey{}
	query := `
		SELECT id, key_hash, user_id, name, tier, rate_limit, requests_today, created_at, last_used, expires_at, revoked
		FROM api_keys
		WHERE key_hash = $1 AND revoked = false
		LIMIT 1
	`
	err := r.db.QueryRowContext(ctx, query, keyHash).
		Scan(&key.ID, &key.KeyHash, &key.UserID, &key.Name, &key.Tier, &key.RateLimit, &key.RequestsToday, &key.CreatedAt, &key.LastUsed, &key.ExpiresAt, &key.Revoked)
	if err == sql.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get api key by hash: %w", err)
	}
	return key, nil
}

// GetByID retrieves an API key by ID.
func (r *APIKeyRepository) GetByID(ctx context.Context, id string) (*domain.APIKey, error) {
	key := &domain.APIKey{}
	query := `
		SELECT id, key_hash, user_id, name, tier, rate_limit, requests_today, created_at, last_used, expires_at, revoked
		FROM api_keys
		WHERE id = $1
	`
	err := r.db.QueryRowContext(ctx, query, id).
		Scan(&key.ID, &key.KeyHash, &key.UserID, &key.Name, &key.Tier, &key.RateLimit, &key.RequestsToday, &key.CreatedAt, &key.LastUsed, &key.ExpiresAt, &key.Revoked)
	if err == sql.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get api key by id: %w", err)
	}
	return key, nil
}

// GetByUserID retrieves all API keys for a user.
func (r *APIKeyRepository) GetByUserID(ctx context.Context, userID string) ([]*domain.APIKey, error) {
	query := `
		SELECT id, key_hash, user_id, name, tier, rate_limit, requests_today, created_at, last_used, expires_at, revoked
		FROM api_keys
		WHERE user_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("get by user id: %w", err)
	}
	defer rows.Close()

	var keys []*domain.APIKey
	for rows.Next() {
		key := &domain.APIKey{}
		err := rows.Scan(&key.ID, &key.KeyHash, &key.UserID, &key.Name, &key.Tier, &key.RateLimit, &key.RequestsToday, &key.CreatedAt, &key.LastUsed, &key.ExpiresAt, &key.Revoked)
		if err != nil {
			return nil, fmt.Errorf("scan api key: %w", err)
		}
		keys = append(keys, key)
	}
	return keys, rows.Err()
}

// Update modifies an existing API key.
func (r *APIKeyRepository) Update(ctx context.Context, id string, key *domain.APIKey) error {
	query := `
		UPDATE api_keys
		SET name = $1, tier = $2, rate_limit = $3, expires_at = $4
		WHERE id = $5
	`
	result, err := r.db.ExecContext(ctx, query, key.Name, key.Tier, key.RateLimit, key.ExpiresAt, id)
	if err != nil {
		return fmt.Errorf("update api key: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// UpdateLastUsed updates the last_used timestamp.
func (r *APIKeyRepository) UpdateLastUsed(ctx context.Context, id string) error {
	query := `UPDATE api_keys SET last_used = NOW() WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("update last used: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// UpdateRequestCount increments the daily request counter.
func (r *APIKeyRepository) UpdateRequestCount(ctx context.Context, id string) error {
	query := `UPDATE api_keys SET requests_today = requests_today + 1 WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("update request count: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// ResetDailyCount resets the requests_today counter.
func (r *APIKeyRepository) ResetDailyCount(ctx context.Context, id string) error {
	query := `UPDATE api_keys SET requests_today = 0 WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("reset daily count: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// Revoke marks an API key as revoked.
func (r *APIKeyRepository) Revoke(ctx context.Context, id string) error {
	query := `UPDATE api_keys SET revoked = true WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("revoke api key: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// Delete removes an API key.
func (r *APIKeyRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM api_keys WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete api key: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// ListByUserID retrieves all non-revoked keys for a user.
func (r *APIKeyRepository) ListByUserID(ctx context.Context, userID string) ([]*domain.APIKey, error) {
	query := `
		SELECT id, key_hash, user_id, name, tier, rate_limit, requests_today, created_at, last_used, expires_at, revoked
		FROM api_keys
		WHERE user_id = $1 AND revoked = false
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list by user id: %w", err)
	}
	defer rows.Close()

	var keys []*domain.APIKey
	for rows.Next() {
		key := &domain.APIKey{}
		err := rows.Scan(&key.ID, &key.KeyHash, &key.UserID, &key.Name, &key.Tier, &key.RateLimit, &key.RequestsToday, &key.CreatedAt, &key.LastUsed, &key.ExpiresAt, &key.Revoked)
		if err != nil {
			return nil, fmt.Errorf("scan api key: %w", err)
		}
		keys = append(keys, key)
	}
	return keys, rows.Err()
}

// ListByTier retrieves all active keys in a tier.
func (r *APIKeyRepository) ListByTier(ctx context.Context, tier string) ([]*domain.APIKey, error) {
	query := `
		SELECT id, key_hash, user_id, name, tier, rate_limit, requests_today, created_at, last_used, expires_at, revoked
		FROM api_keys
		WHERE tier = $1 AND revoked = false
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, tier)
	if err != nil {
		return nil, fmt.Errorf("list by tier: %w", err)
	}
	defer rows.Close()

	var keys []*domain.APIKey
	for rows.Next() {
		key := &domain.APIKey{}
		err := rows.Scan(&key.ID, &key.KeyHash, &key.UserID, &key.Name, &key.Tier, &key.RateLimit, &key.RequestsToday, &key.CreatedAt, &key.LastUsed, &key.ExpiresAt, &key.Revoked)
		if err != nil {
			return nil, fmt.Errorf("scan api key: %w", err)
		}
		keys = append(keys, key)
	}
	return keys, rows.Err()
}

// GetExpiring retrieves keys expiring soon (next N days).
func (r *APIKeyRepository) GetExpiring(ctx context.Context, days int) ([]*domain.APIKey, error) {
	query := `
		SELECT id, key_hash, user_id, name, tier, rate_limit, requests_today, created_at, last_used, expires_at, revoked
		FROM api_keys
		WHERE expires_at IS NOT NULL AND expires_at <= NOW() + INTERVAL '1 day' * $1 AND revoked = false
		ORDER BY expires_at ASC
	`
	rows, err := r.db.QueryContext(ctx, query, days)
	if err != nil {
		return nil, fmt.Errorf("get expiring: %w", err)
	}
	defer rows.Close()

	var keys []*domain.APIKey
	for rows.Next() {
		key := &domain.APIKey{}
		err := rows.Scan(&key.ID, &key.KeyHash, &key.UserID, &key.Name, &key.Tier, &key.RateLimit, &key.RequestsToday, &key.CreatedAt, &key.LastUsed, &key.ExpiresAt, &key.Revoked)
		if err != nil {
			return nil, fmt.Errorf("scan api key: %w", err)
		}
		keys = append(keys, key)
	}
	return keys, rows.Err()
}

// Count returns the total number of active API keys.
func (r *APIKeyRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM api_keys WHERE revoked = false`
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count api keys: %w", err)
	}
	return count, nil
}

// CountByTier returns the number of active keys in a tier.
func (r *APIKeyRepository) CountByTier(ctx context.Context, tier string) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM api_keys WHERE tier = $1 AND revoked = false`
	err := r.db.QueryRowContext(ctx, query, tier).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count by tier: %w", err)
	}
	return count, nil
}
