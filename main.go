package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/tejdeep/linklens/internal/config"
	"github.com/tejdeep/linklens/internal/db"
	"github.com/tejdeep/linklens/internal/handlers"
	"github.com/tejdeep/linklens/internal/middleware"
	"github.com/tejdeep/linklens/internal/repository"
)

func main() {
	// Load .env file (silently skip if missing — prod uses real env vars)
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, reading from environment")
	}

	cfg := config.Load()

	// ── Database ───────────────────────────────────────────────────────────
	pgPool, err := db.NewPostgresDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("PostgreSQL: %v", err)
	}
	defer pgPool.Close()
	log.Println("✓ PostgreSQL connected")

	rdb := db.NewRedisClient(cfg.RedisURL)
	defer rdb.Close()
	log.Println("✓ Redis connected")

	// ── Repositories ──────────────────────────────────────────────────────
	userRepo := repository.NewUserRepository(pgPool)
	urlRepo := repository.NewURLRepository(pgPool)
	clickRepo := repository.NewClickRepository(pgPool)

	// ── Middleware ─────────────────────────────────────────────────────────
	authMW := middleware.NewAuthMiddleware(cfg)
	rateMW := middleware.NewRateLimiter(rdb)

	// ── Handlers ───────────────────────────────────────────────────────────
	authH := handlers.NewAuthHandler(userRepo, cfg)
	urlH := handlers.NewURLHandler(urlRepo, clickRepo, rdb, cfg)
	analyticsH := handlers.NewAnalyticsHandler(clickRepo, urlRepo)

	// ── Router ─────────────────────────────────────────────────────────────
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))

	// Static frontend
	r.Static("/static", "./web/static")
	r.StaticFile("/favicon.ico", "./web/static/favicon.ico")
	r.GET("/", func(c *gin.Context) {
		c.File("./web/static/index.html")
	})

	// ── API routes ─────────────────────────────────────────────────────────
	api := r.Group("/api")
	{
		// Health check
		api.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok", "timestamp": time.Now()})
		})

		// Auth (public)
		auth := api.Group("/auth")
		{
			auth.POST("/register", authH.Register)
			auth.POST("/login", authH.Login)
			auth.POST("/refresh", authH.Refresh)
		}

		// URLs (protected)
		urls := api.Group("/urls")
		urls.Use(authMW.Authenticate())
		urls.Use(rateMW.Limit("url_write", 100, time.Minute))
		{
			urls.POST("", urlH.Create)
			urls.GET("", urlH.List)
			urls.GET("/:id", urlH.Get)
			urls.PUT("/:id", urlH.Update)
			urls.DELETE("/:id", urlH.Delete)
		}

		// Analytics (protected)
		analytics := api.Group("/analytics")
		analytics.Use(authMW.Authenticate())
		{
			analytics.GET("/:id", analyticsH.GetClickStats)
			analytics.GET("/:id/timeline", analyticsH.GetClickTimeline)
		}
	}

	// Redirect short URL — must come AFTER /api to avoid conflicts
	r.GET("/:code", urlH.Redirect)

	// ── Graceful shutdown ──────────────────────────────────────────────────
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		log.Printf("🚀 LinkLens starting on port %s (env: %s)", cfg.Port, cfg.Env)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Forced shutdown:", err)
	}
	log.Println("Server stopped cleanly")
}
