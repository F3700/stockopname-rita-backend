package service

import (
	"context"
	"fmt"
	"stockopname-rita-backend/internal/dto"
	"stockopname-rita-backend/internal/model"
	"stockopname-rita-backend/internal/repository"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProductServiceImpl struct {
	ProductRepository repository.ProductRepository
	Pool              *pgxpool.Pool
	Validator         *validator.Validate
}

func NewProductService(productRepository repository.ProductRepository, pool *pgxpool.Pool, validator *validator.Validate) ProductService {
	return &ProductServiceImpl{
		ProductRepository: productRepository,
		Pool:              pool,
		Validator:         validator,
	}
}

// Create implements [ProductService].
func (p *ProductServiceImpl) Create(ctx context.Context, req dto.ProductCreateRequest) (dto.ProductCreateResponse, error) {
	if err := p.Validator.Struct(req); err != nil {
		return dto.ProductCreateResponse{}, err
	}

	tx, err := p.Pool.Begin(ctx)
	if err != nil {
		return dto.ProductCreateResponse{}, err
	}
	defer tx.Rollback(ctx)

	product := model.Product{
		ProductBarcode:      req.Barcode,
		ProductName:         req.Name,
		ProductBuyPrice:     req.BuyPrice,
		ProductSellPrice:    req.SellPrice,
		ProductCategoryID:   req.CategoryID,
		ProductDepartmentID: req.DepartmentID,
	}

	if err := p.ProductRepository.Save(ctx, tx, &product); err != nil {
		return dto.ProductCreateResponse{}, fmt.Errorf("save product: %w", err)
	}

	productResult, err := p.ProductRepository.FindById(ctx, tx, product.ProductID)
	if err != nil {
		return dto.ProductCreateResponse{}, fmt.Errorf("find created product id %d: %w", product.ProductID, err)
	}

	response := dto.ProductCreateResponse{
		Id:             productResult.ProductID,
		Barcode:        productResult.ProductBarcode,
		Name:           productResult.ProductName,
		BuyPrice:       productResult.ProductBuyPrice,
		SellPrice:      productResult.ProductSellPrice,
		DateCreated:    productResult.ProductCreatedat.Format(time.RFC3339),
		DateUpdated:    productResult.ProductUpdatedat.Format(time.RFC3339),
		CategoryName:   productResult.ProductCategoryName,
		DepartmentCode: productResult.ProductDepartmentCode,
	}

	if err := tx.Commit(ctx); err != nil {
		return dto.ProductCreateResponse{}, err
	}

	return response, nil
}

// Delete implements [ProductService].
func (p *ProductServiceImpl) Delete() {
	panic("unimplemented")
}

// FindAll implements [ProductService].
func (p *ProductServiceImpl) FindAll() {
	panic("unimplemented")
}

// Update implements [ProductService].
func (p *ProductServiceImpl) Update() {
	panic("unimplemented")
}
