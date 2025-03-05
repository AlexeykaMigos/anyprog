package postgresql

import (
	"anyprog/internal/models"
	"database/sql"
	"fmt"
)

type UserRepositoryPostgresql struct {
	db *sql.DB
}

func NewUserRepositoryPostgresql(db *sql.DB) *UserRepositoryPostgresql {
	return &UserRepositoryPostgresql{db: db}
}

func (r *UserRepositoryPostgresql) GetAll() ([]models.Product, error) {
	rows, err := r.db.Query("SELECT id, title, description, price, version FROM products")
	if err != nil {
		return nil, fmt.Errorf("failed to get products: %w", err)
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var p models.Product
		if err := rows.Scan(&p.ID, &p.Title, &p.Description, &p.Price, &p.Version); err != nil {
			return nil, fmt.Errorf("failed to scan product: %w", err)
		}
		products = append(products, p)
	}
	return products, nil
}

func (r *UserRepositoryPostgresql) GetByID(id int) (*models.Product, error) {
	var p models.Product
	err := r.db.QueryRow("SELECT id, title, description, price, version FROM products WHERE id = $1", id).
		Scan(&p.ID, &p.Title, &p.Description, &p.Price, &p.Version)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("product not found")
		}
		return nil, fmt.Errorf("failed to get product: %w", err)
	}
	return &p, nil
}

func (r *UserRepositoryPostgresql) Create(product *models.Product) error {
	err := r.db.QueryRow(
		"INSERT INTO products (title, description, price, version) VALUES ($1, $2, $3, $4) RETURNING id",
		product.Title, product.Description, product.Price, product.Version,
	).Scan(&product.ID)
	if err != nil {
		return fmt.Errorf("failed to create product: %w", err)
	}
	return nil
}

func (r *UserRepositoryPostgresql) Update(product *models.Product) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	//сохр тек верс
	_, err = tx.Exec(
		"INSERT INTO product_versions (product_id, title, description, price, version) SELECT id, title, description, price, version FROM products WHERE id = $1",
		product.ID,
	)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to save product version: %w", err)
	}

	//апдейтим товар
	_, err = tx.Exec(
		"UPDATE products SET title = $1, description = $2, price = $3, version = $4 WHERE id = $5",
		product.Title, product.Description, product.Price, product.Version, product.ID,
	)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update product: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

func (r *UserRepositoryPostgresql) Rollback(productID int, versionID int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	//получаем данные
	var version models.ProductVersion
	err = tx.QueryRow(
		"SELECT title, description, price, version FROM product_versions WHERE product_id = $1 AND id = $2",
		productID, versionID,
	).Scan(&version.Title, &version.Description, &version.Price, &version.Version)
	if err != nil {
		tx.Rollback()
		if err == sql.ErrNoRows {
			return fmt.Errorf("version not found")
		}
		return fmt.Errorf("failed to get version: %w", err)
	}

	//обновляем товар
	_, err = tx.Exec(
		"UPDATE products SET title = $1, description = $2, price = $3, version = $4 WHERE id = $5",
		version.Title, version.Description, version.Price, version.Version, productID,
	)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to rollback product: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

func (r *UserRepositoryPostgresql) GetHistory(productID int) ([]models.ProductVersion, error) {
	rows, err := r.db.Query(
		"SELECT id, title, description, price, version FROM product_versions WHERE product_id = $1 ORDER BY id DESC",
		productID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get history: %w", err)
	}
	defer rows.Close()

	var history []models.ProductVersion
	for rows.Next() {
		var v models.ProductVersion
		if err := rows.Scan(&v.ID, &v.Title, &v.Description, &v.Price, &v.Version); err != nil {
			return nil, fmt.Errorf("failed to scan version: %w", err)
		}
		history = append(history, v)
	}
	return history, nil
}
