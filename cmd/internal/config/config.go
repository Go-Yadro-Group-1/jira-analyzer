package config

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

type Config struct {
	DB      DBConfig      `mapstructure:"db"      validate:"required"`
	App     AppConfig     `mapstructure:"app"     validate:"required"`
	Metrics MetricsConfig `mapstructure:"metrics" validate:"required"`
}

type DBConfig struct {
	Host     string `mapstructure:"host"     validate:"required,hostname|ip"`
	Port     int    `mapstructure:"port"     validate:"required,min=1,max=65535"`
	User     string `mapstructure:"user"     validate:"required,min=1"`
	Password string `mapstructure:"password" validate:"required,min=1"`
	DBName   string `mapstructure:"dbname"   validate:"required,min=1"`
	SSLMode  string `mapstructure:"sslmode"  validate:"required,oneof=disable require verify-ca verify-full"`
}

type AppConfig struct {
	LogLevel string `mapstructure:"log_level" validate:"required,oneof=debug info warn error"`
}

type MetricsConfig struct {
	Port int `mapstructure:"port" validate:"required,min=1,max=65535"`
}

//nolint:gochecknoglobals
var validate = validator.New()

// BindEnvs binds every overridable config key to an environment variable so a
// service can be configured entirely from the environment (.env) — secrets stay
// out of committed YAML, and deploy targets that only accept env vars (e.g.
// Timeweb App Platform) can drive the full config. Explicit BindEnv is required
// because viper's AutomaticEnv does not populate nested keys during Unmarshal
// when the key is absent from the config file. Env names are kept consistent
// across services (DB_NAME, not DB_DBNAME; LOG_LEVEL).
func BindEnvs() error {
	binds := map[string]string{
		"db.host":       "DB_HOST",
		"db.port":       "DB_PORT",
		"db.user":       "DB_USER",
		"db.password":   "DB_PASSWORD",
		"db.dbname":     "DB_NAME",
		"db.sslmode":    "DB_SSLMODE",
		"app.log_level": "LOG_LEVEL",
		"metrics.port":  "METRICS_PORT",
	}

	for key, env := range binds {
		err := viper.BindEnv(key, env)
		if err != nil {
			return fmt.Errorf("bind env %s: %w", env, err)
		}
	}

	return nil
}

func LoadConfig() (*Config, error) {
	var cfg Config

	err := viper.Unmarshal(&cfg)
	if err != nil {
		return nil, fmt.Errorf("unable to decode config: %w", err)
	}

	err = validate.Struct(&cfg)
	if err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return &cfg, nil
}

func (d *DBConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.DBName, d.SSLMode,
	)
}
