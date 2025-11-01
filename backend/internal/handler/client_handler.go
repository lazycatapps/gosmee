// Copyright (c) 2025 Lazycat Apps
// Licensed under the MIT License. See LICENSE file in the project root for details.

package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lazycatapps/gosmee/backend/internal/models"
	"github.com/lazycatapps/gosmee/backend/internal/pkg/logger"
	"github.com/lazycatapps/gosmee/backend/internal/service"
)

// ClientHandler handles HTTP requests for client management.
type ClientHandler struct {
	clientService *service.ClientService
	quotaService  *service.QuotaService
	log           logger.Logger
}

// NewClientHandler creates a new client handler.
func NewClientHandler(
	clientService *service.ClientService,
	quotaService *service.QuotaService,
	log logger.Logger,
) *ClientHandler {
	return &ClientHandler{
		clientService: clientService,
		quotaService:  quotaService,
		log:           log,
	}
}

// Create creates a new client instance.
// POST /api/v1/clients
func (h *ClientHandler) Create(c *gin.Context) {
	var req models.ClientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from context (set by auth middleware)
	userID := getUserID(c)

	// Create client
	client, err := h.clientService.Create(userID, &req)
	if err != nil {
		h.log.Error("Failed to create client: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, client)
}

// List retrieves all clients for the current user.
// GET /api/v1/clients
func (h *ClientHandler) List(c *gin.Context) {
	var req models.ClientListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set defaults
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 20
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	userID := getUserID(c)

	response, err := h.clientService.List(userID, &req)
	if err != nil {
		h.log.Error("Failed to list clients: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// Get retrieves a single client by ID.
// GET /api/v1/clients/:id
func (h *ClientHandler) Get(c *gin.Context) {
	clientID := c.Param("id")

	client, err := h.clientService.Get(clientID)
	if err != nil {
		h.log.Error("Failed to get client: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Client not found"})
		return
	}

	// TODO: Check if user owns this client

	c.JSON(http.StatusOK, client)
}

// Update updates a client instance.
// PUT /api/v1/clients/:id
func (h *ClientHandler) Update(c *gin.Context) {
	clientID := c.Param("id")

	var req models.ClientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	client, err := h.clientService.Update(clientID, &req)
	if err != nil {
		h.log.Error("Failed to update client: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, client)
}

// Delete deletes a client instance.
// DELETE /api/v1/clients/:id
func (h *ClientHandler) Delete(c *gin.Context) {
	clientID := c.Param("id")

	if err := h.clientService.Delete(clientID); err != nil {
		h.log.Error("Failed to delete client: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Client deleted successfully"})
}

// Start starts a client instance.
// POST /api/v1/clients/:id/start
func (h *ClientHandler) Start(c *gin.Context) {
	clientID := c.Param("id")

	if err := h.clientService.Start(clientID); err != nil {
		h.log.Error("Failed to start client: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Client started successfully"})
}

// Stop stops a client instance.
// POST /api/v1/clients/:id/stop
func (h *ClientHandler) Stop(c *gin.Context) {
	clientID := c.Param("id")

	if err := h.clientService.Stop(clientID); err != nil {
		h.log.Error("Failed to stop client: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Client stopped successfully"})
}

// Restart restarts a client instance.
// POST /api/v1/clients/:id/restart
func (h *ClientHandler) Restart(c *gin.Context) {
	clientID := c.Param("id")

	if err := h.clientService.Restart(clientID); err != nil {
		h.log.Error("Failed to restart client: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Client restarted successfully"})
}

// BatchStart starts multiple clients.
// POST /api/v1/clients/batch/start
func (h *ClientHandler) BatchStart(c *gin.Context) {
	var req models.ClientBatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !req.All && len(req.ClientIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "clientIds cannot be empty"})
		return
	}

	userID := getUserID(c)

	response, err := h.clientService.BatchStart(userID, &req)
	if err != nil {
		h.log.Error("Failed to batch start clients: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// BatchStop stops multiple clients.
// POST /api/v1/clients/batch/stop
func (h *ClientHandler) BatchStop(c *gin.Context) {
	var req models.ClientBatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !req.All && len(req.ClientIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "clientIds cannot be empty"})
		return
	}

	userID := getUserID(c)

	response, err := h.clientService.BatchStop(userID, &req)
	if err != nil {
		h.log.Error("Failed to batch stop clients: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetStats retrieves statistics for a client.
// GET /api/v1/clients/:id/stats
func (h *ClientHandler) GetStats(c *gin.Context) {
	clientID := c.Param("id")

	stats, err := h.clientService.GetStats(clientID)
	if err != nil {
		h.log.Error("Failed to get client stats: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// getUserID extracts user ID from context (set by auth middleware).
func getUserID(c *gin.Context) string {
	userID, exists := c.Get("userID")
	if !exists {
		return "default" // Default user when OIDC is disabled
	}
	return userID.(string)
}
