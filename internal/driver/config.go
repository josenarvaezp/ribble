package driver

import (
	"fmt"
	"io/ioutil"

	"github.com/josenarvaezp/displ/internal/objectstore"
	"gopkg.in/yaml.v2"
)

// Config represents the configuration file specified by the user
type Config struct {
	InputBuckets   []*objectstore.Bucket `yaml:"input"`
	OutputBucket   string                `yaml:"output"`
	Region         string                `yaml:"region"`
	MapperFuncName string                `yaml:"mapperFuncName"`
	Local          bool                  `yaml:"local"`
}

// ReadLocalConfigFile reads the config file from the driver's file system
// note that the path can be absolute or relative path
func ReadLocalConfigFile(path string) (*Config, error) {
	var conf Config

	confFile, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	err = yaml.Unmarshal(confFile, &conf)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return &conf, nil
}
