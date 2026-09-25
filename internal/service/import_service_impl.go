package service

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"stockopname-rita-backend/internal/dto"
	"stockopname-rita-backend/internal/model"
	"stockopname-rita-backend/internal/repository"

	"github.com/jackc/pgx/v5"
	"github.com/valentin-kaiser/go-dbase/dbase"
	"golang.org/x/text/encoding/charmap"
)

type ImportServiceImpl struct {
	ProductRepository repository.ProductRepository
	BarcodeRepository repository.BarcodeRepository
	Pool              repository.DBPool
}

type importProduct struct {
	line  int
	plu   string
	name  string
	dept  string
	buy   float64
	sell  float64
	codes []string
}

func openDBF(data []byte) (*dbase.File, error) {
	// Untested: true is required — the legacy files are dBase III (0x03),
	// which go-dbase refuses to open by default.
	table, err := dbase.OpenTable(&dbase.Config{
		Data:       data,
		TrimSpaces: true,
		Untested:   true,
		ReadOnly:   true,
		Converter:  dbase.NewDefaultConverter(charmap.Windows1252),
	})
	if err != nil {
		return nil, err
	}
	return table, nil
}

func requireColumns(table *dbase.File, names ...string) error {
	for _, name := range names {
		if table.ColumnPosByName(name) < 0 {
			return fmt.Errorf("column %s not found (got: %s)", name, strings.Join(table.ColumnNames(), ", "))
		}
	}
	return nil
}

// Import implements [ImportService].
func (s *ImportServiceImpl) Import(ctx context.Context, produkData []byte, barcodeData []byte, dryRun bool) (dto.ImportResultResponse, error) {
	result := dto.ImportResultResponse{
		InvalidRows:            []dto.ImportInvalidRow{},
		OrphanBarcodes:         []string{},
		ProductsWithoutBarcode: []string{},
		DryRun:                 dryRun,
	}
	defer func() {
		sort.Strings(result.OrphanBarcodes)
		sort.Strings(result.ProductsWithoutBarcode)
	}()

	products, invalid := parseProdukFile(produkData)
	result.InvalidRows = append(result.InvalidRows, invalid...)
	if len(products) == 0 {
		return result, &model.ValidationError{Detail: "PRODUK.DBF contains no valid rows"}
	}

	barcodeLines, invalid := parseBarcodeFile(barcodeData)
	result.InvalidRows = append(result.InvalidRows, invalid...)

	// Attach barcodes to products; quarantine orphan barcodes.
	byPLU := make(map[string]*importProduct, len(products))
	for _, p := range products {
		byPLU[p.plu] = p
	}
	seenBarcode := make(map[string]bool)
	for _, b := range barcodeLines {
		p, ok := byPLU[b.plu]
		if !ok {
			result.OrphanBarcodes = append(result.OrphanBarcodes, b.plu+"|"+b.code)
			continue
		}
		key := b.plu + "|" + b.code
		if seenBarcode[key] {
			result.InvalidRows = append(result.InvalidRows, dto.ImportInvalidRow{File: "barcode", Line: b.line, PLU: b.plu, Reason: "duplicate barcode row"})
			continue
		}
		seenBarcode[key] = true
		p.codes = append(p.codes, b.code)
	}

	for _, p := range products {
		if len(p.codes) == 0 {
			result.ProductsWithoutBarcode = append(result.ProductsWithoutBarcode, p.plu)
		}
	}

	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return result, err
	}
	defer tx.Rollback(ctx)

	existing, err := s.ProductRepository.FindAllForImport(ctx, tx)
	if err != nil {
		return result, fmt.Errorf("load existing products: %w", err)
	}
	byExistingPLU := make(map[string]*model.Product, len(existing))
	for _, e := range existing {
		byExistingPLU[e.ProductPLU] = e
	}

	for _, p := range products {
		inserted, updated, added, removed, err := s.upsertProduct(ctx, tx, byExistingPLU[p.plu], p)
		if err != nil {
			return result, err
		}
		switch {
		case inserted:
			result.ProductsInserted++
		case updated:
			result.ProductsUpdated++
		default:
			result.ProductsUnchanged++
		}
		result.BarcodesAdded += added
		result.BarcodesRemoved += removed
	}

	if dryRun {
		return result, nil
	}

	if err := tx.Commit(ctx); err != nil {
		return result, err
	}

	return result, nil
}

// upsertProduct inserts or updates one product + its barcode set.
// updatedat is only touched when product fields actually changed
// (keeps the import idempotent for mobile sync).
func (s *ImportServiceImpl) upsertProduct(ctx context.Context, tx pgx.Tx, old *model.Product, p *importProduct) (inserted bool, updated bool, added int, removed int, err error) {
	if old == nil {
		product := &model.Product{
			ProductPLU:            p.plu,
			ProductName:           p.name,
			ProductDepartmentCode: p.dept,
			ProductBuyPrice:       p.buy,
			ProductSellPrice:      p.sell,
		}
		if err := s.ProductRepository.Save(ctx, tx, product); err != nil {
			return false, false, 0, 0, fmt.Errorf("save product %s: %w", p.plu, err)
		}
		if err := s.BarcodeRepository.SaveBatch(ctx, tx, product.ProductID, p.codes); err != nil {
			return false, false, 0, 0, fmt.Errorf("save barcodes %s: %w", p.plu, err)
		}
		return true, false, len(p.codes), 0, nil
	}

	oldCodes := make(map[string]bool, len(old.ProductBarcodes))
	for _, c := range old.ProductBarcodes {
		oldCodes[c] = true
	}
	newCodes := make(map[string]bool, len(p.codes))
	for _, c := range p.codes {
		newCodes[c] = true
	}
	for c := range newCodes {
		if !oldCodes[c] {
			added++
		}
	}
	for c := range oldCodes {
		if !newCodes[c] {
			removed++
		}
	}
	barcodesChanged := added > 0 || removed > 0

	fieldsChanged := old.ProductName != p.name ||
		old.ProductDepartmentCode != p.dept ||
		old.ProductBuyPrice != p.buy ||
		old.ProductSellPrice != p.sell

	// Any master change (including the barcode set, which is part of
	// /products/sync responses) must bump updatedat so mobile
	// incremental sync picks it up.
	if fieldsChanged || barcodesChanged {
		old.ProductName = p.name
		old.ProductDepartmentCode = p.dept
		old.ProductBuyPrice = p.buy
		old.ProductSellPrice = p.sell
		if err := s.ProductRepository.Update(ctx, tx, old); err != nil {
			return false, false, 0, 0, fmt.Errorf("update product %s: %w", p.plu, err)
		}
	}
	if barcodesChanged {
		if err := s.BarcodeRepository.Replace(ctx, tx, old.ProductID, p.codes); err != nil {
			return false, false, 0, 0, fmt.Errorf("replace barcodes %s: %w", p.plu, err)
		}
	}
	return false, fieldsChanged || barcodesChanged, added, removed, nil
}

