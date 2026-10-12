package config

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
)

type Config struct {
	Addr         string
	DataDir      string
	Password     string
	CookieSecure *bool
	PushContact  string
	GoCache      string
}

func (c Config) DBPath() string {
	return filepath.Join(c.DataDir, "gonteractive.db")
}

func (c Config) GoCacheDir() string {
	if c.GoCache != "" {
		return c.GoCache
	}
	return filepath.Join(c.DataDir, "gocache")
}

func Load() (Config, error) {
	c := Config{
		Addr:        env("APP_ADDR", ":8080"),
		DataDir:     env("APP_DATA_DIR", "data"),
		Password:    os.Getenv("APP_PASSWORD"),
		PushContact: env("APP_PUSH_CONTACT", "gonteractive@example.com"),
		GoCache:     os.Getenv("APP_GO_CACHE"),
	}
	if c.Password == "" {
		return c, errors.New("APP_PASSWORD is required")
	}

	if v := env("APP_COOKIE_SECURE", "auto"); v != "auto" {
		secure, err := strconv.ParseBool(v)
		if err != nil {
			return c, errors.New("APP_COOKIE_SECURE must be auto, true or false")
		}
		c.CookieSecure = &secure
	}
	return c, nil
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
