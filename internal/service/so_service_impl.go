package service

import (
	"context"
	"math"
	"stockopname-rita-backend/internal/dto"
	"stockopname-rita-backend/internal/model"
	"stockopname-rita-backend/internal/repository"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5"
)

type StockOpnameServiceImpl struct {
	Pool                  repository.DBPool
	StockOpnameRepository repository.StockOpnameRepository
	ProductRepository     repository.ProductRepository
	Validator             *validator.Validate
}

func NewStockOpnameService(stockOpnameRepository repository.StockOpnameRepository, productRepository repository.ProductRepository, pool repository.DBPool, validate *validator.Validate) StockOpnameService {
	return &StockOpnameServiceImpl{
		Pool:                  pool,
		StockOpnameRepository: stockOpnameRepository,
		ProductRepository:     productRepository,
		Validator:             validate,
	}
}

// resolveProduct finds the master product by scanned barcode or PLU.
func (s *StockOpnameServiceImpl) resolveProduct(ctx context.Context, tx pgx.Tx, barcode string, plu string) (*model.Product, error) {
	if barcode != "" {
		return s.ProductRepository.FindByBarcode(ctx, tx, barcode)
	}
	if plu != "" {
		return s.ProductRepository.FindByPLU(ctx, tx, plu)
	}
	return nil, &model.ValidationError{Detail: "barcode or plu is required"}
}

// snapshot builds the immutable history row from the resolved product.
// soBarcode is the actually scanned barcode (may differ from the
// product's primary barcode); it falls back to the primary barcode
// (or "" for products without any barcode) on PLU-based input.
func snapshot(quantity int, rakID int, product *model.Product, scannedBarcode string) *model.StockOpname {
	soBarcode := scannedBarcode
	if soBarcode == "" {
		soBarcode = product.PrimaryBarcode()
	}
	return &model.StockOpname{
		StockOpnameQuantity: quantity,
		StockOpnameRakID:    rakID,
		SoProductPLU:        product.ProductPLU,
		SoProductName:       product.ProductName,
		SoBarcode:           soBarcode,
		SoBuyPrice:          product.ProductBuyPrice,
		SoSellPrice:         product.ProductSellPrice,
	}
}

// Create implements [StockOpnameService].
func (s *StockOpnameServiceImpl) Create(ctx context.Context, req dto.CreateStockOpnameRequest) (dto.StockOpnameResponse, error) {
	if err := s.Validator.Struct(req); err != nil {
		return dto.StockOpnameResponse{}, err
	}
	if req.Barcode == "" && req.PLU == "" {
		return dto.StockOpnameResponse{}, &model.ValidationError{Detail: "barcode or plu is required"}
	}

	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return dto.StockOpnameResponse{}, err
	}
	defer tx.Rollback(ctx)

	product, err := s.resolveProduct(ctx, tx, req.Barcode, req.PLU)
	if err != nil {
		return dto.StockOpnameResponse{}, err
	}

	modelSO := snapshot(req.Quantity, req.RakID, product, req.Barcode)

	if err := s.StockOpnameRepository.Save(ctx, tx, modelSO); err != nil {
		return dto.StockOpnameResponse{}, err
	}

	modelSummary, err := s.StockOpnameRepository.FindById(ctx, tx, modelSO.StockOpnameID)
	if err != nil {
		return dto.StockOpnameResponse{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return dto.StockOpnameResponse{}, err
	}

	return dto.StockOpnameResponse{
		Id:              modelSummary.Id,
		Barcode:         modelSummary.Barcode,
		Name:            modelSummary.Name,
		Quantity:        modelSummary.Quantity,
		RackName:        modelSummary.RackName,
		InspectorCode:   modelSummary.InspectorCode,
		CoordinatorCode: modelSummary.CoordinatorCode,
		UpdatedAt:       modelSummary.UpdatedAt.Format(time.RFC3339),
	}, nil
}

