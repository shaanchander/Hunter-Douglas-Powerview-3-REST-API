package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"hunter-douglas-powerview-3-rest-api/internal/api"
)

const defaultAPIPort = "8080"
const usageLine = "usage: powerview-api -H <POWERVIEW_HOST> [-P <API_PORT>]"

func main() {
	cfg, err := loadConfig(os.Args)
	if err != nil {
		log.Fatal(err)
	}

	r := api.NewRouter(cfg)
	if err := r.Run(":" + cfg.APIPort); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

func loadConfig(args []string) (api.Config, error) {
	cfg, err := configFromFlags(args[1:])
	if err != nil {
		return api.Config{}, err
	}

	if cfg.PowerViewHost == "" {
		cfg.PowerViewHost = os.Getenv("POWERVIEW_HOST")
	}
	if cfg.APIPort == "" {
		cfg.APIPort = os.Getenv("API_PORT")
	}

	return normalizeConfig(cfg)
}

func configFromFlags(args []string) (api.Config, error) {
	fs := flag.NewFlagSet("powerview-api", flag.ContinueOnError)
	host := fs.String("H", "", "PowerView hub host URL (required unless POWERVIEW_HOST is set)")
	port := fs.String("P", "", "API port to listen on")
	if err := fs.Parse(args); err != nil {
		return api.Config{}, fmt.Errorf("parse flags: %w; %s", err, usageLine)
	}
	if len(fs.Args()) > 0 {
		return api.Config{}, fmt.Errorf("unexpected args: %v; %s", fs.Args(), usageLine)
	}

	return api.Config{
		PowerViewHost: *host,
		APIPort:       *port,
	}, nil
}

func normalizeConfig(cfg api.Config) (api.Config, error) {
	if cfg.PowerViewHost == "" {
		return api.Config{}, fmt.Errorf("POWERVIEW_HOST is required (use -H or set POWERVIEW_HOST)")
	}
	if cfg.APIPort == "" {
		cfg.APIPort = defaultAPIPort
	}

	return cfg, nil
}
