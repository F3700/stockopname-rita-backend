package service

import (
	"context"
	"encoding/json"
	"stockopname-rita-backend/internal/dto"
	"stockopname-rita-backend/internal/model"
	"stockopname-rita-backend/internal/repository"
	"time"

	"github.com/go-playground/validator/v10"
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

// resolveFromBatch finds a product in the batch that was prefetched for this
// upload. Barcode wins over PLU, matching the previous per-item behaviour.
func resolveFromBatch(byBarcode, byPLU map[string]*model.Product, item dto.UploadItemRequest) (*model.Product, error) {
	if item.Barcode != "" {
		product, ok := byBarcode[item.Barcode]
		if !ok {
			return nil, &model.NotFoundError{Resource: "Product", Detail: "Product with barcode " + item.Barcode + " not found in this session"}
		}
		return product, nil
	}
	if item.PLU != "" {
		product, ok := byPLU[item.PLU]
		if !ok {
			return nil, &model.NotFoundError{Resource: "Product", Detail: "Product with PLU " + item.PLU + " not found in this session"}
		}
		return product, nil
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

	// Pass 1 only collects the referenced codes, so every product is resolved by
	// one query instead of one query per item.
	codes := make([]string, 0, len(req.Racks))
	plus := make([]string, 0, len(req.Racks))
	for _, rackReq := range req.Racks {
		for _, item := range rackReq.Items {
			switch {
			case item.Barcode != "":
				codes = append(codes, item.Barcode)
			case item.PLU != "":
				plus = append(plus, item.PLU)
			}
		}
	}

	found, err := s.ProductRepository.FindByBarcodesOrPLUs(ctx, tx, coordinator.CoorSesiID, codes, plus)
	if err != nil {
		return dto.UploadResponse{}, err
	}

	byBarcode := make(map[string]*model.Product, len(found))
	byPLU := make(map[string]*model.Product, len(found))
	for _, product := range found {
		byPLU[product.ProductPLU] = product
		for _, code := range product.ProductBarcodes {
			if _, exists := byBarcode[code]; !exists {
				byBarcode[code] = product
			}
		}
	}

	// Pass 2 keeps the original per-item order so validation errors surface at
	// the same item they used to.
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
			product, err := resolveFromBatch(byBarcode, byPLU, item)
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
				PLU:       product.ProductPLU,
				Barcode:   soBarcode,
				Name:      product.ProductName,
				Dept:      product.ProductDepartmentCode,
				Qty:       *item.Quantity,
				BuyPrice:  product.ProductBuyPrice,
				SellPrice: product.ProductSellPrice,
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

	// Rows already sitting in the claimed racks join this upload before the
	// batch is written, so a rack submitted with a partial count stays whole.
	if _, err := s.StockOpnameRepository.AdoptUnownedRows(ctx, tx, upload.UploadID, rakIDs); err != nil {
		return dto.UploadResponse{}, err
	}

	ownerID := upload.UploadID
	allResults := make([]*model.StockOpname, 0, totalItems)
	for _, payload := range payloads {
		if err := s.RackRepository.SetSnapshot(ctx, tx, payload.rakID, payload.snapshot); err != nil {
			return dto.UploadResponse{}, err
		}
		for _, result := range payload.results {
			result.SoUploadID = &ownerID
		}
		allResults = append(allResults, payload.results...)
	}
	// One statement for the whole batch instead of one upsert per item.
	if err := s.StockOpnameRepository.SaveMany(ctx, tx, allResults); err != nil {
		return dto.UploadResponse{}, err
	}

	// Counts come from ownership, so adoption and take-over are included even
	// though this request never scanned those rows.
	itemCount, rackCount, err := s.StockOpnameRepository.CountByUpload(ctx, tx, upload.UploadID)
	if err != nil {
		return dto.UploadResponse{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return dto.UploadResponse{}, err
	}

	return dto.UploadResponse{
		UploadID:  upload.UploadID,
		RackCount: rackCount,
		ItemCount: itemCount,
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
			UploadID:     upload.UploadID,
			UploadNumber: upload.UploadNumber,
			RackCount:    upload.RackCount,
			ItemCount:    upload.ItemCount,
			CreatedAt:    upload.CreatedAt.Format(time.RFC3339),
		})
	}
	return responses, nil
}

// FindSummaryByCoordinator implements [UploadService].
func (s *UploadServiceImpl) FindSummaryByCoordinator(ctx context.Context, coordinatorId int) ([]*model.UploadReportRow, error) {
	return s.UploadRepository.FindSummaryByCoordinator(ctx, coordinatorId)
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
				PLU:       item.PLU,
				Barcode:   item.Barcode,
				Name:      item.Name,
				Dept:      item.Dept,
				Qty:       item.Qty,
				BuyPrice:  item.BuyPrice,
				SellPrice: item.SellPrice,
				PerCarton: item.PerCarton,
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
