package service

import (
	"context"
	"techstore/internal/models"
	repository "techstore/product-service/repository"
)

type ProductService interface {
	CreateProduct(ctx context.Context, req *models.CreateProductRequest) (*models.Product, error)
	GetProduct(ctx context.Context, id int) (*models.Product, error)
	ListProducts(ctx context.Context, filter models.ProductFilter) ([]*models.Product, error)
	UpdateProduct(ctx context.Context, id int, req *models.UpdateProductRequest) (*models.Product, error)
	DeleteProduct(ctx context.Context, id int) error
}

type productService struct {
	repo repository.ProductRepository
}

func NewProductService(repo repository.ProductRepository) ProductService {
	return &productService{repo: repo}
}

func (s *productService) CreateProduct(ctx context.Context, req *models.CreateProductRequest) (*models.Product, error) {
	return s.repo.CreateProduct(ctx, req)
}

func (s *productService) GetProduct(ctx context.Context, id int) (*models.Product, error) {
	return s.repo.GetProduct(ctx, id)
}

func (s *productService) ListProducts(ctx context.Context, filter models.ProductFilter) ([]*models.Product, error) {
	if filter.Limit == 0 {
		filter.Limit = 50
	}

	return s.repo.GetAllProducts(ctx, filter)
}

func (s *productService) UpdateProduct(ctx context.Context, id int, req *models.UpdateProductRequest) (*models.Product, error) {
	return s.repo.UpdateProduct(ctx, id, req)
}

func (s *productService) DeleteProduct(ctx context.Context, id int) error {
	return s.repo.DeleteProduct(ctx, id)
}
