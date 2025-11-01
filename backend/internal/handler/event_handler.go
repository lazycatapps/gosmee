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

// EventHandler handles HTTP requests for event management.
type EventHandler struct {
	eventService *service.EventService
	log          logger.Logger
}

// NewEventHandler creates a new event handler.
func NewEventHandler(eventService *service.EventService, log logger.Logger) *EventHandler {
	return &EventHandler{
		eventService: eventService,
		log:          log,
	}
}

// List retrieves events for a client.
// GET /api/v1/clients/:id/events
func (h *EventHandler) List(c *gin.Context) {
	clientID := c.Param("id")

	var req models.EventListRequest
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

	response, err := h.eventService.List(clientID, &req)
	if err != nil {
		h.log.Error("Failed to list events: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// Get retrieves a single event.
// GET /api/v1/clients/:id/events/:eventId
func (h *EventHandler) Get(c *gin.Context) {
	clientID := c.Param("id")
	eventID := c.Param("eventId")

	event, err := h.eventService.Get(clientID, eventID)
	if err != nil {
		h.log.Error("Failed to get event: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	c.JSON(http.StatusOK, event)
}

// Delete deletes an event.
// DELETE /api/v1/clients/:id/events/:eventId
func (h *EventHandler) Delete(c *gin.Context) {
	clientID := c.Param("id")
	eventID := c.Param("eventId")

	if err := h.eventService.Delete(clientID, eventID); err != nil {
		h.log.Error("Failed to delete event: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Event deleted successfully"})
}

// Replay replays events to the target URL.
// POST /api/v1/clients/:id/events/replay
func (h *EventHandler) Replay(c *gin.Context) {
	clientID := c.Param("id")

	var req models.EventReplayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.eventService.Replay(clientID, &req)
	if err != nil {
		h.log.Error("Failed to replay events: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}
