package service

import (
	"context"
	"stockopname-rita-backend/internal/dto"
	"stockopname-rita-backend/internal/helper"
	"stockopname-rita-backend/internal/model"
	"stockopname-rita-backend/internal/repository"
	"time"

	"github.com/jackc/pgx/v5"
)

type CoordinatorServiceImpl struct {
	CoordinatorRepository repository.CoordinatorRepository
	SesiRepository        repository.SesiRepository
	ProductRepository     repository.ProductRepository
	RackRepository        repository.RackRepository
	Pool                  repository.DBPool
}

func NewCoordinatorService(coordinatorRepository repository.CoordinatorRepository, sesiRepository repository.SesiRepository, productRepository repository.ProductRepository, rackRepository repository.RackRepository, pool repository.DBPool) CoordinatorService {
	return &CoordinatorServiceImpl{
		CoordinatorRepository: coordinatorRepository,
		SesiRepository:        sesiRepository,
		ProductRepository:     productRepository,
		RackRepository:        rackRepository,
		Pool:                  pool,
	}
}

func (c *CoordinatorServiceImpl) Update(ctx context.Context, id int, req *dto.UpdateCoordinatorRequest) error {
	tx, err := c.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	current, err := c.CoordinatorRepository.FindById(ctx, tx, id)
	if err != nil {
		return err
	}

	sesi, err := c.SesiRepository.FindByIdForShare(ctx, tx, current.CoorSesiID)
	if err != nil {
		return err
	}
	if err := ensureSessionWritable(sesi); err != nil {
		return err
	}

	coordinatorModel := &model.Coordinator{
		CoorID:     id,
		CoorStatus: req.Status,
	}

	if err := c.CoordinatorRepository.Update(ctx, tx, coordinatorModel); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// checkJoinAllowed enforces D3 + D6 for coordinator joins.
func (c *CoordinatorServiceImpl) checkJoinAllowed(ctx context.Context, tx pgx.Tx, coordinator *model.Coordinator) error {
	sesi, err := c.SesiRepository.FindById(ctx, coordinator.CoorSesiID)
	if err != nil {
		return err
	}
	if err := ensureSessionWritable(sesi); err != nil {
		return err
	}
	if coordinator.CoorStatus != "IN_PROGRESS" {
		return &model.ConflictError{
			Resource: "Coordinator",
			Detail:   "Coordinator is not active",
		}
	}
	total, err := c.ProductRepository.CountBySesi(ctx, tx, coordinator.CoorSesiID)
	if err != nil {
		return err
	}
	if total == 0 {
		return &model.ConflictError{Resource: "Session", Detail: "Session has no products yet"}
	}
	return nil
}

// JoinByCoordinatorQR implements [CoordinatorService].
func (c *CoordinatorServiceImpl) JoinByCoordinatorQR(ctx context.Context, req dto.JoinByCoordinatorQRRequest) (dto.CoordinatorJoinResponse, error) {
	if req.CoordinatorQR == "" {
		return dto.CoordinatorJoinResponse{}, &model.ValidationError{Detail: "coordinator_qr is required"}
	}

	coorID, err := helper.ParseCoordinatorQR(req.CoordinatorQR)
	if err != nil {
		return dto.CoordinatorJoinResponse{}, err
	}

	tx, err := c.Pool.Begin(ctx)
	if err != nil {
		return dto.CoordinatorJoinResponse{}, err
	}
	defer tx.Rollback(ctx)

	coordinator, err := c.CoordinatorRepository.FindById(ctx, tx, coorID)
	if err != nil {
		return dto.CoordinatorJoinResponse{}, err
	}

	if err := c.checkJoinAllowed(ctx, tx, coordinator); err != nil {
		return dto.CoordinatorJoinResponse{}, err
	}

	racks, err := c.RackRepository.FindAll(ctx, &coorID, nil)
	if err != nil {
		return dto.CoordinatorJoinResponse{}, err
	}

	joinRacks := make([]dto.JoinRackResponse, 0, len(racks))
	for _, rack := range racks {
		uploadedAt := ""
		if rack.UploadedAt.Valid {
			uploadedAt = rack.UploadedAt.Time.Format(time.RFC3339)
		}
		joinRacks = append(joinRacks, dto.JoinRackResponse{
			ID:         rack.RakID,
			Name:       rack.RakName,
			UploadedAt: uploadedAt,
		})
	}

	return dto.CoordinatorJoinResponse{
		CoordinatorID: coordinator.CoorID,
		SesiID:        coordinator.CoorSesiID,
		Rak:           joinRacks,
	}, nil
}

// FindAllSummary implements [CoordinatorService].
func (c *CoordinatorServiceImpl) FindAllSummary(ctx context.Context, sesiId *int) ([]*dto.CoordinatorResponse, error) {
	var coordinators []*model.CoordinatorSummary
	var err error

	if sesiId != nil {
		coordinators, err = c.CoordinatorRepository.FindBySesiIdSummary(ctx, *sesiId)
		if err != nil {
			return nil, err
		}
	} else {
		coordinators, err = c.CoordinatorRepository.FindAllSummary(ctx)
		if err != nil {
			return nil, err
		}
	}

	var response []*dto.CoordinatorResponse
	for _, coordinator := range coordinators {
		response = append(response, &dto.CoordinatorResponse{
			ID:            coordinator.ID,
			Code:          coordinator.Code,
			RackAssigned:  coordinator.RackAssigned,
			RackCompleted: coordinator.RackCompleted,
			Status:        coordinator.Status,
		})
	}

	return response, nil
}

// FindByIdSummary implements [CoordinatorService].
func (c *CoordinatorServiceImpl) FindByIdSummary(ctx context.Context, id int) (*dto.CoordinatorResponse, error) {
	coordinator, err := c.CoordinatorRepository.FindByIdSummary(ctx, id)
	if err != nil {
		return nil, err
	}

	return &dto.CoordinatorResponse{
		ID:            coordinator.ID,
		Code:          coordinator.Code,
		RackAssigned:  coordinator.RackAssigned,
		RackCompleted: coordinator.RackCompleted,
		Status:        coordinator.Status,
	}, nil
}

// FindByIdDetail implements [CoordinatorService].
func (c *CoordinatorServiceImpl) FindByIdDetail(ctx context.Context, id int) (*dto.CoordinatorDetailResponse, error) {
	coordinator, err := c.CoordinatorRepository.FindByIdDetail(ctx, id)
	if err != nil {
		return nil, err
	}

	return &dto.CoordinatorDetailResponse{
		ID:              coordinator.ID,
		Code:            coordinator.Code,
		RackAssigned:    coordinator.RackAssigned,
		RackCompleted:   coordinator.RackCompleted,
		Status:          coordinator.Status,
		SessionCode:     coordinator.SessionCode,
		SessionLocation: coordinator.SessionLocation,
	}, nil
}

func (c *CoordinatorServiceImpl) FindByIdReport(ctx context.Context, id int) (*dto.CoordinatorReportResponse, error) {
	report, err := c.CoordinatorRepository.FindByIdReport(ctx, id)
	if err != nil {
		return nil, err
	}
	return &dto.CoordinatorReportResponse{Coordinator: dto.CoordinatorResponse{ID: report.ID, Code: report.Code, RackAssigned: report.RackAssigned, RackCompleted: report.RackCompleted, Status: report.Status}, SessionCode: report.SessionCode}, nil
}
