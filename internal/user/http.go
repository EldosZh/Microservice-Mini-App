package user

import (
	"errors"
	"net/http"

	"example.com/go-microservices-order-app/internal/httpx"
)

type Handler struct{ Store Store }

func (h Handler) Routes() http.Handler {
	m := http.NewServeMux()
	m.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) { httpx.JSON(w, 200, map[string]string{"status": "ok"}) })
	m.HandleFunc("POST /users", h.create)
	m.HandleFunc("GET /users", h.list)
	m.HandleFunc("GET /users/{id}", h.get)
	m.HandleFunc("DELETE /users/{id}", h.delete)
	return m
}
func (h Handler) create(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	if httpx.Decode(w, r, &in) != nil {
		httpx.Error(w, 400, "invalid JSON request")
		return
	}
	n, e, err := Validate(in.Name, in.Email)
	if err != nil {
		httpx.Error(w, 400, err.Error())
		return
	}
	u, err := h.Store.Create(r.Context(), n, e)
	if errors.Is(err, ErrConflict) {
		httpx.Error(w, 409, err.Error())
		return
	}
	if err != nil {
		httpx.Error(w, 500, "could not create user")
		return
	}
	httpx.JSON(w, 201, u)
}
func (h Handler) list(w http.ResponseWriter, r *http.Request) {
	v, e := h.Store.List(r.Context())
	if e != nil {
		httpx.Error(w, 500, "could not list users")
		return
	}
	httpx.JSON(w, 200, v)
}
func (h Handler) get(w http.ResponseWriter, r *http.Request) {
	id, ok := httpx.ID(r)
	if !ok {
		httpx.Error(w, 400, "invalid user id")
		return
	}
	v, e := h.Store.Get(r.Context(), id)
	if errors.Is(e, ErrNotFound) {
		httpx.Error(w, 404, "user not found")
		return
	}
	if e != nil {
		httpx.Error(w, 500, "could not get user")
		return
	}
	httpx.JSON(w, 200, v)
}
func (h Handler) delete(w http.ResponseWriter, r *http.Request) {
	id, ok := httpx.ID(r)
	if !ok {
		httpx.Error(w, 400, "invalid user id")
		return
	}
	e := h.Store.Delete(r.Context(), id)
	if errors.Is(e, ErrNotFound) {
		httpx.Error(w, 404, "user not found")
		return
	}
	if e != nil {
		httpx.Error(w, 500, "could not delete user")
		return
	}
	w.WriteHeader(204)
}
