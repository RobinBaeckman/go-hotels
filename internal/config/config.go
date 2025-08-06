package config

import (
	"log"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv      string `env:"APP_ENV" envDefault:"local"`
	Port        string `env:"PORT" envDefault:"8080"`
	DatabaseURL string `env:"DATABASE_URL,required"`
}

// Production-safe: exits process on error
func Load() Config {
	_ = godotenv.Load(".env", ".env.local")
	cfg, err := parseEnv()
	if err != nil {
		log.Fatalf("Failed to parse env vars: %v", err)
	}
	return cfg
}

// Production-safe: exits process on error
func LoadFrom(paths ...string) Config {
	_ = godotenv.Load(paths...)
	cfg, err := parseEnv()
	if err != nil {
		log.Fatalf("Failed to parse env vars: %v", err)
	}
	return cfg
}

// Test- and CLI-safe: returns error
func TryLoadFrom(paths ...string) (Config, error) {
	_ = godotenv.Load(paths...)
	return parseEnv()
}

// Internal parser
func parseEnv() (Config, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}
