package service

import (
	"context"
	"stockopname-rita-backend/internal/dto"
	"stockopname-rita-backend/internal/model"
	"stockopname-rita-backend/internal/repository"

	"github.com/jackc/pgx/v5/pgxpool"
)

type CoordinatorServiceImpl struct {
	CoordinatorRepository repository.CoordinatorRepository
	Pool                  *pgxpool.Pool
}

func NewCoordinatorService(coordinatorRepository repository.CoordinatorRepository, pool *pgxpool.Pool) CoordinatorService {
	return &CoordinatorServiceImpl{
		CoordinatorRepository: coordinatorRepository,
		Pool:                  pool,
	}
}

// Cancel implements [CoordinatorService].
func (c *CoordinatorServiceImpl) Cancel(ctx context.Context, id int) error {
	tx, err := c.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := c.CoordinatorRepository.Cancel(ctx, tx, id); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// Complete implements [CoordinatorService].
func (c *CoordinatorServiceImpl) Complete(ctx context.Context, id int) error {
	tx, err := c.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := c.CoordinatorRepository.Complete(ctx, tx, id); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// FindAllSummary implements [CoordinatorService].
func (c *CoordinatorServiceImpl) FindAllSummary(ctx context.Context, sesiId *int) ([]*dto.CoordinatorResponse, error) {
	tx, err := c.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var coordinators []*model.CoordinatorSummary

	if sesiId != nil {
		coordinators, err = c.CoordinatorRepository.FindBySesiIdSummary(ctx, tx, *sesiId)
		if err != nil {
			return nil, err
		}
	} else {
		coordinators, err = c.CoordinatorRepository.FindAllSummary(ctx, tx)
		if err != nil {
			return nil, err
		}
	}

	var response []*dto.CoordinatorResponse
	for _, coordinator := range coordinators {
		response = append(response, &dto.CoordinatorResponse{
			ID:            coordinator.ID,
			Code:          coordinator.Code,
			Inspector:     coordinator.Inspector,
			RackAssigned:  coordinator.RackAssigned,
			RackCompleted: coordinator.RackCompleted,
			Status:        coordinator.Status,
		})
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	return response, nil
}

// FindByIdSummary implements [CoordinatorService].
func (c *CoordinatorServiceImpl) FindByIdSummary(ctx context.Context, id int) (*dto.CoordinatorResponse, error) {
	tx, err := c.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	coordinator, err := c.CoordinatorRepository.FindByIdSummary(ctx, tx, id)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &dto.CoordinatorResponse{
		ID:            coordinator.ID,
		Code:          coordinator.Code,
		Inspector:     coordinator.Inspector,
		RackAssigned:  coordinator.RackAssigned,
		RackCompleted: coordinator.RackCompleted,
		Status:        coordinator.Status,
	}, nil
}
