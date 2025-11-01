// Copyright (c) 2025 Lazycat Apps
// Licensed under the MIT License. See LICENSE file in the project root for details.

package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lazycatapps/gosmee/backend/internal/models"
	"github.com/lazycatapps/gosmee/backend/internal/pkg/logger"
	"github.com/lazycatapps/gosmee/backend/internal/repository"
)

// ClientService manages gosmee client instances.
type ClientService struct {
	clientRepo     repository.ClientRepository
	quotaRepo      repository.QuotaRepository
	eventRepo      repository.EventRepository
	processService *ProcessService
	baseDir        string
	log            logger.Logger
}

// NewClientService creates a new client service.
func NewClientService(
	clientRepo repository.ClientRepository,
	quotaRepo repository.QuotaRepository,
	eventRepo repository.EventRepository,
	processService *ProcessService,
	baseDir string,
	log logger.Logger,
) *ClientService {
	return &ClientService{
		clientRepo:     clientRepo,
		quotaRepo:      quotaRepo,
		eventRepo:      eventRepo,
		processService: processService,
		baseDir:        baseDir,
		log:            log,
	}
}

// Create creates a new client instance.
func (s *ClientService) Create(userID string, req *models.ClientRequest) (*models.Client, error) {
	// Check quota first
	quota, err := s.quotaRepo.GetQuota(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check quota: %w", err)
	}

	if !quota.CanCreateClient() {
		return nil, fmt.Errorf("client limit reached: %d/%d", quota.ClientsCount, quota.MaxClients)
	}

	// Generate client ID
	clientID := uuid.New().String()

	// Create client model
	client := models.NewClient(
		clientID,
		userID,
		req.Name,
		req.Description,
		req.SmeeURL,
		req.TargetURL,
	)

	// Apply optional settings
	if req.TargetTimeout > 0 {
		client.TargetTimeout = req.TargetTimeout
	}
	client.HTTPie = req.HTTPie
	client.IgnoreEvents = req.IgnoreEvents
	client.NoReplay = req.NoReplay
	if req.SSEBufferSize > 0 {
		client.SSEBufferSize = req.SSEBufferSize
	}

	// Save to repository
	if err := s.clientRepo.Create(client); err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	s.log.Info("Created client: %s (user: %s, name: %s)", clientID, userID, req.Name)

	// Invalidate quota cache
	s.quotaRepo.(*repository.FileQuotaRepository).InvalidateCache(userID)

	return client, nil
}

// Get retrieves a client by ID.
func (s *ClientService) Get(clientID string) (*models.Client, error) {
	client, err := s.clientRepo.Get(clientID)
	if err != nil {
		return nil, err
	}

	// Update status from process service
	if s.processService.IsRunning(clientID) {
		client.Status = models.ClientStatusRunning
		if processInfo, err := s.processService.GetProcessInfo(clientID); err == nil {
			client.PID = processInfo.PID
			client.StartedAt = &processInfo.StartedAt
		}
	} else {
		client.Status = models.ClientStatusStopped
	}

	if err := s.populateClientLastActivity(client); err != nil {
		s.log.Error("Failed to populate last activity for client %s: %v", clientID, err)
	}

	return client, nil
}

// List retrieves clients with filters and pagination.
func (s *ClientService) List(userID string, req *models.ClientListRequest) (*models.ClientListResponse, error) {
	response, err := s.clientRepo.List(userID, req)
	if err != nil {
		return nil, err
	}

	var filteredSummaries []*models.ClientSummary
	for _, summary := range response.Clients {
		if summary == nil {
			continue
		}

		if strings.EqualFold(summary.Status, string(models.ClientStatusError)) {
			// Preserve error status to surface failed instances after restarts.
		} else if s.processService.IsRunning(summary.ID) {
			summary.Status = string(models.ClientStatusRunning)
		} else {
			summary.Status = string(models.ClientStatusStopped)
		}

		ts, err := s.eventRepo.GetLatestEventTimestamp(summary.ID)
		if err != nil {
			s.log.Error("Failed to fetch last activity for client %s: %v", summary.ID, err)
		} else {
			summary.LastActivity = ts
		}

		if req != nil && req.Status != "" && !strings.EqualFold(summary.Status, req.Status) {
			continue
		}

		filteredSummaries = append(filteredSummaries, summary)
	}

	if req != nil && req.Status != "" {
		response.Clients = filteredSummaries

		total, err := s.countClientsByStatus(userID, req.Status)
		if err != nil {
			s.log.Error("Failed to count clients for status %s: %v", req.Status, err)
		} else {
			response.Total = total
		}
	} else {
		response.Clients = filteredSummaries
	}

	return response, nil
}

