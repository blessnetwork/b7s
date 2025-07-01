package executor

import (
	"github.com/hashicorp/go-metrics"
	"github.com/spf13/afero"

	"github.com/blessnetwork/b7s/models/bls"
)

// defaultConfig used to create Executor.
var defaultConfig = Config{
	WorkDir:         "workspace",
	RuntimeDir:      "",
	ExecutableName:  bls.RuntimeCLI(),
	FS:              afero.NewOsFs(),
	Limiter:         &noopLimiter{},
	DriversRootPath: "",
}

// Config represents the Executor configuration.
type Config struct {
	WorkDir         string           // directory where files needed for the execution are stored
	RuntimeDir      string           // directory where the executable can be found
	ExecutableName  string           // name for the executable
	DriversRootPath string           // where are cgi drivers stored
	FS              afero.Fs         // FS accessor
	Limiter         Limiter          // Resource limiter for executed processes
	Metrics         *metrics.Metrics // Metrics handle
}

type Option func(*Config)

// WithWorkDir sets the workspace directory for the executor.
func WithWorkDir(dir string) Option {
	return func(cfg *Config) {
		cfg.WorkDir = dir
	}
}

// WithRuntimeDir sets the runtime directory for the executor.
func WithRuntimeDir(dir string) Option {
	return func(cfg *Config) {
		cfg.RuntimeDir = dir
	}
}

// WithFS sets the FS handler used by the executor.
func WithFS(fs afero.Fs) Option {
	return func(cfg *Config) {
		cfg.FS = fs
	}
}

// WithExecutableName sets the name of the executable that should be ran.
func WithExecutableName(name string) Option {
	return func(cfg *Config) {
		cfg.ExecutableName = name
	}
}

// WithLimiter sets the resource limiter called for each individual execution.
func WithLimiter(limiter Limiter) Option {
	return func(cfg *Config) {
		cfg.Limiter = limiter
	}
}

// WithMetrics sets the metrics handler.
func WithMetrics(metrics *metrics.Metrics) Option {
	return func(cfg *Config) {
		cfg.Metrics = metrics
	}
}
