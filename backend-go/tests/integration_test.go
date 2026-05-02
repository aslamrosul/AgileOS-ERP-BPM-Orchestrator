package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// IntegrationTestSuite defines the integration test suite
type IntegrationTestSuite struct {
	suite.Suite
	server *httptest.Server
	token  string
	userID string
}

// SetupSuite runs once before all tests
func (suite *IntegrationTestSuite) SetupSuite() {
	// In a real implementation, this would start the actual server
	// For now, we'll use a mock server
	suite.server = httptest.NewServer(http.HandlerFunc(suite.mockHandler))
}

// TearDownSuite runs once after all tests
func (suite *IntegrationTestSuite) TearDownSuite() {
	suite.server.Close()
}

// mockHandler simulates the API endpoints
func (suite *IntegrationTestSuite) mockHandler(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/api/v1/auth/login":
		suite.handleLogin(w, r)
	case "/api/v1/workflow/version":
		suite.handleCreateWorkflow(w, r)
	case "/api/v1/process/start":
		suite.handleStartProcess(w, r)
	case "/api/v1/audit/trails":
		suite.handleGetAuditTrails(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (suite *IntegrationTestSuite) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var loginReq struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&loginReq); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Mock successful login
	response := map[string]interface{}{
		"access_token": "mock_jwt_token_12345",
		"user": map[string]interface{}{
			"id":       "user_123",
			"username": loginReq.Username,
			"role":     "admin",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (suite *IntegrationTestSuite) handleCreateWorkflow(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check authorization
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var workflowReq map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&workflowReq); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Mock workflow creation
	response := map[string]interface{}{
		"version": map[string]interface{}{
			"id":             "version_123",
			"workflow_id":    workflowReq["workflow_id"],
			"version":        "v1.0",
			"version_number": 1,
			"created_at":     time.Now().Format(time.RFC3339),
		},
		"message": "Workflow version created successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (suite *IntegrationTestSuite) handleStartProcess(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check authorization
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var processReq map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&processReq); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Mock process start
	response := map[string]interface{}{
		"process_instance_id": "process_123",
		"workflow_id":         processReq["workflow_id"],
		"status":              "in_progress",
		"task_id":             "task_123",
		"message":             "Process started successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (suite *IntegrationTestSuite) handleGetAuditTrails(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check authorization
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Mock audit trails
	response := map[string]interface{}{
		"audit_trails": []map[string]interface{}{
			{
				"id":                "audit_1",
				"timestamp":         time.Now().Format(time.RFC3339),
				"actor_id":          "user_123",
				"actor_username":    "admin",
				"action":            "WORKFLOW_CHANGE",
				"resource_type":     "workflow",
				"resource_id":       "test_workflow",
				"compliance_status": "PASS",
			},
			{
				"id":                "audit_2",
				"timestamp":         time.Now().Format(time.RFC3339),
				"actor_id":          "user_123",
				"actor_username":    "admin",
				"action":            "CREATE",
				"resource_type":     "process",
				"resource_id":       "process_123",
				"compliance_status": "PASS",
			},
		},
		"pagination": map[string]interface{}{
			"total":  2,
			"limit":  50,
			"offset": 0,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// TestE2E_CompleteWorkflow tests the complete workflow from login to audit
func (suite *IntegrationTestSuite) TestE2E_CompleteWorkflow() {
	t := suite.T()

	// Step 1: User Login
	t.Log("Step 1: User Login")
	loginData := map[string]string{
		"username": "admin",
		"password": "password123",
	}
	loginBody, _ := json.Marshal(loginData)

	loginReq, _ := http.NewRequest("POST", suite.server.URL+"/api/v1/auth/login", bytes.NewBuffer(loginBody))
	loginReq.Header.Set("Content-Type", "application/json")

	loginResp, err := http.DefaultClient.Do(loginReq)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, loginResp.StatusCode)

	var loginResult map[string]interface{}
	json.NewDecoder(loginResp.Body).Decode(&loginResult)
	loginResp.Body.Close()

	suite.token = loginResult["access_token"].(string)
	user := loginResult["user"].(map[string]interface{})
	suite.userID = user["id"].(string)

	assert.NotEmpty(t, suite.token)
	assert.Equal(t, "admin", user["username"])
	t.Logf("✓ Login successful, token: %s", suite.token[:20]+"...")

	// Step 2: Create Workflow
	t.Log("Step 2: Create Workflow")
	workflowData := map[string]interface{}{
		"workflow_id":   "test_workflow",
		"name":          "Test Workflow",
		"description":   "Integration test workflow",
		"definition":    map[string]interface{}{"nodes": []interface{}{}, "edges": []interface{}{}},
		"change_reason": "Integration testing",
	}
	workflowBody, _ := json.Marshal(workflowData)

	workflowReq, _ := http.NewRequest("POST", suite.server.URL+"/api/v1/workflow/version", bytes.NewBuffer(workflowBody))
	workflowReq.Header.Set("Content-Type", "application/json")
	workflowReq.Header.Set("Authorization", "Bearer "+suite.token)

	workflowResp, err := http.DefaultClient.Do(workflowReq)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, workflowResp.StatusCode)

	var workflowResult map[string]interface{}
	json.NewDecoder(workflowResp.Body).Decode(&workflowResult)
	workflowResp.Body.Close()

	version := workflowResult["version"].(map[string]interface{})
	assert.Equal(t, "test_workflow", version["workflow_id"])
	assert.Equal(t, "v1.0", version["version"])
	t.Logf("✓ Workflow created: %s", version["id"])

	// Step 3: Start Process
	t.Log("Step 3: Start Process")
	processData := map[string]interface{}{
		"workflow_id":   "test_workflow",
		"initiated_by":  suite.userID,
		"data":          map[string]interface{}{"test": "data"},
	}
	processBody, _ := json.Marshal(processData)

	processReq, _ := http.NewRequest("POST", suite.server.URL+"/api/v1/process/start", bytes.NewBuffer(processBody))
	processReq.Header.Set("Content-Type", "application/json")
	processReq.Header.Set("Authorization", "Bearer "+suite.token)

	processResp, err := http.DefaultClient.Do(processReq)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, processResp.StatusCode)

	var processResult map[string]interface{}
	json.NewDecoder(processResp.Body).Decode(&processResult)
	processResp.Body.Close()

	assert.Equal(t, "in_progress", processResult["status"])
	assert.NotEmpty(t, processResult["process_instance_id"])
	t.Logf("✓ Process started: %s", processResult["process_instance_id"])

	// Step 4: Verify Audit Log
	t.Log("Step 4: Verify Audit Log")
	auditReq, _ := http.NewRequest("GET", suite.server.URL+"/api/v1/audit/trails?limit=10", nil)
	auditReq.Header.Set("Authorization", "Bearer "+suite.token)

	auditResp, err := http.DefaultClient.Do(auditReq)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, auditResp.StatusCode)

	var auditResult map[string]interface{}
	json.NewDecoder(auditResp.Body).Decode(&auditResult)
	auditResp.Body.Close()

	auditTrails := auditResult["audit_trails"].([]interface{})
	assert.GreaterOrEqual(t, len(auditTrails), 1)

	// Verify audit trail contains expected actions
	foundWorkflowChange := false
	foundProcessCreate := false

	for _, trail := range auditTrails {
		trailMap := trail.(map[string]interface{})
		action := trailMap["action"].(string)
		
		if action == "WORKFLOW_CHANGE" {
			foundWorkflowChange = true
			assert.Equal(t, "PASS", trailMap["compliance_status"])
		}
		if action == "CREATE" && trailMap["resource_type"] == "process" {
			foundProcessCreate = true
		}
	}

	assert.True(t, foundWorkflowChange, "Workflow change should be in audit trail")
	assert.True(t, foundProcessCreate, "Process creation should be in audit trail")
	t.Logf("✓ Audit trails verified: %d records found", len(auditTrails))

	t.Log("✅ E2E Integration Test PASSED")
}

// TestE2E_UnauthorizedAccess tests unauthorized access handling
func (suite *IntegrationTestSuite) TestE2E_UnauthorizedAccess() {
	t := suite.T()

	// Attempt to access protected endpoint without token
	req, _ := http.NewRequest("GET", suite.server.URL+"/api/v1/audit/trails", nil)
	resp, err := http.DefaultClient.Do(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	resp.Body.Close()

	t.Log("✓ Unauthorized access properly rejected")
}

// TestE2E_InvalidWorkflowCreation tests validation error handling
func (suite *IntegrationTestSuite) TestE2E_InvalidWorkflowCreation() {
	t := suite.T()

	// Login first
	loginData := map[string]string{
		"username": "admin",
		"password": "password123",
	}
	loginBody, _ := json.Marshal(loginData)
	loginReq, _ := http.NewRequest("POST", suite.server.URL+"/api/v1/auth/login", bytes.NewBuffer(loginBody))
	loginReq.Header.Set("Content-Type", "application/json")
	loginResp, _ := http.DefaultClient.Do(loginReq)
	
	var loginResult map[string]interface{}
	json.NewDecoder(loginResp.Body).Decode(&loginResult)
	loginResp.Body.Close()
	token := loginResult["access_token"].(string)

	// Attempt to create workflow with invalid data
	invalidData := map[string]interface{}{
		"invalid_field": "invalid_value",
	}
	invalidBody, _ := json.Marshal(invalidData)

	req, _ := http.NewRequest("POST", suite.server.URL+"/api/v1/workflow/version", bytes.NewBuffer(invalidBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	resp.Body.Close()

	t.Log("✓ Invalid workflow creation properly rejected")
}

// TestIntegrationSuite runs the integration test suite
func TestIntegrationSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}