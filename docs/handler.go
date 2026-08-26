package docs

import (
	_ "embed"
	"net/http"

	"github.com/julienschmidt/httprouter"
	v5 "github.com/swaggest/swgui/v5"
)

//go:embed apispec.json
var apiSpec []byte

func RegisterRoutes(router *httprouter.Router) {
	docsUI := v5.New(
		"Rita Stock Opname API",
		"/docs/apispec.json",
		"/docs/",
	)

	router.GET("/docs", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		http.Redirect(w, r, "/docs/", http.StatusMovedPermanently)
	})
	router.Handler(
		http.MethodGet,
		"/docs/*path",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/docs/apispec.json" {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write(apiSpec)
				return
			}

			docsUI.ServeHTTP(w, r)
		}),
	)
}
