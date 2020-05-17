package config

import (
	"os"
)

type Config struct {
	GRPC     GRPC
	HTTP     HTTP
	Database Database
}

type Database struct {
	Host     string
	Database string
	User     string
	Password string
	Port     string
}

type GRPC struct {
	ListenAddress string
	Port          string
}

type HTTP struct {
	ListenAddress string
	Port          string
}

func NewFromEnv() *Config {
	return &Config{
		GRPC: GRPC{
			ListenAddress: os.Getenv("COMMONS_GRPC_ADDRESS"),
			Port:          os.Getenv("COMMONS_GRPC_PORT"),
		},
		HTTP: HTTP{
			ListenAddress: os.Getenv("COMMONS_HTTP_ADDRESS"),
			Port:          os.Getenv("COMMONS_HTTP_PORT"),
		},
		Database: Database{
			Host:     os.Getenv("COMMONS_DB_HOST"),
			Database: os.Getenv("COMMONS_DB_DATABASE"),
			User:     os.Getenv("COMMONS_DB_USER"),
			Password: os.Getenv("COMMONS_DB_PASSWORD"),
			Port:     os.Getenv("COMMONS_DB_PORT"),
		},
	}
}
