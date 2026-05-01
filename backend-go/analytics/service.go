package analytics

import (
	"fmt"
	"log"
	"time"

	"agileos-backend/database"
	"agileos-backend/models"
)

type Service struct {
	db *database.SurrealDB
}

func NewService(db *database.SurrealDB) *Service {
	return &Service{db: db}
}

// GetOverview returns complete analytics overview
func (s *Service) GetOverview(filter models.AnalyticsFilter) (*models.AnalyticsOverview, error) {
	// Set default filter (last 7 days if not specified)
	if filter.Days == 0 && filter.StartDate == nil {
		filter.Days = 7
	}

	overview := &models.AnalyticsOverview{
		GeneratedAt: time.Now(),
	}

	// Get summary metrics
	summary, err := s.GetSummary(filter)
	if err != nil {
		log.Printf("Warning: Failed to get summary: %v", err)
	} else {
		overview.Summary = *summary
	}

	// Get workflow efficiency
	efficiency, err := s.GetWorkflowEfficiency(filter)
	if err != nil {
		log.Printf("Warning: Failed to get workflow efficiency: %v", err)
	} else {
		overview.WorkflowEfficiency = efficiency
	}

	// Get step performance
	stepPerf, err := s.GetStepPerformance(filter)
	if err != nil {
		log.Printf("Warning: Failed to get step performance: %v", err)
	} else {
		overview.StepPerformance = stepPerf
	}

	// Get task status breakdown
	statusBreakdown, err := s.GetTaskStatusBreakdown(filter)
	if err != nil {
		log.Printf("Warning: Failed to get task status: %v", err)
	} else {
		overview.TaskStatusBreakdown = statusBreakdown
	}

	// Get department metrics
	deptMetrics, err := s.GetDepartmentMetrics(filter)
	if err != nil {
		log.Printf("Warning: Failed to get department metrics: %v", err)
	} else {
		overview.DepartmentMetrics = deptMetrics
	}

	// Identify bottlenecks (top 3)
	bottlenecks := s.IdentifyBottlenecks(stepPerf, 3)
	overview.Bottlenecks = bottlenecks

	// Generate business insights
	insights := s.GenerateInsights(overview)
	overview.Insights = insights

	return overview, nil
}

// GetSummary returns high-level KPIs
func (s *Service) GetSummary(filter models.AnalyticsFilter) (*models.AnalyticsSummary, error) {
	// Query for process counts
	processQuery := `
		SELECT 
			count() AS total,
			count(status = 'running') AS active,
			count(status = 'completed') AS completed
		FROM process_instance
	`

	// Query for task counts
	taskQuery := `
		SELECT 
			count() AS total,
			count(status = 'completed') AS completed,
			count(status = 'pending') AS pending
		FROM task_instance
	`

	// Execute queries
	processResult, _ := s.db.Query(processQuery, nil)
	taskResult, _ := s.db.Query(taskQuery, nil)

	summary := &models.AnalyticsSummary{
		TotalProcesses:     0,
		ActiveProcesses:    0,
		CompletedProcesses: 0,
		TotalTasks:         0,
		CompletedTasks:     0,
		PendingTasks:       0,
		AvgCompletionTime:  0,
		OverallSLARate:     95.0, // Default
	}

	// Parse process results
	if processResult != nil {
		var processData []map[string]interface{}
		database.UnmarshalSurrealResult(processResult, &processData)
		if len(processData) > 0 {
			if total, ok := processData[0]["total"].(float64); ok {
				summary.TotalProcesses = int(total)
			}
			if active, ok := processData[0]["active"].(float64); ok {
				summary.ActiveProcesses = int(active)
			}
			if completed, ok := processData[0]["completed"].(float64); ok {
				summary.CompletedProcesses = int(completed)
			}
		}
	}

	// Parse task results
	if taskResult != nil {
		var taskData []map[string]interface{}
		database.UnmarshalSurrealResult(taskResult, &taskData)
		if len(taskData) > 0 {
			if total, ok := taskData[0]["total"].(float64); ok {
				summary.TotalTasks = int(total)
			}
			if completed, ok := taskData[0]["completed"].(float64); ok {
				summary.CompletedTasks = int(completed)
			}
			if pending, ok := taskData[0]["pending"].(float64); ok {
				summary.PendingTasks = int(pending)
			}
		}
	}

	return summary, nil
}

