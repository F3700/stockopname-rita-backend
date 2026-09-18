package service

import (
	"context"
	"errors"
	"testing"

	"github.com/pashagolub/pgxmock/v4"

	"stockopname-rita-backend/internal/dto"
	"stockopname-rita-backend/internal/model"
)

func TestInspectorCreateByCoordinatorQRSuccess(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	t.Cleanup(func() {
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unfulfilled expectations: %v", err)
		}
	})

	mock.ExpectBegin()
	mock.ExpectQuery("WHERE coor_id").WithArgs(12).WillReturnRows(
		pgxmock.NewRows([]string{"id", "code", "sesi_id", "status"}).AddRow(12, "KOR-12", 1, "IN_PROGRESS"),
	)
	mock.ExpectQuery("INSERT INTO inspector").WithArgs("BUDI", 12).WillReturnRows(
		pgxmock.NewRows([]string{"inspector_id"}).AddRow(9),
	)
	mock.ExpectQuery("INSERT INTO rak").WithArgs("A-01", 9).WillReturnRows(
		pgxmock.NewRows([]string{"rak_id"}).AddRow(21),
	)
	mock.ExpectCommit()
	mock.ExpectRollback()

	svc := newInspectorServiceWithMock(mock, t)
	res, err := svc.CreateInspectorByCoordinatorQR(context.Background(), dto.InspectorJoinByCoordinatorQRRequest{
		CoordinatorQR: "RITA-COOR-12",
		InspectorCode: "BUDI",
		Rak:           []string{"A-01"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.InspectorID != 9 {
		t.Errorf("expected inspector id 9, got %d", res.InspectorID)
	}
	if len(res.Rak) != 1 || res.Rak[0].RackID != 21 {
		t.Errorf("unexpected racks %+v", res.Rak)
	}
}

func TestInspectorCreateByCoordinatorQRBadQR(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	svc := newInspectorServiceWithMock(mock, t)
	_, err = svc.CreateInspectorByCoordinatorQR(context.Background(), dto.InspectorJoinByCoordinatorQRRequest{
		CoordinatorQR: "BAD-QR",
		InspectorCode: "BUDI",
		Rak:           []string{"A-01"},
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var validation *model.ValidationError
	if !errors.As(err, &validation) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("no DB interaction expected, got: %v", err)
	}
}

func TestInspectorCreateByCoordinatorQRInactive(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	t.Cleanup(func() {
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unfulfilled expectations: %v", err)
		}
	})

	mock.ExpectBegin()
	mock.ExpectQuery("WHERE coor_id").WithArgs(12).WillReturnRows(
		pgxmock.NewRows([]string{"id", "code", "sesi_id", "status"}).AddRow(12, "KOR-12", 1, "COMPLETED"),
	)
	mock.ExpectRollback()

	svc := newInspectorServiceWithMock(mock, t)
	_, err = svc.CreateInspectorByCoordinatorQR(context.Background(), dto.InspectorJoinByCoordinatorQRRequest{
		CoordinatorQR: "RITA-COOR-12",
		InspectorCode: "BUDI",
		Rak:           []string{"A-01"},
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var conflict *model.ConflictError
	if !errors.As(err, &conflict) {
		t.Errorf("expected ConflictError, got %T: %v", err, err)
	}
}

func TestInspectorCreateByCoordinatorQRNotFound(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	t.Cleanup(func() {
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unfulfilled expectations: %v", err)
		}
	})

	mock.ExpectBegin()
	mock.ExpectQuery("WHERE coor_id").WithArgs(99).WillReturnRows(
		pgxmock.NewRows([]string{"id", "code", "sesi_id", "status"}),
	)
	mock.ExpectRollback()

	svc := newInspectorServiceWithMock(mock, t)
	_, err = svc.CreateInspectorByCoordinatorQR(context.Background(), dto.InspectorJoinByCoordinatorQRRequest{
		CoordinatorQR: "RITA-COOR-99",
		InspectorCode: "BUDI",
		Rak:           []string{"A-01"},
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var notFound *model.NotFoundError
	if !errors.As(err, &notFound) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestInspectorCreateByCoordinatorQRValidation(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	svc := newInspectorServiceWithMock(mock, t)
	_, err = svc.CreateInspectorByCoordinatorQR(context.Background(), dto.InspectorJoinByCoordinatorQRRequest{
		CoordinatorQR: "RITA-COOR-12",
		InspectorCode: "",
		Rak:           []string{"A-01"},
	})
	if err == nil {
		t.Fatal("expected error for empty inspector_code, got nil")
	}
}
