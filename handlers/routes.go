package handlers

import (
	"context"
	"database/sql"
	"github.com/gorilla/mux"
	"net/http"
)

func SetupRoutes(db *sql.DB) *mux.Router {
	router := mux.NewRouter()

	//переменные в пути
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			vars := mux.Vars(r)
			ctx := context.WithValue(r.Context(), "vars", vars)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})

	router.HandleFunc("/api/products", GetProducts(db)).Methods("GET")
	router.HandleFunc("/api/products", AddProduct(db)).Methods("POST")
	router.HandleFunc("/api/products/{id}", UpdateProduct(db)).Methods("PUT")
	router.HandleFunc("/api/products/{id}/rollback", RollbackProduct(db)).Methods("POST")
	router.HandleFunc("/api/products/{id}/history", GetProductHistory(db)).Methods("GET")

	return router
}
