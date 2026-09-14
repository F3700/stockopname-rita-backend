package service

import (
	"context"
	"fmt"
	"math"
	"stockopname-rita-backend/internal/dto"
	"stockopname-rita-backend/internal/model"
	"stockopname-rita-backend/internal/repository"
	"time"

	"github.com/go-playground/validator/v10"
)

type ProductServiceImpl struct {
	ProductRepository        repository.ProductRepository
	DeletedProductRepository repository.DeletedProductRepository
	Pool                     repository.DBPool
	Validator                *validator.Validate
}

func NewProductService(productRepository repository.ProductRepository, deletedProductRepository repository.DeletedProductRepository, pool repository.DBPool, validator *validator.Validate) ProductService {
	return &ProductServiceImpl{
		ProductRepository:        productRepository,
		DeletedProductRepository: deletedProductRepository,
		Pool:                     pool,
		Validator:                validator,
	}
}

// Create implements [ProductService].
func (p *ProductServiceImpl) Create(ctx context.Context, req dto.ProductCreateRequest) (dto.ProductResponse, error) {
	if err := p.Validator.Struct(req); err != nil {
		return dto.ProductResponse{}, err
	}

	tx, err := p.Pool.Begin(ctx)
	if err != nil {
		return dto.ProductResponse{}, err
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
		return dto.ProductResponse{}, fmt.Errorf("save product: %w", err)
	}

	productResult, err := p.ProductRepository.FindById(ctx, tx, product.ProductID)
	if err != nil {
		return dto.ProductResponse{}, fmt.Errorf("find created product id %d: %w", product.ProductID, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return dto.ProductResponse{}, err
	}

	return dto.ProductResponse{
		Id:             productResult.ProductID,
		Barcode:        productResult.ProductBarcode,
		Name:           productResult.ProductName,
		BuyPrice:       productResult.ProductBuyPrice,
		SellPrice:      productResult.ProductSellPrice,
		DateCreated:    productResult.ProductCreatedat.Format(time.RFC3339),
		DateUpdated:    productResult.ProductUpdatedat.Format(time.RFC3339),
		CategoryName:   productResult.ProductCategoryName,
		DepartmentCode: productResult.ProductDepartmentCode,
	}, nil
}

// Delete implements [ProductService].
func (p *ProductServiceImpl) Delete(ctx context.Context, id int) error {
	tx, err := p.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := p.ProductRepository.Delete(ctx, tx, id); err != nil {
		return fmt.Errorf("delete product id %d: %w", id, err)
	}

	//tambahin id yang udah di delete ke table deleted product pake repository deleted product
	if err := p.DeletedProductRepository.Save(ctx, tx, id); err != nil {
		return fmt.Errorf("save deleted product: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// FindAll implements [ProductService].
func (p *ProductServiceImpl) FindAll(ctx context.Context, pagination *dto.Pagination, search string) ([]dto.ProductResponse, error) {
	offset := (pagination.Page - 1) * pagination.Limit

	products, total, err := p.ProductRepository.FindAllInPageSearch(ctx, pagination.Limit, offset, search)
	if err != nil {
		return nil, fmt.Errorf("find all products: %w", err)
	}

	var responses []dto.ProductResponse
	for _, product := range products {
		responses = append(responses, dto.ProductResponse{
			Id:             product.ProductID,
			Barcode:        product.ProductBarcode,
			Name:           product.ProductName,
			BuyPrice:       product.ProductBuyPrice,
			SellPrice:      product.ProductSellPrice,
			DateCreated:    product.ProductCreatedat.Format(time.RFC3339),
			DateUpdated:    product.ProductUpdatedat.Format(time.RFC3339),
			CategoryName:   product.ProductCategoryName,
			DepartmentCode: product.ProductDepartmentCode,
		})
	}

	pagination.TotalItems = total
	pagination.TotalPages = int(math.Ceil(float64(total) / float64(pagination.Limit)))

	return responses, nil
}

// FindAllUpdatedAfter implements [ProductService].
func (p *ProductServiceImpl) FindAllUpdatedAfter(ctx context.Context, pagination *dto.Pagination, date time.Time) ([]dto.ProductResponse, error) {
	offset := (pagination.Page - 1) * pagination.Limit

	products, total, err := p.ProductRepository.FindAllUpdatedAfter(ctx, date, pagination.Limit, offset)
	if err != nil {
		return nil, fmt.Errorf("find all updated after %v: %w", date, err)
	}

	var responses []dto.ProductResponse
	for _, product := range products {
		responses = append(responses, dto.ProductResponse{
			Id:             product.ProductID,
			Barcode:        product.ProductBarcode,
			Name:           product.ProductName,
			BuyPrice:       product.ProductBuyPrice,
			SellPrice:      product.ProductSellPrice,
			DateCreated:    product.ProductCreatedat.Format(time.RFC3339),
			DateUpdated:    product.ProductUpdatedat.Format(time.RFC3339),
			CategoryName:   product.ProductCategoryName,
			DepartmentCode: product.ProductDepartmentCode,
		})
	}

	pagination.TotalItems = total
	pagination.TotalPages = int(math.Ceil(float64(total) / float64(pagination.Limit)))

	return responses, nil
}

// Update implements [ProductService].
func (p *ProductServiceImpl) Update(ctx context.Context, id int, req dto.ProductUpdateRequest) (dto.ProductResponse, error) {
	tx, err := p.Pool.Begin(ctx)
	if err != nil {
		return dto.ProductResponse{}, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	product, err := p.ProductRepository.FindById(ctx, tx, id)
	if err != nil {
		return dto.ProductResponse{}, fmt.Errorf("find product by id %d: %w", id, err)
	}

	if req.Barcode != nil {
		product.ProductBarcode = *req.Barcode
	}
	if req.Name != nil {
		product.ProductName = *req.Name
	}
	if req.BuyPrice != nil {
		product.ProductBuyPrice = *req.BuyPrice
	}
	if req.SellPrice != nil {
		product.ProductSellPrice = *req.SellPrice
	}
	if req.CategoryID != nil {
		product.ProductCategoryID = *req.CategoryID
	}
	if req.DepartmentID != nil {
		product.ProductDepartmentID = *req.DepartmentID
	}

	if err = p.ProductRepository.Update(ctx, tx, product); err != nil {
		return dto.ProductResponse{}, fmt.Errorf("update product %d: %w", id, err)
	}

	updatedProduct, err := p.ProductRepository.FindById(ctx, tx, id)
	if err != nil {
		return dto.ProductResponse{}, fmt.Errorf("find updated product by id %d: %w", id, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return dto.ProductResponse{}, fmt.Errorf("commit transaction: %w", err)
	}

	response := dto.ProductResponse{
		Id:             updatedProduct.ProductID,
		Barcode:        updatedProduct.ProductBarcode,
		Name:           updatedProduct.ProductName,
		BuyPrice:       updatedProduct.ProductBuyPrice,
		SellPrice:      updatedProduct.ProductSellPrice,
		DateCreated:    updatedProduct.ProductCreatedat.Format(time.RFC3339),
		DateUpdated:    updatedProduct.ProductUpdatedat.Format(time.RFC3339),
		CategoryName:   updatedProduct.ProductCategoryName,
		DepartmentCode: updatedProduct.ProductDepartmentCode,
	}

	return response, nil
}

// FindAllLastSession implements [ProductService].
func (p *ProductServiceImpl) FindAllLastSession(ctx context.Context, limit int) ([]dto.ProductLastSessionResponse, error) {
	results, err := p.ProductRepository.FindAllLastSession(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("find all last session: %w", err)
	}

	var responses []dto.ProductLastSessionResponse
	for _, item := range results {
		var lastSessionDate string
		if item.LastSessionDate != nil {
			lastSessionDate = item.LastSessionDate.Format(time.RFC3339)
		}

		responses = append(responses, dto.ProductLastSessionResponse{
			Id:              item.Id,
			Barcode:         item.Barcode,
			Name:            item.Name,
			LastSessionDate: lastSessionDate,
			LastSessionCode: item.LastSessionCode,
		})
	}

	return responses, nil
}
