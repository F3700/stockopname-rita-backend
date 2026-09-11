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

func (c *CoordinatorServiceImpl) Update(ctx context.Context, id int, req *dto.UpdateCoordinatorRequest) error {
	tx, err := c.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	coordinatorModel := &model.Coordinator{
		CoorID:     id,
		CoorStatus: req.Status,
	}

	if err := c.CoordinatorRepository.Update(ctx, tx, coordinatorModel); err != nil {
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

func (c *CoordinatorServiceImpl) FindByIdReport(ctx context.Context, id int) (*dto.CoordinatorReportResponse, error) {
	tx, err := c.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	report, err := c.CoordinatorRepository.FindByIdReport(ctx, tx, id)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &dto.CoordinatorReportResponse{Coordinator: dto.CoordinatorResponse{ID: report.ID, Code: report.Code, Inspector: report.Inspector, RackAssigned: report.RackAssigned, RackCompleted: report.RackCompleted, Status: report.Status}, SessionCode: report.SessionCode}, nil
}
