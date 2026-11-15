package config

import (
	"errors"
	"fmt"
	"io/fs"
	"net/url"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost          string
	DBPort          string
	DBUser          string
	DBPassword      string
	DBName          string
	DBUrl           string
	JWTSecret       string
	Env             Environment
	ResendAPIKey    string
	EmailFrom       string
	FrontendBaseURL string
	EmailLogOnly    bool
}

type Environment string

const defaultFrontendBaseURL = "http://localhost:5173"

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
	// A missing .env is fine (deployed environments set vars directly);
	// requireEnv below catches anything actually missing.
	err := godotenv.Load()
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
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

	// Local-dev escape hatch: log verification links instead of sending email.
	// Must be opted into explicitly, and is refused in production.
	c.EmailLogOnly = os.Getenv("EMAIL_LOG_ONLY") == "true"
	if c.EmailLogOnly && c.Env.IsProduction() {
		return nil, fmt.Errorf("EMAIL_LOG_ONLY cannot be enabled when APP_ENV is %q", Production)
	}

	// Resend credentials are only needed when we actually send email.
	if !c.EmailLogOnly {
		c.ResendAPIKey, err = requireEnv("RESEND_API_KEY")
		if err != nil {
			return nil, err
		}

		c.EmailFrom, err = requireEnv("EMAIL_FROM")
		if err != nil {
			return nil, err
		}
	}

	if v := os.Getenv("FRONTEND_BASE_URL"); v != "" {
		c.FrontendBaseURL = v
	} else if c.Env.IsProduction() {
		// A localhost default would put dead links in production emails.
		return nil, fmt.Errorf("FRONTEND_BASE_URL is required in %s", Production)
	} else {
		c.FrontendBaseURL = defaultFrontendBaseURL
	}

	return &c, nil
}
