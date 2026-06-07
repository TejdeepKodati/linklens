package handlers

import (
	"context"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/tejdeep/linklens/internal/config"
	"github.com/tejdeep/linklens/internal/models"
	"github.com/tejdeep/linklens/internal/repository"
)

const (
	urlCacheTTL  = 10 * time.Minute
	codeAlphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	codeLength   = 7
)

type URLHandler struct {
	urlRepo   *repository.URLRepository
	clickRepo *repository.ClickRepository
	rdb       *redis.Client
	cfg       *config.Config
}

func NewURLHandler(urlRepo *repository.URLRepository, clickRepo *repository.ClickRepository, rdb *redis.Client, cfg *config.Config) *URLHandler {
	return &URLHandler{urlRepo: urlRepo, clickRepo: clickRepo, rdb: rdb, cfg: cfg}
}

// Create godoc — POST /api/urls
func (h *URLHandler) Create(c *gin.Context) {
	var req models.CreateURLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString("user_id")
	ctx := context.Background()

	// Determine short code
	code := strings.TrimSpace(req.CustomCode)
	if code == "" {
		var err error
		code, err = h.generateUniqueCode(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not generate short code"})
			return
		}
	} else {
		exists, _ := h.urlRepo.ShortCodeExists(ctx, code)
		if exists {
			c.JSON(http.StatusConflict, gin.H{"error": "custom code already taken"})
			return
		}
	}

	u := &models.URL{
		ID:          uuid.New().String(),
		UserID:      userID,
		OriginalURL: req.OriginalURL,
		ShortCode:   code,
		Title:       req.Title,
		ExpiresAt:   req.ExpiresAt,
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := h.urlRepo.Create(ctx, u); err != nil {
		if err == repository.ErrDuplicate {
			c.JSON(http.StatusConflict, gin.H{"error": "short code conflict, please try again"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create URL"})
		return
	}

	// Cache in Redis
	h.cacheURL(ctx, code, u.OriginalURL)

	c.JSON(http.StatusCreated, gin.H{
		"url":       u,
		"short_url": h.cfg.BaseURL + "/" + code,
	})
}

// List godoc — GET /api/urls
func (h *URLHandler) List(c *gin.Context) {
	userID := c.GetString("user_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	urls, total, err := h.urlRepo.ListByUser(context.Background(), userID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not list URLs"})
		return
	}

	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))
	c.JSON(http.StatusOK, models.PaginatedResponse{
		Data:       urls,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	})
}

// Get godoc — GET /api/urls/:id
func (h *URLHandler) Get(c *gin.Context) {
	userID := c.GetString("user_id")
	id := c.Param("id")

	u, err := h.urlRepo.GetByID(context.Background(), id, userID)
	if err != nil {
		if err == repository.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "URL not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not retrieve URL"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"url":       u,
		"short_url": h.cfg.BaseURL + "/" + u.ShortCode,
	})
}

// Update godoc — PUT /api/urls/:id
func (h *URLHandler) Update(c *gin.Context) {
	var req models.UpdateURLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString("user_id")
	id := c.Param("id")

	if err := h.urlRepo.Update(context.Background(), id, userID, &req); err != nil {
		if err == repository.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "URL not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not update URL"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "URL updated"})
}

// Delete godoc — DELETE /api/urls/:id
func (h *URLHandler) Delete(c *gin.Context) {
	userID := c.GetString("user_id")
	id := c.Param("id")

	if err := h.urlRepo.Delete(context.Background(), id, userID); err != nil {
		if err == repository.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "URL not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not delete URL"})
		return
	}

	// Invalidate cache (best-effort)
	c.JSON(http.StatusOK, gin.H{"message": "URL deleted"})
}

