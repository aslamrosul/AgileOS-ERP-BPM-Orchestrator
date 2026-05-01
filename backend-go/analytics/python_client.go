package analytics

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"agileos-backend/logger"
)

// PythonAnalyticsClient handles communication with Python analytics microservice
type PythonAnalyticsClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

// WorkflowPrediction represents prediction data from Python service
type WorkflowPrediction struct {
	WorkflowID                string                 `json:"workflow_id"`
	PredictedCompletionTime   time.Time              `json:"predicted_completion_time"`
	ConfidenceScore           float64                `json:"confidence_score"`
	EstimatedDurationMinutes  float64                `json:"estimated_duration_minutes"`
	Factors                   map[string]interface{} `json:"factors"`
}

// AnomalyDetection represents anomaly data from Python service
type AnomalyDetection struct {
	TaskID             string  `json:"task_id"`
	AnomalyType        string  `json:"anomaly_type"`
	Severity           string  `json:"severity"`
	Description        string  `json:"description"`
	ZScore             float64 `json:"z_score"`
	ExpectedDuration   float64 `json:"expected_duration"`
	ActualDuration     float64 `json:"actual_duration"`
	Recommendation     string  `json:"recommendation"`
}

// ComprehensiveAnalytics represents the full analytics response
type ComprehensiveAnalytics struct {
	Predictions        []WorkflowPrediction   `json:"predictions"`
	Anomalies          []AnomalyDetection     `json:"anomalies"`
	Insights           []string               `json:"insights"`
	PerformanceMetrics map[string]interface{} `json:"performance_metrics"`
}

// WorkflowPerformance represents detailed workflow performance metrics
type WorkflowPerformance struct {
	WorkflowID          string                 `json:"workflow_id"`
	TotalInstances      int                    `json:"total_instances"`
	CompletedInstances  int                    `json:"completed_instances"`
	CompletionRate      float64                `json:"completion_rate"`
	AvgDurationMinutes  *float64               `json:"avg_duration_minutes"`
	MinDurationMinutes  *float64               `json:"min_duration_minutes"`
	MaxDurationMinutes  *float64               `json:"max_duration_minutes"`
	StdDurationMinutes  *float64               `json:"std_duration_minutes"`
	StepBreakdown       map[string]interface{} `json:"step_breakdown"`
}

