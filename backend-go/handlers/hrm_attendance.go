package handlers

import (
	"net/http"
	"time"

	"agileos-backend/logger"
	"agileos-backend/models"

	"github.com/gin-gonic/gin"
)

// ============================================
// ATTENDANCE MANAGEMENT HANDLERS
// ============================================

// CheckIn records employee check-in
func (h *HRMHandler) CheckIn(c *gin.Context) {
	var attendance models.Attendance
	if err := c.ShouldBindJSON(&attendance); err != nil {
		logger.LogError("Failed to bind attendance data", err, nil)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if already checked in today
	today := time.Now().Format("2006-01-02")
	existing, err := h.db.QuerySlice(
		"SELECT * FROM attendance WHERE employee_id = $employee_id AND date = $date",
		map[string]interface{}{
			"employee_id": attendance.EmployeeID,
			"date":        today,
		},
	)
	if err == nil && len(existing) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Already checked in today"})
		return
	}

	checkInTime := time.Now()
	attendance.Date = checkInTime
	attendance.CheckInTime = &checkInTime
	attendance.Status = models.AttendanceStatusPresent
	attendance.CreatedAt = time.Now()
	attendance.UpdatedAt = time.Now()

	// Check if late (assuming work starts at 9:00 AM)
	workStartTime := time.Date(checkInTime.Year(), checkInTime.Month(), checkInTime.Day(), 9, 0, 0, 0, checkInTime.Location())
	if checkInTime.After(workStartTime) {
		attendance.IsLate = true
		attendance.LateMinutes = int(checkInTime.Sub(workStartTime).Minutes())
		attendance.Status = models.AttendanceStatusLate
	}

	query := `CREATE attendance CONTENT {
		employee_id: $employee_id,
		employee_code: $employee_code,
		employee_name: $employee_name,
		date: $date,
		check_in_time: $check_in_time,
		check_in_location: $check_in_location,
		status: $status,
		is_late: $is_late,
		late_minutes: $late_minutes,
		work_hours: $work_hours,
		overtime_hours: $overtime_hours,
		created_at: $created_at,
		updated_at: $updated_at
	}`

	params := map[string]interface{}{
		"employee_id":        attendance.EmployeeID,
		"employee_code":      attendance.EmployeeCode,
		"employee_name":      attendance.EmployeeName,
		"date":               attendance.Date,
		"check_in_time":      attendance.CheckInTime,
		"check_in_location":  attendance.CheckInLocation,
		"status":             attendance.Status,
		"is_late":            attendance.IsLate,
		"late_minutes":       attendance.LateMinutes,
		"work_hours":         0.0,
		"overtime_hours":     0.0,
		"created_at":         attendance.CreatedAt,
		"updated_at":         attendance.UpdatedAt,
	}

	result, err := h.db.QuerySlice(query, params)
	if err != nil {
		logger.LogError("Failed to check in", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check in"})
		return
	}

	logger.Log.Info().
		Str("employee_code", attendance.EmployeeCode).
		Time("check_in_time", checkInTime).
		Msg("Employee checked in successfully")

	c.JSON(http.StatusCreated, result[0])
}

// CheckOut records employee check-out
func (h *HRMHandler) CheckOut(c *gin.Context) {
	attendanceID := c.Param("id")

	var req struct {
		CheckOutLocation string `json:"check_out_location"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.LogError("Failed to bind check-out data", err, nil)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get attendance record
	existing, err := h.db.QuerySlice(
		"SELECT * FROM $id",
		map[string]interface{}{"id": attendanceID},
	)
	if err != nil || len(existing) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Attendance record not found"})
		return
	}

	attendance := existing[0].(map[string]interface{})
	if attendance["check_out_time"] != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Already checked out"})
		return
	}

	checkOutTime := time.Now()
	checkInTime := attendance["check_in_time"].(time.Time)
	
	// Calculate work hours
	workHours := checkOutTime.Sub(checkInTime).Hours()
	overtimeHours := 0.0
	if workHours > 8 {
		overtimeHours = workHours - 8
		workHours = 8
	}

	query := `UPDATE $id SET
		check_out_time = $check_out_time,
		check_out_location = $check_out_location,
		work_hours = $work_hours,
		overtime_hours = $overtime_hours,
		updated_at = $updated_at
	`

	params := map[string]interface{}{
		"id":                 attendanceID,
		"check_out_time":     checkOutTime,
		"check_out_location": req.CheckOutLocation,
		"work_hours":         workHours,
		"overtime_hours":     overtimeHours,
		"updated_at":         time.Now(),
	}

	result, err := h.db.QuerySlice(query, params)
	if err != nil || len(result) == 0 {
		logger.LogError("Failed to check out", err, map[string]interface{}{"attendance_id": attendanceID})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check out"})
		return
	}

	logger.Log.Info().
		Str("attendance_id", attendanceID).
		Time("check_out_time", checkOutTime).
		Float64("work_hours", workHours).
		Msg("Employee checked out successfully")

	c.JSON(http.StatusOK, result[0])
}

// GetAttendances retrieves attendance records with filters
func (h *HRMHandler) GetAttendances(c *gin.Context) {
	employeeID := c.Query("employee_id")
	fromDate := c.Query("from_date")
	toDate := c.Query("to_date")
	status := c.Query("status")

	query := "SELECT * FROM attendance"
	params := make(map[string]interface{})

	var conditions []string
	if employeeID != "" {
		conditions = append(conditions, "employee_id = $employee_id")
		params["employee_id"] = employeeID
	}
	if fromDate != "" {
		conditions = append(conditions, "date >= $from_date")
		params["from_date"] = fromDate
	}
	if toDate != "" {
		conditions = append(conditions, "date <= $to_date")
		params["to_date"] = toDate
	}
	if status != "" {
		conditions = append(conditions, "status = $status")
		params["status"] = status
	}

	if len(conditions) > 0 {
		query += " WHERE " + conditions[0]
		for i := 1; i < len(conditions); i++ {
			query += " AND " + conditions[i]
		}
	}

	query += " ORDER BY date DESC, check_in_time DESC"

	attendances, err := h.db.QuerySlice(query, params)
	if err != nil {
		logger.LogError("Failed to get attendances", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve attendances"})
		return
	}

	c.JSON(http.StatusOK, attendances)
}

// GetAttendance retrieves an attendance record by ID
func (h *HRMHandler) GetAttendance(c *gin.Context) {
	attendanceID := c.Param("id")

	attendances, err := h.db.QuerySlice(
		"SELECT * FROM $id",
		map[string]interface{}{"id": attendanceID},
	)
	if err != nil || len(attendances) == 0 {
		logger.LogError("Attendance not found", err, map[string]interface{}{"attendance_id": attendanceID})
		c.JSON(http.StatusNotFound, gin.H{"error": "Attendance not found"})
		return
	}

	c.JSON(http.StatusOK, attendances[0])
}

// UpdateAttendance updates an attendance record (admin only)
func (h *HRMHandler) UpdateAttendance(c *gin.Context) {
	attendanceID := c.Param("id")

	var attendance models.Attendance
	if err := c.ShouldBindJSON(&attendance); err != nil {
		logger.LogError("Failed to bind attendance data", err, nil)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	attendance.UpdatedAt = time.Now()

	query := `UPDATE $id SET
		status = $status,
		work_hours = $work_hours,
		overtime_hours = $overtime_hours,
		notes = $notes,
		updated_at = $updated_at
	`

	params := map[string]interface{}{
		"id":             attendanceID,
		"status":         attendance.Status,
		"work_hours":     attendance.WorkHours,
		"overtime_hours": attendance.OvertimeHours,
		"notes":          attendance.Notes,
		"updated_at":     attendance.UpdatedAt,
	}

	result, err := h.db.QuerySlice(query, params)
	if err != nil || len(result) == 0 {
		logger.LogError("Failed to update attendance", err, map[string]interface{}{"attendance_id": attendanceID})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update attendance"})
		return
	}

	logger.Log.Info().
		Str("attendance_id", attendanceID).
		Msg("Attendance updated successfully")

	c.JSON(http.StatusOK, result[0])
}

// GetAttendanceSummary retrieves attendance summary for an employee
func (h *HRMHandler) GetAttendanceSummary(c *gin.Context) {
	employeeID := c.Query("employee_id")
	month := c.Query("month")
	year := c.Query("year")

	if employeeID == "" || month == "" || year == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "employee_id, month, and year are required"})
		return
	}

	// Get all attendances for the period
	attendances, err := h.db.QuerySlice(
		`SELECT * FROM attendance 
		WHERE employee_id = $employee_id 
		AND EXTRACT(MONTH FROM date) = $month 
		AND EXTRACT(YEAR FROM date) = $year`,
		map[string]interface{}{
			"employee_id": employeeID,
			"month":       month,
			"year":        year,
		},
	)
	if err != nil {
		logger.LogError("Failed to get attendance summary", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get attendance summary"})
		return
	}

	// Calculate summary
	totalDays := len(attendances)
	presentDays := 0
	lateDays := 0
	absentDays := 0
	totalWorkHours := 0.0
	totalOvertimeHours := 0.0

	for _, att := range attendances {
		attendance := att.(map[string]interface{})
		status := attendance["status"].(string)
		
		switch status {
		case "present":
			presentDays++
		case "late":
			lateDays++
			presentDays++
		case "absent":
			absentDays++
		}

		if attendance["work_hours"] != nil {
			totalWorkHours += attendance["work_hours"].(float64)
		}
		if attendance["overtime_hours"] != nil {
			totalOvertimeHours += attendance["overtime_hours"].(float64)
		}
	}

	summary := gin.H{
		"employee_id":         employeeID,
		"month":               month,
		"year":                year,
		"total_days":          totalDays,
		"present_days":        presentDays,
		"late_days":           lateDays,
		"absent_days":         absentDays,
		"total_work_hours":    totalWorkHours,
		"total_overtime_hours": totalOvertimeHours,
	}

	c.JSON(http.StatusOK, summary)
}