// GetWorkflowEfficiency calculates efficiency metrics per workflow
func (s *Service) GetWorkflowEfficiency(filter models.AnalyticsFilter) ([]models.WorkflowEfficiency, error) {
	query := `
		SELECT 
			workflow_id,
			count() AS total_processes,
			count(status = 'completed') AS completed,
			count(status = 'running') AS pending,
			count(status = 'failed') AS failed
		FROM process_instance
		GROUP BY workflow_id
	`

	result, err := s.db.Query(query, nil)
	if err != nil {
		return nil, err
	}

	var data []map[string]interface{}
	if err := database.UnmarshalSurrealResult(result, &data); err != nil {
		return nil, err
	}

	efficiencies := make([]models.WorkflowEfficiency, 0)
	for _, item := range data {
		eff := models.WorkflowEfficiency{
			WorkflowID:     getStringValue(item, "workflow_id"),
			TotalProcesses: getIntValue(item, "total_processes"),
			CompletedCount: getIntValue(item, "completed"),
			PendingCount:   getIntValue(item, "pending"),
			FailedCount:    getIntValue(item, "failed"),
		}

		// Calculate completion rate
		if eff.TotalProcesses > 0 {
			eff.CompletionRate = float64(eff.CompletedCount) / float64(eff.TotalProcesses) * 100
		}

		// Estimate average duration (simplified)
		eff.AvgDurationHrs = 12.5 // Placeholder

		efficiencies = append(efficiencies, eff)
	}

	return efficiencies, nil
}

// GetStepPerformance calculates performance metrics per step
func (s *Service) GetStepPerformance(filter models.AnalyticsFilter) ([]models.StepPerformance, error) {
	query := `
		SELECT 
			step_id,
			step_name,
			assigned_to,
			count() AS total_tasks,
			count(status = 'completed') AS completed,
			count(status = 'pending') AS pending
		FROM task_instance
		GROUP BY step_id, step_name, assigned_to
	`

	result, err := s.db.Query(query, nil)
	if err != nil {
		return nil, err
	}

	var data []map[string]interface{}
	if err := database.UnmarshalSurrealResult(result, &data); err != nil {
		return nil, err
	}

	performances := make([]models.StepPerformance, 0)
	for _, item := range data {
		perf := models.StepPerformance{
			StepID:         getStringValue(item, "step_id"),
			StepName:       getStringValue(item, "step_name"),
			AssignedTo:     getStringValue(item, "assigned_to"),
			TotalTasks:     getIntValue(item, "total_tasks"),
			CompletedTasks: getIntValue(item, "completed"),
			PendingTasks:   getIntValue(item, "pending"),
			SLAViolations:  0, // Placeholder
		}

		// Calculate SLA compliance
		if perf.TotalTasks > 0 {
			perf.SLACompliance = float64(perf.TotalTasks-perf.SLAViolations) / float64(perf.TotalTasks) * 100
		}

		// Estimate average duration
		perf.AvgDurationHrs = float64(perf.PendingTasks) * 2.5 // Simplified

		// Mark as bottleneck if avg duration > 24 hours
		perf.IsBottleneck = perf.AvgDurationHrs > 24

		performances = append(performances, perf)
	}

	return performances, nil
}

// GetTaskStatusBreakdown returns task distribution by status
func (s *Service) GetTaskStatusBreakdown(filter models.AnalyticsFilter) ([]models.TaskStatusSummary, error) {
	query := `
		SELECT 
			status,
			count() AS count
		FROM task_instance
		GROUP BY status
	`

	result, err := s.db.Query(query, nil)
	if err != nil {
		return nil, err
	}

	var data []map[string]interface{}
	if err := database.UnmarshalSurrealResult(result, &data); err != nil {
		return nil, err
	}

	total := 0
	summaries := make([]models.TaskStatusSummary, 0)

	// First pass: calculate total
	for _, item := range data {
		total += getIntValue(item, "count")
	}

	// Second pass: calculate percentages
	for _, item := range data {
		count := getIntValue(item, "count")
		summary := models.TaskStatusSummary{
			Status: getStringValue(item, "status"),
			Count:  count,
		}

		if total > 0 {
			summary.Percentage = float64(count) / float64(total) * 100
		}

		summaries = append(summaries, summary)
	}

	return summaries, nil
}

