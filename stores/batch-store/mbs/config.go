package mbs

type Config struct {
	dbname                    string
	initCollections           bool
	ignoreInitNamespaceErrors bool
}

var defaultConfig = Config{
	dbname:                    "b7s-db",
	initCollections:           true,
	ignoreInitNamespaceErrors: true,
}

type OptionFunc func(*Config)

func DBName(name string) OptionFunc {
	return func(cfg *Config) {
		cfg.dbname = name
	}
}

func InitCollections(b bool) OptionFunc {
	return func(cfg *Config) {
		cfg.initCollections = b
	}
}

func IgnoreInitErrors(b bool) OptionFunc {
	return func(cfg *Config) {
		cfg.ignoreInitNamespaceErrors = b
	}
}
