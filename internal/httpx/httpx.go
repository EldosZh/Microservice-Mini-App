package httpx

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
)

var ErrInvalidJSON = errors.New("invalid JSON request")

func Decode(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	defer r.Body.Close()
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(dst); err != nil {
		return ErrInvalidJSON
	}
	var extra any
	if err := d.Decode(&extra); err != io.EOF {
		return ErrInvalidJSON
	}
	return nil
}

func JSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func Error(w http.ResponseWriter, status int, message string) {
	JSON(w, status, map[string]string{"error": message})
}

func ID(r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	return id, err == nil && id > 0
}
