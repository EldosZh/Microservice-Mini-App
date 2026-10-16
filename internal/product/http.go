package product

import (
	"errors"
	"net/http"

	"example.com/go-microservices-order-app/internal/httpx"
)

type Handler struct{ Store Store }

func (h Handler) Routes() http.Handler {
	m := http.NewServeMux()
	m.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) { httpx.JSON(w, 200, map[string]string{"status": "ok"}) })
	m.HandleFunc("POST /products", h.create)
	m.HandleFunc("GET /products", h.list)
	m.HandleFunc("GET /products/{id}", h.get)
	m.HandleFunc("PATCH /products/{id}/stock", h.stock)
	m.HandleFunc("POST /products/{id}/reserve", h.reserve)
	m.HandleFunc("DELETE /products/{id}", h.delete)
	return m
}
func (h Handler) create(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Name  string  `json:"name"`
		Price float64 `json:"price"`
		Stock int     `json:"stock"`
	}
	if httpx.Decode(w, r, &in) != nil {
		httpx.Error(w, 400, "invalid JSON request")
		return
	}
	c, e := Cents(in.Price)
	if e != nil {
		httpx.Error(w, 400, ErrInvalid.Error())
		return
	}
	p, e := h.Store.Create(r.Context(), in.Name, c, in.Stock)
	if e != nil {
		httpx.Error(w, 400, e.Error())
		return
	}
	httpx.JSON(w, 201, p)
}
func (h Handler) list(w http.ResponseWriter, r *http.Request) {
	v, e := h.Store.List(r.Context())
	if e != nil {
		httpx.Error(w, 500, "could not list products")
		return
	}
	httpx.JSON(w, 200, v)
}
func (h Handler) get(w http.ResponseWriter, r *http.Request) {
	id, ok := httpx.ID(r)
	if !ok {
		httpx.Error(w, 400, "invalid product id")
		return
	}
	v, e := h.Store.Get(r.Context(), id)
	if errors.Is(e, ErrNotFound) {
		httpx.Error(w, 404, "product not found")
		return
	}
	if e != nil {
		httpx.Error(w, 500, "could not get product")
		return
	}
	httpx.JSON(w, 200, v)
}
func (h Handler) stock(w http.ResponseWriter, r *http.Request) {
	id, ok := httpx.ID(r)
	if !ok {
		httpx.Error(w, 400, "invalid product id")
		return
	}
	var in struct {
		Stock int `json:"stock"`
	}
	if httpx.Decode(w, r, &in) != nil {
		httpx.Error(w, 400, "invalid JSON request")
		return
	}
	e := h.Store.Stock(r.Context(), id, in.Stock)
	if errors.Is(e, ErrNotFound) {
		httpx.Error(w, 404, "product not found")
		return
	}
	if e != nil {
		httpx.Error(w, 400, e.Error())
		return
	}
	httpx.JSON(w, 200, map[string]int{"stock": in.Stock})
}
func (h Handler) reserve(w http.ResponseWriter, r *http.Request) {
	id, ok := httpx.ID(r)
	if !ok {
		httpx.Error(w, 400, "invalid product id")
		return
	}
	var in struct {
		Quantity int `json:"quantity"`
	}
	if httpx.Decode(w, r, &in) != nil {
		httpx.Error(w, 400, "invalid JSON request")
		return
	}
	e := h.Store.Reserve(r.Context(), id, in.Quantity)
	switch {
	case errors.Is(e, ErrNotFound):
		httpx.Error(w, 404, "product not found")
	case errors.Is(e, ErrInsufficient):
		httpx.Error(w, 409, ErrInsufficient.Error())
	case e != nil:
		httpx.Error(w, 400, e.Error())
	default:
		httpx.JSON(w, 200, map[string]string{"status": "reserved"})
	}
}
func (h Handler) delete(w http.ResponseWriter, r *http.Request) {
	id, ok := httpx.ID(r)
	if !ok {
		httpx.Error(w, 400, "invalid product id")
		return
	}
	e := h.Store.Delete(r.Context(), id)
	if errors.Is(e, ErrNotFound) {
		httpx.Error(w, 404, "product not found")
		return
	}
	if e != nil {
		httpx.Error(w, 500, "could not delete product")
		return
	}
	w.WriteHeader(204)
}
