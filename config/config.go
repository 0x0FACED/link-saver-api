package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	GRPC     GRPCConfig
}

type ServerConfig struct {
	Host string
	Port string
}

type RedisConfig struct {
	Host string
	Port string
}

type GRPCConfig struct {
	BaseURL string
	Host    string
	Port    string
}

type DatabaseConfig struct {
	Name     string
	Host     string
	Port     string
	Username string
	Password string
	Driver   string
}

func Load() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, err
	}

	return &Config{
		Database: DatabaseConfig{
			Username: os.Getenv("DB_USER"),
			Password: os.Getenv("DB_PASS"),
			Name:     os.Getenv("DB_NAME"),
			Host:     os.Getenv("DB_HOST"),
			Port:     os.Getenv("DB_PORT"),
			Driver:   os.Getenv("DB_DRIVER"),
		},
		Server: ServerConfig{
			Host: os.Getenv("S_HOST"),
			Port: os.Getenv("S_PORT"),
		},
		Redis: RedisConfig{
			Host: os.Getenv("R_HOST"),
			Port: os.Getenv("R_PORT"),
		},
		GRPC: GRPCConfig{
			BaseURL: os.Getenv("BASE_URL"),
			Host:    os.Getenv("GRPC_HOST"),
			Port:    os.Getenv("GRPC_PORT"),
		},
	}, nil
}
