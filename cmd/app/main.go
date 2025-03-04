package main

import (
	_ "anyprog/docs"
	"anyprog/internal/repository/postgresql"
	"anyprog/internal/routes"
	"github.com/swaggo/http-swagger"
	"log"
	"net/http"
)

func main() {
	// Инициализация базы данных
	db := postgresql.InitDB()
	defer db.Close()

	// Инициализация маршрутов

	router := routes.SetupRoutes(db)
	router.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	//запуск сервера
	log.Fatal(http.ListenAndServe(":8080", router))
}
