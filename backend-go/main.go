package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"agileos-backend/database"
	"agileos-backend/handlers"
	"agileos-backend/messaging"

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
		workflowHandler := handlers.NewWorkflowHandler(db)
		api.POST("/workflow", workflowHandler.CreateWorkflow)
		api.GET("/workflows", workflowHandler.GetWorkflows)
		api.GET("/workflow/:id", workflowHandler.GetWorkflow)

		taskHandler := handlers.NewTaskHandler(db, natsClient)
		api.POST("/task/:id/complete", taskHandler.CompleteTask)
		api.GET("/tasks/pending/:assignedTo", taskHandler.GetPendingTasks)
		api.POST("/process/start", taskHandler.StartProcess)
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
