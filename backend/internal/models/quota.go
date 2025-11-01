// Copyright (c) 2025 Lazycat Apps
// Licensed under the MIT License. See LICENSE file in the project root for details.

package models

import (
	"time"
)

// Quota represents user storage quota information.
type Quota struct {
	UserID       string    `json:"userId"`       // User ID
	TotalBytes   int64     `json:"totalBytes"`   // Total quota in bytes
	UsedBytes    int64     `json:"usedBytes"`    // Used storage in bytes
	Percentage   float64   `json:"percentage"`   // Usage percentage (0-100)
	ClientsCount int       `json:"clientsCount"` // Current number of clients
	MaxClients   int       `json:"maxClients"`   // Maximum allowed clients
	UpdatedAt    time.Time `json:"updatedAt"`    // Last update time
}

// NewQuota creates a new Quota instance.
func NewQuota(userID string, totalBytes int64, maxClients int) *Quota {
	return &Quota{
		UserID:       userID,
		TotalBytes:   totalBytes,
		UsedBytes:    0,
		Percentage:   0.0,
		ClientsCount: 0,
		MaxClients:   maxClients,
		UpdatedAt:    time.Now(),
	}
}

// UpdateUsage updates the quota usage statistics.
func (q *Quota) UpdateUsage(usedBytes int64, clientsCount int) {
	q.UsedBytes = usedBytes
	q.ClientsCount = clientsCount
	if q.TotalBytes > 0 {
		q.Percentage = float64(usedBytes) / float64(q.TotalBytes) * 100
	}
	q.UpdatedAt = time.Now()
}

// IsStorageFull checks if storage quota is full (100%).
func (q *Quota) IsStorageFull() bool {
	return q.Percentage >= 100.0
}

// IsStorageWarning checks if storage usage is above 80%.
func (q *Quota) IsStorageWarning() bool {
	return q.Percentage >= 80.0
}

// IsClientsLimitReached checks if the maximum number of clients is reached.
func (q *Quota) IsClientsLimitReached() bool {
	return q.ClientsCount >= q.MaxClients
}

// CanCreateClient checks if a new client can be created.
func (q *Quota) CanCreateClient() bool {
	return !q.IsClientsLimitReached()
}
