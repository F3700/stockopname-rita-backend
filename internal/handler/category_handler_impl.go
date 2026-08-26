package handler

import (
	"net/http"
	"stockopname-rita-backend/internal/dto"
	"stockopname-rita-backend/internal/helper"
	"stockopname-rita-backend/internal/service"

	"github.com/julienschmidt/httprouter"
)

type CategoryHandlerImpl struct {
	CategoryService service.CategoryService
}

func NewCategoryHandler(categoryService service.CategoryService) CategoryHandler {
	return &CategoryHandlerImpl{
		CategoryService: categoryService,
	}
}

// CreateCategory implements [CategoryHandler].
func (c *CategoryHandlerImpl) CreateCategory(writer http.ResponseWriter, req *http.Request, params httprouter.Params) {
	categoryCreateRequest := dto.CategoryCreateRequest{}

	if err := helper.ReadFromRequestBody(req, &categoryCreateRequest); err != nil {
		helper.WriteError(writer, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	categoryCreateResponse, err := c.CategoryService.Create(req.Context(), categoryCreateRequest)
	if err != nil {
		helper.WriteError(writer, http.StatusInternalServerError, "Failed to create category", err)
		return
	}

	response := dto.Response{
		Message: "Category created successfully",
		Data:    categoryCreateResponse,
	}

	if err := helper.ResponseJson(writer, http.StatusCreated, response); err != nil {
		helper.WriteError(writer, http.StatusInternalServerError, "Failed to send response", err)
		return
	}
}

// DeleteCategory implements [CategoryHandler].
func (c *CategoryHandlerImpl) DeleteCategory(writer http.ResponseWriter, req *http.Request, params httprouter.Params) {
	id, err := helper.ParseIntParam(params, "id")
	if err != nil {
		helper.WriteError(writer, http.StatusBadRequest, "Invalid category ID", err)
		return
	}

	if err := c.CategoryService.Delete(req.Context(), id); err != nil {
		helper.WriteError(writer, http.StatusInternalServerError, "Failed to delete category", err)
		return
	}

	response := dto.Response{
		Message: "Category deleted successfully",
		Data:    nil,
	}

	if err := helper.ResponseJson(writer, http.StatusOK, response); err != nil {
		helper.WriteError(writer, http.StatusInternalServerError, "Failed to send response", err)
		return
	}
}

// GetCategories implements [CategoryHandler].
func (c *CategoryHandlerImpl) GetCategories(writer http.ResponseWriter, req *http.Request, params httprouter.Params) {
	categories, err := c.CategoryService.FindAll(req.Context())
	if err != nil {
		helper.WriteError(writer, http.StatusInternalServerError, "Failed to get categories", err)
		return
	}

	response := dto.Response{
		Message: "Categories retrieved successfully",
		Data:    categories,
	}

	if err := helper.ResponseJson(writer, http.StatusOK, response); err != nil {
		helper.WriteError(writer, http.StatusInternalServerError, "Failed to send response", err)
		return
	}
}

// UpdateCategory implements [CategoryHandler].
func (c *CategoryHandlerImpl) UpdateCategory(writer http.ResponseWriter, req *http.Request, params httprouter.Params) {
	id, err := helper.ParseIntParam(params, "id")
	if err != nil {
		helper.WriteError(writer, http.StatusBadRequest, "Invalid category ID", err)
		return
	}

	categoryUpdateRequest := dto.CategoryUpdateRequest{}
	if err := helper.ReadFromRequestBody(req, &categoryUpdateRequest); err != nil {
		helper.WriteError(writer, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	categoryUpdateResponse, err := c.CategoryService.Update(req.Context(), id, categoryUpdateRequest)
	if err != nil {
		helper.WriteError(writer, http.StatusInternalServerError, "Failed to update category", err)
		return
	}

	response := dto.Response{
		Message: "Category updated successfully",
		Data:    categoryUpdateResponse,
	}

	if err := helper.ResponseJson(writer, http.StatusOK, response); err != nil {
		helper.WriteError(writer, http.StatusInternalServerError, "Failed to send response", err)
		return
	}
}
