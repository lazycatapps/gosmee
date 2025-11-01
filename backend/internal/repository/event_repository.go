// Copyright (c) 2025 Lazycat Apps
// Licensed under the MIT License. See LICENSE file in the project root for details.

package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/lazycatapps/gosmee/backend/internal/models"
)

// EventRepository defines the interface for event storage operations.
type EventRepository interface {
	// GetByClientID retrieves events for a specific client
	GetByClientID(clientID string, req *models.EventListRequest) (*models.EventListResponse, error)
	// Get retrieves a single event by ID
	Get(clientID, eventID string) (*models.Event, error)
	// Delete deletes an event
	Delete(clientID, eventID string) error
	// DeleteBatch deletes multiple events
	DeleteBatch(clientID string, eventIDs []string) error
	// CleanupOldEvents removes events older than retention period
	CleanupOldEvents(clientID string, retentionDays int) error
	// GetLatestEventTimestamp returns the latest event timestamp for a client
	GetLatestEventTimestamp(clientID string) (*time.Time, error)
}

// FileEventRepository implements EventRepository using file system storage.
type FileEventRepository struct {
	baseDir string       // Base data directory
	mu      sync.RWMutex // Mutex for thread-safe operations
}

// NewFileEventRepository creates a new file-based event repository.
func NewFileEventRepository(baseDir string) *FileEventRepository {
	return &FileEventRepository{
		baseDir: baseDir,
	}
}

// getEventsDir returns the events directory for a client.
func (r *FileEventRepository) getEventsDir(clientID string) (string, error) {
	// We need to find the client's user directory first
	// This is a simplified approach - in production, you'd want an index
	usersDir := filepath.Join(r.baseDir, "users")
	userDirs, err := os.ReadDir(usersDir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return "", fs.ErrNotExist
		}
		return "", fmt.Errorf("failed to read users directory: %w", err)
	}

	for _, userDir := range userDirs {
		if !userDir.IsDir() {
			continue
		}
		eventsDir := filepath.Join(r.baseDir, "users", userDir.Name(), "clients", clientID, "events")
		if _, err := os.Stat(eventsDir); err == nil {
			return eventsDir, nil
		}
	}

	return "", fmt.Errorf("events directory not found for client %s: %w", clientID, fs.ErrNotExist)
}

// GetByClientID retrieves events for a specific client with filters and pagination.
func (r *FileEventRepository) GetByClientID(clientID string, req *models.EventListRequest) (*models.EventListResponse, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	eventsDir, err := r.getEventsDir(clientID)
	if err != nil {
		return &models.EventListResponse{
			Total:    0,
			Page:     req.Page,
			PageSize: req.PageSize,
			Events:   []*models.EventSummary{},
		}, nil
	}

	// Read all event files
	events, err := r.readAllEvents(eventsDir)
	if err != nil {
		return nil, err
	}

	// Apply filters
	filtered := r.filterEvents(events, req)

	// Sort
	r.sortEvents(filtered, req.SortBy, req.SortOrder)

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
	summaries := make([]*models.EventSummary, len(paged))
	for i, event := range paged {
		summaries[i] = event.ToSummary()
	}

	return &models.EventListResponse{
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
		Events:   summaries,
	}, nil
}

// Get retrieves a single event by ID.
func (r *FileEventRepository) Get(clientID, eventID string) (*models.Event, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	eventsDir, err := r.getEventsDir(clientID)
	if err != nil {
		return nil, err
	}

	// Check flat layout first
	flatPath := filepath.Join(eventsDir, fmt.Sprintf("%s.json", eventID))
	if _, err := os.Stat(flatPath); err == nil {
		return r.readEventFile(flatPath)
	}

	// Search through date directories
	dateDirs, err := os.ReadDir(eventsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read events directory: %w", err)
	}

	for _, dateDir := range dateDirs {
		if !dateDir.IsDir() {
			continue
		}

		eventPath := filepath.Join(eventsDir, dateDir.Name(), fmt.Sprintf("%s.json", eventID))
		if event, err := r.readEventFile(eventPath); err == nil {
			return event, nil
		}
	}

	return nil, fmt.Errorf("event not found: %s", eventID)
}

