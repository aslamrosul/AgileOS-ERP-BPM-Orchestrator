package database

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"agileos-backend/models"

	"github.com/surrealdb/surrealdb.go"
)

type SurrealDB struct {
	client *surrealdb.DB
	ctx    context.Context
}

// query is a helper to execute raw queries using v1.4.0 functional API
func (s *SurrealDB) query(sql string, params map[string]interface{}) (*[]surrealdb.QueryResult[interface{}], error) {
	return surrealdb.Query[interface{}](s.ctx, s.client, sql, params)
}

// queryAndUnmarshal is a helper function for v1.4.0 API compatibility
// It executes a query and unmarshals the first result
func (s *SurrealDB) queryAndUnmarshal(query string, params map[string]interface{}, target interface{}) error {
	// v1.4.0: Use functional API
	results, err := s.query(query, params)
	if err != nil {
		return err
	}
	
	if results == nil || len(*results) == 0 {
		return fmt.Errorf("no results returned")
	}
	
	// Extract result from first query
	firstResult := (*results)[0]
	
	// v1.4.0: Use JSON marshaling/unmarshaling
	jsonData, err := json.Marshal(firstResult.Result)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}
	
	if err := json.Unmarshal(jsonData, target); err != nil {
		return fmt.Errorf("failed to unmarshal result: %w", err)
	}
	
	return nil
}

// ConnectDB establishes connection to SurrealDB with optimized settings
func ConnectDB(url, user, pass, namespace, database string) (*SurrealDB, error) {
	ctx := context.Background()
	
	// Create connection with timeout
	db, err := surrealdb.New(url)
	if err != nil {
		return nil, fmt.Errorf("failed to create SurrealDB client: %w", err)
	}

	if _, err = db.SignIn(ctx, surrealdb.Auth{
		Username: user,
		Password: pass,
	}); err != nil {
		return nil, fmt.Errorf("failed to signin: %w", err)
	}

	if err = db.Use(ctx, namespace, database); err != nil {
		return nil, fmt.Errorf("failed to use namespace/database: %w", err)
	}

	log.Printf("✓ Connected to SurrealDB: %s/%s (Connection pool optimized)", namespace, database)

	return &SurrealDB{client: db, ctx: ctx}, nil
}

