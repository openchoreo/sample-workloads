package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	collab "service-go-collab/internal"
)

func main() {
	cfg := collab.LoadConfig()

	hub := collab.NewHub()
	broker, err := collab.NewNatsBroker(cfg.NATSURL)
	if err != nil {
		log.Fatalf("nats connection failed: %v", err)
	}
	defer broker.Close()

	server := collab.NewServer(hub, broker)
	if err := server.SubscribeNATS(); err != nil {
		log.Fatalf("nats subscribe failed: %v", err)
	}

	httpServer := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           server.Router(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	log.Println("starting collab service...")

	go func() {
		log.Printf("collab service listening on %s", cfg.HTTPAddr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("collab service failed: %v", err)
		}
	}()

	<-ctx.Done()
	shutdownCtx, shutdownCancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer shutdownCancel()
	_ = httpServer.Shutdown(shutdownCtx)
}
