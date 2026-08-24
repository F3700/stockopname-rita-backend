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

	err = helper.ResponseJson(writer, http.StatusCreated, response)
	if err != nil {
		helper.WriteError(writer, http.StatusInternalServerError, "Failed to encode response", err)
		return
	}
}

// DeleteProduct implements [ProductHandler].
func (p *ProductHandlerImpl) DeleteProduct(writer http.ResponseWriter, req *http.Request, params httprouter.Params) {
	panic("unimplemented")
}

// GetProducts implements [ProductHandler].
func (p *ProductHandlerImpl) GetProducts(writer http.ResponseWriter, req *http.Request, params httprouter.Params) {
	panic("unimplemented")
}

// UpdateProduct implements [ProductHandler].
func (p *ProductHandlerImpl) UpdateProduct(writer http.ResponseWriter, req *http.Request, params httprouter.Params) {
	panic("unimplemented")
}