// Delete deletes an event.
func (r *FileEventRepository) Delete(clientID, eventID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	eventsDir, err := r.getEventsDir(clientID)
	if err != nil {
		return err
	}

	// Delete from flat layout if present
	flatJSONPath := filepath.Join(eventsDir, fmt.Sprintf("%s.json", eventID))
	flatShPath := filepath.Join(eventsDir, fmt.Sprintf("%s.sh", eventID))
	if _, err := os.Stat(flatJSONPath); err == nil {
		os.Remove(flatJSONPath)
		os.Remove(flatShPath)
		return nil
	}

	// Search through date directories
	dateDirs, err := os.ReadDir(eventsDir)
	if err != nil {
		return fmt.Errorf("failed to read events directory: %w", err)
	}

	for _, dateDir := range dateDirs {
		if !dateDir.IsDir() {
			continue
		}

		// Delete both JSON and shell script files
		eventJSONPath := filepath.Join(eventsDir, dateDir.Name(), fmt.Sprintf("%s.json", eventID))
		eventShPath := filepath.Join(eventsDir, dateDir.Name(), fmt.Sprintf("%s.sh", eventID))

		if _, err := os.Stat(eventJSONPath); err == nil {
			os.Remove(eventJSONPath)
			os.Remove(eventShPath) // Ignore error if .sh doesn't exist
			return nil
		}
	}

	return fmt.Errorf("event not found: %s", eventID)
}

// DeleteBatch deletes multiple events.
func (r *FileEventRepository) DeleteBatch(clientID string, eventIDs []string) error {
	for _, eventID := range eventIDs {
		if err := r.Delete(clientID, eventID); err != nil {
			// Continue deleting even if one fails
			continue
		}
	}
	return nil
}

// CleanupOldEvents removes events older than retention period.
func (r *FileEventRepository) CleanupOldEvents(clientID string, retentionDays int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if retentionDays == 0 {
		return nil // Keep forever
	}

	eventsDir, err := r.getEventsDir(clientID)
	if err != nil {
		return err
	}

	cutoffDate := time.Now().AddDate(0, 0, -retentionDays)

	// Read date directories
	dateDirs, err := os.ReadDir(eventsDir)
	if err != nil {
		return fmt.Errorf("failed to read events directory: %w", err)
	}

	for _, dateDir := range dateDirs {
		if !dateDir.IsDir() {
			continue
		}

		// Parse date from directory name (YYYY-MM-DD)
		dirDate, err := time.Parse("2006-01-02", dateDir.Name())
		if err != nil {
			continue
		}

		// Delete if older than retention period
		if dirDate.Before(cutoffDate) {
			dateDirPath := filepath.Join(eventsDir, dateDir.Name())
			os.RemoveAll(dateDirPath)
		}
	}

	return nil
}

// GetLatestEventTimestamp returns the most recent event timestamp for a client.
func (r *FileEventRepository) GetLatestEventTimestamp(clientID string) (*time.Time, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	eventsDir, err := r.getEventsDir(clientID)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}

	entries, err := os.ReadDir(eventsDir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read events directory: %w", err)
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() > entries[j].Name()
	})

	var latest *time.Time
	updateLatest := func(path string) {
		var candidate time.Time

		if event, err := r.readEventFile(path); err == nil && !event.Timestamp.IsZero() {
			candidate = event.Timestamp
		} else {
			info, statErr := os.Stat(path)
			if statErr != nil {
				return
			}
			candidate = info.ModTime()
		}

		if latest == nil || candidate.After(*latest) {
			ts := candidate
			latest = &ts
		}
	}

	for _, entry := range entries {
		if entry.IsDir() {
			// Compatibility with legacy per-day directories
			dirPath := filepath.Join(eventsDir, entry.Name())
			files, err := os.ReadDir(dirPath)
			if err != nil {
				continue
			}

			sort.Slice(files, func(i, j int) bool {
				return files[i].Name() > files[j].Name()
			})

			for _, file := range files {
				if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
					continue
				}
				updateLatest(filepath.Join(dirPath, file.Name()))
				if latest != nil {
					return latest, nil
				}
			}
			continue
		}

		if !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		updateLatest(filepath.Join(eventsDir, entry.Name()))
		if latest != nil {
			return latest, nil
		}
	}

	return latest, nil
}

