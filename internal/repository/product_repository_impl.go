package repository

import (
	"context"
	"errors"
	"stockopname-rita-backend/internal/model"
	"time"

	"github.com/jackc/pgx/v5"
)

type ProductRepositoryImpl struct {
	pool DBPool
}

func NewProductRepository(pool DBPool) ProductRepository {
	return &ProductRepositoryImpl{
		pool: pool,
	}
}

// querier is satisfied by both pgx.Tx and DBPool.
type querier interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

const productColumns = `
	p.product_id,
	p.product_plu,
	p.product_name,
	p.product_department_code,
	p.product_buyprice,
	p.product_sellprice,
	p.product_createdat,
	p.product_updatedat
`

func scanProduct(row pgx.Row) (*model.Product, error) {
	var product model.Product
	err := row.Scan(
		&product.ProductID,
		&product.ProductPLU,
		&product.ProductName,
		&product.ProductDepartmentCode,
		&product.ProductBuyPrice,
		&product.ProductSellPrice,
		&product.ProductCreatedat,
		&product.ProductUpdatedat,
	)
	if err != nil {
		return nil, err
	}
	return &product, nil
}

// attachBarcodes loads barcodes (insertion order) for the given products.
func attachBarcodes(ctx context.Context, q querier, products []*model.Product) error {
	if len(products) == 0 {
		return nil
	}
	ids := make([]int, 0, len(products))
	byID := make(map[int]*model.Product, len(products))
	for _, p := range products {
		ids = append(ids, p.ProductID)
		byID[p.ProductID] = p
	}

	const SQL = `
		SELECT barcode_product_id, barcode_code
		FROM barcode
		WHERE barcode_product_id = ANY($1)
		ORDER BY barcode_id
	`
	rows, err := q.Query(ctx, SQL, ids)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var productID int
		var code string
		if err := rows.Scan(&productID, &code); err != nil {
			return err
		}
		if p, ok := byID[productID]; ok {
			p.ProductBarcodes = append(p.ProductBarcodes, code)
		}
	}
	return rows.Err()
}

// Delete implements [ProductRepository].
// Barcodes are removed by the ON DELETE CASCADE foreign key.
func (p *ProductRepositoryImpl) Delete(ctx context.Context, tx pgx.Tx, id int) error {
	const SQL = "DELETE FROM product WHERE product_id = $1"

	res, err := tx.Exec(ctx, SQL, id)
	if err != nil {
		return model.MapPgError("Product", err)
	}
	if res.RowsAffected() == 0 {
		return &model.NotFoundError{Resource: "Product", ID: id}
	}
	return nil
}

// Clear implements [ProductRepository].
func (p *ProductRepositoryImpl) Clear(ctx context.Context, tx pgx.Tx) error {
	const SQL = `TRUNCATE product, barcode, deleted_product RESTART IDENTITY`
	_, err := tx.Exec(ctx, SQL)
	return err
}

// FindById implements [ProductRepository].
func (p *ProductRepositoryImpl) FindById(ctx context.Context, tx pgx.Tx, id int) (*model.Product, error) {
	const SQL = `
		SELECT ` + productColumns + `
		FROM product p
		WHERE p.product_id = $1
	`

	product, err := scanProduct(tx.QueryRow(ctx, SQL, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &model.NotFoundError{Resource: "Product", ID: id}
		}
		return nil, err
	}

	products := []*model.Product{product}
	if err := attachBarcodes(ctx, tx, products); err != nil {
		return nil, err
	}
	return product, nil
}

// FindByPLU implements [ProductRepository].
func (p *ProductRepositoryImpl) FindByPLU(ctx context.Context, tx pgx.Tx, plu string) (*model.Product, error) {
	const SQL = `
		SELECT ` + productColumns + `
		FROM product p
		WHERE p.product_plu = $1
	`

	product, err := scanProduct(tx.QueryRow(ctx, SQL, plu))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &model.NotFoundError{Resource: "Product", Detail: "Product with PLU " + plu + " not found"}
		}
		return nil, err
	}

	products := []*model.Product{product}
	if err := attachBarcodes(ctx, tx, products); err != nil {
		return nil, err
	}
	return product, nil
}

// FindByBarcode implements [ProductRepository].
func (p *ProductRepositoryImpl) FindByBarcode(ctx context.Context, tx pgx.Tx, code string) (*model.Product, error) {
	const SQL = `
		SELECT ` + productColumns + `
		FROM product p
		JOIN barcode b
			ON b.barcode_product_id = p.product_id
		WHERE b.barcode_code = $1
	`

	product, err := scanProduct(tx.QueryRow(ctx, SQL, code))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &model.NotFoundError{Resource: "Product", Detail: "Product with barcode " + code + " not found"}
		}
		return nil, err
	}

	products := []*model.Product{product}
	if err := attachBarcodes(ctx, tx, products); err != nil {
		return nil, err
	}
	return product, nil
}

