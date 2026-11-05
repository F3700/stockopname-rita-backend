package service

import (
	"context"
	"encoding/json"
	"math"
	"stockopname-rita-backend/internal/dto"
	"stockopname-rita-backend/internal/helper"
	"stockopname-rita-backend/internal/model"
	"stockopname-rita-backend/internal/repository"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5"
)

type StockOpnameServiceImpl struct {
	Pool                  repository.DBPool
	StockOpnameRepository repository.StockOpnameRepository
	ProductRepository     repository.ProductRepository
	SesiRepository        repository.SesiRepository
	AuditRepository       repository.AuditRepository
	CoordinatorRepository repository.CoordinatorRepository
	Validator             *validator.Validate
}

func NewStockOpnameService(
	stockOpnameRepository repository.StockOpnameRepository,
	productRepository repository.ProductRepository,
	sesiRepository repository.SesiRepository,
	auditRepository repository.AuditRepository,
	coordinatorRepository repository.CoordinatorRepository,
	pool repository.DBPool,
	validate *validator.Validate,
) StockOpnameService {
	return &StockOpnameServiceImpl{
		Pool:                  pool,
		StockOpnameRepository: stockOpnameRepository,
		ProductRepository:     productRepository,
		SesiRepository:        sesiRepository,
		AuditRepository:       auditRepository,
		CoordinatorRepository: coordinatorRepository,
		Validator:             validate,
	}
}

// resolveProduct finds the master product by scanned barcode or PLU inside
// the given session.
func (s *StockOpnameServiceImpl) resolveProduct(ctx context.Context, tx pgx.Tx, sesiId int, barcode string, plu string) (*model.Product, error) {
	if barcode != "" {
		return s.ProductRepository.FindByBarcode(ctx, tx, sesiId, barcode)
	}
	if plu != "" {
		return s.ProductRepository.FindByPLU(ctx, tx, sesiId, plu)
	}
	return nil, &model.ValidationError{Detail: "barcode or plu is required"}
}

// checkSessionWritable locks the session and enforces D3.
func (s *StockOpnameServiceImpl) checkSessionWritable(ctx context.Context, tx pgx.Tx, sesiId int) error {
	sesi, err := s.SesiRepository.FindByIdForShare(ctx, tx, sesiId)
	if err != nil {
		return err
	}
	return ensureSessionWritable(sesi)
}

// checkSessionReady locks the session, enforces D3, and requires at least
// one product (D6).
func (s *StockOpnameServiceImpl) checkSessionReady(ctx context.Context, tx pgx.Tx, sesiId int) error {
	if err := s.checkSessionWritable(ctx, tx, sesiId); err != nil {
		return err
	}
	total, err := s.ProductRepository.CountBySesi(ctx, tx, sesiId)
	if err != nil {
		return err
	}
	if total == 0 {
		return &model.ConflictError{Resource: "Session", Detail: "Session has no products yet"}
	}
	return nil
}

func ensureCoordinatorWritable(status string) error {
	if status == "COMPLETED" || status == "CANCELLED" {
		return &model.ConflictError{Resource: "Coordinator", Detail: "Coordinator is final"}
	}
	return nil
}

func actorOrDefault(actor string) string {
	actor = strings.TrimSpace(actor)
	if actor == "" {
		return "admin-web"
	}
	runes := []rune(actor)
	if len(runes) > 100 {
		return string(runes[:100])
	}
	return actor
}

func auditJSON(value any) []byte {
	data, err := json.Marshal(value)
	if err != nil {
		return []byte("null")
	}
	return data
}

func auditDataForRack(rack *model.StockOpnameRackContext, so *model.StockOpname) *model.StockOpnameAuditData {
	return &model.StockOpnameAuditData{
		StockOpnameID:     so.StockOpnameID,
		RakID:             rack.RakID,
		RakName:           rack.RakName,
		CoordinatorID:     rack.CoorID,
		CoordinatorCode:   rack.CoorCode,
		CoordinatorStatus: rack.CoorStatus,
		RakUploadID:       rack.UploadID,
		RakUploadedAt:     rack.UploadedAt,
		PLU:               so.SoProductPLU,
		Barcode:           so.SoBarcode,
		Name:              so.SoProductName,
		DepartmentCode:    so.SoProductDept,
		Quantity:          so.StockOpnameQuantity,
		BuyPrice:          so.SoBuyPrice,
		SellPrice:         so.SoSellPrice,
	}
}

