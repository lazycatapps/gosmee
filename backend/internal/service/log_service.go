// Copyright (c) 2025 Lazycat Apps
// Licensed under the MIT License. See LICENSE file in the project root for details.

package service

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/lazycatapps/gosmee/backend/internal/pkg/logger"
)

// LogService manages log files and streaming.
type LogService struct {
	baseDir string
	log     logger.Logger
}

// NewLogService creates a new log service.
func NewLogService(baseDir string, log logger.Logger) *LogService {
	return &LogService{
		baseDir: baseDir,
		log:     log,
	}
}

// GetLogFile returns the path to a log file for a specific date.
func (s *LogService) getLogFile(userID, clientID, date string) (string, error) {
	// Validate date format
	if _, err := time.Parse("2006-01-02", date); err != nil {
		return "", fmt.Errorf("invalid date format: %s", date)
	}

	logPath := filepath.Join(s.baseDir, "users", userID, "clients", clientID, "logs", fmt.Sprintf("%s.log", date))
	return logPath, nil
}

// GetLogs retrieves log lines from a log file with pagination and search.
func (s *LogService) GetLogs(userID, clientID, date string, page, pageSize int, search string) ([]string, int, error) {
	logPath, err := s.getLogFile(userID, clientID, date)
	if err != nil {
		return nil, 0, err
	}

	// Check if file exists
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		return []string{}, 0, nil
	}

	// Read log file
	file, err := os.Open(logPath)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to open log file: %w", err)
	}
	defer file.Close()

	var allLines []string
	scanner := bufio.NewScanner(file)

	// Read all lines
	for scanner.Scan() {
		line := scanner.Text()

		// Apply search filter
		if search != "" && !strings.Contains(strings.ToLower(line), strings.ToLower(search)) {
			continue
		}

		allLines = append(allLines, line)
	}

	if err := scanner.Err(); err != nil {
		return nil, 0, fmt.Errorf("failed to read log file: %w", err)
	}

	total := len(allLines)

	// Apply pagination
	start := (page - 1) * pageSize
	end := start + pageSize
	if start >= total {
		return []string{}, total, nil
	}
	if end > total {
		end = total
	}

	paged := allLines[start:end]

	return paged, total, nil
}

// GetTodayLogs retrieves today's logs.
func (s *LogService) GetTodayLogs(userID, clientID string, page, pageSize int, search string) ([]string, int, error) {
	today := time.Now().Format("2006-01-02")
	return s.GetLogs(userID, clientID, today, page, pageSize, search)
}

// StreamLogs returns a channel for streaming logs in real-time.
func (s *LogService) StreamLogs(clientID string, processService *ProcessService) (chan string, error) {
	// Get process info
	processInfo, err := processService.GetProcessInfo(clientID)
	if err != nil {
		return nil, fmt.Errorf("client not running: %s", clientID)
	}

	// Add log listener
	logChan := processInfo.AddLogListener()

	return logChan, nil
}

// CleanupOldLogs removes log files older than retention period.
func (s *LogService) CleanupOldLogs(userID, clientID string, retentionDays int) error {
	if retentionDays == 0 {
		return nil // Keep forever
	}

	logsDir := filepath.Join(s.baseDir, "users", userID, "clients", clientID, "logs")

	// Check if logs directory exists
	if _, err := os.Stat(logsDir); os.IsNotExist(err) {
		return nil
	}

	cutoffDate := time.Now().AddDate(0, 0, -retentionDays)

	// Read log files
	files, err := os.ReadDir(logsDir)
	if err != nil {
		return fmt.Errorf("failed to read logs directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		// Parse date from filename (YYYY-MM-DD.log)
		filename := file.Name()
		if !strings.HasSuffix(filename, ".log") {
			continue
		}

		dateStr := strings.TrimSuffix(filename, ".log")
		fileDate, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			continue
		}

		// Delete if older than retention period
		if fileDate.Before(cutoffDate) {
			filePath := filepath.Join(logsDir, filename)
			if err := os.Remove(filePath); err != nil {
				s.log.Error("Failed to delete old log file: %s: %v", filePath, err)
			} else {
				s.log.Info("Deleted old log file: %s", filePath)
			}
		}
	}

	return nil
}

// DownloadLog returns the full log file content for download.
func (s *LogService) DownloadLog(userID, clientID, date string) ([]byte, error) {
	logPath, err := s.getLogFile(userID, clientID, date)
	if err != nil {
		return nil, err
	}

	// Read entire file
	data, err := os.ReadFile(logPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read log file: %w", err)
	}

	return data, nil
}
