package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4/pgxpool"
)

type Product struct {
	ID          int       `json:"id"`
	Name        string    `json:"name" binding:"required"`
	Description string    `json:"description"`
	Price       float64   `json:"price" binding:"required"`
	Quantity    int       `json:"quantity" binding:"required"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateProduct(product *Product) error {
	query := `
		INSERT INTO products (name, description, price, quantity)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRow(context.Background(), query,
		product.Name,
		product.Description,
		product.Price,
		product.Quantity,
	).Scan(&product.ID, &product.CreatedAt, &product.UpdatedAt)
}

func (r *Repository) GetProduct(id int) (Product, error) {
	var product Product
	query := "SELECT * FROM products WHERE id = $1"
	err := r.db.QueryRow(context.Background(), query, id).Scan(
		&product.ID,
		&product.Name,
		&product.Description,
		&product.Price,
		&product.Quantity,
		&product.CreatedAt,
		&product.UpdatedAt,
	)
	return product, err
}

func (r *Repository) UpdateProduct(id int, product *Product) error {
	query := `
		UPDATE products
		SET name = $1, description = $2, price = $3, quantity = $4, updated_at = CURRENT_TIMESTAMP
		WHERE id = $5
		RETURNING updated_at
	`
	return r.db.QueryRow(context.Background(), query,
		product.Name,
		product.Description,
		product.Price,
		product.Quantity,
		id,
	).Scan(&product.UpdatedAt)
}

func (r *Repository) DeleteProduct(id int) error {
	query := "DELETE FROM products WHERE id = $1"
	_, err := r.db.Exec(context.Background(), query, id)
	return err
}

func (r *Repository) GetAllProducts() ([]Product, error) {
	query := "SELECT * FROM products ORDER BY id"
	rows, err := r.db.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var p Product
		err = rows.Scan(
			&p.ID,
			&p.Name,
			&p.Description,
			&p.Price,
			&p.Quantity,
			&p.CreatedAt,
			&p.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, nil
}

func main() {
	// Configuração do banco de dados
	db, err := pgxpool.Connect(context.Background(), "postgres://postgres:minha_senha@localhost:5432/stock")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	repository := NewRepository(db)

	// Configuração do servidor
	r := gin.Default()

	// Rotas
	r.POST("/products", func(c *gin.Context) {
		var product Product
		if err := c.ShouldBindJSON(&product); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := repository.CreateProduct(&product); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, product)
	})

	r.GET("/products/:id", func(c *gin.Context) {
		id := c.Param("id")
		var productID int
		if _, err := fmt.Sscan(id, &productID); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
			return
		}

		product, err := repository.GetProduct(productID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Produto não encontrado"})
			return
		}

		c.JSON(http.StatusOK, product)
	})

	r.PUT("/products/:id", func(c *gin.Context) {
		id := c.Param("id")
		var productID int
		if _, err := fmt.Sscan(id, &productID); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
			return
		}

		var product Product
		if err := c.ShouldBindJSON(&product); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := repository.UpdateProduct(productID, &product); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, product)
	})

	r.DELETE("/products/:id", func(c *gin.Context) {
		id := c.Param("id")
		var productID int
		if _, err := fmt.Sscan(id, &productID); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
			return
		}

		if err := repository.DeleteProduct(productID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.Status(http.StatusNoContent)
	})

	r.GET("/products", func(c *gin.Context) {
		products, err := repository.GetAllProducts()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, products)
	})

	r.Run(":8080")
}
