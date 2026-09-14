package service

import (
	"context"
	"math"
	"stockopname-rita-backend/internal/dto"
	"stockopname-rita-backend/internal/model"
	"stockopname-rita-backend/internal/repository"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
)

type StockOpnameServiceImpl struct {
	Pool                  *pgxpool.Pool
	StockOpnameRepository repository.StockOpnameRepository
	Validator             *validator.Validate
}

func NewStockOpnameService(stockOpnameRepository repository.StockOpnameRepository, pool *pgxpool.Pool, validate *validator.Validate) StockOpnameService {
	return &StockOpnameServiceImpl{
		Pool:                  pool,
		StockOpnameRepository: stockOpnameRepository,
		Validator:             validate,
	}
}

// Create implements [StockOpnameService].
func (s *StockOpnameServiceImpl) Create(ctx context.Context, req dto.CreateStockOpnameRequest) (dto.StockOpnameResponse, error) {
	if err := s.Validator.Struct(req); err != nil {
		return dto.StockOpnameResponse{}, err
	}

	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return dto.StockOpnameResponse{}, err
	}
	defer tx.Rollback(ctx)

	modelSO := &model.StockOpname{
		StockOpnameQuantity:  req.Quantity,
		StockOpnameProductID: req.ProductID,
		StockOpnameRakID:     req.RakID,
	}

	if err := s.StockOpnameRepository.Save(ctx, tx, modelSO); err != nil {
		return dto.StockOpnameResponse{}, err
	}

	modelSummary, err := s.StockOpnameRepository.FindById(ctx, tx, modelSO.StockOpnameID)
	if err != nil {
		return dto.StockOpnameResponse{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return dto.StockOpnameResponse{}, err
	}

	return dto.StockOpnameResponse{
		Id:              modelSummary.Id,
		Barcode:         modelSummary.Barcode,
		Name:            modelSummary.Name,
		Quantity:        modelSummary.Quantity,
		RackName:        modelSummary.RackName,
		InspectorCode:   modelSummary.InspectorCode,
		CoordinatorCode: modelSummary.CoordinatorCode,
		UpdatedAt:       modelSummary.UpdatedAt.Format(time.RFC3339),
	}, nil
}

// Delete implements [StockOpnameService].
func (s *StockOpnameServiceImpl) Delete(ctx context.Context, id int) error {
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := s.StockOpnameRepository.Delete(ctx, tx, id); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// FindAll implements [StockOpnameService].
func (s *StockOpnameServiceImpl) FindAll(ctx context.Context, pagination *dto.Pagination, search string, coorId *int, sesiId *int) ([]dto.StockOpnameResponse, error) {
	offset := (pagination.Page - 1) * pagination.Limit

	stockOpnames, total, err := s.StockOpnameRepository.FindAll(ctx, pagination.Limit, offset, search, sesiId, coorId)
	if err != nil {
		return nil, err
	}

	var responses []dto.StockOpnameResponse
	for _, so := range stockOpnames {
		responses = append(responses, dto.StockOpnameResponse{
			Id:              so.Id,
			Barcode:         so.Barcode,
			Name:            so.Name,
			Quantity:        so.Quantity,
			RackName:        so.RackName,
			InspectorCode:   so.InspectorCode,
			CoordinatorCode: so.CoordinatorCode,
			UpdatedAt:       so.UpdatedAt.Format(time.RFC3339),
		})
	}

	pagination.TotalItems = total
	pagination.TotalPages = int(math.Ceil(float64(total) / float64(pagination.Limit)))

	return responses, nil
}

// FindById implements [StockOpnameService].
func (s *StockOpnameServiceImpl) FindById(ctx context.Context, id int) (dto.StockOpnameResponse, error) {
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return dto.StockOpnameResponse{}, err
	}
	defer tx.Rollback(ctx)

	modelSummary, err := s.StockOpnameRepository.FindById(ctx, tx, id)
	if err != nil {
		return dto.StockOpnameResponse{}, err
	}

	return dto.StockOpnameResponse{
		Id:              modelSummary.Id,
		Barcode:         modelSummary.Barcode,
		Name:            modelSummary.Name,
		Quantity:        modelSummary.Quantity,
		RackName:        modelSummary.RackName,
		InspectorCode:   modelSummary.InspectorCode,
		CoordinatorCode: modelSummary.CoordinatorCode,
		UpdatedAt:       modelSummary.UpdatedAt.Format(time.RFC3339),
	}, nil
}

// Update implements [StockOpnameService].
func (s *StockOpnameServiceImpl) Update(ctx context.Context, id int, req dto.UpdateStockOpnameRequest) (dto.StockOpnameResponse, error) {
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return dto.StockOpnameResponse{}, err
	}
	defer tx.Rollback(ctx)

	modelSO := &model.StockOpname{
		StockOpnameID:       id,
		StockOpnameQuantity: req.Quantity,
	}

	if err := s.StockOpnameRepository.Update(ctx, tx, modelSO); err != nil {
		return dto.StockOpnameResponse{}, err
	}

	result, err := s.StockOpnameRepository.FindById(ctx, tx, id)
	if err != nil {
		return dto.StockOpnameResponse{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return dto.StockOpnameResponse{}, err
	}

	return dto.StockOpnameResponse{
		Id:              result.Id,
		Barcode:         result.Barcode,
		Name:            result.Name,
		Quantity:        result.Quantity,
		RackName:        result.RackName,
		InspectorCode:   result.InspectorCode,
		CoordinatorCode: result.CoordinatorCode,
		UpdatedAt:       result.UpdatedAt.Format(time.RFC3339),
	}, nil
}

func (s *StockOpnameServiceImpl) CreateByRack(ctx context.Context, req dto.CreateStockOpnameByRackRequest) error {
	if err := s.Validator.Struct(req); err != nil {
		return err
	}

	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, item := range req.Items {
		modelSO := &model.StockOpname{
			StockOpnameQuantity:  item.Quantity,
			StockOpnameProductID: item.ProductID,
			StockOpnameRakID:     req.RackID,
		}
		if err := s.StockOpnameRepository.Save(ctx, tx, modelSO); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}
