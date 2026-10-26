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
func (b *BarcodeRepositoryImpl) SaveBatch(ctx context.Context, tx pgx.Tx, sesiId int, productID int, codes []string) error {
	const SQL = `INSERT INTO barcode (barcode_sesi_id, barcode_product_id, barcode_code) VALUES ($1, $2, $3)`
	for _, code := range codes {
		if _, err := tx.Exec(ctx, SQL, sesiId, productID, code); err != nil {
			return model.MapPgError("Barcode", err)
		}
	}
	return nil
}

// Replace implements [BarcodeRepository].
func (b *BarcodeRepositoryImpl) Replace(ctx context.Context, tx pgx.Tx, sesiId int, productID int, codes []string) error {
	const SQL = `DELETE FROM barcode WHERE barcode_product_id = $1 AND barcode_sesi_id = $2`
	if _, err := tx.Exec(ctx, SQL, productID, sesiId); err != nil {
		return model.MapPgError("Barcode", err)
	}
	return b.SaveBatch(ctx, tx, sesiId, productID, codes)
}