// Update updates a client instance.
func (s *ClientService) Update(clientID string, req *models.ClientRequest) (*models.Client, error) {
	// Get existing client
	client, err := s.clientRepo.Get(clientID)
	if err != nil {
		return nil, err
	}

	// Check if running - must stop first
	if s.processService.IsRunning(clientID) {
		return nil, fmt.Errorf("cannot update running client - stop it first")
	}

	// Update fields
	client.Name = req.Name
	client.Description = req.Description
	client.TargetURL = req.TargetURL
	client.TargetTimeout = req.TargetTimeout
	client.HTTPie = req.HTTPie
	client.IgnoreEvents = req.IgnoreEvents
	client.NoReplay = req.NoReplay
	client.SSEBufferSize = req.SSEBufferSize
	client.UpdatedAt = time.Now()

	// Save updates
	if err := s.clientRepo.Update(client); err != nil {
		return nil, fmt.Errorf("failed to update client: %w", err)
	}

	s.log.Info("Updated client: %s", clientID)

	return client, nil
}

// Delete deletes a client instance.
func (s *ClientService) Delete(clientID string) error {
	// Get client first to get userID
	client, err := s.clientRepo.Get(clientID)
	if err != nil {
		return err
	}

	// Stop if running
	if s.processService.IsRunning(clientID) {
		if err := s.processService.Stop(clientID); err != nil {
			s.log.Error("Failed to stop client before deletion: %v", err)
		}
	}

	// Delete from repository
	if err := s.clientRepo.Delete(clientID); err != nil {
		return fmt.Errorf("failed to delete client: %w", err)
	}

	s.log.Info("Deleted client: %s", clientID)

	// Invalidate quota cache
	s.quotaRepo.(*repository.FileQuotaRepository).InvalidateCache(client.UserID)

	return nil
}

// Start starts a client instance.
func (s *ClientService) Start(clientID string) error {
	// Get client
	client, err := s.clientRepo.Get(clientID)
	if err != nil {
		return err
	}

	// Check if already running
	if s.processService.IsRunning(clientID) {
		return fmt.Errorf("client already running: %s", clientID)
	}

	// Start process
	if err := s.processService.Start(client, s.baseDir); err != nil {
		return fmt.Errorf("failed to start client: %w", err)
	}

	// Update client status
	now := time.Now()
	client.Status = models.ClientStatusRunning
	client.StartedAt = &now
	client.UpdatedAt = now

	if err := s.clientRepo.Update(client); err != nil {
		s.log.Error("Failed to update client status: %v", err)
	}

	s.log.Info("Started client: %s", clientID)

	return nil
}

// Stop stops a client instance.
func (s *ClientService) Stop(clientID string) error {
	// Get client
	client, err := s.clientRepo.Get(clientID)
	if err != nil {
		return err
	}

	// Check if running
	if !s.processService.IsRunning(clientID) {
		return fmt.Errorf("client not running: %s", clientID)
	}

	// Stop process
	if err := s.processService.Stop(clientID); err != nil {
		return fmt.Errorf("failed to stop client: %w", err)
	}

	// Update client status
	now := time.Now()
	client.Status = models.ClientStatusStopped
	client.StoppedAt = &now
	client.UpdatedAt = now

	if err := s.clientRepo.Update(client); err != nil {
		s.log.Error("Failed to update client status: %v", err)
	}

	s.log.Info("Stopped client: %s", clientID)

	return nil
}

// Restart restarts a client instance.
func (s *ClientService) Restart(clientID string) error {
	// Get client
	client, err := s.clientRepo.Get(clientID)
	if err != nil {
		return err
	}

	// Restart process
	if err := s.processService.Restart(client, s.baseDir); err != nil {
		return fmt.Errorf("failed to restart client: %w", err)
	}

	// Update client status
	now := time.Now()
	client.Status = models.ClientStatusRunning
	client.StartedAt = &now
	client.RestartCount++
	client.UpdatedAt = now

	if err := s.clientRepo.Update(client); err != nil {
		s.log.Error("Failed to update client status: %v", err)
	}

	s.log.Info("Restarted client: %s (count: %d)", clientID, client.RestartCount)

	return nil
}

// GetStats retrieves statistics for a client.
func (s *ClientService) GetStats(clientID string) (*models.ClientStats, error) {
	client, err := s.clientRepo.Get(clientID)
	if err != nil {
		return nil, err
	}

	if err := s.populateClientLastActivity(client); err != nil {
		s.log.Error("Failed to populate last activity for stats of client %s: %v", clientID, err)
	}

	stats := &models.ClientStats{
		TodayEvents:   client.TodayEvents,
		TotalEvents:   client.TotalEvents,
		LastEventTime: client.LastActivity,
	}

	// Calculate running time
	if client.StartedAt != nil && client.Status == models.ClientStatusRunning {
		stats.RunningTime = int64(time.Since(*client.StartedAt).Seconds())
	}

	// TODO: Calculate success rate, average latency from event data

	return stats, nil
}

