package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"

	"stockopname-rita-backend/internal/database"
	"stockopname-rita-backend/internal/middleware"
	"stockopname-rita-backend/internal/router"
)

const (
	// 3 minutes leaves room for PDF/Excel/DBF exports, which are slow until the
	// export path gets a row limit. Override with HTTP_WRITE_TIMEOUT.
	defaultWriteTimeout = 3 * time.Minute
	// 64 MiB fits both DBF files of a large store's master data.
	defaultMaxBodyBytes = 64 << 20
	shutdownTimeout     = 30 * time.Second
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	host := os.Getenv("APP_HOST")
	port := os.Getenv("APP_PORT")
	databaseURL := os.Getenv("DATABASE_URL")

	addr := host + ":" + port

	ctx := context.Background()
	validate := validator.New()

	pool, err := database.NewPostgresPool(ctx, databaseURL)
	if err != nil {
		log.Fatal("Failed to connect to the database:", err)
	}
	defer pool.Close()

	r := router.NewRouter(
		validate,
		pool,
		router.WithExportLimits(
			int(envInt64("REPORT_MAX_ROWS", router.DefaultMaxExportRows)),
			int(envInt64("REPORT_MAX_CONCURRENT", router.DefaultMaxConcurrentExport)),
		),
	)

	// Order matters: Recovery stays outermost so a panic anywhere is caught,
	// and CORS sits outside LimitBody so a 413 still carries CORS headers.
	m := middleware.Recovery(
		middleware.CORS(
			middleware.LimitBody(r, envInt64("HTTP_MAX_BODY_BYTES", defaultMaxBodyBytes)),
		),
	)

	srv := &http.Server{
		Addr:              addr,
		Handler:           m,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       60 * time.Second,
		WriteTimeout:      envDuration("HTTP_WRITE_TIMEOUT", defaultWriteTimeout),
		IdleTimeout:       2 * time.Minute,
		MaxHeaderBytes:    1 << 20,
	}

	// SIGTERM arrives on container stop / rolling deploy; Ctrl+C on Windows
	// sends os.Interrupt.
	stopCtx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	serveErr := make(chan error, 1)
	go func() {
		fmt.Println("Server is running on", addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serveErr <- err
		}
	}()

	select {
	case err := <-serveErr:
		log.Fatal("Server failed:", err)
	case <-stopCtx.Done():
		log.Println("Shutdown signal received, draining connections")
	}

	shutdownCtx, cancel := context.WithTimeout(ctx, shutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Println("Graceful shutdown timed out, closing connections:", err)
		_ = srv.Close()
	}
	log.Println("Server stopped")
}

// envDuration reads a Go duration ("90s", "3m"). 0 disables the timeout.
func envDuration(key string, fallback time.Duration) time.Duration {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	value, err := time.ParseDuration(raw)
	if err != nil || value < 0 {
		log.Printf("Ignoring invalid %s=%q, using %s", key, raw, fallback)
		return fallback
	}
	return value
}

// envInt64 reads a non-negative integer. 0 is accepted so the limit can be
// turned off.
func envInt64(key string, fallback int64) int64 {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || value < 0 {
		log.Printf("Ignoring invalid %s=%q, using %d", key, raw, fallback)
		return fallback
	}
	return value
}
