package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tejdeep/linklens/internal/repository"
)

type AnalyticsHandler struct {
	clickRepo *repository.ClickRepository
	urlRepo   *repository.URLRepository
}

func NewAnalyticsHandler(clickRepo *repository.ClickRepository, urlRepo *repository.URLRepository) *AnalyticsHandler {
	return &AnalyticsHandler{clickRepo: clickRepo, urlRepo: urlRepo}
}

// GetClickStats godoc — GET /api/analytics/:id
func (h *AnalyticsHandler) GetClickStats(c *gin.Context) {
	id := c.Param("id")
	userID := c.GetString("user_id")

	// Verify ownership
	_, err := h.urlRepo.GetByID(context.Background(), id, userID)
	if err != nil {
		if err == repository.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "URL not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server error"})
		return
	}

	stats, err := h.clickRepo.GetStats(context.Background(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not retrieve stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetClickTimeline godoc — GET /api/analytics/:id/timeline
func (h *AnalyticsHandler) GetClickTimeline(c *gin.Context) {
	id := c.Param("id")
	userID := c.GetString("user_id")

	_, err := h.urlRepo.GetByID(context.Background(), id, userID)
	if err != nil {
		if err == repository.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "URL not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server error"})
		return
	}

	timeline, err := h.clickRepo.GetTimeline(context.Background(), id, 30)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not retrieve timeline"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"timeline": timeline})
}
