package repository

import (
	"context"
	"errors"
	"stockopname-rita-backend/internal/model"

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
	p.product_sesi_id,
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
		&product.ProductSesiID,
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
func (p *ProductRepositoryImpl) Clear(ctx context.Context, tx pgx.Tx, sesiId int) error {
	const SQL = `DELETE FROM product WHERE product_sesi_id = $1`
	_, err := tx.Exec(ctx, SQL, sesiId)
	return err
}

// CountBySesi implements [ProductRepository].
func (p *ProductRepositoryImpl) CountBySesi(ctx context.Context, tx pgx.Tx, sesiId int) (int, error) {
	const SQL = `SELECT COUNT(*) FROM product WHERE product_sesi_id = $1`
	var total int
	if err := tx.QueryRow(ctx, SQL, sesiId).Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

// FindById implements [ProductRepository].
// product_id is globally unique; callers that need session isolation must
// check ProductSesiID after loading.
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
func (p *ProductRepositoryImpl) FindByPLU(ctx context.Context, tx pgx.Tx, sesiId int, plu string) (*model.Product, error) {
	const SQL = `
		SELECT ` + productColumns + `
		FROM product p
		WHERE p.product_sesi_id = $1
		  AND p.product_plu = $2
	`

	product, err := scanProduct(tx.QueryRow(ctx, SQL, sesiId, plu))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &model.NotFoundError{Resource: "Product", Detail: "Product with PLU " + plu + " not found in this session"}
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
func (p *ProductRepositoryImpl) FindByBarcode(ctx context.Context, tx pgx.Tx, sesiId int, code string) (*model.Product, error) {
	const SQL = `
		SELECT ` + productColumns + `
		FROM product p
		JOIN barcode b
			ON b.barcode_product_id = p.product_id
		   AND b.barcode_sesi_id = p.product_sesi_id
		WHERE b.barcode_sesi_id = $1
		  AND b.barcode_code = $2
	`

	product, err := scanProduct(tx.QueryRow(ctx, SQL, sesiId, code))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &model.NotFoundError{Resource: "Product", Detail: "Product with barcode " + code + " not found in this session"}
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
func (p *ProductRepositoryImpl) FindAllForImport(ctx context.Context, tx pgx.Tx, sesiId int) ([]*model.Product, error) {
	const SQL = `
		SELECT ` + productColumns + `
		FROM product p
		WHERE p.product_sesi_id = $1
		ORDER BY p.product_id
	`

	rows, err := tx.Query(ctx, SQL, sesiId)
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

func (p *ProductRepositoryImpl) FindAllInPageSearch(ctx context.Context, sesiId int, limit int, offset int, search string) ([]*model.Product, int, error) {
	const SQL = `
		SELECT ` + productColumns + `
		FROM product p
		WHERE p.product_sesi_id = $1
		AND (
			p.product_plu ILIKE $2
			OR p.product_name ILIKE $2
			OR p.product_department_code ILIKE $2
			OR EXISTS (
				SELECT 1 FROM barcode b
				WHERE b.barcode_product_id = p.product_id
				  AND b.barcode_code ILIKE $2
			)
		)
		ORDER BY p.product_id DESC
		LIMIT $3 OFFSET $4
	`

	rows, err := p.pool.Query(ctx, SQL, sesiId, "%"+search+"%", limit, offset)
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
		WHERE p.product_sesi_id = $1
		AND (
			p.product_plu ILIKE $2
			OR p.product_name ILIKE $2
			OR p.product_department_code ILIKE $2
			OR EXISTS (
				SELECT 1 FROM barcode b
				WHERE b.barcode_product_id = p.product_id
				  AND b.barcode_code ILIKE $2
			)
		)
	`
	if err := p.pool.QueryRow(ctx, countSQL, sesiId, "%"+search+"%").Scan(&total); err != nil {
		return nil, 0, err
	}

	if err := attachBarcodes(ctx, p.pool, products); err != nil {
		return nil, 0, err
	}
	return products, total, nil
}

// Save implements [ProductRepository].
func (p *ProductRepositoryImpl) Save(ctx context.Context, tx pgx.Tx, product *model.Product) error {
	const SQL = "INSERT INTO product (product_sesi_id, product_plu, product_name, product_department_code, product_buyprice, product_sellprice) VALUES ($1, $2, $3, $4, $5, $6) RETURNING product_id"

	err := tx.QueryRow(ctx, SQL, product.ProductSesiID, product.ProductPLU, product.ProductName, product.ProductDepartmentCode, product.ProductBuyPrice, product.ProductSellPrice).Scan(&product.ProductID)
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
