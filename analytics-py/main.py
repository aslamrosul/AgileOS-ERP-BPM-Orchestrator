"""
AgileOS Analytics Microservice
FastAPI-based AI/ML service for BPM analytics, predictions, and anomaly detection
"""

from fastapi import FastAPI, HTTPException, Depends
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel
from typing import List, Optional, Dict, Any
import pandas as pd
import numpy as np
from sklearn.linear_model import LinearRegression
from sklearn.preprocessing import StandardScaler
from scipy import stats
import httpx
import asyncio
import logging
from datetime import datetime, timedelta
import os
from contextlib import asynccontextmanager

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# Configuration
SURREAL_URL = os.getenv("SURREAL_URL", "http://agileos-db:8000")
GO_BACKEND_URL = os.getenv("GO_BACKEND_URL", "http://agileos-backend:8081")
SERVICE_PORT = int(os.getenv("SERVICE_PORT", "8001"))

# Global variables for caching
analytics_cache = {}
cache_timestamp = None
CACHE_DURATION = 300  # 5 minutes

@asynccontextmanager
async def lifespan(app: FastAPI):
    """Application lifespan events"""
    logger.info("🚀 Starting AgileOS Analytics Microservice")
    logger.info(f"   SurrealDB URL: {SURREAL_URL}")
    logger.info(f"   Go Backend URL: {GO_BACKEND_URL}")
    yield
    logger.info("🛑 Shutting down Analytics Microservice")

# FastAPI app initialization
app = FastAPI(
    title="AgileOS Analytics Microservice",
    description="AI/ML-powered analytics for BPM workflow optimization",
    version="1.0.0",
    lifespan=lifespan
)

# CORS middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Pydantic models
class TaskData(BaseModel):
    task_id: str
    workflow_id: str
    step_name: str
    assigned_to: str
    started_at: datetime
    completed_at: Optional[datetime] = None
    duration_minutes: Optional[float] = None
    status: str

class WorkflowPrediction(BaseModel):
    workflow_id: str
    predicted_completion_time: datetime
    confidence_score: float
    estimated_duration_minutes: float
    factors: Dict[str, Any]

class AnomalyDetection(BaseModel):
    task_id: str
    anomaly_type: str
    severity: str  # low, medium, high, critical
    description: str
    z_score: float
    expected_duration: float
    actual_duration: float
    recommendation: str

class AnalyticsResponse(BaseModel):
    predictions: List[WorkflowPrediction]
    anomalies: List[AnomalyDetection]
    insights: List[str]
    performance_metrics: Dict[str, Any]

# Utility functions
async def fetch_data_from_go_backend(endpoint: str) -> Dict[str, Any]:
    """Fetch data from Go backend API"""
    try:
        async with httpx.AsyncClient(timeout=30.0) as client:
            response = await client.get(f"{GO_BACKEND_URL}/api/v1/{endpoint}")
            response.raise_for_status()
            return response.json()
    except httpx.RequestError as e:
        logger.error(f"Error fetching data from Go backend: {e}")
        raise HTTPException(status_code=503, detail="Backend service unavailable")
    except httpx.HTTPStatusError as e:
        logger.error(f"HTTP error from Go backend: {e}")
        raise HTTPException(status_code=e.response.status_code, detail="Backend API error")

async def get_task_data() -> pd.DataFrame:
    """Fetch and process task data into pandas DataFrame"""
    global analytics_cache, cache_timestamp
    
    # Check cache
    if cache_timestamp and (datetime.now() - cache_timestamp).seconds < CACHE_DURATION:
        if 'task_data' in analytics_cache:
            return analytics_cache['task_data']
    
    try:
        # Fetch data from Go backend (this would call the analytics endpoints)
        # For now, we'll simulate data - in production, this would call actual endpoints
        tasks_data = await simulate_task_data()
        
        # Convert to DataFrame
        df = pd.DataFrame(tasks_data)
        
        # Data preprocessing
        if not df.empty:
            df['started_at'] = pd.to_datetime(df['started_at'])
            df['completed_at'] = pd.to_datetime(df['completed_at'])
            df['duration_minutes'] = df.apply(
                lambda row: (row['completed_at'] - row['started_at']).total_seconds() / 60 
                if pd.notna(row['completed_at']) else None, axis=1
            )
        
        # Cache the data
        analytics_cache['task_data'] = df
        cache_timestamp = datetime.now()
        
        return df
        
    except Exception as e:
        logger.error(f"Error processing task data: {e}")
        raise HTTPException(status_code=500, detail="Data processing error")

