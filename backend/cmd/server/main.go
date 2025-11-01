// Copyright (c) 2025 Lazycat Apps
// Licensed under the MIT License. See LICENSE file in the project root for details.

// Package main is the entry point for the Gosmee Web UI server application.
package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/lazycatapps/gosmee/backend/internal/handler"
	"github.com/lazycatapps/gosmee/backend/internal/pkg/logger"
	"github.com/lazycatapps/gosmee/backend/internal/repository"
	"github.com/lazycatapps/gosmee/backend/internal/router"
	"github.com/lazycatapps/gosmee/backend/internal/service"
	"github.com/lazycatapps/gosmee/backend/internal/types"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// rootCmd is the root command for the CLI application.
var rootCmd = &cobra.Command{
	Use:   "gosmee-web",
	Short: "Gosmee Web UI - Webhook relay management with web interface",
	Long:  `A web service for managing multiple gosmee client instances.`,
	Run:   runServer,
}

// init initializes command-line flags and environment variable bindings.
func init() {
	rootCmd.Flags().String("host", "0.0.0.0", "Server host")
	rootCmd.Flags().IntP("port", "p", 8080, "Server port")
	rootCmd.Flags().StringSlice("cors-allowed-origins", []string{"*"}, "CORS allowed origins")
	rootCmd.Flags().String("data-dir", "/data", "Base data directory for all user data")

	// Gosmee configuration
	rootCmd.Flags().Int("max-clients-per-user", 1000, "Maximum number of clients per user")
	rootCmd.Flags().Int64("max-storage-per-user", 10737418240, "Maximum storage per user in bytes (default: 10GB)")
	rootCmd.Flags().Int("event-retention-days", 30, "Days to retain events (0 = forever)")
	rootCmd.Flags().Int("log-retention-days", 30, "Days to retain logs (0 = forever)")
	rootCmd.Flags().Bool("auto-restart", false, "Auto restart crashed clients")
	rootCmd.Flags().Int("max-restart-attempts", 3, "Maximum restart attempts")

	// OIDC configuration
	rootCmd.Flags().String("oidc-client-id", "", "OIDC client ID")
	rootCmd.Flags().String("oidc-client-secret", "", "OIDC client secret")
	rootCmd.Flags().String("oidc-issuer", "", "OIDC issuer URL")
	rootCmd.Flags().String("oidc-redirect-url", "", "OIDC redirect URL")

	viper.BindPFlags(rootCmd.Flags())

	// Set environment variable prefix to "GOSMEE"
	viper.SetEnvPrefix("GOSMEE")
	viper.AutomaticEnv()
	// Replace hyphens with underscores in environment variable names
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
}

