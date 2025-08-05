package config

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/knadh/koanf/providers/structs"
	"github.com/spf13/pflag"
)

func createFlags(fields []ConfigOption) (*pflag.FlagSet, map[string]string, error) {

	fs := pflag.NewFlagSet("b7s-node", pflag.ExitOnError)
	fs.SortFlags = false

	for _, field := range fields {
		err := addFlag(fs, field.CLI)
		if err != nil {
			return nil, nil, fmt.Errorf("could not add flag for config (name: %v, flag: %v, type: %v)", field.FullPath, field.CLI.Flag, field.kind.String())
		}
	}

	mapping, err := mapCLIFlagsToConfig(fields)
	if err != nil {
		return nil, nil, fmt.Errorf("could not get mapping of CLI flags to config: %w", err)
	}

	return fs, mapping, nil
}

func addFlag(fs *pflag.FlagSet, fc CLIFlag) error {

	if fc.Flag == "" {
		return nil
	}

	switch def := fc.Default.(type) {
	case uint:
		fs.UintP(fc.Flag, fc.Shorthand, def, fc.Description)

	case string:
		fs.StringP(fc.Flag, fc.Shorthand, def, fc.Description)

	case float64:
		fs.Float64P(fc.Flag, fc.Shorthand, def, fc.Description)

	case int64:
		fs.Int64P(fc.Flag, fc.Shorthand, def, fc.Description)

	case bool:
		fs.BoolP(fc.Flag, fc.Shorthand, def, fc.Description)

	case []string:
		fs.StringSliceP(fc.Flag, fc.Shorthand, nil, fc.Description)

	case time.Duration:
		fs.DurationP(fc.Flag, fc.Shorthand, def, fc.Description)

	default:
		return errors.New("unsupported type for a CLI flag. Extend support by adding handling for the new flag type")
	}

	return nil
}

func getFlagFromTag(tag string) (string, string) {

	tag = strings.TrimSpace(tag)

	fields := strings.Split(tag, ",")
	switch len(fields) {
	case 0:
		return "", ""
	case 1:
		return fields[0], ""
	default:
		return fields[0], fields[1]
	}
}

// return mapping of CLI flag to the config path used by koanf. E.g. address => connectivity.address.
// We don't have to enfore uniqueness of CLI flags as pflag does that for us.
func mapCLIFlagsToConfig(fields []ConfigOption) (map[string]string, error) {

	flags := make(map[string]string)
	for _, field := range fields {
		if field.CLI.Flag == "" {
			continue
		}
		flags[field.CLI.Flag] = field.FullPath
	}

	return flags, nil
}

func getDefaultFlagValues() map[string]any {

	cfg := structs.Provider(DefaultConfig, "koanf")
	defaults, err := cfg.Read()
	if err != nil {
		return nil
	}

	flat := make(map[string]any)
	flattenMap("", defaults, flat)
	return flat
}
