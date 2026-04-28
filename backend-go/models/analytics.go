package models

import "time"

// WorkflowEfficiency represents workflow performance metrics
type WorkflowEfficiency struct {
	WorkflowID      string        `json:"workflow_id"`
	WorkflowName    string        `json:"workflow_name"`
	TotalProcesses  int           `json:"total_processes"`
	CompletedCount  int           `json:"completed_count"`
	PendingCount    int           `json:"pending_count"`
	FailedCount     int           `json:"failed_count"`
	AvgDuration     time.Duration `json:"avg_duration"`
	AvgDurationHrs  float64       `json:"avg_duration_hours"`
	CompletionRate  float64       `json:"completion_rate"`
}

// StepPerformance represents performance metrics for a workflow step
type StepPerformance struct {
	StepID         string        `json:"step_id"`
	StepName       string        `json:"step_name"`
	AssignedTo     string        `json:"assigned_to"`
	TotalTasks     int           `json:"total_tasks"`
	CompletedTasks int           `json:"completed_tasks"`
	PendingTasks   int           `json:"pending_tasks"`
	AvgDuration    time.Duration `json:"avg_duration"`
	AvgDurationHrs float64       `json:"avg_duration_hours"`
	SLAViolations  int           `json:"sla_violations"`
	SLACompliance  float64       `json:"sla_compliance"`
	IsBottleneck   bool          `json:"is_bottleneck"`
}

// TaskStatusSummary represents task status distribution
type TaskStatusSummary struct {
	Status string `json:"status"`
	Count  int    `json:"count"`
	Percentage float64 `json:"percentage"`
}

// DepartmentMetrics represents department-level performance
type DepartmentMetrics struct {
	Department     string  `json:"department"`
	TotalTasks     int     `json:"total_tasks"`
	CompletedTasks int     `json:"completed_tasks"`
	PendingTasks   int     `json:"pending_tasks"`
	AvgDurationHrs float64 `json:"avg_duration_hours"`
	SLAViolations  int     `json:"sla_violations"`
	HighLatency    bool    `json:"high_latency"`
	Recommendation string  `json:"recommendation,omitempty"`
}

// AnalyticsOverview represents the complete analytics dashboard data
type AnalyticsOverview struct {
	Summary            AnalyticsSummary      `json:"summary"`
	WorkflowEfficiency []WorkflowEfficiency  `json:"workflow_efficiency"`
	StepPerformance    []StepPerformance     `json:"step_performance"`
	TaskStatusBreakdown []TaskStatusSummary  `json:"task_status_breakdown"`
	DepartmentMetrics  []DepartmentMetrics   `json:"department_metrics"`
	Bottlenecks        []StepPerformance     `json:"bottlenecks"`
	Insights           []BusinessInsight     `json:"insights"`
	GeneratedAt        time.Time             `json:"generated_at"`
}

// AnalyticsSummary represents high-level KPIs
type AnalyticsSummary struct {
	TotalProcesses     int     `json:"total_processes"`
	ActiveProcesses    int     `json:"active_processes"`
	CompletedProcesses int     `json:"completed_processes"`
	TotalTasks         int     `json:"total_tasks"`
	CompletedTasks     int     `json:"completed_tasks"`
	PendingTasks       int     `json:"pending_tasks"`
	AvgCompletionTime  float64 `json:"avg_completion_time_hours"`
	OverallSLARate     float64 `json:"overall_sla_compliance_rate"`
}

// BusinessInsight represents AI-generated business recommendations
type BusinessInsight struct {
	Type        string `json:"type"` // warning, info, success
	Category    string `json:"category"` // bottleneck, efficiency, sla
	Title       string `json:"title"`
	Description string `json:"description"`
	Recommendation string `json:"recommendation"`
	Priority    string `json:"priority"` // high, medium, low
}

// TimeSeriesData represents time-based metrics
type TimeSeriesData struct {
	Date           string `json:"date"`
	CompletedTasks int    `json:"completed_tasks"`
	PendingTasks   int    `json:"pending_tasks"`
	FailedTasks    int    `json:"failed_tasks"`
}

// AnalyticsFilter represents filter options for analytics queries
type AnalyticsFilter struct {
	StartDate  *time.Time `json:"start_date"`
	EndDate    *time.Time `json:"end_date"`
	WorkflowID string     `json:"workflow_id"`
	Department string     `json:"department"`
	Days       int        `json:"days"` // Last N days
}
