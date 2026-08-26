package handler

import (
	"net/http"
	"stockopname-rita-backend/internal/dto"
	"stockopname-rita-backend/internal/helper"
	"stockopname-rita-backend/internal/service"

	"github.com/julienschmidt/httprouter"
)

type DeletedProductHandlerImpl struct {
	DeletedProductService service.DeletedProductService
}

func NewDeletedProductHandler(deletedProductService service.DeletedProductService) DeletedProductHandler {
	return &DeletedProductHandlerImpl{
		DeletedProductService: deletedProductService,
	}
}

// DeleteDeletedProduct implements [DeletedProductHandler].
func (d *DeletedProductHandlerImpl) DeleteDeletedProducts(writer http.ResponseWriter, req *http.Request, params httprouter.Params) {
	if err := d.DeletedProductService.Delete(req.Context()); err != nil {
		helper.WriteError(writer, http.StatusInternalServerError, "failed to delete deleted products", err)
		return
	}
	response := dto.Response{
		Message: "Deleted products deleted successfully",
		Data:    nil,
	}

	if err := helper.ResponseJson(writer, http.StatusOK, response); err != nil {
		helper.WriteError(writer, http.StatusInternalServerError, "failed to respond", err)
		return
	}
}

// GetDeletedProducts implements [DeletedProductHandler].
func (d *DeletedProductHandlerImpl) GetDeletedProducts(writer http.ResponseWriter, req *http.Request, params httprouter.Params) {
	updatedAfter, err := helper.ParseDate(req.URL.Query().Get("updated_after"))
	if err != nil {
		helper.WriteError(writer, http.StatusBadRequest, "invalid date format for updated_after", err)
		return
	}

	deletedProducts, err := d.DeletedProductService.FindAll(req.Context(), updatedAfter)
	if err != nil {
		helper.WriteError(writer, http.StatusInternalServerError, "failed to get deleted products", err)
		return
	}

	response := dto.Response{
		Message: "Deleted products fetched successfully",
		Data:    deletedProducts,
	}

	if err := helper.ResponseJson(writer, http.StatusOK, response); err != nil {
		helper.WriteError(writer, http.StatusInternalServerError, "failed to respond", err)
		return
	}
}
