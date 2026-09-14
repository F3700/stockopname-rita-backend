package repository

import (
	"context"
	"stockopname-rita-backend/internal/model"
	"time"

	"github.com/jackc/pgx/v5"
)

type DeletedProductRepositoryImpl struct {
	pool DBPool
}

func NewDeletedProductRepository(pool DBPool) DeletedProductRepository {
	return &DeletedProductRepositoryImpl{
		pool: pool,
	}
}

// Delete implements [DeletedProductRepository].
func (d *DeletedProductRepositoryImpl) Delete(ctx context.Context) error {
	const SQL = `
		DELETE FROM deleted_product
	`
	_, err := d.pool.Exec(ctx, SQL)
	return err
}

// FindAll implements [DeletedProductRepository].
func (d *DeletedProductRepositoryImpl) FindAll(ctx context.Context) ([]*model.DeletedProduct, error) {
	const SQL = `
		SELECT product_id, deleted_at
		FROM deleted_product
	`
	rows, err := d.pool.Query(ctx, SQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deletedProducts []*model.DeletedProduct
	for rows.Next() {
		var deletedProduct model.DeletedProduct
		if err := rows.Scan(&deletedProduct.ProductID, &deletedProduct.DeletedAt); err != nil {
			return nil, err
		}
		deletedProducts = append(deletedProducts, &deletedProduct)
	}
	return deletedProducts, nil
}

// FindAllUpdatedAfter implements [DeletedProductRepository].
func (d *DeletedProductRepositoryImpl) FindAllUpdatedAfter(ctx context.Context, date time.Time) ([]*model.DeletedProduct, error) {
	const SQL = `
		SELECT product_id, deleted_at
		FROM deleted_product
		WHERE deleted_at > $1
	`
	rows, err := d.pool.Query(ctx, SQL, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deletedProducts []*model.DeletedProduct
	for rows.Next() {
		var deletedProduct model.DeletedProduct
		if err := rows.Scan(&deletedProduct.ProductID, &deletedProduct.DeletedAt); err != nil {
			return nil, err
		}
		deletedProducts = append(deletedProducts, &deletedProduct)
	}
	return deletedProducts, nil
}

// Save implements [DeletedProductRepository].
func (d *DeletedProductRepositoryImpl) Save(ctx context.Context, tx pgx.Tx, id int) error {
	const SQL = `
		INSERT INTO deleted_product (product_id)
		VALUES ($1)
	`
	_, err := tx.Exec(ctx, SQL, id)
	if err != nil {
		return model.MapPgError("Deleted Product", err)
	}
	return nil
}
