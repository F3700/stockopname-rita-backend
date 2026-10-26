package service

import (
	"context"
	"stockopname-rita-backend/internal/dto"
	"stockopname-rita-backend/internal/model"
	"stockopname-rita-backend/internal/repository"
	"time"
)

type RackServiceImpl struct {
	RackRepository        repository.RackRepository
	CoordinatorRepository repository.CoordinatorRepository
	SesiRepository        repository.SesiRepository
	StockOpnameRepository repository.StockOpnameRepository
	Pool                  repository.DBPool
}

func NewRackService(rackRepository repository.RackRepository, coordinatorRepository repository.CoordinatorRepository, sesiRepository repository.SesiRepository, stockOpnameRepository repository.StockOpnameRepository, pool repository.DBPool) RackService {
	return &RackServiceImpl{
		RackRepository:        rackRepository,
		CoordinatorRepository: coordinatorRepository,
		SesiRepository:        sesiRepository,
		StockOpnameRepository: stockOpnameRepository,
		Pool:                  pool,
	}
}

// FindAllSummary implements [RackService].
func (r *RackServiceImpl) FindAllSummary(ctx context.Context, coordinatorId *int, sessionId *int) ([]dto.RackResponse, error) {
	racks, err := r.RackRepository.FindAll(ctx, coordinatorId, sessionId)
	if err != nil {
		return nil, err
	}

	var responses []dto.RackResponse
	for _, rack := range racks {
		uploadedAt := ""
		if rack.UploadedAt.Valid {
			uploadedAt = rack.UploadedAt.Time.Format(time.RFC3339)
		}
		responses = append(responses, dto.RackResponse{
			RackID:     rack.RakID,
			RackName:   rack.RakName,
			UploadedAt: uploadedAt,
		})
	}
	return responses, nil
}

// FindProgress implements [RackService].
func (r *RackServiceImpl) FindProgress(ctx context.Context, coordinatorId *int, sessionId *int) ([]dto.RackProgressResponse, error) {
	rackProgresses, err := r.RackRepository.FindProgress(ctx, coordinatorId, sessionId)
	if err != nil {
		return nil, err
	}

	var responses []dto.RackProgressResponse
	for _, rackProgress := range rackProgresses {
		responses = append(responses, dto.RackProgressResponse{
			RackAssigned:  rackProgress.RackAssigned,
			RackCompleted: rackProgress.RackCompleted,
			TotalItems:    rackProgress.TotalItems,
		})
	}
	return responses, nil
}

// DeleteRack implements [RackService].
func (r *RackServiceImpl) DeleteRack(ctx context.Context, rakID int) error {
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	rack, err := r.RackRepository.FindById(ctx, tx, rakID)
	if err != nil {
		return err
	}

	coordinator, err := r.CoordinatorRepository.FindById(ctx, tx, rack.CoorID)
	if err != nil {
		return err
	}

	sesi, err := r.SesiRepository.FindByIdForShare(ctx, tx, coordinator.CoorSesiID)
	if err != nil {
		return err
	}
	if err := ensureSessionWritable(sesi); err != nil {
		return err
	}

	locked, err := r.RackRepository.FindByIdForUpdate(ctx, tx, rakID)
	if err != nil {
		return err
	}
	if locked.UploadedAt.Valid {
		return &model.ConflictError{Resource: "Rack", Detail: "Rack already uploaded"}
	}

	items, err := r.StockOpnameRepository.CountByRak(ctx, tx, rakID)
	if err != nil {
		return err
	}
	if items > 0 {
		return &model.ConflictError{Resource: "Rack", Detail: "Rack has items"}
	}

	if err := r.CoordinatorRepository.LockById(ctx, tx, rack.CoorID); err != nil {
		return err
	}

	total, err := r.RackRepository.CountByCoordinator(ctx, tx, rack.CoorID)
	if err != nil {
		return err
	}
	if total <= 1 {
		return &model.ConflictError{Resource: "Coordinator", Detail: "Coordinator must have at least one rack"}
	}

	if err := r.RackRepository.Delete(ctx, tx, rakID); err != nil {
		return err
	}
	if err := r.CoordinatorRepository.RecalcStatus(ctx, tx, rack.CoorID); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