// Delete implements [StockOpnameService].
func (s *StockOpnameServiceImpl) Delete(ctx context.Context, id int) error {
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := s.StockOpnameRepository.Delete(ctx, tx, id); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// FindAll implements [StockOpnameService].
func (s *StockOpnameServiceImpl) FindAll(ctx context.Context, pagination *dto.Pagination, search string, coorId *int, sesiId *int) ([]dto.StockOpnameResponse, error) {
	offset := (pagination.Page - 1) * pagination.Limit

	stockOpnames, total, err := s.StockOpnameRepository.FindAll(ctx, pagination.Limit, offset, search, sesiId, coorId)
	if err != nil {
		return nil, err
	}

	var responses []dto.StockOpnameResponse
	for _, so := range stockOpnames {
		responses = append(responses, dto.StockOpnameResponse{
			Id:              so.Id,
			Barcode:         so.Barcode,
			Name:            so.Name,
			Quantity:        so.Quantity,
			RackName:        so.RackName,
			InspectorCode:   so.InspectorCode,
			CoordinatorCode: so.CoordinatorCode,
			UpdatedAt:       so.UpdatedAt.Format(time.RFC3339),
		})
	}

	pagination.TotalItems = total
	pagination.TotalPages = int(math.Ceil(float64(total) / float64(pagination.Limit)))

	return responses, nil
}

// FindById implements [StockOpnameService].
func (s *StockOpnameServiceImpl) FindById(ctx context.Context, id int) (dto.StockOpnameResponse, error) {
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return dto.StockOpnameResponse{}, err
	}
	defer tx.Rollback(ctx)

	modelSummary, err := s.StockOpnameRepository.FindById(ctx, tx, id)
	if err != nil {
		return dto.StockOpnameResponse{}, err
	}

	return dto.StockOpnameResponse{
		Id:              modelSummary.Id,
		Barcode:         modelSummary.Barcode,
		Name:            modelSummary.Name,
		Quantity:        modelSummary.Quantity,
		RackName:        modelSummary.RackName,
		InspectorCode:   modelSummary.InspectorCode,
		CoordinatorCode: modelSummary.CoordinatorCode,
		UpdatedAt:       modelSummary.UpdatedAt.Format(time.RFC3339),
	}, nil
}

// Update implements [StockOpnameService].
func (s *StockOpnameServiceImpl) Update(ctx context.Context, id int, req dto.UpdateStockOpnameRequest) (dto.StockOpnameResponse, error) {
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return dto.StockOpnameResponse{}, err
	}
	defer tx.Rollback(ctx)

	modelSO := &model.StockOpname{
		StockOpnameID:       id,
		StockOpnameQuantity: req.Quantity,
	}

	if err := s.StockOpnameRepository.Update(ctx, tx, modelSO); err != nil {
		return dto.StockOpnameResponse{}, err
	}

	result, err := s.StockOpnameRepository.FindById(ctx, tx, id)
	if err != nil {
		return dto.StockOpnameResponse{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return dto.StockOpnameResponse{}, err
	}

	return dto.StockOpnameResponse{
		Id:              result.Id,
		Barcode:         result.Barcode,
		Name:            result.Name,
		Quantity:        result.Quantity,
		RackName:        result.RackName,
		InspectorCode:   result.InspectorCode,
		CoordinatorCode: result.CoordinatorCode,
		UpdatedAt:       result.UpdatedAt.Format(time.RFC3339),
	}, nil
}

func (s *StockOpnameServiceImpl) CreateByRack(ctx context.Context, req dto.CreateStockOpnameByRackRequest) error {
	if err := s.Validator.Struct(req); err != nil {
		return err
	}
	for _, item := range req.Items {
		if item.Barcode == "" && item.PLU == "" {
			return &model.ValidationError{Detail: "barcode or plu is required"}
		}
	}

	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, item := range req.Items {
		product, err := s.resolveProduct(ctx, tx, item.Barcode, item.PLU)
		if err != nil {
			return err
		}
		modelSO := snapshot(item.Quantity, req.RackID, product, item.Barcode)
		if err := s.StockOpnameRepository.Save(ctx, tx, modelSO); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

// FindAllForExport implements [StockOpnameService].
func (s *StockOpnameServiceImpl) FindAllForExport(ctx context.Context, sesiId int) ([]dto.StockOpnameExportResponse, error) {
	exports, err := s.StockOpnameRepository.FindAllForExport(ctx, sesiId)
	if err != nil {
		return nil, err
	}

	var responses []dto.StockOpnameExportResponse
	for _, export := range exports {
		responses = append(responses, dto.StockOpnameExportResponse{
			Id:              export.Id,
			Barcode:         export.Barcode,
			Name:            export.Name,
			BuyPrice:        export.BuyPrice,
			SellPrice:       export.SellPrice,
			Quantity:        export.Quantity,
			RackName:        export.RackName,
			InspectorCode:   export.InspectorCode,
			CoordinatorCode: export.CoordinatorCode,
			UpdatedAt:       export.UpdatedAt.Format(time.RFC3339),
		})
	}

	return responses, nil
}