// runServer is the main server execution function.
func runServer(cmd *cobra.Command, args []string) {
	// Load configuration from viper
	oidcClientID := viper.GetString("oidc-client-id")
	oidcClientSecret := viper.GetString("oidc-client-secret")
	oidcIssuer := viper.GetString("oidc-issuer")
	oidcRedirectURL := viper.GetString("oidc-redirect-url")

	cfg := &types.Config{
		Server: types.ServerConfig{
			Host: viper.GetString("host"),
			Port: viper.GetInt("port"),
		},
		Gosmee: types.GosmeeConfig{
			MaxClientsPerUser:  viper.GetInt("max-clients-per-user"),
			MaxStoragePerUser:  viper.GetInt64("max-storage-per-user"),
			EventRetentionDays: viper.GetInt("event-retention-days"),
			LogRetentionDays:   viper.GetInt("log-retention-days"),
			AutoRestart:        viper.GetBool("auto-restart"),
			MaxRestartAttempts: viper.GetInt("max-restart-attempts"),
		},
		CORS: types.CORSConfig{
			AllowedOrigins: viper.GetStringSlice("cors-allowed-origins"),
		},
		Storage: types.StorageConfig{
			DataDir: viper.GetString("data-dir"),
		},
		OIDC: types.OIDCConfig{
			ClientID:     oidcClientID,
			ClientSecret: oidcClientSecret,
			Issuer:       oidcIssuer,
			RedirectURL:  oidcRedirectURL,
			Enabled:      oidcClientID != "" && oidcClientSecret != "" && oidcIssuer != "",
		},
	}

	// Initialize logger
	log := logger.New()

	log.Info("Starting Gosmee Web UI server")
	log.Info("=================================")

	// Log configuration
	log.Info("Gosmee Configuration:")
	log.Info("  Max Clients Per User: %d", cfg.Gosmee.MaxClientsPerUser)
	log.Info("  Max Storage Per User: %d bytes (%.2f GB)", cfg.Gosmee.MaxStoragePerUser, float64(cfg.Gosmee.MaxStoragePerUser)/1024/1024/1024)
	log.Info("  Event Retention: %d days", cfg.Gosmee.EventRetentionDays)
	log.Info("  Log Retention: %d days", cfg.Gosmee.LogRetentionDays)
	log.Info("  Auto Restart: %v", cfg.Gosmee.AutoRestart)

	// Log OIDC configuration status
	if cfg.OIDC.Enabled {
		log.Info("OIDC authentication: ENABLED")
		log.Info("  Issuer: %s", cfg.OIDC.Issuer)
		log.Info("  Client ID: %s", cfg.OIDC.ClientID)
		log.Info("  Redirect URL: %s", cfg.OIDC.RedirectURL)
	} else {
		log.Info("OIDC authentication: DISABLED")
	}

	// Initialize repositories
	log.Info("Initializing repositories...")
	log.Info("  Data directory: %s", cfg.Storage.DataDir)

	clientRepo, err := repository.NewFileClientRepository(cfg.Storage.DataDir)
	if err != nil {
		log.Error("Failed to initialize client repository: %v", err)
		return
	}

	eventRepo := repository.NewFileEventRepository(cfg.Storage.DataDir)
	quotaRepo := repository.NewFileQuotaRepository(
		cfg.Storage.DataDir,
		cfg.Gosmee.MaxStoragePerUser,
		cfg.Gosmee.MaxClientsPerUser,
	)

	log.Info("Repositories initialized successfully")

	// Initialize services
	processService := service.NewProcessService(cfg.Gosmee.AutoRestart, cfg.Gosmee.MaxRestartAttempts, log)
	clientService := service.NewClientService(clientRepo, quotaRepo, eventRepo, processService, cfg.Storage.DataDir, log)
	logService := service.NewLogService(cfg.Storage.DataDir, log)
	eventService := service.NewEventService(eventRepo, clientRepo, log)
	quotaService := service.NewQuotaService(quotaRepo, log)
	sessionService := service.NewSessionService(7 * 24 * time.Hour) // 7 days session TTL

	// Initialize HTTP handlers
	clientHandler := handler.NewClientHandler(clientService, quotaService, log)
	logHandler := handler.NewLogHandler(logService, processService, log)
	eventHandler := handler.NewEventHandler(eventService, log)
	quotaHandler := handler.NewQuotaHandler(quotaService, log)

	// Initialize auth handler
	authHandler, err := handler.NewAuthHandler(&cfg.OIDC, sessionService, log)
	if err != nil {
		log.Error("Failed to initialize auth handler: %v", err)
		return
	}

	// Set up router and middleware
	r := router.New(clientHandler, logHandler, eventHandler, quotaHandler, authHandler, sessionService)
	engine := r.Setup(cfg)

	// Set up graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Start HTTP server in goroutine
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Info("=================================")
	log.Info("Server listening on %s", addr)
	log.Info("Press Ctrl+C to stop")

	go func() {
		if err := engine.Run(addr); err != nil {
			log.Error("Server failed: %v", err)
			quit <- syscall.SIGTERM
		}
	}()

	// Wait for interrupt signal
	<-quit
	log.Info("Shutting down server...")

	// Stop all running processes
	processService.StopAll()

	log.Info("Goodbye!")
}

// main is the application entry point.
func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
