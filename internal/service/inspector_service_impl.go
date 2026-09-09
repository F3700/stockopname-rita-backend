package service

import (
	"context"
	"stockopname-rita-backend/internal/dto"
	"stockopname-rita-backend/internal/model"
	"stockopname-rita-backend/internal/repository"
)

type InspectorServiceImpl struct {
	InspectorRepository repository.InspectorRepository
}

func NewInspectorService(inspectorRepository repository.InspectorRepository) *InspectorServiceImpl {
	return &InspectorServiceImpl{
		InspectorRepository: inspectorRepository,
	}
}

// FindAllSummary implements [InspectorService].
func (i *InspectorServiceImpl) FindAllSummary(ctx context.Context, coorId *int) ([]dto.InspectorResponse, error) {
	var inspectorSummaries []*model.InspectorSummary
	var err error

	if coorId != nil {
		inspectorSummaries, err = i.InspectorRepository.FindByCoorId(ctx, *coorId)
		if err != nil {
			return nil, err
		}
	} else {
		inspectorSummaries, err = i.InspectorRepository.FindAll(ctx)
		if err != nil {
			return nil, err
		}
	}

	var inspectorResponses []dto.InspectorResponse
	for _, inspectorSummary := range inspectorSummaries {
		inspectorResponse := dto.InspectorResponse{
			ID:            inspectorSummary.InspectorID,
			Code:          inspectorSummary.InspectorCode,
			RackAssigned:  inspectorSummary.RackAssigned,
			RackCompleted: inspectorSummary.RackCompleted,
			TotalItems:    inspectorSummary.TotalItems,
		}
		inspectorResponses = append(inspectorResponses, inspectorResponse)
	}

	return inspectorResponses, nil
}
