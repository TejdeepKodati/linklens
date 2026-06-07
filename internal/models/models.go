package models

import "time"

// ─────────────────────────────────────────────
//  User
// ─────────────────────────────────────────────

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	Name         string    `json:"name"`
	PasswordHash string    `json:"-"` // never serialized to JSON
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type RegisterRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Name     string `json:"name"     binding:"required,min=2"`
	Password string `json:"password" binding:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	User         *User  `json:"user"`
}

// ─────────────────────────────────────────────
//  URL
// ─────────────────────────────────────────────

type URL struct {
	ID          string     `json:"id"`
	UserID      string     `json:"user_id"`
	OriginalURL string     `json:"original_url"`
	ShortCode   string     `json:"short_code"`
	Title       string     `json:"title"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	IsActive    bool       `json:"is_active"`
	ClickCount  int64      `json:"click_count"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type CreateURLRequest struct {
	OriginalURL string     `json:"original_url" binding:"required,url"`
	CustomCode  string     `json:"custom_code,omitempty"`
	Title       string     `json:"title,omitempty"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}

type UpdateURLRequest struct {
	Title     string     `json:"title,omitempty"`
	IsActive  *bool      `json:"is_active,omitempty"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

// ─────────────────────────────────────────────
//  Click
// ─────────────────────────────────────────────

type Click struct {
	ID         string    `json:"id"`
	URLID      string    `json:"url_id"`
	IPAddress  string    `json:"ip_address"`
	UserAgent  string    `json:"user_agent"`
	Referer    string    `json:"referer"`
	Country    string    `json:"country"`
	DeviceType string    `json:"device_type"`
	Browser    string    `json:"browser"`
	OS         string    `json:"os"`
	ClickedAt  time.Time `json:"clicked_at"`
}

type ClickStats struct {
	TotalClicks   int64            `json:"total_clicks"`
	UniqueClicks  int64            `json:"unique_clicks"`
	ByCountry     []CountStat      `json:"by_country"`
	ByDevice      []CountStat      `json:"by_device"`
	ByBrowser     []CountStat      `json:"by_browser"`
	TopReferers   []CountStat      `json:"top_referers"`
	DailyClicks   []DailyClickStat `json:"daily_clicks"`
}

type CountStat struct {
	Label string `json:"label"`
	Count int64  `json:"count"`
}

type DailyClickStat struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

// ─────────────────────────────────────────────
//  Generic responses
// ─────────────────────────────────────────────

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalPages int         `json:"total_pages"`
}
