package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

var ErrUnavailable = errors.New("dependency unavailable")
var ErrNotFound = errors.New("dependency resource not found")
var ErrInsufficient = errors.New("insufficient stock")

type UserClient struct {
	BaseURL string
	HTTP    *http.Client
}
type ProductClient struct {
	BaseURL string
	HTTP    *http.Client
}
type Product struct {
	ID    int64   `json:"id"`
	Price float64 `json:"price"`
}

func request(ctx context.Context, c *http.Client, method, url string, body any, out any) (int, error) {
	var rd io.Reader
	if body != nil {
		b, e := json.Marshal(body)
		if e != nil {
			return 0, e
		}
		rd = bytes.NewReader(b)
	}
	req, e := http.NewRequestWithContext(ctx, method, url, rd)
	if e != nil {
		return 0, e
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	res, e := c.Do(req)
	if e != nil {
		return 0, ErrUnavailable
	}
	defer res.Body.Close()
	if res.StatusCode == 404 {
		return 404, ErrNotFound
	}
	if res.StatusCode == 409 {
		return 409, ErrInsufficient
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return res.StatusCode, fmt.Errorf("dependency returned %d", res.StatusCode)
	}
	if out != nil {
		if e = json.NewDecoder(io.LimitReader(res.Body, 1<<20)).Decode(out); e != nil {
			return res.StatusCode, ErrUnavailable
		}
	}
	return res.StatusCode, nil
}
func (c UserClient) Exists(ctx context.Context, id int64) error {
	_, e := request(ctx, c.HTTP, http.MethodGet, strings.TrimRight(c.BaseURL, "/")+fmt.Sprintf("/users/%d", id), nil, &struct {
		ID int64 `json:"id"`
	}{})
	return e
}
func (c ProductClient) Get(ctx context.Context, id int64) (Product, error) {
	var p Product
	_, e := request(ctx, c.HTTP, http.MethodGet, strings.TrimRight(c.BaseURL, "/")+fmt.Sprintf("/products/%d", id), nil, &p)
	return p, e
}
func (c ProductClient) Reserve(ctx context.Context, id int64, q int) error {
	_, e := request(ctx, c.HTTP, http.MethodPost, strings.TrimRight(c.BaseURL, "/")+fmt.Sprintf("/products/%d/reserve", id), map[string]int{"quantity": q}, nil)
	return e
}
