package models

import "time"

// TaskInstance represents a running task for a specific step
type TaskInstance struct {
	ID                string                 `json:"id,omitempty"`
	ProcessInstanceID string                 `json:"process_instance_id"`
	StepID            string                 `json:"step_id"`
	StepName          string                 `json:"step_name"`
	Status            TaskStatus             `json:"status"`
	AssignedTo        string                 `json:"assigned_to"`
	ExecutedBy        string                 `json:"executed_by,omitempty"`
	CreatedAt         time.Time              `json:"created_at"`
	StartedAt         *time.Time             `json:"started_at,omitempty"`
	CompletedAt       *time.Time             `json:"completed_at,omitempty"`
	DueAt             time.Time              `json:"due_at"`
	Data              interface{}            `json:"data,omitempty"`
	Result            interface{}            `json:"result,omitempty"`
	DigitalSignature  string                 `json:"digital_signature,omitempty"`
	SignatureMetadata map[string]interface{} `json:"signature_metadata,omitempty"`
}

// TaskStatus defines the status of a task instance
type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusFailed     TaskStatus = "failed"
	TaskStatusCancelled  TaskStatus = "cancelled"
)
