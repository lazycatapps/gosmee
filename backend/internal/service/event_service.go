// Copyright (c) 2025 Lazycat Apps
// Licensed under the MIT License. See LICENSE file in the project root for details.

package service

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/lazycatapps/gosmee/backend/internal/models"
	"github.com/lazycatapps/gosmee/backend/internal/pkg/logger"
	"github.com/lazycatapps/gosmee/backend/internal/repository"
)

// EventService manages webhook events.
type EventService struct {
	eventRepo  repository.EventRepository
	clientRepo repository.ClientRepository
	log        logger.Logger
}

// NewEventService creates a new event service.
func NewEventService(
	eventRepo repository.EventRepository,
	clientRepo repository.ClientRepository,
	log logger.Logger,
) *EventService {
	return &EventService{
		eventRepo:  eventRepo,
		clientRepo: clientRepo,
		log:        log,
	}
}

// List retrieves events for a client with filters and pagination.
func (s *EventService) List(clientID string, req *models.EventListRequest) (*models.EventListResponse, error) {
	return s.eventRepo.GetByClientID(clientID, req)
}

// Get retrieves a single event.
func (s *EventService) Get(clientID, eventID string) (*models.Event, error) {
	return s.eventRepo.Get(clientID, eventID)
}

// Delete deletes an event.
func (s *EventService) Delete(clientID, eventID string) error {
	if err := s.eventRepo.Delete(clientID, eventID); err != nil {
		return fmt.Errorf("failed to delete event: %w", err)
	}

	s.log.Info("Deleted event: %s (client: %s)", eventID, clientID)
	return nil
}

// Replay replays events to the target URL.
func (s *EventService) Replay(clientID string, req *models.EventReplayRequest) (*models.EventReplayResponse, error) {
	// Get client to get target URL
	client, err := s.clientRepo.Get(clientID)
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	response := &models.EventReplayResponse{
		Total:   len(req.EventIDs),
		Results: make([]*models.EventReplayResult, 0, len(req.EventIDs)),
	}

	// Replay each event
	for _, eventID := range req.EventIDs {
		result := s.replayEvent(client, eventID)
		response.Results = append(response.Results, result)

		if result.Success {
			response.Successful++
		} else {
			response.Failed++
		}
	}

	s.log.Info("Replayed %d events for client %s (%d successful, %d failed)",
		response.Total, clientID, response.Successful, response.Failed)

	return response, nil
}

// replayEvent replays a single event.
func (s *EventService) replayEvent(client *models.Client, eventID string) *models.EventReplayResult {
	result := &models.EventReplayResult{
		EventID: eventID,
	}

	// Get event
	event, err := s.eventRepo.Get(client.ID, eventID)
	if err != nil {
		result.Success = false
		result.ErrorMessage = fmt.Sprintf("failed to get event: %v", err)
		return result
	}

	// Log payload for debugging
	s.log.Info("Replaying event %s: payload length=%d bytes", eventID, len(event.Payload))
	if len(event.Payload) < 500 {
		s.log.Debug("Payload content: %s", event.Payload)
	}

	// Prepare HTTP request
	req, err := http.NewRequest("POST", client.TargetURL, bytes.NewBufferString(event.Payload))
	if err != nil {
		result.Success = false
		result.ErrorMessage = fmt.Sprintf("failed to create request: %v", err)
		return result
	}

	// Set default Content-Type if not present in original headers
	hasContentType := false
	for key := range event.Headers {
		if strings.EqualFold(key, "Content-Type") {
			hasContentType = true
			break
		}
	}
	if !hasContentType {
		req.Header.Set("Content-Type", "application/json")
		s.log.Debug("Set default Content-Type: application/json")
	}

	// Copy headers from original event
	for key, value := range event.Headers {
		req.Header.Set(key, value)
	}
	s.log.Debug("Replay request headers: %d headers copied from original event", len(event.Headers))

	// Log final headers for debugging
	s.log.Debug("Final request headers: Content-Type=%s, Total=%d",
		req.Header.Get("Content-Type"), len(req.Header))
	for key, values := range req.Header {
		s.log.Debug("  %s: %s", key, strings.Join(values, ", "))
	}

	// Send request
	httpClient := &http.Client{
		Timeout: time.Duration(client.TargetTimeout) * time.Second,
	}

	s.log.Info("Sending replay request to %s", client.TargetURL)
	startTime := time.Now()
	resp, err := httpClient.Do(req)
	latency := time.Since(startTime)

	if err != nil {
		result.Success = false
		result.ErrorMessage = fmt.Sprintf("failed to send request: %v", err)
		s.log.Error("Replay request failed: %v", err)
		return result
	}
	defer resp.Body.Close()

	// Read response
	body, _ := io.ReadAll(resp.Body)

	result.Success = resp.StatusCode >= 200 && resp.StatusCode < 300
	result.StatusCode = resp.StatusCode
	result.LatencyMs = int(latency.Milliseconds())

	s.log.Info("Replay response: status=%d, latency=%dms, body_length=%d bytes",
		resp.StatusCode, result.LatencyMs, len(body))
	if len(body) < 500 {
		s.log.Debug("Response body: %s", string(body))
	}

	if !result.Success {
		result.ErrorMessage = fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	return result
}

// CleanupOldEvents removes events older than retention period.
func (s *EventService) CleanupOldEvents(clientID string, retentionDays int) error {
	if err := s.eventRepo.CleanupOldEvents(clientID, retentionDays); err != nil {
		return fmt.Errorf("failed to cleanup old events: %w", err)
	}

	s.log.Info("Cleaned up old events for client: %s (retention: %d days)", clientID, retentionDays)
	return nil
}
