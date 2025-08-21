package main

import (
	"flag"
	"go-tunnel/internal/server"
	"go-tunnel/pkg/config"
	"go-tunnel/pkg/logger"
	"log"
)

func main() {
	configPath := flag.String("config", "tunnel-server-config.yaml", "path to config file")
	flag.Parse()

	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	logger.Init(cfg.Log.Path, cfg.Log.Enable)

	logger.InfoLogger.Println("Starting server...")
	srv := server.NewServer(cfg)
	if err := srv.Start(); err != nil {
		logger.ErrorLogger.Fatalf("Failed to start server: %v", err)
	}
}
