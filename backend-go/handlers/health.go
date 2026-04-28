package handlers

import (
	"net/http"
	"time"

	"agileos-backend/database"
	"agileos-backend/models"

	"github.com/gin-gonic/gin"
	"github.com/nats-io/nats.go"
)

var startTime = time.Now()

type HealthHandler struct {
	db   *database.SurrealDB
	nats *nats.Conn
}

func NewHealthHandler(db *database.SurrealDB, nats *nats.Conn) *HealthHandler {
	return &HealthHandler{
		db:   db,
		nats: nats,
	}
}

// GetHealth performs comprehensive health checks
func (h *HealthHandler) GetHealth(c *gin.Context) {
	health := models.SystemHealth{
		Timestamp: time.Now(),
		Uptime:    int64(time.Since(startTime).Seconds()),
		Version:   "1.0.0",
	}

	// Check database
	dbHealth := h.checkDatabase()
	health.Database = dbHealth

	// Check message broker
	natsHealth := h.checkNATS()
	health.MessageBroker = natsHealth

	// Determine overall status
	if dbHealth.Status == "up" && natsHealth.Status == "up" {
		health.Status = "healthy"
		c.JSON(http.StatusOK, health)
	} else if dbHealth.Status == "down" || natsHealth.Status == "down" {
		health.Status = "unhealthy"
		c.JSON(http.StatusServiceUnavailable, health)
	} else {
		health.Status = "degraded"
		c.JSON(http.StatusOK, health)
	}
}

// GetHealthLive is a lightweight liveness probe
func (h *HealthHandler) GetHealthLive(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "alive",
		"timestamp": time.Now().Unix(),
	})
}

// GetHealthReady checks if the service is ready to accept traffic
func (h *HealthHandler) GetHealthReady(c *gin.Context) {
	dbHealth := h.checkDatabase()
	natsHealth := h.checkNATS()

	if dbHealth.Status == "up" && natsHealth.Status == "up" {
		c.JSON(http.StatusOK, gin.H{
			"status": "ready",
			"database": "connected",
			"message_broker": "connected",
		})
	} else {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "not_ready",
			"database": dbHealth.Status,
			"message_broker": natsHealth.Status,
		})
	}
}

// checkDatabase performs database health check
func (h *HealthHandler) checkDatabase() models.HealthCheck {
	start := time.Now()
	
	// Try a simple query
	_, err := h.db.Query("SELECT 1", nil)
	responseTime := time.Since(start).Milliseconds()

	if err != nil {
		return models.HealthCheck{
			Status:       "down",
			ResponseTime: responseTime,
			Message:      "Database connection failed",
			Details: map[string]interface{}{
				"error": err.Error(),
			},
		}
	}

	if responseTime > 1000 {
		return models.HealthCheck{
			Status:       "degraded",
			ResponseTime: responseTime,
			Message:      "Database response time is slow",
		}
	}

	return models.HealthCheck{
		Status:       "up",
		ResponseTime: responseTime,
		Message:      "Database is healthy",
	}
}

// checkNATS performs NATS health check
func (h *HealthHandler) checkNATS() models.HealthCheck {
	start := time.Now()
	
	if h.nats == nil {
		return models.HealthCheck{
			Status:       "down",
			ResponseTime: 0,
			Message:      "NATS connection is nil",
		}
	}

	if !h.nats.IsConnected() {
		return models.HealthCheck{
			Status:       "down",
			ResponseTime: time.Since(start).Milliseconds(),
			Message:      "NATS is not connected",
		}
	}

	responseTime := time.Since(start).Milliseconds()

	return models.HealthCheck{
		Status:       "up",
		ResponseTime: responseTime,
		Message:      "NATS is healthy",
		Details: map[string]interface{}{
			"servers": h.nats.ConnectedUrl(),
		},
	}
}
