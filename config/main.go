package config

import (
	"strconv"

	"github.com/caarlos0/env/v10"
)

type Config struct {
	DatabaseUser string `env:"DATABASE_USER" envDefault:"root"`
	DatabasePass string `env:"DATABASE_PASS" envDefault:"mysql"`
	DatabaseHost string `env:"DATABASE_HOST" envDefault:"localhost"`
	DatabasePort int    `env:"DATABASE_PORT" envDefault:"3306"`
	DatabaseName string `env:"DATABASE_NAME" envDefault:"auto_trade"`
}

func NewConfig() (Config, error) {
	var config Config
	if error := env.Parse(&config); error != nil {
		return Config{}, error
	}
	return config, nil
}

func (config Config) DatabaseURL() string {
	return config.DatabaseUser + ":" + config.DatabasePass +
		"@tcp(" + config.DatabaseHost + ":" + strconv.Itoa(config.DatabasePort) + ")" +
		"/" + config.DatabaseName +
		"?multiStatements=true"
}
