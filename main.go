package main

import (
	"anyprog/database"
	_ "anyprog/docs"
	"anyprog/handlers"
	"github.com/swaggo/http-swagger"
	"log"
	"net/http"
)

func main() {
	// Инициализация базы данных
	db := database.InitDB()
	defer db.Close()

	// Инициализация маршрутов

	router := handlers.SetupRoutes(db)
	router.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	// Запуск сервера
	log.Fatal(http.ListenAndServe(":8080", router))
}
