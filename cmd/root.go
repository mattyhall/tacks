/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "tacks",
	Short: "A time tracking application",
	RunE: func(cmd *cobra.Command, args []string) error {
		err := validateConfig()
		if err != nil {
			return err
		}

		return nil
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.config/tacks/config.toml)")
}

func initConfig() {
	viper.SetDefault("connection", "couchbase://localhost:8091")

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(filepath.Join(home, ".config", "tacks"))
		viper.SetConfigType("toml")
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

type ErrMissingFields struct {
	fields []string
}

func (e *ErrMissingFields) Error() string {
	return fmt.Sprintf("missing the following fields in the config: %q", strings.Join(e.fields, ", "))
}

var _ error = (*ErrMissingFields)(nil)

func validateConfig() error {
	missing := []string{}
	for _, field := range []string{"username", "password", "bucket"} {
		if !viper.IsSet(field) {
			missing = append(missing, field)
		}
	}

	if len(missing) > 0 {
		return &ErrMissingFields{fields: missing}
	}

	return nil
}
