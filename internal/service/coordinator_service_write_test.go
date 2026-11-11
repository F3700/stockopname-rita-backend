package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/pashagolub/pgxmock/v4"

	"stockopname-rita-backend/internal/dto"
	"stockopname-rita-backend/internal/model"
	"stockopname-rita-backend/internal/repository"
)

func newCoordinatorServiceWithMock(mock pgxmock.PgxPoolIface, t *testing.T) CoordinatorService {
	t.Helper()
	return NewCoordinatorService(repository.NewCoordinatorRepository(mock), repository.NewSesiRepository(mock), repository.NewProductRepository(mock), repository.NewRackRepository(mock), mock)
}

func coordinatorRow(id int, sesiId int) *pgxmock.Rows {
	return pgxmock.NewRows([]string{"id", "code", "sesi_id", "status"}).
		AddRow(id, "KOR-01", sesiId, "IN_PROGRESS")
}

func sesiStatusRow(id int, status string) *pgxmock.Rows {
	started := time.Date(2026, 9, 10, 8, 0, 0, 0, time.UTC)
	return pgxmock.NewRows([]string{"sesi_id", "sesi_location", "sesi_code", "sesi_status", "sesi_startedat", "sesi_endedat"}).
		AddRow(id, "Gudang A", "SESI-01", status, started, nil)
}

