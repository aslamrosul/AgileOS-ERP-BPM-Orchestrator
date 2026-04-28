package main

import (
	"os"
	"time"

	"agileos-backend/analytics"
	"agileos-backend/database"
	"agileos-backend/handlers"
	"agileos-backend/logger"
	"agileos-backend/messaging"
	"agileos-backend/middleware"

	"github.com/gin-gonic/gin"
	"github.com/nats-io/nats.go"
)

var (
	db         *database.SurrealDB
	nc         *nats.Conn
	natsClient *messaging.NATSClient
)

func main() {
	// Initialize structured logging
	logLevel := getEnv("LOG_LEVEL", "info")
	logToFile := getEnv("LOG_TO_FILE", "true") == "true"
	logFilePath := getEnv("LOG_FILE_PATH", "./logs/agileos.log")
	
	if err := logger.InitLogger(logLevel, logToFile, logFilePath); err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}

	logger.Log.Info().Msg("🚀 Starting AgileOS BPM Engine...")

	// Initialize SurrealDB connection
	dbURL := getEnv("SURREAL_URL", "ws://agileos-db:8000/rpc")
	dbUser := getEnv("SURREAL_USER", "root")
	dbPass := getEnv("SURREAL_PASS", "root")

	var err error
	db, err = database.ConnectDB(dbURL, dbUser, dbPass, "agileos", "main")
	if err != nil {
		logger.LogFatal("Failed to connect to SurrealDB", err, map[string]interface{}{
			"url": dbURL,
		})
	}
	defer db.Close()

	logger.Log.Info().Str("url", dbURL).Msg("✓ Connected to SurrealDB")

	// Initialize NATS connection
	natsURL := getEnv("NATS_URL", "nats://agileos-nats:4222")
	nc, err = nats.Connect(natsURL,
		nats.RetryOnFailedConnect(true),
		nats.MaxReconnects(10),
		nats.ReconnectWait(2*time.Second),
	)
	if err != nil {
		logger.LogFatal("Failed to connect to NATS", err, map[string]interface{}{
			"url": natsURL,
		})
	}
	defer nc.Close()

	logger.Log.Info().Str("url", natsURL).Msg("✓ Connected to NATS")

	// Initialize NATS Client with orchestration
	natsClient, err = messaging.InitNATS(natsURL, db)
	if err != nil {
		logger.LogFatal("Failed to initialize NATS client", err, nil)
	}
	defer natsClient.Close()

	// Subscribe to task events
	if err := natsClient.SubscribeTaskEvents(); err != nil {
		logger.LogFatal("Failed to subscribe to task events", err, nil)
	}

	// Start NATS worker in background
	natsClient.StartWorker()

	// Initialize Gin router
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// Logging middleware
	r.Use(func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		duration := time.Since(start)
		statusCode := c.Writer.Status()

		logger.Log.Info().
			Str("method", method).
			Str("path", path).
			Int("status", statusCode).
			Dur("duration_ms", duration).
			Str("ip", c.ClientIP()).
			Msg("HTTP Request")
	})

	// CORS middleware
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		
		c.Next()
	})

	// Health check endpoints
	healthHandler := handlers.NewHealthHandler(db, nc)
	r.GET("/health", healthHandler.GetHealth)
	r.GET("/health/live", healthHandler.GetHealthLive)
	r.GET("/health/ready", healthHandler.GetHealthReady)

	// API routes
	api := r.Group("/api/v1")
	{
		// Public routes (no authentication required)
		authHandler := handlers.NewAuthHandler(db)
		api.POST("/auth/login", authHandler.Login)
		api.POST("/auth/register", authHandler.Register)
		api.POST("/auth/refresh", authHandler.RefreshToken)

		// Protected routes (authentication required)
		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware())
		{
			// User profile
			protected.GET("/auth/profile", authHandler.GetProfile)

			// Workflow routes (all authenticated users)
			workflowHandler := handlers.NewWorkflowHandler(db)
			protected.GET("/workflows", workflowHandler.GetWorkflows)
			protected.GET("/workflow/:id", workflowHandler.GetWorkflow)

			// Workflow creation (admin only)
			protected.POST("/workflow", middleware.AuthorizeRole("admin"), workflowHandler.CreateWorkflow)

			// Task routes (all authenticated users)
			taskHandler := handlers.NewTaskHandler(db, natsClient)
			protected.GET("/tasks/pending/:assignedTo", taskHandler.GetPendingTasks)
			protected.POST("/task/:id/complete", taskHandler.CompleteTask)

			// Process routes (manager and above)
			protected.POST("/process/start", middleware.AuthorizeRole("admin", "manager"), taskHandler.StartProcess)

			// Admin routes (admin only)
			admin := protected.Group("")
			admin.Use(middleware.AuthorizeRole("admin"))
			{
				admin.GET("/users", authHandler.ListUsers)
			}

			// Analytics routes (manager and admin)
			analyticsService := analytics.NewService(db)
			analyticsHandler := handlers.NewAnalyticsHandler(analyticsService)
			
			analytics := protected.Group("/analytics")
			analytics.Use(middleware.AuthorizeRole("admin", "manager"))
			{
				analytics.GET("/overview", analyticsHandler.GetOverview)
				analytics.GET("/workflows", analyticsHandler.GetWorkflowEfficiency)
				analytics.GET("/steps", analyticsHandler.GetStepPerformance)
				analytics.GET("/departments", analyticsHandler.GetDepartmentMetrics)
				analytics.GET("/summary", analyticsHandler.GetSummary)
				analytics.GET("/insights", analyticsHandler.GetInsights)
			}

			// Digital Signature routes (all authenticated users)
			signatureHandler := handlers.NewSignatureHandler(db)
			signature := protected.Group("/signature")
			{
				signature.POST("/verify", signatureHandler.VerifySignature)
				signature.GET("/task/:id", signatureHandler.GetTaskSignature)
				signature.GET("/task/:id/integrity", signatureHandler.VerifyTaskIntegrity)
				signature.GET("/task/:id/receipt", signatureHandler.GenerateTaskReceipt)
			}
		}
	}

	// Start server
	port := getEnv("PORT", "8080")
	logger.Log.Info().Str("port", port).Msg("🚀 AgileOS Engine running")
	if err := r.Run(":" + port); err != nil {
		logger.LogFatal("Failed to start server", err, map[string]interface{}{
			"port": port,
		})
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
