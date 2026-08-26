package repository

import (
	"context"
	"fmt"
	"stockopname-rita-backend/internal/model"
	"time"

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
	const SQL = "DELETE FROM product WHERE product_id = $1"

	res, err := tx.Exec(ctx, SQL, id)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("product with ID %d not found", id)
	}
	return nil
}

// FindAll implements [ProductRepository].
func (p *ProductRepositoryImpl) FindAll(ctx context.Context) ([]*model.Product, error) {
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
	`

	var products []*model.Product

	rows, err := p.pool.Query(ctx, SQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var product model.Product
		err := rows.Scan(
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
		products = append(products, &product)
	}

	return products, nil
}

func (p *ProductRepositoryImpl) FindAllUpdatedAfter(ctx context.Context, date time.Time) ([]*model.Product, error) {
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
		WHERE p.product_updatedat > $1
	`

	var products []*model.Product

	rows, err := p.pool.Query(ctx, SQL, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var product model.Product
		err := rows.Scan(
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
		products = append(products, &product)
	}

	return products, nil
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
	const SQL = "UPDATE product SET product_barcode = $1, product_name = $2, product_buyprice = $3, product_sellprice = $4, product_updatedat = NOW(), product_category_id = $5, product_department_id = $6 WHERE product_id = $7"

	_, err := tx.Exec(ctx, SQL, product.ProductBarcode, product.ProductName, product.ProductBuyPrice, product.ProductSellPrice, product.ProductCategoryID, product.ProductDepartmentID, product.ProductID)
	if err != nil {
		return err
	}
	return nil
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
			p.product_category_id,
			p.product_department_id,
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
		&product.ProductCategoryID,
		&product.ProductDepartmentID,
		&product.ProductCategoryName,
		&product.ProductDepartmentCode,
	)

	if err != nil {
		return nil, err
	}

	return &product, nil
}
