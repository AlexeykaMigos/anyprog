package routes

import (
	"anyprog/internal/api/handlers"
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

	router.HandleFunc("/api/products", handlers.GetProducts(db)).Methods("GET")
	router.HandleFunc("/api/products", handlers.AddProduct(db)).Methods("POST")
	router.HandleFunc("/api/products/{id}", handlers.UpdateProduct(db)).Methods("PUT")
	router.HandleFunc("/api/products/{id}/rollback", handlers.RollbackProduct(db)).Methods("POST")
	router.HandleFunc("/api/products/{id}/history", handlers.GetProductHistory(db)).Methods("GET")

	return router
}
