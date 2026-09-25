package service

import (
	"context"
	"stockopname-rita-backend/internal/dto"
	"stockopname-rita-backend/internal/repository"
)

type ImportService interface {
	// Import upserts master data from a PRODUK.DBF + BARCODE.DBF pair.
	// It is idempotent: re-importing the same files changes nothing.
	// When dryRun is true nothing is written.
	Import(ctx context.Context, produkData []byte, barcodeData []byte, dryRun bool) (dto.ImportResultResponse, error)
}

func NewImportService(productRepository repository.ProductRepository, barcodeRepository repository.BarcodeRepository, pool repository.DBPool) ImportService {
	return &ImportServiceImpl{
		ProductRepository: productRepository,
		BarcodeRepository: barcodeRepository,
		Pool:              pool,
	}
}
