// Copyright (c) 2025 Lazycat Apps
// Licensed under the MIT License. See LICENSE file in the project root for details.

package repository

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/lazycatapps/gosmee/backend/internal/models"
)

// ClientRepository defines the interface for client instance storage operations.
type ClientRepository interface {
	// Create creates a new client instance
	Create(client *models.Client) error
	// Get retrieves a client by ID
	Get(id string) (*models.Client, error)
	// GetByUserID retrieves all clients for a user
	GetByUserID(userID string) ([]*models.Client, error)
	// Update updates an existing client
	Update(client *models.Client) error
	// Delete deletes a client by ID
	Delete(id string) error
	// List retrieves clients with filters and pagination
	List(userID string, req *models.ClientListRequest) (*models.ClientListResponse, error)
}

// FileClientRepository implements ClientRepository using file system storage.
type FileClientRepository struct {
	baseDir string     // Base data directory
	mu      sync.RWMutex // Mutex for thread-safe operations
}

// NewFileClientRepository creates a new file-based client repository.
func NewFileClientRepository(baseDir string) (*FileClientRepository, error) {
	// Ensure base directory exists
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}

	return &FileClientRepository{
		baseDir: baseDir,
	}, nil
}

// getClientConfigPath returns the path to client config file.
func (r *FileClientRepository) getClientConfigPath(userID, clientID string) string {
	return filepath.Join(r.baseDir, "users", userID, "clients", clientID, "config.json")
}

// getUserClientsDir returns the directory containing all clients for a user.
func (r *FileClientRepository) getUserClientsDir(userID string) string {
	return filepath.Join(r.baseDir, "users", userID, "clients")
}

// Create creates a new client instance.
func (r *FileClientRepository) Create(client *models.Client) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	configPath := r.getClientConfigPath(client.UserID, client.ID)

	// Check if client already exists
	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("client already exists: %s", client.ID)
	}

	// Create directory structure
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("failed to create client directory: %w", err)
	}

	// Create events and logs directories
	clientDir := filepath.Dir(configPath)
	if err := os.MkdirAll(filepath.Join(clientDir, "events"), 0755); err != nil {
		return fmt.Errorf("failed to create events directory: %w", err)
	}
	if err := os.MkdirAll(filepath.Join(clientDir, "logs"), 0755); err != nil {
		return fmt.Errorf("failed to create logs directory: %w", err)
	}

	// Write config file
	return r.writeClientConfig(configPath, client)
}

// Get retrieves a client by ID.
func (r *FileClientRepository) Get(id string) (*models.Client, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// We need to search through all users to find the client
	// This is inefficient but acceptable for MVP
	// TODO: Add index for faster lookups
	usersDir := filepath.Join(r.baseDir, "users")
	userDirs, err := os.ReadDir(usersDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("client not found: %s", id)
		}
		return nil, fmt.Errorf("failed to read users directory: %w", err)
	}

	for _, userDir := range userDirs {
		if !userDir.IsDir() {
			continue
		}
		configPath := r.getClientConfigPath(userDir.Name(), id)
		if client, err := r.readClientConfig(configPath); err == nil {
			return client, nil
		}
	}

	return nil, fmt.Errorf("client not found: %s", id)
}

// GetByUserID retrieves all clients for a user.
func (r *FileClientRepository) GetByUserID(userID string) ([]*models.Client, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	clientsDir := r.getUserClientsDir(userID)
	clientDirs, err := os.ReadDir(clientsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []*models.Client{}, nil
		}
		return nil, fmt.Errorf("failed to read clients directory: %w", err)
	}

	var clients []*models.Client
	for _, clientDir := range clientDirs {
		if !clientDir.IsDir() {
			continue
		}
		configPath := r.getClientConfigPath(userID, clientDir.Name())
		client, err := r.readClientConfig(configPath)
		if err != nil {
			// Skip invalid configs
			continue
		}
		clients = append(clients, client)
	}

	return clients, nil
}

// Update updates an existing client.
func (r *FileClientRepository) Update(client *models.Client) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	configPath := r.getClientConfigPath(client.UserID, client.ID)

	// Check if client exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("client not found: %s", client.ID)
	}

	// Write updated config
	return r.writeClientConfig(configPath, client)
}

// Delete deletes a client by ID.
func (r *FileClientRepository) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Find client first
	client, err := r.Get(id)
	if err != nil {
		return err
	}

	// Delete entire client directory
	clientDir := filepath.Join(r.baseDir, "users", client.UserID, "clients", id)
	if err := os.RemoveAll(clientDir); err != nil {
		return fmt.Errorf("failed to delete client directory: %w", err)
	}

	return nil
}

// List retrieves clients with filters and pagination.
func (r *FileClientRepository) List(userID string, req *models.ClientListRequest) (*models.ClientListResponse, error) {
	// Get all clients for user
	clients, err := r.GetByUserID(userID)
	if err != nil {
		return nil, err
	}

	// Apply filters
	filtered := r.filterClients(clients, req)

	// Sort
	r.sortClients(filtered, req.SortBy, req.SortOrder)

	// Apply pagination
	total := len(filtered)
	start := (req.Page - 1) * req.PageSize
	end := start + req.PageSize
	if start >= total {
		start = 0
		end = 0
	}
	if end > total {
		end = total
	}

	paged := filtered[start:end]

	// Convert to summaries
	summaries := make([]*models.ClientSummary, len(paged))
	for i, client := range paged {
		summaries[i] = client.ToSummary()
	}

	return &models.ClientListResponse{
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
		Clients:  summaries,
	}, nil
}

// filterClients applies filters to client list.
func (r *FileClientRepository) filterClients(clients []*models.Client, req *models.ClientListRequest) []*models.Client {
	var filtered []*models.Client

	for _, client := range clients {
		// Filter by status
		if req.Status != "" && string(client.Status) != req.Status {
			continue
		}

		// Filter by search (name contains)
		if req.Search != "" && !strings.Contains(strings.ToLower(client.Name), strings.ToLower(req.Search)) {
			continue
		}

		filtered = append(filtered, client)
	}

	return filtered
}

// sortClients sorts clients by field and order.
func (r *FileClientRepository) sortClients(clients []*models.Client, sortBy, sortOrder string) {
	sort.Slice(clients, func(i, j int) bool {
		var less bool
		switch sortBy {
		case "name":
			less = clients[i].Name < clients[j].Name
		case "status":
			less = clients[i].Status < clients[j].Status
		case "createdAt":
			less = clients[i].CreatedAt.Before(clients[j].CreatedAt)
		default: // default to createdAt
			less = clients[i].CreatedAt.Before(clients[j].CreatedAt)
		}

		if sortOrder == "asc" {
			return less
		}
		return !less
	})
}

// readClientConfig reads client config from file.
func (r *FileClientRepository) readClientConfig(path string) (*models.Client, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var client models.Client
	if err := json.Unmarshal(data, &client); err != nil {
		return nil, fmt.Errorf("failed to parse client config: %w", err)
	}

	return &client, nil
}

// writeClientConfig writes client config to file.
func (r *FileClientRepository) writeClientConfig(path string, client *models.Client) error {
	data, err := json.MarshalIndent(client, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal client config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write client config: %w", err)
	}

	return nil
}