// FindAllForImport implements [ProductRepository].
func (p *ProductRepositoryImpl) FindAllForImport(ctx context.Context, tx pgx.Tx) ([]*model.Product, error) {
	const SQL = `
		SELECT ` + productColumns + `
		FROM product p
		ORDER BY p.product_id
	`

	rows, err := tx.Query(ctx, SQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []*model.Product
	for rows.Next() {
		product, err := scanProduct(rows)
		if err != nil {
			return nil, err
		}
		products = append(products, product)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if err := attachBarcodes(ctx, tx, products); err != nil {
		return nil, err
	}
	return products, nil
}

func (p *ProductRepositoryImpl) FindAllInPageSearch(ctx context.Context, limit int, offset int, search string) ([]*model.Product, int, error) {
	const SQL = `
		SELECT ` + productColumns + `
		FROM product p
		WHERE p.product_plu ILIKE $1
			OR p.product_name ILIKE $1
			OR p.product_department_code ILIKE $1
			OR EXISTS (
				SELECT 1 FROM barcode b
				WHERE b.barcode_product_id = p.product_id
				  AND b.barcode_code ILIKE $1
			)
		ORDER BY p.product_id DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := p.pool.Query(ctx, SQL, "%"+search+"%", limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var products []*model.Product
	for rows.Next() {
		product, err := scanProduct(rows)
		if err != nil {
			return nil, 0, err
		}
		products = append(products, product)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	var total int
	countSQL := `
		SELECT COUNT(*)
		FROM product p
		WHERE p.product_plu ILIKE $1
			OR p.product_name ILIKE $1
			OR p.product_department_code ILIKE $1
			OR EXISTS (
				SELECT 1 FROM barcode b
				WHERE b.barcode_product_id = p.product_id
				  AND b.barcode_code ILIKE $1
			)
	`
	if err := p.pool.QueryRow(ctx, countSQL, "%"+search+"%").Scan(&total); err != nil {
		return nil, 0, err
	}

	if err := attachBarcodes(ctx, p.pool, products); err != nil {
		return nil, 0, err
	}
	return products, total, nil
}

func (p *ProductRepositoryImpl) FindAllUpdatedAfter(ctx context.Context, date time.Time, limit int, offset int) ([]*model.Product, int, error) {
	const SQL = `
		SELECT ` + productColumns + `
		FROM product p
		WHERE p.product_updatedat > $1
		ORDER BY p.product_id DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := p.pool.Query(ctx, SQL, date, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var products []*model.Product
	for rows.Next() {
		product, err := scanProduct(rows)
		if err != nil {
			return nil, 0, err
		}
		products = append(products, product)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	var total int
	countSQL := `SELECT COUNT(*) FROM product p WHERE p.product_updatedat > $1`
	if err := p.pool.QueryRow(ctx, countSQL, date).Scan(&total); err != nil {
		return nil, 0, err
	}

	if err := attachBarcodes(ctx, p.pool, products); err != nil {
		return nil, 0, err
	}
	return products, total, nil
}

// Save implements [ProductRepository].
func (p *ProductRepositoryImpl) Save(ctx context.Context, tx pgx.Tx, product *model.Product) error {
	const SQL = "INSERT INTO product (product_plu, product_name, product_department_code, product_buyprice, product_sellprice) VALUES ($1, $2, $3, $4, $5) RETURNING product_id"

	err := tx.QueryRow(ctx, SQL, product.ProductPLU, product.ProductName, product.ProductDepartmentCode, product.ProductBuyPrice, product.ProductSellPrice).Scan(&product.ProductID)
	if err != nil {
		return model.MapPgError("Product", err)
	}
	return nil
}

// Update implements [ProductRepository].
func (p *ProductRepositoryImpl) Update(ctx context.Context, tx pgx.Tx, product *model.Product) error {
	const SQL = "UPDATE product SET product_plu = $1, product_name = $2, product_department_code = $3, product_buyprice = $4, product_sellprice = $5, product_updatedat = NOW() WHERE product_id = $6"

	res, err := tx.Exec(ctx, SQL, product.ProductPLU, product.ProductName, product.ProductDepartmentCode, product.ProductBuyPrice, product.ProductSellPrice, product.ProductID)
	if err != nil {
		return model.MapPgError("Product", err)
	}
	if res.RowsAffected() == 0 {
		return &model.NotFoundError{Resource: "Product", ID: product.ProductID}
	}
	return nil
}