func auditDataForTarget(rack *model.StockOpnameRackContext, data *model.StockOpnameAuditData) *model.StockOpnameAuditData {
	data.RakID = rack.RakID
	data.RakName = rack.RakName
	data.CoordinatorID = rack.CoorID
	data.CoordinatorCode = rack.CoorCode
	data.CoordinatorStatus = rack.CoorStatus
	data.RakUploadID = rack.UploadID
	data.RakUploadedAt = rack.UploadedAt
	return data
}

func (s *StockOpnameServiceImpl) writeAudit(ctx context.Context, tx pgx.Tx, audit *model.StockOpnameAudit) error {
	audit.Actor = actorOrDefault(audit.Actor)
	return s.AuditRepository.Save(ctx, tx, audit)
}

func toStockOpnameResponse(summary *model.StockOpnameSummary) dto.StockOpnameResponse {
	return dto.StockOpnameResponse{
		Id:              summary.Id,
		Barcode:         summary.Barcode,
		PLU:             summary.ProductPLU,
		Name:            summary.Name,
		Quantity:        summary.Quantity,
		RackID:          summary.RackID,
		RackName:        summary.RackName,
		CoordinatorID:   summary.CoordinatorID,
		CoordinatorCode: summary.CoordinatorCode,
		UpdatedAt:       summary.UpdatedAt.Format(time.RFC3339),
	}
}

// snapshot builds the immutable history row from the resolved product.
// soBarcode is the actually scanned barcode (may differ from the
// product's primary barcode); it falls back to the primary barcode
// (or "" for products without any barcode) on PLU-based input.
func snapshot(quantity float64, rakID int, product *model.Product, scannedBarcode string) *model.StockOpname {
	soBarcode := scannedBarcode
	if soBarcode == "" {
		soBarcode = product.FirstBarcode()
	}
	return &model.StockOpname{
		StockOpnameQuantity: quantity,
		StockOpnameRakID:    rakID,
		SoProductPLU:        product.ProductPLU,
		SoProductName:       product.ProductName,
		SoProductDept:       product.ProductDepartmentCode,
		SoBarcode:           soBarcode,
		SoBuyPrice:          product.ProductBuyPrice,
		SoSellPrice:         product.ProductSellPrice,
	}
}

// Create implements [StockOpnameService]. It is create-only: an existing
// (rack, plu) pair is rejected with 409 and must be changed via Update.
func (s *StockOpnameServiceImpl) Create(ctx context.Context, req dto.CreateStockOpnameRequest) (dto.StockOpnameResponse, error) {
	if err := s.Validator.Struct(req); err != nil {
		return dto.StockOpnameResponse{}, err
	}
	if req.Barcode == "" && req.PLU == "" {
		return dto.StockOpnameResponse{}, &model.ValidationError{Detail: "barcode or plu is required"}
	}
	if err := validateQuantity(*req.Quantity); err != nil {
		return dto.StockOpnameResponse{}, err
	}

	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return dto.StockOpnameResponse{}, err
	}
	defer tx.Rollback(ctx)

	sesiId, err := s.StockOpnameRepository.FindSesiIdByRak(ctx, tx, req.RakID)
	if err != nil {
		return dto.StockOpnameResponse{}, err
	}
	if err := s.checkSessionReady(ctx, tx, sesiId); err != nil {
		return dto.StockOpnameResponse{}, err
	}

	rack, err := s.StockOpnameRepository.FindRackContextForUpdate(ctx, tx, req.RakID)
	if err != nil {
		return dto.StockOpnameResponse{}, err
	}
	if err := ensureCoordinatorWritable(rack.CoorStatus); err != nil {
		return dto.StockOpnameResponse{}, err
	}

	product, err := s.resolveProduct(ctx, tx, sesiId, req.Barcode, req.PLU)
	if err != nil {
		return dto.StockOpnameResponse{}, err
	}

	modelSO := snapshot(*req.Quantity, req.RakID, product, req.Barcode)

	if err := s.StockOpnameRepository.Create(ctx, tx, modelSO); err != nil {
		return dto.StockOpnameResponse{}, err
	}

	audit := &model.StockOpnameAudit{
		StockOpnameID: modelSO.StockOpnameID,
		SesiID:        rack.SesiID,
		SourceRakID:   &rack.RakID,
		TargetRakID:   &rack.RakID,
		SourceCoorID:  &rack.CoorID,
		TargetCoorID:  &rack.CoorID,
		Action:        "CREATE",
		AfterData:     auditJSON(auditDataForRack(rack, modelSO)),
		Actor:         req.Actor,
		Note:          req.Note,
	}
	if err := s.writeAudit(ctx, tx, audit); err != nil {
		return dto.StockOpnameResponse{}, err
	}
	if err := s.CoordinatorRepository.RecalcStatusAfterResultChange(ctx, tx, rack.CoorID); err != nil {
		return dto.StockOpnameResponse{}, err
	}

	modelSummary, err := s.StockOpnameRepository.FindById(ctx, tx, modelSO.StockOpnameID)
	if err != nil {
		return dto.StockOpnameResponse{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return dto.StockOpnameResponse{}, err
	}

	return toStockOpnameResponse(modelSummary), nil
}

