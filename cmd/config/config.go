package config

import (
	"encoding/json"
	"github.com/rs/zerolog/log"
	"io/ioutil"
)

type Config struct {
	Sentinels []string          `json:"sentinels"`
	Databases []*DatabaseConfig `json:"databases"`
}

type DatabaseConfig struct {
	Name string `json:"name"`
	Port uint16 `json:"port"`
}

func CreateFromFile(path string) (*Config, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		log.Error().Err(err).Msgf("read config file: %s", path)

		return nil, err
	}

	conf := &Config{}
	err = json.Unmarshal(b, conf)
	if err != nil {
		log.Error().Err(err).Msgf("parse config file: %s", path)

		return nil, err
	}

	return conf, err
}
