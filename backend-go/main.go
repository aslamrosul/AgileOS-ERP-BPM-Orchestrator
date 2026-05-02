package main

import (
	"os"
	"time"

	"agileos-backend/analytics"
	"agileos-backend/database"
	"agileos-backend/handlers"
	"agileos-backend/internal/audit"
	"agileos-backend/internal/ws"
	"agileos-backend/logger"
	"agileos-backend/messaging"
	"agileos-backend/middleware"

	"github.com/gin-gonic/gin"
	"github.com/nats-io/nats.go"
	
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "agileos-backend/docs" // Import generated docs
)

// @title AgileOS BPM API
// @version 1.0
// @description Enterprise Business Process Management System with Workflow Engine, Real-time Notifications, and Analytics
// @termsOfService https://agileos.com/terms

// @contact.name API Support
// @contact.url https://agileos.com/support
// @contact.email support@agileos.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

var (
	db         *database.SurrealDB
	nc         *nats.Conn
	natsClient *messaging.NATSClient
	wsHub      *ws.Hub
	notifier   *ws.Notifier
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

	// Initialize security middleware
	middleware.InitRateLimiters()
	middleware.InitIPFilter()

	// Initialize SurrealDB connection
	dbURL := getEnv("SURREAL_URL", "ws://localhost:8002/rpc")
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
	natsURL := getEnv("NATS_URL", "nats://localhost:4223")
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

	// Initialize WebSocket Hub
	wsHub = ws.NewHub()
	go wsHub.Run()

	// Initialize NATS to WebSocket notifier
	notifier = ws.NewNotifier(wsHub, natsClient)
	if err := notifier.Start(); err != nil {
		logger.LogFatal("Failed to start WebSocket notifier", err, nil)
	}

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

	// Security middleware
	r.Use(middleware.SecurityHeaders())
	r.Use(middleware.CORSSecure())
	r.Use(middleware.IPFilterMiddleware())
	r.Use(middleware.GlobalRateLimit())

	// Health check endpoints
	healthHandler := handlers.NewHealthHandler(db, nc)
	r.GET("/health", healthHandler.GetHealth)
	r.GET("/health/live", healthHandler.GetHealthLive)
	r.GET("/health/ready", healthHandler.GetHealthReady)

	// Swagger documentation
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// WebSocket endpoint
	r.GET("/ws", func(c *gin.Context) {
		ws.ServeWS(wsHub, c)
	})

	// API routes
	api := r.Group("/api/v1")
	{
		// Public routes with strict auth rate limiting
		authHandler := handlers.NewAuthHandler(db)
		
		// Apply auth rate limiter to login/register endpoints
		authGroup := api.Group("/auth")
		authGroup.Use(middleware.AuthRateLimit())
		{
			authGroup.POST("/login", authHandler.Login)
			authGroup.POST("/register", authHandler.Register)
		}
		
		// Refresh token endpoint (less strict)
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

			// AI Analytics routes (manager and admin) - Python FastAPI integration
			aiAnalyticsHandler := handlers.NewAIAnalyticsHandler()
			
			aiAnalytics := protected.Group("/ai-analytics")
			aiAnalytics.Use(middleware.AuthorizeRole("admin", "manager"))
			{
				aiAnalytics.GET("/status", aiAnalyticsHandler.GetAIServiceStatus)
				aiAnalytics.GET("/predict/workflow/:workflow_id", aiAnalyticsHandler.GetWorkflowPrediction)
				aiAnalytics.GET("/anomalies", aiAnalyticsHandler.GetAnomalies)
				aiAnalytics.GET("/comprehensive", aiAnalyticsHandler.GetComprehensiveAIAnalytics)
				aiAnalytics.GET("/workflow/:workflow_id/performance", aiAnalyticsHandler.GetWorkflowPerformanceAI)
				aiAnalytics.POST("/refresh-cache", aiAnalyticsHandler.RefreshAICache)
			}

			// Audit & Governance routes (admin and manager)
			auditService := audit.NewAuditService(db)
			auditHandler := handlers.NewAuditHandler(auditService)
			
			auditRoutes := protected.Group("/audit")
			auditRoutes.Use(middleware.AuthorizeRole("admin", "manager"))
			{
				auditRoutes.GET("/trails", auditHandler.GetAuditTrails)
				auditRoutes.GET("/violations", auditHandler.GetComplianceViolations)
				auditRoutes.GET("/export", auditHandler.ExportAuditTrails)
				auditRoutes.GET("/statistics", auditHandler.GetAuditStatistics)
				auditRoutes.GET("/workflow/:workflow_id/versions", auditHandler.GetWorkflowVersionHistory)
				auditRoutes.POST("/workflow/version", auditHandler.CreateWorkflowVersion)
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

			// Accounting routes (admin, manager, finance)
			accountingHandler := handlers.NewAccountingHandler(db)
			accounting := protected.Group("/accounting")
			accounting.Use(middleware.AuthorizeRole("admin", "manager", "finance"))
			{
				// Chart of Accounts
				accounting.POST("/accounts", accountingHandler.CreateAccount)
				accounting.GET("/accounts", accountingHandler.GetAccounts)
				accounting.GET("/accounts/tree", accountingHandler.GetAccountTree)
				accounting.GET("/accounts/:id", accountingHandler.GetAccount)
				accounting.PUT("/accounts/:id", accountingHandler.UpdateAccount)
				accounting.DELETE("/accounts/:id", accountingHandler.DeleteAccount)

				// Journal Entries
				accounting.POST("/journal-entries", accountingHandler.CreateJournalEntry)
				accounting.GET("/journal-entries", accountingHandler.GetJournalEntries)
				accounting.GET("/journal-entries/:id", accountingHandler.GetJournalEntry)
				accounting.POST("/journal-entries/:id/post", accountingHandler.PostJournalEntry)
				accounting.POST("/journal-entries/:id/reverse", accountingHandler.ReverseJournalEntry)
				accounting.DELETE("/journal-entries/:id", accountingHandler.DeleteJournalEntry)

				// Vendors
				accounting.POST("/vendors", accountingHandler.CreateVendor)
				accounting.GET("/vendors", accountingHandler.GetVendors)
				accounting.GET("/vendors/:id", accountingHandler.GetVendor)
				accounting.PUT("/vendors/:id", accountingHandler.UpdateVendor)
				accounting.DELETE("/vendors/:id", accountingHandler.DeleteVendor)

				// Purchase Invoices
				accounting.POST("/purchase-invoices", accountingHandler.CreatePurchaseInvoice)
				accounting.GET("/purchase-invoices", accountingHandler.GetPurchaseInvoices)
				accounting.GET("/purchase-invoices/:id", accountingHandler.GetPurchaseInvoice)
				accounting.PUT("/purchase-invoices/:id", accountingHandler.UpdatePurchaseInvoice)
				accounting.DELETE("/purchase-invoices/:id", accountingHandler.DeletePurchaseInvoice)
				accounting.POST("/purchase-invoices/:id/approve", accountingHandler.ApprovePurchaseInvoice)
				accounting.POST("/purchase-invoices/:id/cancel", accountingHandler.CancelPurchaseInvoice)

				// Account Receivable (AR) - Customers & Sales Invoices
				accounting.POST("/customers", accountingHandler.CreateCustomer)
				accounting.GET("/customers", accountingHandler.GetCustomers)
				accounting.GET("/customers/:id", accountingHandler.GetCustomer)
				accounting.PUT("/customers/:id", accountingHandler.UpdateCustomer)
				accounting.DELETE("/customers/:id", accountingHandler.DeleteCustomer)

				accounting.POST("/sales-invoices", accountingHandler.CreateSalesInvoice)
				accounting.GET("/sales-invoices", accountingHandler.GetSalesInvoices)
				accounting.GET("/sales-invoices/:id", accountingHandler.GetSalesInvoice)
				accounting.PUT("/sales-invoices/:id", accountingHandler.UpdateSalesInvoice)
				accounting.DELETE("/sales-invoices/:id", accountingHandler.DeleteSalesInvoice)
				accounting.POST("/sales-invoices/:id/approve", accountingHandler.ApproveSalesInvoice)
				accounting.POST("/sales-invoices/:id/cancel", accountingHandler.CancelSalesInvoice)

				// Financial Reports
				accounting.GET("/reports/balance-sheet", accountingHandler.GetBalanceSheet)
				accounting.GET("/reports/profit-loss", accountingHandler.GetProfitLoss)
				accounting.GET("/reports/cash-flow", accountingHandler.GetCashFlow)
				accounting.GET("/reports/trial-balance", accountingHandler.GetTrialBalance)
				accounting.GET("/reports/general-ledger", accountingHandler.GetGeneralLedger)
				accounting.GET("/reports/ar-aging", accountingHandler.GetARAgingReport)
				accounting.GET("/reports/ap-aging", accountingHandler.GetAPAgingReport)

				// Budget Management
				accounting.POST("/budgets", accountingHandler.CreateBudget)
				accounting.GET("/budgets", accountingHandler.GetBudgets)
				accounting.GET("/budgets/:id", accountingHandler.GetBudget)
				accounting.PUT("/budgets/:id", accountingHandler.UpdateBudget)
				accounting.DELETE("/budgets/:id", accountingHandler.DeleteBudget)
				accounting.POST("/budgets/:id/approve", accountingHandler.ApproveBudget)
				accounting.GET("/budgets/:id/variance", accountingHandler.GetBudgetVariance)

				// Payment Management
				accounting.POST("/payments", accountingHandler.CreatePayment)
				accounting.GET("/payments", accountingHandler.GetPayments)
				accounting.GET("/payments/:id", accountingHandler.GetPayment)
				accounting.PUT("/payments/:id", accountingHandler.UpdatePayment)
				accounting.DELETE("/payments/:id", accountingHandler.DeletePayment)
				accounting.POST("/payments/:id/clear", accountingHandler.ClearPayment)
				accounting.POST("/payments/:id/cancel", accountingHandler.CancelPayment)

				// Settings
				accounting.GET("/settings", accountingHandler.GetSettings)
				accounting.POST("/settings", accountingHandler.SaveSettings)
			}

			// HRM routes (admin, manager, hr)
			hrmHandler := handlers.NewHRMHandler(db)
			hrm := protected.Group("/hrm")
			hrm.Use(middleware.AuthorizeRole("admin", "manager", "hr"))
			{
				// Employee Management
				hrm.POST("/employees", hrmHandler.CreateEmployee)
				hrm.GET("/employees", hrmHandler.GetEmployees)
				hrm.GET("/employees/:id", hrmHandler.GetEmployee)
				hrm.PUT("/employees/:id", hrmHandler.UpdateEmployee)
				hrm.DELETE("/employees/:id", hrmHandler.DeleteEmployee)

				// Payroll Management
				hrm.POST("/payrolls", hrmHandler.CreatePayroll)
				hrm.GET("/payrolls", hrmHandler.GetPayrolls)
				hrm.GET("/payroll-details", hrmHandler.GetPayrollDetails)
				hrm.GET("/payrolls/:id", hrmHandler.GetPayroll)
				hrm.POST("/payrolls/:id/process", hrmHandler.ProcessPayroll)
				hrm.POST("/payrolls/:id/approve", hrmHandler.ApprovePayroll)
				hrm.POST("/payrolls/:id/pay", hrmHandler.PayPayroll)
				hrm.GET("/payroll/:payroll_id/employee/:employee_id/payslip", hrmHandler.GetEmployeePayslip)

				// Attendance Management
				hrm.POST("/attendance/check-in", hrmHandler.CheckIn)
				hrm.POST("/attendance/:id/check-out", hrmHandler.CheckOut)
				hrm.GET("/attendances", hrmHandler.GetAttendances)
				hrm.GET("/attendances/:id", hrmHandler.GetAttendance)
				hrm.PUT("/attendances/:id", hrmHandler.UpdateAttendance)
				hrm.GET("/attendance/summary", hrmHandler.GetAttendanceSummary)

				// Leave Management
				hrm.POST("/leave-requests", hrmHandler.CreateLeaveRequest)
				hrm.GET("/leave-requests", hrmHandler.GetLeaveRequests)
				hrm.GET("/leave-requests/:id", hrmHandler.GetLeaveRequest)
				hrm.POST("/leave-requests/:id/approve", hrmHandler.ApproveLeaveRequest)
				hrm.POST("/leave-requests/:id/reject", hrmHandler.RejectLeaveRequest)
				hrm.POST("/leave-requests/:id/cancel", hrmHandler.CancelLeaveRequest)
				hrm.GET("/leave-balance", hrmHandler.GetLeaveBalance)
			}

			// Inventory routes (admin, manager, inventory)
			inventoryHandler := handlers.NewInventoryHandler(db)
			inventory := protected.Group("/inventory")
			inventory.Use(middleware.AuthorizeRole("admin", "manager", "inventory"))
			{
				// Product Management
				inventory.POST("/products", inventoryHandler.CreateProduct)
				inventory.GET("/products", inventoryHandler.GetProducts)
				inventory.GET("/products/:id", inventoryHandler.GetProduct)
				inventory.PUT("/products/:id", inventoryHandler.UpdateProduct)
				inventory.DELETE("/products/:id", inventoryHandler.DeleteProduct)

				// Stock Management
				inventory.GET("/stocks", inventoryHandler.GetStocks)
				inventory.GET("/stocks/:id", inventoryHandler.GetStock)
				inventory.POST("/stock-movements", inventoryHandler.CreateStockMovement)
				inventory.GET("/stock-movements", inventoryHandler.GetStockMovements)
				inventory.POST("/stock-adjustments", inventoryHandler.CreateStockAdjustment)
				inventory.GET("/stock-adjustments", inventoryHandler.GetStockAdjustments)
				inventory.GET("/stocks/low-stock", inventoryHandler.GetLowStockProducts)

				// Warehouse Management
				inventory.POST("/warehouses", inventoryHandler.CreateWarehouse)
				inventory.GET("/warehouses", inventoryHandler.GetWarehouses)
				inventory.GET("/warehouses/:id", inventoryHandler.GetWarehouse)
				inventory.PUT("/warehouses/:id", inventoryHandler.UpdateWarehouse)
				inventory.DELETE("/warehouses/:id", inventoryHandler.DeleteWarehouse)

				// Purchasing - Purchase Requisition
				inventory.POST("/purchase-requisitions", inventoryHandler.CreatePurchaseRequisition)
				inventory.GET("/purchase-requisitions", inventoryHandler.GetPurchaseRequisitions)
				inventory.POST("/purchase-requisitions/:id/approve", inventoryHandler.ApprovePurchaseRequisition)

				// Purchasing - Purchase Order
				inventory.POST("/purchase-orders", inventoryHandler.CreatePurchaseOrder)
				inventory.GET("/purchase-orders", inventoryHandler.GetPurchaseOrders)
				inventory.GET("/purchase-orders/:id", inventoryHandler.GetPurchaseOrder)
				inventory.POST("/purchase-orders/:id/approve", inventoryHandler.ApprovePurchaseOrder)

				// Purchasing - Goods Receipt
				inventory.POST("/goods-receipts", inventoryHandler.CreateGoodsReceipt)
				inventory.GET("/goods-receipts", inventoryHandler.GetGoodsReceipts)
				inventory.POST("/goods-receipts/:id/confirm", inventoryHandler.ConfirmGoodsReceipt)
			}

			// CRM routes (admin, manager, sales)
			crmHandler := handlers.NewCRMHandler(db)
			crm := protected.Group("/crm")
			crm.Use(middleware.AuthorizeRole("admin", "manager", "sales"))
			{
				// Contact Management
				crm.POST("/contacts", crmHandler.CreateContact)
				crm.GET("/contacts", crmHandler.GetContacts)
				crm.GET("/contacts/:id", crmHandler.GetContact)
				crm.PUT("/contacts/:id", crmHandler.UpdateContact)
				crm.DELETE("/contacts/:id", crmHandler.DeleteContact)

				// Lead Management
				crm.POST("/leads", crmHandler.CreateLead)
				crm.GET("/leads", crmHandler.GetLeads)
				crm.GET("/leads/:id", crmHandler.GetLead)
				crm.PUT("/leads/:id", crmHandler.UpdateLead)
				crm.DELETE("/leads/:id", crmHandler.DeleteLead)
				crm.POST("/leads/:id/qualify", crmHandler.QualifyLead)
				crm.POST("/leads/:id/convert", crmHandler.ConvertLead)
				crm.PUT("/leads/:id/score", crmHandler.UpdateLeadScore)

				// Opportunity Management
				crm.POST("/opportunities", crmHandler.CreateOpportunity)
				crm.GET("/opportunities", crmHandler.GetOpportunities)
				crm.GET("/opportunities/:id", crmHandler.GetOpportunity)
				crm.PUT("/opportunities/:id", crmHandler.UpdateOpportunity)
				crm.DELETE("/opportunities/:id", crmHandler.DeleteOpportunity)
				crm.POST("/opportunities/:id/move-stage", crmHandler.MoveOpportunityStage)
				crm.POST("/opportunities/:id/win", crmHandler.WinOpportunity)
				crm.POST("/opportunities/:id/lose", crmHandler.LoseOpportunity)
				crm.GET("/opportunities/pipeline", crmHandler.GetOpportunityPipeline)
				crm.GET("/opportunities/forecast", crmHandler.GetOpportunityForecast)

				// Quotation Management
				crm.POST("/quotations", crmHandler.CreateQuotation)
				crm.GET("/quotations", crmHandler.GetQuotations)
				crm.GET("/quotations/:id", crmHandler.GetQuotation)
				crm.POST("/quotations/:id/send", crmHandler.SendQuotation)
				crm.POST("/quotations/:id/accept", crmHandler.AcceptQuotation)

				// Sales Order Management
				crm.POST("/sales-orders", crmHandler.CreateSalesOrder)
				crm.GET("/sales-orders", crmHandler.GetSalesOrders)
				crm.GET("/sales-orders/:id", crmHandler.GetSalesOrder)
				crm.POST("/sales-orders/:id/confirm", crmHandler.ConfirmSalesOrder)
				crm.POST("/sales-orders/:id/deliver", crmHandler.DeliverSalesOrder)
				crm.POST("/sales-orders/:id/cancel", crmHandler.CancelSalesOrder)
			}

			// Manufacturing routes (admin, manager, production)
			manufacturingHandler := handlers.NewManufacturingHandler(db)
			manufacturing := protected.Group("/manufacturing")
			manufacturing.Use(middleware.AuthorizeRole("admin", "manager", "production"))
			{
				// BOM Management
				manufacturing.POST("/boms", manufacturingHandler.CreateBOM)
				manufacturing.GET("/boms", manufacturingHandler.GetBOMs)
				manufacturing.GET("/boms/:id", manufacturingHandler.GetBOM)
				manufacturing.PUT("/boms/:id", manufacturingHandler.UpdateBOM)
				manufacturing.DELETE("/boms/:id", manufacturingHandler.DeleteBOM)
				manufacturing.POST("/boms/:id/set-default", manufacturingHandler.SetDefaultBOM)
				manufacturing.POST("/boms/:id/copy", manufacturingHandler.CopyBOM)
				manufacturing.GET("/boms/:id/cost-breakdown", manufacturingHandler.GetBOMCostBreakdown)

				// Production Planning
				manufacturing.POST("/production-plans", manufacturingHandler.CreateProductionPlan)
				manufacturing.GET("/production-plans", manufacturingHandler.GetProductionPlans)
				manufacturing.GET("/production-plans/:id", manufacturingHandler.GetProductionPlan)
				manufacturing.POST("/production-plans/:id/approve", manufacturingHandler.ApproveProductionPlan)

				// Production Order Management
				manufacturing.POST("/production-orders", manufacturingHandler.CreateProductionOrder)
				manufacturing.GET("/production-orders", manufacturingHandler.GetProductionOrders)
				manufacturing.GET("/production-orders/:id", manufacturingHandler.GetProductionOrder)
				manufacturing.POST("/production-orders/:id/start", manufacturingHandler.StartProductionOrder)
				manufacturing.POST("/production-orders/:id/complete", manufacturingHandler.CompleteProductionOrder)
				manufacturing.POST("/production-orders/:id/cancel", manufacturingHandler.CancelProductionOrder)
				manufacturing.GET("/production/schedule", manufacturingHandler.GetProductionSchedule)
				manufacturing.GET("/production/capacity", manufacturingHandler.GetProductionCapacity)
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
