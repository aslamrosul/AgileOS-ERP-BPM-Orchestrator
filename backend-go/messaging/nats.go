package messaging

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"agileos-backend/database"
	"agileos-backend/models"

	"github.com/nats-io/nats.go"
)

type NATSClient struct {
	conn *nats.Conn
	db   *database.SurrealDB
}

// TaskCompletedEvent represents a task completion event
type TaskCompletedEvent struct {
	TaskID            string                 `json:"task_id"`
	ProcessInstanceID string                 `json:"process_instance_id"`
	CurrentStepID     string                 `json:"current_step_id"`
	ExecutedBy        string                 `json:"executed_by"`
	Result            map[string]interface{} `json:"result"`
	CompletedAt       time.Time              `json:"completed_at"`
}

// TaskStartedEvent represents a new task being started
type TaskStartedEvent struct {
	TaskID            string    `json:"task_id"`
	ProcessInstanceID string    `json:"process_instance_id"`
	StepID            string    `json:"step_id"`
	StepName          string    `json:"step_name"`
	AssignedTo        string    `json:"assigned_to"`
	StartedAt         time.Time `json:"started_at"`
}

const (
	SubjectTaskCompleted = "task.completed"
	SubjectTaskStarted   = "task.started"
	SubjectTaskFailed    = "task.failed"
)

// InitNATS initializes NATS connection and returns client
func InitNATS(natsURL string, db *database.SurrealDB) (*NATSClient, error) {
	nc, err := nats.Connect(natsURL,
		nats.RetryOnFailedConnect(true),
		nats.MaxReconnects(10),
		nats.ReconnectWait(2*time.Second),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			log.Printf("[NATS] Disconnected: %v", err)
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			log.Printf("[NATS] Reconnected to %s", nc.ConnectedUrl())
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	log.Printf("[NATS] Connected to %s", natsURL)

	return &NATSClient{
		conn: nc,
		db:   db,
	}, nil
}

// Close closes NATS connection
func (n *NATSClient) Close() {
	if n.conn != nil {
		n.conn.Close()
		log.Println("[NATS] Connection closed")
	}
}

// PublishTaskCompleted publishes task completion event
func (n *NATSClient) PublishTaskCompleted(event TaskCompletedEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	if err := n.conn.Publish(SubjectTaskCompleted, data); err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	log.Printf("[NATS] Published: Task %s completed (Step: %s)", event.TaskID, event.CurrentStepID)
	return nil
}

// PublishTaskStarted publishes task started event
func (n *NATSClient) PublishTaskStarted(event TaskStartedEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	if err := n.conn.Publish(SubjectTaskStarted, data); err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	log.Printf("[NATS] Published: Task %s started (Step: %s - %s)", 
		event.TaskID, event.StepID, event.StepName)
	return nil
}

// SubscribeTaskEvents subscribes to task events and handles orchestration
func (n *NATSClient) SubscribeTaskEvents() error {
	// Subscribe to task.completed
	_, err := n.conn.Subscribe(SubjectTaskCompleted, func(msg *nats.Msg) {
		var event TaskCompletedEvent
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			log.Printf("[NATS] Error unmarshaling event: %v", err)
			return
		}

		log.Printf("[NATS] ⚡ Received: Task %s completed at step %s", 
			event.TaskID, event.CurrentStepID)

		// Handle orchestration
		if err := n.handleTaskCompletion(event); err != nil {
			log.Printf("[NATS] ❌ Orchestration error: %v", err)
		}
	})

	if err != nil {
		return fmt.Errorf("failed to subscribe to %s: %w", SubjectTaskCompleted, err)
	}

	log.Printf("[NATS] 📡 Subscribed to %s", SubjectTaskCompleted)

	// Subscribe to task.started (for logging/monitoring)
	_, err = n.conn.Subscribe(SubjectTaskStarted, func(msg *nats.Msg) {
		var event TaskStartedEvent
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			log.Printf("[NATS] Error unmarshaling event: %v", err)
			return
		}

		log.Printf("[NATS] 🚀 Task %s started: %s (Assigned to: %s)", 
			event.TaskID, event.StepName, event.AssignedTo)
	})

	if err != nil {
		return fmt.Errorf("failed to subscribe to %s: %w", SubjectTaskStarted, err)
	}

	log.Printf("[NATS] 📡 Subscribed to %s", SubjectTaskStarted)

	return nil
}

