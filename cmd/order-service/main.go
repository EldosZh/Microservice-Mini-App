package main

import (
	"context"
	"example.com/go-microservices-order-app/internal/client"
	"example.com/go-microservices-order-app/internal/config"
	"example.com/go-microservices-order-app/internal/database"
	"example.com/go-microservices-order-app/internal/order"
	"example.com/go-microservices-order-app/internal/server"
	ordermigration "example.com/go-microservices-order-app/migrations/order"
	"log"
	"net/http"
	"time"
)

func main() {
	ctx, c := context.WithTimeout(context.Background(), 30*time.Second)
	defer c()
	db, e := database.Open(ctx, config.Get("DATABASE_URL", "postgres://app:app@localhost:5435/orders_db?sslmode=disable"))
	if e != nil {
		log.Fatal(e)
	}
	defer db.Close()
	if e = ordermigration.Apply(db); e != nil {
		log.Fatal("database initialization failed")
	}
	hc := &http.Client{Timeout: 5 * time.Second}
	h := order.Handler{Store: order.SQLStore{DB: db}, Users: client.UserClient{BaseURL: config.Get("USER_SERVICE_URL", "http://localhost:8081"), HTTP: hc}, Products: client.ProductClient{BaseURL: config.Get("PRODUCT_SERVICE_URL", "http://localhost:8082"), HTTP: hc}}
	server.Run(":"+config.Get("PORT", "8083"), h.Routes())
}
