package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"techstore/internal/models"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

const (
	defaultRedisTime = 5 * time.Minute
)

var (
	ErrProductNotFound = errors.New("product not found")
	ErrNoProducts      = errors.New("product list is empty")
	ErrBadRequest      = errors.New("incorrect request for this operation")
)

type ProductRepository struct {
	db      *pgxpool.Pool
	redis   *redis.Client
	builder squirrel.StatementBuilderType
}

func NewProductRepository(db *pgxpool.Pool, redisClient *redis.Client) *ProductRepository {
	return &ProductRepository{
		db:      db,
		redis:   redisClient,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (r *ProductRepository) CreateProduct(ctx context.Context, req *models.CreateProductRequest) (*models.Product, error) {
	if req.Name == "" {
		return nil, errors.New("product name cannot be empty")
	}
	if req.Price <= 0 {
		return nil, errors.New("product price must be positive")
	}
	if req.Stock < 0 {
		return nil, errors.New("product stock cannot be negative")
	}

	query, args, err := r.builder.Insert("products").
		Columns("name", "description", "category", "brand", "model", "price", "stock", "image_url").
		Values(req.Name, req.Description, req.Category, req.Brand, req.Model, req.Price, req.Stock, req.ImageURL).
		Suffix("RETURNING id, name, description, category, brand, model, price, stock, image_url, created_at, updated_at").
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("build insert query: %w", err)
	}

	product := &models.Product{}
	err = r.db.QueryRow(ctx, query, args...).Scan(
		&product.ID, &product.Name, &product.Description, &product.Category, &product.Brand,
		&product.Model, &product.Price, &product.Stock, &product.ImageURL, &product.CreatedAt, &product.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("insert product: %w", err)
	}

	r.redis.Del(ctx, "products:all")

	return product, nil
}

func (r *ProductRepository) GetProduct(ctx context.Context, id int) (*models.Product, error) {
	cacheKey := fmt.Sprintf("product:%d", id)

	cachedProduct, err := r.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var product models.Product
		if err := json.Unmarshal([]byte(cachedProduct), &product); err == nil {
			return &product, nil
		}
	}

	query, args, err := r.builder.Select(
		"id", "name", "description", "category", "brand", "model", "price", "stock", "image_url", "created_at", "updated_at").
		From("products").
		Where(squirrel.Eq{"id": id}).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("build select query: %w", err)
	}

	product := &models.Product{}
	err = r.db.QueryRow(ctx, query, args...).Scan(
		&product.ID, &product.Name, &product.Description, &product.Category, &product.Brand,
		&product.Model, &product.Price, &product.Stock, &product.ImageURL, &product.CreatedAt, &product.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrProductNotFound
		}
		return nil, fmt.Errorf("get product: %w", err)
	}

	productJSON, err := json.Marshal(product)
	if err == nil {
		r.redis.Set(ctx, cacheKey, productJSON, defaultRedisTime)
	}

	return product, nil
}

func (r *ProductRepository) GetAllProducts(ctx context.Context, filter models.ProductFilter) ([]*models.Product, error) {
	cacheKey := fmt.Sprintf("products:category:%s:brand:%s:min_price:%.2f:max_price:%.2f:limit:%d:offset:%d",
		filter.Category, filter.Brand, filter.MinPrice, filter.MaxPrice, filter.Limit, filter.Offset)

	cachedProducts, err := r.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var products []*models.Product
		if err := json.Unmarshal([]byte(cachedProducts), &products); err == nil {
			return products, nil
		}
	}

	queryBuilder := r.builder.Select(
		"id", "name", "description", "category", "brand", "model", "price", "stock", "image_url", "created_at", "updated_at").
		From("products")

	if filter.Category != "" {
		queryBuilder = queryBuilder.Where(squirrel.Eq{"category": filter.Category})
	}
	if filter.Brand != "" {
		queryBuilder = queryBuilder.Where(squirrel.Eq{"brand": filter.Brand})
	}
	if filter.MinPrice > 0 {
		queryBuilder = queryBuilder.Where(squirrel.GtOrEq{"price": filter.MinPrice})
	}
	if filter.MaxPrice > 0 {
		queryBuilder = queryBuilder.Where(squirrel.LtOrEq{"price": filter.MaxPrice})
	}

	queryBuilder = queryBuilder.OrderBy("created_at DESC") // последние добавленные

	if filter.Limit > 0 {
		queryBuilder = queryBuilder.Limit(uint64(filter.Limit))
	}
	if filter.Offset > 0 {
		queryBuilder = queryBuilder.Offset(uint64(filter.Offset))
	}

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build select query: %w", err)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("get products: %w", err)
	}
	defer rows.Close()

	var products []*models.Product
	for rows.Next() {
		product := &models.Product{}
		err := rows.Scan(
			&product.ID, &product.Name, &product.Description, &product.Category, &product.Brand,
			&product.Model, &product.Price, &product.Stock, &product.ImageURL, &product.CreatedAt, &product.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan product: %w", err)
		}
		products = append(products, product)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	if len(products) == 0 {
		return nil, ErrNoProducts
	}

	productsJSON, err := json.Marshal(products)
	if err == nil {
		r.redis.Set(ctx, cacheKey, productsJSON, defaultRedisTime)
	}

	return products, nil
}

func (r *ProductRepository) UpdateProduct(ctx context.Context, id int, update *models.UpdateProductRequest) (*models.Product, error) {
	updateBuilder := r.builder.Update("products")

	if update.Name != nil {
		updateBuilder = updateBuilder.Set("name", *update.Name)
	}
	if update.Description != nil {
		updateBuilder = updateBuilder.Set("description", *update.Description)
	}
	if update.Category != nil {
		updateBuilder = updateBuilder.Set("category", *update.Category)
	}
	if update.Brand != nil {
		updateBuilder = updateBuilder.Set("brand", *update.Brand)
	}
	if update.Model != nil {
		updateBuilder = updateBuilder.Set("model", *update.Model)
	}
	if update.Price != nil {
		updateBuilder = updateBuilder.Set("price", *update.Price)
	}
	if update.Stock != nil {
		updateBuilder = updateBuilder.Set("stock", *update.Stock)
	}
	if update.ImageURL != nil {
		updateBuilder = updateBuilder.Set("image_url", *update.ImageURL)
	}

	updateBuilder = updateBuilder.Set("updated_at", squirrel.Expr("NOW()")).
		Where(squirrel.Eq{"id": id}).
		Suffix("RETURNING id, name, description, category, brand, model, price, stock, image_url, created_at, updated_at")

	query, args, err := updateBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build update query: %w", err)
	}

	product := &models.Product{}
	err = r.db.QueryRow(ctx, query, args...).Scan(
		&product.ID, &product.Name, &product.Description, &product.Category, &product.Brand,
		&product.Model, &product.Price, &product.Stock, &product.ImageURL, &product.CreatedAt, &product.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrProductNotFound
		}
		return nil, fmt.Errorf("update product: %w", err)
	}

	cacheKey := fmt.Sprintf("product:%d", id)
	r.redis.Del(ctx, cacheKey, "products:all")

	return product, nil
}

func (r *ProductRepository) DeleteProduct(ctx context.Context, id int) error {
	query, args, err := r.builder.
		Delete("products").
		Where(squirrel.Eq{"id": id}).
		ToSql()

	if err != nil {
		return fmt.Errorf("build delete query: %w", err)
	}

	result, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("delete product: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrProductNotFound
	}

	cacheKey := fmt.Sprintf("product:%d", id)
	r.redis.Del(ctx, cacheKey, "products:all")

	return nil
}
