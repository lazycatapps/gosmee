// Copyright (c) 2025 Lazycat Apps
// Licensed under the MIT License. See LICENSE file in the project root for details.

package service

import (
	"bufio"
	"fmt"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/lazycatapps/gosmee/backend/internal/models"
	"github.com/lazycatapps/gosmee/backend/internal/pkg/logger"
)

// ProcessService manages gosmee client processes.
type ProcessService struct {
	processes       map[string]*processContext // clientID -> process context
	mu              sync.RWMutex               // Mutex for thread-safe operations
	log             logger.Logger
	autoRestart     bool
	maxRestartCount int
}

// processContext holds information about a running process.
type processContext struct {
	client       *models.Client
	cmd          *exec.Cmd
	processInfo  *models.ProcessInfo
	stopChan     chan struct{}
	restartCount int
}

// NewProcessService creates a new process service.
func NewProcessService(autoRestart bool, maxRestartCount int, log logger.Logger) *ProcessService {
	return &ProcessService{
		processes:       make(map[string]*processContext),
		log:             log,
		autoRestart:     autoRestart,
		maxRestartCount: maxRestartCount,
	}
}

// Start starts a gosmee client process.
func (s *ProcessService) Start(client *models.Client, baseDir string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if already running
	if ctx, exists := s.processes[client.ID]; exists {
		if ctx.cmd.Process != nil {
			return fmt.Errorf("client already running: %s", client.ID)
		}
	}

	// Build gosmee command
	cmd, err := s.buildGosmeeCommand(client, baseDir)
	if err != nil {
		return fmt.Errorf("failed to build gosmee command: %w", err)
	}

	// Create pipes for stdout/stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Start the process
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start gosmee process: %w", err)
	}

	// Create process info
	processInfo := models.NewProcessInfo(client.ID, cmd.Process.Pid)

	// Create process context
	ctx := &processContext{
		client:      client,
		cmd:         cmd,
		processInfo: processInfo,
		stopChan:    make(chan struct{}),
	}

	s.processes[client.ID] = ctx

	// Start log collectors
	go s.collectLogs(ctx, stdout, "stdout")
	go s.collectLogs(ctx, stderr, "stderr")

	// Start process monitor
	go s.monitorProcess(ctx)

	s.log.Info("Started gosmee client process: %s (PID: %d)", client.ID, cmd.Process.Pid)

	return nil
}

// Stop stops a gosmee client process.
func (s *ProcessService) Stop(clientID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	ctx, exists := s.processes[clientID]
	if !exists {
		return fmt.Errorf("client not running: %s", clientID)
	}

	// Signal stop
	close(ctx.stopChan)

	// Try graceful shutdown (SIGTERM)
	if ctx.cmd.Process != nil {
		if err := ctx.cmd.Process.Signal(syscall.SIGTERM); err != nil {
			s.log.Error("Failed to send SIGTERM to process %d: %v", ctx.cmd.Process.Pid, err)
		}

		// Wait for graceful shutdown (5 seconds timeout)
		done := make(chan error, 1)
		go func() {
			done <- ctx.cmd.Wait()
		}()

		select {
		case <-done:
			s.log.Info("Process %d terminated gracefully", ctx.cmd.Process.Pid)
		case <-time.After(5 * time.Second):
			// Force kill if not stopped
			s.log.Info("Process %d did not stop gracefully, force killing", ctx.cmd.Process.Pid)
			ctx.cmd.Process.Kill()
		}
	}

	// Close log listeners
	ctx.processInfo.CloseAllLogListeners()

	// Remove from map
	delete(s.processes, clientID)

	s.log.Info("Stopped gosmee client process: %s", clientID)

	return nil
}

// Restart restarts a gosmee client process.
func (s *ProcessService) Restart(client *models.Client, baseDir string) error {
	// Stop first
	s.Stop(client.ID)

	// Wait a moment
	time.Sleep(500 * time.Millisecond)

	// Start again
	return s.Start(client, baseDir)
}

