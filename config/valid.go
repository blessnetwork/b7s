package config

import (
	"errors"
)

// TODO: Move more of the validation into this function.
func (c *Config) valid() error {

	if (c.Head.BatchDBServer != "") != (c.Head.BatchDBName != "") {
		return errors.New("batch db server and database name have to be either set or unset")
	}

	return nil
}
