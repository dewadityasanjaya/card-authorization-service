package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	App      AppConfig
	Database DatabaseConfig
}

type AppConfig struct {
	Port string
	Env  string
}

type DatabaseConfig struct {
	Host     string
	User     string
	Password string
	Name     string
	Port     string
	SSLMode  string
}

func Load() *Config {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv() // also read from real environment variables

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	return &Config{
		App: AppConfig{
			Port: viper.GetString("APP_PORT"),
			Env:  viper.GetString("APP_ENV"),
		},
		Database: DatabaseConfig{
			Host:     viper.GetString("DB_HOST"),
			User:     viper.GetString("DB_USER"),
			Password: viper.GetString("DB_PASSWORD"),
			Name:     viper.GetString("DB_NAME"),
			Port:     viper.GetString("DB_PORT"),
			SSLMode:  viper.GetString("DB_SSLMODE"),
		},
	}
}

// DSN builds the Postgres connection string
func (d *DatabaseConfig) DSN() string {
	return "host=" + d.Host +
		" user=" + d.User +
		" password=" + d.Password +
		" dbname=" + d.Name +
		" port=" + d.Port +
		" sslmode=" + d.SSLMode
}
