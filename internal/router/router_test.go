package router

import (
	"testing"

	"github.com/go-playground/validator/v10"
)

// NewRouter must register without panicking (httprouter panics on
// conflicting static/param routes, e.g. DELETE /products vs :id).
func TestNewRouterRegisters(t *testing.T) {
	router := NewRouter(validator.New(), nil)
	for _, method := range []string{"GET", "POST", "PATCH", "DELETE"} {
		handle, _, _ := router.Lookup(method, "/products")
		_ = handle
	}
	for _, route := range []struct{ method, path string }{
		{"POST", "/products/import"},
		{"POST", "/products/clear"},
		{"DELETE", "/products"},
		{"DELETE", "/products/:id"},
		{"POST", "/stockopname/results"},
		{"POST", "/stockopname/results/racks"},
	} {
		handle, _, _ := router.Lookup(route.method, route.path)
		if handle == nil {
			t.Errorf("expected handler for %s %s", route.method, route.path)
		}
	}
}
