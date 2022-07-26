package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/couchbase/gocb/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/exp/slices"
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

func connectToCluster() (*gocb.Bucket, error) {
	cluster, err := gocb.Connect(viper.GetString("connection"), gocb.ClusterOptions{
		Authenticator: gocb.PasswordAuthenticator{
			Username: viper.GetString("username"),
			Password: viper.GetString("password"),
		},
		SecurityConfig: gocb.SecurityConfig{TLSSkipVerify: true},
	})
	if err != nil {
		return nil, fmt.Errorf("could not connect to cluster: %w", err)
	}

	bucket := cluster.Bucket(viper.GetString("bucket"))
	err = bucket.WaitUntilReady(time.Second, nil)
	if err != nil {
		return nil, fmt.Errorf("could not get bucket: %w", err)
	}

	return bucket, nil
}

func setupScopesAndConnections(bucket *gocb.Bucket, settingUp *bool) error {
	cm := bucket.Collections()

	scopes, err := cm.GetAllScopes(nil)
	if err != nil {
		return fmt.Errorf("could not get scopes: %w", err)
	}

	var collections []gocb.CollectionSpec

	i := slices.IndexFunc(scopes, func(s gocb.ScopeSpec) bool { return s.Name == "tacks" })
	if i == -1 {
		*settingUp = true
		fmt.Println("Setting up database")

		if err = cm.CreateScope("tacks", nil); err != nil {
			return fmt.Errorf("could not create 'tacks' scope: %w", err)
		}
	} else {
		collections = scopes[i].Collections
	}

	for _, collection := range []string{"internal", "stretches"} {
		if i := slices.IndexFunc(collections, func(c gocb.CollectionSpec) bool { return c.Name == collection }); i == -1 {
			if !*settingUp {
				*settingUp = true
				fmt.Println("Setting up database")
			}

			err = cm.CreateCollection(gocb.CollectionSpec{ScopeName: "tacks", Name: collection}, nil)
			if err != nil {
				return fmt.Errorf("could not create collection: %w", err)
			}
		}
	}

	return nil
}

func setupSDK() (*gocb.Scope, error) {
	bucket, err := connectToCluster()
	if err != nil {
		return nil, err
	}

	settingUp := false
	if err = setupScopesAndConnections(bucket, &settingUp); err != nil {
		return nil, err
	}

	scope := bucket.Scope("tacks")

	_, err = scope.Collection("internal").Binary().Increment("next-id", &gocb.IncrementOptions{Initial: 1})

	return scope, nil
}
