package main

import (
	_ "anyprog/docs"
	"anyprog/internal/api/handler"
	"anyprog/internal/api/migrations"
	"anyprog/internal/config"
	"anyprog/internal/repository/postgresql"
	"anyprog/internal/usecase"
	"database/sql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/swaggo/http-swagger"
	"log"
	"net/http"
)

func main() {
	cfg, err := config.LoadConfig(".")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := sql.Open("postgres", cfg.DBConnectionString())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	if err := migrations.RunMigrations(db); err != nil {
		log.Fatalf("Failed to apply migrations: %v", err)
	}

	productRepo := postgresql.NewUserRepositoryPostgresql(db)

	productUseCase := usecase.NewProductUseCase(productRepo)

	productHandler := handler.NewProductHandler(productUseCase)

	r := mux.NewRouter()

	r.HandleFunc("/api/products", productHandler.GetAll).Methods("GET")
	r.HandleFunc("/api/products/{id}", productHandler.GetByID).Methods("GET")
	r.HandleFunc("/api/products", productHandler.Create).Methods("POST")
	r.HandleFunc("/api/products/{id}", productHandler.Update).Methods("PUT")
	r.HandleFunc("/api/products/{id}/rollback", productHandler.Rollback).Methods("POST")
	r.HandleFunc("/api/products/{id}/history", productHandler.GetHistory).Methods("GET")

	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	log.Fatal(http.ListenAndServe(":8080", r))
}
