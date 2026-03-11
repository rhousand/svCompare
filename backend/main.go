package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/rhousand/svcompare/internal/auth"
	"github.com/rhousand/svcompare/internal/db"
	"github.com/rhousand/svcompare/internal/handlers"
	mw "github.com/rhousand/svcompare/internal/middleware"
)

func main() {
	port := getEnv("PORT", "8080")
	dbPath := getEnv("DATABASE_PATH", "./data/svcompare.db")
	jwtSecret := getEnv("JWT_SECRET", "dev-secret-change-in-prod")
	goEnv := getEnv("GO_ENV", "development")
	adminUsername := getEnv("SEED_ADMIN_USERNAME", "admin")
	adminPassword := getEnv("SEED_ADMIN_PASSWORD", "admin")

	database, err := db.Open(dbPath)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer database.Close()

	hashedPassword, err := auth.HashPassword(adminPassword)
	if err != nil {
		log.Fatalf("hash admin password: %v", err)
	}
	if err := database.SeedAdmin(adminUsername, hashedPassword); err != nil {
		log.Fatalf("seed admin: %v", err)
	}

	authenticator := auth.NewLocalAuthenticator(database, jwtSecret)
	h := handlers.New(database, authenticator, jwtSecret)
	authMW := mw.NewAuthMiddleware(jwtSecret)

	r := chi.NewRouter()
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.StripSlashes)

	if goEnv == "development" {
		r.Use(cors.Handler(cors.Options{
			AllowedOrigins:   []string{"http://localhost:5173"},
			AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Accept", "Content-Type"},
			AllowCredentials: true,
			MaxAge:           300,
		}))
	}

	// Public routes
	r.Post("/api/auth/login", h.Login)
	r.Post("/api/auth/logout", h.Logout)
	r.Get("/api/share/{token}", h.GetShare)

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(authMW)
		r.Get("/api/auth/me", h.Me)

		r.Get("/api/comparisons", h.ListComparisons)
		r.Post("/api/comparisons", h.CreateComparison)
		r.Get("/api/comparisons/{id}", h.GetComparison)
		r.Patch("/api/comparisons/{id}", h.UpdateComparison)
		r.Delete("/api/comparisons/{id}", h.DeleteComparison)

		r.Post("/api/comparisons/{id}/boats", h.AddBoat)
		r.Patch("/api/comparisons/{id}/boats/{bid}", h.UpdateBoat)
		r.Delete("/api/comparisons/{id}/boats/{bid}", h.DeleteBoat)

		r.Put("/api/comparisons/{id}/boats/{bid}/scores", h.UpsertScores)
	})

	// SPA handler — defined in embed_prod.go (production) or embed_dev.go (dev build tag)
	r.Get("/*", spaHandler())

	// Background goroutine: delete expired comparisons every hour.
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			n, err := database.DeleteExpiredComparisons()
			if err != nil {
				log.Printf("cleanup error: %v", err)
			} else if n > 0 {
				log.Printf("deleted %d expired comparisons", n)
			}
		}
	}()

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("svCompare listening on :%s [env=%s]", port, goEnv)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-quit
	log.Println("shutting down…")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("shutdown error: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
