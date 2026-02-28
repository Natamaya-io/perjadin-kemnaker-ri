package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	App      AppConfig
	Database DatabaseConfig
	JWT      JWTConfig
}

type AppConfig struct {
	Port string
	Env  string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

type JWTConfig struct {
	Secret string
	Expiry int // Hours
}

func LoadConfig() *Config {
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./backend")

	// Defaults
	viper.SetDefault("APP_PORT", "8080")
	viper.SetDefault("APP_ENV", "development")
	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", "5432")
	viper.SetDefault("DB_SSLMODE", "disable")
	viper.SetDefault("JWT_EXPIRY", 24)

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Warning: .env file not found or couldn't be loaded: %v. Using environment variables.", err)
	}

	var cfg Config
	cfg.App.Port = viper.GetString("APP_PORT")
	cfg.App.Env = viper.GetString("APP_ENV")
	
	cfg.Database.Host = viper.GetString("DB_HOST")
	cfg.Database.Port = viper.GetString("DB_PORT")
	cfg.Database.User = viper.GetString("DB_USER")
	cfg.Database.Password = viper.GetString("DB_PASSWORD")
	cfg.Database.Name = viper.GetString("DB_NAME")
	cfg.Database.SSLMode = viper.GetString("DB_SSLMODE")

	cfg.JWT.Secret = viper.GetString("JWT_SECRET")
	cfg.JWT.Expiry = viper.GetInt("JWT_EXPIRY")

	return &cfg
}
