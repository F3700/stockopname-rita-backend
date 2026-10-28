package service

import (
	"context"
	"encoding/json"
	"stockopname-rita-backend/internal/dto"
	"stockopname-rita-backend/internal/model"
	"stockopname-rita-backend/internal/repository"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5"
)

type UploadServiceImpl struct {
	UploadRepository      repository.UploadRepository
	RackRepository        repository.RackRepository
	CoordinatorRepository repository.CoordinatorRepository
	SesiRepository        repository.SesiRepository
	ProductRepository     repository.ProductRepository
	StockOpnameRepository repository.StockOpnameRepository
	Pool                  repository.DBPool
	Validator             *validator.Validate
}

func NewUploadService(
	uploadRepository repository.UploadRepository,
	rackRepository repository.RackRepository,
	coordinatorRepository repository.CoordinatorRepository,
	sesiRepository repository.SesiRepository,
	productRepository repository.ProductRepository,
	stockOpnameRepository repository.StockOpnameRepository,
	pool repository.DBPool,
	validate *validator.Validate,
) UploadService {
	return &UploadServiceImpl{
		UploadRepository:      uploadRepository,
		RackRepository:        rackRepository,
		CoordinatorRepository: coordinatorRepository,
		SesiRepository:        sesiRepository,
		ProductRepository:     productRepository,
		StockOpnameRepository: stockOpnameRepository,
		Pool:                  pool,
		Validator:             validate,
	}
}

type uploadRackPayload struct {
	rakID     int
	snapshot  string
	results   []*model.StockOpname
	itemCount int
}

// resolveProduct finds the master product by scanned barcode or PLU.
func (s *UploadServiceImpl) resolveProduct(ctx context.Context, tx pgx.Tx, sesiId int, item dto.UploadItemRequest) (*model.Product, error) {
	if item.Barcode != "" {
		return s.ProductRepository.FindByBarcode(ctx, tx, sesiId, item.Barcode)
	}
	if item.PLU != "" {
		return s.ProductRepository.FindByPLU(ctx, tx, sesiId, item.PLU)
	}
	return nil, &model.ValidationError{Detail: "barcode or plu is required"}
}

