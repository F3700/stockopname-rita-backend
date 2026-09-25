package repository

import (
	"context"
	"stockopname-rita-backend/internal/model"

	"github.com/jackc/pgx/v5"
)

type BarcodeRepositoryImpl struct {
	pool DBPool
}

func NewBarcodeRepository(pool DBPool) BarcodeRepository {
	return &BarcodeRepositoryImpl{
		pool: pool,
	}
}

// SaveBatch implements [BarcodeRepository].
func (b *BarcodeRepositoryImpl) SaveBatch(ctx context.Context, tx pgx.Tx, productID int, codes []string) error {
	const SQL = `INSERT INTO barcode (barcode_product_id, barcode_code) VALUES ($1, $2)`
	for _, code := range codes {
		if _, err := tx.Exec(ctx, SQL, productID, code); err != nil {
			return model.MapPgError("Barcode", err)
		}
	}
	return nil
}

// Replace implements [BarcodeRepository].
func (b *BarcodeRepositoryImpl) Replace(ctx context.Context, tx pgx.Tx, productID int, codes []string) error {
	const SQL = `DELETE FROM barcode WHERE barcode_product_id = $1`
	if _, err := tx.Exec(ctx, SQL, productID); err != nil {
		return model.MapPgError("Barcode", err)
	}
	return b.SaveBatch(ctx, tx, productID, codes)
}
