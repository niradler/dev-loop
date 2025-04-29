package server

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

type Config struct {
	ScriptFolders        []string          `json:"scriptFolders"`
	ExtensionCommands    map[string]string `json:"extensionCommands"`
	EnvironmentVariables map[string]string `json:"environmentVariables,omitempty"`
	APIKey               string            `json:"apiKey,omitempty"`
	Editor               string            `json:"editor,omitempty"`
}

var configCache *Config

func getConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "./config.json"
	}
	return filepath.Join(homeDir, ".dev-loop", "config.json")
}

func LoadConfig() (*Config, error) {
	if configCache != nil {
		return configCache, nil
	}

	configPath := getConfigPath()

	// Create config directory if it doesn't exist
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, err
	}

	// Read existing config or create new one
	config := defaultConfig()
	if _, err := os.Stat(configPath); err == nil {
		data, err := os.ReadFile(configPath)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(data, &config); err != nil {
			return nil, err
		}
	} else {
		// Create new config file with defaults
		data, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			return nil, err
		}
		if err := os.WriteFile(configPath, data, 0644); err != nil {
			return nil, err
		}
	}

	if config.Editor == "" {
		config.Editor = "code"
	}

	configCache = config
	return config, nil
}

func SaveConfig(config *Config) error {
	configPath := getConfigPath()
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return err
	}
	configCache = config
	return nil
}

func defaultConfig() *Config {
	return &Config{
		ScriptFolders: []string{"~/.dev-loop/scripts"},
		ExtensionCommands: map[string]string{
			".py":   "python",
			".js":   "node",
			".ts":   "ts-node",
			".go":   "go run",
			".sh":   "bash",
			".bash": "bash",
			".zsh":  "zsh",
			".zx":   "zx",
		},
		EnvironmentVariables: make(map[string]string),
		APIKey:               "",
		Editor:               "code",
	}
}

func getConfigHandler(c *gin.Context) {
	cfg, err := LoadConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load config"})
		return
	}
	c.JSON(http.StatusOK, cfg)
}

func updateConfigHandler(c *gin.Context) {
	var cfg Config
	if err := c.ShouldBindJSON(&cfg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid config"})
		return
	}
	if err := SaveConfig(&cfg); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save config"})
		return
	}
	c.JSON(http.StatusOK, cfg)
}
