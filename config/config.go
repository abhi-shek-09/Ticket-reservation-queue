package config

import (
	"log"
	"os"
	"github.com/joho/godotenv"
)

type Config struct {
	PostgresURL   string
	RedisAddr     string
	RedisPassword string
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}
	log.Print(os.Getenv("POSTGRES_URL"))
	return &Config{
		PostgresURL:   os.Getenv("POSTGRES_URL"),
		RedisAddr:     os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT"),
		RedisPassword: os.Getenv("REDIS_PASSWORD"),
	}
}