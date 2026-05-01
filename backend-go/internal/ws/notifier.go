package ws

import (
	"encoding/json"
	"fmt"
	"time"

	"agileos-backend/logger"
	"agileos-backend/messaging"

	"github.com/nats-io/nats.go"
)

// Notifier handles NATS to WebSocket message bridging
type Notifier struct {
	hub        *Hub
	natsClient *messaging.NATSClient
}

// NewNotifier creates a new NATS to WebSocket notifier
func NewNotifier(hub *Hub, natsClient *messaging.NATSClient) *Notifier {
	return &Notifier{
		hub:        hub,
		natsClient: natsClient,
	}
}

// Start begins listening for NATS events and forwarding them to WebSocket clients
func (n *Notifier) Start() error {
	// Subscribe to task assignment events
	if err := n.subscribeToTaskAssigned(); err != nil {
		return fmt.Errorf("failed to subscribe to task.assigned: %w", err)
	}

	// Subscribe to task completion events
	if err := n.subscribeToTaskCompleted(); err != nil {
		return fmt.Errorf("failed to subscribe to task.completed: %w", err)
	}

	// Subscribe to process events
	if err := n.subscribeToProcessEvents(); err != nil {
		return fmt.Errorf("failed to subscribe to process events: %w", err)
	}

	logger.Log.Info().Msg("🔔 NATS to WebSocket notifier started")
	return nil
}

// subscribeToTaskAssigned subscribes to task assignment events
func (n *Notifier) subscribeToTaskAssigned() error {
	_, err := n.natsClient.Conn.Subscribe("task.assigned", func(msg *nats.Msg) {
		var event messaging.TaskStartedEvent
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			logger.LogError("Failed to unmarshal task.assigned event", err, nil)
			return
		}

		// Create notification for assigned user
		notification := NotificationMessage{
			ID:         fmt.Sprintf("task_assigned_%s", event.TaskID),
			Type:       "task_assigned",
			Title:      "New Task Assigned",
			Message:    fmt.Sprintf("You have been assigned a new task: %s", event.StepName),
			UserID:     event.AssignedTo,
			TaskID:     event.TaskID,
			ProcessID:  event.ProcessInstanceID,
			WorkflowID: event.StepID,
			Priority:   "medium",
			ActionURL:  fmt.Sprintf("/tasks/%s", event.TaskID),
			Data: map[string]interface{}{
				"step_name":           event.StepName,
				"process_instance_id": event.ProcessInstanceID,
				"assigned_to":         event.AssignedTo,
				"started_at":          event.StartedAt,
			},
			Timestamp: getCurrentTimestamp(),
		}

		// Send notification to assigned user
		n.hub.SendToUser(event.AssignedTo, notification)

		logger.Log.Info().
			Str("task_id", event.TaskID).
			Str("assigned_to", event.AssignedTo).
			Str("step_name", event.StepName).
			Msg("Task assignment notification sent")
	})

	return err
}

// subscribeToTaskCompleted subscribes to task completion events
func (n *Notifier) subscribeToTaskCompleted() error {
	_, err := n.natsClient.Conn.Subscribe("task.completed", func(msg *nats.Msg) {
		var event messaging.TaskCompletedEvent
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			logger.LogError("Failed to unmarshal task.completed event", err, nil)
			return
		}

		// Create notification for process initiator and stakeholders
		notification := NotificationMessage{
			ID:         fmt.Sprintf("task_completed_%s", event.TaskID),
			Type:       "task_completed",
			Title:      "Task Completed",
			Message:    fmt.Sprintf("Task has been completed by %s", event.ExecutedBy),
			TaskID:     event.TaskID,
			ProcessID:  event.ProcessInstanceID,
			Priority:   "low",
			ActionURL:  fmt.Sprintf("/tasks/%s", event.TaskID),
			Data: map[string]interface{}{
				"executed_by":         event.ExecutedBy,
				"result":              event.Result,
				"completed_at":        event.CompletedAt,
				"process_instance_id": event.ProcessInstanceID,
			},
			Timestamp: getCurrentTimestamp(),
		}

		// Send to process stakeholders (you might want to get this from database)
		// For now, we'll send to the executor
		notification.UserID = event.ExecutedBy
		n.hub.SendToUser(event.ExecutedBy, notification)

		logger.Log.Info().
			Str("task_id", event.TaskID).
			Str("executed_by", event.ExecutedBy).
			Msg("Task completion notification sent")
	})

	return err
}