// Delete implements [StockOpnameService].
func (s *StockOpnameServiceImpl) Delete(ctx context.Context, id int, actor string) error {
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	sesiId, err := s.StockOpnameRepository.FindSesiIdByStockOpnameID(ctx, tx, id)
	if err != nil {
		return err
	}
	if err := s.checkSessionWritable(ctx, tx, sesiId); err != nil {
		return err
	}

	before, err := s.StockOpnameRepository.FindByIdForUpdate(ctx, tx, id)
	if err != nil {
		return err
	}
	if err := ensureCoordinatorWritable(before.CoordinatorStatus); err != nil {
		return err
	}

	if err := s.StockOpnameRepository.Delete(ctx, tx, id); err != nil {
		return err
	}

	audit := &model.StockOpnameAudit{
		StockOpnameID: id,
		SesiID:        sesiId,
		SourceRakID:   &before.RakID,
		TargetRakID:   &before.RakID,
		SourceCoorID:  &before.CoordinatorID,
		TargetCoorID:  &before.CoordinatorID,
		Action:        "DELETE",
		BeforeData:    auditJSON(before),
		Actor:         actor,
	}
	if err := s.writeAudit(ctx, tx, audit); err != nil {
		return err
	}
	if err := s.CoordinatorRepository.RecalcStatusAfterResultChange(ctx, tx, before.CoordinatorID); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// FindAll implements [StockOpnameService].
func (s *StockOpnameServiceImpl) FindAll(ctx context.Context, pagination *dto.Pagination, search string, coorId *int, sesiId *int, rakId *int) ([]dto.StockOpnameResponse, error) {
	offset := (pagination.Page - 1) * pagination.Limit

	stockOpnames, total, err := s.StockOpnameRepository.FindAll(ctx, pagination.Limit, offset, search, sesiId, coorId, rakId)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.StockOpnameResponse, 0, len(stockOpnames))
	for _, so := range stockOpnames {
		responses = append(responses, toStockOpnameResponse(so))
	}

	pagination.TotalItems = total
	pagination.TotalPages = int(math.Ceil(float64(total) / float64(pagination.Limit)))

	return responses, nil
}

// FindById implements [StockOpnameService].
func (s *StockOpnameServiceImpl) FindById(ctx context.Context, id int) (dto.StockOpnameResponse, error) {
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return dto.StockOpnameResponse{}, err
	}
	defer tx.Rollback(ctx)

	modelSummary, err := s.StockOpnameRepository.FindById(ctx, tx, id)
	if err != nil {
		return dto.StockOpnameResponse{}, err
	}

	return toStockOpnameResponse(modelSummary), nil
}

