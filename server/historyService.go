package server

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type ExecutionHistory struct {
	ID             string         `json:"id"`
	ScriptID       string         `json:"script_id"`
	ExecutedAt     time.Time      `json:"executed_at"`
	FinishedAt     time.Time      `json:"finished_at"`
	ExecuteRequest ExecuteRequest `json:"execute_request"`
	Output         string         `json:"output"`
	ExitCode       int            `json:"exitcode"`
	Incognito      bool           `json:"incognito"`
	Command        string         `json:"command"`
}

func listScriptHistoryHandler(c *gin.Context) {
	id := c.Param("id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * limit
	history, _ := storage.ListExecutionHistory(id, offset, limit)
	c.JSON(http.StatusOK, history)
}

func getHistoryByIDHandler(c *gin.Context) {
	id := c.Param("id")
	history, err := storage.GetHistoryByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, history)
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

func recentHistoryScriptsHandler(c *gin.Context) {
	limit := 10
	if l := c.Query("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			limit = n
		}
	}
	db, ok := storage.(*SQLiteStorage)
	if !ok {
		c.JSON(500, gin.H{"error": "storage not supported"})
		return
	}
	scripts, err := db.GetRecentScriptsWithHistory(limit)
	if err != nil {
		c.JSON(500, gin.H{"error": "db error"})
		return
	}
	c.JSON(200, scripts)
}
