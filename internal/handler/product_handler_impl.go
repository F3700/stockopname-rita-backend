package handler

import (
	"fmt"
	"io"
	"net/http"
	"stockopname-rita-backend/internal/dto"
	"stockopname-rita-backend/internal/helper"
	"stockopname-rita-backend/internal/service"

	"github.com/julienschmidt/httprouter"
)

type ProductHandlerImpl struct {
	ProductService service.ProductService
	ImportService  service.ImportService
}

func NewProductHandler(productService service.ProductService, importService service.ImportService) ProductHandler {
	return &ProductHandlerImpl{
		ProductService: productService,
		ImportService:  importService,
	}
}

// CreateProduct implements [ProductHandler].
func (p *ProductHandlerImpl) CreateProduct(writer http.ResponseWriter, req *http.Request, params httprouter.Params) {
	productCreateRequest := dto.ProductCreateRequest{}
	err := helper.ReadFromRequestBody(req, &productCreateRequest)
	if err != nil {
		helper.WriteError(writer, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	productCreateResponse, err := p.ProductService.Create(req.Context(), productCreateRequest)
	if err != nil {
		helper.WriteServiceError(writer, err)
		return
	}

	response := dto.Response{
		Message: "Product created successfully",
		Data:    productCreateResponse,
	}

	if err := helper.ResponseJson(writer, http.StatusCreated, response); err != nil {
		helper.WriteError(writer, http.StatusInternalServerError, "Failed to encode response", err)
		return
	}
}

// DeleteProduct implements [ProductHandler].
func (p *ProductHandlerImpl) DeleteProduct(writer http.ResponseWriter, req *http.Request, params httprouter.Params) {
	id, err := helper.ParseIntParam(params, "id")
	if err != nil {
		helper.WriteError(writer, http.StatusBadRequest, "Invalid product ID", err)
		return
	}

	if err := p.ProductService.Delete(req.Context(), id); err != nil {
		helper.WriteServiceError(writer, err)
		return
	}

	response := dto.Response{
		Message: "Product deleted successfully",
		Data:    nil,
	}

	if err := helper.ResponseJson(writer, http.StatusOK, response); err != nil {
		helper.WriteError(writer, http.StatusInternalServerError, "Failed to encode response", err)
		return
	}
}

// GetProducts implements [ProductHandler].
func (p *ProductHandlerImpl) GetProducts(writer http.ResponseWriter, req *http.Request, params httprouter.Params) {
	page, err := helper.ParseIntQueryNotNull(req, "page")
	if err != nil {
		helper.WriteError(writer, http.StatusBadRequest, "Invalid query page parameter", err)
		return
	}

	limit, err := helper.ParseIntQueryNotNull(req, "limit")
	if err != nil {
		helper.WriteError(writer, http.StatusBadRequest, "Invalid query limit parameter", err)
		return
	}

	search := req.URL.Query().Get("search")

	pagination := &dto.Pagination{
		Page:  page,
		Limit: limit,
	}

	products, err := p.ProductService.FindAll(req.Context(), pagination, search)
	if err != nil {
		helper.WriteServiceError(writer, err)
		return
	}

	response := dto.ResponsePagination{
		Message:    "Products fetched successfully",
		Data:       products,
		Pagination: pagination,
	}

	if err := helper.ResponseJson(writer, http.StatusOK, response); err != nil {
		helper.WriteError(writer, http.StatusInternalServerError, "Failed to encode response", err)
		return
	}
}

// GetProductsLastSession implements [ProductHandler].
func (p *ProductHandlerImpl) GetProductsLastSession(writer http.ResponseWriter, req *http.Request, params httprouter.Params) {
	limit := 20
	if l := req.URL.Query().Get("limit"); l != "" {
		parsed, err := helper.ParseIntQueryNotNull(req, "limit")
		if err != nil {
			helper.WriteError(writer, http.StatusBadRequest, "Invalid query limit parameter", err)
			return
		}
		limit = parsed
	}

	products, err := p.ProductService.FindAllLastSession(req.Context(), limit)
	if err != nil {
		helper.WriteServiceError(writer, err)
		return
	}

	response := dto.Response{
		Message: "Products last session retrieved successfully",
		Data:    products,
	}

	if err := helper.ResponseJson(writer, http.StatusOK, response); err != nil {
		helper.WriteError(writer, http.StatusInternalServerError, "Failed to encode response", err)
		return
	}
}

// SyncProducts implements [ProductHandler].
func (p *ProductHandlerImpl) SyncProducts(writer http.ResponseWriter, req *http.Request, params httprouter.Params) {
	page, err := helper.ParseIntQueryNotNull(req, "page")
	if err != nil {
		helper.WriteError(writer, http.StatusBadRequest, "Invalid query page parameter", err)
		return
	}

	limit, err := helper.ParseIntQueryNotNull(req, "limit")
	if err != nil {
		helper.WriteError(writer, http.StatusBadRequest, "Invalid query limit parameter", err)
		return
	}

	date, err := helper.ParseDate(req.URL.Query().Get("updated_after"))
	if err != nil {
		helper.WriteError(writer, http.StatusBadRequest, "Invalid date format", err)
		return
	}

	if date == nil {
		helper.WriteError(writer, http.StatusBadRequest, "updated_after parameter is required", nil)
		return
	}

	pagination := &dto.Pagination{
		Page:  page,
		Limit: limit,
	}

	products, err := p.ProductService.FindAllUpdatedAfter(req.Context(), pagination, *date)
	if err != nil {
		helper.WriteServiceError(writer, err)
		return
	}

	response := dto.ResponsePagination{
		Message:    "Products synced successfully",
		Data:       products,
		Pagination: pagination,
	}

	if err := helper.ResponseJson(writer, http.StatusOK, response); err != nil {
		helper.WriteError(writer, http.StatusInternalServerError, "Failed to encode response", err)
		return
	}
}

// ImportProducts implements [ProductHandler].
// Expects multipart form with "produk" (PRODUK.DBF) and "barcode"
// (BARCODE.DBF) files. Optional ?dry_run=true validates without writing.
func (p *ProductHandlerImpl) ImportProducts(writer http.ResponseWriter, req *http.Request, params httprouter.Params) {
	if err := req.ParseMultipartForm(32 << 20); err != nil {
		helper.WriteError(writer, http.StatusBadRequest, "Invalid multipart form", err)
		return
	}

	produkData, err := readUpload(req, "produk")
	if err != nil {
		helper.WriteError(writer, http.StatusBadRequest, "Missing or invalid produk file", err)
		return
	}
	barcodeData, err := readUpload(req, "barcode")
	if err != nil {
		helper.WriteError(writer, http.StatusBadRequest, "Missing or invalid barcode file", err)
		return
	}

	dryRun := req.URL.Query().Get("dry_run") == "true"

	result, err := p.ImportService.Import(req.Context(), produkData, barcodeData, dryRun)
	if err != nil {
		helper.WriteServiceError(writer, err)
		return
	}

	message := "Master data imported successfully"
	if dryRun {
		message = "Master data validation completed (dry run, nothing written)"
	}
	if err := helper.ResponseJson(writer, http.StatusOK, dto.Response{Message: message, Data: result}); err != nil {
		helper.WriteError(writer, http.StatusInternalServerError, "Failed to encode response", err)
	}
}

func readUpload(req *http.Request, field string) ([]byte, error) {
	file, _, err := req.FormFile(field)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("empty file: %s", field)
	}
	return data, nil
}

// ClearProducts implements [ProductHandler].
// Removes ALL master data (products, barcodes, tombstones).
// Requires ?confirm=true. Stock opname history is untouched.
func (p *ProductHandlerImpl) ClearProducts(writer http.ResponseWriter, req *http.Request, params httprouter.Params) {
	if req.URL.Query().Get("confirm") != "true" {
		helper.WriteError(writer, http.StatusBadRequest, "Master data clear requires ?confirm=true", nil)
		return
	}

	if err := p.ProductService.Clear(req.Context()); err != nil {
		helper.WriteServiceError(writer, err)
		return
	}

	if err := helper.ResponseJson(writer, http.StatusOK, dto.Response{Message: "Master data cleared successfully", Data: nil}); err != nil {
		helper.WriteError(writer, http.StatusInternalServerError, "Failed to encode response", err)
	}
}

// UpdateProduct implements [ProductHandler].
func (p *ProductHandlerImpl) UpdateProduct(writer http.ResponseWriter, req *http.Request, params httprouter.Params) {
	id, err := helper.ParseIntParam(params, "id")
	if err != nil {
		helper.WriteError(writer, http.StatusBadRequest, "Invalid product ID", err)
		return
	}

	productUpdateRequest := dto.ProductUpdateRequest{}
	if err := helper.ReadFromRequestBody(req, &productUpdateRequest); err != nil {
		helper.WriteError(writer, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	productResponse, err := p.ProductService.Update(req.Context(), id, productUpdateRequest)
	if err != nil {
		helper.WriteServiceError(writer, err)
		return
	}

	response := dto.Response{
		Message: "Product updated successfully",
		Data:    productResponse,
	}

	if err := helper.ResponseJson(writer, http.StatusOK, response); err != nil {
		helper.WriteError(writer, http.StatusInternalServerError, "Failed to encode response", err)
		return
	}
}
