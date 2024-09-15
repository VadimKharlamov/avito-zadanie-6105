package config

import (
	"github.com/spf13/viper"
	"log"
	"time"
)

type Config struct {
	Env         string        `mapstructure:"ENV"`
	StoragePath string        `mapstructure:"POSTGRES_CONN"`
	Address     string        `mapstructure:"SERVER_ADDRESS"`
	Timeout     time.Duration `mapstructure:"SERVER_TIMEOUT"`
	IdleTimeout time.Duration `mapstructure:"SERVER_IDLE_TIMEOUT"`
	DbUsername  string        `mapstructure:"POSTGRES_USERNAME"`
	DbPassword  string        `mapstructure:"POSTGRES_PASSWORD"`
	DbHost      string        `mapstructure:"POSTGRES_HOST"`
	DbPort      string        `mapstructure:"POSTGRES_PORT"`
	DbDatabase  string        `mapstructure:"POSTGRES_DATABASE"`
}

func NewConfig() *Config {
	config := &Config{}
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		log.Fatal("Can't find the file .env : ", err)
	}

	err = viper.Unmarshal(&config)
	if err != nil {
		log.Fatal("Environment can't be loaded: ", err)
	}

	if config.Env == "local" {
		log.Println("The App is running in local env")
	}

	return config
}
