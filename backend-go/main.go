package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"agileos-backend/database"
	"agileos-backend/handlers"
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
	// Initialize SurrealDB connection
	dbURL := getEnv("SURREAL_URL", "ws://agileos-db:8000/rpc")
	dbUser := getEnv("SURREAL_USER", "root")
	dbPass := getEnv("SURREAL_PASS", "root")

	var err error
	db, err = database.ConnectDB(dbURL, dbUser, dbPass, "agileos", "main")
	if err != nil {
		log.Fatalf("Failed to connect to SurrealDB: %v", err)
	}
	defer db.Close()

	// Initialize NATS connection
	natsURL := getEnv("NATS_URL", "nats://agileos-nats:4222")
	nc, err = nats.Connect(natsURL,
		nats.RetryOnFailedConnect(true),
		nats.MaxReconnects(10),
		nats.ReconnectWait(2*time.Second),
	)
	if err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}
	defer nc.Close()

	log.Println("✓ Connected to NATS")

	// Initialize NATS Client with orchestration
	natsClient, err = messaging.InitNATS(natsURL, db)
	if err != nil {
		log.Fatalf("Failed to initialize NATS client: %v", err)
	}
	defer natsClient.Close()

	// Subscribe to task events
	if err := natsClient.SubscribeTaskEvents(); err != nil {
		log.Fatalf("Failed to subscribe to task events: %v", err)
	}

	// Start NATS worker in background
	natsClient.StartWorker()

	// Initialize Gin router
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

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

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":         "engine_running",
			"database":       db != nil,
			"message_broker": nc.IsConnected(),
			"timestamp":      time.Now().Unix(),
		})
	})

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
		}
	}

	// Start server
	port := getEnv("PORT", "8080")
	log.Printf("🚀 AgileOS Engine running on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
