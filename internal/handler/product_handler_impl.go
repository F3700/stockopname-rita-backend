package handler

import (
	"net/http"
	"stockopname-rita-backend/internal/dto"
	"stockopname-rita-backend/internal/helper"
	"stockopname-rita-backend/internal/service"

	"github.com/julienschmidt/httprouter"
)

type ProductHandlerImpl struct {
	ProductService service.ProductService
}

func NewProductHandler(productService service.ProductService) ProductHandler {
	return &ProductHandlerImpl{
		ProductService: productService,
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
