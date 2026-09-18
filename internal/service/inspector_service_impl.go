package service

import (
	"context"
	"stockopname-rita-backend/internal/dto"
	"stockopname-rita-backend/internal/helper"
	"stockopname-rita-backend/internal/model"
	"stockopname-rita-backend/internal/repository"
)

type InspectorServiceImpl struct {
	InspectorRepository   repository.InspectorRepository
	CoordinatorRepository repository.CoordinatorRepository
	RackRepository        repository.RackRepository
	Pool                  repository.DBPool
}

func NewInspectorService(inspectorRepository repository.InspectorRepository, coordinatorRepository repository.CoordinatorRepository, rackRepository repository.RackRepository, pool repository.DBPool) InspectorService {
	return &InspectorServiceImpl{
		InspectorRepository:   inspectorRepository,
		CoordinatorRepository: coordinatorRepository,
		RackRepository:        rackRepository,
		Pool:                  pool,
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

func (i *InspectorServiceImpl) CreateInspector(ctx context.Context, request dto.InspectorRequest) (dto.InspectorJoinResponse, error) {
	tx, err := i.Pool.Begin(ctx)
	if err != nil {
		return dto.InspectorJoinResponse{}, err
	}
	defer tx.Rollback(ctx)

	coordinator, err := i.CoordinatorRepository.FindBySesiAndCoorCode(ctx, tx, request.SesiCode, request.CoorCode)
	if err != nil {
		return dto.InspectorJoinResponse{}, err
	}

	if coordinator.CoorStatus != "IN_PROGRESS" {
		return dto.InspectorJoinResponse{}, &model.ConflictError{
			Resource: "Coordinator",
			Detail:   "Coordinator is not active",
		}
	}

	inspectorModel := model.Inspector{
		InspectorCode: request.InspectorCode,
		CoordinatorID: coordinator.CoorID,
	}

	if err := i.InspectorRepository.Save(ctx, tx, &inspectorModel); err != nil {
		return dto.InspectorJoinResponse{}, err
	}

	var rakResponses []dto.RackResponse

	for _, rak := range request.Rak {
		rackModel := model.Rack{
			RackName:    rak,
			InspectorID: inspectorModel.InspectorID,
		}
		if err := i.RackRepository.Save(ctx, tx, &rackModel); err != nil {
			return dto.InspectorJoinResponse{}, err
		}
		rakResponse := dto.RackResponse{
			RackID:   rackModel.RackID,
			RackName: rackModel.RackName,
		}
		rakResponses = append(rakResponses, rakResponse)
	}

	inspectorJoinResponse := dto.InspectorJoinResponse{
		InspectorID: inspectorModel.InspectorID,
		Rak:         rakResponses,
	}

	err = tx.Commit(ctx)
	if err != nil {
		return dto.InspectorJoinResponse{}, err
	}

	return inspectorJoinResponse, nil
}

// CreateInspectorByCoordinatorQR joins a coordinator by scanning its QR code
// (e.g. "RITA-COOR-12"). Only inspector_code + rak are required from mobile.
func (i *InspectorServiceImpl) CreateInspectorByCoordinatorQR(ctx context.Context, request dto.InspectorJoinByCoordinatorQRRequest) (dto.InspectorJoinResponse, error) {
	if request.CoordinatorQR == "" || request.InspectorCode == "" || len(request.Rak) == 0 {
		return dto.InspectorJoinResponse{}, &model.ValidationError{Detail: "coordinator_qr, inspector_code and rak are required"}
	}

	coorID, err := helper.ParseCoordinatorQR(request.CoordinatorQR)
	if err != nil {
		return dto.InspectorJoinResponse{}, err
	}

	tx, err := i.Pool.Begin(ctx)
	if err != nil {
		return dto.InspectorJoinResponse{}, err
	}
	defer tx.Rollback(ctx)

	coordinator, err := i.CoordinatorRepository.FindById(ctx, tx, coorID)
	if err != nil {
		return dto.InspectorJoinResponse{}, err
	}

	if coordinator.CoorStatus != "IN_PROGRESS" {
		return dto.InspectorJoinResponse{}, &model.ConflictError{
			Resource: "Coordinator",
			Detail:   "Coordinator is not active",
		}
	}

	inspectorModel := model.Inspector{
		InspectorCode: request.InspectorCode,
		CoordinatorID: coordinator.CoorID,
	}

	if err := i.InspectorRepository.Save(ctx, tx, &inspectorModel); err != nil {
		return dto.InspectorJoinResponse{}, err
	}

	var rakResponses []dto.RackResponse

	for _, rak := range request.Rak {
		rackModel := model.Rack{
			RackName:    rak,
			InspectorID: inspectorModel.InspectorID,
		}
		if err := i.RackRepository.Save(ctx, tx, &rackModel); err != nil {
			return dto.InspectorJoinResponse{}, err
		}
		rakResponse := dto.RackResponse{
			RackID:   rackModel.RackID,
			RackName: rackModel.RackName,
		}
		rakResponses = append(rakResponses, rakResponse)
	}

	inspectorJoinResponse := dto.InspectorJoinResponse{
		InspectorID: inspectorModel.InspectorID,
		Rak:         rakResponses,
	}

	if err := tx.Commit(ctx); err != nil {
		return dto.InspectorJoinResponse{}, err
	}

	return inspectorJoinResponse, nil
}
