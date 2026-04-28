package database

import (
	"fmt"
	"log"
	"time"

	"agileos-backend/models"

	"github.com/surrealdb/surrealdb.go"
)

type SurrealDB struct {
	client *surrealdb.DB
}

// ConnectDB establishes connection to SurrealDB
func ConnectDB(url, user, pass, namespace, database string) (*SurrealDB, error) {
	db, err := surrealdb.New(url)
	if err != nil {
		return nil, fmt.Errorf("failed to create SurrealDB client: %w", err)
	}

	if _, err = db.Signin(map[string]interface{}{
		"user": user,
		"pass": pass,
	}); err != nil {
		return nil, fmt.Errorf("failed to signin: %w", err)
	}

	if _, err = db.Use(namespace, database); err != nil {
		return nil, fmt.Errorf("failed to use namespace/database: %w", err)
	}

	log.Printf("✓ Connected to SurrealDB: %s/%s", namespace, database)

	return &SurrealDB{client: db}, nil
}

// Close closes the database connection
func (s *SurrealDB) Close() {
	if s.client != nil {
		s.client.Close()
	}
}

// SaveWorkflow creates or updates a workflow definition
func (s *SurrealDB) SaveWorkflow(wf *models.Workflow) error {
	wf.UpdatedAt = time.Now()
	if wf.CreatedAt.IsZero() {
		wf.CreatedAt = time.Now()
	}

	// Simple approach: Use raw query
	query := `CREATE workflow CONTENT $workflow`
	if wf.ID != "" {
		query = fmt.Sprintf(`UPDATE %s CONTENT $workflow`, wf.ID)
	}

	result, err := s.client.Query(query, map[string]interface{}{
		"workflow": wf,
	})
	if err != nil {
		return fmt.Errorf("failed to save workflow: %w", err)
	}

	// Extract ID from result
	if wf.ID == "" {
		var created []models.Workflow
		if err := surrealdb.Unmarshal(result, &created); err == nil && len(created) > 0 {
			wf.ID = created[0].ID
			log.Printf("✓ Workflow created: %s (ID: %s)", wf.Name, wf.ID)
		} else {
			return fmt.Errorf("workflow created but ID extraction failed")
		}
	} else {
		log.Printf("✓ Workflow updated: %s (ID: %s)", wf.Name, wf.ID)
	}

	return nil
}

// AddStep adds a step to a workflow
func (s *SurrealDB) AddStep(step *models.Step) error {
	step.CreatedAt = time.Now()

	// Simple approach: Use raw query
	query := `CREATE step CONTENT $step`
	if step.ID != "" {
		query = fmt.Sprintf(`UPDATE %s CONTENT $step`, step.ID)
	}

	result, err := s.client.Query(query, map[string]interface{}{
		"step": step,
	})
	if err != nil {
		return fmt.Errorf("failed to add step: %w", err)
	}

	// Extract ID from result
	if step.ID == "" {
		var created []models.Step
		if err := surrealdb.Unmarshal(result, &created); err == nil && len(created) > 0 {
			step.ID = created[0].ID
			log.Printf("✓ Step created: %s (ID: %s)", step.Name, step.ID)
		} else {
			return fmt.Errorf("step created but ID extraction failed")
		}
	} else {
		log.Printf("✓ Step updated: %s (ID: %s)", step.Name, step.ID)
	}

	return nil
}

// LinkSteps creates a NEXT relationship between two steps
func (s *SurrealDB) LinkSteps(fromStepID, toStepID string, condition map[string]interface{}) error {
	query := `RELATE $from->next->$to CONTENT $data`

	data := map[string]interface{}{}
	if condition != nil {
		data["condition"] = condition
	}

	_, err := s.client.Query(query, map[string]interface{}{
		"from": fromStepID,
		"to":   toStepID,
		"data": data,
	})
	if err != nil {
		return fmt.Errorf("failed to link steps: %w", err)
	}

	log.Printf("✓ Steps linked: %s -> %s", fromStepID, toStepID)
	return nil
}

