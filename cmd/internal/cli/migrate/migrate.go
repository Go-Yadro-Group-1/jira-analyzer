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

const (
	defaultConfig  = "config/dev.yaml"
	noVersionValue = 0
)

var errInvalidDownVersion = errors.New("--down value must be >= 0")

//nolint:exhaustruct
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Run database migrations",
		Long: "Apply or roll back database migrations.\n" +
			"  --version N   apply up to version N (0 = apply all)\n" +
			"  --down N      roll back to version N (0 = roll back all)\n" +
			"  --force N     force-set version N to clear dirty state",
		RunE:         run,
		SilenceUsage: true,
	}

	cmd.Flags().String("config", defaultConfig, "path to config file")
	cmd.Flags().Uint(
		"version", noVersionValue, "target migration version to apply up to (0 = apply all up)",
	)
	cmd.Flags().Int(
		"down", -1, "roll back to this migration version (0 = roll back all, omit = no rollback)",
	)
	cmd.Flags().Int(
		"force", -1, "force-set schema version without running migrations, clears dirty state",
	)

	return cmd
}

type direction int

const (
	directionUp    direction = iota
	directionDown  direction = iota
	directionForce direction = iota
)

type migrateParams struct {
	direction direction
	version   uint
}

func run(cmd *cobra.Command, _ []string) error {
	cfg, err := loadConfig(cmd)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	params, err := parseFlags(cmd)
	if err != nil {
		return err
	}

	database, err := sql.Open("postgres", cfg.DB.DSN())
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer database.Close()

	err = runMigrations(database, params)
	if err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}

	return nil
}

func parseFlags(cmd *cobra.Command) (migrateParams, error) {
	if cmd.Flags().Changed("force") {
		forceVal, err := cmd.Flags().GetInt("force")
		if err != nil {
			return migrateParams{}, fmt.Errorf("get force flag: %w", err)
		}

		if forceVal < 0 {
			return migrateParams{}, errors.New("--force value must be >= 0") //nolint:err113
		}

		return migrateParams{direction: directionForce, version: uint(forceVal)}, nil //nolint:gosec
	}

	if cmd.Flags().Changed("down") {
		downVal, err := cmd.Flags().GetInt("down")
		if err != nil {
			return migrateParams{}, fmt.Errorf("get down flag: %w", err)
		}

		if downVal < 0 {
			return migrateParams{}, errInvalidDownVersion
		}

		return migrateParams{direction: directionDown, version: uint(downVal)}, nil //nolint:gosec
	}

	upVersion, err := cmd.Flags().GetUint("version")
	if err != nil {
		return migrateParams{}, fmt.Errorf("get version flag: %w", err)
	}

	return migrateParams{direction: directionUp, version: upVersion}, nil
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

func runMigrations(database *sql.DB, params migrateParams) error {
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

	err = applyMigration(migrator, params)

	switch {
	case err == nil:
		log.Println("migrations applied successfully")
	case errors.Is(err, migrate.ErrNoChange):
		log.Println("no changes to apply")
	default:
		return fmt.Errorf("migrate: %w", err)
	}

	return nil
}

func applyMigration(migrator *migrate.Migrate, params migrateParams) error {
	switch params.direction {
	case directionForce:
		err := migrator.Force(int(params.version)) //nolint:gosec
		if err != nil {
			return fmt.Errorf("force version %d: %w", params.version, err)
		}

		return nil

	case directionDown:
		if params.version == noVersionValue {
			return migrator.Down() //nolint:wrapcheck
		}

		return migrator.Migrate(params.version) //nolint:wrapcheck

	case directionUp:
		if params.version == noVersionValue {
			return migrator.Up() //nolint:wrapcheck
		}

		return migrator.Migrate(params.version) //nolint:wrapcheck
	}

	return nil
}
