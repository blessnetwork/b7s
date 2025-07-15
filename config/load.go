package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/camelcase"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/posflag"
	"github.com/knadh/koanf/v2"
	"github.com/spf13/pflag"

	"github.com/blessnetwork/b7s/models/bls"
)

const (
	defaultDelimiter = "."
	EnvDelimiter     = "_"
)

func Load() (*Config, error) {
	return load(os.Args[1:])
}

func load(args []string) (*Config, error) {

	var configPath string

	configOptions := flattenConfigOptions(getConfigOptions())
	flags, mapping, err := createFlags(configOptions)
	if err != nil {
		return nil, fmt.Errorf("could not create CLI flags for config: %w", err)
	}

	flags.StringVar(&configPath, "config", "", "path to a config file")

	// General flags.
	flags.Parse(args)

	delimiter := defaultDelimiter
	konfig := koanf.New(delimiter)

	err = konfig.Load(env.ProviderWithValue(bls.EnvPrefix, EnvDelimiter, envClean), nil)
	if err != nil {
		return nil, fmt.Errorf("could not load configuration from env: %w", err)
	}

	if configPath != "" {
		err = konfig.Load(file.Provider(configPath), yaml.Parser())
		if err != nil {
			return nil, fmt.Errorf("could not load config file: %w", err)
		}
	}

	// For the sake of usability flags have a flat structure - e.g. port or cpu-percentage-limit.
	// For use in config files, we prefer a structured layout, e.g. connectivity=>port or worker=>cpu-percentage-limit.
	// This callback translates the flag names from a flat layout to the structured one, so that koanf knows how to match
	// analogous values.
	translate := cliFlagTranslate(mapping, flags)

	err = konfig.Load(posflag.ProviderWithFlag(flags, delimiter, konfig, translate), nil)
	if err != nil {
		return nil, fmt.Errorf("could not load config: %w", err)
	}

	var cfg Config
	err = konfig.Unmarshal("", &cfg)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal konfig: %w", err)
	}

	err = cfg.valid()
	if err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}

func cliFlagTranslate(mapping map[string]string, fs *pflag.FlagSet) func(*pflag.Flag) (string, any) {

	return func(flag *pflag.Flag) (string, any) {
		key := flag.Name
		val := posflag.FlagVal(fs, flag)

		// Should not happen.
		skey, ok := mapping[key]
		if !ok {
			return key, val
		}

		return skey, val
	}
}

// envClean will do the following:
//
// - split environment variables to parts ("B7S_Connectivity_DialbackAddress" => [ "Connectivity", "DialbackAddress"])
// - translate individual parts from CamelCase to Kebab-Case ("DialbackAddress" => "Dialback-Address")
// - lowercase parts ("Dialback-Address" => "dialback-address")
// - join back parts using the environment variable delimiter (underscore) ("Connectivity_DialbackAddress" => "connectivity_dialback-address")
//
// Koanf then uses the underscore to determine structure and in which section the config option belongs.
func envClean(key string, value string) (string, any) {

	key = strings.TrimPrefix(key, bls.EnvPrefix)

	sections := strings.Split(key, EnvDelimiter)
	cleaned := make([]string, 0, len(sections))
	for _, part := range sections {
		p := strings.ToLower(strings.Join(camelcase.Split(part), "-"))
		cleaned = append(cleaned, p)
	}

	ss := strings.Join(cleaned, EnvDelimiter)

	switch ss {

	default:
		return ss, value

	// Kludge: For boot nodes and topics, return type should be a string slice.
	case "boot-nodes", "topics":
		return ss, strings.Split(value, ",")
	}
}