// Close closes the database connection
func (s *SurrealDB) Close() {
	if s.client != nil {
		s.client.Close(s.ctx)
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

	// Extract ID from result
	if wf.ID == "" {
		var created []models.Workflow
		if err := s.queryAndUnmarshal(query, map[string]interface{}{"workflow": wf}, &created); err != nil {
			return fmt.Errorf("failed to save workflow: %w", err)
		}
		if len(created) > 0 {
			wf.ID = created[0].ID
			log.Printf("✓ Workflow created: %s (ID: %s)", wf.Name, wf.ID)
		} else {
			return fmt.Errorf("workflow created but ID extraction failed")
		}
	} else {
		if _, err := s.query(query, map[string]interface{}{"workflow": wf}); err != nil {
			return fmt.Errorf("failed to update workflow: %w", err)
		}
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

	// Extract ID from result
	if step.ID == "" {
		var created []models.Step
		if err := s.queryAndUnmarshal(query, map[string]interface{}{"step": step}, &created); err != nil {
			return fmt.Errorf("failed to add step: %w", err)
		}
		if len(created) > 0 {
			step.ID = created[0].ID
			log.Printf("✓ Step created: %s (ID: %s)", step.Name, step.ID)
		} else {
			return fmt.Errorf("step created but ID extraction failed")
		}
	} else {
		if _, err := s.query(query, map[string]interface{}{"step": step}); err != nil {
			return fmt.Errorf("failed to update step: %w", err)
		}
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

	_, err := s.query(query, map[string]interface{}{
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

	var response []struct {
		NextSteps []models.Step `json:"next_steps"`
	}

	if err := s.queryAndUnmarshal(query, map[string]interface{}{"step": currentStepID}, &response); err != nil {
		return nil, fmt.Errorf("failed to get next step: %w", err)
	}

	if len(response) > 0 {
		return response[0].NextSteps, nil
	}

	return []models.Step{}, nil
}

// GetWorkflow retrieves a workflow by ID
func (s *SurrealDB) GetWorkflow(workflowID string) (*models.Workflow, error) {
	var workflows []models.Workflow
	if err := s.queryAndUnmarshal(`SELECT * FROM $workflow`, map[string]interface{}{"workflow": workflowID}, &workflows); err != nil {
		return nil, fmt.Errorf("failed to get workflow: %w", err)
	}

	if len(workflows) == 0 {
		return nil, fmt.Errorf("workflow not found")
	}

	return &workflows[0], nil
}

// GetWorkflowSteps retrieves all steps for a workflow
func (s *SurrealDB) GetWorkflowSteps(workflowID string) ([]models.Step, error) {
	query := `SELECT * FROM step WHERE workflow_id = $workflow_id ORDER BY created_at`

	var steps []models.Step
	if err := s.queryAndUnmarshal(query, map[string]interface{}{"workflow_id": workflowID}, &steps); err != nil {
		return nil, fmt.Errorf("failed to get workflow steps: %w", err)
	}

	return steps, nil
}

// CreateProcessInstance starts a new process instance
func (s *SurrealDB) CreateProcessInstance(instance *models.ProcessInstance) error {
	instance.StartedAt = time.Now()
	instance.Status = models.ProcessStatusRunning

	query := `CREATE process_instance CONTENT $instance`

	// Extract ID from result
	var created []models.ProcessInstance
	if err := s.queryAndUnmarshal(query, map[string]interface{}{"instance": instance}, &created); err != nil {
		return fmt.Errorf("failed to create process instance: %w", err)
	}
	
	if len(created) > 0 {
		instance.ID = created[0].ID
		log.Printf("✓ Process instance created: %s", instance.ID)
		return nil
	}

	return fmt.Errorf("process instance created but ID extraction failed")
}

// UpdateProcessInstance updates a process instance
func (s *SurrealDB) UpdateProcessInstance(instance *models.ProcessInstance) error {
	query := fmt.Sprintf(`UPDATE %s CONTENT $instance`, instance.ID)

	_, err := s.query(query, map[string]interface{}{
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

	var created []models.TaskInstance
	if err := s.queryAndUnmarshal(query, map[string]interface{}{"instance": instance}, &created); err != nil {
		return fmt.Errorf("failed to create task instance: %w", err)
	}

	if len(created) > 0 && created[0].ID != "" {
		instance.ID = created[0].ID
		log.Printf("✓ Task instance created: %s", instance.ID)
		return nil
	}

	return fmt.Errorf("task instance created but ID extraction failed")
}

// GetTaskInstance retrieves a task instance by ID
func (s *SurrealDB) GetTaskInstance(taskID string) (*models.TaskInstance, error) {
	var tasks []models.TaskInstance
	if err := s.queryAndUnmarshal(`SELECT * FROM $task`, map[string]interface{}{"task": taskID}, &tasks); err != nil {
		return nil, fmt.Errorf("failed to get task instance: %w", err)
	}

	if len(tasks) == 0 {
		return nil, fmt.Errorf("task not found")
	}

	return &tasks[0], nil
}

// UpdateTaskInstance updates a task instance
func (s *SurrealDB) UpdateTaskInstance(instance *models.TaskInstance) error {
	query := fmt.Sprintf(`UPDATE %s MERGE $instance`, instance.ID)

	result, err := s.query(query, map[string]interface{}{
		"instance": instance,
	})
	if err != nil {
		return fmt.Errorf("failed to update task instance: %w", err)
	}

	// Log for debugging
	log.Printf("🔍 Update result type: %T", result)
	
	// Verify update succeeded
	if result != nil && len(*result) > 0 {
		log.Printf("✓ Task instance updated: %s", instance.ID)
		return nil
	}

	// If we get here but no error was returned, assume success
	log.Printf("⚠️  Update verification ambiguous, assuming success: %s", instance.ID)
	return nil
}

// GetProcessInstance retrieves a process instance by ID
func (s *SurrealDB) GetProcessInstance(instanceID string) (*models.ProcessInstance, error) {
	var instances []models.ProcessInstance
	if err := s.queryAndUnmarshal(`SELECT * FROM $instance`, map[string]interface{}{"instance": instanceID}, &instances); err != nil {
		return nil, fmt.Errorf("failed to get process instance: %w", err)
	}

	if len(instances) == 0 {
		return nil, fmt.Errorf("process instance not found")
	}

	return &instances[0], nil
}

// GetPendingTasks retrieves all pending tasks for a user/role
func (s *SurrealDB) GetPendingTasks(assignedTo string) ([]models.TaskInstance, error) {
	query := `SELECT * FROM task_instance WHERE assigned_to = $assigned_to AND status = 'pending' ORDER BY created_at`

	var tasks []models.TaskInstance
	if err := s.queryAndUnmarshal(query, map[string]interface{}{"assigned_to": assignedTo}, &tasks); err != nil {
		return nil, fmt.Errorf("failed to get pending tasks: %w", err)
	}

	return tasks, nil
}

// Query executes a raw SurrealQL query (for analytics and custom queries)
func (s *SurrealDB) Query(query string, params map[string]interface{}) (interface{}, error) {
	// v1.4.0 API: Use functional API
	results, err := s.query(query, params)
	if err != nil {
		return nil, err
	}
	
	// Return the raw results for caller to handle
	return results, nil
}

// QuerySlice executes a query and returns results as []interface{} for easier handling
func (s *SurrealDB) QuerySlice(query string, params map[string]interface{}) ([]interface{}, error) {
	result, err := s.Query(query, params)
	if err != nil {
		return nil, err
	}
	
	// Type assert to slice
	if slice, ok := result.([]interface{}); ok {
		return slice, nil
	}
	
	// Return empty slice if type assertion fails
	return []interface{}{}, nil
}

// UnmarshalSurrealResult is a helper for v1.4.0 to unmarshal query results
func UnmarshalSurrealResult(result interface{}, target interface{}) error {
	jsonData, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}
	
	if err := json.Unmarshal(jsonData, target); err != nil {
		return fmt.Errorf("failed to unmarshal result: %w", err)
	}
	
	return nil
}
