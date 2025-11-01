// Copyright (c) 2025 Lazycat Apps
// Licensed under the MIT License. See LICENSE file in the project root for details.

// Package types defines configuration types for the Gosmee Web UI application.
package types

// Config represents the complete application configuration.
type Config struct {
	Server  ServerConfig  // HTTP server configuration
	Gosmee  GosmeeConfig  // Gosmee client management configuration
	CORS    CORSConfig    // CORS policy configuration
	Storage StorageConfig // Storage configuration
	OIDC    OIDCConfig    // OIDC authentication configuration
}

// ServerConfig defines HTTP server listening configuration.
type ServerConfig struct {
	Host string // Server listening address (e.g., "0.0.0.0", "127.0.0.1")
	Port int    // Server listening port (e.g., 8080)
}

// GosmeeConfig defines gosmee client management configuration.
type GosmeeConfig struct {
    MaxClientsPerUser  int   // Maximum number of clients per user (default: 1000)
	MaxStoragePerUser  int64 // Maximum storage per user in bytes (default: 10GB = 10737418240)
	EventRetentionDays int   // Days to retain events (default: 30, 0 = forever)
	LogRetentionDays   int   // Days to retain logs (default: 30, 0 = forever)
	AutoRestart        bool  // Auto restart crashed clients (default: false)
	MaxRestartAttempts int   // Maximum restart attempts (default: 3)
}

// CORSConfig defines Cross-Origin Resource Sharing policy.
type CORSConfig struct {
	AllowedOrigins []string // Allowed origins (e.g., ["*"], ["https://app.example.com"])
}

// StorageConfig defines storage configuration.
type StorageConfig struct {
	DataDir string // Base data directory for all user data (default: "/data")
}

// OIDCConfig defines OIDC authentication configuration.
type OIDCConfig struct {
	ClientID     string // OIDC client ID
	ClientSecret string // OIDC client secret
	Issuer       string // OIDC issuer URL
	RedirectURL  string // OIDC redirect URL after authentication
	Enabled      bool   // Whether OIDC authentication is enabled
}
