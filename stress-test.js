/**
 * AgileOS Load Testing Script (k6)
 * Tests system performance under various load conditions
 * 
 * Installation:
 * - Windows: choco install k6
 * - Mac: brew install k6
 * - Linux: sudo apt-get install k6
 * 
 * Usage:
 * k6 run stress-test.js
 * k6 run --vus 100 --duration 5m stress-test.js
 */

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');
const loginDuration = new Trend('login_duration');
const workflowDuration = new Trend('workflow_duration');
const analyticsDuration = new Trend('analytics_duration');
const processDuration = new Trend('process_duration');
const requestCounter = new Counter('requests_total');

// Test configuration
export const options = {
  stages: [
    // Ramp-up: Gradually increase load
    { duration: '2m', target: 50 },   // Ramp up to 50 users over 2 minutes
    { duration: '3m', target: 100 },  // Ramp up to 100 users over 3 minutes
    { duration: '5m', target: 100 },  // Stay at 100 users for 5 minutes
    { duration: '2m', target: 200 },  // Spike to 200 users
    { duration: '3m', target: 200 },  // Stay at 200 users for 3 minutes
    { duration: '2m', target: 0 },    // Ramp down to 0 users
  ],
  thresholds: {
    // Performance thresholds
    'http_req_duration': ['p(95)<500', 'p(99)<1000'], // 95% of requests under 500ms, 99% under 1s
    'http_req_failed': ['rate<0.05'],                  // Error rate under 5%
    'errors': ['rate<0.05'],                           // Custom error rate under 5%
    'login_duration': ['p(95)<300'],                   // Login under 300ms for 95%
    'workflow_duration': ['p(95)<800'],                // Workflow operations under 800ms
    'analytics_duration': ['p(95)<1000'],              // Analytics under 1s
  },
};

// Base URL configuration
const BASE_URL = __ENV.BASE_URL || 'http://localhost:8081';
const API_BASE = `${BASE_URL}/api/v1`;

// Test data
const testUsers = [
  { username: 'admin', password: 'password123', role: 'admin' },
  { username: 'manager', password: 'password123', role: 'manager' },
  { username: 'employee', password: 'password123', role: 'employee' },
  { username: 'finance', password: 'password123', role: 'finance' },
];

// Helper function to get random user
function getRandomUser() {
  return testUsers[Math.floor(Math.random() * testUsers.length)];
}

// Helper function to login and get token
function login() {
  const user = getRandomUser();
  const loginPayload = JSON.stringify({
    username: user.username,
    password: user.password,
  });

  const params = {
    headers: {
      'Content-Type': 'application/json',
    },
  };

  const startTime = new Date();
  const response = http.post(`${API_BASE}/auth/login`, loginPayload, params);
  const duration = new Date() - startTime;

  loginDuration.add(duration);
  requestCounter.add(1);

  const success = check(response, {
    'login status is 200': (r) => r.status === 200,
    'login has token': (r) => r.json('access_token') !== undefined,
  });

  if (!success) {
    errorRate.add(1);
    console.error(`Login failed: ${response.status} - ${response.body}`);
    return null;
  }

  errorRate.add(0);
  return response.json('access_token');
}

// Test scenario: Get workflows
function testGetWorkflows(token) {
  const params = {
    headers: {
      'Authorization': `Bearer ${token}`,
    },
  };

  const startTime = new Date();
  const response = http.get(`${API_BASE}/workflows`, params);
  const duration = new Date() - startTime;

  workflowDuration.add(duration);
  requestCounter.add(1);

  const success = check(response, {
    'get workflows status is 200': (r) => r.status === 200,
    'workflows response has data': (r) => r.json('workflows') !== undefined,
  });

  errorRate.add(success ? 0 : 1);
  return success;
}

// Test scenario: Get analytics
function testGetAnalytics(token) {
  const params = {
    headers: {
      'Authorization': `Bearer ${token}`,
    },
  };

  const startTime = new Date();
  const response = http.get(`${API_BASE}/analytics/overview`, params);
  const duration = new Date() - startTime;

  analyticsDuration.add(duration);
  requestCounter.add(1);

  const success = check(response, {
    'get analytics status is 200': (r) => r.status === 200,
    'analytics response has data': (r) => r.body.length > 0,
  });

  errorRate.add(success ? 0 : 1);
  return success;
}

// Test scenario: Start process
function testStartProcess(token) {
  const processPayload = JSON.stringify({
    workflow_id: 'purchase_approval',
    initiated_by: 'admin',
    data: {
      amount: Math.floor(Math.random() * 10000) + 1000,
      description: `Load test purchase ${Date.now()}`,
      department: 'IT',
    },
  });

  const params = {
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json',
    },
  };

  const startTime = new Date();
  const response = http.post(`${API_BASE}/process/start`, processPayload, params);
  const duration = new Date() - startTime;

  processDuration.add(duration);
  requestCounter.add(1);

  const success = check(response, {
    'start process status is 201 or 200': (r) => r.status === 201 || r.status === 200,
    'process response has instance id': (r) => {
      try {
        const json = r.json();
        return json.process_instance_id !== undefined || json.task_id !== undefined;
      } catch (e) {
        return false;
      }
    },
  });

  errorRate.add(success ? 0 : 1);
  return success;
}

