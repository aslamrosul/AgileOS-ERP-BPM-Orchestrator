package models

import "time"

// Workflow represents a business process definition
type Workflow struct {
	ID          string    `json:"id,omitempty"`
	Name        string    `json:"name"`
	Version     string    `json:"version"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	IsActive    bool      `json:"is_active"`
}

// Step represents a single step in a workflow
type Step struct {
	ID          string        `json:"id,omitempty"`
	WorkflowID  string        `json:"workflow_id"`
	Name        string        `json:"name"`
	Type        StepType      `json:"type"`
	AssignedTo  string        `json:"assigned_to,omitempty"` // User/Role ID
	SLA         time.Duration `json:"sla"`                    // Max duration
	Description string        `json:"description"`
	Config      interface{}   `json:"config,omitempty"` // Flexible config for step-specific data
	CreatedAt   time.Time     `json:"created_at"`
}

// StepType defines the type of step
type StepType string

const (
	StepTypeApproval StepType = "approval"
	StepTypeAction   StepType = "action"
	StepTypeDecision StepType = "decision"
	StepTypeNotify   StepType = "notify"
)

// StepRelation represents the NEXT relationship between steps
type StepRelation struct {
	ID        string                 `json:"id,omitempty"`
	In        string                 `json:"in"`  // From step
	Out       string                 `json:"out"` // To step
	Condition map[string]interface{} `json:"condition,omitempty"`
}

// ProcessInstance represents a running instance of a workflow
type ProcessInstance struct {
	ID              string                 `json:"id,omitempty"`
	WorkflowID      string                 `json:"workflow_id"`
	CurrentStepID   string                 `json:"current_step_id"`
	Status          ProcessStatus          `json:"status"`
	StartedAt       time.Time              `json:"started_at"`
	CompletedAt     *time.Time             `json:"completed_at,omitempty"`
	InitiatedBy     string                 `json:"initiated_by"`
	Data            interface{}            `json:"data"` // Process variables
	ExecutionHistory []ExecutionLog        `json:"execution_history,omitempty"`
}

// ProcessStatus defines the status of a process instance
type ProcessStatus string

const (
	ProcessStatusPending   ProcessStatus = "pending"
	ProcessStatusRunning   ProcessStatus = "running"
	ProcessStatusCompleted ProcessStatus = "completed"
	ProcessStatusFailed    ProcessStatus = "failed"
	ProcessStatusCancelled ProcessStatus = "cancelled"
)

// ExecutionLog tracks step execution history
type ExecutionLog struct {
	StepID      string                 `json:"step_id"`
	ExecutedAt  time.Time              `json:"executed_at"`
	ExecutedBy  string                 `json:"executed_by"`
	Action      string                 `json:"action"`
	Result      interface{}            `json:"result,omitempty"`
	Duration    time.Duration          `json:"duration"`
	CompletedAt time.Time              `json:"completed_at"`
}
