package service

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"stockopname-rita-backend/internal/dto"
	"stockopname-rita-backend/internal/model"
	"stockopname-rita-backend/internal/repository"

	"github.com/valentin-kaiser/go-dbase/dbase"
	"golang.org/x/text/encoding/charmap"
)

type ImportServiceImpl struct {
	ProductRepository repository.ProductRepository
	BarcodeRepository repository.BarcodeRepository
	SesiRepository    repository.SesiRepository
	Pool              repository.DBPool
}

type importProduct struct {
	line  int
	plu   string
	name  string
	dept  string
	buy   float64
	sell  float64
	codes []importCode
}

type importCode struct {
	code string
	line int
}

func (p *importProduct) codeStrings() []string {
	codes := make([]string, 0, len(p.codes))
	for _, c := range p.codes {
		codes = append(codes, c.code)
	}
	return codes
}

func (p *importProduct) hasCode(code string) bool {
	for _, c := range p.codes {
		if c.code == code {
			return true
		}
	}
	return false
}

func (p *importProduct) removeCode(code string) {
	filtered := p.codes[:0]
	for _, c := range p.codes {
		if c.code != code {
			filtered = append(filtered, c)
		}
	}
	p.codes = filtered
}

func (p *importProduct) codeLine(code string) int {
	for _, c := range p.codes {
		if c.code == code {
			return c.line
		}
	}
	return p.line
}

