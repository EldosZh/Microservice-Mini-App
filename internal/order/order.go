package order

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var ErrNotFound = errors.New("order not found")

type Order struct {
	ID         int64     `json:"id"`
	UserID     int64     `json:"user_id"`
	ProductID  int64     `json:"product_id"`
	Quantity   int       `json:"quantity"`
	TotalPrice float64   `json:"total_price"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
}
type Store interface {
	Create(context.Context, int64, int64, int, int64) (Order, error)
	List(context.Context) ([]Order, error)
	Get(context.Context, int64) (Order, error)
	Cancel(context.Context, int64) error
}
type SQLStore struct{ DB *sql.DB }

func scan(row interface{ Scan(...any) error }) (Order, error) {
	var o Order
	var cents int64
	err := row.Scan(&o.ID, &o.UserID, &o.ProductID, &o.Quantity, &cents, &o.Status, &o.CreatedAt)
	o.TotalPrice = float64(cents) / 100
	return o, err
}
func (s SQLStore) Create(ctx context.Context, u, p int64, q int, c int64) (Order, error) {
	return scan(s.DB.QueryRowContext(ctx, "INSERT INTO orders(user_id,product_id,quantity,total_cents,status) VALUES($1,$2,$3,$4,'created') RETURNING id,user_id,product_id,quantity,total_cents,status,created_at", u, p, q, c))
}
func (s SQLStore) List(ctx context.Context) ([]Order, error) {
	rows, e := s.DB.QueryContext(ctx, "SELECT id,user_id,product_id,quantity,total_cents,status,created_at FROM orders ORDER BY id")
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []Order{}
	for rows.Next() {
		o, e := scan(rows)
		if e != nil {
			return nil, e
		}
		out = append(out, o)
	}
	return out, rows.Err()
}
func (s SQLStore) Get(ctx context.Context, id int64) (Order, error) {
	o, e := scan(s.DB.QueryRowContext(ctx, "SELECT id,user_id,product_id,quantity,total_cents,status,created_at FROM orders WHERE id=$1", id))
	if errors.Is(e, sql.ErrNoRows) {
		return o, ErrNotFound
	}
	return o, e
}
func (s SQLStore) Cancel(ctx context.Context, id int64) error {
	res, e := s.DB.ExecContext(ctx, "UPDATE orders SET status='cancelled' WHERE id=$1", id)
	if e != nil {
		return e
	}
	n, e := res.RowsAffected()
	if e == nil && n == 0 {
		return ErrNotFound
	}
	return e
}
