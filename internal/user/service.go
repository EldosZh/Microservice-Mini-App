package user

import (
	"context"
	"database/sql"
	"errors"
	"net/mail"
	"strings"
	"time"
)

var ErrNotFound = errors.New("user not found")
var ErrConflict = errors.New("email already exists")
var ErrInvalid = errors.New("name and a valid email are required")

type User struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}
type Store interface {
	Create(context.Context, string, string) (User, error)
	List(context.Context) ([]User, error)
	Get(context.Context, int64) (User, error)
	Delete(context.Context, int64) error
}
type SQLStore struct{ DB *sql.DB }

func Validate(name, email string) (string, string, error) {
	name = strings.TrimSpace(name)
	email = strings.ToLower(strings.TrimSpace(email))
	parsed, err := mail.ParseAddress(email)
	if name == "" || err != nil || parsed.Address != email {
		return "", "", ErrInvalid
	}
	return name, email, nil
}
func (s SQLStore) Create(ctx context.Context, n, e string) (User, error) {
	var u User
	err := s.DB.QueryRowContext(ctx, "INSERT INTO users(name,email) VALUES($1,$2) RETURNING id,name,email,created_at", n, e).Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt)
	if err != nil && strings.Contains(err.Error(), "users_email_key") {
		return u, ErrConflict
	}
	return u, err
}
func (s SQLStore) List(ctx context.Context) ([]User, error) {
	rows, err := s.DB.QueryContext(ctx, "SELECT id,name,email,created_at FROM users ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []User{}
	for rows.Next() {
		var u User
		if err = rows.Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}
func (s SQLStore) Get(ctx context.Context, id int64) (User, error) {
	var u User
	err := s.DB.QueryRowContext(ctx, "SELECT id,name,email,created_at FROM users WHERE id=$1", id).Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return u, ErrNotFound
	}
	return u, err
}
func (s SQLStore) Delete(ctx context.Context, id int64) error {
	res, err := s.DB.ExecContext(ctx, "DELETE FROM users WHERE id=$1", id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err == nil && n == 0 {
		return ErrNotFound
	}
	return err
}