type barcodeWrite struct {
	productID int
	oldCodes  []string
	newCodes  []string
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
func (s *ImportServiceImpl) Import(ctx context.Context, sesiId int, produkData []byte, barcodeData []byte, dryRun bool) (dto.ImportResultResponse, error) {
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
		p.codes = append(p.codes, importCode{code: b.code, line: b.line})
	}

	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return result, err
	}
	defer tx.Rollback(ctx)

	sesi, err := s.SesiRepository.FindByIdForShare(ctx, tx, sesiId)
	if err != nil {
		return result, err
	}
	if err := ensureSessionWritable(sesi); err != nil {
		return result, err
	}

	existing, err := s.ProductRepository.FindAllForImport(ctx, tx, sesiId)
	if err != nil {
		return result, fmt.Errorf("load existing products: %w", err)
	}
	byExistingPLU := make(map[string]*model.Product, len(existing))
	ownerByCode := make(map[string]string)
	for _, e := range existing {
		byExistingPLU[e.ProductPLU] = e
		for _, code := range e.ProductBarcodes {
			ownerByCode[code] = e.ProductPLU
		}
	}

	s.resolveBarcodeClaims(products, byPLU, ownerByCode, &result)

	for _, p := range products {
		if len(p.codes) == 0 {
			result.ProductsWithoutBarcode = append(result.ProductsWithoutBarcode, p.plu)
		}
	}

	writes := make([]barcodeWrite, 0, len(products))
	for _, p := range products {
		old := byExistingPLU[p.plu]
		newCodes := p.codeStrings()

		if old == nil {
			product := &model.Product{
				ProductSesiID:         sesiId,
				ProductPLU:            p.plu,
				ProductName:           p.name,
				ProductDepartmentCode: p.dept,
				ProductBuyPrice:       p.buy,
				ProductSellPrice:      p.sell,
			}
			if err := s.ProductRepository.Save(ctx, tx, product); err != nil {
				return result, fmt.Errorf("save product %s: %w", p.plu, err)
			}
			result.ProductsInserted++
			result.BarcodesAdded += len(newCodes)
			writes = append(writes, barcodeWrite{productID: product.ProductID, oldCodes: nil, newCodes: newCodes})
			continue
		}

		oldSet := toSet(old.ProductBarcodes)
		newSet := toSet(newCodes)
		added, removed := diffSets(oldSet, newSet)

		fieldsChanged := old.ProductName != p.name ||
			old.ProductDepartmentCode != p.dept ||
			old.ProductBuyPrice != p.buy ||
			old.ProductSellPrice != p.sell
		barcodesChanged := added > 0 || removed > 0

		if fieldsChanged {
			old.ProductName = p.name
			old.ProductDepartmentCode = p.dept
			old.ProductBuyPrice = p.buy
			old.ProductSellPrice = p.sell
			if err := s.ProductRepository.Update(ctx, tx, old); err != nil {
				return result, fmt.Errorf("update product %s: %w", p.plu, err)
			}
		}

		if barcodesChanged {
			result.ProductsUpdated++
		} else if fieldsChanged {
			result.ProductsUpdated++
		} else {
			result.ProductsUnchanged++
		}
		result.BarcodesAdded += added
		result.BarcodesRemoved += removed
		if barcodesChanged {
			writes = append(writes, barcodeWrite{productID: old.ProductID, oldCodes: old.ProductBarcodes, newCodes: newCodes})
		}
	}

	if dryRun {
		return result, nil
	}

	// Phase 2: delete first for every changed product, then insert. This
	// lets a barcode move from one PLU to another without tripping the
	// immediate unique constraint.
	for _, w := range writes {
		if len(w.oldCodes) == 0 {
			continue
		}
		if err := s.BarcodeRepository.Replace(ctx, tx, sesiId, w.productID, nil); err != nil {
			return result, fmt.Errorf("clear barcodes product %d: %w", w.productID, err)
		}
	}
	for _, w := range writes {
		if len(w.newCodes) == 0 {
			continue
		}
		if err := s.BarcodeRepository.SaveBatch(ctx, tx, sesiId, w.productID, w.newCodes); err != nil {
			return result, fmt.Errorf("save barcodes product %d: %w", w.productID, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return result, err
	}

	return result, nil
}

// resolveBarcodeClaims implements D4-A: decide which barcode/PLU claims are
// valid without depending on file order. Invalid claims are removed from the
// product's code set and reported in invalid_rows.
func (s *ImportServiceImpl) resolveBarcodeClaims(products []*importProduct, byPLU map[string]*importProduct, ownerByCode map[string]string, result *dto.ImportResultResponse) {
	candidates := make(map[string][]string)
	for _, p := range products {
		for _, c := range p.codes {
			candidates[c.code] = append(candidates[c.code], p.plu)
		}
	}

	invalid := func(p *importProduct, code string, reason string) {
		result.InvalidRows = append(result.InvalidRows, dto.ImportInvalidRow{
			File:   "barcode",
			Line:   p.codeLine(code),
			PLU:    p.plu,
			Reason: reason,
		})
		p.removeCode(code)
	}

	for code, plus := range candidates {
		owner, hasOwner := ownerByCode[code]

		if len(plus) == 1 {
			plu := plus[0]
			p := byPLU[plu]
			if !hasOwner || plu == owner {
				continue
			}
			ownerProduct, ownerInFile := byPLU[owner]
			if ownerInFile && !ownerProduct.hasCode(code) {
				continue // valid transfer
			}
			invalid(p, code, "barcode already used by another product in this session")
			continue
		}

		ownerKeeps := false
		if hasOwner {
			if ownerProduct, ok := byPLU[owner]; ok && ownerProduct.hasCode(code) {
				ownerKeeps = true
			}
		}

		for _, plu := range plus {
			if ownerKeeps && plu == owner {
				continue
			}
			reason := "barcode claimed by multiple products"
			if ownerKeeps {
				reason = "barcode already used by another product in this session"
			}
			invalid(byPLU[plu], code, reason)
		}
	}
}

func toSet(codes []string) map[string]bool {
	set := make(map[string]bool, len(codes))
	for _, c := range codes {
		set[c] = true
	}
	return set
}

func diffSets(oldSet, newSet map[string]bool) (added int, removed int) {
	for c := range newSet {
		if !oldSet[c] {
			added++
		}
	}
	for c := range oldSet {
		if !newSet[c] {
			removed++
		}
	}
	return added, removed
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
		// Legacy names: HRG_JUAL is the selling price, RP_BELI the purchase price.
		sell, sellErr := row.FloatValueByName("HRG_JUAL")
		buy, buyErr := row.FloatValueByName("RP_BELI")

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
			reject("invalid RP_BELI")
			continue
		case sellErr != nil:
			reject("invalid HRG_JUAL")
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
