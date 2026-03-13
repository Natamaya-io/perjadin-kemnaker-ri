package config

import (
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	App      AppConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	Fonnte   FonnteConfig
}

type AppConfig struct {
	Port string
	Env  string
}

type FonnteConfig struct {
	Token string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
}

type JWTConfig struct {
	Secret string
	Expiry int // Hours
}

func LoadConfig() *Config {
	// Defaults
	viper.SetDefault("APP_PORT", "8080")
	viper.SetDefault("APP_ENV", "development")
	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", "5432")
	viper.SetDefault("DB_SSLMODE", "disable")
	viper.SetDefault("REDIS_HOST", "localhost")
	viper.SetDefault("REDIS_PORT", "6379")
	viper.SetDefault("REDIS_PASSWORD", "")
	viper.SetDefault("JWT_EXPIRY", 24)

	viper.AutomaticEnv()

	var cfg Config
	cfg.App.Port = viper.GetString("APP_PORT")
	cfg.App.Env = viper.GetString("APP_ENV")

	dbHost := viper.GetString("DB_HOST")
	if dbHost == "localhost" || dbHost == "127.0.0.1" {
		if os.Getenv("container") == "podman" {
			dbHost = "host.containers.internal"
		} else if _, err := os.Stat("/.dockerenv"); err == nil {
			dbHost = "host.docker.internal"
		}
	}
	cfg.Database.Host = dbHost
	cfg.Database.Port = viper.GetString("DB_PORT")
	cfg.Database.User = viper.GetString("DB_USER")
	cfg.Database.Password = viper.GetString("DB_PASSWORD")
	cfg.Database.Name = viper.GetString("DB_NAME")
	cfg.Database.SSLMode = viper.GetString("DB_SSLMODE")

	cfg.Redis.Host = viper.GetString("REDIS_HOST")
	cfg.Redis.Port = viper.GetString("REDIS_PORT")
	cfg.Redis.Password = viper.GetString("REDIS_PASSWORD")

	cfg.JWT.Secret = viper.GetString("JWT_SECRET")
	cfg.JWT.Expiry = viper.GetInt("JWT_EXPIRY")

	cfg.Fonnte.Token = viper.GetString("FONNTE_API")

	return &cfg
}