// NewPythonAnalyticsClient creates a new Python analytics client
func NewPythonAnalyticsClient() *PythonAnalyticsClient {
	baseURL := os.Getenv("PYTHON_ANALYTICS_URL")
	if baseURL == "" {
		baseURL = "http://agileos-analytics:8001"
	}

	return &PythonAnalyticsClient{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// IsHealthy checks if the Python analytics service is healthy
func (c *PythonAnalyticsClient) IsHealthy() bool {
	resp, err := c.HTTPClient.Get(c.BaseURL + "/health")
	if err != nil {
		logger.LogError("Python analytics health check failed", err, map[string]interface{}{
			"url": c.BaseURL,
		})
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// PredictWorkflowCompletion gets completion prediction for a specific workflow
func (c *PythonAnalyticsClient) PredictWorkflowCompletion(workflowID string) (*WorkflowPrediction, error) {
	url := fmt.Sprintf("%s/predict/workflow/%s", c.BaseURL, workflowID)
	
	resp, err := c.HTTPClient.Get(url)
	if err != nil {
		logger.LogError("Failed to get workflow prediction", err, map[string]interface{}{
			"workflow_id": workflowID,
			"url":         url,
		})
		return nil, fmt.Errorf("failed to call prediction service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logger.LogError("Python analytics prediction failed", nil, map[string]interface{}{
			"workflow_id":   workflowID,
			"status_code":   resp.StatusCode,
			"response_body": string(body),
		})
		return nil, fmt.Errorf("prediction service returned status %d", resp.StatusCode)
	}

	var prediction WorkflowPrediction
	if err := json.NewDecoder(resp.Body).Decode(&prediction); err != nil {
		logger.LogError("Failed to decode prediction response", err, map[string]interface{}{
			"workflow_id": workflowID,
		})
		return nil, fmt.Errorf("failed to decode prediction response: %w", err)
	}

	logger.Log.Info().
		Str("workflow_id", workflowID).
		Float64("estimated_duration", prediction.EstimatedDurationMinutes).
		Float64("confidence", prediction.ConfidenceScore).
		Msg("Received workflow prediction from Python service")

	return &prediction, nil
}

// GetAnomalies retrieves current anomalies from the Python service
func (c *PythonAnalyticsClient) GetAnomalies() ([]AnomalyDetection, error) {
	url := fmt.Sprintf("%s/anomalies", c.BaseURL)
	
	resp, err := c.HTTPClient.Get(url)
	if err != nil {
		logger.LogError("Failed to get anomalies", err, map[string]interface{}{
			"url": url,
		})
		return nil, fmt.Errorf("failed to call anomaly service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logger.LogError("Python analytics anomaly detection failed", nil, map[string]interface{}{
			"status_code":   resp.StatusCode,
			"response_body": string(body),
		})
		return nil, fmt.Errorf("anomaly service returned status %d", resp.StatusCode)
	}

	var anomalies []AnomalyDetection
	if err := json.NewDecoder(resp.Body).Decode(&anomalies); err != nil {
		logger.LogError("Failed to decode anomalies response", err, nil)
		return nil, fmt.Errorf("failed to decode anomalies response: %w", err)
	}

	logger.Log.Info().
		Int("anomaly_count", len(anomalies)).
		Msg("Received anomalies from Python service")

	return anomalies, nil
}

// GetComprehensiveAnalytics retrieves comprehensive analytics including predictions, anomalies, and insights
func (c *PythonAnalyticsClient) GetComprehensiveAnalytics() (*ComprehensiveAnalytics, error) {
	url := fmt.Sprintf("%s/analytics/comprehensive", c.BaseURL)
	
	resp, err := c.HTTPClient.Get(url)
	if err != nil {
		logger.LogError("Failed to get comprehensive analytics", err, map[string]interface{}{
			"url": url,
		})
		return nil, fmt.Errorf("failed to call comprehensive analytics service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logger.LogError("Python analytics comprehensive service failed", nil, map[string]interface{}{
			"status_code":   resp.StatusCode,
			"response_body": string(body),
		})
		return nil, fmt.Errorf("comprehensive analytics service returned status %d", resp.StatusCode)
	}

	var analytics ComprehensiveAnalytics
	if err := json.NewDecoder(resp.Body).Decode(&analytics); err != nil {
		logger.LogError("Failed to decode comprehensive analytics response", err, nil)
		return nil, fmt.Errorf("failed to decode comprehensive analytics response: %w", err)
	}

	logger.Log.Info().
		Int("predictions_count", len(analytics.Predictions)).
		Int("anomalies_count", len(analytics.Anomalies)).
		Int("insights_count", len(analytics.Insights)).
		Msg("Received comprehensive analytics from Python service")

	return &analytics, nil
}

// GetWorkflowPerformance retrieves detailed performance metrics for a specific workflow
func (c *PythonAnalyticsClient) GetWorkflowPerformance(workflowID string) (*WorkflowPerformance, error) {
	url := fmt.Sprintf("%s/analytics/workflow/%s/performance", c.BaseURL, workflowID)
	
	resp, err := c.HTTPClient.Get(url)
	if err != nil {
		logger.LogError("Failed to get workflow performance", err, map[string]interface{}{
			"workflow_id": workflowID,
			"url":         url,
		})
		return nil, fmt.Errorf("failed to call workflow performance service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logger.LogError("Python analytics workflow performance failed", nil, map[string]interface{}{
			"workflow_id":   workflowID,
			"status_code":   resp.StatusCode,
			"response_body": string(body),
		})
		return nil, fmt.Errorf("workflow performance service returned status %d", resp.StatusCode)
	}

	var performance WorkflowPerformance
	if err := json.NewDecoder(resp.Body).Decode(&performance); err != nil {
		logger.LogError("Failed to decode workflow performance response", err, map[string]interface{}{
			"workflow_id": workflowID,
		})
		return nil, fmt.Errorf("failed to decode workflow performance response: %w", err)
	}

	logger.Log.Info().
		Str("workflow_id", workflowID).
		Int("total_instances", performance.TotalInstances).
		Float64("completion_rate", performance.CompletionRate).
		Msg("Received workflow performance from Python service")

	return &performance, nil
}

// RefreshCache triggers a cache refresh in the Python service
func (c *PythonAnalyticsClient) RefreshCache() error {
	url := fmt.Sprintf("%s/analytics/refresh", c.BaseURL)
	
	resp, err := c.HTTPClient.Post(url, "application/json", bytes.NewBuffer([]byte("{}")))
	if err != nil {
		logger.LogError("Failed to refresh Python analytics cache", err, map[string]interface{}{
			"url": url,
		})
		return fmt.Errorf("failed to call cache refresh service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logger.LogError("Python analytics cache refresh failed", nil, map[string]interface{}{
			"status_code":   resp.StatusCode,
			"response_body": string(body),
		})
		return fmt.Errorf("cache refresh service returned status %d", resp.StatusCode)
	}

	logger.Log.Info().Msg("Successfully refreshed Python analytics cache")
	return nil
}