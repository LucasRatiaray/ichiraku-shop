// Command server démarre le serveur HTTP et se connecte à la base de données.
package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/LucasRatiaray/ichiraku-shop/internal/adapter/httpapi"
	"github.com/LucasRatiaray/ichiraku-shop/internal/adapter/mailer"
	"github.com/LucasRatiaray/ichiraku-shop/internal/adapter/postgres"
	"github.com/LucasRatiaray/ichiraku-shop/internal/usecase"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("Failed to ping the database: %v", err)
	}

	if err := postgres.Migrate(ctx, db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	log.Println("Connected to the database successfully")

	userRepo := postgres.NewUserRepository(db)
	codeRepo := postgres.NewVerificationCodeRepository(db)
	mail := mailer.NewLogMailer()
	sessionRepo := postgres.NewSessionRepository(db)

	authService := usecase.NewAuthService(userRepo, codeRepo, mail, sessionRepo)

	adminEmail := os.Getenv("ADMIN_EMAIL")
	adminPassword := os.Getenv("ADMIN_PASSWORD")
	if adminEmail != "" && adminPassword != "" {
		if err := authService.EnsureAdmin(ctx, adminEmail, adminPassword); err != nil {
			log.Fatalf("Failed to ensure admin: %v", err)
		}
	} else {
		log.Println("ADMIN_EMAIL/ADMIN_PASSWORD not set: admin seed skipped")
	}

	authHandler := httpapi.NewAuthHandler(authService)
	productRepo := postgres.NewProductRepository(db)
	productService := usecase.NewProductService(productRepo)
	productHandler := httpapi.NewProductHandler(productService)

	basketRepo := postgres.NewBasketRepository(db)
	basketService := usecase.NewBasketService(basketRepo, productRepo)
	basketHandler := httpapi.NewBasketHandler(basketService)

	orderRepo := postgres.NewOrderRepository(db)
	orderService := usecase.NewOrderService(orderRepo, basketRepo, userRepo, productRepo)
	orderHandler := httpapi.NewOrderHandler(orderService)

	userAdminService := usecase.NewUserAdminService(userRepo)
	userAdminHandler := httpapi.NewUserAdminHandler(userAdminService)

	srv := &http.Server{
		Addr:              ":8080",
		Handler:           httpapi.NewRouter(authHandler, productHandler, basketHandler, orderHandler, userAdminHandler),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Println("Listening on port 8080")

	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