// Update implements [StockOpnameService].
func (s *StockOpnameServiceImpl) Update(ctx context.Context, id int, req dto.UpdateStockOpnameRequest) (dto.StockOpnameResponse, error) {
	if err := s.Validator.Struct(req); err != nil {
		return dto.StockOpnameResponse{}, err
	}
	if err := validateQuantity(*req.Quantity); err != nil {
		return dto.StockOpnameResponse{}, err
	}

	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return dto.StockOpnameResponse{}, err
	}
	defer tx.Rollback(ctx)

	sesiId, err := s.StockOpnameRepository.FindSesiIdByStockOpnameID(ctx, tx, id)
	if err != nil {
		return dto.StockOpnameResponse{}, err
	}
	if err := s.checkSessionWritable(ctx, tx, sesiId); err != nil {
		return dto.StockOpnameResponse{}, err
	}

	before, err := s.StockOpnameRepository.FindByIdForUpdate(ctx, tx, id)
	if err != nil {
		return dto.StockOpnameResponse{}, err
	}
	if err := ensureCoordinatorWritable(before.CoordinatorStatus); err != nil {
		return dto.StockOpnameResponse{}, err
	}

	modelSO := &model.StockOpname{
		StockOpnameID:       id,
		StockOpnameQuantity: *req.Quantity,
	}
	if err := s.StockOpnameRepository.Update(ctx, tx, modelSO); err != nil {
		return dto.StockOpnameResponse{}, err
	}

	after := *before
	after.Quantity = *req.Quantity

	audit := &model.StockOpnameAudit{
		StockOpnameID: id,
		SesiID:        sesiId,
		SourceRakID:   &before.RakID,
		TargetRakID:   &before.RakID,
		SourceCoorID:  &before.CoordinatorID,
		TargetCoorID:  &before.CoordinatorID,
		Action:        "UPDATE",
		BeforeData:    auditJSON(before),
		AfterData:     auditJSON(&after),
		Actor:         req.Actor,
		Note:          req.Note,
	}
	if err := s.writeAudit(ctx, tx, audit); err != nil {
		return dto.StockOpnameResponse{}, err
	}
	if err := s.CoordinatorRepository.RecalcStatusAfterResultChange(ctx, tx, before.CoordinatorID); err != nil {
		return dto.StockOpnameResponse{}, err
	}

	result, err := s.StockOpnameRepository.FindById(ctx, tx, id)
	if err != nil {
		return dto.StockOpnameResponse{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return dto.StockOpnameResponse{}, err
	}

	return toStockOpnameResponse(result), nil
}