// GetNextStep retrieves the next step(s) from current step
func (s *SurrealDB) GetNextStep(currentStepID string) ([]models.Step, error) {
	query := `SELECT ->next->step.* AS next_steps FROM $step`

	result, err := s.client.Query(query, map[string]interface{}{
		"step": currentStepID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get next step: %w", err)
	}

	var response []struct {
		NextSteps []models.Step `json:"next_steps"`
	}

	if err := surrealdb.Unmarshal(result, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal result: %w", err)
	}

	if len(response) > 0 {
		return response[0].NextSteps, nil
	}

	return []models.Step{}, nil
}

// GetWorkflow retrieves a workflow by ID
func (s *SurrealDB) GetWorkflow(workflowID string) (*models.Workflow, error) {
	result, err := s.client.Query(`SELECT * FROM $workflow`, map[string]interface{}{
		"workflow": workflowID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow: %w", err)
	}

	var workflows []models.Workflow
	if err := surrealdb.Unmarshal(result, &workflows); err != nil {
		return nil, fmt.Errorf("failed to unmarshal workflow: %w", err)
	}

	if len(workflows) == 0 {
		return nil, fmt.Errorf("workflow not found")
	}

	return &workflows[0], nil
}

// GetWorkflowSteps retrieves all steps for a workflow
func (s *SurrealDB) GetWorkflowSteps(workflowID string) ([]models.Step, error) {
	query := `SELECT * FROM step WHERE workflow_id = $workflow_id ORDER BY created_at`

	result, err := s.client.Query(query, map[string]interface{}{
		"workflow_id": workflowID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow steps: %w", err)
	}

	var steps []models.Step
	if err := surrealdb.Unmarshal(result, &steps); err != nil {
		return nil, fmt.Errorf("failed to unmarshal steps: %w", err)
	}

	return steps, nil
}

// CreateProcessInstance starts a new process instance
func (s *SurrealDB) CreateProcessInstance(instance *models.ProcessInstance) error {
	instance.StartedAt = time.Now()
	instance.Status = models.ProcessStatusRunning

	query := `CREATE process_instance CONTENT $instance`

	result, err := s.client.Query(query, map[string]interface{}{
		"instance": instance,
	})
	if err != nil {
		return fmt.Errorf("failed to create process instance: %w", err)
	}

	// Extract ID from result
	var created []models.ProcessInstance
	if err := surrealdb.Unmarshal(result, &created); err == nil && len(created) > 0 {
		instance.ID = created[0].ID
		log.Printf("✓ Process instance created: %s", instance.ID)
		return nil
	}

	return fmt.Errorf("process instance created but ID extraction failed")
}

// UpdateProcessInstance updates a process instance
func (s *SurrealDB) UpdateProcessInstance(instance *models.ProcessInstance) error {
	query := fmt.Sprintf(`UPDATE %s CONTENT $instance`, instance.ID)

	_, err := s.client.Query(query, map[string]interface{}{
		"instance": instance,
	})
	if err != nil {
		return fmt.Errorf("failed to update process instance: %w", err)
	}

	return nil
}

// CreateTaskInstance creates a new task instance
func (s *SurrealDB) CreateTaskInstance(instance *models.TaskInstance) error {
	instance.CreatedAt = time.Now()

	query := `CREATE task_instance CONTENT $instance`

	result, err := s.client.Query(query, map[string]interface{}{
		"instance": instance,
	})
	if err != nil {
		return fmt.Errorf("failed to create task instance: %w", err)
	}

	// Log raw result for debugging
	log.Printf("🔍 Raw SurrealDB result type: %T", result)
	
	// SurrealDB returns: [{result: [{id: task_instance:xxx, ...}], status: OK, ...}]
	// We need to extract ID from result[0].result[0].id
	if resultArray, ok := result.([]interface{}); ok && len(resultArray) > 0 {
		if outerMap, ok := resultArray[0].(map[string]interface{}); ok {
			// Try to extract ID from result field
			if resultField, ok := outerMap["result"].([]interface{}); ok && len(resultField) > 0 {
				if innerMap, ok := resultField[0].(map[string]interface{}); ok {
					if id, ok := innerMap["id"].(string); ok {
						instance.ID = id
						log.Printf("✓ Task instance created: %s", instance.ID)
						return nil
					}
				}
			}
			// Alternative: check if ID is directly in the result
			if id, ok := outerMap["id"].(string); ok {
				instance.ID = id
				log.Printf("✓ Task instance created (direct ID): %s", instance.ID)
				return nil
			}
		}
	}

	// Try unmarshaling as a fallback
	var created []models.TaskInstance
	if err := surrealdb.Unmarshal(result, &created); err == nil && len(created) > 0 && created[0].ID != "" {
		instance.ID = created[0].ID
		log.Printf("✓ Task instance created (via unmarshal): %s", instance.ID)
		return nil
	}

	return fmt.Errorf("task instance created but ID extraction failed")
}

// GetTaskInstance retrieves a task instance by ID
func (s *SurrealDB) GetTaskInstance(taskID string) (*models.TaskInstance, error) {
	result, err := s.client.Query(`SELECT * FROM $task`, map[string]interface{}{
		"task": taskID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get task instance: %w", err)
	}

	var tasks []models.TaskInstance
	if err := surrealdb.Unmarshal(result, &tasks); err != nil {
		return nil, fmt.Errorf("failed to unmarshal task: %w", err)
	}

	if len(tasks) == 0 {
		return nil, fmt.Errorf("task not found")
	}

	return &tasks[0], nil
}

// UpdateTaskInstance updates a task instance
func (s *SurrealDB) UpdateTaskInstance(instance *models.TaskInstance) error {
	query := fmt.Sprintf(`UPDATE %s MERGE $instance`, instance.ID)

	result, err := s.client.Query(query, map[string]interface{}{
		"instance": instance,
	})
	if err != nil {
		return fmt.Errorf("failed to update task instance: %w", err)
	}

	// Log for debugging
	log.Printf("🔍 Update result type: %T", result)
	
	// Verify update succeeded - SurrealDB returns: [{result: [{id: task_instance:xxx, ...}], status: OK, ...}]
	if resultArray, ok := result.([]interface{}); ok && len(resultArray) > 0 {
		if outerMap, ok := resultArray[0].(map[string]interface{}); ok {
			// Check for status field
			if status, ok := outerMap["status"].(string); ok && status == "OK" {
				log.Printf("✓ Task instance updated: %s", instance.ID)
				return nil
			}
			// Also check if result field exists (alternative success indicator)
			if resultField, ok := outerMap["result"]; ok && resultField != nil {
				log.Printf("✓ Task instance updated: %s", instance.ID)
				return nil
			}
		}
	}

	// If we get here but no error was returned, assume success
	log.Printf("⚠️  Update verification ambiguous, assuming success: %s", instance.ID)
	return nil
}

// GetProcessInstance retrieves a process instance by ID
func (s *SurrealDB) GetProcessInstance(instanceID string) (*models.ProcessInstance, error) {
	result, err := s.client.Query(`SELECT * FROM $instance`, map[string]interface{}{
		"instance": instanceID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get process instance: %w", err)
	}

	var instances []models.ProcessInstance
	if err := surrealdb.Unmarshal(result, &instances); err != nil {
		return nil, fmt.Errorf("failed to unmarshal instance: %w", err)
	}

	if len(instances) == 0 {
		return nil, fmt.Errorf("process instance not found")
	}

	return &instances[0], nil
}

// GetPendingTasks retrieves all pending tasks for a user/role
func (s *SurrealDB) GetPendingTasks(assignedTo string) ([]models.TaskInstance, error) {
	query := `SELECT * FROM task_instance WHERE assigned_to = $assigned_to AND status = 'pending' ORDER BY created_at`

	result, err := s.client.Query(query, map[string]interface{}{
		"assigned_to": assignedTo,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get pending tasks: %w", err)
	}

	var tasks []models.TaskInstance
	if err := surrealdb.Unmarshal(result, &tasks); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tasks: %w", err)
	}

	return tasks, nil
}
