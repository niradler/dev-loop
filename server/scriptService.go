package server

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
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
	"github.com/google/uuid"
)

// --- Script Types ---
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
	Category    string   `json:"category"`
	Tags        []string `json:"tags"`
	Inputs      []Input  `json:"inputs"`
	Path        string   `json:"path"`
}

type ExecuteRequest struct {
	Args    []string          `json:"args"`
	Env     map[string]string `json:"env"`
	Command string            `json:"command"`
	Backoff int               `json:"backoff"` // milliseconds
	Repeat  int               `json:"repeat"`  // number of times to repeat execution
	Retry   int               `json:"retry"`   // number of retries on failure
}

func loadScriptsHandler(c *gin.Context) {
	var req struct {
		Folders []string `json:"folders"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || len(req.Folders) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "folders required"})
		return
	}

	cfg, err := LoadConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load config"})
		return
	}
	folderSet := make(map[string]struct{})
	for _, f := range req.Folders {
		folderSet[f] = struct{}{}
	}
	for _, f := range cfg.ScriptFolders {
		folderSet[f] = struct{}{}
	}
	var allFolders []string
	for f := range folderSet {
		allFolders = append(allFolders, f)
	}
	storage.ClearScripts()
	count := 0
	for _, folder := range allFolders {
		err := filepath.Walk(folder, func(path string, info fs.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return nil
			}
			ext := filepath.Ext(path)
			if _, ok := cfg.ExtensionCommands[ext]; !ok {
				return nil
			}
			content, err := os.ReadFile(path)
			if err != nil {
				return nil
			}
			script, err := parseScript(path, string(content))
			if err != nil {
				log.Printf("parseScript err: %v", err)
				return nil
			}
			if script.Category == "" {
				script.Category = "uncategorized"
			}

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

func parseScript(path string, content string) (*Script, error) {
	script := &Script{Path: path, ID: md5Hash(path)}
	scanner := bufio.NewScanner(strings.NewReader(content))
	var (
		inInputsBlock bool
		inputsLines   []string
		commentPrefix string
	)

	// Determine comment prefix based on file extension
	switch filepath.Ext(path) {
	case ".py":
		commentPrefix = "# @"
	case ".js", ".ts", ".go", ".zx":
		commentPrefix = "// @"
	case ".sh", ".bash", ".zsh":
		commentPrefix = "# @"
	default:
		commentPrefix = "# @" // Default to bash-style comments
	}

	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, commentPrefix) && !inInputsBlock {
			continue
		}
		if inInputsBlock {
			trimmed := strings.TrimSpace(strings.TrimPrefix(line, strings.TrimSuffix(commentPrefix, "@")))
			inputsLines = append(inputsLines, trimmed)
			if strings.Contains(trimmed, "]") {
				inInputsBlock = false
				inputsStr := strings.Join(inputsLines, "\n")
				inputsStr = strings.TrimSpace(inputsStr)
				json.Unmarshal([]byte(inputsStr), &script.Inputs)
			}
			continue
		}
		line = strings.TrimPrefix(line, commentPrefix)
		if strings.HasPrefix(line, "name:") {
			script.Name = strings.TrimSpace(strings.TrimPrefix(line, "name:"))
		} else if strings.HasPrefix(line, "description:") {
			script.Description = strings.TrimSpace(strings.TrimPrefix(line, "description:"))
		} else if strings.HasPrefix(line, "author:") {
			script.Author = strings.TrimSpace(strings.TrimPrefix(line, "author:"))
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
		script.Name = filepath.Base(path)
	}

	return script, nil
}

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

	// Set default values if not provided
	if req.Backoff == 0 {
		req.Backoff = 500 // default 500ms backoff
	}
	if req.Repeat == 0 {
		req.Repeat = 1 // default to 1 execution
	}
	if req.Retry < 0 {
		req.Retry = 0 // ensure retry is not negative
	}

	log.Printf("execScriptHandler: user request: %+v", req)

	args := append([]string{script.Path}, req.Args...)

	// Load config and add environment variables from config if present
	cfg, err := LoadConfig()
	if err == nil && cfg != nil && cfg.EnvironmentVariables != nil {
		for k, v := range cfg.EnvironmentVariables {
			req.Env[k] = v
		}
	}

	if req.Command == "" {
		req.Command = cfg.ExtensionCommands[filepath.Ext(script.Path)]
	}

	log.Printf("execScriptHandler: running command: %s %s", req.Command, strings.Join(args, " "))

	// Split the command into parts if it contains spaces
	commandParts := strings.Fields(req.Command)

	// Function to execute a single run with retries
	executeWithRetry := func() (string, int, error) {
		var lastErr error
		var output string
		var exitCode int

		for attempt := 0; attempt <= req.Retry; attempt++ {
			cmd := exec.Command(commandParts[0], append(commandParts[1:], args...)...)
			cmd.Env = os.Environ()

			for k, v := range req.Env {
				cmd.Env = append(cmd.Env, k+"="+v)
			}

			var stdoutBuf, stderrBuf strings.Builder
			cmd.Stdout = &stdoutBuf
			cmd.Stderr = &stderrBuf

			if err := cmd.Start(); err != nil {
				lastErr = err
				if attempt < req.Retry {
					time.Sleep(time.Duration(req.Backoff) * time.Millisecond)
					continue
				}
				return "", -1, err
			}

			waitErr := cmd.Wait()
			output = stdoutBuf.String() + stderrBuf.String()

			if waitErr != nil {
				if exitErr, ok := waitErr.(*exec.ExitError); ok {
					exitCode = exitErr.ExitCode()
				} else {
					exitCode = -1
				}
				lastErr = waitErr
				if attempt < req.Retry {
					time.Sleep(time.Duration(req.Backoff) * time.Millisecond)
					continue
				}
				return output, exitCode, waitErr
			}

			if cmd.ProcessState != nil {
				exitCode = cmd.ProcessState.ExitCode()
			}
			return output, exitCode, nil
		}
		return output, exitCode, lastErr
	}

	// Execute the script multiple times if requested
	var allOutputs []string
	var allExitCodes []int

	for i := 0; i < req.Repeat; i++ {
		if i > 0 {
			time.Sleep(time.Duration(req.Backoff) * time.Millisecond)
		}

		output, exitCode, err := executeWithRetry()
		allOutputs = append(allOutputs, output)
		allExitCodes = append(allExitCodes, exitCode)

		if err != nil && i < req.Repeat-1 {
			time.Sleep(time.Duration(req.Backoff) * time.Millisecond)
		}
	}

	// Combine all outputs
	combinedOutput := strings.Join(allOutputs, "\n")

	// Stream the output to the client
	c.String(http.StatusOK, combinedOutput)

	// Save execution history
	incognito := c.Query("incognito") == "true"

	if incognito {
		maskedArgs := make([]string, len(req.Args))
		for i := range req.Args {
			maskedArgs[i] = "*****"
		}
		maskedEnv := make(map[string]string)
		for k := range req.Env {
			maskedEnv[k] = "*****"
		}
		req.Args = maskedArgs
		req.Env = maskedEnv
		combinedOutput = "*****"
	}

	// Save the last execution's details
	storage.SaveExecutionHistory(&ExecutionHistory{
		ID:             uuid.New().String(),
		ScriptID:       id,
		ExecutedAt:     time.Now(),
		FinishedAt:     time.Now(),
		ExecuteRequest: req,
		Output:         combinedOutput,
		ExitCode:       allExitCodes[len(allExitCodes)-1],
		Incognito:      incognito,
		Command:        req.Command,
	})
}

func listScriptsHandler(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * limit

	search := strings.TrimSpace(strings.ToLower(c.DefaultQuery("search", "")))
	category := strings.TrimSpace(strings.ToLower(c.DefaultQuery("category", "")))
	tag := strings.TrimSpace(strings.ToLower(c.DefaultQuery("tag", "")))

	scripts, err := storage.ListScripts(offset, limit, search, category, tag)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}

	type ScriptMeta struct {
		ID          string   `json:"id"`
		Name        string   `json:"name"`
		Description string   `json:"description"`
		Author      string   `json:"author"`
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
	cfg, err := LoadConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load config"})
		return
	}
	go func() {
		exec.Command(cfg.Editor, script.Path).Run()
	}()
	c.JSON(http.StatusOK, gin.H{"message": "Script opened in editor"})
}

func md5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}
