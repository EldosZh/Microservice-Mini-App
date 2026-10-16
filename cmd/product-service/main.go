package main

import (
	"context"
	"example.com/go-microservices-order-app/internal/config"
	"example.com/go-microservices-order-app/internal/database"
	"example.com/go-microservices-order-app/internal/product"
	"example.com/go-microservices-order-app/internal/server"
	productmigration "example.com/go-microservices-order-app/migrations/product"
	"log"
	"time"
)

func main() {
	ctx, c := context.WithTimeout(context.Background(), 30*time.Second)
	defer c()
	db, e := database.Open(ctx, config.Get("DATABASE_URL", "postgres://app:app@localhost:5434/products_db?sslmode=disable"))
	if e != nil {
		log.Fatal(e)
	}
	defer db.Close()
	if e = productmigration.Apply(db); e != nil {
		log.Fatal("database initialization failed")
	}
	server.Run(":"+config.Get("PORT", "8082"), product.Handler{Store: product.SQLStore{DB: db}}.Routes())
}
