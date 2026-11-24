package app

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env/v2"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

type Config struct {
	DatabaseURL string
	AppSecret   string
	Addr        string
	Dev         bool
	TemplateDir string
	StaticDir   string
	// ViteURL is the Vite dev server pages load the frontend from, if set. Otherwise it's served from StaticDir.
	ViteURL string
}

// rawConfig mirrors the keys in application.yaml.
type rawConfig struct {
	AppSecret   string `koanf:"app.secret"`
	Port        int    `koanf:"server.port"`
	DatabaseURL string `koanf:"database.url"`
	LogDev      bool   `koanf:"log.development"`
	TemplateDir string `koanf:"templates.dir"`
	StaticDir   string `koanf:"static.dir"`
	ViteURL     string `koanf:"vite.url"`
}

func NewConfig() (Config, error) {
	return loadConfig(".", os.Getenv("APP_PROFILE"), os.Environ)
}

// loadConfig layers, lowest precedence first: application.yaml, application-<profile>.yaml,
// then environment variables for keys that already exist, e.g. DATABASE_URL -> database.url.
func loadConfig(dir, profile string, environ func() []string) (Config, error) {
	k := koanf.New(".")

	if err := k.Load(file.Provider(filepath.Join(dir, "application.yaml")), yaml.Parser()); err != nil {
		return Config{}, fmt.Errorf("failed loading base config: %w", err)
	}

	if profile != "" {
		path := filepath.Join(dir, "application-"+profile+".yaml")
		if err := k.Load(file.Provider(path), yaml.Parser()); err != nil {
			return Config{}, fmt.Errorf("failed loading profile %q: %w", profile, err)
		}
	}

	envProvider := env.Provider(".", env.Opt{
		EnvironFunc: environ,
		TransformFunc: func(name, value string) (string, any) {
			key := strings.ReplaceAll(strings.ToLower(name), "_", ".")
			if !k.Exists(key) {
				return "", nil
			}
			return key, value
		},
	})
	if err := k.Load(envProvider, nil); err != nil {
		return Config{}, fmt.Errorf("failed loading environment: %w", err)
	}

	var raw rawConfig
	if err := k.UnmarshalWithConf("", &raw, koanf.UnmarshalConf{FlatPaths: true}); err != nil {
		return Config{}, fmt.Errorf("failed parsing config: %w", err)
	}

	if raw.AppSecret == "" {
		return Config{}, errors.New("app.secret (APP_SECRET) must be set")
	}
	if raw.DatabaseURL == "" {
		return Config{}, errors.New("database.url (DATABASE_URL) must be set")
	}

	return Config{
		DatabaseURL: raw.DatabaseURL,
		AppSecret:   raw.AppSecret,
		Addr:        ":" + strconv.Itoa(raw.Port),
		Dev:         raw.LogDev,
		TemplateDir: raw.TemplateDir,
		StaticDir:   raw.StaticDir,
		ViteURL:     raw.ViteURL,
	}, nil
}
