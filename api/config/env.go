package config

import (
	"fmt"
	"net/url"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBUrl      string
	JWTSecret  string
	Env        Environment
}

type Environment string

const (
	Development = "development"
	Production  = "production"
)

func (e Environment) IsProduction() bool {
	return e == Production
}

func buildDBURL(c Config) string {
	u := &url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(c.DBUser, c.DBPassword),
		Host:     fmt.Sprintf("%s:%s", c.DBHost, c.DBPort),
		Path:     fmt.Sprintf("/%s", c.DBName),
		RawQuery: "sslmode=disable",
	}
	return u.String()
}

func requireEnv(key string) (string, error) {
	v := os.Getenv(key)
	if v == "" {
		return "", fmt.Errorf("%v is required", key)
	}
	return v, nil
}

func LoadEnv() (*Config, error) {
	c := Config{}
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("godotenv load: %w", err)
	}

	// TODO: evaluate which DB fields should be made required
	// also add a default fallback value for specifically these fields
	c.DBHost = os.Getenv("DB_HOST")
	c.DBPort = os.Getenv("DB_PORT")
	c.DBUser = os.Getenv("DB_USER")
	c.DBPassword = os.Getenv("DB_PASSWORD")
	c.DBName = os.Getenv("DB_NAME")
	c.DBUrl = buildDBURL(c)

	c.JWTSecret, err = requireEnv("JWT_SECRET")
	if err != nil {
		return nil, err
	}

	if v := os.Getenv("APP_ENV"); v != "" {
		c.Env = Environment(v)
	} else {
		c.Env = Development
	}

	return &c, nil
}
