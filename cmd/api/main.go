package main

import (
	"log"
	"os"

	"hunter-douglas-powerview-3-rest-api/internal/api"
)

func main() {
	cfg := api.Config{
		PowerViewHost: os.Getenv("POWERVIEW_HOST"),
		APIPort:       os.Getenv("API_PORT"),
	}

	if cfg.PowerViewHost == "" {
		log.Fatal("POWERVIEW_HOST is required")
	}
	if cfg.APIPort == "" {
		cfg.APIPort = "8080"
	}

	r := api.NewRouter(cfg)
	if err := r.Run(":" + cfg.APIPort); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
