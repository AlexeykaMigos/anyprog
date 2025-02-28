package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func InitDB() *sql.DB {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"),
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS products (id SERIAL PRIMARY KEY,title VARCHAR(255) NOT NULL,description VARCHAR(255) NOT NULL,price DECIMAL(10, 2) NOT NULL,version INT NOT NULL)")
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS product_versions (id SERIAL PRIMARY KEY,product_id INT NOT NULL REFERENCES products(id),title TEXT NOT NULL,description TEXT,price DECIMAL(10, 2) NOT NULL,version INT NOT NULL)")
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Connected to the database")
	return db
}