// subscribeToProcessEvents subscribes to process-related events
func (n *Notifier) subscribeToProcessEvents() error {
	// Subscribe to process started events
	_, err := n.natsClient.Conn.Subscribe("process.started", func(msg *nats.Msg) {
		// Handle process started notifications
		logger.Log.Debug().Msg("Process started event received")
	})

	if err != nil {
		return err
	}

	// Subscribe to process completed events
	_, err = n.natsClient.Conn.Subscribe("process.completed", func(msg *nats.Msg) {
		// Handle process completion notifications
		logger.Log.Debug().Msg("Process completed event received")
	})

	return err
}

// SendTaskAssignmentNotification sends a notification when a task is assigned
func (n *Notifier) SendTaskAssignmentNotification(taskID, assignedTo, stepName, processID string) {
	notification := NotificationMessage{
		ID:         fmt.Sprintf("task_assigned_%s_%d", taskID, time.Now().Unix()),
		Type:       "task_assigned",
		Title:      "New Task Assignment",
		Message:    fmt.Sprintf("You have been assigned: %s", stepName),
		UserID:     assignedTo,
		TaskID:     taskID,
		ProcessID:  processID,
		Priority:   "high",
		ActionURL:  fmt.Sprintf("/tasks/%s", taskID),
		Data: map[string]interface{}{
			"step_name":  stepName,
			"process_id": processID,
		},
		Timestamp: getCurrentTimestamp(),
	}

	n.hub.SendToUser(assignedTo, notification)

	logger.LogAudit("task_assignment_notification", assignedTo, taskID, map[string]interface{}{
		"step_name":  stepName,
		"process_id": processID,
	})
}

// SendApprovalNotification sends a notification for approval requests
func (n *Notifier) SendApprovalNotification(taskID, approverID, requesterName, amount string) {
	notification := NotificationMessage{
		ID:        fmt.Sprintf("approval_request_%s", taskID),
		Type:      "approval_request",
		Title:     "Approval Required",
		Message:   fmt.Sprintf("Approval request from %s for amount %s", requesterName, amount),
		UserID:    approverID,
		TaskID:    taskID,
		Priority:  "high",
		ActionURL: fmt.Sprintf("/approvals/%s", taskID),
		Data: map[string]interface{}{
			"requester": requesterName,
			"amount":    amount,
		},
		Timestamp: getCurrentTimestamp(),
	}

	n.hub.SendToUser(approverID, notification)
}

// SendSystemNotification sends a system-wide notification
func (n *Notifier) SendSystemNotification(title, message, priority string) {
	notification := NotificationMessage{
		ID:        fmt.Sprintf("system_%d", time.Now().Unix()),
		Type:      "system_notification",
		Title:     title,
		Message:   message,
		Priority:  priority,
		Timestamp: getCurrentTimestamp(),
	}

	n.hub.BroadcastToAll(notification)

	logger.LogAudit("system_notification", "system", "", map[string]interface{}{
		"title":    title,
		"message":  message,
		"priority": priority,
	})
}

// SendDigitalSignatureNotification sends notification when a digital signature is generated
func (n *Notifier) SendDigitalSignatureNotification(taskID, userID, signature string) {
	notification := NotificationMessage{
		ID:        fmt.Sprintf("signature_generated_%s", taskID),
		Type:      "signature_generated",
		Title:     "Digital Signature Generated",
		Message:   "Your approval has been digitally signed and secured",
		UserID:    userID,
		TaskID:    taskID,
		Priority:  "medium",
		ActionURL: fmt.Sprintf("/signatures/%s", taskID),
		Data: map[string]interface{}{
			"signature": signature[:16] + "...", // Show only first 16 chars
		},
		Timestamp: getCurrentTimestamp(),
	}

	n.hub.SendToUser(userID, notification)
}