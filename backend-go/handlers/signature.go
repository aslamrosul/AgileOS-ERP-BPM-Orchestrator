package handlers

import (
	"net/http"
	"time"

	"agileos-backend/database"
	"agileos-backend/internal/crypto"
	"agileos-backend/logger"
	"agileos-backend/models"

	"github.com/gin-gonic/gin"
)

type SignatureHandler struct {
	db *database.SurrealDB
}

func NewSignatureHandler(db *database.SurrealDB) *SignatureHandler {
	return &SignatureHandler{db: db}
}

// VerifySignature handles signature verification requests
func (h *SignatureHandler) VerifySignature(c *gin.Context) {
	var req models.SignatureVerificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get task instance
	task, err := h.db.GetTaskInstance(req.TaskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	// Verify signature
	result := crypto.VerifyTaskSignature(
		task.ID,
		task.ExecutedBy,
		task.ProcessInstanceID,
		"completed",
		req.Signature,
		*task.CompletedAt,
		task.Result.(map[string]interface{}),
	)

	// Log verification attempt
	logger.LogAudit("signature_verified", c.GetString("user_id"), req.TaskID, map[string]interface{}{
		"valid":     result.Valid,
		"signature": req.Signature,
	})

	c.JSON(http.StatusOK, result)
}

// GetTaskSignature retrieves signature information for a task
func (h *SignatureHandler) GetTaskSignature(c *gin.Context) {
	taskID := c.Param("id")

	task, err := h.db.GetTaskInstance(taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	// Check if task has a signature
	if task.DigitalSignature == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task has no digital signature"})
		return
	}

	// Generate QR code data
	baseURL := c.Request.Host
	qrData := crypto.GenerateQRCodeData(task.ID, task.DigitalSignature, "https://"+baseURL)

	response := gin.H{
		"task_id":           task.ID,
		"digital_signature": task.DigitalSignature,
		"signed_by":         task.ExecutedBy,
		"signed_at":         task.CompletedAt,
		"workflow_id":       task.ProcessInstanceID,
		"status":            task.Status,
		"qr_code_data":      qrData,
		"signature_metadata": task.SignatureMetadata,
	}

	c.JSON(http.StatusOK, response)
}

// VerifyTaskIntegrity checks if task data has been tampered with
func (h *SignatureHandler) VerifyTaskIntegrity(c *gin.Context) {
	taskID := c.Param("id")

	task, err := h.db.GetTaskInstance(taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	if task.DigitalSignature == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Task has no digital signature"})
		return
	}

	// Verify current data against stored signature
	signatureData := crypto.SignatureData{
		TaskID:     task.ID,
		UserID:     task.ExecutedBy,
		Timestamp:  *task.CompletedAt,
		WorkflowID: task.ProcessInstanceID,
		Action:     "completed",
		Data:       task.Result.(map[string]interface{}),
	}

	isValid := crypto.VerifySignature(task.DigitalSignature, signatureData)

	result := gin.H{
		"task_id":           task.ID,
		"integrity_valid":   isValid,
		"stored_signature":  task.DigitalSignature,
		"verified_at":       time.Now(),
	}

	if !isValid {
		// Log security event - potential tampering detected
		logger.LogSecurity("signature_mismatch", task.ExecutedBy, c.ClientIP(), map[string]interface{}{
			"task_id":          task.ID,
			"stored_signature": task.DigitalSignature,
			"expected_signature": crypto.GenerateSignature(signatureData),
		})

		result["message"] = "SECURITY ALERT: Task data may have been tampered with"
		result["expected_signature"] = crypto.GenerateSignature(signatureData)
	} else {
		result["message"] = "Task integrity verified - data is authentic"
	}

	c.JSON(http.StatusOK, result)
}

// GenerateTaskReceipt generates a digital receipt for a completed task
func (h *SignatureHandler) GenerateTaskReceipt(c *gin.Context) {
	taskID := c.Param("id")

	task, err := h.db.GetTaskInstance(taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	if task.Status != models.TaskStatusCompleted {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Task is not completed"})
		return
	}

	// Get process instance for additional context
	process, err := h.db.GetProcessInstance(task.ProcessInstanceID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Process not found"})
		return
	}

	// Get workflow for process name
	workflow, err := h.db.GetWorkflow(process.WorkflowID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Workflow not found"})
		return
	}

	// Generate receipt
	receipt := models.DocumentReceipt{
		ID:                taskID + "_receipt",
		DocumentID:        taskID,
		DocumentTitle:     "Task Completion Receipt - " + task.StepName,
		ProcessName:       workflow.Name,
		WorkflowID:        workflow.ID,
		ProcessInstanceID: process.ID,
		Signatures: []models.DocumentSignature{
			{
				SignedBy:     task.ExecutedBy,
				SignedByName: task.ExecutedBy, // In real app, get actual name
				SignedAt:     *task.CompletedAt,
				Action:       models.SignatureActionApproved,
				Signature:    task.DigitalSignature,
				TaskID:       task.ID,
				Comments:     "Task completed successfully",
			},
		},
		GeneratedAt:     time.Now(),
		VerificationURL: "https://" + c.Request.Host + "/verify/" + taskID,
		QRCodeData:      crypto.GenerateQRCodeData(task.ID, task.DigitalSignature, "https://"+c.Request.Host),
		Status:          "final",
	}

	c.JSON(http.StatusOK, receipt)
}