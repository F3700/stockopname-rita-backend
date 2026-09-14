package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"stockopname-rita-backend/internal/dto"
	"stockopname-rita-backend/internal/model"
)

func sesiModel() *model.Sesi {
	return &model.Sesi{
		SesiID:        1,
		SesiLocation:  "Gudang A",
		SesiCode:      "SESI-01",
		SesiStatus:    "IN_PROGRESS",
		SesiStartedAt: time.Date(2026, 9, 10, 8, 0, 0, 0, time.UTC),
		SesiEndedAt:   pgtype.Timestamptz{Valid: false},
	}
}

func TestSesiFindById(t *testing.T) {
	svc := NewSesiService(&fakeSesiRepository{
		findByIdFn: func(ctx context.Context, id int) (*model.Sesi, error) {
			if id != 1 {
				t.Errorf("expected id 1, got %d", id)
			}
			return sesiModel(), nil
		},
	}, nil, nil, newTestValidator(t))

	res, err := svc.FindById(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.ID != 1 || res.Code != "SESI-01" || res.Location != "Gudang A" || res.Status != "IN_PROGRESS" {
		t.Errorf("unexpected response %+v", res)
	}
	if res.StartDate != "2026-09-10T08:00:00Z" {
		t.Errorf("unexpected startDate %q", res.StartDate)
	}
	if res.EndDate != "0001-01-01T00:00:00Z" {
		t.Errorf("expected zero-time endDate for invalid timestamptz, got %q", res.EndDate)
	}
}

func TestSesiFindByIdNotFound(t *testing.T) {
	svc := NewSesiService(&fakeSesiRepository{
		findByIdFn: func(ctx context.Context, id int) (*model.Sesi, error) {
			return nil, &model.NotFoundError{Resource: "Session", ID: id}
		},
	}, nil, nil, newTestValidator(t))

	_, err := svc.FindById(context.Background(), 99)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var notFound *model.NotFoundError
	if !errors.As(err, &notFound) {
		t.Errorf("expected NotFoundError to propagate, got %T: %v", err, err)
	}
}

func TestSesiFindAll(t *testing.T) {
	svc := NewSesiService(&fakeSesiRepository{
		findAllInPageSearchFn: func(ctx context.Context, limit int, offset int, search string) ([]*model.Sesi, int, error) {
			if limit != 10 || offset != 0 {
				t.Errorf("expected limit 10 offset 0, got limit %d offset %d", limit, offset)
			}
			if search != "gudang" {
				t.Errorf("expected search %q, got %q", "gudang", search)
			}
			return []*model.Sesi{sesiModel()}, 1, nil
		},
	}, nil, nil, newTestValidator(t))

	pagination := &dto.Pagination{Page: 1, Limit: 10}
	responses, err := svc.FindAll(context.Background(), pagination, "gudang")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(responses) != 1 || responses[0].Code != "SESI-01" {
		t.Errorf("unexpected responses %+v", responses)
	}
	if pagination.TotalItems != 1 || pagination.TotalPages != 1 {
		t.Errorf("unexpected pagination %+v", pagination)
	}
}

func TestSesiFindAllRepoError(t *testing.T) {
	svc := NewSesiService(&fakeSesiRepository{
		findAllInPageSearchFn: func(ctx context.Context, limit int, offset int, search string) ([]*model.Sesi, int, error) {
			return nil, 0, errors.New("boom")
		},
	}, nil, nil, newTestValidator(t))

	_, err := svc.FindAll(context.Background(), &dto.Pagination{Page: 1, Limit: 10}, "")
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestSesiCreateValidationFailure(t *testing.T) {
	svc := NewSesiService(&fakeSesiRepository{}, nil, nil, newTestValidator(t))

	_, err := svc.Create(context.Background(), dto.CreateSesiRequest{})
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
}
