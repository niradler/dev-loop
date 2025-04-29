package server

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

var storage Storage

func getConfigFolderPath() string {
	home, err := os.UserHomeDir()
	configFolderPath := "~/.dev-loop"
	if err == nil {
		configFolderPath = filepath.Join(home, ".dev-loop")
	}

	if _, err := os.Stat(configFolderPath); os.IsNotExist(err) {
		if err := os.MkdirAll(configFolderPath, 0755); err != nil {
			log.Printf("Failed to create config folder: %v", err)
		}
	}
	return configFolderPath
}

func getDBPath() string {
	return filepath.Join(getConfigFolderPath(), "devloop.db")
}

// --- API Key Middleware ---
func apiKeyMiddleware() gin.HandlerFunc {
	cfg, err := LoadConfig()
	if err != nil {
		log.Printf("Failed to load config: %v", err)
		return func(c *gin.Context) {
			c.Next()
		}
	}
	apiKey := os.Getenv("DEV_LOOP_API_KEY")
	if apiKey == "" {
		apiKey = cfg.APIKey
	}
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

func StartServer() {
	// Load environment variables from .env file if present
	_ = godotenv.Load()

	var err error
	storage, err = NewSQLiteStorage(getDBPath())
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

	port := os.Getenv("DEV_LOOP_PORT")
	if port == "" {
		port = "8997"
	}
	r.Run("localhost:" + port)
}