func TestCoordinatorCreateTx(t *testing.T) {
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
	mock.ExpectQuery("FOR UPDATE").WithArgs(3).WillReturnRows(sesiStatusRow(3, "IN_PROGRESS"))
	mock.ExpectQuery("INSERT INTO coordinator").WithArgs("KOR-03", 3, "IN_PROGRESS").
		WillReturnRows(pgxmock.NewRows([]string{"coor_id"}).AddRow(7))
	mock.ExpectQuery("INSERT INTO rak").WithArgs("Rak 1", 7).
		WillReturnRows(pgxmock.NewRows([]string{"rak_id"}).AddRow(11))
	mock.ExpectQuery("INSERT INTO rak").WithArgs("Rak 2", 7).
		WillReturnRows(pgxmock.NewRows([]string{"rak_id"}).AddRow(12))
	mock.ExpectCommit()
	mock.ExpectRollback()

	svc := newCoordinatorServiceWithMock(mock, t)
	res, err := svc.Create(context.Background(), dto.CreateCoordinatorRequest{
		SessionID: 3,
		Code:      "KOR-03",
		RackStart: 1,
		RackEnd:   2,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.ID != 7 || res.Code != "KOR-03" || res.RackAssigned != 2 || res.Status != "IN_PROGRESS" {
		t.Errorf("unexpected response %+v", res)
	}
}

func TestCoordinatorCreateRejectsOversizedRange(t *testing.T) {
	svc := newCoordinatorServiceWithMock(nil, t)
	_, err := svc.Create(context.Background(), dto.CreateCoordinatorRequest{
		SessionID: 3,
		Code:      "KOR-03",
		RackStart: 1,
		RackEnd:   201,
	})
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
	var validation *model.ValidationError
	if !errors.As(err, &validation) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestCoordinatorUpdateTx(t *testing.T) {
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
	mock.ExpectQuery("FROM coordinator").WithArgs(2).WillReturnRows(coordinatorRow(2, 1))
	mock.ExpectQuery("FOR SHARE").WithArgs(1).WillReturnRows(sesiStatusRow(1, "IN_PROGRESS"))
	mock.ExpectExec("UPDATE coordinator").WithArgs("COMPLETED", 2).WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	mock.ExpectCommit()
	mock.ExpectRollback()

	svc := newCoordinatorServiceWithMock(mock, t)
	if err := svc.Update(context.Background(), 2, &dto.UpdateCoordinatorRequest{Status: "COMPLETED"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCoordinatorUpdateTxNotFound(t *testing.T) {
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
	mock.ExpectQuery("FROM coordinator").WithArgs(99).WillReturnRows(
		pgxmock.NewRows([]string{"id", "code", "sesi_id", "status"}),
	)
	mock.ExpectRollback()

	svc := newCoordinatorServiceWithMock(mock, t)
	err = svc.Update(context.Background(), 99, &dto.UpdateCoordinatorRequest{Status: "COMPLETED"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var notFound *model.NotFoundError
	if !errors.As(err, &notFound) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestCoordinatorUpdateTxFinalSession(t *testing.T) {
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
	mock.ExpectQuery("FROM coordinator").WithArgs(2).WillReturnRows(coordinatorRow(2, 1))
	mock.ExpectQuery("FOR SHARE").WithArgs(1).WillReturnRows(sesiStatusRow(1, "COMPLETED"))
	mock.ExpectRollback()

	svc := newCoordinatorServiceWithMock(mock, t)
	err = svc.Update(context.Background(), 2, &dto.UpdateCoordinatorRequest{Status: "CANCELLED"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var conflict *model.ConflictError
	if !errors.As(err, &conflict) {
		t.Errorf("expected ConflictError, got %T: %v", err, err)
	}
}

func TestCoordinatorJoinByQR(t *testing.T) {
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
	mock.ExpectQuery("FROM coordinator").WithArgs(12).WillReturnRows(coordinatorRow(12, 4))
	mock.ExpectQuery("FROM sesi").WithArgs(4).WillReturnRows(sesiStatusRow(4, "IN_PROGRESS"))
	mock.ExpectQuery("COUNT").WithArgs(4).WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery("FROM rak").WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).WillReturnRows(
		pgxmock.NewRows([]string{"RakID", "RakName", "CoorID", "UploadedAt", "ItemCount", "UnownedCount", "CanUpload"}).
			AddRow(31, "Rak 1", 12, nil, 0, 0, true).
			AddRow(32, "Rak 2", 12, time.Now(), 2, 0, false),
	)
	mock.ExpectRollback()

	svc := newCoordinatorServiceWithMock(mock, t)
	res, err := svc.JoinByCoordinatorQR(context.Background(), dto.JoinByCoordinatorQRRequest{CoordinatorQR: "RITA-COOR-12"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.CoordinatorID != 12 || res.SesiID != 4 || len(res.Rak) != 2 {
		t.Fatalf("unexpected response %+v", res)
	}
	if res.Rak[0].UploadedAt != "" {
		t.Error("expected first rack to be not uploaded")
	}
	if res.Rak[1].UploadedAt == "" {
		t.Error("expected second rack to be uploaded")
	}
	if res.Rak[0].ItemCount != 0 || !res.Rak[0].IsEmpty || res.Rak[1].ItemCount != 2 || res.Rak[1].IsEmpty {
		t.Errorf("unexpected live rack counts %+v", res.Rak)
	}
	if !res.Rak[0].CanUpload || res.Rak[1].CanUpload {
		t.Errorf("unexpected can_upload flags %+v", res.Rak)
	}
}

func TestCoordinatorJoinInactive(t *testing.T) {
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
	mock.ExpectQuery("FROM coordinator").WithArgs(12).WillReturnRows(
		pgxmock.NewRows([]string{"id", "code", "sesi_id", "status"}).AddRow(12, "KOR-12", 4, "IN_REVIEW"),
	)
	mock.ExpectQuery("FROM sesi").WithArgs(4).WillReturnRows(sesiStatusRow(4, "IN_PROGRESS"))
	mock.ExpectRollback()

	svc := newCoordinatorServiceWithMock(mock, t)
	_, err = svc.JoinByCoordinatorQR(context.Background(), dto.JoinByCoordinatorQRRequest{CoordinatorQR: "RITA-COOR-12"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var conflict *model.ConflictError
	if !errors.As(err, &conflict) {
		t.Errorf("expected ConflictError, got %T: %v", err, err)
	}
}

func TestCoordinatorJoinInvalidQR(t *testing.T) {
	svc := newCoordinatorServiceWithMock(nil, t)
	_, err := svc.JoinByCoordinatorQR(context.Background(), dto.JoinByCoordinatorQRRequest{CoordinatorQR: "not-a-qr"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
