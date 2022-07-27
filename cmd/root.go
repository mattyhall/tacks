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
	Use:          "tacks",
	Short:        "A time tracking application",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {

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
	viper.SetDefault("connection", "couchbase://localhost")
	viper.SetDefault("scope", "tacks")

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(".")
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
