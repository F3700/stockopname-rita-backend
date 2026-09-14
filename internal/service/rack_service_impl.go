package service

import (
	"context"
	"stockopname-rita-backend/internal/dto"
	"stockopname-rita-backend/internal/model"
	"stockopname-rita-backend/internal/repository"

	"github.com/jackc/pgx/v5/pgxpool"
)

type RackServiceImpl struct {
	RackRepository repository.RackRepository
	Pool           *pgxpool.Pool
}

func NewRackService(rackRepository repository.RackRepository, pool *pgxpool.Pool) RackService {
	return &RackServiceImpl{
		RackRepository: rackRepository,
		Pool:           pool,
	}
}

// FindAllSummary implements [RackService].
func (r *RackServiceImpl) FindAllSummary(ctx context.Context, inspectorId *int, coordinatorId *int, sessionId *int) ([]dto.RackResponse, error) {
	racks, err := r.RackRepository.FindAll(ctx, inspectorId, coordinatorId, sessionId)
	if err != nil {
		return nil, err
	}

	var responses []dto.RackResponse
	for _, rack := range racks {
		responses = append(responses, dto.RackResponse{
			RackID:   rack.RackID,
			RackName: rack.RackName,
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

func (r *RackServiceImpl) CreateRack(ctx context.Context, request dto.CreateRackRequest) (dto.RackResponse, error) {
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return dto.RackResponse{}, err
	}
	defer tx.Rollback(ctx)

	rack := &model.Rack{
		RackName:    request.RackName,
		InspectorID: request.InspectorID,
	}

	if err := r.RackRepository.Save(ctx, tx, rack); err != nil {
		return dto.RackResponse{}, err
	}

	response := dto.RackResponse{
		RackID:   rack.RackID,
		RackName: rack.RackName,
	}

	if err := tx.Commit(ctx); err != nil {
		return dto.RackResponse{}, err
	}

	return response, nil
}
