package service

import (
	"context"
	"fmt"
	"stockopname-rita-backend/internal/dto"
	"stockopname-rita-backend/internal/model"
	"stockopname-rita-backend/internal/repository"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DepartmentServiceImpl struct {
	departmentRepository repository.DepartmentRepository
	pool                 *pgxpool.Pool
	validator            *validator.Validate
}

func NewDepartmentService(departmentRepository repository.DepartmentRepository, pool *pgxpool.Pool, validator *validator.Validate) DepartmentService {
	return &DepartmentServiceImpl{
		departmentRepository: departmentRepository,
		pool:                 pool,
		validator:            validator,
	}
}

// Create implements [DepartmentService].
func (d *DepartmentServiceImpl) Create(ctx context.Context, req dto.DepartmentCreateRequest) (dto.DepartmentResponse, error) {
	if err := d.validator.Struct(req); err != nil {
		return dto.DepartmentResponse{}, fmt.Errorf("validation error: %w", err)
	}

	tx, err := d.pool.Begin(ctx)
	if err != nil {
		return dto.DepartmentResponse{}, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	department := model.Department{
		DepartmentName: req.Name,
		DepartmentCode: req.Code,
		DepartmentDesc: req.Description,
	}

	if err = d.departmentRepository.Save(ctx, tx, &department); err != nil {
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
	tx, err := d.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if err = d.departmentRepository.Delete(ctx, tx, id); err != nil {
		return fmt.Errorf("failed to delete department: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// FindAll implements [DepartmentService].
func (d *DepartmentServiceImpl) FindAll(ctx context.Context) ([]dto.DepartmentResponse, error) {
	departments, err := d.departmentRepository.FindAll(ctx)
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
	return responses, nil
}

// Update implements [DepartmentService].
func (d *DepartmentServiceImpl) Update(ctx context.Context, id int, req dto.DepartmentUpdateRequest) (dto.DepartmentResponse, error) {
	tx, err := d.pool.Begin(ctx)
	if err != nil {
		return dto.DepartmentResponse{}, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	department, err := d.departmentRepository.FindById(ctx, tx, id)
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

	if err = d.departmentRepository.Update(ctx, tx, department); err != nil {
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
