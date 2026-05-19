package server

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/Go-Yadro-Group-1/Jira-Analyzer/cmd/internal/config"
	"github.com/Go-Yadro-Group-1/Jira-Analyzer/internal/app"
	"github.com/Go-Yadro-Group-1/Jira-Analyzer/internal/logger"
	"github.com/Go-Yadro-Group-1/Jira-Analyzer/internal/metrics"
	"github.com/golang-migrate/migrate/v4"
	migratepg "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file" // migrate file source
	_ "github.com/lib/pq"                                // postgres driver
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	defaultHost = "0.0.0.0"
	defaultPort = 50051

	defaultConfig     = "config/dev.yaml"
	defaultMigrations = "migrations"
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
	cmd.Flags().String("migrations", defaultMigrations, "path to migrations directory")

	return cmd
}

func run(cmd *cobra.Command, _ []string) error {
	cfg, err := loadConfig(cmd)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	log, err := logger.New(os.Stdout, cfg.App.LogLevel)
	if err != nil {
		return fmt.Errorf("init logger: %w", err)
	}

	slog.SetDefault(log)

	mtr := metrics.New()
	mtr.RegisterRuntimeCollectors()

	migrationsDir, err := cmd.Flags().GetString("migrations")
	if err != nil {
		return fmt.Errorf("get migrations flag: %w", err)
	}

	err = runMigrations(&cfg.DB, migrationsDir, log)
	if err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}

	conn, err := connectDB(cmd.Context(), cfg)
	if err != nil {
		return fmt.Errorf("connect db: %w", err)
	}
	defer conn.Close()

	metricsSrv := startMetricsServer(cfg.Metrics.Port, mtr, log)
	defer shutdownMetricsServer(metricsSrv, log)

	err = startServer(cmd, conn, log, mtr)
	if err != nil {
		return fmt.Errorf("start server: %w", err)
	}

	return nil
}

const (
	metricsShutdownTimeout = 5 * time.Second
	metricsReadTimeout     = 5 * time.Second
)

func startMetricsServer(port int, mtr *metrics.Metrics, log *slog.Logger) *http.Server {
	handlerOpts := promhttp.HandlerOpts{Registry: mtr.Registry} //nolint:exhaustruct

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(mtr.Registry, handlerOpts))

	srv := &http.Server{ //nolint:exhaustruct
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           mux,
		ReadHeaderTimeout: metricsReadTimeout,
	}

	go func() {
		log.Info("starting metrics server", slog.String("addr", srv.Addr))

		err := srv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("metrics server failed", slog.String("error", err.Error()))
		}
	}()

	return srv
}

func shutdownMetricsServer(srv *http.Server, log *slog.Logger) {
	shutdownCtx, cancel := context.WithTimeout(context.Background(), metricsShutdownTimeout)
	defer cancel()

	err := srv.Shutdown(shutdownCtx)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Warn("metrics server shutdown", slog.String("error", err.Error()))
	}
}

func runMigrations(cfg *config.DBConfig, dir string, log *slog.Logger) error {
	conn, err := sql.Open("postgres", cfg.DSN())
	if err != nil {
		return fmt.Errorf("open migrate db: %w", err)
	}
	defer conn.Close()

	driver, err := migratepg.WithInstance(conn, &migratepg.Config{}) //nolint:exhaustruct
	if err != nil {
		return fmt.Errorf("init migrate driver: %w", err)
	}

	migr, err := migrate.NewWithDatabaseInstance("file://"+dir, "postgres", driver)
	if err != nil {
		return fmt.Errorf("init migrate: %w", err)
	}
	defer migr.Close()

	err = migr.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrate up: %w", err)
	}

	log.Info("migrations applied", slog.String("dir", dir))

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

func startServer(cmd *cobra.Command, conn *sql.DB, log *slog.Logger, mtr *metrics.Metrics) error {
	host, err := cmd.Flags().GetString("host")
	if err != nil {
		return fmt.Errorf("get host flag: %w", err)
	}

	port, err := cmd.Flags().GetInt("port")
	if err != nil {
		return fmt.Errorf("get port flag: %w", err)
	}

	addr := fmt.Sprintf("%s:%d", host, port)

	var lc net.ListenConfig

	listener, err := lc.Listen(cmd.Context(), "tcp", addr)
	if err != nil {
		return fmt.Errorf("listen %s: %w", addr, err)
	}

	server := app.NewGRPCServer(conn, log, mtr)

	go func() {
		<-cmd.Context().Done()
		server.GracefulStop()
	}()

	log.Info("starting gRPC server", slog.String("addr", addr))

	err = server.Serve(listener)
	if err != nil {
		return fmt.Errorf("serve: %w", err)
	}

	return nil
}