async def simulate_task_data() -> List[Dict]:
    """Simulate task data for demonstration - replace with actual API calls"""
    base_time = datetime.now() - timedelta(days=30)
    
    workflows = ["purchase_approval", "expense_approval", "hr_onboarding", "it_request"]
    steps = ["initial_review", "manager_approval", "finance_approval", "final_processing"]
    users = ["admin", "manager", "finance", "employee"]
    
    tasks = []
    for i in range(200):  # Generate 200 sample tasks
        workflow = np.random.choice(workflows)
        step = np.random.choice(steps)
        user = np.random.choice(users)
        
        # Simulate realistic durations based on step type
        base_duration = {
            "initial_review": 30,
            "manager_approval": 120,
            "finance_approval": 180,
            "final_processing": 60
        }.get(step, 60)
        
        # Add some randomness and occasional anomalies
        if np.random.random() < 0.05:  # 5% anomalies
            duration = base_duration * np.random.uniform(5, 20)  # Very long duration
        else:
            duration = base_duration * np.random.uniform(0.5, 2.0)  # Normal variation
        
        start_time = base_time + timedelta(minutes=i * 30)
        end_time = start_time + timedelta(minutes=duration)
        
        tasks.append({
            "task_id": f"task_{i+1:03d}",
            "workflow_id": workflow,
            "step_name": step,
            "assigned_to": user,
            "started_at": start_time.isoformat(),
            "completed_at": end_time.isoformat() if np.random.random() > 0.1 else None,  # 10% incomplete
            "status": "completed" if np.random.random() > 0.1 else "in_progress"
        })
    
    return tasks

def predict_completion_time(df: pd.DataFrame, workflow_id: str) -> WorkflowPrediction:
    """Predict workflow completion time using linear regression"""
    try:
        # Filter data for specific workflow
        workflow_data = df[df['workflow_id'] == workflow_id].copy()
        
        if len(workflow_data) < 5:  # Need minimum data points
            # Use overall average if insufficient workflow-specific data
            avg_duration = df['duration_minutes'].mean() if not df.empty else 120
            predicted_time = datetime.now() + timedelta(minutes=avg_duration)
            
            return WorkflowPrediction(
                workflow_id=workflow_id,
                predicted_completion_time=predicted_time,
                confidence_score=0.3,  # Low confidence due to insufficient data
                estimated_duration_minutes=avg_duration,
                factors={"data_points": len(workflow_data), "method": "global_average"}
            )
        
        # Prepare features for regression
        workflow_data = workflow_data.dropna(subset=['duration_minutes'])
        
        # Create features: hour of day, day of week, step complexity
        workflow_data['hour'] = workflow_data['started_at'].dt.hour
        workflow_data['day_of_week'] = workflow_data['started_at'].dt.dayofweek
        workflow_data['step_complexity'] = workflow_data['step_name'].map({
            'initial_review': 1,
            'manager_approval': 2,
            'finance_approval': 3,
            'final_processing': 2
        }).fillna(2)
        
        # Features and target
        features = ['hour', 'day_of_week', 'step_complexity']
        X = workflow_data[features].values
        y = workflow_data['duration_minutes'].values
        
        # Train linear regression model
        model = LinearRegression()
        model.fit(X, y)
        
        # Predict for current time
        current_hour = datetime.now().hour
        current_day = datetime.now().weekday()
        avg_complexity = workflow_data['step_complexity'].mean()
        
        prediction_features = np.array([[current_hour, current_day, avg_complexity]])
        predicted_duration = model.predict(prediction_features)[0]
        
        # Calculate confidence based on R² score
        confidence = max(0.1, min(0.95, model.score(X, y)))
        
        predicted_time = datetime.now() + timedelta(minutes=predicted_duration)
        
        return WorkflowPrediction(
            workflow_id=workflow_id,
            predicted_completion_time=predicted_time,
            confidence_score=confidence,
            estimated_duration_minutes=predicted_duration,
            factors={
                "data_points": len(workflow_data),
                "r2_score": model.score(X, y),
                "avg_historical_duration": workflow_data['duration_minutes'].mean(),
                "method": "linear_regression"
            }
        )
        
    except Exception as e:
        logger.error(f"Error in prediction: {e}")
        # Fallback to simple average
        avg_duration = df['duration_minutes'].mean() if not df.empty else 120
        return WorkflowPrediction(
            workflow_id=workflow_id,
            predicted_completion_time=datetime.now() + timedelta(minutes=avg_duration),
            confidence_score=0.2,
            estimated_duration_minutes=avg_duration,
            factors={"error": str(e), "method": "fallback_average"}
        )

