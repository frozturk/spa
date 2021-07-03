package spa

import (
	"errors"
	"os"
)

type Config struct {
	SPADirectory string
	NPMScript    string
}

var (
	defaultScript = "start"
)

func (c *Config) validate() error {
	if _, err := os.Stat(c.SPADirectory); os.IsNotExist(err) {
		return errors.New("spa directory does not exist")
	}
	return nil
}

func newConfig(config *Config) (*Config, error) {
	spaConfig := &Config{NPMScript: defaultScript, SPADirectory: config.SPADirectory}
	if config.NPMScript != "" {
		spaConfig.NPMScript = config.NPMScript
	}
	err := spaConfig.validate()
	if err != nil {
		return nil, err
	}
	return spaConfig, nil
}