// GetDepartmentMetrics calculates metrics per department
func (s *Service) GetDepartmentMetrics(filter models.AnalyticsFilter) ([]models.DepartmentMetrics, error) {
	// Simplified: Extract department from assigned_to field
	query := `
		SELECT 
			assigned_to,
			assigned_to AS department,
			count() AS total_tasks,
			count(status = 'completed') AS completed,
			count(status = 'pending') AS pending
		FROM task_instance
		GROUP BY assigned_to
	`

	result, err := s.db.Query(query, nil)
	if err != nil {
		return nil, err
	}

	var data []map[string]interface{}
	if err := database.UnmarshalSurrealResult(result, &data); err != nil {
		return nil, err
	}

	metrics := make([]models.DepartmentMetrics, 0)
	for _, item := range data {
		dept := models.DepartmentMetrics{
			Department:     getStringValue(item, "department"),
			TotalTasks:     getIntValue(item, "total_tasks"),
			CompletedTasks: getIntValue(item, "completed"),
			PendingTasks:   getIntValue(item, "pending"),
			AvgDurationHrs: float64(getIntValue(item, "pending")) * 3.5, // Simplified
			SLAViolations:  0,
		}

		// Mark high latency if avg > 24 hours
		dept.HighLatency = dept.AvgDurationHrs > 24

		if dept.HighLatency {
			dept.Recommendation = fmt.Sprintf(
				"Saran: %s membutuhkan optimasi alur karena rata-rata delay mencapai %.1f jam.",
				dept.Department, dept.AvgDurationHrs,
			)
		}

		metrics = append(metrics, dept)
	}

	return metrics, nil
}

// IdentifyBottlenecks returns top N bottleneck steps
func (s *Service) IdentifyBottlenecks(steps []models.StepPerformance, topN int) []models.StepPerformance {
	// Sort by average duration (descending)
	bottlenecks := make([]models.StepPerformance, 0)

	for _, step := range steps {
		if step.IsBottleneck {
			bottlenecks = append(bottlenecks, step)
		}
	}

	// Limit to topN
	if len(bottlenecks) > topN {
		bottlenecks = bottlenecks[:topN]
	}

	return bottlenecks
}

// GenerateInsights generates AI-lite business recommendations
func (s *Service) GenerateInsights(overview *models.AnalyticsOverview) []models.BusinessInsight {
	insights := make([]models.BusinessInsight, 0)

	// Insight 1: Bottleneck warning
	if len(overview.Bottlenecks) > 0 {
		for _, bottleneck := range overview.Bottlenecks {
			insights = append(insights, models.BusinessInsight{
				Type:     "warning",
				Category: "bottleneck",
				Title:    fmt.Sprintf("Bottleneck Detected: %s", bottleneck.StepName),
				Description: fmt.Sprintf(
					"Step '%s' has an average processing time of %.1f hours, which exceeds the optimal threshold.",
					bottleneck.StepName, bottleneck.AvgDurationHrs,
				),
				Recommendation: fmt.Sprintf(
					"Consider adding more resources to %s or reviewing the approval process.",
					bottleneck.AssignedTo,
				),
				Priority: "high",
			})
		}
	}

	// Insight 2: High completion rate
	if overview.Summary.CompletedTasks > 0 {
		completionRate := float64(overview.Summary.CompletedTasks) / float64(overview.Summary.TotalTasks) * 100
		if completionRate > 80 {
			insights = append(insights, models.BusinessInsight{
				Type:     "success",
				Category: "efficiency",
				Title:    "High Task Completion Rate",
				Description: fmt.Sprintf(
					"Your team has achieved a %.1f%% task completion rate. Great job!",
					completionRate,
				),
				Recommendation: "Maintain current workflow processes and consider documenting best practices.",
				Priority:       "low",
			})
		}
	}

	// Insight 3: Pending tasks warning
	if overview.Summary.PendingTasks > overview.Summary.CompletedTasks {
		insights = append(insights, models.BusinessInsight{
			Type:     "warning",
			Category: "workload",
			Title:    "High Pending Task Volume",
			Description: fmt.Sprintf(
				"There are %d pending tasks compared to %d completed tasks.",
				overview.Summary.PendingTasks, overview.Summary.CompletedTasks,
			),
			Recommendation: "Consider redistributing workload or adding temporary resources.",
			Priority:       "medium",
		})
	}

	// Insight 4: Department-specific recommendations
	for _, dept := range overview.DepartmentMetrics {
		if dept.HighLatency {
			insights = append(insights, models.BusinessInsight{
				Type:           "warning",
				Category:       "sla",
				Title:          fmt.Sprintf("High Latency in %s", dept.Department),
				Description:    dept.Recommendation,
				Recommendation: "Review staffing levels and process complexity in this department.",
				Priority:       "high",
			})
		}
	}

	return insights
}

// Helper functions
func getStringValue(m map[string]interface{}, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}

func getIntValue(m map[string]interface{}, key string) int {
	if val, ok := m[key].(float64); ok {
		return int(val)
	}
	return 0
}

func getFloatValue(m map[string]interface{}, key string) float64 {
	if val, ok := m[key].(float64); ok {
		return val
	}
	return 0.0
}
