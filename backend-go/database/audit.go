package database

import (
	"fmt"
	"time"

	"agileos-backend/models"

	"github.com/surrealdb/surrealdb.go"
)

// CreateAuditLog saves an audit log entry to the database
func (s *SurrealDB) CreateAuditLog(log *models.AuditLog) error {
	log.Timestamp = time.Now()

	query := `CREATE audit_log CONTENT $log`

	result, err := s.client.Query(query, map[string]interface{}{
		"log": log,
	})
	if err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}

	// Extract ID from result
	if resultArray, ok := result.([]interface{}); ok && len(resultArray) > 0 {
		if outerMap, ok := resultArray[0].(map[string]interface{}); ok {
			if resultField, ok := outerMap["result"].([]interface{}); ok && len(resultField) > 0 {
				if innerMap, ok := resultField[0].(map[string]interface{}); ok {
					if id, ok := innerMap["id"].(string); ok {
						log.ID = id
						return nil
					}
				}
			}
		}
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

	result, err := s.client.Query(query, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get audit logs: %w", err)
	}

	var logs []models.AuditLog
	if err := surrealdb.Unmarshal(result, &logs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal audit logs: %w", err)
	}

	return logs, nil
}

// CreateWorkflowHistory saves a workflow snapshot for audit trail
func (s *SurrealDB) CreateWorkflowHistory(history *models.WorkflowHistory) error {
	history.ChangedAt = time.Now()

	query := `CREATE workflow_history CONTENT $history`

	result, err := s.client.Query(query, map[string]interface{}{
		"history": history,
	})
	if err != nil {
		return fmt.Errorf("failed to create workflow history: %w", err)
	}

	// Extract ID
	if resultArray, ok := result.([]interface{}); ok && len(resultArray) > 0 {
		if outerMap, ok := resultArray[0].(map[string]interface{}); ok {
			if resultField, ok := outerMap["result"].([]interface{}); ok && len(resultField) > 0 {
				if innerMap, ok := resultField[0].(map[string]interface{}); ok {
					if id, ok := innerMap["id"].(string); ok {
						history.ID = id
						return nil
					}
				}
			}
		}
	}

	return fmt.Errorf("workflow history created but ID extraction failed")
}

// GetWorkflowHistory retrieves workflow change history
func (s *SurrealDB) GetWorkflowHistory(workflowID string, limit int) ([]models.WorkflowHistory, error) {
	query := `SELECT * FROM workflow_history WHERE workflow_id = $workflow_id ORDER BY changed_at DESC`
	
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	result, err := s.client.Query(query, map[string]interface{}{
		"workflow_id": workflowID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow history: %w", err)
	}

	var history []models.WorkflowHistory
	if err := surrealdb.Unmarshal(result, &history); err != nil {
		return nil, fmt.Errorf("failed to unmarshal workflow history: %w", err)
	}

	return history, nil
}

// CreatePerformanceMetric saves a performance metric
func (s *SurrealDB) CreatePerformanceMetric(metric *models.PerformanceMetric) error {
	metric.Timestamp = time.Now()

	query := `CREATE performance_metric CONTENT $metric`

	_, err := s.client.Query(query, map[string]interface{}{
		"metric": metric,
	})
	if err != nil {
		return fmt.Errorf("failed to create performance metric: %w", err)
	}

	return nil
}
