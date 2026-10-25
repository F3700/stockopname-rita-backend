package router

import (
	"testing"

	"github.com/go-playground/validator/v10"
)

// NewRouter must register without panicking (httprouter panics on
// conflicting static/param routes, e.g. static /products/clear next to
// DELETE /products/:id).
func TestNewRouterRegisters(t *testing.T) {
	router := NewRouter(validator.New(), nil)

	for _, route := range []struct{ method, path string }{
		{"GET", "/products"},
		{"POST", "/products"},
		{"PATCH", "/products/:id"},
		{"DELETE", "/products/:id"},
		{"POST", "/products/import"},
		{"POST", "/products/clear"},
		{"DELETE", "/stockopname/sessions/:id"},
		{"POST", "/stockopname/results"},
		{"POST", "/stockopname/results/racks"},
	} {
		handle, _, _ := router.Lookup(route.method, route.path)
		if handle == nil {
			t.Errorf("expected handler for %s %s", route.method, route.path)
		}
	}

	for _, route := range []struct{ method, path string }{
		{"DELETE", "/products"},
		{"GET", "/products/sync"},
		{"GET", "/products/last-session"},
		{"GET", "/deleted/products"},
	} {
		handle, _, _ := router.Lookup(route.method, route.path)
		if handle != nil {
			t.Errorf("expected no handler for %s %s", route.method, route.path)
		}
	}
}