type importBarcodeLine struct {
	line int
	plu  string
	code string
}

func parseProdukFile(data []byte) ([]*importProduct, []dto.ImportInvalidRow) {
	var products []*importProduct
	var invalid []dto.ImportInvalidRow

	table, err := openDBF(data)
	if err != nil {
		return nil, []dto.ImportInvalidRow{{File: "produk", Reason: "cannot parse DBF: " + err.Error()}}
	}
	defer table.Close()

	if err := requireColumns(table, "PLU", "NAMABRG", "KODE_DPT", "HRG_JUAL", "RP_BELI"); err != nil {
		return nil, []dto.ImportInvalidRow{{File: "produk", Reason: err.Error()}}
	}

	rows, err := table.Rows(false, true)
	if err != nil {
		return nil, []dto.ImportInvalidRow{{File: "produk", Reason: "cannot read rows: " + err.Error()}}
	}

	seen := make(map[string]int)
	for i, row := range rows {
		line := i + 1
		plu, _ := row.StringValueByName("PLU")
		name, _ := row.StringValueByName("NAMABRG")
		dept, _ := row.StringValueByName("KODE_DPT")
		buy, buyErr := row.FloatValueByName("HRG_JUAL")
		sell, sellErr := row.FloatValueByName("RP_BELI")

		reject := func(reason string) {
			invalid = append(invalid, dto.ImportInvalidRow{File: "produk", Line: line, PLU: plu, Reason: reason})
		}

		switch {
		case strings.TrimSpace(plu) == "":
			reject("empty PLU")
			continue
		case len(plu) > 10:
			reject("PLU longer than 10 characters")
			continue
		case strings.TrimSpace(name) == "":
			reject("empty product name")
			continue
		case len(name) > 150:
			reject("product name longer than 150 characters")
			continue
		case strings.TrimSpace(dept) == "":
			reject("empty department code")
			continue
		case len(dept) > 10:
			reject("department code longer than 10 characters")
			continue
		case buyErr != nil:
			reject("invalid HRG_JUAL")
			continue
		case sellErr != nil:
			reject("invalid RP_BELI")
			continue
		case buy < 0 || sell < 0:
			reject("negative price")
			continue
		}

		if prev, dup := seen[plu]; dup {
			invalid = append(invalid, dto.ImportInvalidRow{File: "produk", Line: prev, PLU: plu, Reason: "duplicate PLU, kept last occurrence"})
			for _, p := range products {
				if p.plu == plu {
					p.line = line
					p.name = name
					p.dept = dept
					p.buy = buy
					p.sell = sell
					break
				}
			}
			seen[plu] = line
			continue
		}
		seen[plu] = line
		products = append(products, &importProduct{line: line, plu: plu, name: name, dept: dept, buy: buy, sell: sell})
	}
	return products, invalid
}

func parseBarcodeFile(data []byte) ([]importBarcodeLine, []dto.ImportInvalidRow) {
	var lines []importBarcodeLine
	var invalid []dto.ImportInvalidRow

	table, err := openDBF(data)
	if err != nil {
		return nil, []dto.ImportInvalidRow{{File: "barcode", Reason: "cannot parse DBF: " + err.Error()}}
	}
	defer table.Close()

	if err := requireColumns(table, "PLU", "BARCODE"); err != nil {
		return nil, []dto.ImportInvalidRow{{File: "barcode", Reason: err.Error()}}
	}

	rows, err := table.Rows(false, true)
	if err != nil {
		return nil, []dto.ImportInvalidRow{{File: "barcode", Reason: "cannot read rows: " + err.Error()}}
	}

	for i, row := range rows {
		line := i + 1
		plu, _ := row.StringValueByName("PLU")
		code, _ := row.StringValueByName("BARCODE")

		switch {
		case strings.TrimSpace(plu) == "":
			invalid = append(invalid, dto.ImportInvalidRow{File: "barcode", Line: line, Reason: "empty PLU"})
			continue
		case strings.TrimSpace(code) == "":
			invalid = append(invalid, dto.ImportInvalidRow{File: "barcode", Line: line, PLU: plu, Reason: "empty barcode"})
			continue
		case len(code) > 15:
			invalid = append(invalid, dto.ImportInvalidRow{File: "barcode", Line: line, PLU: plu, Reason: "barcode longer than 15 characters"})
			continue
		}
		lines = append(lines, importBarcodeLine{line: line, plu: plu, code: code})
	}
	return lines, invalid
}
