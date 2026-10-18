// Command server starts the HTTP server and connects to the database.
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
	authService := usecase.NewAuthService(userRepo, codeRepo, mail)
	authHandler := httpapi.NewAuthHandler(authService)

	srv := &http.Server{
		Addr:              ":8080",
		Handler:           httpapi.NewRouter(authHandler),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Println("Listening on port 8080")

	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
