package tests

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/overthinkinglabs/sqd/src/services/default_config"
)

func TestInitCreatesConfigAndIgnore(t *testing.T) {
	tmpHome := t.TempDir()
	tmpCwd := t.TempDir()

	homeDir := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", homeDir)

	loader := default_config.NewLoader()
	initializer := default_config.NewInitializer(loader)
	configPath, configCreated, ignorePath, ignoreCreated, err := initializer.InitResult(tmpCwd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !configCreated {
		t.Error("expected global default_config to be created")
	}

	if !ignoreCreated {
		t.Error("expected local ignore file to be created")
	}

	if _, statErr := os.Stat(configPath); statErr != nil {
		t.Errorf("default_config file should exist: %v", statErr)
	}

	if _, statErr := os.Stat(ignorePath); statErr != nil {
		t.Errorf("ignore file should exist: %v", statErr)
	}
}

func TestInitDoesNotOverwriteExistingFiles(t *testing.T) {
	tmpHome := t.TempDir()
	tmpCwd := t.TempDir()

	homeDir := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", homeDir)

	configDir := filepath.Join(tmpHome, ".config", "sqd")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("failed to create default_config dir: %v", err)
	}

	existingConfig := filepath.Join(configDir, "config.json")
	if err := os.WriteFile(existingConfig, []byte("{}"), 0o644); err != nil {
		t.Fatalf("failed to write default_config: %v", err)
	}

	existingIgnore := filepath.Join(tmpCwd, ".sqdignore")
	if err := os.WriteFile(existingIgnore, []byte("# existing"), 0o644); err != nil {
		t.Fatalf("failed to write ignore: %v", err)
	}

	loader := default_config.NewLoader()
	initializer := default_config.NewInitializer(loader)
	configPath, configCreated, ignorePath, ignoreCreated, err := initializer.InitResult(tmpCwd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if configCreated {
		t.Error("expected existing default_config not to be overwritten")
	}

	if ignoreCreated {
		t.Error("expected existing ignore not to be overwritten")
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read default_config: %v", err)
	}

	if string(data) != "{}" {
		t.Errorf("default_config content should remain {}, got %s", string(data))
	}

	if configPath != existingConfig {
		t.Errorf("default_config path mismatch: %s vs %s", configPath, existingConfig)
	}

	if ignorePath != existingIgnore {
		t.Errorf("ignore path mismatch: %s vs %s", ignorePath, existingIgnore)
	}
}
