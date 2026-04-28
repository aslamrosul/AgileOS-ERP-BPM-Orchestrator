package handlers

import (
	"net/http"
	"time"

	"agileos-backend/database"
	"agileos-backend/models"

	"github.com/gin-gonic/gin"
)

type WorkflowHandler struct {
	db *database.SurrealDB
}

func NewWorkflowHandler(db *database.SurrealDB) *WorkflowHandler {
	return &WorkflowHandler{db: db}
}

type CreateWorkflowRequest struct {
	Workflow  WorkflowInput  `json:"workflow"`
	Steps     []StepInput    `json:"steps"`
	Relations []RelationInput `json:"relations"`
}

type WorkflowInput struct {
	Name        string `json:"name" binding:"required"`
	Version     string `json:"version"`
	Description string `json:"description"`
	IsActive    bool   `json:"is_active"`
}

type StepInput struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name" binding:"required"`
	Type       string                 `json:"type" binding:"required"`
	AssignedTo string                 `json:"assigned_to"`
	SLA        string                 `json:"sla"`
	Position   map[string]interface{} `json:"position"`
}

type RelationInput struct {
	From      string                 `json:"from" binding:"required"`
	To        string                 `json:"to" binding:"required"`
	Condition map[string]interface{} `json:"condition"`
}

type CreateWorkflowResponse struct {
	WorkflowID       string `json:"workflow_id"`
	StepsCreated     int    `json:"steps_created"`
	RelationsCreated int    `json:"relations_created"`
	Message          string `json:"message"`
}

// CreateWorkflow handles POST /api/v1/workflow
func (h *WorkflowHandler) CreateWorkflow(c *gin.Context) {
	var req CreateWorkflowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create workflow
	workflow := &models.Workflow{
		Name:        req.Workflow.Name,
		Version:     req.Workflow.Version,
		Description: req.Workflow.Description,
		IsActive:    req.Workflow.IsActive,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := h.db.SaveWorkflow(workflow); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to save workflow",
			"detail": err.Error(),
		})
		return
	}

	// Check if workflow ID was set
	if workflow.ID == "" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Workflow created but ID is empty",
		})
		return
	}

	// Create steps
	stepIDMap := make(map[string]string) // frontend ID -> backend ID
	for _, stepInput := range req.Steps {
		sla, _ := time.ParseDuration(stepInput.SLA)
		if sla == 0 {
			sla = 24 * time.Hour
		}

		step := &models.Step{
			WorkflowID:  workflow.ID,
			Name:        stepInput.Name,
			Type:        models.StepType(stepInput.Type),
			AssignedTo:  stepInput.AssignedTo,
			SLA:         sla,
			Description: "",
			Config:      stepInput.Position,
			CreatedAt:   time.Now(),
		}

		if err := h.db.AddStep(step); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add step"})
			return
		}

		stepIDMap[stepInput.ID] = step.ID
	}

	// Create relations
	relationsCreated := 0
	for _, rel := range req.Relations {
		fromID := stepIDMap[rel.From]
		toID := stepIDMap[rel.To]

		if fromID == "" || toID == "" {
			continue
		}

		if err := h.db.LinkSteps(fromID, toID, rel.Condition); err != nil {
			// Log error but continue
			continue
		}
		relationsCreated++
	}

	c.JSON(http.StatusCreated, CreateWorkflowResponse{
		WorkflowID:       workflow.ID,
		StepsCreated:     len(req.Steps),
		RelationsCreated: relationsCreated,
		Message:          "Workflow created successfully",
	})
}

// GetWorkflows handles GET /api/v1/workflows
func (h *WorkflowHandler) GetWorkflows(c *gin.Context) {
	// TODO: Implement list workflows
	c.JSON(http.StatusOK, gin.H{
		"workflows": []interface{}{},
		"message":   "List workflows - coming soon",
	})
}

// GetWorkflow handles GET /api/v1/workflow/:id
func (h *WorkflowHandler) GetWorkflow(c *gin.Context) {
	workflowID := c.Param("id")

	workflow, err := h.db.GetWorkflow(workflowID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Workflow not found"})
		return
	}

	steps, err := h.db.GetWorkflowSteps(workflow.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get steps"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"workflow": workflow,
		"steps":    steps,
	})
}
