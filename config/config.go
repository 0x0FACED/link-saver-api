package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   serverConfig
	Database databaseConfig
}

type serverConfig struct {
	Host string
	Port string
}

type databaseConfig struct {
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
		Database: databaseConfig{
			Username: os.Getenv("DB_USER"),
			Password: os.Getenv("DB_PASS"),
			Name:     os.Getenv("DB_NAME"),
			Host:     os.Getenv("DB_HOST"),
			Port:     os.Getenv("DB_PORT"),
			Driver:   os.Getenv("DB_DRIVER"),
		},
		Server: serverConfig{
			Host: os.Getenv("S_HOST"),
			Port: os.Getenv("S_PORT"),
		},
	}, nil
}
