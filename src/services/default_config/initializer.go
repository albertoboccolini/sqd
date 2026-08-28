package default_config

import (
	"os"
	"path/filepath"
)

type Initializer struct {
	loader *Loader
}

func NewInitializer(loader *Loader) *Initializer {
	return &Initializer{
		loader,
	}
}

func (initializer *Initializer) InitConfig() (string, bool, error) {
	fullPath := initializer.loader.configPath()
	configDir := filepath.Dir(fullPath)

	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return "", false, err
	}

	if _, statErr := os.Stat(fullPath); statErr == nil {
		return fullPath, false, nil
	}

	defaultConfigJSON := `{
		"output": {
			"color": "blue",
			"show_stats": true
		},
		"from_aliases": {}
	}
	`

	return fullPath, true, os.WriteFile(fullPath, []byte(defaultConfigJSON), 0o644)
}

func (initializer *Initializer) InitIgnore(cwd string) (ignorePath string, created bool, err error) {
	ignorePath = filepath.Join(cwd, ".sqdignore")

	if _, statErr := os.Stat(ignorePath); statErr == nil {
		return ignorePath, false, nil
	}
	defaultIgnoreContent := `# Files and directories ignored by sqd.
	# Use shell glob syntax; "**" matches any depth.
	# Examples:
	# *.min.js
	# vendor/
	# node_modules/**
	# .git/
	`
	return ignorePath, true, os.WriteFile(ignorePath, []byte(defaultIgnoreContent), 0o644)
}

func (initializer *Initializer) InitResult(cwd string) (configPath string, configCreated bool, ignorePath string, ignoreCreated bool, err error) {
	configPath, configCreated, err = initializer.InitConfig()
	if err != nil {
		return "", false, "", false, err
	}

	ignorePath, ignoreCreated, err = initializer.InitIgnore(cwd)
	if err != nil {
		return configPath, configCreated, "", false, err
	}

	return configPath, configCreated, ignorePath, ignoreCreated, nil
}
