// Copyright (c) 2025 Lazycat Apps
// Licensed under the MIT License. See LICENSE file in the project root for details.

package models

import (
	"sync"
	"time"
)

// ProcessInfo represents gosmee client process information.
type ProcessInfo struct {
	ClientID     string       `json:"clientId"`
	PID          int          `json:"pid"`
	Status       ClientStatus `json:"status"`
	StartedAt    time.Time    `json:"startedAt"`
	RestartCount int          `json:"restartCount"`
	LastError    string       `json:"lastError,omitempty"`

	// Log streaming
	LogLines     []string      `json:"-"` // In-memory log lines (not serialized)
	LogListeners []chan string `json:"-"` // Active log stream subscribers (SSE)
	logMu        sync.Mutex    // Mutex for thread-safe log operations
}

// NewProcessInfo creates a new ProcessInfo instance.
func NewProcessInfo(clientID string, pid int) *ProcessInfo {
	return &ProcessInfo{
		ClientID:     clientID,
		PID:          pid,
		Status:       ClientStatusRunning,
		StartedAt:    time.Now(),
		RestartCount: 0,
		LogLines:     []string{},
		LogListeners: []chan string{},
	}
}

// AddLog appends a log line to the process and broadcasts it to all active listeners.
// Thread-safe for concurrent access.
func (p *ProcessInfo) AddLog(line string) {
	p.logMu.Lock()
	defer p.logMu.Unlock()

	p.LogLines = append(p.LogLines, line)

	// Broadcast to all SSE listeners
	for _, ch := range p.LogListeners {
		select {
		case ch <- line:
			// Successfully sent
		default:
			// Channel is full or closed, skip this listener
		}
	}
}

// AddLogListener creates a new log listener channel for SSE streaming.
// Returns a buffered channel (100 messages) that will receive new log lines.
func (p *ProcessInfo) AddLogListener() chan string {
	p.logMu.Lock()
	defer p.logMu.Unlock()

	ch := make(chan string, 100)
	p.LogListeners = append(p.LogListeners, ch)
	return ch
}

// RemoveLogListener removes and closes a log listener channel.
// Should be called when an SSE client disconnects.
func (p *ProcessInfo) RemoveLogListener(ch chan string) {
	p.logMu.Lock()
	defer p.logMu.Unlock()

	for i, listener := range p.LogListeners {
		if listener == ch {
			p.LogListeners = append(p.LogListeners[:i], p.LogListeners[i+1:]...)
			close(ch)
			break
		}
	}
}

// CloseAllLogListeners closes all active log listener channels.
// Called when process stops to notify all SSE clients.
func (p *ProcessInfo) CloseAllLogListeners() {
	p.logMu.Lock()
	defer p.logMu.Unlock()

	for _, ch := range p.LogListeners {
		close(ch)
	}
	p.LogListeners = []chan string{}
}

// GetLogLines returns a copy of all log lines.
// Thread-safe for concurrent access.
func (p *ProcessInfo) GetLogLines() []string {
	p.logMu.Lock()
	defer p.logMu.Unlock()

	logs := make([]string, len(p.LogLines))
	copy(logs, p.LogLines)
	return logs
}

// ClientStats represents statistics for a client instance.
type ClientStats struct {
	RunningTime      int64      `json:"runningTime"`      // Running time in seconds
	TodayEvents      int        `json:"todayEvents"`      // Events today
	TotalEvents      int        `json:"totalEvents"`      // Total events
	SuccessRate      float64    `json:"successRate"`      // Success rate percentage
	AverageLatency   int        `json:"averageLatency"`   // Average response latency in ms
	SSEConnected     bool       `json:"sseConnected"`     // SSE connection status
	ReconnectCount   int        `json:"reconnectCount"`   // SSE reconnect count
	LastEventTime    *time.Time `json:"lastEventTime,omitempty"` // Last event time
}
