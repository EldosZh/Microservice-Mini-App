package main

import (
	"context"
	"example.com/go-microservices-order-app/internal/config"
	"example.com/go-microservices-order-app/internal/database"
	"example.com/go-microservices-order-app/internal/server"
	"example.com/go-microservices-order-app/internal/user"
	usermigration "example.com/go-microservices-order-app/migrations/user"
	"log"
	"time"
)

func main() {
	ctx, c := context.WithTimeout(context.Background(), 30*time.Second)
	defer c()
	db, e := database.Open(ctx, config.Get("DATABASE_URL", "postgres://app:app@localhost:5433/users_db?sslmode=disable"))
	if e != nil {
		log.Fatal(e)
	}
	defer db.Close()
	if e = usermigration.Apply(db); e != nil {
		log.Fatal("database initialization failed")
	}
	server.Run(":"+config.Get("PORT", "8081"), user.Handler{Store: user.SQLStore{DB: db}}.Routes())
}
