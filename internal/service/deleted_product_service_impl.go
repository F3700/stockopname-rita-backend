package service

import (
	"context"
	"fmt"
	"stockopname-rita-backend/internal/dto"
	"stockopname-rita-backend/internal/repository"
	"time"
)

type DeletedProductServiceImpl struct {
	DeletedProductRepository repository.DeletedProductRepository
}

func NewDeletedProductService(deletedProductRepository repository.DeletedProductRepository) DeletedProductService {
	return &DeletedProductServiceImpl{
		DeletedProductRepository: deletedProductRepository,
	}
}

// Delete implements [DeletedProductService].
func (d *DeletedProductServiceImpl) Delete(ctx context.Context) error {
	err := d.DeletedProductRepository.Delete(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete deleted products: %w", err)
	}

	return nil
}

// FindAll implements [DeletedProductService].
func (d *DeletedProductServiceImpl) FindAll(ctx context.Context, date *time.Time) ([]dto.DeletedProductResponse, error) {
	var deletedProducts []dto.DeletedProductResponse
	if date != nil {
		products, err := d.DeletedProductRepository.FindAllUpdatedAfter(ctx, *date)
		if err != nil {
			return nil, fmt.Errorf("failed to find all deleted products updated after %v: %w", *date, err)
		}

		for _, product := range products {
			response := dto.DeletedProductResponse{
				ProductID: product.ProductID,
			}
			deletedProducts = append(deletedProducts, response)
		}
		return deletedProducts, nil
	}

	products, err := d.DeletedProductRepository.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find all deleted products: %w", err)
	}
	for _, product := range products {
		response := dto.DeletedProductResponse{
			ProductID: product.ProductID,
		}
		deletedProducts = append(deletedProducts, response)
	}
	return deletedProducts, nil
}