// readAllEvents reads all events from the events directory.
func (r *FileEventRepository) readAllEvents(eventsDir string) ([]*models.Event, error) {
	var events []*models.Event

	err := filepath.WalkDir(eventsDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			if errors.Is(walkErr, fs.ErrNotExist) {
				return nil
			}
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(d.Name(), ".json") {
			return nil
		}

		event, err := r.readEventFile(path)
		if err != nil {
			return nil
		}
		events = append(events, event)
		return nil
	})
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return []*models.Event{}, nil
		}
		return nil, fmt.Errorf("failed to read events: %w", err)
	}

	return events, nil
}

// readEventFile reads an event from a JSON file.
func (r *FileEventRepository) readEventFile(path string) (*models.Event, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var event models.Event
	if err := json.Unmarshal(data, &event); err != nil {
		return r.buildEventFromRaw(path, data), nil
	}

	r.enrichEventFromPath(&event, path, data)

	return &event, nil
}

func (r *FileEventRepository) buildEventFromRaw(path string, data []byte) *models.Event {
	payload := strings.TrimSpace(string(data))

	event := &models.Event{
		Payload: payload,
		Status:  models.EventStatusNotReplayed,
	}

	r.enrichEventFromPath(event, path, data)

	return event
}

// filterEvents applies filters to event list.
func (r *FileEventRepository) filterEvents(events []*models.Event, req *models.EventListRequest) []*models.Event {
	var filtered []*models.Event

	for _, event := range events {
		// Filter by event type
		if req.EventType != "" && event.EventType != req.EventType {
			continue
		}

		// Filter by status
		if req.Status != "" && string(event.Status) != req.Status {
			continue
		}

		// Filter by search (source contains)
		if req.Search != "" && !strings.Contains(strings.ToLower(event.Source), strings.ToLower(req.Search)) {
			continue
		}

		// Filter by date range
		if !req.DateFrom.IsZero() && event.Timestamp.Before(req.DateFrom) {
			continue
		}
		if !req.DateTo.IsZero() && event.Timestamp.After(req.DateTo) {
			continue
		}

		filtered = append(filtered, event)
	}

	return filtered
}

// sortEvents sorts events by field and order.
func (r *FileEventRepository) sortEvents(events []*models.Event, sortBy, sortOrder string) {
	sort.Slice(events, func(i, j int) bool {
		var less bool
		switch sortBy {
		case "eventType":
			less = events[i].EventType < events[j].EventType
		case "status":
			less = events[i].Status < events[j].Status
		case "timestamp":
			less = events[i].Timestamp.Before(events[j].Timestamp)
		default: // default to timestamp
			less = events[i].Timestamp.Before(events[j].Timestamp)
		}

		if sortOrder == "asc" {
			return less
		}
		return !less
	})
}

func (r *FileEventRepository) enrichEventFromPath(event *models.Event, path string, data []byte) {
	if event == nil {
		return
	}

	eventID := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	if event.ID == "" {
		event.ID = eventID
	}

	if event.ClientID == "" {
		event.ClientID = inferClientIDFromPath(path)
	}

	if event.Timestamp.IsZero() {
		if ts, ok := parseTimestampFromEventID(event.ID); ok {
			event.Timestamp = ts
		}
	}

	if event.Timestamp.IsZero() {
		if info, err := os.Stat(path); err == nil {
			event.Timestamp = info.ModTime()
		}
	}

	if event.Payload == "" {
		event.Payload = strings.TrimSpace(string(data))
	}

	// Try to load headers from corresponding .sh file if headers are empty
	if len(event.Headers) == 0 {
		if headers := r.loadHeadersFromShellScript(path); len(headers) > 0 {
			event.Headers = headers
		} else {
			event.Headers = nil
		}
	}

	if event.Status == "" {
		event.Status = models.EventStatusNotReplayed
	}
}