// Create implements [UploadService]. One request creates one upload row that
// groups every rack in the batch; all racks are claimed atomically.
func (s *UploadServiceImpl) Create(ctx context.Context, req dto.CreateUploadRequest) (dto.UploadResponse, error) {
	if err := s.Validator.Struct(req); err != nil {
		return dto.UploadResponse{}, err
	}

	seenRacks := make(map[int]struct{}, len(req.Racks))
	for _, rack := range req.Racks {
		if _, exists := seenRacks[rack.RakID]; exists {
			return dto.UploadResponse{}, &model.ValidationError{Detail: "duplicate rack in upload"}
		}
		seenRacks[rack.RakID] = struct{}{}
	}

	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return dto.UploadResponse{}, err
	}
	defer tx.Rollback(ctx)

	coordinator, err := s.CoordinatorRepository.FindById(ctx, tx, req.CoordinatorID)
	if err != nil {
		return dto.UploadResponse{}, err
	}

	sesi, err := s.SesiRepository.FindByIdForShare(ctx, tx, coordinator.CoorSesiID)
	if err != nil {
		return dto.UploadResponse{}, err
	}
	if err := ensureSessionWritable(sesi); err != nil {
		return dto.UploadResponse{}, err
	}
	if coordinator.CoorStatus != "IN_PROGRESS" && coordinator.CoorStatus != "IN_REVIEW" {
		return dto.UploadResponse{}, &model.ConflictError{Resource: "Coordinator", Detail: "Coordinator is not active"}
	}

	products, err := s.ProductRepository.CountBySesi(ctx, tx, coordinator.CoorSesiID)
	if err != nil {
		return dto.UploadResponse{}, err
	}
	if products == 0 {
		return dto.UploadResponse{}, &model.ConflictError{Resource: "Session", Detail: "Session has no products yet"}
	}

	payloads := make([]uploadRackPayload, 0, len(req.Racks))
	rakIDs := make([]int32, 0, len(req.Racks))
	totalItems := 0

	for _, rackReq := range req.Racks {
		seenProducts := make(map[string]struct{}, len(rackReq.Items))
		uploadItems := make([]model.UploadItem, 0, len(rackReq.Items))
		results := make([]*model.StockOpname, 0, len(rackReq.Items))

		for _, item := range rackReq.Items {
			if err := validateQuantity(*item.Quantity); err != nil {
				return dto.UploadResponse{}, err
			}
			product, err := s.resolveProduct(ctx, tx, coordinator.CoorSesiID, item)
			if err != nil {
				return dto.UploadResponse{}, err
			}
			if _, exists := seenProducts[product.ProductPLU]; exists {
				return dto.UploadResponse{}, &model.ValidationError{Detail: "duplicate product in rack upload"}
			}
			seenProducts[product.ProductPLU] = struct{}{}

			soBarcode := item.Barcode
			if soBarcode == "" {
				soBarcode = product.FirstBarcode()
			}
			uploadItems = append(uploadItems, model.UploadItem{
				PLU:     product.ProductPLU,
				Barcode: soBarcode,
				Name:    product.ProductName,
				Qty:     *item.Quantity,
			})
			results = append(results, snapshot(*item.Quantity, rackReq.RakID, product, item.Barcode))
		}

		snapshotJSON, err := json.Marshal(uploadItems)
		if err != nil {
			return dto.UploadResponse{}, err
		}
		payloads = append(payloads, uploadRackPayload{
			rakID:     rackReq.RakID,
			snapshot:  string(snapshotJSON),
			results:   results,
			itemCount: len(uploadItems),
		})
		rakIDs = append(rakIDs, int32(rackReq.RakID))
		totalItems += len(uploadItems)
	}

	upload := &model.Upload{CoorID: coordinator.CoorID}
	if err := s.UploadRepository.Save(ctx, tx, upload); err != nil {
		return dto.UploadResponse{}, err
	}

	claimed, err := s.RackRepository.ClaimForUpload(ctx, tx, upload.UploadID, rakIDs, coordinator.CoorID)
	if err != nil {
		return dto.UploadResponse{}, err
	}
	if claimed != int64(len(rakIDs)) {
		return dto.UploadResponse{}, &model.ConflictError{Resource: "Rack", Detail: "One or more racks are already uploaded or invalid"}
	}

	for _, payload := range payloads {
		if err := s.RackRepository.SetSnapshot(ctx, tx, payload.rakID, payload.snapshot); err != nil {
			return dto.UploadResponse{}, err
		}
		for _, result := range payload.results {
			if err := s.StockOpnameRepository.Save(ctx, tx, result); err != nil {
				return dto.UploadResponse{}, err
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return dto.UploadResponse{}, err
	}

	return dto.UploadResponse{
		UploadID:  upload.UploadID,
		RackCount: len(payloads),
		ItemCount: totalItems,
		CreatedAt: upload.CreatedAt.Format(time.RFC3339),
	}, nil
}

// FindAllSummary implements [UploadService].
func (s *UploadServiceImpl) FindAllSummary(ctx context.Context, coordinatorId int) ([]dto.UploadResponse, error) {
	uploads, err := s.UploadRepository.FindAllSummary(ctx, coordinatorId)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.UploadResponse, 0, len(uploads))
	for _, upload := range uploads {
		responses = append(responses, dto.UploadResponse{
			UploadID:  upload.UploadID,
			RackCount: upload.RackCount,
			ItemCount: upload.ItemCount,
			CreatedAt: upload.CreatedAt.Format(time.RFC3339),
		})
	}
	return responses, nil
}

// FindDetail implements [UploadService].
func (s *UploadServiceImpl) FindDetail(ctx context.Context, uploadID int) (dto.UploadDetailResponse, error) {
	detail, err := s.UploadRepository.FindDetail(ctx, uploadID)
	if err != nil {
		return dto.UploadDetailResponse{}, err
	}

	racks := make([]dto.UploadDetailRackResponse, 0, len(detail.Racks))
	for _, rack := range detail.Racks {
		items := make([]dto.UploadItemResponse, 0, len(rack.Items))
		for _, item := range rack.Items {
			items = append(items, dto.UploadItemResponse{
				PLU:     item.PLU,
				Barcode: item.Barcode,
				Name:    item.Name,
				Qty:     item.Qty,
			})
		}
		racks = append(racks, dto.UploadDetailRackResponse{
			RakID:     rack.RakID,
			RakName:   rack.RakName,
			ItemCount: rack.ItemCount,
			Items:     items,
		})
	}

	return dto.UploadDetailResponse{
		UploadResponse: dto.UploadResponse{
			UploadID:  detail.UploadID,
			RackCount: detail.RackCount,
			ItemCount: detail.ItemCount,
			CreatedAt: detail.CreatedAt.Format(time.RFC3339),
		},
		CoordinatorCode: detail.CoordinatorCode,
		SessionCode:     detail.SessionCode,
		SessionLocation: detail.SessionLocation,
		Racks:           racks,
	}, nil
}
