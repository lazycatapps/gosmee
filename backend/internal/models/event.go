// Copyright (c) 2025 Lazycat Apps
// Licensed under the MIT License. See LICENSE file in the project root for details.

package models

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// EventStatus represents the forwarding status of an event.
type EventStatus string

const (
	EventStatusSuccess     EventStatus = "success"      // Successfully forwarded
	EventStatusFailed      EventStatus = "failed"       // Forward failed
	EventStatusNotReplayed EventStatus = "not_replayed" // Saved but not forwarded (noReplay mode)
)

// Event represents a webhook event received and forwarded by gosmee.
type Event struct {
	ID           string            `json:"id"`                     // Event ID
	ClientID     string            `json:"clientId"`               // Client instance ID
	Timestamp    time.Time         `json:"timestamp"`              // Event received time
	EventType    string            `json:"eventType"`              // Event type (e.g., "push", "pull_request")
	Source       string            `json:"source"`                 // Event source (e.g., "github.com/myorg/myrepo")
	Status       EventStatus       `json:"status"`                 // Forward status
	StatusCode   int               `json:"statusCode"`             // HTTP status code from target
	LatencyMs    int               `json:"latencyMs"`              // Response latency in milliseconds
	Headers      map[string]string `json:"headers"`                // Request headers
	Payload      string            `json:"payload"`                // Request payload (JSON string)
	Response     string            `json:"response,omitempty"`     // Response body (if available)
	ErrorMessage string            `json:"errorMessage,omitempty"` // Error message (if failed)
}

// UnmarshalJSON implements custom decoding to support multiple event file formats.
func (e *Event) UnmarshalJSON(data []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	e.ID = extractString(raw, "id")
	e.ClientID = firstNonEmptyString(raw, "clientId", "client_id")
	e.EventType = firstNonEmptyString(raw, "eventType", "event_type")
	e.Source = extractString(raw, "source")
	e.Status = EventStatus(firstNonEmptyString(raw, "status", "forward_status"))

	if ts := firstNonEmptyString(raw, "timestamp", "time", "created_at"); ts != "" {
		if parsed, err := time.Parse(time.RFC3339, ts); err == nil {
			e.Timestamp = parsed
		}
	}

	e.StatusCode = firstNonZeroInt(raw, "statusCode", "status_code")
	e.LatencyMs = firstNonZeroInt(raw, "latencyMs", "latency_ms")

	if resp := extractMap(raw, "response"); len(resp) > 0 {
		if e.StatusCode == 0 {
			e.StatusCode = firstNonZeroInt(resp, "status_code")
		}
		if e.LatencyMs == 0 {
			e.LatencyMs = firstNonZeroInt(resp, "latency_ms")
		}
		if e.Response == "" {
			e.Response = stringifyValue(resp)
		}
		if e.ErrorMessage == "" {
			e.ErrorMessage = extractString(resp, "error")
		}
	}

	if e.Status == "" {
		switch {
		case e.StatusCode >= 200 && e.StatusCode < 300:
			e.Status = EventStatusSuccess
		case e.StatusCode > 0:
			e.Status = EventStatusFailed
		}
	}

	if headers := extractStringMap(raw, "headers"); len(headers) > 0 {
		e.Headers = headers
	} else {
		e.Headers = nil
	}

	if payload := stringifyValue(raw["payload"]); payload != "" {
		e.Payload = payload
	}

	if response := stringifyValue(raw["response"]); response != "" && e.Response == "" {
		e.Response = response
	}

	if errMsg := firstNonEmptyString(raw, "errorMessage", "error_message"); errMsg != "" {
		e.ErrorMessage = errMsg
	}

	return nil
}

// EventSummary represents a summarized view of an event (for list queries).
type EventSummary struct {
	ID         string      `json:"id"`
	Timestamp  time.Time   `json:"timestamp"`
	EventType  string      `json:"eventType"`
	Source     string      `json:"source"`
	Status     EventStatus `json:"status"`
	StatusCode int         `json:"statusCode"`
	LatencyMs  int         `json:"latencyMs"`
}

// ToSummary converts an Event to EventSummary.
func (e *Event) ToSummary() *EventSummary {
	return &EventSummary{
		ID:         e.ID,
		Timestamp:  e.Timestamp,
		EventType:  e.EventType,
		Source:     e.Source,
		Status:     e.Status,
		StatusCode: e.StatusCode,
		LatencyMs:  e.LatencyMs,
	}
}

func extractString(data map[string]interface{}, key string) string {
	if data == nil {
		return ""
	}
	if value, ok := data[key]; ok {
		return toString(value)
	}
	return ""
}

