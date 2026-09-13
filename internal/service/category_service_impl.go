package service

import (
	"context"
	"fmt"
	"math"
	"stockopname-rita-backend/internal/dto"
	"stockopname-rita-backend/internal/model"
	"stockopname-rita-backend/internal/repository"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CategoryServiceImpl struct {
	CategoryRepository repository.CategoryRepository
	Pool               *pgxpool.Pool
	Validator          *validator.Validate
}

func NewCategoryService(categoryRepository repository.CategoryRepository, pool *pgxpool.Pool, validator *validator.Validate) CategoryService {
	return &CategoryServiceImpl{
		CategoryRepository: categoryRepository,
		Pool:               pool,
		Validator:          validator,
	}
}

// Create implements [CategoryService].
func (c *CategoryServiceImpl) Create(ctx context.Context, req dto.CategoryCreateRequest) (dto.CategoryResponse, error) {
	if err := c.Validator.Struct(req); err != nil {
		return dto.CategoryResponse{}, fmt.Errorf("validation error: %w", err)
	}

	tx, err := c.Pool.Begin(ctx)
	if err != nil {
		return dto.CategoryResponse{}, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	category := model.Category{
		CategoryName:        req.CategoryName,
		CategoryDescription: req.CategoryDescription,
	}

	if err = c.CategoryRepository.Save(ctx, tx, &category); err != nil {
		return dto.CategoryResponse{}, fmt.Errorf("failed to save category: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return dto.CategoryResponse{}, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return dto.CategoryResponse{
		CategoryID:          category.CategoryID,
		CategoryName:        category.CategoryName,
		CategoryDescription: category.CategoryDescription,
	}, nil
}

// Delete implements [CategoryService].
func (c *CategoryServiceImpl) Delete(ctx context.Context, id int) error {
	tx, err := c.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if err = c.CategoryRepository.Delete(ctx, tx, id); err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// FindAll implements [CategoryService].
func (c *CategoryServiceImpl) FindAll(ctx context.Context, pagination *dto.Pagination, search string) ([]dto.CategoryResponse, error) {
	offset := (pagination.Page - 1) * pagination.Limit

	categories, total, err := c.CategoryRepository.FindAllInPageSearch(ctx, pagination.Limit, offset, search)
	if err != nil {
		return nil, fmt.Errorf("failed to find all categories: %w", err)
	}

	var responses []dto.CategoryResponse
	for _, category := range categories {
		responses = append(responses, dto.CategoryResponse{
			CategoryID:          category.CategoryID,
			CategoryName:        category.CategoryName,
			CategoryDescription: category.CategoryDescription,
		})
	}

	pagination.TotalItems = total
	pagination.TotalPages = int(math.Ceil(float64(total) / float64(pagination.Limit)))

	return responses, nil
}

// Update implements [CategoryService].
func (c *CategoryServiceImpl) Update(ctx context.Context, id int, req dto.CategoryUpdateRequest) (dto.CategoryResponse, error) {
	tx, err := c.Pool.Begin(ctx)
	if err != nil {
		return dto.CategoryResponse{}, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	category, err := c.CategoryRepository.FindById(ctx, tx, id)
	if err != nil {
		return dto.CategoryResponse{}, fmt.Errorf("failed to find category: %w", err)
	}

	if req.CategoryName != nil {
		category.CategoryName = *req.CategoryName
	}
	if req.CategoryDescription != nil {
		category.CategoryDescription = *req.CategoryDescription
	}

	if err = c.CategoryRepository.Update(ctx, tx, category); err != nil {
		return dto.CategoryResponse{}, fmt.Errorf("failed to update category: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return dto.CategoryResponse{}, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return dto.CategoryResponse{
		CategoryID:          category.CategoryID,
		CategoryName:        category.CategoryName,
		CategoryDescription: category.CategoryDescription,
	}, nil
}
