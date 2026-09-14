package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pashagolub/pgxmock/v4"

	"stockopname-rita-backend/internal/model"
)

var inspectorSummaryColumns = []string{"inspector_id", "inspector_code", "rack_assigned", "rack_completed", "total_items"}

func TestInspectorFindAll(t *testing.T) {
	mock := mustMockPool(t)
	mock.ExpectQuery("FROM inspector").WillReturnRows(
		pgxmock.NewRows(inspectorSummaryColumns).
			AddRow(1, "INSP-01", 3, 2, 40).
			AddRow(2, "INSP-02", 1, 1, 10),
	)

	repo := NewInspectorRepository(mock)
	results, err := repo.FindAll(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 || results[0].InspectorCode != "INSP-01" || results[1].TotalItems != 10 {
		t.Errorf("unexpected results %+v", results)
	}
}

func TestInspectorFindAllQueryError(t *testing.T) {
	mock := mustMockPool(t)
	mock.ExpectQuery("FROM inspector").WillReturnError(errors.New("boom"))

	repo := NewInspectorRepository(mock)
	_, err := repo.FindAll(context.Background())
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestInspectorFindByCoorId(t *testing.T) {
	mock := mustMockPool(t)
	mock.ExpectQuery("inspector_coor_id").WithArgs(5).WillReturnRows(
		pgxmock.NewRows(inspectorSummaryColumns).
			AddRow(3, "INSP-03", 2, 0, 0),
	)

	repo := NewInspectorRepository(mock)
	results, err := repo.FindByCoorId(context.Background(), 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 || results[0].InspectorID != 3 {
		t.Errorf("unexpected results %+v", results)
	}
}

func TestInspectorSave(t *testing.T) {
	mock := mustMockPool(t)
	tx := mustBeginTx(t, mock)
	mock.ExpectQuery("INSERT INTO inspector").WithArgs("INSP-09", 5).WillReturnRows(
		pgxmock.NewRows([]string{"inspector_id"}).AddRow(9),
	)

	repo := NewInspectorRepository(mock)
	inspector := &model.Inspector{InspectorCode: "INSP-09", CoordinatorID: 5}
	if err := repo.Save(context.Background(), tx, inspector); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inspector.InspectorID != 9 {
		t.Errorf("expected id 9, got %d", inspector.InspectorID)
	}
}

func TestInspectorSaveConflict(t *testing.T) {
	mock := mustMockPool(t)
	tx := mustBeginTx(t, mock)
	mock.ExpectQuery("INSERT INTO inspector").WithArgs("INSP-09", 5).WillReturnError(
		&pgconn.PgError{Code: "23505", ConstraintName: "uq_inspector_coor_code"},
	)

	repo := NewInspectorRepository(mock)
	err := repo.Save(context.Background(), tx, &model.Inspector{InspectorCode: "INSP-09", CoordinatorID: 5})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var conflict *model.ConflictError
	if !errors.As(err, &conflict) {
		t.Errorf("expected ConflictError, got %T: %v", err, err)
	}
}
