/**
 * AI Analytics API Client
 * Integrates with Python FastAPI analytics microservice via Go backend
 */

import { api } from './api';

export interface WorkflowPrediction {
  workflow_id: string;
  predicted_completion_time: string;
  confidence_score: number;
  estimated_duration_minutes: number;
  factors: Record<string, any>;
}

export interface AnomalyDetection {
  task_id: string;
  anomaly_type: string;
  severity: 'low' | 'medium' | 'high' | 'critical';
  description: string;
  z_score: number;
  expected_duration: number;
  actual_duration: number;
  recommendation: string;
}

export interface ComprehensiveAnalytics {
  predictions: WorkflowPrediction[];
  anomalies: AnomalyDetection[];
  insights: string[];
  performance_metrics: Record<string, any>;
}

export interface AIServiceStatus {
  service: string;
  status: 'healthy' | 'unhealthy';
  ai_capabilities: {
    predictions: boolean;
    anomaly_detection: boolean;
    comprehensive_analytics: boolean;
    workflow_performance: boolean;
  };
  fallback_available: boolean;
}

export interface WorkflowPerformance {
  workflow_id: string;
  total_instances: number;
  completed_instances: number;
  completion_rate: number;
  avg_duration_minutes?: number;
  min_duration_minutes?: number;
  max_duration_minutes?: number;
  std_duration_minutes?: number;
  step_breakdown: Record<string, any>;
}

class AIAnalyticsAPI {
  private baseUrl = '/api/v1/ai-analytics';

  /**
   * Check AI service status
   */
  async getServiceStatus(): Promise<AIServiceStatus> {
    const response = await api.get(`${this.baseUrl}/status`);
    return response.data;
  }

  /**
   * Get workflow completion prediction
   */
  async predictWorkflowCompletion(workflowId: string): Promise<{
    workflow_id: string;
    prediction: WorkflowPrediction;
    ai_powered: boolean;
    generated_at: string;
  }> {
    const response = await api.get(`${this.baseUrl}/predict/workflow/${workflowId}`);
    return response.data;
  }

  /**
   * Get current anomalies
   */
  async getAnomalies(options?: {
    severity?: 'low' | 'medium' | 'high' | 'critical';
    limit?: number;
  }): Promise<{
    anomalies: AnomalyDetection[];
    total_found: number;
    ai_powered: boolean;
    filters: Record<string, any>;
  }> {
    const params = new URLSearchParams();
    if (options?.severity) params.append('severity', options.severity);
    if (options?.limit) params.append('limit', options.limit.toString());

    const response = await api.get(`${this.baseUrl}/anomalies?${params.toString()}`);
    return response.data;
  }

  /**
   * Get comprehensive AI analytics
   */
  async getComprehensiveAnalytics(): Promise<{
    ai_analytics: ComprehensiveAnalytics;
    ai_powered: boolean;
    generated_at: Record<string, any>;
    summary: {
      predictions_count: number;
      anomalies_count: number;
      insights_count: number;
      has_performance_metrics: boolean;
    };
    anomaly_severity_breakdown: Record<string, number>;
    prediction_confidence?: {
      average: number;
      quality: 'low' | 'medium' | 'high';
    };
  }> {
    const response = await api.get(`${this.baseUrl}/comprehensive`);
    return response.data;
  }

  /**
   * Get AI-powered workflow performance analysis
   */
  async getWorkflowPerformance(workflowId: string): Promise<{
    workflow_performance: WorkflowPerformance;
    ai_powered: boolean;
    ai_insights: {
      efficiency_rating: 'excellent' | 'good' | 'fair' | 'needs_improvement';
      recommendations: string[];
    };
  }> {
    const response = await api.get(`${this.baseUrl}/workflow/${workflowId}/performance`);
    return response.data;
  }

  /**
   * Refresh AI analytics cache
   */
  async refreshCache(): Promise<{
    status: string;
    message: string;
    timestamp: string;
  }> {
    const response = await api.post(`${this.baseUrl}/refresh-cache`);
    return response.data;
  }

  /**
   * Format prediction time for display
   */
  formatPredictionTime(prediction: WorkflowPrediction): string {
    const completionTime = new Date(prediction.predicted_completion_time);
    const now = new Date();
    const diffMs = completionTime.getTime() - now.getTime();
    const diffHours = Math.round(diffMs / (1000 * 60 * 60));

    if (diffHours < 1) {
      const diffMinutes = Math.round(diffMs / (1000 * 60));
      return `${diffMinutes} minutes`;
    } else if (diffHours < 24) {
      return `${diffHours} hours`;
    } else {
      const diffDays = Math.round(diffHours / 24);
      return `${diffDays} days`;
    }
  }

  /**
   * Get severity color for anomalies
   */
  getSeverityColor(severity: string): string {
    switch (severity) {
      case 'critical':
        return 'text-red-600 bg-red-100';
      case 'high':
        return 'text-orange-600 bg-orange-100';
      case 'medium':
        return 'text-yellow-600 bg-yellow-100';
      case 'low':
      default:
        return 'text-blue-600 bg-blue-100';
    }
  }

  /**
   * Get confidence level description
   */
  getConfidenceDescription(score: number): string {
    if (score >= 0.8) return 'High confidence';
    if (score >= 0.6) return 'Medium confidence';
    if (score >= 0.4) return 'Low confidence';
    return 'Very low confidence';
  }

  /**
   * Format duration for display
   */
  formatDuration(minutes: number): string {
    if (minutes < 60) {
      return `${Math.round(minutes)} min`;
    } else if (minutes < 1440) { // Less than 24 hours
      const hours = Math.round(minutes / 60);
      return `${hours} hr`;
    } else {
      const days = Math.round(minutes / 1440);
      return `${days} day${days > 1 ? 's' : ''}`;
    }
  }
}

export const aiAnalytics = new AIAnalyticsAPI();
export default aiAnalytics;