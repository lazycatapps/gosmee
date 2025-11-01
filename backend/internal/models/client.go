// Copyright (c) 2025 Lazycat Apps
// Licensed under the MIT License. See LICENSE file in the project root for details.

// Package models defines data structures for the Gosmee Web UI application.
package models

import (
	"time"
)

// ClientStatus represents the current state of a gosmee client instance.
type ClientStatus string

const (
	ClientStatusRunning ClientStatus = "running" // Client process is running
	ClientStatusStopped ClientStatus = "stopped" // Client process is stopped
	ClientStatusError   ClientStatus = "error"   // Client process encountered an error
)

// Client represents a gosmee client instance configuration and status.
type Client struct {
	ID          string       `json:"id"`          // Unique client identifier (UUID)
	UserID      string       `json:"userId"`      // User ID (for OIDC multi-tenancy)
	Name        string       `json:"name"`        // User-friendly name
	Description string       `json:"description"` // Instance description
	Status      ClientStatus `json:"status"`      // Current status

	// Gosmee configuration
	SmeeURL       string   `json:"smeeUrl"`                // Gosmee server event source URL
	TargetURL     string   `json:"targetUrl"`              // Target webhook receiver URL
	TargetTimeout int      `json:"targetTimeout"`          // Target connection timeout in seconds
	HTTPie        bool     `json:"httpie"`                 // Generate HTTPie scripts instead of cURL
	IgnoreEvents  []string `json:"ignoreEvents,omitempty"` // Event types to filter
	NoReplay      bool     `json:"noReplay"`               // Save only, don't forward events
	SSEBufferSize int      `json:"sseBufferSize"`          // SSE buffer size in bytes

	// Process information
	PID          int        `json:"pid,omitempty"`       // Process ID (when running)
	StartedAt    *time.Time `json:"startedAt,omitempty"` // Last start time
	StoppedAt    *time.Time `json:"stoppedAt,omitempty"` // Last stop time
	RestartCount int        `json:"restartCount"`        // Number of restarts
	LastError    string     `json:"lastError,omitempty"` // Last error message

	// Statistics
	TodayEvents  int        `json:"todayEvents"`            // Events forwarded today
	TotalEvents  int        `json:"totalEvents"`            // Total events forwarded
	LastActivity *time.Time `json:"lastActivity,omitempty"` // Last event time

	// Metadata
	CreatedAt time.Time `json:"createdAt"` // Creation timestamp
	UpdatedAt time.Time `json:"updatedAt"` // Last update timestamp
}

// NewClient creates a new client instance with default values.
func NewClient(id, userID, name, description, smeeURL, targetURL string) *Client {
	now := time.Now()
	return &Client{
		ID:            id,
		UserID:        userID,
		Name:          name,
		Description:   description,
		Status:        ClientStatusStopped,
		SmeeURL:       smeeURL,
		TargetURL:     targetURL,
		TargetTimeout: 60, // Default 60 seconds
		HTTPie:        false,
		NoReplay:      false,
		SSEBufferSize: 1048576, // Default 1MB
		RestartCount:  0,
		TodayEvents:   0,
		TotalEvents:   0,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

// ToSummary converts a Client to ClientSummary (for list queries).
func (c *Client) ToSummary() *ClientSummary {
	return &ClientSummary{
		ID:           c.ID,
		Name:         c.Name,
		Status:       string(c.Status),
		SmeeURL:      c.SmeeURL,
		TargetURL:    c.TargetURL,
		TodayEvents:  c.TodayEvents,
		TotalEvents:  c.TotalEvents,
		LastActivity: c.LastActivity,
	}
}

// ClientSummary represents a summarized view of a client (for list queries).
type ClientSummary struct {
	ID           string     `json:"id"`
	Name         string     `json:"name"`
	Status       string     `json:"status"`
	SmeeURL      string     `json:"smeeUrl"`
	TargetURL    string     `json:"targetUrl"`
	TodayEvents  int        `json:"todayEvents"`
	TotalEvents  int        `json:"totalEvents"`
	LastActivity *time.Time `json:"lastActivity,omitempty"`
}

// ClientRequest represents the request body for creating/updating a client.
type ClientRequest struct {
	Name          string   `json:"name" binding:"required"`      // Instance name (required)
	Description   string   `json:"description"`                  // Instance description (optional)
	SmeeURL       string   `json:"smeeUrl" binding:"required"`   // Smee server URL (required)
	TargetURL     string   `json:"targetUrl" binding:"required"` // Target URL (required)
	TargetTimeout int      `json:"targetTimeout"`                // Target timeout (optional, default: 60)
	HTTPie        bool     `json:"httpie"`                       // Use HTTPie format (optional)
	IgnoreEvents  []string `json:"ignoreEvents"`                 // Events to ignore (optional)
	NoReplay      bool     `json:"noReplay"`                     // Save only mode (optional)
	SSEBufferSize int      `json:"sseBufferSize"`                // SSE buffer size (optional, default: 1048576)
}

// ClientListRequest represents query parameters for listing clients.
type ClientListRequest struct {
	Page      int    `form:"page,default=1"`           // Page number (default: 1)
	PageSize  int    `form:"pageSize,default=20"`      // Items per page (default: 20, max: 100)
	Status    string `form:"status"`                   // Filter by status (optional)
	Search    string `form:"search"`                   // Search by name (optional)
	SortBy    string `form:"sortBy,default=createdAt"` // Sort field (default: createdAt)
	SortOrder string `form:"sortOrder,default=desc"`   // Sort order: asc/desc (default: desc)
}

// ClientListResponse represents the response for client list queries.
type ClientListResponse struct {
	Total    int              `json:"total"`    // Total number of clients matching filter
	Page     int              `json:"page"`     // Current page number
	PageSize int              `json:"pageSize"` // Items per page
	Clients  []*ClientSummary `json:"clients"`  // Client summaries for current page
}

// ClientBatchRequest represents a batch operation request for clients.
type ClientBatchRequest struct {
	ClientIDs []string `json:"clientIds"`     // Client IDs to operate on
	All       bool     `json:"all,omitempty"` // Whether to operate on all clients
}

// ClientBatchResult represents the result of a batch operation for a single client.
type ClientBatchResult struct {
	ClientID string `json:"clientId"`          // Client ID
	Success  bool   `json:"success"`           // Whether operation succeeded
	Message  string `json:"message,omitempty"` // Optional error or info message
}

// ClientBatchResponse represents the aggregated result of a batch operation.
type ClientBatchResponse struct {
	Total      int                  `json:"total"`      // Total number of clients processed
	Successful int                  `json:"successful"` // Number of successful operations
	Failed     int                  `json:"failed"`     // Number of failed operations
	Results    []*ClientBatchResult `json:"results"`    // Per-client results
}
