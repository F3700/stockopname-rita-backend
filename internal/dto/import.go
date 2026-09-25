package dto

// ImportInvalidRow describes one skipped input row.
type ImportInvalidRow struct {
	File   string `json:"file"`
	Line   int    `json:"line"`
	PLU    string `json:"plu"`
	Reason string `json:"reason"`
}

type ImportResultResponse struct {
	ProductsInserted       int                `json:"products_inserted"`
	ProductsUpdated        int                `json:"products_updated"`
	ProductsUnchanged      int                `json:"products_unchanged"`
	BarcodesAdded          int                `json:"barcodes_added"`
	BarcodesRemoved        int                `json:"barcodes_removed"`
	InvalidRows            []ImportInvalidRow `json:"invalid_rows"`
	OrphanBarcodes         []string           `json:"orphan_barcodes"`
	ProductsWithoutBarcode []string           `json:"products_without_barcode"`
	DryRun                 bool               `json:"dry_run"`
}
