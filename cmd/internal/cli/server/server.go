package server

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/Go-Yadro-Group-1/Jira-Analyzer/cmd/internal/config"
	_ "github.com/lib/pq" // postgres driver
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	defaultHost = "0.0.0.0"
	defaultPort = 50051

	defaultConfig = "config/dev.yaml"
)

//nolint:exhaustruct
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "serve",
		Short:        "Start the Jira Analyzer gRPC server",
		Long:         "Start the Jira Analyzer gRPC server to analyze Jira issues.",
		RunE:         run,
		SilenceUsage: true,
	}

	cmd.Flags().String("host", defaultHost, "gRPC server host")
	cmd.Flags().Int("port", defaultPort, "gRPC server port")
	cmd.Flags().String("config", defaultConfig, "path to config file")

	return cmd
}

func run(cmd *cobra.Command, _ []string) error {
	cfg, err := loadConfig(cmd)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	conn, err := connectDB(cmd.Context(), cfg)
	if err != nil {
		return fmt.Errorf("connect db: %w", err)
	}
	defer conn.Close()

	err = startServer(cmd, cfg)
	if err != nil {
		return fmt.Errorf("start server: %w", err)
	}

	return nil
}

func loadConfig(cmd *cobra.Command) (*config.Config, error) {
	cfgFile, err := cmd.Flags().GetString("config")
	if err != nil {
		return nil, fmt.Errorf("get config flag: %w", err)
	}

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	}

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	return cfg, nil
}

func connectDB(ctx context.Context, cfg *config.Config) (*sql.DB, error) {
	conn, err := sql.Open("postgres", cfg.DB.DSN())
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	err = conn.PingContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}

	return conn, nil
}

func startServer(cmd *cobra.Command, _ *config.Config) error {
	host, err := cmd.Flags().GetString("host")
	if err != nil {
		return fmt.Errorf("get host flag: %w", err)
	}

	port, err := cmd.Flags().GetInt("port")
	if err != nil {
		return fmt.Errorf("get port flag: %w", err)
	}

	log.Printf("starting gRPC server on %s:%d\n", host, port)

	return nil
}
