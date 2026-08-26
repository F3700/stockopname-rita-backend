package repository

import (
	"context"
	"fmt"
	"stockopname-rita-backend/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CategoryRepositoryImpl struct {
	pool *pgxpool.Pool
}

func NewCategoryRepositoryImpl(pool *pgxpool.Pool) CategoryRepository {
	return &CategoryRepositoryImpl{
		pool: pool,
	}
}

// Delete implements [CategoryRepository].
func (c *CategoryRepositoryImpl) Delete(ctx context.Context, tx pgx.Tx, id int) error {
	const SQL = `DELETE FROM category WHERE category_id = $1`

	res, err := tx.Exec(ctx, SQL, id)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("category with ID %d not found", id)
	}

	return nil
}

// FindAll implements [CategoryRepository].
func (c *CategoryRepositoryImpl) FindAll(ctx context.Context) ([]*model.Category, error) {
	const SQL = `
		SELECT
			category_id,
			category_name,
			category_desc
		FROM category
	`
	rows, err := c.pool.Query(ctx, SQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []*model.Category
	for rows.Next() {
		var category model.Category
		err := rows.Scan(
			&category.CategoryID,
			&category.CategoryName,
			&category.CategoryDescription,
		)
		if err != nil {
			return nil, err
		}
		categories = append(categories, &category)
	}

	return categories, nil
}

// FindById implements [CategoryRepository].
func (c *CategoryRepositoryImpl) FindById(ctx context.Context, tx pgx.Tx, id int) (*model.Category, error) {
	const SQL = `
		SELECT
			category_id,
			category_name,
			category_desc
		FROM category
		WHERE category_id = $1
	`

	var category model.Category
	err := tx.QueryRow(ctx, SQL, id).Scan(
		&category.CategoryID,
		&category.CategoryName,
		&category.CategoryDescription,
	)

	if err != nil {
		return nil, err
	}

	return &category, nil
}

// Save implements [CategoryRepository].
func (c *CategoryRepositoryImpl) Save(ctx context.Context, tx pgx.Tx, category *model.Category) error {
	const SQL = `
		INSERT INTO category (category_name, category_desc)
		VALUES ($1, $2)
		RETURNING category_id
	`

	err := tx.QueryRow(ctx, SQL, category.CategoryName, category.CategoryDescription).Scan(&category.CategoryID)
	if err != nil {
		return err
	}
	return nil
}

// Update implements [CategoryRepository].
func (c *CategoryRepositoryImpl) Update(ctx context.Context, tx pgx.Tx, category *model.Category) error {
	const SQL = `
		UPDATE category
		SET category_name = $1, category_desc = $2
		WHERE category_id = $3
	`
	_, err := tx.Exec(ctx, SQL, category.CategoryName, category.CategoryDescription, category.CategoryID)
	if err != nil {
		return err
	}
	return nil
}
