package tests

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/overthinkinglabs/sqd/src/services/default_config"
)

func writeConfigFile(t *testing.T, configDir string, content string) {
	t.Helper()
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	configFile := filepath.Join(configDir, "config.json")
	if err := os.WriteFile(configFile, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}
}

func TestLoadConfigReturnsDefaultsWhenMissing(t *testing.T) {
	tmpDir := t.TempDir()

	homeDir := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", homeDir)

	loader := default_config.NewLoader()
	loadedConfig, err := loader.LoadConfig()
	if err != nil {
		t.Fatalf("expected no error for missing config, got %v", err)
	}

	if loadedConfig.Output.Color != "blue" {
		t.Errorf("expected default color blue, got %s", loadedConfig.Output.Color)
	}

	if !loadedConfig.Output.ShowStats {
		t.Error("expected show_stats true by default")
	}
}

func TestLoadConfigReadsGlobalFile(t *testing.T) {
	tmpDir := t.TempDir()

	homeDir := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", homeDir)

	content := `{
		"output": {"color": "green", "show_stats": false},
		"from_aliases": {"md": "*.md"}
	}`
	writeConfigFile(t, filepath.Join(tmpDir, ".config", "sqd"), content)

	loader := default_config.NewLoader()
	loadedConfig, err := loader.LoadConfig()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if loadedConfig.Output.Color != "green" {
		t.Errorf("expected color green, got %s", loadedConfig.Output.Color)
	}

	if loadedConfig.Output.ShowStats {
		t.Error("expected show_stats false")
	}

	if loadedConfig.FromAliases["md"] != "*.md" {
		t.Errorf("expected md alias *.md, got %s", loadedConfig.FromAliases["md"])
	}
}

func TestLoadConfigReturnsErrorForInvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()

	homeDir := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", homeDir)

	writeConfigFile(t, filepath.Join(tmpDir, ".config", "sqd"), "not json")

	loader := default_config.NewLoader()
	_, err := loader.LoadConfig()
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}