def detect_anomalies(df: pd.DataFrame) -> List[AnomalyDetection]:
    """Detect anomalies using Z-score analysis"""
    anomalies = []
    
    try:
        # Filter completed tasks with valid durations
        completed_tasks = df.dropna(subset=['duration_minutes']).copy()
        
        if len(completed_tasks) < 10:  # Need minimum data for anomaly detection
            return anomalies
        
        # Group by workflow and step for more accurate anomaly detection
        for (workflow_id, step_name), group in completed_tasks.groupby(['workflow_id', 'step_name']):
            if len(group) < 3:  # Need minimum samples per group
                continue
            
            durations = group['duration_minutes'].values
            mean_duration = np.mean(durations)
            std_duration = np.std(durations)
            
            if std_duration == 0:  # Avoid division by zero
                continue
            
            # Calculate Z-scores
            z_scores = np.abs((durations - mean_duration) / std_duration)
            
            # Identify anomalies (Z-score > 2.5)
            anomaly_indices = np.where(z_scores > 2.5)[0]
            
            for idx in anomaly_indices:
                task_row = group.iloc[idx]
                z_score = z_scores[idx]
                actual_duration = durations[idx]
                
                # Determine severity
                if z_score > 4:
                    severity = "critical"
                elif z_score > 3.5:
                    severity = "high"
                elif z_score > 3:
                    severity = "medium"
                else:
                    severity = "low"
                
                # Generate recommendation
                if actual_duration > mean_duration:
                    recommendation = f"Task taking {actual_duration/mean_duration:.1f}x longer than average. Consider process optimization or resource reallocation."
                else:
                    recommendation = f"Task completed {mean_duration/actual_duration:.1f}x faster than average. Investigate for quality assurance."
                
                anomalies.append(AnomalyDetection(
                    task_id=task_row['task_id'],
                    anomaly_type="duration_outlier",
                    severity=severity,
                    description=f"Task duration ({actual_duration:.1f} min) significantly differs from expected ({mean_duration:.1f} min)",
                    z_score=z_score,
                    expected_duration=mean_duration,
                    actual_duration=actual_duration,
                    recommendation=recommendation
                ))
        
        # Sort by severity and Z-score
        severity_order = {"critical": 4, "high": 3, "medium": 2, "low": 1}
        anomalies.sort(key=lambda x: (severity_order[x.severity], x.z_score), reverse=True)
        
        return anomalies[:20]  # Return top 20 anomalies
        
    except Exception as e:
        logger.error(f"Error in anomaly detection: {e}")
        return []

def generate_insights(df: pd.DataFrame, predictions: List[WorkflowPrediction], anomalies: List[AnomalyDetection]) -> List[str]:
    """Generate business insights from analytics"""
    insights = []
    
    try:
        if df.empty:
            return ["Insufficient data for generating insights"]
        
        # Workflow performance insights
        workflow_stats = df.groupby('workflow_id')['duration_minutes'].agg(['mean', 'count']).reset_index()
        
        if not workflow_stats.empty:
            fastest_workflow = workflow_stats.loc[workflow_stats['mean'].idxmin()]
            slowest_workflow = workflow_stats.loc[workflow_stats['mean'].idxmax()]
            
            insights.append(f"Fastest workflow: {fastest_workflow['workflow_id']} (avg: {fastest_workflow['mean']:.1f} min)")
            insights.append(f"Slowest workflow: {slowest_workflow['workflow_id']} (avg: {slowest_workflow['mean']:.1f} min)")
        
        # Anomaly insights
        if anomalies:
            critical_anomalies = [a for a in anomalies if a.severity == "critical"]
            if critical_anomalies:
                insights.append(f"⚠️ {len(critical_anomalies)} critical performance anomalies detected requiring immediate attention")
        
        # Prediction insights
        if predictions:
            avg_confidence = np.mean([p.confidence_score for p in predictions])
            insights.append(f"Prediction confidence: {avg_confidence:.1%} (based on historical data patterns)")
        
        # Time-based insights
        completed_tasks = df.dropna(subset=['duration_minutes'])
        if not completed_tasks.empty:
            peak_hour = completed_tasks.groupby(completed_tasks['started_at'].dt.hour)['duration_minutes'].mean().idxmin()
            insights.append(f"Optimal processing time: {peak_hour:02d}:00 (fastest average completion)")
        
        # Bottleneck insights
        step_performance = df.groupby('step_name')['duration_minutes'].mean().sort_values(ascending=False)
        if not step_performance.empty:
            bottleneck_step = step_performance.index[0]
            insights.append(f"Process bottleneck: '{bottleneck_step}' step (avg: {step_performance.iloc[0]:.1f} min)")
        
        return insights
        
    except Exception as e:
        logger.error(f"Error generating insights: {e}")
        return [f"Error generating insights: {str(e)}"]

