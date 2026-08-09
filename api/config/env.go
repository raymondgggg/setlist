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

func LoadEnv() (*Config, error) {
	c := Config{}
	err := godotenv.Load()
	if err != nil {
		return nil, err
	}

	c.DBHost = os.Getenv("DB_HOST")
	c.DBPort = os.Getenv("DB_PORT")
	c.DBUser = os.Getenv("DB_USER")
	c.DBPassword = os.Getenv("DB_PASSWORD")
	c.DBName = os.Getenv("DB_NAME")
	c.DBUrl = buildDBURL(c)

	return &c, nil
}
