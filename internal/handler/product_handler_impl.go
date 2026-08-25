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
		helper.WriteError(writer, http.StatusInternalServerError, "Failed to create product", err)
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
		helper.WriteError(writer, http.StatusInternalServerError, "Failed to delete product", err)
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
	date, err := helper.ParseDate(req.URL.Query().Get("updated_after"))
	if err != nil {
		helper.WriteError(writer, http.StatusBadRequest, "Invalid date format", err)
		return
	}

	products, err := p.ProductService.FindAll(req.Context(), date)
	if err != nil {
		helper.WriteError(writer, http.StatusInternalServerError, "Failed to fetch products", err)
		return
	}

	response := dto.Response{
		Message: "Products fetched successfully",
		Data:    products,
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
		helper.WriteError(writer, http.StatusInternalServerError, "Failed to update product", err)
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
