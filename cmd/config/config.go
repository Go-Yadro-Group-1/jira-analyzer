/*
Copyright © 2026 German-Feskov
*/
package config

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

type Config struct {
	DB  DBConfig  `mapstructure:"db" validate:"required"`
	App AppConfig `mapstructure:"app" validate:"required"`
}

type DBConfig struct {
	Host     string `mapstructure:"host" validate:"required,hostname|ip"`
	Port     int    `mapstructure:"port" validate:"required,min=1,max=65535"`
	User     string `mapstructure:"user" validate:"required,min=1"`
	Password string `mapstructure:"password" validate:"required,min=1"`
	DBName   string `mapstructure:"dbname" validate:"required,min=1"`
	SSLMode  string `mapstructure:"sslmode" validate:"required,oneof=disable require verify-ca verify-full"`
}

// TODO add some vars
type AppConfig struct {
	LogLevel string `mapstructure:"log_level" validate:"required,oneof=debug info warn error"`
}

//TODO add config for gRPC

var (
	appConfig *Config
	validate  *validator.Validate
)

func init() {
	validate = validator.New()
}

func LoadConfig() (*Config, error) {
	var cfg Config

	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unable to decode config: %w", err)
	}

	if err := validate.Struct(&cfg); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	if err := cfg.customValidate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	appConfig = &cfg
	return &cfg, nil
}

func GetConfig() *Config {
	if appConfig == nil {
		panic("config not loaded")
	}
	return appConfig
}

func (c *Config) customValidate() error {
	// Example of custom validation, TODO edit
	return nil
}

func (d *DBConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.DBName, d.SSLMode,
	)
}
