package usecase

import (
	"anyprog/internal/models"
	"anyprog/internal/repository"
)

type ProductUseCase struct {
	repo repository.ProductRepository
}

func NewProductUseCase(repo repository.ProductRepository) *ProductUseCase {
	return &ProductUseCase{repo: repo}
}

func (uc *ProductUseCase) GetAll() ([]models.Product, error) {
	return uc.repo.GetAll()
}

func (uc *ProductUseCase) GetByID(id int) (*models.Product, error) {
	return uc.repo.GetByID(id)
}

func (uc *ProductUseCase) Create(product *models.Product) error {
	return uc.repo.Create(product)
}

func (uc *ProductUseCase) Update(product *models.Product) error {
	return uc.repo.Update(product)
}

func (uc *ProductUseCase) Rollback(productID int, versionID int) error {
	return uc.repo.Rollback(productID, versionID)
}

func (uc *ProductUseCase) GetHistory(productID int) ([]models.ProductVersion, error) {
	return uc.repo.GetHistory(productID)
}
