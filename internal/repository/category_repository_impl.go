package repository

import (
	"context"
	"errors"
	"stockopname-rita-backend/internal/model"

	"github.com/jackc/pgx/v5"
)

type CategoryRepositoryImpl struct {
	pool DBPool
}

func NewCategoryRepositoryImpl(pool DBPool) CategoryRepository {
	return &CategoryRepositoryImpl{
		pool: pool,
	}
}

// Delete implements [CategoryRepository].
func (c *CategoryRepositoryImpl) Delete(ctx context.Context, tx pgx.Tx, id int) error {
	const SQL = `DELETE FROM category WHERE category_id = $1`

	res, err := tx.Exec(ctx, SQL, id)
	if err != nil {
		return model.MapPgError("Category", err)
	}
	if res.RowsAffected() == 0 {
		return &model.NotFoundError{Resource: "Category", ID: id}
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

// FindAllInPageSearch implements [CategoryRepository].
func (c *CategoryRepositoryImpl) FindAllInPageSearch(ctx context.Context, limit int, offset int, search string) ([]*model.Category, int, error) {
	const SQL = `
		SELECT
			category_id,
			category_name,
			category_desc
		FROM category
		WHERE category_name ILIKE $1
		ORDER BY category_id DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := c.pool.Query(ctx, SQL, "%"+search+"%", limit, offset)
	if err != nil {
		return nil, 0, err
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
			return nil, 0, err
		}
		categories = append(categories, &category)
	}

	var total int
	countSQL := `SELECT COUNT(*) FROM category WHERE category_name ILIKE $1`
	if err := c.pool.QueryRow(ctx, countSQL, "%"+search+"%").Scan(&total); err != nil {
		return nil, 0, err
	}

	return categories, total, nil
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
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &model.NotFoundError{Resource: "Category", ID: id}
		}
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
		return model.MapPgError("Category", err)
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
