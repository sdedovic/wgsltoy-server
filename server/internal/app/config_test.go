package app

import (
	"testing"
)

func environ(vars ...string) func() []string {
	return func() []string { return vars }
}

func TestLoadConfig_DevProfile(t *testing.T) {
	config, err := loadConfig("../..", "dev", environ())
	if err != nil {
		t.Fatal(err)
	}

	if config.DatabaseURL != "postgres://postgres:postgres@localhost:5432/default?sslmode=disable" {
		t.Errorf("unexpected database url %q", config.DatabaseURL)
	}
	if !config.Dev {
		t.Error("expected development logging")
	}
	if config.Addr != ":8080" {
		t.Errorf("expected default addr, got %q", config.Addr)
	}
}

func TestLoadConfig_EnvOverridesProfile(t *testing.T) {
	config, err := loadConfig("../..", "dev", environ(
		"DATABASE_URL=postgres://prod/db",
		"SERVER_PORT=9090",
		"LOG_DEVELOPMENT=false",
		"PATH=/usr/bin",
	))
	if err != nil {
		t.Fatal(err)
	}

	if config.DatabaseURL != "postgres://prod/db" {
		t.Errorf("expected env override, got %q", config.DatabaseURL)
	}
	if config.Addr != ":9090" {
		t.Errorf("expected env override, got %q", config.Addr)
	}
	if config.Dev {
		t.Error("expected env to disable development logging")
	}
}

func TestLoadConfig_RequiresSecretAndDatabase(t *testing.T) {
	if _, err := loadConfig("../..", "", environ("DATABASE_URL=postgres://x")); err == nil {
		t.Error("expected missing secret error")
	}
	if _, err := loadConfig("../..", "", environ("APP_SECRET=s")); err == nil {
		t.Error("expected missing database error")
	}
}

func TestLoadConfig_UnknownProfile(t *testing.T) {
	if _, err := loadConfig("../..", "nope", environ()); err == nil {
		t.Error("expected error for missing profile file")
	}
}

func TestLoadConfig_InviteCodeFromEnv(t *testing.T) {
	config, err := loadConfig("../..", "dev", environ("INVITE_CODE=letmein"))
	if err != nil {
		t.Fatal(err)
	}
	if config.InviteCode != "letmein" {
		t.Errorf("expected invite code from env, got %q", config.InviteCode)
	}
}