// handleTaskCompletion is the orchestration brain
func (n *NATSClient) handleTaskCompletion(event TaskCompletedEvent) error {
	log.Printf("[ORCHESTRATOR] 🧠 Processing completion of step: %s", event.CurrentStepID)

	// 1. Get next steps from graph
	nextSteps, err := n.db.GetNextStep(event.CurrentStepID)
	if err != nil {
		return fmt.Errorf("failed to get next steps: %w", err)
	}

	if len(nextSteps) == 0 {
		log.Printf("[ORCHESTRATOR] ✅ Workflow completed - no more steps after %s", 
			event.CurrentStepID)
		
		// Update process instance to completed
		instance, err := n.db.GetProcessInstance(event.ProcessInstanceID)
		if err == nil {
			now := time.Now()
			instance.Status = models.ProcessStatusCompleted
			instance.CompletedAt = &now
			n.db.UpdateProcessInstance(instance)
		}
		
		return nil
	}

	// 2. For each next step, create task instance and trigger
	for _, nextStep := range nextSteps {
		log.Printf("[ORCHESTRATOR] ➡️  Triggering next step: %s (%s)", 
			nextStep.Name, nextStep.Type)

		// Create task instance
		taskInstance := &models.TaskInstance{
			ProcessInstanceID: event.ProcessInstanceID,
			StepID:            nextStep.ID,
			StepName:          nextStep.Name,
			Status:            models.TaskStatusPending,
			AssignedTo:        nextStep.AssignedTo,
			CreatedAt:         time.Now(),
			DueAt:             time.Now().Add(nextStep.SLA),
		}

		if err := n.db.CreateTaskInstance(taskInstance); err != nil {
			log.Printf("[ORCHESTRATOR] ❌ Failed to create task instance: %v", err)
			continue
		}

		// Update process instance current step
		instance, err := n.db.GetProcessInstance(event.ProcessInstanceID)
		if err == nil {
			instance.CurrentStepID = nextStep.ID
			
			// Add to execution history
			instance.ExecutionHistory = append(instance.ExecutionHistory, models.ExecutionLog{
				StepID:      event.CurrentStepID,
				ExecutedAt:  event.CompletedAt,
				ExecutedBy:  event.ExecutedBy,
				Action:      "completed",
				Result:      event.Result,
				Duration:    0, // Calculate if needed
				CompletedAt: event.CompletedAt,
			})
			
			n.db.UpdateProcessInstance(instance)
		}

		// Publish task started event
		startedEvent := TaskStartedEvent{
			TaskID:            taskInstance.ID,
			ProcessInstanceID: event.ProcessInstanceID,
			StepID:            nextStep.ID,
			StepName:          nextStep.Name,
			AssignedTo:        nextStep.AssignedTo,
			StartedAt:         time.Now(),
		}

		if err := n.PublishTaskStarted(startedEvent); err != nil {
			log.Printf("[ORCHESTRATOR] ⚠️  Failed to publish task started: %v", err)
		}

		log.Printf("[ORCHESTRATOR] ✅ Task %s created for step: %s (Assigned to: %s)", 
			taskInstance.ID, nextStep.Name, nextStep.AssignedTo)
	}

	return nil
}

// StartWorker starts NATS worker in background goroutine
func (n *NATSClient) StartWorker() {
	go func() {
		log.Println("[NATS] 🔄 Worker started - listening for events...")
		
		// Keep worker alive
		select {}
	}()
}
