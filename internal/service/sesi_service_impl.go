package service

import (
	"context"
	"fmt"
	"math"
	"stockopname-rita-backend/internal/dto"
	"stockopname-rita-backend/internal/helper"
	"stockopname-rita-backend/internal/model"
	"stockopname-rita-backend/internal/repository"
	"time"

	"github.com/go-playground/validator/v10"
)

type SesiServiceImpl struct {
	Validator             *validator.Validate
	Pool                  repository.DBPool
	SesiRepository        repository.SesiRepository
	CoordinatorRepository repository.CoordinatorRepository
}

func NewSesiService(sesiRepository repository.SesiRepository, coordinatorRepository repository.CoordinatorRepository, pool repository.DBPool, validate *validator.Validate) SesiService {
	return &SesiServiceImpl{
		Validator:             validate,
		Pool:                  pool,
		SesiRepository:        sesiRepository,
		CoordinatorRepository: coordinatorRepository,
	}
}

func (s *SesiServiceImpl) FindById(ctx context.Context, id int) (dto.SesiResponse, error) {
	sesiModel, err := s.SesiRepository.FindById(ctx, id)
	if err != nil {
		return dto.SesiResponse{}, fmt.Errorf("failed to find sesi with id %d: %w", id, err)
	}

	response := dto.SesiResponse{
		ID:        sesiModel.SesiID,
		Location:  sesiModel.SesiLocation,
		Code:      sesiModel.SesiCode,
		Status:    sesiModel.SesiStatus,
		StartDate: sesiModel.SesiStartedAt.Format(time.RFC3339),
		EndDate:   helper.ParseDateNullable(&sesiModel.SesiEndedAt.Time),
	}

	return response, nil
}

// Create implements [SesiService].
func (s *SesiServiceImpl) Create(ctx context.Context, req dto.CreateSesiRequest) (dto.SesiResponse, error) {
	if err := s.Validator.Struct(req); err != nil {
		return dto.SesiResponse{}, err
	}

	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return dto.SesiResponse{}, err
	}
	defer tx.Rollback(ctx)

	sesiModel := &model.Sesi{
		SesiLocation: req.Location,
		SesiCode:     req.Code,
		SesiStatus:   "IN_PROGRESS",
	}

	if err := s.SesiRepository.Save(ctx, tx, sesiModel); err != nil {
		return dto.SesiResponse{}, err
	}

	for _, coordinator := range req.Coordinator {
		coordinatorModel := &model.Coordinator{
			CoorCode:   coordinator,
			CoorStatus: "IN_PROGRESS",
			CoorSesiID: sesiModel.SesiID,
		}
		if err := s.CoordinatorRepository.Save(ctx, tx, coordinatorModel); err != nil {
			return dto.SesiResponse{}, fmt.Errorf("failed to create coordinator %s: %w", coordinator, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return dto.SesiResponse{}, err
	}

	return dto.SesiResponse{
		ID:        sesiModel.SesiID,
		Location:  sesiModel.SesiLocation,
		Code:      sesiModel.SesiCode,
		Status:    sesiModel.SesiStatus,
		StartDate: sesiModel.SesiStartedAt.Format(time.RFC3339),
		EndDate:   helper.ParseDateNullable(&sesiModel.SesiEndedAt.Time),
	}, nil
}

// Delete implements [SesiService].
func (s *SesiServiceImpl) Delete(ctx context.Context, id int) error {
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := s.SesiRepository.Delete(ctx, tx, id); err != nil {
		return fmt.Errorf("failed to delete sesi with id %d: %w", id, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Update implements [SesiService].
func (s *SesiServiceImpl) Update(ctx context.Context, id int, req dto.UpdateSesiRequest) error {
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	sesiModel := &model.Sesi{
		SesiID:     id,
		SesiStatus: req.Status,
	}

	if err := s.SesiRepository.Update(ctx, tx, sesiModel); err != nil {
		return fmt.Errorf("failed to update sesi with id %d: %w", id, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *SesiServiceImpl) FindAll(ctx context.Context, pagination *dto.Pagination, search string) ([]dto.SesiResponse, error) {
	offset := (pagination.Page - 1) * pagination.Limit

	sesiModels, total, err := s.SesiRepository.FindAllInPageSearch(ctx, pagination.Limit, offset, search)
	if err != nil {
		return nil, fmt.Errorf("failed to find all sesi: %w", err)
	}

	var sesiResponses []dto.SesiResponse
	for _, sesiModel := range sesiModels {
		sesiResponse := dto.SesiResponse{
			ID:        sesiModel.SesiID,
			Location:  sesiModel.SesiLocation,
			Code:      sesiModel.SesiCode,
			Status:    sesiModel.SesiStatus,
			StartDate: sesiModel.SesiStartedAt.Format(time.RFC3339),
			EndDate:   helper.ParseDateNullable(&sesiModel.SesiEndedAt.Time),
		}
		sesiResponses = append(sesiResponses, sesiResponse)
	}

	pagination.TotalItems = total
	pagination.TotalPages = int(math.Ceil(float64(total) / float64(pagination.Limit)))

	return sesiResponses, nil
}
