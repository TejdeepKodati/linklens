package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tejdeep/linklens/internal/models"
)

type ClickRepository struct {
	db *pgxpool.Pool
}

func NewClickRepository(db *pgxpool.Pool) *ClickRepository {
	return &ClickRepository{db: db}
}

func (r *ClickRepository) Record(ctx context.Context, click *models.Click) error {
	query := `
		INSERT INTO clicks (id, url_id, ip_address, user_agent, referer, country, device_type, browser, os, clicked_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := r.db.Exec(ctx, query,
		click.ID, click.URLID, click.IPAddress, click.UserAgent,
		click.Referer, click.Country, click.DeviceType, click.Browser,
		click.OS, click.ClickedAt,
	)
	return err
}

// GetStats returns aggregated analytics for a URL.
func (r *ClickRepository) GetStats(ctx context.Context, urlID string) (*models.ClickStats, error) {
	stats := &models.ClickStats{}

	// Total clicks
	err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM clicks WHERE url_id = $1`, urlID,
	).Scan(&stats.TotalClicks)
	if err != nil {
		return nil, fmt.Errorf("total clicks: %w", err)
	}

	// Unique IPs
	err = r.db.QueryRow(ctx,
		`SELECT COUNT(DISTINCT ip_address) FROM clicks WHERE url_id = $1`, urlID,
	).Scan(&stats.UniqueClicks)
	if err != nil {
		return nil, fmt.Errorf("unique clicks: %w", err)
	}

	// By country
	stats.ByCountry, err = r.aggregateBy(ctx, urlID, "country")
	if err != nil {
		return nil, err
	}

	// By device
	stats.ByDevice, err = r.aggregateBy(ctx, urlID, "device_type")
	if err != nil {
		return nil, err
	}

	// By browser
	stats.ByBrowser, err = r.aggregateBy(ctx, urlID, "browser")
	if err != nil {
		return nil, err
	}

	// Top referers
	stats.TopReferers, err = r.aggregateBy(ctx, urlID, "referer")
	if err != nil {
		return nil, err
	}

	// Daily clicks (last 30 days)
	stats.DailyClicks, err = r.dailyClicks(ctx, urlID)
	if err != nil {
		return nil, err
	}

	return stats, nil
}

// GetTimeline returns daily click counts for charts.
func (r *ClickRepository) GetTimeline(ctx context.Context, urlID string, days int) ([]models.DailyClickStat, error) {
	return r.dailyClicks(ctx, urlID)
}

func (r *ClickRepository) aggregateBy(ctx context.Context, urlID, column string) ([]models.CountStat, error) {
	// Whitelist columns to prevent SQL injection
	allowed := map[string]bool{"country": true, "device_type": true, "browser": true, "referer": true}
	if !allowed[column] {
		return nil, fmt.Errorf("invalid column: %s", column)
	}

	query := fmt.Sprintf(`
		SELECT COALESCE(NULLIF(%s, ''), 'Unknown') as label, COUNT(*) as count
		FROM clicks WHERE url_id = $1
		GROUP BY %s ORDER BY count DESC LIMIT 10
	`, column, column)

	rows, err := r.db.Query(ctx, query, urlID)
	if err != nil {
		return nil, fmt.Errorf("aggregate by %s: %w", column, err)
	}
	defer rows.Close()

	var result []models.CountStat
	for rows.Next() {
		var s models.CountStat
		if err := rows.Scan(&s.Label, &s.Count); err != nil {
			return nil, err
		}
		result = append(result, s)
	}
	return result, nil
}

func (r *ClickRepository) dailyClicks(ctx context.Context, urlID string) ([]models.DailyClickStat, error) {
	query := `
		SELECT TO_CHAR(DATE_TRUNC('day', clicked_at), 'YYYY-MM-DD') as date,
		       COUNT(*) as count
		FROM clicks
		WHERE url_id = $1 AND clicked_at >= NOW() - INTERVAL '30 days'
		GROUP BY date ORDER BY date ASC
	`
	rows, err := r.db.Query(ctx, query, urlID)
	if err != nil {
		return nil, fmt.Errorf("daily clicks: %w", err)
	}
	defer rows.Close()

	var result []models.DailyClickStat
	for rows.Next() {
		var s models.DailyClickStat
		if err := rows.Scan(&s.Date, &s.Count); err != nil {
			return nil, err
		}
		result = append(result, s)
	}
	return result, nil
}

// isUniqueViolation checks for PostgreSQL unique constraint error code 23505.
func isUniqueViolation(err error) bool {
	return err != nil && strings.Contains(err.Error(), "23505")
}
