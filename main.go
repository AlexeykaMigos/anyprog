package main

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"log"
)

func main() {
	connStr := "postgres://postgres:secret@localhost:5432/pgdb?sslmode=disable"

	db, err := sql.Open("postgres", connStr)

	defer db.Close()
	if err != nil {
		log.Fatal(err)
	}
	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	createTableProduct(db)

	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Hello World",
		})
	})
	//r.POST("/api/products", addProduct)

	if err := r.Run(":8080"); err != nil {
		log.Fatal("error running server", err.Error())
	}
}

func createTableProduct(db *sql.DB) {
	query := `CREATE TABLE IF NOT EXISTS product (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    price DECIMAL(10, 2) NOT NULL,
    description TEXT
	);`

	_, err := db.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}
