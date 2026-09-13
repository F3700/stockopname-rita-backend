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

type DepartmentServiceImpl struct {
	DepartmentRepository repository.DepartmentRepository
	Pool                 *pgxpool.Pool
	Validator            *validator.Validate
}

func NewDepartmentService(departmentRepository repository.DepartmentRepository, pool *pgxpool.Pool, validator *validator.Validate) DepartmentService {
	return &DepartmentServiceImpl{
		DepartmentRepository: departmentRepository,
		Pool:                 pool,
		Validator:            validator,
	}
}

// Create implements [DepartmentService].
func (d *DepartmentServiceImpl) Create(ctx context.Context, req dto.DepartmentCreateRequest) (dto.DepartmentResponse, error) {
	if err := d.Validator.Struct(req); err != nil {
		return dto.DepartmentResponse{}, fmt.Errorf("validation error: %w", err)
	}

	tx, err := d.Pool.Begin(ctx)
	if err != nil {
		return dto.DepartmentResponse{}, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	department := model.Department{
		DepartmentName: req.Name,
		DepartmentCode: req.Code,
		DepartmentDesc: req.Description,
	}

	if err = d.DepartmentRepository.Save(ctx, tx, &department); err != nil {
		return dto.DepartmentResponse{}, fmt.Errorf("failed to save department: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return dto.DepartmentResponse{}, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return dto.DepartmentResponse{
		ID:          department.DepartmentID,
		Name:        department.DepartmentName,
		Code:        department.DepartmentCode,
		Description: department.DepartmentDesc,
	}, nil
}

// Delete implements [DepartmentService].
func (d *DepartmentServiceImpl) Delete(ctx context.Context, id int) error {
	tx, err := d.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if err = d.DepartmentRepository.Delete(ctx, tx, id); err != nil {
		return fmt.Errorf("failed to delete department: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// FindAll implements [DepartmentService].
func (d *DepartmentServiceImpl) FindAll(ctx context.Context, pagination *dto.Pagination, search string) ([]dto.DepartmentResponse, error) {
	offset := (pagination.Page - 1) * pagination.Limit

	departments, total, err := d.DepartmentRepository.FindAllInPageSearch(ctx, pagination.Limit, offset, search)
	if err != nil {
		return nil, fmt.Errorf("failed to find all departments: %w", err)
	}

	var responses []dto.DepartmentResponse
	for _, department := range departments {
		responses = append(responses, dto.DepartmentResponse{
			ID:          department.DepartmentID,
			Name:        department.DepartmentName,
			Code:        department.DepartmentCode,
			Description: department.DepartmentDesc,
		})
	}

	pagination.TotalItems = total
	pagination.TotalPages = int(math.Ceil(float64(total) / float64(pagination.Limit)))

	return responses, nil
}

// Update implements [DepartmentService].
func (d *DepartmentServiceImpl) Update(ctx context.Context, id int, req dto.DepartmentUpdateRequest) (dto.DepartmentResponse, error) {
	tx, err := d.Pool.Begin(ctx)
	if err != nil {
		return dto.DepartmentResponse{}, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	department, err := d.DepartmentRepository.FindById(ctx, tx, id)
	if err != nil {
		return dto.DepartmentResponse{}, fmt.Errorf("failed to find department: %w", err)
	}

	if req.Code != nil {
		department.DepartmentCode = *req.Code
	}
	if req.Name != nil {
		department.DepartmentName = *req.Name
	}
	if req.Description != nil {
		department.DepartmentDesc = *req.Description
	}

	if err = d.DepartmentRepository.Update(ctx, tx, department); err != nil {
		return dto.DepartmentResponse{}, fmt.Errorf("failed to update department: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return dto.DepartmentResponse{}, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return dto.DepartmentResponse{
		ID:          department.DepartmentID,
		Name:        department.DepartmentName,
		Code:        department.DepartmentCode,
		Description: department.DepartmentDesc,
	}, nil
}
