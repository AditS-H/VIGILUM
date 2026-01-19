// Package repositories implements data persistence for all entities.
package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/vigilum/backend/internal/domain"
)

// ============================================================
// USER REPOSITORY
// ============================================================

// UserRepository implements domain.UserRepository using PostgreSQL.
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new user repository.
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create inserts a new user.
func (r *UserRepository) Create(ctx context.Context, wallet string) (*domain.User, error) {
	id := uuid.New().String()
	user := &domain.User{
		ID:            id,
		WalletAddress: wallet,
		CreatedAt:     time.Now(),
		RiskScore:     0.0,
	}

	query := `
		INSERT INTO users (id, wallet_address, created_at, risk_score)
		VALUES ($1, $2, $3, $4)
		RETURNING id, wallet_address, created_at, risk_score, is_blacklisted
	`
	err := r.db.QueryRowContext(ctx, query, user.ID, wallet, user.CreatedAt, 0.0).
		Scan(&user.ID, &user.WalletAddress, &user.CreatedAt, &user.RiskScore, &user.IsBlacklisted)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	return user, nil
}

// GetByWallet retrieves a user by wallet address.
func (r *UserRepository) GetByWallet(ctx context.Context, wallet string) (*domain.User, error) {
	query := `
		SELECT id, wallet_address, created_at, last_activity, risk_score, is_blacklisted
		FROM users
		WHERE wallet_address = $1
	`
	user := &domain.User{}
	err := r.db.QueryRowContext(ctx, query, wallet).
		Scan(&user.ID, &user.WalletAddress, &user.CreatedAt, &user.LastActivity, &user.RiskScore, &user.IsBlacklisted)
	if err == sql.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get user by wallet: %w", err)
	}
	return user, nil
}

// GetByID retrieves a user by ID.
func (r *UserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	query := `
		SELECT id, wallet_address, created_at, last_activity, risk_score, is_blacklisted
		FROM users
		WHERE id = $1
	`
	user := &domain.User{}
	err := r.db.QueryRowContext(ctx, query, id).
		Scan(&user.ID, &user.WalletAddress, &user.CreatedAt, &user.LastActivity, &user.RiskScore, &user.IsBlacklisted)
	if err == sql.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return user, nil
}

// Update modifies an existing user.
func (r *UserRepository) Update(ctx context.Context, id string, u *domain.User) error {
	query := `
		UPDATE users
		SET wallet_address = $1, last_activity = $2, risk_score = $3, is_blacklisted = $4
		WHERE id = $5
	`
	result, err := r.db.ExecContext(ctx, query, u.WalletAddress, u.LastActivity, u.RiskScore, u.IsBlacklisted, id)
	if err != nil {
		return fmt.Errorf("update user: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("update user rows affected: %w", err)
	}
	if rows == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// UpdateRiskScore updates the risk score for a user.
func (r *UserRepository) UpdateRiskScore(ctx context.Context, id string, score float64) error {
	query := `UPDATE users SET risk_score = $1 WHERE id = $2`
	result, err := r.db.ExecContext(ctx, query, score, id)
	if err != nil {
		return fmt.Errorf("update risk score: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// UpdateLastActivity updates the last activity timestamp.
func (r *UserRepository) UpdateLastActivity(ctx context.Context, id string) error {
	query := `UPDATE users SET last_activity = NOW() WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("update last activity: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// Blacklist marks a user as blacklisted.
func (r *UserRepository) Blacklist(ctx context.Context, id string) error {
	query := `UPDATE users SET is_blacklisted = true WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("blacklist user: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// RemoveBlacklist removes a user from blacklist.
func (r *UserRepository) RemoveBlacklist(ctx context.Context, id string) error {
	query := `UPDATE users SET is_blacklisted = false WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("remove blacklist: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// Delete removes a user.
func (r *UserRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM users WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// ListByRiskScore returns users above a risk threshold.
func (r *UserRepository) ListByRiskScore(ctx context.Context, threshold float64, limit int) ([]*domain.User, error) {
	query := `
		SELECT id, wallet_address, created_at, last_activity, risk_score, is_blacklisted
		FROM users
		WHERE risk_score >= $1
		ORDER BY risk_score DESC
		LIMIT $2
	`
	rows, err := r.db.QueryContext(ctx, query, threshold, limit)
	if err != nil {
		return nil, fmt.Errorf("list by risk score: %w", err)
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		user := &domain.User{}
		err := rows.Scan(&user.ID, &user.WalletAddress, &user.CreatedAt, &user.LastActivity, &user.RiskScore, &user.IsBlacklisted)
		if err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, user)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return users, nil
}

// ListBlacklisted returns all blacklisted users.
func (r *UserRepository) ListBlacklisted(ctx context.Context, limit int, offset int) ([]*domain.User, error) {
	query := `
		SELECT id, wallet_address, created_at, last_activity, risk_score, is_blacklisted
		FROM users
		WHERE is_blacklisted = true
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list blacklisted: %w", err)
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		user := &domain.User{}
		err := rows.Scan(&user.ID, &user.WalletAddress, &user.CreatedAt, &user.LastActivity, &user.RiskScore, &user.IsBlacklisted)
		if err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, user)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return users, nil
}

// Count returns the total number of users.
func (r *UserRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM users`
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count users: %w", err)
	}
	return count, nil
}
