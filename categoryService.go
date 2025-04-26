package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func listCategoriesHandler(c *gin.Context) {
	// Use SQL-based aggregation for categories
	counts, err := storage.(*SQLiteStorage).ListCategoryCounts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}

	if counts == nil {
		counts = []CategoryCount{}
	}
	c.JSON(http.StatusOK, counts)
}
