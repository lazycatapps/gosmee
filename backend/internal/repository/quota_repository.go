// Copyright (c) 2025 Lazycat Apps
// Licensed under the MIT License. See LICENSE file in the project root for details.

package repository

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/lazycatapps/gosmee/backend/internal/models"
)

// QuotaRepository defines the interface for quota management operations.
type QuotaRepository interface {
	// GetQuota retrieves quota information for a user
	GetQuota(userID string) (*models.Quota, error)
	// CalculateUsage calculates current storage usage for a user
	CalculateUsage(userID string) (int64, error)
	// CountClients counts the number of clients for a user
	CountClients(userID string) (int, error)
}

// FileQuotaRepository implements QuotaRepository using file system storage.
type FileQuotaRepository struct {
	baseDir           string       // Base data directory
	maxStoragePerUser int64        // Maximum storage per user in bytes
	maxClientsPerUser int          // Maximum clients per user
	cache             sync.Map     // Cache of quota information (key: userID, value: *quotaCache)
	cacheTTL          time.Duration // Cache TTL
	mu                sync.RWMutex // Mutex for thread-safe operations
}

// quotaCache represents cached quota information.
type quotaCache struct {
	quota     *models.Quota
	expiresAt time.Time
}

// NewFileQuotaRepository creates a new file-based quota repository.
func NewFileQuotaRepository(baseDir string, maxStoragePerUser int64, maxClientsPerUser int) *FileQuotaRepository {
	return &FileQuotaRepository{
		baseDir:           baseDir,
		maxStoragePerUser: maxStoragePerUser,
		maxClientsPerUser: maxClientsPerUser,
		cacheTTL:          1 * time.Hour, // Cache for 1 hour
	}
}

// GetQuota retrieves quota information for a user.
func (r *FileQuotaRepository) GetQuota(userID string) (*models.Quota, error) {
	// Check cache first
	if cached, ok := r.cache.Load(userID); ok {
		cache := cached.(*quotaCache)
		if time.Now().Before(cache.expiresAt) {
			return cache.quota, nil
		}
	}

	// Calculate fresh quota
	quota, err := r.calculateQuota(userID)
	if err != nil {
		return nil, err
	}

	// Update cache
	r.cache.Store(userID, &quotaCache{
		quota:     quota,
		expiresAt: time.Now().Add(r.cacheTTL),
	})

	return quota, nil
}

// CalculateUsage calculates current storage usage for a user.
func (r *FileQuotaRepository) CalculateUsage(userID string) (int64, error) {
	userDir := filepath.Join(r.baseDir, "users", userID)

	// Check if user directory exists
	if _, err := os.Stat(userDir); os.IsNotExist(err) {
		return 0, nil
	}

	var totalSize int64

	// Walk through user directory and sum file sizes
	err := filepath.WalkDir(userDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			info, err := d.Info()
			if err != nil {
				return err
			}
			totalSize += info.Size()
		}
		return nil
	})

	if err != nil {
		return 0, fmt.Errorf("failed to calculate storage usage: %w", err)
	}

	return totalSize, nil
}

// CountClients counts the number of clients for a user.
func (r *FileQuotaRepository) CountClients(userID string) (int, error) {
	clientsDir := filepath.Join(r.baseDir, "users", userID, "clients")

	// Check if clients directory exists
	if _, err := os.Stat(clientsDir); os.IsNotExist(err) {
		return 0, nil
	}

	clientDirs, err := os.ReadDir(clientsDir)
	if err != nil {
		return 0, fmt.Errorf("failed to read clients directory: %w", err)
	}

	count := 0
	for _, dir := range clientDirs {
		if dir.IsDir() {
			count++
		}
	}

	return count, nil
}

// calculateQuota calculates fresh quota information.
func (r *FileQuotaRepository) calculateQuota(userID string) (*models.Quota, error) {
	quota := models.NewQuota(userID, r.maxStoragePerUser, r.maxClientsPerUser)

	// Calculate storage usage
	usedBytes, err := r.CalculateUsage(userID)
	if err != nil {
		return nil, err
	}

	// Count clients
	clientsCount, err := r.CountClients(userID)
	if err != nil {
		return nil, err
	}

	quota.UpdateUsage(usedBytes, clientsCount)

	return quota, nil
}

// InvalidateCache invalidates the quota cache for a user.
func (r *FileQuotaRepository) InvalidateCache(userID string) {
	r.cache.Delete(userID)
}
