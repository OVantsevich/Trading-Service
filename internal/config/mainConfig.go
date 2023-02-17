// Package config main config
package config

import (
	"fmt"

	"github.com/caarlos0/env/v7"
)

// MainConfig with init data
type MainConfig struct {
	PostgresPort     string `env:"POSTGRES_DB_URL,notEmpty" envDefault:"5434"`
	PostgresHost     string `env:"POSTGRES_DB_URL,notEmpty" envDefault:"localhost"`
	PostgresPassword string `env:"POSTGRES_PASSWORD,notEmpty" envDefault:"postgres"`
	PostgresUser     string `env:"POSTGRES_USER,notEmpty" envDefault:"postgres"`
	PostgresDB       string `env:"POSTGRES_DB,notEmpty" envDefault:"postgres"`
	Port             string `env:"PORT,notEmpty" envDefault:"40000"`
}

// NewMainConfig parsing config from environment
func NewMainConfig() (*MainConfig, error) {
	mainConfig := &MainConfig{}

	err := env.Parse(mainConfig)
	if err != nil {
		return nil, fmt.Errorf("config - NewMainConfig - Parse:%w", err)
	}

	return mainConfig, nil
}
