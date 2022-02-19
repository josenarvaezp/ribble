package driver

import (
	"io/ioutil"

	"github.com/josenarvaezp/displ/internal/objectstore"
	"gopkg.in/yaml.v2"
)

// Config represents the configuration file specified by the user
type Config struct {
	InputBuckets []*objectstore.Bucket `yaml:"input"`
	OutputBucket string                `yaml:"output"`
	Region       string                `yaml:"region"`
	Local        bool                  `yaml:"local"`
	LogLevel     int                   `yaml:"logLevel"`
	AccountID    string                `yaml:"accountID"`
}

// ReadLocalConfigFile reads the config file from the driver's file system
// note that the path can be absolute or relative path
func ReadLocalConfigFile(path string) (*Config, error) {
	var conf Config

	confFile, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(confFile, &conf)
	if err != nil {
		return nil, err
	}

	return &conf, nil
}
