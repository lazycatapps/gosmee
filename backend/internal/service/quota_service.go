// Copyright (c) 2025 Lazycat Apps
// Licensed under the MIT License. See LICENSE file in the project root for details.

package service

import (
	"fmt"

	"github.com/lazycatapps/gosmee/backend/internal/models"
	"github.com/lazycatapps/gosmee/backend/internal/pkg/logger"
	"github.com/lazycatapps/gosmee/backend/internal/repository"
)

// QuotaService manages user quotas.
type QuotaService struct {
	quotaRepo repository.QuotaRepository
	log       logger.Logger
}

// NewQuotaService creates a new quota service.
func NewQuotaService(quotaRepo repository.QuotaRepository, log logger.Logger) *QuotaService {
	return &QuotaService{
		quotaRepo: quotaRepo,
		log:       log,
	}
}

// GetQuota retrieves quota information for a user.
func (s *QuotaService) GetQuota(userID string) (*models.Quota, error) {
	quota, err := s.quotaRepo.GetQuota(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get quota: %w", err)
	}

	return quota, nil
}

// CheckCanCreateClient checks if a user can create a new client.
func (s *QuotaService) CheckCanCreateClient(userID string) error {
	quota, err := s.GetQuota(userID)
	if err != nil {
		return err
	}

	if !quota.CanCreateClient() {
		return fmt.Errorf("client limit reached: %d/%d", quota.ClientsCount, quota.MaxClients)
	}

	return nil
}

// CheckStorageQuota checks if user has enough storage.
func (s *QuotaService) CheckStorageQuota(userID string) error {
	quota, err := s.GetQuota(userID)
	if err != nil {
		return err
	}

	if quota.IsStorageFull() {
		return fmt.Errorf("storage quota exceeded: %.2f%% used", quota.Percentage)
	}

	return nil
}

// GetStorageWarning returns a warning message if storage is above 80%.
func (s *QuotaService) GetStorageWarning(userID string) (string, error) {
	quota, err := s.GetQuota(userID)
	if err != nil {
		return "", err
	}

	if quota.IsStorageWarning() {
		return fmt.Sprintf("Warning: Storage usage is at %.2f%% - consider cleaning up old events and logs", quota.Percentage), nil
	}

	return "", nil
}