func inferClientIDFromPath(path string) string {
	eventsDir := filepath.Dir(path)
	clientDir := filepath.Dir(eventsDir)
	return filepath.Base(clientDir)
}

func parseTimestampFromEventID(eventID string) (time.Time, bool) {
	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15.04.05.000",
		"2006-01-02T15.04.05",
	}

	for _, layout := range layouts {
		if ts, err := time.Parse(layout, eventID); err == nil {
			return ts.UTC(), true
		}
	}

	return time.Time{}, false
}

// loadHeadersFromShellScript parses headers from the companion .sh file
func (r *FileEventRepository) loadHeadersFromShellScript(jsonPath string) map[string]string {
	// Replace .json extension with .sh
	shPath := strings.TrimSuffix(jsonPath, ".json") + ".sh"

	// Check if .sh file exists
	if _, err := os.Stat(shPath); err != nil {
		return nil
	}

	// Read shell script
	content, err := os.ReadFile(shPath)
	if err != nil {
		return nil
	}

	headers := make(map[string]string)

	// Find the curl command line (contains 'curl' and multiple '-H' flags)
	// Example: curl $curl_flags -H "Content-Type: application/json" -H 'X-Forwarded-For: 212.50.251.184' ...
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Only process lines that contain 'curl' command
		if !strings.Contains(line, "curl") {
			continue
		}

		// Parse all -H flags in the curl command
		// Match patterns: -H "Header: Value" or -H 'Header: Value'
		remaining := line
		for {
			// Find next -H flag
			hIdx := strings.Index(remaining, "-H")
			if hIdx == -1 {
				break
			}

			// Skip past "-H "
			remaining = remaining[hIdx+2:]
			remaining = strings.TrimSpace(remaining)

			// Determine quote type (single or double)
			var quoteChar byte
			if len(remaining) > 0 {
				if remaining[0] == '"' {
					quoteChar = '"'
				} else if remaining[0] == '\'' {
					quoteChar = '\''
				} else {
					// No quote found, skip this -H
					continue
				}
			} else {
				break
			}

			// Find the closing quote
			closeIdx := strings.Index(remaining[1:], string(quoteChar))
			if closeIdx == -1 {
				// No closing quote found
				break
			}

			// Extract header content (between quotes)
			headerContent := remaining[1 : closeIdx+1]

			// Parse header name and value
			colonIdx := strings.Index(headerContent, ":")
			if colonIdx > 0 {
				key := strings.TrimSpace(headerContent[:colonIdx])
				value := strings.TrimSpace(headerContent[colonIdx+1:])

				// Only add if it looks like a valid HTTP header
				// (key contains only alphanumeric, dash, underscore)
				if isValidHeaderName(key) {
					headers[key] = value
				}
			}

			// Move past this header for next iteration
			remaining = remaining[closeIdx+2:]
		}

		// If we found headers in this line, we're done
		if len(headers) > 0 {
			break
		}
	}

	return headers
}

// isValidHeaderName checks if a string is a valid HTTP header name
func isValidHeaderName(name string) bool {
	if len(name) == 0 {
		return false
	}

	for _, ch := range name {
		// HTTP header names can contain: letters, digits, dash, underscore
		// Common headers: Content-Type, X-Forwarded-For, User-Agent, etc.
		if !((ch >= 'A' && ch <= 'Z') ||
			(ch >= 'a' && ch <= 'z') ||
			(ch >= '0' && ch <= '9') ||
			ch == '-' || ch == '_') {
			return false
		}
	}

	return true
}
