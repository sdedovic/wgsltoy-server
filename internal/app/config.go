package app

import (
	"cmp"
	"errors"
	"os"
)

type Config struct {
	DatabaseURL string
	AppSecret   string
	Addr        string
	Dev         bool
	TemplateDir string
}

func NewConfig() (Config, error) {
	config := Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		AppSecret:   os.Getenv("APP_SECRET"),
		Addr:        ":" + cmp.Or(os.Getenv("PORT"), "8080"),
		Dev:         os.Getenv("APP_ENV") == "dev",
		TemplateDir: cmp.Or(os.Getenv("TEMPLATE_DIR"), "templates"),
	}

	if config.AppSecret == "" {
		return Config{}, errors.New("APP_SECRET must be set")
	}

	return config, nil
}