// Move implements [StockOpnameService].
func (s *StockOpnameServiceImpl) Move(ctx context.Context, req dto.MoveStockOpnameRequest) (dto.MoveStockOpnameResponse, error) {
	if err := s.Validator.Struct(req); err != nil {
		return dto.MoveStockOpnameResponse{}, err
	}
	if req.SourceRakID == req.TargetRakID {
		return dto.MoveStockOpnameResponse{}, &model.ValidationError{Detail: "source and target rack must differ"}
	}
	if req.ResultIDs != nil && len(*req.ResultIDs) == 0 {
		return dto.MoveStockOpnameResponse{}, &model.ValidationError{Detail: "result_ids must not be empty"}
	}

	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return dto.MoveStockOpnameResponse{}, err
	}
	defer tx.Rollback(ctx)

	sesiId, err := s.StockOpnameRepository.FindSesiIdByRak(ctx, tx, req.SourceRakID)
	if err != nil {
		return dto.MoveStockOpnameResponse{}, err
	}
	if err := s.checkSessionWritable(ctx, tx, sesiId); err != nil {
		return dto.MoveStockOpnameResponse{}, err
	}

	var sourceCtx, targetCtx *model.StockOpnameRackContext
	if req.SourceRakID < req.TargetRakID {
		sourceCtx, err = s.StockOpnameRepository.FindRackContextForUpdate(ctx, tx, req.SourceRakID)
		if err != nil {
			return dto.MoveStockOpnameResponse{}, err
		}
		targetCtx, err = s.StockOpnameRepository.FindRackContextForUpdate(ctx, tx, req.TargetRakID)
		if err != nil {
			return dto.MoveStockOpnameResponse{}, err
		}
	} else {
		targetCtx, err = s.StockOpnameRepository.FindRackContextForUpdate(ctx, tx, req.TargetRakID)
		if err != nil {
			return dto.MoveStockOpnameResponse{}, err
		}
		sourceCtx, err = s.StockOpnameRepository.FindRackContextForUpdate(ctx, tx, req.SourceRakID)
		if err != nil {
			return dto.MoveStockOpnameResponse{}, err
		}
	}

	if sourceCtx.SesiID != sesiId || targetCtx.SesiID != sesiId {
		return dto.MoveStockOpnameResponse{}, &model.ValidationError{Detail: "source and target rack must belong to the same session"}
	}
	if err := ensureCoordinatorWritable(sourceCtx.CoorStatus); err != nil {
		return dto.MoveStockOpnameResponse{}, err
	}
	if err := ensureCoordinatorWritable(targetCtx.CoorStatus); err != nil {
		return dto.MoveStockOpnameResponse{}, err
	}

	var ids []int
	if req.ResultIDs != nil {
		seen := make(map[int]struct{}, len(*req.ResultIDs))
		ids = make([]int, 0, len(*req.ResultIDs))
		for _, id := range *req.ResultIDs {
			if _, exists := seen[id]; exists {
				continue
			}
			seen[id] = struct{}{}
			ids = append(ids, id)
		}
	}

	sources, err := s.StockOpnameRepository.FindForMove(ctx, tx, req.SourceRakID, ids)
	if err != nil {
		return dto.MoveStockOpnameResponse{}, err
	}
	if len(sources) == 0 {
		return dto.MoveStockOpnameResponse{}, &model.ConflictError{Resource: "Rack", Detail: "Source rack has no results"}
	}
	if ids != nil && len(sources) != len(ids) {
		return dto.MoveStockOpnameResponse{}, &model.ValidationError{Detail: "result_ids must belong to the source rack"}
	}

	batchID := helper.NewUUID()
	movedCount := 0
	mergedCount := 0

	for _, source := range sources {
		existing, err := s.StockOpnameRepository.FindByRakAndPLUForUpdate(ctx, tx, targetCtx.RakID, source.PLU)
		if err != nil {
			return dto.MoveStockOpnameResponse{}, err
		}
		if existing != nil {
			mergedCount++
		}

		moved := &model.StockOpname{
			StockOpnameQuantity: source.Quantity,
			StockOpnameRakID:    targetCtx.RakID,
			SoProductPLU:        source.PLU,
			SoProductName:       source.Name,
			SoProductDept:       source.DepartmentCode,
			SoBarcode:           source.Barcode,
			SoBuyPrice:          source.BuyPrice,
			SoSellPrice:         source.SellPrice,
		}
		target, err := s.StockOpnameRepository.UpsertForMove(ctx, tx, moved)
		if err != nil {
			return dto.MoveStockOpnameResponse{}, err
		}
		target = auditDataForTarget(targetCtx, target)

		if err := s.StockOpnameRepository.Delete(ctx, tx, source.StockOpnameID); err != nil {
			return dto.MoveStockOpnameResponse{}, err
		}

		audit := &model.StockOpnameAudit{
			BatchID:       batchID,
			StockOpnameID: target.StockOpnameID,
			SesiID:        sesiId,
			SourceRakID:   &source.RakID,
			TargetRakID:   &targetCtx.RakID,
			SourceCoorID:  &source.CoordinatorID,
			TargetCoorID:  &targetCtx.CoorID,
			Action:        "MOVE",
			BeforeData:    auditJSON(source),
			AfterData:     auditJSON(target),
			Actor:         req.Actor,
			Note:          req.Note,
		}
		if err := s.writeAudit(ctx, tx, audit); err != nil {
			return dto.MoveStockOpnameResponse{}, err
		}
		movedCount++
	}

	if err := s.CoordinatorRepository.RecalcStatusAfterResultChange(ctx, tx, sourceCtx.CoorID); err != nil {
		return dto.MoveStockOpnameResponse{}, err
	}
	if targetCtx.CoorID != sourceCtx.CoorID {
		if err := s.CoordinatorRepository.RecalcStatusAfterResultChange(ctx, tx, targetCtx.CoorID); err != nil {
			return dto.MoveStockOpnameResponse{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return dto.MoveStockOpnameResponse{}, err
	}

	return dto.MoveStockOpnameResponse{MovedCount: movedCount, MergedCount: mergedCount}, nil
}

// FindAllForExport implements [StockOpnameService].
func (s *StockOpnameServiceImpl) FindAllForExport(ctx context.Context, sesiId *int, coorId *int) ([]dto.StockOpnameExportResponse, error) {
	exports, err := s.StockOpnameRepository.FindAllForExport(ctx, sesiId, coorId)
	if err != nil {
		return nil, err
	}

	var responses []dto.StockOpnameExportResponse
	for _, export := range exports {
		responses = append(responses, dto.StockOpnameExportResponse{
			Id:              export.Id,
			Barcode:         export.Barcode,
			PLU:             export.ProductPLU,
			Name:            export.Name,
			DepartmentCode:  export.DepartmentCode,
			BuyPrice:        export.BuyPrice,
			SellPrice:       export.SellPrice,
			Quantity:        export.Quantity,
			RackName:        export.RackName,
			CoordinatorCode: export.CoordinatorCode,
			UpdatedAt:       export.UpdatedAt.Format(time.RFC3339),
		})
	}

	return responses, nil
}