// populateClientLastActivity refreshes the last activity timestamp from stored events.
func (s *ClientService) populateClientLastActivity(client *models.Client) error {
	if client == nil || s.eventRepo == nil {
		return nil
	}

	ts, err := s.eventRepo.GetLatestEventTimestamp(client.ID)
	if err != nil {
		return err
	}

	client.LastActivity = ts
	return nil
}

func (s *ClientService) countClientsByStatus(userID, status string) (int, error) {
	clients, err := s.clientRepo.GetByUserID(userID)
	if err != nil {
		return 0, err
	}

	targetStatus := strings.ToLower(status)
	count := 0

	for _, client := range clients {
		if client == nil {
			continue
		}

		actualStatus := string(client.Status)
		if !strings.EqualFold(actualStatus, string(models.ClientStatusError)) {
			if s.processService.IsRunning(client.ID) {
				actualStatus = string(models.ClientStatusRunning)
			} else {
				actualStatus = string(models.ClientStatusStopped)
			}
		}

		if strings.ToLower(actualStatus) == targetStatus {
			count++
		}
	}

	return count, nil
}

// getBatchTargetClientIDs resolves the list of client IDs for a batch operation.
func (s *ClientService) getBatchTargetClientIDs(userID string, req *models.ClientBatchRequest) ([]string, error) {
	if req == nil {
		return []string{}, nil
	}

	if req.All || len(req.ClientIDs) == 0 {
		clients, err := s.clientRepo.GetByUserID(userID)
		if err != nil {
			return nil, fmt.Errorf("failed to list clients: %w", err)
		}
		ids := make([]string, 0, len(clients))
		for _, client := range clients {
			ids = append(ids, client.ID)
		}
		return ids, nil
	}

	seen := make(map[string]struct{}, len(req.ClientIDs))
	ids := make([]string, 0, len(req.ClientIDs))
	for _, id := range req.ClientIDs {
		trimmed := strings.TrimSpace(id)
		if trimmed == "" {
			continue
		}
		if _, exists := seen[trimmed]; exists {
			continue
		}
		seen[trimmed] = struct{}{}
		ids = append(ids, trimmed)
	}
	return ids, nil
}

// BatchStart starts multiple clients for a user.
func (s *ClientService) BatchStart(userID string, req *models.ClientBatchRequest) (*models.ClientBatchResponse, error) {
	clientIDs, err := s.getBatchTargetClientIDs(userID, req)
	if err != nil {
		return nil, err
	}

	response := &models.ClientBatchResponse{
		Total:   len(clientIDs),
		Results: make([]*models.ClientBatchResult, 0, len(clientIDs)),
	}

	if len(clientIDs) == 0 {
		return response, nil
	}

	for _, clientID := range clientIDs {
		result := &models.ClientBatchResult{
			ClientID: clientID,
		}

		client, err := s.clientRepo.Get(clientID)
		if err != nil {
			result.Message = fmt.Sprintf("failed to load client: %v", err)
			response.Failed++
			response.Results = append(response.Results, result)
			continue
		}

		if client.UserID != userID {
			result.Message = "client does not belong to current user"
			response.Failed++
			response.Results = append(response.Results, result)
			continue
		}

		if err := s.Start(clientID); err != nil {
			result.Message = err.Error()
			response.Failed++
		} else {
			result.Success = true
			response.Successful++
		}

		response.Results = append(response.Results, result)
	}

	s.log.Info("Batch start completed: user=%s, total=%d, successful=%d, failed=%d",
		userID, response.Total, response.Successful, response.Failed)

	return response, nil
}

// BatchStop stops multiple clients for a user.
func (s *ClientService) BatchStop(userID string, req *models.ClientBatchRequest) (*models.ClientBatchResponse, error) {
	clientIDs, err := s.getBatchTargetClientIDs(userID, req)
	if err != nil {
		return nil, err
	}

	response := &models.ClientBatchResponse{
		Total:   len(clientIDs),
		Results: make([]*models.ClientBatchResult, 0, len(clientIDs)),
	}

	if len(clientIDs) == 0 {
		return response, nil
	}

	for _, clientID := range clientIDs {
		result := &models.ClientBatchResult{
			ClientID: clientID,
		}

		client, err := s.clientRepo.Get(clientID)
		if err != nil {
			result.Message = fmt.Sprintf("failed to load client: %v", err)
			response.Failed++
			response.Results = append(response.Results, result)
			continue
		}

		if client.UserID != userID {
			result.Message = "client does not belong to current user"
			response.Failed++
			response.Results = append(response.Results, result)
			continue
		}

		if err := s.Stop(clientID); err != nil {
			result.Message = err.Error()
			response.Failed++
		} else {
			result.Success = true
			response.Successful++
		}

		response.Results = append(response.Results, result)
	}

	s.log.Info("Batch stop completed: user=%s, total=%d, successful=%d, failed=%d",
		userID, response.Total, response.Successful, response.Failed)

	return response, nil
}
