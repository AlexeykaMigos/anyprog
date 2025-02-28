package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"os"
	"strconv"
)

var db *sql.DB

type Product struct {
	ID          int     `db:"id" json:"id"`
	Title       string  `db:"title" json:"title"`
	Description *string `db:"description" json:"description"`
	Price       float64 `db:"price" json:"price"`
	Version     int     `db:"version" json:"version"`
}

func initDB() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName,
	)

	db, err = sql.Open("postgres", connStr)
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
}

func main() {

	initDB()

	router := mux.NewRouter()
	router.HandleFunc("/api/products", getProducts(db)).Methods("GET")
	router.HandleFunc("/api/products", addProduct(db)).Methods("POST")
	router.HandleFunc("/api/products/{id}", updateProduct(db)).Methods("PUT")
	router.HandleFunc("/api/products/{id}/rollback", rollbackProduct(db)).Methods("POST")
	router.HandleFunc("/api/products/{id}/history", getProductHistory(db)).Methods("GET")
	log.Fatal(http.ListenAndServe(":8080", jsonContentTypeMiddleware(router)))
}

func jsonContentTypeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func getProducts(db *sql.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query(`SELECT * FROM products`)
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		products := []Product{}
		for rows.Next() {
			var p Product
			if err := rows.Scan(&p.ID, &p.Title, &p.Description, &p.Price, &p.Version); err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				log.Fatal(err)
			}
			products = append(products, p)
		}
		if err := rows.Err(); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			log.Fatal(err)
		}
		json.NewEncoder(w).Encode(products)

	})
}

func addProduct(db *sql.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var p Product
		err := json.NewDecoder(r.Body).Decode(&p)
		if err != nil {
			log.Println("Error decoding JSON:", err)
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		err = db.QueryRow("INSERT INTO products (title, description, price, version) VALUES ($1, $2, $3, $4) RETURNING id", p.Title, p.Description, p.Price, p.Version).Scan(&p.ID)
		if err != nil {
			log.Println("Error inserting product:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(p)
	})
}

func updateProduct(db *sql.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		var updatedProduct Product
		if err := json.NewDecoder(r.Body).Decode(&updatedProduct); err != nil {
			http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
			return
		}

		//стартуем транзакцию
		tx, err := db.Begin()
		if err != nil {
			log.Println("Failed to start transaction:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		defer tx.Rollback() // если всё плохо, то роллбэк

		//получаем текущую инфу
		var currentProduct Product
		err = tx.QueryRow("SELECT id, title, description, price, version FROM products WHERE id = $1", id).Scan(
			&currentProduct.ID, &currentProduct.Title, &currentProduct.Description, &currentProduct.Price, &currentProduct.Version,
		)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "Product not found", http.StatusNotFound)
			} else {
				log.Println("Failed to retrieve current product:", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}

		//сохраняем старую версию
		_, err = tx.Exec(
			"INSERT INTO product_versions (product_id, title, description, price, version) VALUES ($1, $2, $3, $4, $5)",
			currentProduct.ID, currentProduct.Title, currentProduct.Description, currentProduct.Price, currentProduct.Version,
		)
		if err != nil {
			log.Println("Failed to save old version:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		//добавляем версию + обновляем инфу в бдшке
		updatedProduct.Version = currentProduct.Version + 1
		_, err = tx.Exec(
			"UPDATE products SET title = $1, description = $2, price = $3, version = $4 WHERE id = $5",
			updatedProduct.Title, updatedProduct.Description, updatedProduct.Price, updatedProduct.Version, id,
		)
		if err != nil {
			log.Println("Failed to update product:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if err := tx.Commit(); err != nil {
			log.Println("Failed to commit transaction:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	})
}

func rollbackProduct(db *sql.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		vars := mux.Vars(r)
		id := vars["id"]
		version := r.URL.Query().Get("version") //получаем версию по строке

		if version == "" {
			http.Error(w, "Version parameter is required", http.StatusBadRequest)
			return
		}

		versionInt, err := strconv.Atoi(version)
		if err != nil {
			http.Error(w, "Invalid version parameter", http.StatusBadRequest)
			return
		}

		tx, err := db.Begin()
		if err != nil {
			log.Println("Failed to start transaction:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		defer tx.Rollback() //если всё плохо

		var rollbackProduct Product
		err = tx.QueryRow(
			"SELECT title, description, price, version FROM product_versions WHERE product_id = $1 AND version = $2",
			id, versionInt,
		).Scan(&rollbackProduct.Title, &rollbackProduct.Description, &rollbackProduct.Price, &rollbackProduct.Version)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "Specified version not found", http.StatusNotFound)
			} else {
				log.Println("Failed to retrieve rollback version:", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}

		//обновляем в таблице
		_, err = tx.Exec(
			"UPDATE products SET title = $1, description = $2, price = $3, version = $4 WHERE id = $5",
			rollbackProduct.Title, rollbackProduct.Description, rollbackProduct.Price, rollbackProduct.Version, id,
		)
		if err != nil {
			log.Println("Failed to update product with rollback data:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		//коммитим транзакцию
		if err := tx.Commit(); err != nil {
			log.Println("Failed to commit transaction:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		//вернем успех
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "Product rolled back successfully"})
	})
}
func getProductHistory(db *sql.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		//получаем все версии
		rows, err := db.Query(
			"SELECT id, title, description, price, version FROM product_versions WHERE product_id = $1 ORDER BY id DESC",
			id,
		)
		if err != nil {
			log.Println("Failed to retrieve product history:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		//срез для храненияя истории
		var history []map[string]interface{}

		//итерируем
		for rows.Next() {
			var (
				versionID   int
				title       string
				description sql.NullString
				price       float64
				version     int
			)
			if err := rows.Scan(&versionID, &title, &description, &price, &version); err != nil {
				log.Println("Failed to scan row:", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			// добавляем в историю версию
			history = append(history, map[string]interface{}{
				"version_id":  versionID,
				"title":       title,
				"description": description.String,
				"price":       price,
				"version":     version,
			})
		}

		if err := rows.Err(); err != nil {
			log.Println("Error iterating over rows:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(history)
	})
}
