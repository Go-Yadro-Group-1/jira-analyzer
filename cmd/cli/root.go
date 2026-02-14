package cli

import (
	"fmt"
	"os"

	"github.com/Go-Yadro-Group-1/Jira-Analyzer/cmd/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	configDir = "config"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "jira-analyzer",
	Short: "Jira Analyzer",
	Long:  "Jira Analyzer",
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.Jira-Connector.yaml)")
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		workDir, err := os.Getwd()
		if err != nil {
			cobra.CheckErr(err)
		}

		configPath := workDir + "/" + configDir

		viper.AddConfigPath(configPath)
		viper.SetConfigType("yaml")
		viper.SetConfigName("local")
	}

	viper.AutomaticEnv() // read in environment variables that match

	if err := viper.ReadInConfig(); err != nil {
		fmt.Fprintf(os.Stderr, "Config read failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintln(os.Stderr, "Config file:", viper.ConfigFileUsed())

	if _, err := config.LoadConfig(); err != nil {
		fmt.Fprintf(os.Stderr, "Config validation failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintln(os.Stderr, "Config loaded")
}
