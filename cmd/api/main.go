package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	"hunter-douglas-powerview-3-rest-api/internal/api"

	"gopkg.in/yaml.v3"
)

const configFilePath = "config.yaml"

func main() {
	cfg, err := loadConfig(configFilePath)
	if err != nil {
		log.Fatal(err)
	}

	r := api.NewRouter(cfg)
	if err := r.Run(":" + cfg.APIPort); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

func loadConfig(path string) (api.Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return api.Config{}, fmt.Errorf("missing %s: copy config.sample.yaml to %s and fill in your values", path, path)
		}
		return api.Config{}, fmt.Errorf("read config file %q: %w", path, err)
	}

	var cfg api.Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return api.Config{}, fmt.Errorf("parse config file %q: %w", path, err)
	}

	if cfg.PowerViewHost == "" {
		return api.Config{}, fmt.Errorf("%s: POWERVIEW_HOST is required", path)
	}
	if cfg.APIPort == "" {
		cfg.APIPort = "8080"
	}

	return cfg, nil
}
