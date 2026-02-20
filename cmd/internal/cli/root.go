package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/Go-Yadro-Group-1/Jira-Analyzer/cmd/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	defaultConfigDir  = "config"
	defaultConfigName = "dev"
	defaultConfigType = "yaml"
)

// nolint: gochecknoglobals
var (
	cfgFile    string
	ErrService = errors.New("error on initialize service")
)

// nolint: gochecknoglobals
var rootCmd = &cobra.Command{
	Use:   "Jira-analyzer",
	Short: "Jira Analyzer is analyzer service for Jira tasks",
	Long: `Jira Analyzer need config file,
that contains information about DB,
Gateway service, and app configuration.\n
You can check config file in config path`,
}

func Execute() error {
	err := rootCmd.Execute()
	if err != nil {
		return fmt.Errorf("%w: %w", ErrService, err)
	}

	return nil
}

// nolint: gochecknoinits
func init() {
	cobra.OnInitialize(initConfig)

	configMsg := fmt.Sprintf(
		"default config is $%s/%s.%s",
		defaultConfigDir,
		defaultConfigName,
		defaultConfigType,
	)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", configMsg)

	toggleMsg := "Help message for toggle"
	rootCmd.Flags().BoolP("toggle", "t", false, toggleMsg)
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		workDir, err := os.Getwd()
		if err != nil {
			cobra.CheckErr(err)
		}

		configPath := workDir + "/" + defaultConfigDir

		viper.AddConfigPath(configPath)
		viper.SetConfigType(defaultConfigType)
		viper.SetConfigName(defaultConfigName)
	}

	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Config read failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintln(os.Stderr, "Config file:", viper.ConfigFileUsed())

	_, err = config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Config validation failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintln(os.Stderr, "Config loaded")
}
