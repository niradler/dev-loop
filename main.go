package main

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type ExecutionHistory struct {
	ScriptID       string         `json:"script_id"`
	ExecutedAt     time.Time      `json:"executed_at"`
	FinishedAt     time.Time      `json:"finished_at"`
	ExecuteRequest ExecuteRequest `json:"execute_request"`
	Output         string         `json:"output"`
	ExitCode       int            `json:"exitcode"`
}

type Input struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Type        string      `json:"type"`
	Required    bool        `json:"required"`
	Default     interface{} `json:"default"`
}

type Script struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Author      string   `json:"author"`
	Version     string   `json:"version"`
	Category    string   `json:"category"`
	Tags        []string `json:"tags"`
	Inputs      []Input  `json:"inputs"`
	Path        string   `json:"path"`
}

type ExecuteRequest struct {
	Args    []string          `json:"args"`
	Env     map[string]string `json:"env"`
	Command string            `json:"command"`
}

var (
	storage Storage
)

func main() {
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

	// Serve static files
	r.Static("/public", "./public")
	r.GET("/", func(c *gin.Context) {
		c.File("./public/index.html")
	})

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

	r.Run("localhost:8081")
}

// --- Script loading ---

func loadScriptsHandler(c *gin.Context) {
	var req struct {
		Folders []string `json:"folders"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || len(req.Folders) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "folders required"})
		return
	}
	count := 0
	for _, folder := range req.Folders {
		err := filepath.Walk(folder, func(path string, info fs.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return nil
			}
			content, err := os.ReadFile(path)
			if err != nil {
				return nil
			}
			script, err := parseScript(string(content))
			if err != nil {
				return nil
			}
			script.Path = path
			script.ID = md5Hash(path)
			storage.SaveScript(script)
			count++
			return nil
		})
		if err != nil {
			continue
		}
	}
	c.JSON(http.StatusOK, gin.H{"message": "Scripts loaded successfully", "count": count})
}

func parseScript(content string) (*Script, error) {
	script := &Script{}
	scanner := bufio.NewScanner(strings.NewReader(content))
	var (
		inInputsBlock bool
		inputsLines   []string
	)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "# @") && !inInputsBlock {
			continue
		}
		if inInputsBlock {
			trimmed := strings.TrimSpace(strings.TrimPrefix(line, "#"))
			inputsLines = append(inputsLines, trimmed)
			if strings.Contains(trimmed, "]") {
				inInputsBlock = false
				// Join and unmarshal
				inputsStr := strings.Join(inputsLines, "\n")
				inputsStr = strings.TrimSpace(inputsStr)
				json.Unmarshal([]byte(inputsStr), &script.Inputs)
			}
			continue
		}
		line = strings.TrimPrefix(line, "# @")
		if strings.HasPrefix(line, "name:") {
			script.Name = strings.TrimSpace(strings.TrimPrefix(line, "name:"))
		} else if strings.HasPrefix(line, "description:") {
			script.Description = strings.TrimSpace(strings.TrimPrefix(line, "description:"))
		} else if strings.HasPrefix(line, "author:") {
			script.Author = strings.TrimSpace(strings.TrimPrefix(line, "author:"))
		} else if strings.HasPrefix(line, "version:") {
			script.Version = strings.TrimSpace(strings.TrimPrefix(line, "version:"))
		} else if strings.HasPrefix(line, "category:") {
			script.Category = strings.TrimSpace(strings.TrimPrefix(line, "category:"))
		} else if strings.HasPrefix(line, "tags:") {
			tags := strings.TrimSpace(strings.TrimPrefix(line, "tags:"))
			json.Unmarshal([]byte(tags), &script.Tags)
		} else if strings.HasPrefix(line, "inputs:") {
			// Start of multi-line inputs block
			after := strings.TrimSpace(strings.TrimPrefix(line, "inputs:"))
			if strings.HasPrefix(after, "[") && !strings.HasSuffix(after, "]") {
				inInputsBlock = true
				inputsLines = []string{after}
			} else {
				// Single-line inputs
				json.Unmarshal([]byte(after), &script.Inputs)
			}
		}
	}
	if script.Name == "" {
		return nil, errors.New("missing name")
	}
	return script, nil
}

// --- Script execution ---

func execScriptHandler(c *gin.Context) {
	id := c.Param("id")
	script, err := storage.GetScript(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "script not found"})
		return
	}
	var req ExecuteRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if req.Command == "" {
		req.Command = "sh"
	}
	args := append([]string{script.Path}, req.Args...)
	cmd := exec.Command(req.Command, args...)
	cmd.Env = os.Environ()
	for k, v := range req.Env {
		cmd.Env = append(cmd.Env, k+"="+v)
	}
	var stdoutBuf, stderrBuf strings.Builder
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf
	executedAt := time.Now()
	if err := cmd.Start(); err != nil {
		log.Printf("Error starting command: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to start script"})
		return
	}
	waitErr := cmd.Wait()
	if waitErr != nil {
		log.Printf("Error executing command: %v", waitErr)
	}
	output := stdoutBuf.String() + stderrBuf.String() // Combine stdout and stderr
	c.String(http.StatusOK, output)                   // Stream the output to the client

	req.Args = args

	exitCode := 0
	if waitErr != nil {
		if exitErr, ok := waitErr.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = -1
		}
	} else if cmd.ProcessState != nil {
		exitCode = cmd.ProcessState.ExitCode()
	}

	if c.Query("incognito") == "true" {
		maskedArgs := []string{script.Path}
		maskedEnv := make(map[string]string)
		for k := range req.Env {
			maskedEnv[k] = "*****"
		}
		req.Args = maskedArgs
		req.Env = maskedEnv
		output = "*****"
	}

	storage.SaveExecutionHistory(&ExecutionHistory{
		ScriptID:       id,
		ExecutedAt:     executedAt,
		FinishedAt:     time.Now(),
		ExecuteRequest: req,
		Output:         output,
		ExitCode:       exitCode,
	})
}

// --- Script metadata ---

func listScriptsHandler(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * limit
	scripts, _ := storage.ListScripts(offset, limit)
	// Remove content field from response, only return metadata and path
	type ScriptMeta struct {
		ID          string   `json:"id"`
		Name        string   `json:"name"`
		Description string   `json:"description"`
		Author      string   `json:"author"`
		Version     string   `json:"version"`
		Category    string   `json:"category"`
		Tags        []string `json:"tags"`
		Inputs      []Input  `json:"inputs"`
		Path        string   `json:"path"`
	}
	var metas []ScriptMeta
	for _, s := range scripts {
		metas = append(metas, ScriptMeta{
			ID:          s.ID,
			Name:        s.Name,
			Description: s.Description,
			Author:      s.Author,
			Version:     s.Version,
			Category:    s.Category,
			Tags:        s.Tags,
			Inputs:      s.Inputs,
			Path:        s.Path,
		})
	}
	c.JSON(http.StatusOK, metas)
}

func getScriptHandler(c *gin.Context) {
	id := c.Param("id")
	script, err := storage.GetScript(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	content, err := os.ReadFile(script.Path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read script file"})
		return
	}
	type ScriptWithContent struct {
		ID          string   `json:"id"`
		Name        string   `json:"name"`
		Description string   `json:"description"`
		Author      string   `json:"author"`
		Version     string   `json:"version"`
		Category    string   `json:"category"`
		Tags        []string `json:"tags"`
		Inputs      []Input  `json:"inputs"`
		Path        string   `json:"path"`
		Content     string   `json:"content"`
	}
	resp := ScriptWithContent{
		ID:          script.ID,
		Name:        script.Name,
		Description: script.Description,
		Author:      script.Author,
		Version:     script.Version,
		Category:    script.Category,
		Tags:        script.Tags,
		Inputs:      script.Inputs,
		Path:        script.Path,
		Content:     string(content),
	}
	c.JSON(http.StatusOK, resp)
}

func deleteScriptHandler(c *gin.Context) {
	id := c.Param("id")
	script, err := storage.GetScript(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "script not found"})
		return
	}

	// Check for "rm" query parameter
	if c.Query("rm") == "true" {
		if err := os.Remove(script.Path); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove script file"})
			return
		}
	}

	err = storage.DeleteScript(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "delete failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Script deleted"})
}

func openScriptHandler(c *gin.Context) {
	id := c.Param("id")
	script, err := storage.GetScript(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "code"
	}
	go func() {
		exec.Command(editor, script.Path).Run()
	}()
	c.JSON(http.StatusOK, gin.H{"message": "Script opened in editor"})
}

// --- History ---

func listScriptHistoryHandler(c *gin.Context) {
	id := c.Param("id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * limit
	histories, _ := storage.ListExecutionHistory(id, offset, limit)
	if histories == nil {
		histories = []*ExecutionHistory{}
	}
	c.JSON(http.StatusOK, histories)
}

func getHistoryByIDHandler(c *gin.Context) {
	id := c.Param("id")
	h, err := storage.GetHistoryByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, h)
}

func deleteHistoryByIDHandler(c *gin.Context) {
	id := c.Param("id")
	err := storage.DeleteHistoryByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "delete failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "History deleted"})
}

// --- Utility ---

func md5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}
