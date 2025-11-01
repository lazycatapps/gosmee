// Copyright (c) 2025 Lazycat Apps
// Licensed under the MIT License. See LICENSE file in the project root for details.

package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lazycatapps/gosmee/backend/internal/pkg/logger"
	"github.com/lazycatapps/gosmee/backend/internal/service"
)

// QuotaHandler handles HTTP requests for quota management.
type QuotaHandler struct {
	quotaService *service.QuotaService
	log          logger.Logger
}

// NewQuotaHandler creates a new quota handler.
func NewQuotaHandler(quotaService *service.QuotaService, log logger.Logger) *QuotaHandler {
	return &QuotaHandler{
		quotaService: quotaService,
		log:          log,
	}
}

// GetQuota retrieves quota information for the current user.
// GET /api/v1/quota
func (h *QuotaHandler) GetQuota(c *gin.Context) {
	userID := getUserID(c)

	quota, err := h.quotaService.GetQuota(userID)
	if err != nil {
		h.log.Error("Failed to get quota: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Add warning if needed
	warning, _ := h.quotaService.GetStorageWarning(userID)

	response := gin.H{
		"quota": quota,
	}

	if warning != "" {
		response["warning"] = warning
	}

	c.JSON(http.StatusOK, response)
}
