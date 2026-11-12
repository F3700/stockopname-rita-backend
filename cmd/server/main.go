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

	"stockopname-rita-backend/internal/auth"
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

	// one shift. The cookie is re-signed on every login, so a longer ttl only
	// means a shared laptop stays unlocked longer.
	defaultSessionTTL = 12 * time.Hour

	// The SPA's origin in production. Only used when CORS_ALLOWED_ORIGINS is
	// unset; dev goes through the Vite proxy, so it needs no CORS entry.
	defaultAllowedOrigin = "https://rita.albertt.my.id"
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

	// Built before the database connection: a missing credential is a
	// configuration mistake, and discovering it after a successful connect only
	// delays the same failure.
	username, passwordHash, err := authCredentials()
	if err != nil {
		log.Fatal(err)
	}

	sessionManager, err := auth.NewManager(
		[]byte(os.Getenv("SESSION_SECRET")),
		envDuration("SESSION_TTL", defaultSessionTTL),
	)
	if err != nil {
		log.Fatal("Invalid session configuration: ", err)
	}

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
		router.WithAuth(router.AuthConfig{
			Manager:      sessionManager,
			Username:     username,
			PasswordHash: passwordHash,
			CookieSecure: envBool("COOKIE_SECURE", true),
			CookieDomain: strings.TrimSpace(os.Getenv("COOKIE_DOMAIN")),
		}),
	)

	// Order matters: AccessLog stays outermost so a recovered panic is logged as
	// the 500 Recovery wrote, Recovery keeps a panic anywhere from killing the
	// process, CORS sits outside RequireAuth so a preflight still gets an answer,
	// and RequireAuth sits outside LimitBody so an anonymous request cannot make
	// the server read a large body.
	m := middleware.AccessLog(
		middleware.Recovery(
			middleware.CORS(
				envList("CORS_ALLOWED_ORIGINS", []string{defaultAllowedOrigin}),
				middleware.RequireAuth(
					sessionManager,
					middleware.LimitBody(r, envInt64("HTTP_MAX_BODY_BYTES", defaultMaxBodyBytes)),
				),
			),
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

// envBool reads a boolean flag. The fallback is used when the variable is
// missing or unparsable.
func envBool(key string, fallback bool) bool {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	value, err := strconv.ParseBool(raw)
	if err != nil {
		log.Printf("Ignoring invalid %s=%q, using %t", key, raw, fallback)
		return fallback
	}
	return value
}

// envList reads a comma separated list. Blank entries are dropped, and an
// entirely blank value falls back to the default.
func envList(key string, fallback []string) []string {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}

	values := make([]string, 0, strings.Count(raw, ",")+1)
	for _, part := range strings.Split(raw, ",") {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			values = append(values, trimmed)
		}
	}

	if len(values) == 0 {
		return fallback
	}

	return values
}

// authCredentials refuses to start without a usable login. A server that
// accepts every request is worse than one that does not come up: the failure
// would be silent and only visible to whoever reads the audit log.
func authCredentials() (string, string, error) {
	username := strings.TrimSpace(os.Getenv("AUTH_USERNAME"))
	passwordHash := strings.TrimSpace(os.Getenv("AUTH_PASSWORD_HASH"))

	if username == "" || passwordHash == "" {
		return "", "", errors.New(
			"AUTH_USERNAME and AUTH_PASSWORD_HASH must both be set; generate the hash with go run ./cmd/hashpw",
		)
	}

	if err := auth.ValidatePasswordHash(passwordHash); err != nil {
		return "", "", fmt.Errorf("AUTH_PASSWORD_HASH is unusable: %w", err)
	}

	return username, passwordHash, nil
}
