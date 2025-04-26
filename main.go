package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

type Config struct {
	ScriptFolders        []string          `json:"scriptFolders"`
	ExtensionCommands    map[string]string `json:"extensionCommands"`
	EnvironmentVariables map[string]string `json:"environmentVariables,omitempty"`
}

var (
	storage     Storage
	configCache *Config
)

func getConfigFilePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "config.json" // fallback
	}
	return filepath.Join(home, ".dev-loop", "config.json")
}

// --- API Key Middleware ---
func apiKeyMiddleware() gin.HandlerFunc {
	apiKey := os.Getenv("DEV_LOOP_API_KEY")
	if apiKey == "" {
		return func(c *gin.Context) {
			c.Next()
		}
	}
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		if path == "/" || strings.HasPrefix(path, "/public") {
			c.Next()
			return
		}
		header := c.GetHeader("Authorization")
		if header == "Bearer "+apiKey || header == apiKey {
			c.Next()
			return
		}
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or missing API key"})
	}
}

func main() {
	// Load environment variables from .env file if present
	_ = godotenv.Load()

	var err error
	storage, err = NewSQLiteStorage("devloop.db")
	if err != nil {
		log.Fatalf("Failed to open db: %v", err)
	}

	r := gin.Default()

	// Add CORS middleware
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	r.Static("/public", "./public")
	r.GET("/", func(c *gin.Context) {
		c.File("./dist/index.html")
	})

	r.Use(apiKeyMiddleware())

	// API endpoints
	r.POST("/api/actions/scripts/load", loadScriptsHandler)
	r.POST("/api/actions/exec/scripts/:id", execScriptHandler)

	r.GET("/api/scripts", listScriptsHandler)
	r.GET("/api/scripts/:id", getScriptHandler)
	r.DELETE("/api/scripts/:id", deleteScriptHandler)
	r.PATCH("/api/scripts/:id", openScriptHandler)

	r.GET("/api/history/scripts/:id", listScriptHistoryHandler)
	r.GET("/api/history/:id", getHistoryByIDHandler)
	r.DELETE("/api/history/:id", deleteHistoryByIDHandler)

	r.GET("/api/config", getConfigHandler)
	r.POST("/api/config", updateConfigHandler)

	r.GET("/api/history/scripts/recent", recentHistoryScriptsHandler)
	r.GET("/api/categories", listCategoriesHandler)

	r.Run("localhost:8081")
}

// --- Config persistence ---

func loadConfig() (*Config, error) {
	if configCache != nil {
		return configCache, nil
	}
	f, err := os.Open(getConfigFilePath())
	if err != nil {
		if os.IsNotExist(err) {
			configCache = defaultConfig()
			return configCache, nil
		}
		return nil, err
	}
	defer f.Close()
	var cfg Config
	if err := json.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, err
	}
	if cfg.ScriptFolders == nil || len(cfg.ScriptFolders) == 0 {
		cfg.ScriptFolders = defaultConfig().ScriptFolders
	}
	if cfg.ExtensionCommands == nil || len(cfg.ExtensionCommands) == 0 {
		cfg.ExtensionCommands = defaultConfig().ExtensionCommands
	}
	configCache = &cfg
	return configCache, nil
}

func saveConfig(cfg *Config) error {
	configPath := getConfigFilePath()
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	f, err := os.Create(configPath)
	if err != nil {
		return err
	}
	defer f.Close()
	err = json.NewEncoder(f).Encode(cfg)
	if err != nil {
		return err
	}
	configCache = cfg // update in-memory config
	return nil
}

func defaultConfig() *Config {
	return &Config{
		ScriptFolders: []string{"~/.dev-loop/scripts"},
		ExtensionCommands: map[string]string{
			".py": "python",
			".sh": "sh",
			".js": "node",
			".ts": "ts-node",
		},
	}
}

func getConfigHandler(c *gin.Context) {
	cfg, err := loadConfig()
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
	if cfg.ScriptFolders == nil || len(cfg.ScriptFolders) == 0 {
		cfg.ScriptFolders = defaultConfig().ScriptFolders
	}
	if cfg.ExtensionCommands == nil || len(cfg.ExtensionCommands) == 0 {
		cfg.ExtensionCommands = defaultConfig().ExtensionCommands
	}
	if err := saveConfig(&cfg); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save config"})
		return
	}
	c.JSON(http.StatusOK, cfg)
}

// --- Utility ---

func md5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}
