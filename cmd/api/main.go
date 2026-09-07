package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/driver-hub/driver-hub/internal/auth"
	"github.com/driver-hub/driver-hub/internal/config"
	"github.com/driver-hub/driver-hub/internal/esia"
	httpapi "github.com/driver-hub/driver-hub/internal/handler/http"
	"github.com/driver-hub/driver-hub/internal/repository/postgres"
	"github.com/driver-hub/driver-hub/internal/seed"
	"github.com/driver-hub/driver-hub/internal/service"
	"github.com/driver-hub/driver-hub/internal/vision"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	pool, err := postgres.WaitReady(ctx, cfg.DatabaseURL, 30)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer pool.Close()

	migPath := os.Getenv("MIGRATIONS_PATH")
	if migPath == "" {
		migPath = "migrations/001_init.sql"
	}
	if err := postgres.Migrate(ctx, pool, migPath); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	store := postgres.NewStore(pool)
	if os.Getenv("SEED_DEMO") != "false" {
		if err := seed.Run(ctx, store); err != nil {
			log.Printf("seed warning: %v", err)
		}
	}
	tokens := auth.NewTokenManager(cfg.JWTSecret, cfg.JWTExpiry)

	var esiaClient esia.Client
	var mock *esia.MockClient
	switch cfg.ESIAMode {
	case "live":
		live, err := esia.NewLiveClient(cfg.ESIABaseURL, cfg.ESIAClientID, cfg.ESIAClientSecret, cfg.ESIARedirectURI, cfg.ESIAScopes, cfg.ESIAPrivateKeyPath)
		if err != nil {
			log.Fatalf("esia live: %v", err)
		}
		esiaClient = live
	default:
		mock = esia.NewMockClient(cfg.PublicBaseURL)
		esiaClient = mock
	}

	var visionScanner vision.Scanner
	switch cfg.VisionMode {
	case "openai":
		visionScanner = vision.NewOpenAIScanner(cfg.OpenAIAPIKey, cfg.OpenAIBaseURL, cfg.OpenAIVisionModel)
	default:
		visionScanner = vision.NewMockScanner()
	}

	svc := service.New(store, tokens, esiaClient, visionScanner)
	h := httpapi.NewHandler(svc, mock)

	_ = os.MkdirAll(cfg.UploadDir, 0o755)

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           httpapi.NewRouter(h, tokens),
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("Driver Hub listening on %s (esia=%s vision=%s)", cfg.HTTPAddr, cfg.ESIAMode, cfg.VisionMode)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	<-ctx.Done()
	shutdownCtx, c := context.WithTimeout(context.Background(), 10*time.Second)
	defer c()
	_ = srv.Shutdown(shutdownCtx)
}
