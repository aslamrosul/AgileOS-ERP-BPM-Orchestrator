package handlers

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"agileos-backend/database"
	"agileos-backend/internal/crypto"
	"agileos-backend/logger"
	"agileos-backend/messaging"
	"agileos-backend/models"

	"github.com/gin-gonic/gin"
)

type TaskHandler struct {
	db         *database.SurrealDB
	natsClient *messaging.NATSClient
}

func NewTaskHandler(db *database.SurrealDB, natsClient *messaging.NATSClient) *TaskHandler {
	return &TaskHandler{
		db:         db,
		natsClient: natsClient,
	}
}

type CompleteTaskRequest struct {
	ExecutedBy string                 `json:"executed_by" binding:"required"`
	Result     map[string]interface{} `json:"result"`
}

type StartProcessRequest struct {
	WorkflowID  string                 `json:"workflow_id" binding:"required"`
	InitiatedBy string                 `json:"initiated_by" binding:"required"`
	Data        map[string]interface{} `json:"data"`
}

// CompleteTask handles POST /api/v1/task/:id/complete
func (h *TaskHandler) CompleteTask(c *gin.Context) {
	taskID := c.Param("id")

	var req CompleteTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get task instance
	task, err := h.db.GetTaskInstance(taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	// Update task status
	now := time.Now()
	task.Status = models.TaskStatusCompleted
	task.CompletedAt = &now
	task.ExecutedBy = req.ExecutedBy
	task.Result = req.Result

	// Generate digital signature
	signatureData := crypto.SignatureData{
		TaskID:     task.ID,
		UserID:     req.ExecutedBy,
		Timestamp:  now,
		WorkflowID: task.ProcessInstanceID,
		Action:     "completed",
		Data:       req.Result,
	}

	task.DigitalSignature = crypto.GenerateSignature(signatureData)
	task.SignatureMetadata = map[string]interface{}{
		"signed_by":    req.ExecutedBy,
		"signed_at":    now,
		"action":       "completed",
		"ip_address":   c.ClientIP(),
		"user_agent":   c.GetHeader("User-Agent"),
	}

	if err := h.db.UpdateTaskInstance(task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update task"})
		return
	}

	// Log audit trail
	logger.LogAudit("task_completed", req.ExecutedBy, task.ID, map[string]interface{}{
		"step_name":         task.StepName,
		"process_id":        task.ProcessInstanceID,
		"digital_signature": task.DigitalSignature,
		"result":            req.Result,
	})

	// Log BPM event
	logger.LogBPM("task_completed", task.ProcessInstanceID, task.ID, map[string]interface{}{
		"step_name":         task.StepName,
		"executed_by":       req.ExecutedBy,
		"digital_signature": task.DigitalSignature,
		"completion_time":   now,
	})

	// Publish task completed event to NATS
	event := messaging.TaskCompletedEvent{
		TaskID:            taskID,
		ProcessInstanceID: task.ProcessInstanceID,
		CurrentStepID:     task.StepID,
		ExecutedBy:        req.ExecutedBy,
		Result:            req.Result,
		CompletedAt:       now,
	}

	if err := h.natsClient.PublishTaskCompleted(event); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to publish event"})
		return
	}

	// Generate QR code data for verification
	baseURL := c.Request.Host
	qrData := crypto.GenerateQRCodeData(task.ID, task.DigitalSignature, "https://"+baseURL)

	c.JSON(http.StatusOK, gin.H{
		"message":           "Task completed successfully",
		"task_id":           taskID,
		"status":            "completed",
		"digital_signature": task.DigitalSignature,
		"qr_code_data":      qrData,
		"signed_by":         req.ExecutedBy,
		"signed_at":         now,
	})
}

// GetPendingTasks handles GET /api/v1/tasks/pending/:assignedTo
func (h *TaskHandler) GetPendingTasks(c *gin.Context) {
	assignedTo := c.Param("assignedTo")

	tasks, err := h.db.GetPendingTasks(assignedTo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get tasks"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tasks": tasks,
		"count": len(tasks),
	})
}

// StartProcess handles POST /api/v1/process/start
func (h *TaskHandler) StartProcess(c *gin.Context) {
	var req StartProcessRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get workflow
	workflow, err := h.db.GetWorkflow(req.WorkflowID)
	if err != nil {
		log.Printf("❌ Failed to get workflow %s: %v", req.WorkflowID, err)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Workflow not found",
			"detail": err.Error(),
		})
		return
	}

	// Get first step (should be start node)
	steps, err := h.db.GetWorkflowSteps(workflow.ID)
	if err != nil || len(steps) == 0 {
		log.Printf("❌ Failed to get workflow steps: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Workflow has no steps",
			"detail": fmt.Sprintf("Steps count: %d, error: %v", len(steps), err),
		})
		return
	}

	firstStep := steps[0]
	log.Printf("📋 Starting process with workflow: %s, first step: %s", workflow.Name, firstStep.Name)

	// Create process instance
	instance := &models.ProcessInstance{
		WorkflowID:       workflow.ID,
		CurrentStepID:    firstStep.ID,
		Status:           models.ProcessStatusRunning,
		StartedAt:        time.Now(),
		InitiatedBy:      req.InitiatedBy,
		Data:             req.Data,
		ExecutionHistory: []models.ExecutionLog{},
	}

	if err := h.db.CreateProcessInstance(instance); err != nil {
		log.Printf("❌ Failed to create process instance: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create process",
			"detail": err.Error(),
		})
		return
	}

	log.Printf("✓ Process instance created: %s", instance.ID)

	// Create first task instance
	taskInstance := &models.TaskInstance{
		ProcessInstanceID: instance.ID,
		StepID:            firstStep.ID,
		StepName:          firstStep.Name,
		Status:            models.TaskStatusPending,
		AssignedTo:        firstStep.AssignedTo,
		CreatedAt:         time.Now(),
		DueAt:             time.Now().Add(firstStep.SLA),
	}

	if err := h.db.CreateTaskInstance(taskInstance); err != nil {
		log.Printf("❌ Failed to create task instance: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create task",
			"detail": err.Error(),
		})
		return
	}

	log.Printf("✓ Task instance created: %s", taskInstance.ID)

	// Publish task started event
	startedEvent := messaging.TaskStartedEvent{
		TaskID:            taskInstance.ID,
		ProcessInstanceID: instance.ID,
		StepID:            firstStep.ID,
		StepName:          firstStep.Name,
		AssignedTo:        firstStep.AssignedTo,
		StartedAt:         time.Now(),
	}

	h.natsClient.PublishTaskStarted(startedEvent)

	c.JSON(http.StatusCreated, gin.H{
		"message":             "Process started successfully",
		"process_instance_id": instance.ID,
		"first_task_id":       taskInstance.ID,
		"current_step":        firstStep.Name,
	})
}
