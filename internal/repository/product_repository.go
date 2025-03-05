package repository

import "anyprog/internal/models"

type ProductRepository interface {
	GetAll() ([]models.Product, error)
	GetByID(id int) (*models.Product, error)
	Create(product *models.Product) error
	Update(product *models.Product) error
	Rollback(productID int, versionID int) error
	GetHistory(productID int) ([]models.ProductVersion, error)
}
