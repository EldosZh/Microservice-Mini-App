package product

import (
	"context"
	"database/sql"
	"errors"
	"math"
	"strings"
	"time"
)

var ErrNotFound = errors.New("product not found")
var ErrInvalid = errors.New("name, positive price, and non-negative stock are required")
var ErrInsufficient = errors.New("insufficient stock")

type Product struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Price     float64   `json:"price"`
	Stock     int       `json:"stock"`
	CreatedAt time.Time `json:"created_at"`
}
type Store interface {
	Create(context.Context, string, int64, int) (Product, error)
	List(context.Context) ([]Product, error)
	Get(context.Context, int64) (Product, error)
	Stock(context.Context, int64, int) error
	Reserve(context.Context, int64, int) error
	Delete(context.Context, int64) error
}
type SQLStore struct{ DB *sql.DB }

func Cents(price float64) (int64, error) {
	if price <= 0 || math.IsNaN(price) || math.IsInf(price, 0) {
		return 0, ErrInvalid
	}
	cents := math.Round(price * 100)
	if math.Abs(price*100-cents) > 1e-7 || cents > math.MaxInt64 {
		return 0, ErrInvalid
	}
	return int64(cents), nil
}
func (s SQLStore) Create(ctx context.Context, n string, c int64, stock int) (Product, error) {
	n = strings.TrimSpace(n)
	if n == "" || c <= 0 || stock < 0 {
		return Product{}, ErrInvalid
	}
	var p Product
	var cents int64
	err := s.DB.QueryRowContext(ctx, "INSERT INTO products(name,price_cents,stock) VALUES($1,$2,$3) RETURNING id,name,price_cents,stock,created_at", n, c, stock).Scan(&p.ID, &p.Name, &cents, &p.Stock, &p.CreatedAt)
	p.Price = float64(cents) / 100
	return p, err
}
func (s SQLStore) List(ctx context.Context) ([]Product, error) {
	rows, err := s.DB.QueryContext(ctx, "SELECT id,name,price_cents,stock,created_at FROM products ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Product{}
	for rows.Next() {
		var p Product
		var c int64
		if err = rows.Scan(&p.ID, &p.Name, &c, &p.Stock, &p.CreatedAt); err != nil {
			return nil, err
		}
		p.Price = float64(c) / 100
		out = append(out, p)
	}
	return out, rows.Err()
}
func (s SQLStore) Get(ctx context.Context, id int64) (Product, error) {
	var p Product
	var c int64
	err := s.DB.QueryRowContext(ctx, "SELECT id,name,price_cents,stock,created_at FROM products WHERE id=$1", id).Scan(&p.ID, &p.Name, &c, &p.Stock, &p.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return p, ErrNotFound
	}
	p.Price = float64(c) / 100
	return p, err
}
func (s SQLStore) Stock(ctx context.Context, id int64, n int) error {
	if n < 0 {
		return ErrInvalid
	}
	res, err := s.DB.ExecContext(ctx, "UPDATE products SET stock=$1 WHERE id=$2", n, id)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}
func (s SQLStore) Reserve(ctx context.Context, id int64, n int) error {
	if n < 1 {
		return ErrInvalid
	}
	res, err := s.DB.ExecContext(ctx, "UPDATE products SET stock=stock-$1 WHERE id=$2 AND stock >= $1", n, id)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows > 0 {
		return nil
	}
	var exists bool
	err = s.DB.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM products WHERE id=$1)", id).Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		return ErrNotFound
	}
	return ErrInsufficient
}
func (s SQLStore) Delete(ctx context.Context, id int64) error {
	res, err := s.DB.ExecContext(ctx, "DELETE FROM products WHERE id=$1", id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err == nil && n == 0 {
		return ErrNotFound
	}
	return err
}
