package server

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func Run(addr string, h http.Handler) {
	s := &http.Server{Addr: addr, Handler: h, ReadTimeout: 5 * time.Second, WriteTimeout: 10 * time.Second, IdleTimeout: 60 * time.Second}
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-done
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = s.Shutdown(ctx)
	}()
	log.Printf("listening on %s", addr)
	if e := s.ListenAndServe(); e != nil && !errors.Is(e, http.ErrServerClosed) {
		log.Fatal(e)
	}
}
