package models

import "time"

// AuditLog represents an audit trail entry for compliance and security
type AuditLog struct {
	ID         string                 `json:"id,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
	Type       string                 `json:"type"` // audit, security, bpm, performance
	Action     string                 `json:"action"`
	UserID     string                 `json:"user_id,omitempty"`
	Username   string                 `json:"username,omitempty"`
	IPAddress  string                 `json:"ip_address,omitempty"`
	Resource   string                 `json:"resource,omitempty"`
	ResourceID string                 `json:"resource_id,omitempty"`
	Details    map[string]interface{} `json:"details,omitempty"`
	Severity   string                 `json:"severity"` // info, warn, error, critical
	Success    bool                   `json:"success"`
}

// WorkflowHistory represents historical snapshots of workflow changes
type WorkflowHistory struct {
	ID           string                 `json:"id,omitempty"`
	WorkflowID   string                 `json:"workflow_id"`
	Version      string                 `json:"version"`
	ChangedBy    string                 `json:"changed_by"`
	ChangedAt    time.Time              `json:"changed_at"`
	ChangeType   string                 `json:"change_type"` // created, updated, deleted
	Snapshot     map[string]interface{} `json:"snapshot"`     // Full workflow state
	ChangeReason string                 `json:"change_reason,omitempty"`
}

// SystemHealth represents system health check results
type SystemHealth struct {
	Status       string                 `json:"status"` // healthy, degraded, unhealthy
	Timestamp    time.Time              `json:"timestamp"`
	Database     HealthCheck            `json:"database"`
	MessageBroker HealthCheck           `json:"message_broker"`
	Dependencies map[string]HealthCheck `json:"dependencies,omitempty"`
	Uptime       int64                  `json:"uptime_seconds"`
	Version      string                 `json:"version"`
}

// HealthCheck represents a single health check result
type HealthCheck struct {
	Status       string                 `json:"status"` // up, down, degraded
	ResponseTime int64                  `json:"response_time_ms"`
	Message      string                 `json:"message,omitempty"`
	Details      map[string]interface{} `json:"details,omitempty"`
}

// PerformanceMetric represents performance monitoring data
type PerformanceMetric struct {
	ID        string                 `json:"id,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Operation string                 `json:"operation"`
	Duration  int64                  `json:"duration_ms"`
	Success   bool                   `json:"success"`
	Details   map[string]interface{} `json:"details,omitempty"`
}
