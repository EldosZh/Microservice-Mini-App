package order

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"example.com/go-microservices-order-app/internal/client"
)

type fakeStore struct {
	created bool
	total   int64
}

func (f *fakeStore) Create(_ context.Context, u, p int64, q int, c int64) (Order, error) {
	f.created = true
	f.total = c
	return Order{ID: 1, UserID: u, ProductID: p, Quantity: q, TotalPrice: float64(c) / 100, Status: "created", CreatedAt: time.Now()}, nil
}
func (f *fakeStore) List(context.Context) ([]Order, error)     { return nil, nil }
func (f *fakeStore) Get(context.Context, int64) (Order, error) { return Order{}, ErrNotFound }
func (f *fakeStore) Cancel(context.Context, int64) error       { return nil }

type fakeUsers struct{ err error }

func (f fakeUsers) Exists(context.Context, int64) error { return f.err }

type fakeProducts struct {
	p                  client.Product
	getErr, reserveErr error
	reserved           int
}

func (f *fakeProducts) Get(context.Context, int64) (client.Product, error) { return f.p, f.getErr }
func (f *fakeProducts) Reserve(_ context.Context, _ int64, n int) error {
	f.reserved = n
	return f.reserveErr
}
func perform(h Handler, body string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(body))
	h.Routes().ServeHTTP(w, r)
	return w
}
func TestCreateOrderTotalAndReserve(t *testing.T) {
	s := &fakeStore{}
	p := &fakeProducts{p: client.Product{ID: 1, Price: 79.99}}
	h := Handler{Store: s, Users: fakeUsers{}, Products: p}
	w := perform(h, `{"user_id":1,"product_id":1,"quantity":2}`)
	if w.Code != 201 || s.total != 15998 || p.reserved != 2 || !s.created {
		t.Fatalf("status=%d total=%d reserved=%d body=%s", w.Code, s.total, p.reserved, w.Body.String())
	}
}
func TestCreateRejectsInvalidQuantity(t *testing.T) {
	s := &fakeStore{}
	w := perform(Handler{Store: s, Users: fakeUsers{}, Products: &fakeProducts{}}, `{"user_id":1,"product_id":1,"quantity":0}`)
	if w.Code != 400 || s.created {
		t.Fatalf("status=%d created=%v", w.Code, s.created)
	}
}
func TestCreateMapsMissingUserAndUnavailableDependency(t *testing.T) {
	for _, tc := range []struct {
		err    error
		status int
	}{{client.ErrNotFound, 404}, {errors.New("offline"), 503}} {
		w := perform(Handler{Store: &fakeStore{}, Users: fakeUsers{tc.err}, Products: &fakeProducts{}}, `{"user_id":1,"product_id":1,"quantity":1}`)
		if w.Code != tc.status {
			t.Fatalf("got %d want %d", w.Code, tc.status)
		}
	}
}
func TestCreateMapsMissingProductAndInsufficientStock(t *testing.T) {
	for _, tc := range []struct {
		get, reserve error
		status       int
	}{{get: client.ErrNotFound, status: 404}, {reserve: client.ErrInsufficient, status: 409}} {
		w := perform(Handler{Store: &fakeStore{}, Users: fakeUsers{}, Products: &fakeProducts{p: client.Product{Price: 1}, getErr: tc.get, reserveErr: tc.reserve}}, `{"user_id":1,"product_id":1,"quantity":1}`)
		if w.Code != tc.status {
			t.Fatalf("got %d want %d", w.Code, tc.status)
		}
	}
}
