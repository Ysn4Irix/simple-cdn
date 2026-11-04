package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"simple-cdn/internal/api"
	"simple-cdn/internal/config"
	"simple-cdn/internal/delivery"
	"simple-cdn/internal/security"
	"simple-cdn/internal/storage"
	"simple-cdn/internal/web"
)

func main() {
	configPath := os.Getenv("CONFIG_PATH")
	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("Fatal: Failed to load configuration: %v", err)
	}

	storageEng, err := storage.New(cfg.StorageDir)
	if err != nil {
		log.Fatalf("Fatal: Failed to initialize storage engine: %v", err)
	}

	signer := security.NewURLSigner(cfg.SecretKey)
	limiter := security.NewRateLimiter(cfg.RateLimitRPS, cfg.RateLimitBurst)
	delHandler := delivery.NewHandler(cfg, storageEng)
	apiServer := api.New(cfg, storageEng, delHandler, signer)

	// API Mux with Auth Guard
	apiMux := http.NewServeMux()
	apiServer.RegisterRoutes(apiMux)
	apiGuarded := security.AuthGuard(cfg.APIKey, apiMux)

	// Delivery Handler with Optional Signed URL Guard
	deliveryGuarded := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("token")
		if cfg.RequireSignedURLs || token != "" {
			relPath := strings.TrimPrefix(r.URL.Path, "/files/")
			if err := signer.Verify(relPath, r.URL.Query()); err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				fmt.Fprintf(w, `{"error":"Forbidden: %s"}`, err.Error())
				return
			}
		}
		delHandler.ServeHTTP(w, r)
	})

	deliveryPipeline := security.ReferrerMiddleware(
		cfg.AllowedReferrers,
		delivery.CompressionMiddleware(cfg.EnableGzip, deliveryGuarded),
	)

	// Main Root Router
	mainMux := http.NewServeMux()

	// 1. Files delivery route
	mainMux.Handle("/files/", deliveryPipeline)

	// 2. Protected API routes
	mainMux.Handle("/api/", apiGuarded)

	// 3. Web UI Dashboard
	mainMux.Handle("/", web.Handler())

	// Top-level global pipeline: Rate Limiter + CORS
	handler := limiter.Middleware(security.CORSMiddleware(cfg.AllowedOrigins, mainMux))

	serverAddr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	srv := &http.Server{
		Addr:         serverAddr,
		Handler:      handler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		log.Printf("Simple Static CDN listening on http://%s", serverAddr)
		log.Printf("Storage directory: %s", cfg.StorageDir)
		log.Printf("Public Base URL: %s", cfg.BaseURL)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)
	<-stopChan

	log.Println("Shutting down Simple Static CDN server gracefully...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	log.Println("Server exiting.")
}