// GetProcessInfo returns process information for a client.
func (s *ProcessService) GetProcessInfo(clientID string) (*models.ProcessInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ctx, exists := s.processes[clientID]
	if !exists {
		return nil, fmt.Errorf("client not running: %s", clientID)
	}

	return ctx.processInfo, nil
}

// IsRunning checks if a client process is running.
func (s *ProcessService) IsRunning(clientID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ctx, exists := s.processes[clientID]
	if !exists {
		return false
	}

	return ctx.cmd.Process != nil
}

// StopAll stops all running processes.
func (s *ProcessService) StopAll() {
	s.mu.Lock()
	clientIDs := make([]string, 0, len(s.processes))
	for clientID := range s.processes {
		clientIDs = append(clientIDs, clientID)
	}
	s.mu.Unlock()

	for _, clientID := range clientIDs {
		s.Stop(clientID)
	}
}

// buildGosmeeCommand builds the gosmee command with all parameters.
func (s *ProcessService) buildGosmeeCommand(client *models.Client, baseDir string) (*exec.Cmd, error) {
	args := []string{"client"}

	// Add target connection timeout
	if client.TargetTimeout > 0 {
		args = append(args, "--target-connection-timeout", fmt.Sprintf("%d", client.TargetTimeout))
	}

	// Add save directory
	eventsDir := filepath.Join(baseDir, "users", client.UserID, "clients", client.ID, "events")
	args = append(args, "--saveDir", eventsDir)

	// Add HTTPie flag if enabled
	if client.HTTPie {
		args = append(args, "--httpie")
	}

	// Add ignore events
	for _, event := range client.IgnoreEvents {
		args = append(args, "--ignore-event", event)
	}

	// Add noReplay flag if enabled
	if client.NoReplay {
		args = append(args, "--noReplay")
	}

	// Add SSE buffer size
	if client.SSEBufferSize > 0 {
		args = append(args, "--sse-buffer-size", fmt.Sprintf("%d", client.SSEBufferSize))
	}

	// Add Smee URL and Target URL (positional arguments)
	args = append(args, client.SmeeURL, client.TargetURL)

	cmd := exec.Command("gosmee", args...)

	return cmd, nil
}

// collectLogs collects logs from stdout/stderr and broadcasts to listeners.
func (s *ProcessService) collectLogs(ctx *processContext, pipe interface{}, source string) {
	scanner := bufio.NewScanner(pipe.(interface{ Read([]byte) (int, error) }))

	for scanner.Scan() {
		line := scanner.Text()
		timestamp := time.Now().Format("2006-01-02 15:04:05")
		logLine := fmt.Sprintf("[%s] [%s] %s", timestamp, source, line)

		// Add to process info
		ctx.processInfo.AddLog(logLine)

		// Also log to application logger
		s.log.Debug("[Client %s] %s", ctx.client.ID, logLine)
	}

	if err := scanner.Err(); err != nil {
		s.log.Error("Error reading logs from %s: %v", source, err)
	}
}

// monitorProcess monitors the process and handles restarts.
func (s *ProcessService) monitorProcess(ctx *processContext) {
	// Wait for process to finish
	err := ctx.cmd.Wait()

	// Check if it was a normal stop
	select {
	case <-ctx.stopChan:
		// Normal stop, don't restart
		s.log.Info("Client %s stopped normally", ctx.client.ID)
		return
	default:
	}

	// Process crashed
	if err != nil {
		s.log.Error("Client %s process crashed: %v", ctx.client.ID, err)
		ctx.processInfo.LastError = err.Error()
		ctx.processInfo.Status = models.ClientStatusError
	}

	// Auto restart if enabled
	if s.autoRestart && ctx.restartCount < s.maxRestartCount {
		ctx.restartCount++
		s.log.Info("Auto-restarting client %s (attempt %d/%d)", ctx.client.ID, ctx.restartCount, s.maxRestartCount)

		// Wait a moment before restart
		time.Sleep(2 * time.Second)

		// Restart (this requires client object and baseDir, which we need to pass through)
		// For now, we'll just log - actual restart should be triggered from ClientService
		s.log.Info("Auto-restart not implemented yet - please restart manually")
	}
}
