package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type BarcodeRepository interface {
	SaveBatch(ctx context.Context, tx pgx.Tx, sesiId int, productID int, codes []string) error
	// Replace deletes all barcodes of a product in one session and inserts
	// the given ones.
	Replace(ctx context.Context, tx pgx.Tx, sesiId int, productID int, codes []string) error
}
