package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	document "service-go-document/internal"
)

func main() {
	cfg := document.LoadConfig()
	if cfg.DatabaseURL == "" {
		log.Fatal("DOCLET_DATABASE_URL is required")
	}

	db, err := document.OpenDatabase(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	if err := document.RunMigrations(db); err != nil {
		log.Fatalf("database migration failed: %v", err)
	}

	store := document.NewStore(db)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	nc, err := document.StartSnapshotConsumer(ctx, store, cfg.NATSURL)
	if err != nil {
		log.Fatalf("nats connection failed: %v", err)
	}
	defer nc.Close()

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           document.NewServer(store).Router(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("starting document service.....")

	go func() {
		log.Printf("document service listening on %s", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("document service failed: %v", err)
		}
	}()

	<-ctx.Done()
	shutdownCtx, shutdownCancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer shutdownCancel()
	_ = srv.Shutdown(shutdownCtx)
}
