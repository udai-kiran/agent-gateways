package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"
	"time"

	"github.com/udai-kiran/agent-gateways/gateways/internal/core"
	"github.com/udai-kiran/agent-gateways/gateways/internal/whatsapp"
)

func main() {
	agent, err := core.NewHTTPAgentClient()
	if err != nil {
		log.Fatalf("Failed to create agent client: %v", err)
	}

	cfg, err := whatsapp.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	gw := whatsapp.NewGateway(cfg, agent)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := gw.Start(ctx); err != nil {
			log.Fatalf("Gateway error: %v", err)
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := gw.Stop(shutdownCtx); err != nil {
		log.Printf("Shutdown error: %v", err)
	}
}
