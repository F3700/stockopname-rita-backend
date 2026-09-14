package repository

import (
	"context"
	"errors"
	"stockopname-rita-backend/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DepartmentRepositoryImpl struct {
	pool *pgxpool.Pool
}

func NewDepartmentRepositoryImpl(pool *pgxpool.Pool) DepartmentRepository {
	return &DepartmentRepositoryImpl{
		pool: pool,
	}
}

// Delete implements [DepartmentRepository].
func (d *DepartmentRepositoryImpl) Delete(ctx context.Context, tx pgx.Tx, id int) error {
	const SQL = `DELETE FROM department WHERE department_id = $1`

	res, err := tx.Exec(ctx, SQL, id)
	if err != nil {
		return model.MapPgError("Department", err)
	}
	if res.RowsAffected() == 0 {
		return &model.NotFoundError{Resource: "Department", ID: id}
	}
	return nil
}

// FindAll implements [DepartmentRepository].
func (d *DepartmentRepositoryImpl) FindAll(ctx context.Context) ([]*model.Department, error) {
	const SQL = `
		SELECT
			department_id,
			department_code,
			department_name,
			department_desc
		FROM department
	`
	rows, err := d.pool.Query(ctx, SQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var departments []*model.Department
	for rows.Next() {
		var department model.Department
		err := rows.Scan(
			&department.DepartmentID,
			&department.DepartmentCode,
			&department.DepartmentName,
			&department.DepartmentDesc,
		)
		if err != nil {
			return nil, err
		}
		departments = append(departments, &department)
	}

	return departments, nil
}

// FindAllInPageSearch implements [DepartmentRepository].
func (d *DepartmentRepositoryImpl) FindAllInPageSearch(ctx context.Context, limit int, offset int, search string) ([]*model.Department, int, error) {
	const SQL = `
		SELECT
			department_id,
			department_code,
			department_name,
			department_desc
		FROM department
		WHERE department_code ILIKE $1 OR department_name ILIKE $1
		ORDER BY department_id DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := d.pool.Query(ctx, SQL, "%"+search+"%", limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var departments []*model.Department
	for rows.Next() {
		var department model.Department
		err := rows.Scan(
			&department.DepartmentID,
			&department.DepartmentCode,
			&department.DepartmentName,
			&department.DepartmentDesc,
		)
		if err != nil {
			return nil, 0, err
		}
		departments = append(departments, &department)
	}

	var total int
	countSQL := `SELECT COUNT(*) FROM department WHERE department_code ILIKE $1 OR department_name ILIKE $1`
	if err := d.pool.QueryRow(ctx, countSQL, "%"+search+"%").Scan(&total); err != nil {
		return nil, 0, err
	}

	return departments, total, nil
}

// FindById implements [DepartmentRepository].
func (d *DepartmentRepositoryImpl) FindById(ctx context.Context, tx pgx.Tx, id int) (*model.Department, error) {
	const SQL = `
		SELECT
			department_id,
			department_code,
			department_name,
			department_desc
		FROM department
		WHERE department_id = $1
	`

	var department model.Department
	err := tx.QueryRow(ctx, SQL, id).Scan(
		&department.DepartmentID,
		&department.DepartmentCode,
		&department.DepartmentName,
		&department.DepartmentDesc,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &model.NotFoundError{Resource: "Department", ID: id}
		}
		return nil, err
	}
	return &department, nil
}

// Save implements [DepartmentRepository].
func (d *DepartmentRepositoryImpl) Save(ctx context.Context, tx pgx.Tx, department *model.Department) error {
	const SQL = `
		INSERT INTO department (department_code, department_name, department_desc)
		VALUES ($1, $2, $3)
		RETURNING department_id
	`
	err := tx.QueryRow(ctx, SQL, department.DepartmentCode, department.DepartmentName, department.DepartmentDesc).Scan(&department.DepartmentID)
	if err != nil {
		return model.MapPgError("Department", err)
	}
	return nil
}

// Update implements [DepartmentRepository].
func (d *DepartmentRepositoryImpl) Update(ctx context.Context, tx pgx.Tx, department *model.Department) error {
	const SQL = `
		UPDATE department
		SET department_code = $1, department_name = $2, department_desc = $3
		WHERE department_id = $4
	`
	_, err := tx.Exec(ctx, SQL, department.DepartmentCode, department.DepartmentName, department.DepartmentDesc, department.DepartmentID)
	if err != nil {
		return err
	}
	return nil
}
