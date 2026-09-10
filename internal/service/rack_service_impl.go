package service

import (
	"context"
	"stockopname-rita-backend/internal/dto"
	"stockopname-rita-backend/internal/repository"
)

type RackServiceImpl struct {
	RackRepository repository.RackRepository
}

func NewRackService(rackRepository repository.RackRepository) *RackServiceImpl {
	return &RackServiceImpl{
		RackRepository: rackRepository,
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
