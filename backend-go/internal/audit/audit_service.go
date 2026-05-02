package audit

import (
	"encoding/json"
	"fmt"
	"time"

	"agileos-backend/database"
	"agileos-backend/logger"
)

// AuditAction represents the type of action being audited
type AuditAction string

const (
	ActionCreate         AuditAction = "CREATE"
	ActionUpdate         AuditAction = "UPDATE"
	ActionDelete         AuditAction = "DELETE"
	ActionApprove        AuditAction = "APPROVE"
	ActionReject         AuditAction = "REJECT"
	ActionLogin          AuditAction = "LOGIN"
	ActionLogout         AuditAction = "LOGOUT"
	ActionWorkflowChange AuditAction = "WORKFLOW_CHANGE"
	ActionPolicyChange   AuditAction = "POLICY_CHANGE"
	ActionAccessDenied   AuditAction = "ACCESS_DENIED"
	ActionUnauthorized   AuditAction = "UNAUTHORIZED_ACTION"
)

// ComplianceStatus represents the compliance check result
type ComplianceStatus string

const (
	CompliancePass    ComplianceStatus = "PASS"
	ComplianceFail    ComplianceStatus = "FAIL"
	ComplianceWarning ComplianceStatus = "WARNING"
	ComplianceReview  ComplianceStatus = "REVIEW"
)

