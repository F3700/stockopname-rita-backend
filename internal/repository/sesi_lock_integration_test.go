package repository

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

// TestSesiFindByIdForShareBlocksUpdate is a real two-connection lock test.
// It needs a migrated database: TEST_DATABASE_URL=... go test ./internal/repository -run TestSesiFindByIdForShareBlocksUpdate
func TestSesiFindByIdForShareBlocksUpdate(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}

	ctx := context.Background()
	conn1, err := pgx.Connect(ctx, dsn)
	if err != nil {
		t.Fatalf("connect conn1: %v", err)
	}
	defer conn1.Close(ctx)

	conn2, err := pgx.Connect(ctx, dsn)
	if err != nil {
		t.Fatalf("connect conn2: %v", err)
	}
	defer conn2.Close(ctx)

	code := fmt.Sprintf("LK-%012d", time.Now().UnixNano()%1_000_000_000_000)
	var sesiId int
	if err := conn1.QueryRow(ctx, `
		INSERT INTO sesi (sesi_location, sesi_code, sesi_status)
		VALUES ('Test lock', $1, 'IN_PROGRESS')
		RETURNING sesi_id
	`, code).Scan(&sesiId); err != nil {
		t.Fatalf("insert session: %v", err)
	}
	t.Cleanup(func() {
		_, _ = conn1.Exec(context.Background(), `DELETE FROM sesi WHERE sesi_id = $1`, sesiId)
	})

	tx1, err := conn1.Begin(ctx)
	if err != nil {
		t.Fatalf("begin tx1: %v", err)
	}
	defer tx1.Rollback(ctx)

	repo := NewSesiRepository(conn1)
	if _, err := repo.FindByIdForShare(ctx, tx1, sesiId); err != nil {
		t.Fatalf("lock session: %v", err)
	}

	tx2, err := conn2.Begin(ctx)
	if err != nil {
		t.Fatalf("begin tx2: %v", err)
	}
	defer tx2.Rollback(ctx)

	blockedCtx, cancel := context.WithTimeout(ctx, 300*time.Millisecond)
	defer cancel()
	_, err = tx2.Exec(blockedCtx, `
		UPDATE sesi
		SET sesi_status = 'COMPLETED', sesi_endedat = NOW()
		WHERE sesi_id = $1
	`, sesiId)
	if err == nil {
		t.Fatal("UPDATE was not blocked by FOR SHARE")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected lock wait to time out, got %v", err)
	}

	if err := tx1.Rollback(ctx); err != nil {
		t.Fatalf("release lock: %v", err)
	}

	if _, err := conn1.Exec(ctx, `
		UPDATE sesi
		SET sesi_status = 'COMPLETED', sesi_endedat = NOW()
		WHERE sesi_id = $1
	`, sesiId); err != nil {
		t.Fatalf("update after lock release: %v", err)
	}
}
