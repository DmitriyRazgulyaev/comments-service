package config

import "github.com/ilyakaznacheev/cleanenv"

type Config struct {
	SrvConfig ServerConfig
	DBConfig  DataBaseConfig
}

type ServerConfig struct {
	Port    string `env:"SERVER_PORT" env-default:"8080"`
	Storage string `env:"STORAGE" env-default:"postgres"`
}

type DataBaseConfig struct {
	Port     string `env:"POSTGRES_PORT" env-default:"5432"`
	Host     string `env:"POSTGRES_HOST" env-default:"localhost"`
	DBName   string `env:"POSTGRES_DB" env-default:"comment-service"`
	User     string `env:"POSTGRES_USER" env-default:"postgres"`
	Password string `env:"POSTGRES_PASSWORD"`
}

func ParseConfigFromEnv(path string) (*Config, error) {
	cfg := &Config{}

	if err := cleanenv.ReadConfig(path, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
