package internal

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type Config struct {
	Domain     string `yaml:"domain"`
	TunnelName string `yaml:"tunnel_name"`
	Interval   int    `yaml:"interval"`
	IPVersion  string `yaml:"ip_version"`
}

func loadConfig(file string) (*Config, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	config := &Config{}
	err = yaml.Unmarshal(data, config)
	if err != nil {
		return nil, err
	}

	return config, nil
}
