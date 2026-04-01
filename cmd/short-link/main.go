package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/reservation-v/short-link/internal/app"
	"github.com/reservation-v/short-link/internal/config"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Parse(os.Args[1:])
	if err != nil {
		log.Fatalf("parse config: %v", err)
	}

	if err := app.Run(ctx, cfg); err != nil {
		log.Fatalf("run app: %v", err)
	}
}
