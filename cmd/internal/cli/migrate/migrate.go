package migrate

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/Go-Yadro-Group-1/Jira-Analyzer/cmd/internal/config"
	"github.com/Go-Yadro-Group-1/Jira-Analyzer/migrations"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/lib/pq" // postgres driver
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const defaultConfig = "config/dev.yaml"

//nolint:exhaustruct
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "migrate",
		Short:        "Run database migrations",
		Long:         "Apply database migrations. Without --version applies all available migrations (up).",
		RunE:         run,
		SilenceUsage: true,
	}

	cmd.Flags().String("config", defaultConfig, "path to config file")
	cmd.Flags().Uint("version", 0, "target migration version (0 = apply all up)")

	return cmd
}

func run(cmd *cobra.Command, _ []string) error {
	cfg, err := loadConfig(cmd)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	version, err := cmd.Flags().GetUint("version")
	if err != nil {
		return fmt.Errorf("get version flag: %w", err)
	}

	database, err := sql.Open("postgres", cfg.DB.DSN())
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer database.Close()

	err = runMigrations(database, version)
	if err != nil {
		return fmt.Errorf("run migrations: %w", err)
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

func runMigrations(database *sql.DB, version uint) error {
	sourceDriver, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return fmt.Errorf("create iofs source: %w", err)
	}

	dbDriver, err := postgres.WithInstance(database, &postgres.Config{}) //nolint:exhaustruct
	if err != nil {
		return fmt.Errorf("create postgres driver: %w", err)
	}

	migrator, err := migrate.NewWithInstance("iofs", sourceDriver, "postgres", dbDriver)
	if err != nil {
		return fmt.Errorf("create migrator: %w", err)
	}
	defer migrator.Close()

	if version == 0 {
		err = migrator.Up()
	} else {
		err = migrator.Migrate(version)
	}

	switch {
	case err == nil:
		log.Println("migrations applied successfully")
	case errors.Is(err, migrate.ErrNoChange):
		log.Println("no new migrations to apply")
	default:
		return fmt.Errorf("migrate: %w", err)
	}

	return nil
}
