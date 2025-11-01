// Copyright (c) 2025 Lazycat Apps
// Licensed under the MIT License. See LICENSE file in the project root for details.

// Package router provides HTTP routing configuration for the Gosmee Web UI server.
package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lazycatapps/gosmee/backend/internal/handler"
	"github.com/lazycatapps/gosmee/backend/internal/middleware"
	"github.com/lazycatapps/gosmee/backend/internal/types"
)

// Router manages HTTP request routing and handler registration.
type Router struct {
	clientHandler    *handler.ClientHandler
	logHandler       *handler.LogHandler
	eventHandler     *handler.EventHandler
	quotaHandler     *handler.QuotaHandler
	authHandler      *handler.AuthHandler
	sessionValidator middleware.SessionValidator
}

// New creates a new Router instance with the provided handlers.
func New(
	clientHandler *handler.ClientHandler,
	logHandler *handler.LogHandler,
	eventHandler *handler.EventHandler,
	quotaHandler *handler.QuotaHandler,
	authHandler *handler.AuthHandler,
	sessionValidator middleware.SessionValidator,
) *Router {
	return &Router{
		clientHandler:    clientHandler,
		logHandler:       logHandler,
		eventHandler:     eventHandler,
		quotaHandler:     quotaHandler,
		authHandler:      authHandler,
		sessionValidator: sessionValidator,
	}
}

// Setup initializes the Gin engine with middleware and routes.
func (r *Router) Setup(cfg *types.Config) *gin.Engine {
	engine := gin.New()
	engine.Use(gin.Logger())
	engine.Use(gin.Recovery())
	engine.Use(middleware.CORS(cfg.CORS.AllowedOrigins))
	engine.Use(middleware.Auth(cfg.OIDC.Enabled, r.sessionValidator))

	// Disable trusted proxy feature for security
	engine.SetTrustedProxies(nil)

	r.registerRoutes(engine)

	return engine
}

// registerRoutes registers all API routes under /api/v1 prefix.
func (r *Router) registerRoutes(engine *gin.Engine) {
	api := engine.Group("/api/v1")
	{
		// Public endpoints
		api.GET("/health", r.healthCheck)

		// Auth endpoints
		auth := api.Group("/auth")
		{
			auth.GET("/login", r.authHandler.Login)
			auth.GET("/callback", r.authHandler.Callback)
			auth.POST("/logout", r.authHandler.Logout)
			auth.GET("/userinfo", r.authHandler.UserInfo)
		}

		// Protected endpoints (require auth if OIDC enabled)

		// Client management endpoints
		api.POST("/clients", r.clientHandler.Create)
		api.GET("/clients", r.clientHandler.List)
		api.GET("/clients/:id", r.clientHandler.Get)
		api.PUT("/clients/:id", r.clientHandler.Update)
		api.DELETE("/clients/:id", r.clientHandler.Delete)

		// Client control endpoints
		api.POST("/clients/batch/start", r.clientHandler.BatchStart)
		api.POST("/clients/batch/stop", r.clientHandler.BatchStop)
		api.POST("/clients/:id/start", r.clientHandler.Start)
		api.POST("/clients/:id/stop", r.clientHandler.Stop)
		api.POST("/clients/:id/restart", r.clientHandler.Restart)

		// Client stats endpoints
		api.GET("/clients/:id/stats", r.clientHandler.GetStats)

		// Log endpoints
		api.GET("/clients/:id/logs", r.logHandler.GetLogs)
		api.GET("/clients/:id/logs/stream", r.logHandler.StreamLogs)
		api.GET("/clients/:id/logs/download", r.logHandler.DownloadLog)

		// Event endpoints
		api.GET("/clients/:id/events", r.eventHandler.List)
		api.GET("/clients/:id/events/:eventId", r.eventHandler.Get)
		api.DELETE("/clients/:id/events/:eventId", r.eventHandler.Delete)
		api.POST("/clients/:id/events/replay", r.eventHandler.Replay)

		// Quota endpoints
		api.GET("/quota", r.quotaHandler.GetQuota)
	}
}

// healthCheck returns a simple health status.
func (r *Router) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "gosmee-webui",
	})
}
