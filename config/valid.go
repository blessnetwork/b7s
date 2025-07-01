package config

import (
	"errors"
)

func (c *Config) valid() error {

	if (c.Head.Batch.Server != "") != (c.Head.Batch.DBName != "") {
		return errors.New("batch db server and database name have to be either set or unset")
	}

	return nil
}
