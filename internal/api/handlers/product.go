package handlers

import (
	"anyprog/internal/models"
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
)

func GetProducts(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT * FROM products")
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var products []models.Product
		for rows.Next() {
			var p models.Product
			if err := rows.Scan(&p.ID, &p.Title, &p.Description, &p.Price, &p.Version); err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			products = append(products, p)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(products)
	}
}

func AddProduct(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var p models.Product
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		err := db.QueryRow(
			"INSERT INTO products (title, description, price, version) VALUES ($1, $2, $3, $4) RETURNING id",
			p.Title, p.Description, p.Price, p.Version,
		).Scan(&p.ID)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(p)
	}
}

func UpdateProduct(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := r.Context().Value("vars").(map[string]string)
		id := vars["id"]

		var updatedProduct models.Product
		if err := json.NewDecoder(r.Body).Decode(&updatedProduct); err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		tx, err := db.Begin()
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		defer tx.Rollback()

		var currentProduct models.Product
		if err := tx.QueryRow(
			"SELECT id, title, description, price, version FROM products WHERE id = $1", id,
		).Scan(&currentProduct.ID, &currentProduct.Title, &currentProduct.Description, &currentProduct.Price, &currentProduct.Version); err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "Product not found", http.StatusNotFound)
			} else {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}

		if _, err := tx.Exec(
			"INSERT INTO product_versions (product_id, title, description, price, version) VALUES ($1, $2, $3, $4, $5)",
			currentProduct.ID, currentProduct.Title, currentProduct.Description, currentProduct.Price, currentProduct.Version,
		); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		updatedProduct.Version = currentProduct.Version + 1
		if _, err := tx.Exec(
			"UPDATE products SET title = $1, description = $2, price = $3, version = $4 WHERE id = $5",
			updatedProduct.Title, updatedProduct.Description, updatedProduct.Price, updatedProduct.Version, id,
		); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if err := tx.Commit(); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "Product updated successfully"})
	}
}

func RollbackProduct(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := r.Context().Value("vars").(map[string]string)
		id := vars["id"]
		version := r.URL.Query().Get("version")

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
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		defer tx.Rollback()

		var rollbackProduct models.Product
		if err := tx.QueryRow(
			"SELECT title, description, price, version FROM product_versions WHERE product_id = $1 AND version = $2",
			id, versionInt,
		).Scan(&rollbackProduct.Title, &rollbackProduct.Description, &rollbackProduct.Price, &rollbackProduct.Version); err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "Specified version not found", http.StatusNotFound)
			} else {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}

		if _, err := tx.Exec(
			"UPDATE products SET title = $1, description = $2, price = $3, version = $4 WHERE id = $5",
			rollbackProduct.Title, rollbackProduct.Description, rollbackProduct.Price, rollbackProduct.Version, id,
		); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if err := tx.Commit(); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "Product rolled back successfully"})
	}
}

func GetProductHistory(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := r.Context().Value("vars").(map[string]string)
		id := vars["id"]

		rows, err := db.Query(
			"SELECT id, title, description, price, version FROM product_versions WHERE product_id = $1 ORDER BY id DESC",
			id,
		)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var history []map[string]interface{}
		for rows.Next() {
			var (
				versionID   int
				title       string
				description sql.NullString
				price       float64
				version     int
			)
			if err := rows.Scan(&versionID, &title, &description, &price, &version); err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			history = append(history, map[string]interface{}{
				"version_id":  versionID,
				"title":       title,
				"description": description.String,
				"price":       price,
				"version":     version,
			})
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(history)
	}
}
