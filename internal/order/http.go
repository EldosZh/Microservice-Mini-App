package order

import (
	"context"
	"errors"
	"math"
	"net/http"

	"example.com/go-microservices-order-app/internal/client"
	"example.com/go-microservices-order-app/internal/httpx"
)

type Users interface {
	Exists(context.Context, int64) error
}
type Products interface {
	Get(context.Context, int64) (client.Product, error)
	Reserve(context.Context, int64, int) error
}
type Handler struct {
	Store    Store
	Users    Users
	Products Products
}

func (h Handler) Routes() http.Handler {
	m := http.NewServeMux()
	m.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) { httpx.JSON(w, 200, map[string]string{"status": "ok"}) })
	m.HandleFunc("POST /orders", h.create)
	m.HandleFunc("GET /orders", h.list)
	m.HandleFunc("GET /orders/{id}", h.get)
	m.HandleFunc("DELETE /orders/{id}", h.cancel)
	return m
}
func dependencyError(w http.ResponseWriter, err error, kind string) {
	if errors.Is(err, client.ErrNotFound) {
		httpx.Error(w, 404, kind+" not found")
		return
	}
	if errors.Is(err, client.ErrInsufficient) {
		httpx.Error(w, 409, "insufficient stock")
		return
	}
	httpx.Error(w, 503, kind+" service unavailable")
}
func (h Handler) create(w http.ResponseWriter, r *http.Request) {
	var in struct {
		UserID    int64 `json:"user_id"`
		ProductID int64 `json:"product_id"`
		Quantity  int   `json:"quantity"`
	}
	if httpx.Decode(w, r, &in) != nil {
		httpx.Error(w, 400, "invalid JSON request")
		return
	}
	if in.UserID < 1 || in.ProductID < 1 || in.Quantity < 1 {
		httpx.Error(w, 400, "user_id, product_id, and positive quantity are required")
		return
	}
	if err := h.Users.Exists(r.Context(), in.UserID); err != nil {
		dependencyError(w, err, "user")
		return
	}
	p, err := h.Products.Get(r.Context(), in.ProductID)
	if err != nil {
		dependencyError(w, err, "product")
		return
	}
	cents := int64(math.Round(p.Price * 100))
	if cents < 1 || int64(in.Quantity) > math.MaxInt64/cents {
		httpx.Error(w, 400, "order total is invalid")
		return
	}
	if err = h.Products.Reserve(r.Context(), in.ProductID, in.Quantity); err != nil {
		dependencyError(w, err, "product")
		return
	}
	o, err := h.Store.Create(r.Context(), in.UserID, in.ProductID, in.Quantity, cents*int64(in.Quantity))
	if err != nil {
		httpx.Error(w, 500, "could not create order")
		return
	}
	httpx.JSON(w, 201, o)
}
func (h Handler) list(w http.ResponseWriter, r *http.Request) {
	v, e := h.Store.List(r.Context())
	if e != nil {
		httpx.Error(w, 500, "could not list orders")
		return
	}
	httpx.JSON(w, 200, v)
}
func (h Handler) get(w http.ResponseWriter, r *http.Request) {
	id, ok := httpx.ID(r)
	if !ok {
		httpx.Error(w, 400, "invalid order id")
		return
	}
	v, e := h.Store.Get(r.Context(), id)
	if errors.Is(e, ErrNotFound) {
		httpx.Error(w, 404, "order not found")
		return
	}
	if e != nil {
		httpx.Error(w, 500, "could not get order")
		return
	}
	httpx.JSON(w, 200, v)
}
func (h Handler) cancel(w http.ResponseWriter, r *http.Request) {
	id, ok := httpx.ID(r)
	if !ok {
		httpx.Error(w, 400, "invalid order id")
		return
	}
	e := h.Store.Cancel(r.Context(), id)
	if errors.Is(e, ErrNotFound) {
		httpx.Error(w, 404, "order not found")
		return
	}
	if e != nil {
		httpx.Error(w, 500, "could not cancel order")
		return
	}
	httpx.JSON(w, 200, map[string]string{"status": "cancelled"})
}
