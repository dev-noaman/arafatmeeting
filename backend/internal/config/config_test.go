package config

import (
	"os"
	"testing"
)

func TestGetEnv_DefaultValue(t *testing.T) {
	key := "TEST_CONFIG_NONEXISTENT_KEY_" + t.Name()
	os.Unsetenv(key)
	result := getEnv(key, "default-val")
	if result != "default-val" {
		t.Errorf("expected 'default-val', got '%s'", result)
	}
}

func TestGetEnv_EnvValue(t *testing.T) {
	key := "TEST_CONFIG_EXISTING_KEY_" + t.Name()
	os.Setenv(key, "custom-value")
	defer os.Unsetenv(key)
	result := getEnv(key, "default-val")
	if result != "custom-value" {
		t.Errorf("expected 'custom-value', got '%s'", result)
	}
}

func TestGetEnv_NoDefault(t *testing.T) {
	key := "TEST_CONFIG_NO_DEFAULT_" + t.Name()
	os.Unsetenv(key)
	result := getEnv(key)
	if result != "" {
		t.Errorf("expected empty string, got '%s'", result)
	}
}

func TestGetEnvAsInt_DefaultValue(t *testing.T) {
	key := "TEST_CONFIG_INT_" + t.Name()
	os.Unsetenv(key)
	result := getEnvAsInt(key, 42)
	if result != 42 {
		t.Errorf("expected 42, got %d", result)
	}
}

func TestGetEnvAsInt_ValidValue(t *testing.T) {
	key := "TEST_CONFIG_INT_" + t.Name()
	os.Setenv(key, "100")
	defer os.Unsetenv(key)
	result := getEnvAsInt(key, 42)
	if result != 100 {
		t.Errorf("expected 100, got %d", result)
	}
}

func TestGetEnvAsInt_InvalidValue(t *testing.T) {
	key := "TEST_CONFIG_INT_INVALID_" + t.Name()
	os.Setenv(key, "not-a-number")
	defer os.Unsetenv(key)
	result := getEnvAsInt(key, 42)
	if result != 42 {
		t.Errorf("expected default 42 for invalid value, got %d", result)
	}
}

func TestGetEnvAsInt_ZeroValue(t *testing.T) {
	key := "TEST_CONFIG_INT_ZERO_" + t.Name()
	os.Setenv(key, "0")
	defer os.Unsetenv(key)
	result := getEnvAsInt(key, 42)
	if result != 0 {
		t.Errorf("expected 0, got %d", result)
	}
}

func TestDatabaseConfigDSN_UsesURLWhenSet(t *testing.T) {
	cfg := DatabaseConfig{
		URL:      "postgresql://user:pass@host:5432/db",
		Host:     "localhost",
		Port:     "5432",
		User:     "postgres",
		Password: "root",
		DBName:   "mini_meeting",
	}
	result := cfg.DSN()
	if result != "postgresql://user:pass@host:5432/db" {
		t.Errorf("expected URL to be used, got '%s'", result)
	}
}

func TestDatabaseConfigDSN_BuildsFromParts(t *testing.T) {
	cfg := DatabaseConfig{
		URL:      "",
		Host:     "myhost",
		Port:     "5433",
		User:     "myuser",
		Password: "mypass",
		DBName:   "mydb",
	}
	result := cfg.DSN()
	expected := "host=myhost port=5433 user=myuser password=mypass dbname=mydb sslmode=disable"
	if result != expected {
		t.Errorf("expected '%s', got '%s'", expected, result)
	}
}

func TestDatabaseConfigIsCloud_WithURL(t *testing.T) {
	cfg := DatabaseConfig{URL: "postgresql://user:pass@host:5432/db"}
	if !cfg.IsCloud() {
		t.Error("expected IsCloud() true when URL is set")
	}
}

func TestDatabaseConfigIsCloud_NoURL(t *testing.T) {
	cfg := DatabaseConfig{URL: ""}
	if cfg.IsCloud() {
		t.Error("expected IsCloud() false when URL is empty")
	}
}