// Test scenario: Get audit trails
function testGetAuditTrails(token) {
  const params = {
    headers: {
      'Authorization': `Bearer ${token}`,
    },
  };

  const response = http.get(`${API_BASE}/audit/trails?limit=10`, params);
  requestCounter.add(1);

  const success = check(response, {
    'get audit trails status is 200': (r) => r.status === 200,
  });

  errorRate.add(success ? 0 : 1);
  return success;
}

// Test scenario: Health check
function testHealthCheck() {
  const response = http.get(`${BASE_URL}/health`);
  requestCounter.add(1);

  const success = check(response, {
    'health check status is 200': (r) => r.status === 200,
    'health check has status': (r) => r.json('status') !== undefined,
  });

  errorRate.add(success ? 0 : 1);
  return success;
}

// Main test scenario
export default function () {
  // Health check (10% of requests)
  if (Math.random() < 0.1) {
    testHealthCheck();
    sleep(1);
    return;
  }

  // Login
  const token = login();
  if (!token) {
    sleep(2); // Wait before retry
    return;
  }

  sleep(0.5); // Think time

  // Simulate user behavior with weighted scenarios
  const scenario = Math.random();

  if (scenario < 0.3) {
    // 30% - View workflows
    testGetWorkflows(token);
  } else if (scenario < 0.5) {
    // 20% - View analytics
    testGetAnalytics(token);
  } else if (scenario < 0.7) {
    // 20% - Start process
    testStartProcess(token);
  } else if (scenario < 0.9) {
    // 20% - View audit trails
    testGetAuditTrails(token);
  } else {
    // 10% - Mixed operations
    testGetWorkflows(token);
    sleep(0.3);
    testGetAnalytics(token);
  }

  // Think time between requests (1-3 seconds)
  sleep(Math.random() * 2 + 1);
}

// Setup function (runs once at start)
export function setup() {
  console.log('🚀 Starting AgileOS Load Test');
  console.log(`Target: ${BASE_URL}`);
  console.log('Test stages:');
  console.log('  - Ramp up to 50 users (2m)');
  console.log('  - Ramp up to 100 users (3m)');
  console.log('  - Sustain 100 users (5m)');
  console.log('  - Spike to 200 users (2m)');
  console.log('  - Sustain 200 users (3m)');
  console.log('  - Ramp down (2m)');
  console.log('');

  // Verify system is accessible
  const healthCheck = http.get(`${BASE_URL}/health`);
  if (healthCheck.status !== 200) {
    throw new Error(`System not accessible: ${healthCheck.status}`);
  }

  console.log('✓ System is accessible');
  console.log('');
}

// Teardown function (runs once at end)
export function teardown(data) {
  console.log('');
  console.log('🏁 Load Test Completed');
  console.log('Check the summary above for detailed metrics');
}

// Handle summary for custom reporting
export function handleSummary(data) {
  const timestamp = new Date().toISOString().replace(/[:.]/g, '-');
  
  return {
    'stdout': textSummary(data, { indent: ' ', enableColors: true }),
    [`load-test-results-${timestamp}.json`]: JSON.stringify(data, null, 2),
  };
}

// Text summary helper
function textSummary(data, options) {
  const indent = options.indent || '';
  const enableColors = options.enableColors || false;

  let summary = '\n';
  summary += `${indent}📊 Load Test Summary\n`;
  summary += `${indent}${'='.repeat(50)}\n\n`;

  // Test duration
  const duration = data.state.testRunDurationMs / 1000;
  summary += `${indent}Duration: ${duration.toFixed(2)}s\n`;

  // Request metrics
  const requests = data.metrics.requests_total?.values?.count || 0;
  const rps = requests / duration;
  summary += `${indent}Total Requests: ${requests}\n`;
  summary += `${indent}Requests/sec: ${rps.toFixed(2)}\n\n`;

  // HTTP metrics
  const httpDuration = data.metrics.http_req_duration;
  if (httpDuration) {
    summary += `${indent}HTTP Request Duration:\n`;
    summary += `${indent}  avg: ${httpDuration.values.avg.toFixed(2)}ms\n`;
    summary += `${indent}  p95: ${httpDuration.values['p(95)'].toFixed(2)}ms\n`;
    summary += `${indent}  p99: ${httpDuration.values['p(99)'].toFixed(2)}ms\n`;
    summary += `${indent}  max: ${httpDuration.values.max.toFixed(2)}ms\n\n`;
  }

  // Error rate
  const errorRate = data.metrics.errors?.values?.rate || 0;
  const httpFailRate = data.metrics.http_req_failed?.values?.rate || 0;
  summary += `${indent}Error Rates:\n`;
  summary += `${indent}  Custom errors: ${(errorRate * 100).toFixed(2)}%\n`;
  summary += `${indent}  HTTP failures: ${(httpFailRate * 100).toFixed(2)}%\n\n`;

  // Custom metrics
  if (data.metrics.login_duration) {
    summary += `${indent}Login Duration (p95): ${data.metrics.login_duration.values['p(95)'].toFixed(2)}ms\n`;
  }
  if (data.metrics.workflow_duration) {
    summary += `${indent}Workflow Duration (p95): ${data.metrics.workflow_duration.values['p(95)'].toFixed(2)}ms\n`;
  }
  if (data.metrics.analytics_duration) {
    summary += `${indent}Analytics Duration (p95): ${data.metrics.analytics_duration.values['p(95)'].toFixed(2)}ms\n`;
  }

  summary += '\n';
  return summary;
}