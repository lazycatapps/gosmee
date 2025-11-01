// Copyright (c) 2025 Lazycat Apps
// Licensed under the MIT License. See LICENSE file in the project root for details.

package handler

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/lazycatapps/gosmee/backend/internal/pkg/logger"
	"github.com/lazycatapps/gosmee/backend/internal/service"
)

// LogHandler handles HTTP requests for log management.
type LogHandler struct {
	logService     *service.LogService
	processService *service.ProcessService
	log            logger.Logger
}

// NewLogHandler creates a new log handler.
func NewLogHandler(
	logService *service.LogService,
	processService *service.ProcessService,
	log logger.Logger,
) *LogHandler {
	return &LogHandler{
		logService:     logService,
		processService: processService,
		log:            log,
	}
}

// GetLogs retrieves historical logs for a client.
// GET /api/v1/clients/:id/logs
func (h *LogHandler) GetLogs(c *gin.Context) {
	clientID := c.Param("id")
	date := c.DefaultQuery("date", "")
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("pageSize", "100")
	search := c.Query("search")

	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 1000 {
		pageSize = 100
	}

	userID := getUserID(c)

	var logs []string
	var total int
	var err error

	if date == "" {
		// Get today's logs
		logs, total, err = h.logService.GetTodayLogs(userID, clientID, page, pageSize, search)
	} else {
		// Get logs for specific date
		logs, total, err = h.logService.GetLogs(userID, clientID, date, page, pageSize, search)
	}

	if err != nil {
		h.log.Error("Failed to get logs: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
		"logs":     logs,
	})
}

// StreamLogs streams real-time logs via SSE.
// GET /api/v1/clients/:id/logs/stream
func (h *LogHandler) StreamLogs(c *gin.Context) {
	clientID := c.Param("id")

	// Get log stream channel
	logChan, err := h.logService.StreamLogs(clientID, h.processService)
	if err != nil {
		h.log.Error("Failed to start log stream: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Transfer-Encoding", "chunked")

	// Stream logs
	c.Stream(func(w io.Writer) bool {
		select {
		case log, ok := <-logChan:
			if !ok {
				return false
			}
			c.SSEvent("log", log)
			return true
		case <-c.Request.Context().Done():
			// Client disconnected
			return false
		}
	})
}

// DownloadLog downloads a log file.
// GET /api/v1/clients/:id/logs/download
func (h *LogHandler) DownloadLog(c *gin.Context) {
	clientID := c.Param("id")
	date := c.Query("date")

	if date == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "date parameter is required"})
		return
	}

	userID := getUserID(c)

	data, err := h.logService.DownloadLog(userID, clientID, date)
	if err != nil {
		h.log.Error("Failed to download log: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	filename := fmt.Sprintf("gosmee-%s-%s.log", clientID, date)

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Type", "text/plain")
	c.Data(http.StatusOK, "text/plain", data)
}