func firstNonEmptyString(data map[string]interface{}, keys ...string) string {
	for _, key := range keys {
		if value := extractString(data, key); value != "" {
			return value
		}
	}
	return ""
}

func firstNonZeroInt(data map[string]interface{}, keys ...string) int {
	for _, key := range keys {
		if value, ok := data[key]; ok {
			if num := toInt(value); num != 0 {
				return num
			}
		}
	}
	return 0
}

func toString(value interface{}) string {
	switch v := value.(type) {
	case nil:
		return ""
	case string:
		return strings.TrimSpace(v)
	case fmt.Stringer:
		return strings.TrimSpace(v.String())
	case []byte:
		return strings.TrimSpace(string(v))
	default:
		return strings.TrimSpace(fmt.Sprint(v))
	}
}

func toInt(value interface{}) int {
	switch v := value.(type) {
	case nil:
		return 0
	case int:
		return v
	case int8:
		return int(v)
	case int16:
		return int(v)
	case int32:
		return int(v)
	case int64:
		return int(v)
	case uint:
		return int(v)
	case uint8:
		return int(v)
	case uint16:
		return int(v)
	case uint32:
		return int(v)
	case uint64:
		return int(v)
	case float32:
		return int(v)
	case float64:
		return int(v)
	case json.Number:
		if i, err := v.Int64(); err == nil {
			return int(i)
		}
		if f, err := v.Float64(); err == nil {
			return int(f)
		}
	case string:
		if i, err := strconv.Atoi(strings.TrimSpace(v)); err == nil {
			return i
		}
	}
	return 0
}

func extractMap(data map[string]interface{}, key string) map[string]interface{} {
	if data == nil {
		return nil
	}
	if value, ok := data[key]; ok {
		return toMap(value)
	}
	return nil
}

func toMap(value interface{}) map[string]interface{} {
	switch v := value.(type) {
	case map[string]interface{}:
		return v
	case map[interface{}]interface{}:
		result := make(map[string]interface{}, len(v))
		for key, val := range v {
			result[toString(key)] = val
		}
		return result
	default:
		return nil
	}
}

func extractStringMap(data map[string]interface{}, key string) map[string]string {
	raw := extractMap(data, key)
	if len(raw) == 0 {
		return nil
	}
	result := make(map[string]string, len(raw))
	for k, v := range raw {
		result[k] = toString(v)
	}
	return result
}

func stringifyValue(value interface{}) string {
	switch v := value.(type) {
	case nil:
		return ""
	case string:
		return strings.TrimSpace(v)
	case json.RawMessage:
		if len(v) == 0 {
			return ""
		}
		return strings.TrimSpace(string(v))
	default:
		bytes, err := json.Marshal(v)
		if err != nil {
			return toString(v)
		}
		return string(bytes)
	}
}

// EventListRequest represents query parameters for listing events.
type EventListRequest struct {
	Page      int       `form:"page,default=1"`           // Page number
	PageSize  int       `form:"pageSize,default=20"`      // Items per page
	EventType string    `form:"eventType"`                // Filter by event type
	Status    string    `form:"status"`                   // Filter by status
	Search    string    `form:"search"`                   // Search in source
	DateFrom  time.Time `form:"dateFrom"`                 // Filter by date range (from)
	DateTo    time.Time `form:"dateTo"`                   // Filter by date range (to)
	SortBy    string    `form:"sortBy,default=timestamp"` // Sort field
	SortOrder string    `form:"sortOrder,default=desc"`   // Sort order
}

// EventListResponse represents the response for event list queries.
type EventListResponse struct {
	Total    int             `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"pageSize"`
	Events   []*EventSummary `json:"events"`
}

// EventReplayRequest represents the request body for replaying an event.
type EventReplayRequest struct {
	EventIDs []string `json:"eventIds" binding:"required"` // Event IDs to replay
}

// EventReplayResponse represents the response for event replay.
type EventReplayResponse struct {
	Total      int                  `json:"total"`      // Total events to replay
	Successful int                  `json:"successful"` // Successfully replayed
	Failed     int                  `json:"failed"`     // Failed to replay
	Results    []*EventReplayResult `json:"results"`    // Detailed results
}

// EventReplayResult represents the result of replaying a single event.
type EventReplayResult struct {
	EventID      string `json:"eventId"`
	Success      bool   `json:"success"`
	StatusCode   int    `json:"statusCode,omitempty"`
	LatencyMs    int    `json:"latencyMs,omitempty"`
	ErrorMessage string `json:"errorMessage,omitempty"`
}