# API Endpoints

@app.get("/")
async def root():
    """Health check endpoint"""
    return {
        "service": "AgileOS Analytics Microservice",
        "status": "healthy",
        "version": "1.0.0",
        "capabilities": ["predictions", "anomaly_detection", "insights"]
    }

@app.get("/health")
async def health_check():
    """Detailed health check"""
    return {
        "status": "healthy",
        "timestamp": datetime.now().isoformat(),
        "cache_status": "active" if cache_timestamp else "empty",
        "cache_age_seconds": (datetime.now() - cache_timestamp).seconds if cache_timestamp else None
    }

@app.get("/predict/workflow/{workflow_id}", response_model=WorkflowPrediction)
async def predict_workflow_completion(workflow_id: str):
    """Predict completion time for a specific workflow"""
    df = await get_task_data()
    prediction = predict_completion_time(df, workflow_id)
    
    logger.info(f"Generated prediction for workflow {workflow_id}: {prediction.estimated_duration_minutes:.1f} min")
    return prediction

@app.get("/anomalies", response_model=List[AnomalyDetection])
async def get_anomalies():
    """Detect and return current anomalies"""
    df = await get_task_data()
    anomalies = detect_anomalies(df)
    
    logger.info(f"Detected {len(anomalies)} anomalies")
    return anomalies

@app.get("/analytics/comprehensive", response_model=AnalyticsResponse)
async def get_comprehensive_analytics():
    """Get comprehensive analytics including predictions, anomalies, and insights"""
    df = await get_task_data()
    
    # Get unique workflows for predictions
    workflows = df['workflow_id'].unique().tolist() if not df.empty else []
    predictions = []
    
    for workflow_id in workflows[:5]:  # Limit to top 5 workflows
        prediction = predict_completion_time(df, workflow_id)
        predictions.append(prediction)
    
    # Detect anomalies
    anomalies = detect_anomalies(df)
    
    # Generate insights
    insights = generate_insights(df, predictions, anomalies)
    
    # Calculate performance metrics
    performance_metrics = {}
    if not df.empty:
        completed_tasks = df.dropna(subset=['duration_minutes'])
        if not completed_tasks.empty:
            performance_metrics = {
                "total_tasks": len(df),
                "completed_tasks": len(completed_tasks),
                "avg_completion_time": completed_tasks['duration_minutes'].mean(),
                "completion_rate": len(completed_tasks) / len(df),
                "anomaly_rate": len(anomalies) / len(completed_tasks) if len(completed_tasks) > 0 else 0
            }
    
    logger.info(f"Generated comprehensive analytics: {len(predictions)} predictions, {len(anomalies)} anomalies")
    
    return AnalyticsResponse(
        predictions=predictions,
        anomalies=anomalies,
        insights=insights,
        performance_metrics=performance_metrics
    )

@app.post("/analytics/refresh")
async def refresh_cache():
    """Manually refresh the analytics cache"""
    global analytics_cache, cache_timestamp
    
    analytics_cache.clear()
    cache_timestamp = None
    
    # Trigger data reload
    await get_task_data()
    
    return {"status": "cache_refreshed", "timestamp": datetime.now().isoformat()}

@app.get("/analytics/workflow/{workflow_id}/performance")
async def get_workflow_performance(workflow_id: str):
    """Get detailed performance analytics for a specific workflow"""
    df = await get_task_data()
    
    workflow_data = df[df['workflow_id'] == workflow_id]
    
    if workflow_data.empty:
        raise HTTPException(status_code=404, detail="Workflow not found or no data available")
    
    completed_tasks = workflow_data.dropna(subset=['duration_minutes'])
    
    performance = {
        "workflow_id": workflow_id,
        "total_instances": len(workflow_data),
        "completed_instances": len(completed_tasks),
        "completion_rate": len(completed_tasks) / len(workflow_data) if len(workflow_data) > 0 else 0,
        "avg_duration_minutes": completed_tasks['duration_minutes'].mean() if not completed_tasks.empty else None,
        "min_duration_minutes": completed_tasks['duration_minutes'].min() if not completed_tasks.empty else None,
        "max_duration_minutes": completed_tasks['duration_minutes'].max() if not completed_tasks.empty else None,
        "std_duration_minutes": completed_tasks['duration_minutes'].std() if not completed_tasks.empty else None,
        "step_breakdown": workflow_data.groupby('step_name')['duration_minutes'].agg(['count', 'mean']).to_dict() if not workflow_data.empty else {}
    }
    
    return performance

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(
        "main:app",
        host="0.0.0.0",
        port=SERVICE_PORT,
        reload=True,
        log_level="info"
    )