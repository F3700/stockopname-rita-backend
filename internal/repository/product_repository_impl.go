package repository

import (
	"context"
	"stockopname-rita-backend/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProductRepositoryImpl struct {
	pool *pgxpool.Pool
}

func NewProductRepository(pool *pgxpool.Pool) ProductRepository {
	return &ProductRepositoryImpl{
		pool: pool,
	}
}

// Delete implements [ProductRepository].
func (p *ProductRepositoryImpl) Delete(ctx context.Context, tx pgx.Tx, id int) error {
	panic("unimplemented")
}

// FindAll implements [ProductRepository].
func (p *ProductRepositoryImpl) FindAll(ctx context.Context) ([]*model.Product, error) {
	panic("unimplemented")
}

// Save implements [ProductRepository].
func (p *ProductRepositoryImpl) Save(ctx context.Context, tx pgx.Tx, product *model.Product) error {
	const SQL = "INSERT INTO product (product_barcode, product_name, product_buyprice, product_sellprice, product_category_id, product_department_id) VALUES ($1, $2, $3, $4, $5, $6) RETURNING product_id"

	err := tx.QueryRow(ctx, SQL, product.ProductBarcode, product.ProductName, product.ProductBuyPrice, product.ProductSellPrice, product.ProductCategoryID, product.ProductDepartmentID).Scan(&product.ProductID)
	if err != nil {
		return err
	}
	return nil
}

// Update implements [ProductRepository].
func (p *ProductRepositoryImpl) Update(ctx context.Context, tx pgx.Tx, product *model.Product) error {
	panic("unimplemented")
}

func (p *ProductRepositoryImpl) FindById(ctx context.Context, tx pgx.Tx, id int) (*model.Product, error) {
	const SQL = `
        SELECT
            p.product_id,
            p.product_barcode,
            p.product_name,
            p.product_buyprice,
            p.product_sellprice,
            p.product_createdat,
            p.product_updatedat,
            c.category_name,
            d.department_code
        FROM product p
        JOIN category c
            ON c.category_id = p.product_category_id
        JOIN department d
            ON d.department_id = p.product_department_id
        WHERE p.product_id = $1
    `

	var product model.Product

	err := tx.QueryRow(ctx, SQL, id).Scan(
		&product.ProductID,
		&product.ProductBarcode,
		&product.ProductName,
		&product.ProductBuyPrice,
		&product.ProductSellPrice,
		&product.ProductCreatedat,
		&product.ProductUpdatedat,
		&product.ProductCategoryName,
		&product.ProductDepartmentCode,
	)

	if err != nil {
		return nil, err
	}

	return &product, nil
}
