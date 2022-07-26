package cmd

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/mattyhall/tacks/internal"
	"github.com/spf13/cobra"
)

func parseAttrs(attrs []string) (map[string]string, error) {
	ret := map[string]string{}

	for _, attr :=  range attrs {
		split := strings.SplitN(attr, ":", 2)
		if len(split) != 2 {
			return nil, fmt.Errorf("attributes must be in the format 'key:value'")
		}

		ret[split[0]] = split[1]
	}

	return ret, nil
}

func run(cmd *cobra.Command) error {
	err := validateConfig()
	if err != nil {
		return err
	}

	scope, err := setupSDK()
	if err != nil {
		return err
	}

	internalCol := scope.Collection("internal")
	col := scope.Collection("stretches")

	id, err := internal.GetID(internalCol)
	if err != nil {
		return fmt.Errorf("could not get id for stretch: %w", err)
	}

	description, _ := cmd.Flags().GetString("description")
	tags, _ := cmd.Flags().GetStringSlice("tag")
	attrs, _ := cmd.Flags().GetStringSlice("attr")

	realAttrs, err := parseAttrs(attrs)
	if err != nil {
		return err
	}

	stretch := internal.Stretch{
		ID:          id,
		Description: description,
		Start:       time.Now(),
		Tags:        tags,
		Attributes:  realAttrs,
	}

	_, err = col.Insert(strconv.FormatUint(id, 10), &stretch, nil)
	if err != nil {
		return fmt.Errorf("could not insert stretch: %w", err)
	}

	fmt.Printf("Recording stretch %d\n", id)

	return nil
}

var recordCmd = &cobra.Command{
	Use:   "record",
	Short: "Block whilst tracking time",
	RunE:  func(cmd *cobra.Command, args []string) error { return run(cmd) },
}

func init() {
	rootCmd.AddCommand(recordCmd)

	recordCmd.PersistentFlags().String("description", "", "Description of what will be done in this stretch")
	recordCmd.PersistentFlags().StringSlice("tag", []string{}, "Used to set the tags of the stretch")
	recordCmd.PersistentFlags().StringSlice("attr", []string{}, "Used to set the attributes of the stretch. In the form 'key:value'")
}
