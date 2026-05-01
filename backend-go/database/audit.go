package database

import (
	"fmt"
	"time"

	"agileos-backend/models"
)

// CreateAuditLog saves an audit log entry to the database
func (s *SurrealDB) CreateAuditLog(log *models.AuditLog) error {
	log.Timestamp = time.Now()

	query := `CREATE audit_log CONTENT $log`

	var created []models.AuditLog
	if err := s.queryAndUnmarshal(query, map[string]interface{}{"log": log}, &created); err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}

	if len(created) > 0 {
		log.ID = created[0].ID
		return nil
	}

	return fmt.Errorf("audit log created but ID extraction failed")
}

// GetAuditLogs retrieves audit logs with optional filters
func (s *SurrealDB) GetAuditLogs(userID string, logType string, limit int) ([]models.AuditLog, error) {
	query := `SELECT * FROM audit_log`
	params := make(map[string]interface{})

	// Add filters
	conditions := []string{}
	if userID != "" {
		conditions = append(conditions, "user_id = $user_id")
		params["user_id"] = userID
	}
	if logType != "" {
		conditions = append(conditions, "type = $type")
		params["type"] = logType
	}

	if len(conditions) > 0 {
		query += " WHERE "
		for i, cond := range conditions {
			if i > 0 {
				query += " AND "
			}
			query += cond
		}
	}

	query += " ORDER BY timestamp DESC"

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	var logs []models.AuditLog
	if err := s.queryAndUnmarshal(query, params, &logs); err != nil {
		return nil, fmt.Errorf("failed to get audit logs: %w", err)
	}

	return logs, nil
}

// CreateWorkflowHistory saves a workflow snapshot for audit trail
func (s *SurrealDB) CreateWorkflowHistory(history *models.WorkflowHistory) error {
	history.ChangedAt = time.Now()

	query := `CREATE workflow_history CONTENT $history`

	var created []models.WorkflowHistory
	if err := s.queryAndUnmarshal(query, map[string]interface{}{"history": history}, &created); err != nil {
		return fmt.Errorf("failed to create workflow history: %w", err)
	}

	if len(created) > 0 {
		history.ID = created[0].ID
		return nil
	}

	return fmt.Errorf("workflow history created but ID extraction failed")
}

// GetWorkflowHistory retrieves workflow change history
func (s *SurrealDB) GetWorkflowHistory(workflowID string, limit int) ([]models.WorkflowHistory, error) {
	query := `SELECT * FROM workflow_history WHERE workflow_id = $workflow_id ORDER BY changed_at DESC`
	
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	var history []models.WorkflowHistory
	if err := s.queryAndUnmarshal(query, map[string]interface{}{"workflow_id": workflowID}, &history); err != nil {
		return nil, fmt.Errorf("failed to get workflow history: %w", err)
	}

	return history, nil
}

// CreatePerformanceMetric saves a performance metric
func (s *SurrealDB) CreatePerformanceMetric(metric *models.PerformanceMetric) error {
	metric.Timestamp = time.Now()

	query := `CREATE performance_metric CONTENT $metric`

	_, err := s.query(query, map[string]interface{}{
		"metric": metric,
	})
	if err != nil {
		return fmt.Errorf("failed to create performance metric: %w", err)
	}

	return nil
}
