package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tejdeep/linklens/internal/models"
)

type URLRepository struct {
	db *pgxpool.Pool
}

func NewURLRepository(db *pgxpool.Pool) *URLRepository {
	return &URLRepository{db: db}
}

func (r *URLRepository) Create(ctx context.Context, u *models.URL) error {
	query := `
		INSERT INTO urls (id, user_id, original_url, short_code, title, expires_at, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.db.Exec(ctx, query,
		u.ID, u.UserID, u.OriginalURL, u.ShortCode,
		u.Title, u.ExpiresAt, u.IsActive, u.CreatedAt, u.UpdatedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return ErrDuplicate
		}
		return fmt.Errorf("create url: %w", err)
	}
	return nil
}

func (r *URLRepository) GetByShortCode(ctx context.Context, code string) (*models.URL, error) {
	query := `
		SELECT id, user_id, original_url, short_code, title, expires_at, is_active,
		       (SELECT COUNT(*) FROM clicks WHERE url_id = urls.id) as click_count,
		       created_at, updated_at
		FROM urls WHERE short_code = $1
	`
	return r.scanURL(r.db.QueryRow(ctx, query, code))
}

func (r *URLRepository) GetByID(ctx context.Context, id, userID string) (*models.URL, error) {
	query := `
		SELECT id, user_id, original_url, short_code, title, expires_at, is_active,
		       (SELECT COUNT(*) FROM clicks WHERE url_id = urls.id) as click_count,
		       created_at, updated_at
		FROM urls WHERE id = $1 AND user_id = $2
	`
	return r.scanURL(r.db.QueryRow(ctx, query, id, userID))
}

func (r *URLRepository) ListByUser(ctx context.Context, userID string, page, pageSize int) ([]*models.URL, int64, error) {
	offset := (page - 1) * pageSize

	// Total count
	var total int64
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM urls WHERE user_id = $1`, userID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count urls: %w", err)
	}

	query := `
		SELECT id, user_id, original_url, short_code, title, expires_at, is_active,
		       (SELECT COUNT(*) FROM clicks WHERE url_id = urls.id) as click_count,
		       created_at, updated_at
		FROM urls WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(ctx, query, userID, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list urls: %w", err)
	}
	defer rows.Close()

	var urls []*models.URL
	for rows.Next() {
		u := &models.URL{}
		err := rows.Scan(
			&u.ID, &u.UserID, &u.OriginalURL, &u.ShortCode,
			&u.Title, &u.ExpiresAt, &u.IsActive, &u.ClickCount,
			&u.CreatedAt, &u.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("scan url row: %w", err)
		}
		urls = append(urls, u)
	}
	return urls, total, nil
}

func (r *URLRepository) Update(ctx context.Context, id, userID string, req *models.UpdateURLRequest) error {
	query := `UPDATE urls SET title = COALESCE(NULLIF($1,''), title), updated_at = NOW()`
	args := []interface{}{req.Title}
	argIdx := 2

	if req.IsActive != nil {
		query += fmt.Sprintf(", is_active = $%d", argIdx)
		args = append(args, *req.IsActive)
		argIdx++
	}
	if req.ExpiresAt != nil {
		query += fmt.Sprintf(", expires_at = $%d", argIdx)
		args = append(args, req.ExpiresAt)
		argIdx++
	}

	query += fmt.Sprintf(" WHERE id = $%d AND user_id = $%d", argIdx, argIdx+1)
	args = append(args, id, userID)

	tag, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("update url: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *URLRepository) Delete(ctx context.Context, id, userID string) error {
	tag, err := r.db.Exec(ctx,
		`DELETE FROM urls WHERE id = $1 AND user_id = $2`, id, userID,
	)
	if err != nil {
		return fmt.Errorf("delete url: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *URLRepository) ShortCodeExists(ctx context.Context, code string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM urls WHERE short_code = $1)`, code,
	).Scan(&exists)
	return exists, err
}

func (r *URLRepository) scanURL(row pgx.Row) (*models.URL, error) {
	u := &models.URL{}
	err := row.Scan(
		&u.ID, &u.UserID, &u.OriginalURL, &u.ShortCode,
		&u.Title, &u.ExpiresAt, &u.IsActive, &u.ClickCount,
		&u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("scan url: %w", err)
	}
	return u, nil
}
