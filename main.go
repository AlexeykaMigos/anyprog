package main

import (
	"anyprog/database"
	"anyprog/handlers"
	"log"
	"net/http"
)

func main() {
	// Инициализация базы данных
	db := database.InitDB()
	defer db.Close()

	// Инициализация маршрутов
	router := handlers.SetupRoutes(db)

	// Запуск сервера
	log.Fatal(http.ListenAndServe(":8080", router))
}
