package handler

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func GetProducts(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Retrieve all products!"))
}