// Redirect godoc — GET /:code
// This is the hot path: Redis first, Postgres fallback.
func (h *URLHandler) Redirect(c *gin.Context) {
	code := c.Param("code")
	ctx := context.Background()

	// 1. Redis cache lookup (O(1))
	if dest := h.getCachedURL(ctx, code); dest != "" {
		go h.recordClick(code, c) // async click recording
		c.Redirect(http.StatusMovedPermanently, dest)
		return
	}

	// 2. Postgres fallback
	u, err := h.urlRepo.GetByShortCode(ctx, code)
	if err != nil {
		if err == repository.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "link not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server error"})
		return
	}

	if !u.IsActive {
		c.JSON(http.StatusGone, gin.H{"error": "link is inactive"})
		return
	}

	if u.ExpiresAt != nil && time.Now().After(*u.ExpiresAt) {
		c.JSON(http.StatusGone, gin.H{"error": "link has expired"})
		return
	}

	// Warm the cache
	h.cacheURL(ctx, code, u.OriginalURL)
	go h.recordClickByURLID(u.ID, c)

	c.Redirect(http.StatusMovedPermanently, u.OriginalURL)
}

// ─── Helpers ────────────────────────────────────────────────────────────────

func (h *URLHandler) generateUniqueCode(ctx context.Context) (string, error) {
	for i := 0; i < 10; i++ {
		code := randCode(codeLength)
		exists, err := h.urlRepo.ShortCodeExists(ctx, code)
		if err != nil {
			return "", err
		}
		if !exists {
			return code, nil
		}
	}
	return "", nil
}

func randCode(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = codeAlphabet[rand.Intn(len(codeAlphabet))]
	}
	return string(b)
}

func (h *URLHandler) cacheURL(ctx context.Context, code, dest string) {
	h.rdb.Set(ctx, "url:"+code, dest, urlCacheTTL)
}

func (h *URLHandler) getCachedURL(ctx context.Context, code string) string {
	v, err := h.rdb.Get(ctx, "url:"+code).Result()
	if err != nil {
		return ""
	}
	return v
}

func (h *URLHandler) recordClick(code string, c *gin.Context) {
	ctx := context.Background()
	u, err := h.urlRepo.GetByShortCode(ctx, code)
	if err != nil {
		return
	}
	h.recordClickByURLID(u.ID, c)
}

func (h *URLHandler) recordClickByURLID(urlID string, c *gin.Context) {
	ua := c.GetHeader("User-Agent")
	click := &models.Click{
		ID:         uuid.New().String(),
		URLID:      urlID,
		IPAddress:  c.ClientIP(),
		UserAgent:  ua,
		Referer:    c.GetHeader("Referer"),
		DeviceType: detectDevice(ua),
		Browser:    detectBrowser(ua),
		OS:         detectOS(ua),
		ClickedAt:  time.Now(),
	}
	_ = h.clickRepo.Record(context.Background(), click)
}

// Lightweight UA parsing without heavy dependencies
func detectDevice(ua string) string {
	ua = strings.ToLower(ua)
	if strings.Contains(ua, "mobile") || strings.Contains(ua, "android") || strings.Contains(ua, "iphone") {
		return "mobile"
	}
	if strings.Contains(ua, "tablet") || strings.Contains(ua, "ipad") {
		return "tablet"
	}
	return "desktop"
}

func detectBrowser(ua string) string {
	ua = strings.ToLower(ua)
	switch {
	case strings.Contains(ua, "edg"):
		return "Edge"
	case strings.Contains(ua, "chrome"):
		return "Chrome"
	case strings.Contains(ua, "firefox"):
		return "Firefox"
	case strings.Contains(ua, "safari"):
		return "Safari"
	case strings.Contains(ua, "opera"):
		return "Opera"
	default:
		return "Other"
	}
}

func detectOS(ua string) string {
	ua = strings.ToLower(ua)
	switch {
	case strings.Contains(ua, "windows"):
		return "Windows"
	case strings.Contains(ua, "mac os"):
		return "macOS"
	case strings.Contains(ua, "linux"):
		return "Linux"
	case strings.Contains(ua, "android"):
		return "Android"
	case strings.Contains(ua, "iphone") || strings.Contains(ua, "ipad"):
		return "iOS"
	default:
		return "Other"
	}
}
