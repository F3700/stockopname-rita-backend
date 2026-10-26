package service

import (
	"context"
	"fmt"
	"math"
	"sort"
	"stockopname-rita-backend/internal/dto"
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
	ProductRepository     repository.ProductRepository
	StockOpnameRepository repository.StockOpnameRepository
	RackRepository        repository.RackRepository
}

func NewSesiService(sesiRepository repository.SesiRepository, coordinatorRepository repository.CoordinatorRepository, productRepository repository.ProductRepository, stockOpnameRepository repository.StockOpnameRepository, rackRepository repository.RackRepository, pool repository.DBPool, validate *validator.Validate) SesiService {
	return &SesiServiceImpl{
		Validator:             validate,
		Pool:                  pool,
		SesiRepository:        sesiRepository,
		CoordinatorRepository: coordinatorRepository,
		ProductRepository:     productRepository,
		StockOpnameRepository: stockOpnameRepository,
		RackRepository:        rackRepository,
	}
}

func sesiEndDate(sesi *model.Sesi) string {
	if !sesi.SesiEndedAt.Valid {
		return ""
	}
	return sesi.SesiEndedAt.Time.Format(time.RFC3339)
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
		EndDate:   sesiEndDate(sesiModel),
	}

	return response, nil
}

// Create implements [SesiService].
func (s *SesiServiceImpl) Create(ctx context.Context, req dto.CreateSesiRequest) (dto.SesiResponse, error) {
	if err := s.Validator.Struct(req); err != nil {
		return dto.SesiResponse{}, err
	}
	if err := validateRackRanges(req.Coordinators); err != nil {
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

	for _, input := range req.Coordinators {
		coordinatorModel := &model.Coordinator{
			CoorCode:   input.Code,
			CoorStatus: "IN_PROGRESS",
			CoorSesiID: sesiModel.SesiID,
		}
		if err := s.CoordinatorRepository.Save(ctx, tx, coordinatorModel); err != nil {
			return dto.SesiResponse{}, fmt.Errorf("failed to create coordinator %s: %w", input.Code, err)
		}

		for n := input.RackStart; n <= input.RackEnd; n++ {
			rack := &model.Rack{
				RackName: fmt.Sprintf("Rak %d", n),
				CoorID:   coordinatorModel.CoorID,
			}
			if err := s.RackRepository.Save(ctx, tx, rack); err != nil {
				return dto.SesiResponse{}, fmt.Errorf("failed to create rack %s for coordinator %s: %w", rack.RackName, input.Code, err)
			}
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
		EndDate:   sesiEndDate(sesiModel),
	}, nil
}

// Delete implements [SesiService].
func (s *SesiServiceImpl) Delete(ctx context.Context, id int) error {
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	current, err := s.SesiRepository.FindByIdForUpdate(ctx, tx, id)
	if err != nil {
		return err
	}

	if current.SesiStatus == "IN_PROGRESS" {
		products, err := s.ProductRepository.CountBySesi(ctx, tx, id)
		if err != nil {
			return err
		}
		if products > 0 {
			return &model.ConflictError{Resource: "Session", Detail: "Session still has data"}
		}

		results, err := s.StockOpnameRepository.CountBySesi(ctx, tx, id)
		if err != nil {
			return err
		}
		if results > 0 {
			return &model.ConflictError{Resource: "Session", Detail: "Session still has data"}
		}
	}

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

	current, err := s.SesiRepository.FindByIdForUpdate(ctx, tx, id)
	if err != nil {
		return err
	}

	if err := ensureSessionWritable(current); err != nil {
		return err
	}

	if req.Status == "COMPLETED" {
		products, err := s.ProductRepository.CountBySesi(ctx, tx, id)
		if err != nil {
			return err
		}
		if products == 0 {
			return &model.ConflictError{Resource: "Session", Detail: "Session has no products yet"}
		}
	}

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
			EndDate:   sesiEndDate(sesiModel),
		}
		sesiResponses = append(sesiResponses, sesiResponse)
	}

	pagination.TotalItems = total
	pagination.TotalPages = int(math.Ceil(float64(total) / float64(pagination.Limit)))

	return sesiResponses, nil
}

// maxRackSpan caps how many racks one coordinator may own.
const maxRackSpan = 200

// validateRackRanges rejects overlapping or oversized rack ranges. Overlap is
// checked here (not by a DB constraint) because rack names are unique per
// coordinator, while numbering is session-wide.
func validateRackRanges(coordinators []dto.CoordinatorInput) error {
	ranges := make([]dto.CoordinatorInput, len(coordinators))
	copy(ranges, coordinators)

	for _, coordinator := range ranges {
		span := coordinator.RackEnd - coordinator.RackStart + 1
		if span < 1 || span > maxRackSpan {
			return &model.ValidationError{Detail: fmt.Sprintf(
				"rack range for %s must cover 1 to %d racks", coordinator.Code, maxRackSpan,
			)}
		}
	}

	sort.Slice(ranges, func(i, j int) bool {
		return ranges[i].RackStart < ranges[j].RackStart
	})
	for i := 1; i < len(ranges); i++ {
		if ranges[i].RackStart <= ranges[i-1].RackEnd {
			return &model.ValidationError{Detail: fmt.Sprintf(
				"rack ranges of %s and %s overlap", ranges[i-1].Code, ranges[i].Code,
			)}
		}
	}
	return nil
}
