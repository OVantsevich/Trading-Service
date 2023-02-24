// Package config main config
package config

import (
	"fmt"

	"github.com/caarlos0/env/v7"
)

// MainConfig with init data
type MainConfig struct {
	PostgresPort     string `env:"POSTGRES_DB_PORT,notEmpty" envDefault:"5435"`
	PostgresHost     string `env:"POSTGRES_DB_HOST,notEmpty" envDefault:"localhost"`
	PostgresPassword string `env:"POSTGRES_PASSWORD,notEmpty" envDefault:"postgres"`
	PostgresUser     string `env:"POSTGRES_USER,notEmpty" envDefault:"postgres"`
	PostgresDB       string `env:"POSTGRES_DB,notEmpty" envDefault:"postgres"`
	Port             string `env:"PORT,notEmpty" envDefault:"40000"`

	PriceServicePort string `env:"PRICE_SERVICE_PORT,notEmpty" envDefault:"10000"`
	PriceServiceHost string `env:"PRICE_SERVICE_HOST,notEmpty" envDefault:"localhost"`

	PaymentServicePort string `env:"PAYMENT_SERVICE_PORT,notEmpty" envDefault:"30000"`
	PaymentServiceHost string `env:"PAYMENT_SERVICE_HOST,notEmpty" envDefault:"localhost"`
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
