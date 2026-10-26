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

func newUploadServiceWithMock(mock pgxmock.PgxPoolIface, t *testing.T) UploadService {
	t.Helper()
	return NewUploadService(
		repository.NewUploadRepository(mock),
		repository.NewRackRepository(mock),
		repository.NewCoordinatorRepository(mock),
		repository.NewSesiRepository(mock),
		repository.NewProductRepository(mock),
		repository.NewStockOpnameRepository(mock),
		mock,
		newTestValidator(t),
	)
}

func uploadRequest() dto.CreateUploadRequest {
	return dto.CreateUploadRequest{
		CoordinatorID: 7,
		Racks: []dto.UploadRackRequest{
			{RakID: 31, Items: []dto.UploadItemRequest{{Quantity: intPtr(10), Barcode: "8991001010016"}}},
			{RakID: 32, Items: []dto.UploadItemRequest{{Quantity: intPtr(11), Barcode: "8991001010016"}}},
		},
	}
}

func TestUploadCreateBatch(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	t.Cleanup(func() {
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unfulfilled expectations: %v", err)
		}
	})

	createdAt := time.Date(2026, 10, 26, 10, 14, 0, 0, time.UTC)
	mock.ExpectBegin()
	mock.ExpectQuery("FROM coordinator").WithArgs(7).WillReturnRows(coordinatorRows(7, 1))
	mock.ExpectQuery("FOR SHARE").WithArgs(1).WillReturnRows(sesiStatusRow(1, "IN_PROGRESS"))
	mock.ExpectQuery("COUNT").WithArgs(1).WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery("JOIN barcode").WithArgs(1, "8991001010016").WillReturnRows(soResolveProductRows())
	mock.ExpectQuery("FROM barcode").WithArgs(pgxmock.AnyArg()).WillReturnRows(soResolveBarcodeRows())
	mock.ExpectQuery("JOIN barcode").WithArgs(1, "8991001010016").WillReturnRows(soResolveProductRows())
	mock.ExpectQuery("FROM barcode").WithArgs(pgxmock.AnyArg()).WillReturnRows(soResolveBarcodeRows())
	mock.ExpectQuery("INSERT INTO upload").WithArgs(7).WillReturnRows(
		pgxmock.NewRows([]string{"upload_id", "upload_createdat"}).AddRow(3, createdAt),
	)
	mock.ExpectExec("UPDATE rak").WithArgs(3, []int32{31, 32}, 7).WillReturnResult(pgxmock.NewResult("UPDATE", 2))
	mock.ExpectExec("UPDATE rak").WithArgs(31, pgxmock.AnyArg()).WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	mock.ExpectQuery("ON CONFLICT").WithArgs(10, 31, "100251", "Sari Roti", "8991001010016", 14500.0, 17200.0).WillReturnRows(
		pgxmock.NewRows([]string{"stock_opname_id"}).AddRow(1),
	)
	mock.ExpectExec("UPDATE rak").WithArgs(32, pgxmock.AnyArg()).WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	mock.ExpectQuery("ON CONFLICT").WithArgs(11, 32, "100251", "Sari Roti", "8991001010016", 14500.0, 17200.0).WillReturnRows(
		pgxmock.NewRows([]string{"stock_opname_id"}).AddRow(2),
	)
	mock.ExpectCommit()

	svc := newUploadServiceWithMock(mock, t)
	res, err := svc.Create(context.Background(), uploadRequest())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.UploadID != 3 || res.RackCount != 2 || res.ItemCount != 2 {
		t.Errorf("unexpected response %+v", res)
	}
}

func TestUploadCreateAlreadyUploaded(t *testing.T) {
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
	mock.ExpectQuery("FROM coordinator").WithArgs(7).WillReturnRows(coordinatorRows(7, 1))
	mock.ExpectQuery("FOR SHARE").WithArgs(1).WillReturnRows(sesiStatusRow(1, "IN_PROGRESS"))
	mock.ExpectQuery("COUNT").WithArgs(1).WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery("JOIN barcode").WithArgs(1, "8991001010016").WillReturnRows(soResolveProductRows())
	mock.ExpectQuery("FROM barcode").WithArgs(pgxmock.AnyArg()).WillReturnRows(soResolveBarcodeRows())
	mock.ExpectQuery("JOIN barcode").WithArgs(1, "8991001010016").WillReturnRows(soResolveProductRows())
	mock.ExpectQuery("FROM barcode").WithArgs(pgxmock.AnyArg()).WillReturnRows(soResolveBarcodeRows())
	mock.ExpectQuery("INSERT INTO upload").WithArgs(7).WillReturnRows(
		pgxmock.NewRows([]string{"upload_id", "upload_createdat"}).AddRow(3, time.Now()),
	)
	mock.ExpectExec("UPDATE rak").WithArgs(3, []int32{31, 32}, 7).WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	mock.ExpectRollback()

	svc := newUploadServiceWithMock(mock, t)
	_, err = svc.Create(context.Background(), uploadRequest())
	if err == nil {
		t.Fatal("expected conflict, got nil")
	}
	var conflict *model.ConflictError
	if !errors.As(err, &conflict) {
		t.Errorf("expected ConflictError, got %T: %v", err, err)
	}
}

func TestUploadCreateDuplicateRack(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	t.Cleanup(func() {
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unfulfilled expectations: %v", err)
		}
	})

	req := uploadRequest()
	req.Racks = append(req.Racks, req.Racks[0])

	svc := newUploadServiceWithMock(mock, t)
	_, err = svc.Create(context.Background(), req)
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
	var validation *model.ValidationError
	if !errors.As(err, &validation) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestUploadCreateFinalSession(t *testing.T) {
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
	mock.ExpectQuery("FROM coordinator").WithArgs(7).WillReturnRows(coordinatorRows(7, 1))
	mock.ExpectQuery("FOR SHARE").WithArgs(1).WillReturnRows(sesiStatusRow(1, "COMPLETED"))
	mock.ExpectRollback()

	svc := newUploadServiceWithMock(mock, t)
	_, err = svc.Create(context.Background(), uploadRequest())
	if err == nil {
		t.Fatal("expected conflict, got nil")
	}
	var conflict *model.ConflictError
	if !errors.As(err, &conflict) {
		t.Errorf("expected ConflictError, got %T: %v", err, err)
	}
}
