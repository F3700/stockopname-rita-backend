package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pashagolub/pgxmock/v4"

	"stockopname-rita-backend/internal/model"
)

func TestRackFindAll(t *testing.T) {
	mock := mustMockPool(t)
	inspectorId, coordinatorId, sessionId := 1, 2, 3
	mock.ExpectQuery("FROM rak").WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).WillReturnRows(
		pgxmock.NewRows([]string{"RackID", "RackName", "InspectorID"}).
			AddRow(1, "R1", 1).
			AddRow(2, "R2", 1),
	)

	repo := NewRackRepository(mock)
	results, err := repo.FindAll(context.Background(), &inspectorId, &coordinatorId, &sessionId)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 || results[0].RackName != "R1" || results[1].RackID != 2 {
		t.Errorf("unexpected results %+v", results)
	}
}

func TestRackFindAllQueryError(t *testing.T) {
	mock := mustMockPool(t)
	mock.ExpectQuery("FROM rak").WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).WillReturnError(errors.New("boom"))

	repo := NewRackRepository(mock)
	_, err := repo.FindAll(context.Background(), nil, nil, nil)
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestRackFindProgress(t *testing.T) {
	mock := mustMockPool(t)
	coorId, sesiId := 2, 3
	mock.ExpectQuery("FROM rak").WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).WillReturnRows(
		pgxmock.NewRows([]string{"rack_assigned", "rack_completed", "total_items"}).
			AddRow(5, 3, 42),
	)

	repo := NewRackRepository(mock)
	results, err := repo.FindProgress(context.Background(), &coorId, &sesiId)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	got := results[0]
	if got.RackAssigned != 5 || got.RackCompleted != 3 || got.TotalItems != 42 {
		t.Errorf("unexpected result %+v", got)
	}
}

func TestRackSave(t *testing.T) {
	mock := mustMockPool(t)
	tx := mustBeginTx(t, mock)
	mock.ExpectQuery("INSERT INTO rak").WithArgs("R9", 3).WillReturnRows(
		pgxmock.NewRows([]string{"rak_id"}).AddRow(9),
	)

	repo := NewRackRepository(mock)
	rack := &model.Rack{RackName: "R9", InspectorID: 3}
	if err := repo.Save(context.Background(), tx, rack); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rack.RackID != 9 {
		t.Errorf("expected id 9, got %d", rack.RackID)
	}
}

func TestRackSaveConflict(t *testing.T) {
	mock := mustMockPool(t)
	tx := mustBeginTx(t, mock)
	mock.ExpectQuery("INSERT INTO rak").WithArgs("R9", 3).WillReturnError(
		&pgconn.PgError{Code: "23505", ConstraintName: "uq_rak_inspector_name"},
	)

	repo := NewRackRepository(mock)
	err := repo.Save(context.Background(), tx, &model.Rack{RackName: "R9", InspectorID: 3})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var conflict *model.ConflictError
	if !errors.As(err, &conflict) {
		t.Errorf("expected ConflictError, got %T: %v", err, err)
	}
}
