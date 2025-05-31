package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type PostgresConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DB       string `yaml:"db"`
}

type RedisConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
	DB   int    `yaml:"db"`
}

type RabbitMQConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Queue    string `yaml:"queue"`
}

type JWTConfig struct {
	Secret     string `yaml:"secret"`
	TTLMinutes int    `yaml:"ttl_minutes"`
}

type AppConfig struct {
	Port int `yaml:"port"`
}

type Config struct {
	Postgres PostgresConfig `yaml:"postgres"`
	Redis    RedisConfig    `yaml:"redis"`
	RabbitMQ RabbitMQConfig `yaml:"rabbitmq"`
	JWT      JWTConfig      `yaml:"jwt"`
	App      AppConfig      `yaml:"app"`
}

func LoadConfig() (*Config, error) {
	cfg := &Config{}
	if _, err := os.Stat("config.yaml"); err != nil {
		return nil, err
	}

	data, err := os.ReadFile("config.yaml")
	if err != nil {
		return nil, err
	}

	if err = yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