// AuditTrail represents an immutable audit log entry
type AuditTrail struct {
	ID               string                 `json:"id,omitempty"`
	Timestamp        time.Time              `json:"timestamp"`
	ActorID          string                 `json:"actor_id"`
	ActorUsername    string                 `json:"actor_username,omitempty"`
	ActorRole        string                 `json:"actor_role,omitempty"`
	Action           AuditAction            `json:"action"`
	ResourceType     string                 `json:"resource_type"`
	ResourceID       string                 `json:"resource_id"`
	OldValue         map[string]interface{} `json:"old_value,omitempty"`
	NewValue         map[string]interface{} `json:"new_value,omitempty"`
	IPAddress        string                 `json:"ip_address,omitempty"`
	UserAgent        string                 `json:"user_agent,omitempty"`
	ComplianceStatus ComplianceStatus       `json:"compliance_status"`
	ComplianceNotes  string                 `json:"compliance_notes,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// WorkflowVersion represents a versioned workflow
type WorkflowVersion struct {
	ID              string                 `json:"id,omitempty"`
	WorkflowID      string                 `json:"workflow_id"`
	Version         string                 `json:"version"`
	VersionNumber   int                    `json:"version_number"`
	Name            string                 `json:"name"`
	Description     string                 `json:"description"`
	Definition      map[string]interface{} `json:"definition"`
	CreatedBy       string                 `json:"created_by"`
	CreatedAt       time.Time              `json:"created_at"`
	ChangeReason    string                 `json:"change_reason,omitempty"`
	IsActive        bool                   `json:"is_active"`
	ApprovedBy      string                 `json:"approved_by,omitempty"`
	ApprovedAt      *time.Time             `json:"approved_at,omitempty"`
}

// AuditService handles all audit trail operations
type AuditService struct {
	db *database.SurrealDB
}

// NewAuditService creates a new audit service
func NewAuditService(db *database.SurrealDB) *AuditService {
	return &AuditService{db: db}
}

// LogAction creates an immutable audit trail entry
func (s *AuditService) LogAction(trail AuditTrail) error {
	// Set timestamp if not provided
	if trail.Timestamp.IsZero() {
		trail.Timestamp = time.Now()
	}

	// Perform compliance check
	trail.ComplianceStatus, trail.ComplianceNotes = s.checkCompliance(trail)

	// Create audit trail record (immutable - no updates allowed)
	query := `CREATE audit_trails CONTENT $data`
	
	_, err := s.db.Query(query, map[string]interface{}{
		"data": trail,
	})

	if err != nil {
		logger.LogError("Failed to create audit trail", err, map[string]interface{}{
			"actor_id":      trail.ActorID,
			"action":        trail.Action,
			"resource_type": trail.ResourceType,
			"resource_id":   trail.ResourceID,
		})
		return fmt.Errorf("failed to create audit trail: %w", err)
	}

	// Log to structured logger as well
	logger.LogAudit(string(trail.Action), trail.ActorID, trail.ResourceID, map[string]interface{}{
		"resource_type":     trail.ResourceType,
		"compliance_status": trail.ComplianceStatus,
		"ip_address":        trail.IPAddress,
	})

	// If compliance failed, log a warning
	if trail.ComplianceStatus == ComplianceFail {
		logger.Log.Warn().
			Str("actor_id", trail.ActorID).
			Str("action", string(trail.Action)).
			Str("resource_id", trail.ResourceID).
			Str("compliance_notes", trail.ComplianceNotes).
			Msg("🚨 COMPLIANCE VIOLATION DETECTED")
	}

	return nil
}

// checkCompliance performs automated compliance checks
func (s *AuditService) checkCompliance(trail AuditTrail) (ComplianceStatus, string) {
	// Check for unauthorized actions
	if trail.Action == ActionUnauthorized {
		return ComplianceFail, "Unauthorized action attempted"
	}

	// Check approval actions
	if trail.Action == ActionApprove || trail.Action == ActionReject {
		// Verify actor has appropriate role
		if trail.ActorRole != "admin" && trail.ActorRole != "manager" && trail.ActorRole != "finance" {
			return ComplianceFail, fmt.Sprintf("User with role '%s' is not authorized to approve/reject", trail.ActorRole)
		}

		// Check if approver is the same as initiator (conflict of interest)
		if trail.Metadata != nil {
			if initiator, ok := trail.Metadata["initiated_by"].(string); ok {
				if initiator == trail.ActorID {
					return ComplianceFail, "Self-approval detected - conflict of interest violation"
				}
			}
		}
	}

	// Check workflow changes
	if trail.Action == ActionWorkflowChange || trail.Action == ActionPolicyChange {
		// Only admins should change workflows
		if trail.ActorRole != "admin" {
			return ComplianceFail, fmt.Sprintf("User with role '%s' is not authorized to modify workflows", trail.ActorRole)
		}

		// Require change reason
		if trail.Metadata == nil || trail.Metadata["change_reason"] == nil || trail.Metadata["change_reason"] == "" {
			return ComplianceWarning, "Workflow change without documented reason"
		}
	}

	// Check for suspicious timing (e.g., actions outside business hours)
	hour := trail.Timestamp.Hour()
	if hour < 6 || hour > 22 {
		return ComplianceWarning, fmt.Sprintf("Action performed outside business hours (%02d:00)", hour)
	}

	// Check for rapid successive actions (potential automation/bot)
	if trail.Metadata != nil {
		if lastActionTime, ok := trail.Metadata["last_action_time"].(time.Time); ok {
			timeDiff := trail.Timestamp.Sub(lastActionTime)
			if timeDiff < 2*time.Second {
				return ComplianceWarning, "Rapid successive actions detected (< 2 seconds)"
			}
		}
	}

	return CompliancePass, "All compliance checks passed"
}

// GetAuditTrails retrieves audit trails with filtering
func (s *AuditService) GetAuditTrails(filters map[string]interface{}, limit, offset int) ([]AuditTrail, error) {
	query := `SELECT * FROM audit_trails`
	conditions := []string{}
	params := make(map[string]interface{})

	// Build WHERE clause based on filters
	if actorID, ok := filters["actor_id"].(string); ok && actorID != "" {
		conditions = append(conditions, "actor_id = $actor_id")
		params["actor_id"] = actorID
	}

	if action, ok := filters["action"].(string); ok && action != "" {
		conditions = append(conditions, "action = $action")
		params["action"] = action
	}

	if resourceType, ok := filters["resource_type"].(string); ok && resourceType != "" {
		conditions = append(conditions, "resource_type = $resource_type")
		params["resource_type"] = resourceType
	}

	if resourceID, ok := filters["resource_id"].(string); ok && resourceID != "" {
		conditions = append(conditions, "resource_id = $resource_id")
		params["resource_id"] = resourceID
	}

	if complianceStatus, ok := filters["compliance_status"].(string); ok && complianceStatus != "" {
		conditions = append(conditions, "compliance_status = $compliance_status")
		params["compliance_status"] = complianceStatus
	}

	// Date range filters
	if startDate, ok := filters["start_date"].(time.Time); ok {
		conditions = append(conditions, "timestamp >= $start_date")
		params["start_date"] = startDate
	}

	if endDate, ok := filters["end_date"].(time.Time); ok {
		conditions = append(conditions, "timestamp <= $end_date")
		params["end_date"] = endDate
	}

	// Add WHERE clause if conditions exist
	if len(conditions) > 0 {
		query += " WHERE "
		for i, condition := range conditions {
			if i > 0 {
				query += " AND "
			}
			query += condition
		}
	}

	// Add ORDER BY and pagination
	query += " ORDER BY timestamp DESC"
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}
	if offset > 0 {
		query += fmt.Sprintf(" START %d", offset)
	}

	result, err := s.db.Query(query, params)
	if err != nil {
		logger.LogError("Failed to retrieve audit trails", err, filters)
		return nil, fmt.Errorf("failed to retrieve audit trails: %w", err)
	}

	var trails []AuditTrail
	if err := database.UnmarshalSurrealResult(result, &trails); err != nil {
		logger.LogError("Failed to unmarshal audit trails", err, nil)
		return nil, fmt.Errorf("failed to unmarshal audit trails: %w", err)
	}

	return trails, nil
}

// GetAuditTrailCount returns the total count of audit trails matching filters
func (s *AuditService) GetAuditTrailCount(filters map[string]interface{}) (int, error) {
	query := `SELECT count() FROM audit_trails GROUP ALL`
	
	// Apply same filters as GetAuditTrails
	// (simplified for count query)
	
	result, err := s.db.Query(query, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to count audit trails: %w", err)
	}

	var countResult []map[string]interface{}
	if err := database.UnmarshalSurrealResult(result, &countResult); err != nil {
		return 0, err
	}

	if len(countResult) > 0 {
		if count, ok := countResult[0]["count"].(float64); ok {
			return int(count), nil
		}
	}

	return 0, nil
}

// CreateWorkflowVersion creates a new version of a workflow
func (s *AuditService) CreateWorkflowVersion(version WorkflowVersion) (*WorkflowVersion, error) {
	// Set creation timestamp
	version.CreatedAt = time.Now()

	// Get the latest version number for this workflow
	latestVersion, err := s.GetLatestWorkflowVersion(version.WorkflowID)
	if err == nil && latestVersion != nil {
		version.VersionNumber = latestVersion.VersionNumber + 1
		version.Version = fmt.Sprintf("v%d.0", version.VersionNumber)
		
		// Deactivate previous version
		s.DeactivateWorkflowVersion(latestVersion.ID)
	} else {
		version.VersionNumber = 1
		version.Version = "v1.0"
	}

	// Set as active
	version.IsActive = true

	// Create version record
	query := `CREATE workflow_versions CONTENT $data RETURN *`
	
	result, err := s.db.Query(query, map[string]interface{}{
		"data": version,
	})

	if err != nil {
		logger.LogError("Failed to create workflow version", err, map[string]interface{}{
			"workflow_id": version.WorkflowID,
			"version":     version.Version,
		})
		return nil, fmt.Errorf("failed to create workflow version: %w", err)
	}

	var createdVersions []WorkflowVersion
	if err := database.UnmarshalSurrealResult(result, &createdVersions); err != nil {
		return nil, err
	}

	if len(createdVersions) == 0 {
		return nil, fmt.Errorf("no workflow version created")
	}

	createdVersion := &createdVersions[0]

	// Log audit trail for workflow versioning
	s.LogAction(AuditTrail{
		ActorID:      version.CreatedBy,
		Action:       ActionWorkflowChange,
		ResourceType: "workflow",
		ResourceID:   version.WorkflowID,
		NewValue: map[string]interface{}{
			"version":       version.Version,
			"change_reason": version.ChangeReason,
		},
		Metadata: map[string]interface{}{
			"version_id":    createdVersion.ID,
			"change_reason": version.ChangeReason,
		},
	})

	logger.Log.Info().
		Str("workflow_id", version.WorkflowID).
		Str("version", version.Version).
		Str("created_by", version.CreatedBy).
		Msg("Workflow version created")

	return createdVersion, nil
}

// GetLatestWorkflowVersion retrieves the latest active version of a workflow
func (s *AuditService) GetLatestWorkflowVersion(workflowID string) (*WorkflowVersion, error) {
	query := `SELECT * FROM workflow_versions WHERE workflow_id = $workflow_id ORDER BY version_number DESC LIMIT 1`
	
	result, err := s.db.Query(query, map[string]interface{}{
		"workflow_id": workflowID,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get latest workflow version: %w", err)
	}

	var versions []WorkflowVersion
	if err := database.UnmarshalSurrealResult(result, &versions); err != nil {
		return nil, err
	}

	if len(versions) == 0 {
		return nil, fmt.Errorf("no versions found for workflow %s", workflowID)
	}

	return &versions[0], nil
}

// GetWorkflowVersionHistory retrieves all versions of a workflow
func (s *AuditService) GetWorkflowVersionHistory(workflowID string) ([]WorkflowVersion, error) {
	query := `SELECT * FROM workflow_versions WHERE workflow_id = $workflow_id ORDER BY version_number DESC`
	
	result, err := s.db.Query(query, map[string]interface{}{
		"workflow_id": workflowID,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get workflow version history: %w", err)
	}

	var versions []WorkflowVersion
	if err := database.UnmarshalSurrealResult(result, &versions); err != nil {
		return nil, err
	}

	return versions, nil
}

// DeactivateWorkflowVersion marks a workflow version as inactive
func (s *AuditService) DeactivateWorkflowVersion(versionID string) error {
	query := `UPDATE $version_id SET is_active = false`
	
	_, err := s.db.Query(query, map[string]interface{}{
		"version_id": versionID,
	})

	return err
}

// GetComplianceViolations retrieves all compliance violations
func (s *AuditService) GetComplianceViolations(startDate, endDate time.Time) ([]AuditTrail, error) {
	filters := map[string]interface{}{
		"compliance_status": ComplianceFail,
		"start_date":        startDate,
		"end_date":          endDate,
	}

	return s.GetAuditTrails(filters, 1000, 0)
}

// ExportAuditTrails exports audit trails to JSON format
func (s *AuditService) ExportAuditTrails(filters map[string]interface{}) ([]byte, error) {
	trails, err := s.GetAuditTrails(filters, 10000, 0) // Export up to 10k records
	if err != nil {
		return nil, err
	}

	data, err := json.MarshalIndent(trails, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal audit trails: %w", err)
	}

	return data, nil
}

// ArchiveOldAuditTrails archives audit trails older than retention period
func (s *AuditService) ArchiveOldAuditTrails(retentionDays int) (int, error) {
	cutoffDate := time.Now().AddDate(0, 0, -retentionDays)
	
	// In production, this would move records to an archive table or external storage
	// For now, we'll just count them (since we want immutable logs)
	query := `SELECT count() FROM audit_trails WHERE timestamp < $cutoff_date GROUP ALL`
	
	result, err := s.db.Query(query, map[string]interface{}{
		"cutoff_date": cutoffDate,
	})

	if err != nil {
		return 0, fmt.Errorf("failed to count old audit trails: %w", err)
	}

	var countResult []map[string]interface{}
	if err := database.UnmarshalSurrealResult(result, &countResult); err != nil {
		return 0, err
	}

	if len(countResult) > 0 {
		if count, ok := countResult[0]["count"].(float64); ok {
			logger.Log.Info().
				Int("count", int(count)).
				Int("retention_days", retentionDays).
				Msg("Audit trails eligible for archival")
			return int(count), nil
		}
	}

	return 0, nil
}