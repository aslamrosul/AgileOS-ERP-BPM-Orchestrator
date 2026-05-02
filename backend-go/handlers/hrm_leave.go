package handlers

import (
	"fmt"
	"net/http"
	"time"

	"agileos-backend/logger"
	"agileos-backend/models"

	"github.com/gin-gonic/gin"
)

// ============================================
// LEAVE MANAGEMENT HANDLERS
// ============================================

// CreateLeaveRequest creates a new leave request
func (h *HRMHandler) CreateLeaveRequest(c *gin.Context) {
	var leaveRequest models.LeaveRequest
	if err := c.ShouldBindJSON(&leaveRequest); err != nil {
		logger.LogError("Failed to bind leave request data", err, nil)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Auto-generate request number
	year := time.Now().Year()
	requests, err := h.db.QuerySlice(
		fmt.Sprintf("SELECT request_number FROM leave_request WHERE request_number LIKE 'LR-%d-%%' ORDER BY request_number DESC LIMIT 1", year),
		nil,
	)
	if err != nil {
		logger.LogError("Failed to get last request number", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate request number"})
		return
	}

	requestNumber := fmt.Sprintf("LR-%d-0001", year)
	if len(requests) > 0 {
		lastNumber := requests[0].(map[string]interface{})["request_number"].(string)
		var lastNum int
		fmt.Sscanf(lastNumber, fmt.Sprintf("LR-%d-%%d", year), &lastNum)
		requestNumber = fmt.Sprintf("LR-%d-%04d", year, lastNum+1)
	}

	// Calculate total days
	totalDays := int(leaveRequest.EndDate.Sub(leaveRequest.StartDate).Hours()/24) + 1

	leaveRequest.RequestNumber = requestNumber
	leaveRequest.TotalDays = totalDays
	leaveRequest.CreatedBy = userID.(string)
	leaveRequest.CreatedAt = time.Now()
	leaveRequest.UpdatedAt = time.Now()
	leaveRequest.Status = models.LeaveStatusPending

	query := `CREATE leave_request CONTENT {
		request_number: $request_number,
		employee_id: $employee_id,
		employee_code: $employee_code,
		employee_name: $employee_name,
		leave_type_id: $leave_type_id,
		leave_type_name: $leave_type_name,
		start_date: $start_date,
		end_date: $end_date,
		total_days: $total_days,
		reason: $reason,
		document_url: $document_url,
		status: $status,
		created_by: $created_by,
		created_at: $created_at,
		updated_at: $updated_at
	}`

	params := map[string]interface{}{
		"request_number":  leaveRequest.RequestNumber,
		"employee_id":     leaveRequest.EmployeeID,
		"employee_code":   leaveRequest.EmployeeCode,
		"employee_name":   leaveRequest.EmployeeName,
		"leave_type_id":   leaveRequest.LeaveTypeID,
		"leave_type_name": leaveRequest.LeaveTypeName,
		"start_date":      leaveRequest.StartDate,
		"end_date":        leaveRequest.EndDate,
		"total_days":      leaveRequest.TotalDays,
		"reason":          leaveRequest.Reason,
		"document_url":    leaveRequest.DocumentURL,
		"status":          leaveRequest.Status,
		"created_by":      leaveRequest.CreatedBy,
		"created_at":      leaveRequest.CreatedAt,
		"updated_at":      leaveRequest.UpdatedAt,
	}

	result, err := h.db.QuerySlice(query, params)
	if err != nil {
		logger.LogError("Failed to create leave request", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create leave request"})
		return
	}

	logger.Log.Info().
		Str("request_number", leaveRequest.RequestNumber).
		Str("employee_name", leaveRequest.EmployeeName).
		Msg("Leave request created successfully")

	c.JSON(http.StatusCreated, result[0])
}

// GetLeaveRequests retrieves all leave requests with filters
func (h *HRMHandler) GetLeaveRequests(c *gin.Context) {
	employeeID := c.Query("employee_id")
	status := c.Query("status")
	leaveTypeID := c.Query("leave_type_id")

	query := "SELECT * FROM leave_request"
	params := make(map[string]interface{})

	var conditions []string
	if employeeID != "" {
		conditions = append(conditions, "employee_id = $employee_id")
		params["employee_id"] = employeeID
	}
	if status != "" {
		conditions = append(conditions, "status = $status")
		params["status"] = status
	}
	if leaveTypeID != "" {
		conditions = append(conditions, "leave_type_id = $leave_type_id")
		params["leave_type_id"] = leaveTypeID
	}

	if len(conditions) > 0 {
		query += " WHERE " + conditions[0]
		for i := 1; i < len(conditions); i++ {
			query += " AND " + conditions[i]
		}
	}

	query += " ORDER BY created_at DESC"

	requests, err := h.db.QuerySlice(query, params)
	if err != nil {
		logger.LogError("Failed to get leave requests", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve leave requests"})
		return
	}

	c.JSON(http.StatusOK, requests)
}

// GetLeaveRequest retrieves a leave request by ID
func (h *HRMHandler) GetLeaveRequest(c *gin.Context) {
	requestID := c.Param("id")

	requests, err := h.db.QuerySlice(
		"SELECT * FROM $id",
		map[string]interface{}{"id": requestID},
	)
	if err != nil || len(requests) == 0 {
		logger.LogError("Leave request not found", err, map[string]interface{}{"request_id": requestID})
		c.JSON(http.StatusNotFound, gin.H{"error": "Leave request not found"})
		return
	}

	c.JSON(http.StatusOK, requests[0])
}

// ApproveLeaveRequest approves a leave request
func (h *HRMHandler) ApproveLeaveRequest(c *gin.Context) {
	requestID := c.Param("id")

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req struct {
		ApprovalNotes string `json:"approval_notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.LogError("Failed to bind approval data", err, nil)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if request is pending
	existing, err := h.db.QuerySlice(
		"SELECT * FROM $id",
		map[string]interface{}{"id": requestID},
	)
	if err != nil || len(existing) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Leave request not found"})
		return
	}

	existingRequest := existing[0].(map[string]interface{})
	if existingRequest["status"] != "pending" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only pending requests can be approved"})
		return
	}

	query := `UPDATE $id SET 
		status = 'approved',
		approver_id = $approver_id,
		approval_notes = $approval_notes,
		approved_at = $approved_at,
		updated_at = $updated_at`
	
	params := map[string]interface{}{
		"id":             requestID,
		"approver_id":    userID.(string),
		"approval_notes": req.ApprovalNotes,
		"approved_at":    time.Now(),
		"updated_at":     time.Now(),
	}

	result, err := h.db.QuerySlice(query, params)
	if err != nil || len(result) == 0 {
		logger.LogError("Failed to approve leave request", err, map[string]interface{}{"request_id": requestID})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to approve leave request"})
		return
	}

	logger.Log.Info().
		Str("request_id", requestID).
		Msg("Leave request approved successfully")

	c.JSON(http.StatusOK, result[0])
}

// RejectLeaveRequest rejects a leave request
func (h *HRMHandler) RejectLeaveRequest(c *gin.Context) {
	requestID := c.Param("id")

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req struct {
		ApprovalNotes string `json:"approval_notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.LogError("Failed to bind rejection data", err, nil)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if request is pending
	existing, err := h.db.QuerySlice(
		"SELECT * FROM $id",
		map[string]interface{}{"id": requestID},
	)
	if err != nil || len(existing) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Leave request not found"})
		return
	}

	existingRequest := existing[0].(map[string]interface{})
	if existingRequest["status"] != "pending" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only pending requests can be rejected"})
		return
	}

	query := `UPDATE $id SET 
		status = 'rejected',
		approver_id = $approver_id,
		approval_notes = $approval_notes,
		updated_at = $updated_at`
	
	params := map[string]interface{}{
		"id":             requestID,
		"approver_id":    userID.(string),
		"approval_notes": req.ApprovalNotes,
		"updated_at":     time.Now(),
	}

	result, err := h.db.QuerySlice(query, params)
	if err != nil || len(result) == 0 {
		logger.LogError("Failed to reject leave request", err, map[string]interface{}{"request_id": requestID})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reject leave request"})
		return
	}

	logger.Log.Info().
		Str("request_id", requestID).
		Msg("Leave request rejected successfully")

	c.JSON(http.StatusOK, result[0])
}

// CancelLeaveRequest cancels a leave request
func (h *HRMHandler) CancelLeaveRequest(c *gin.Context) {
	requestID := c.Param("id")

	// Check if request can be cancelled
	existing, err := h.db.QuerySlice(
		"SELECT * FROM $id",
		map[string]interface{}{"id": requestID},
	)
	if err != nil || len(existing) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Leave request not found"})
		return
	}

	existingRequest := existing[0].(map[string]interface{})
	status := existingRequest["status"].(string)
	if status == "cancelled" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Request is already cancelled"})
		return
	}

	query := `UPDATE $id SET status = 'cancelled', updated_at = $updated_at`
	params := map[string]interface{}{
		"id":         requestID,
		"updated_at": time.Now(),
	}

	result, err := h.db.QuerySlice(query, params)
	if err != nil || len(result) == 0 {
		logger.LogError("Failed to cancel leave request", err, map[string]interface{}{"request_id": requestID})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel leave request"})
		return
	}

	logger.Log.Info().
		Str("request_id", requestID).
		Msg("Leave request cancelled successfully")

	c.JSON(http.StatusOK, result[0])
}

// GetLeaveBalance retrieves leave balance for an employee
func (h *HRMHandler) GetLeaveBalance(c *gin.Context) {
	employeeID := c.Query("employee_id")
	year := c.Query("year")

	if employeeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "employee_id is required"})
		return
	}

	if year == "" {
		year = fmt.Sprintf("%d", time.Now().Year())
	}

	balances, err := h.db.QuerySlice(
		"SELECT * FROM leave_balance WHERE employee_id = $employee_id AND year = $year",
		map[string]interface{}{
			"employee_id": employeeID,
			"year":        year,
		},
	)
	if err != nil {
		logger.LogError("Failed to get leave balance", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve leave balance"})
		return
	}

	c.JSON(http.StatusOK, balances)
}
