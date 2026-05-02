package handlers

import (
	"fmt"
	"net/http"
	"time"

	"agileos-backend/logger"
	"agileos-backend/models"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

// ============================================
// PAYROLL MANAGEMENT HANDLERS
// ============================================

// CreatePayroll creates a new payroll period
func (h *HRMHandler) CreatePayroll(c *gin.Context) {
	var payroll models.Payroll
	if err := c.ShouldBindJSON(&payroll); err != nil {
		logger.LogError("Failed to bind payroll data", err, nil)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Auto-generate payroll number: PAY-YYYY-MM
	payrollNumber := fmt.Sprintf("PAY-%d-%02d", payroll.PeriodYear, payroll.PeriodMonth)

	// Check if payroll already exists for this period
	existing, err := h.db.QuerySlice(
		"SELECT * FROM payroll WHERE payroll_number = $payroll_number",
		map[string]interface{}{"payroll_number": payrollNumber},
	)
	if err == nil && len(existing) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Payroll for this period already exists"})
		return
	}

	payroll.PayrollNumber = payrollNumber
	payroll.CreatedBy = userID.(string)
	payroll.CreatedAt = time.Now()
	payroll.UpdatedAt = time.Now()
	payroll.Status = models.PayrollStatusDraft
	payroll.TotalGrossPay = decimal.Zero
	payroll.TotalDeductions = decimal.Zero
	payroll.TotalNetPay = decimal.Zero

	query := `CREATE payroll CONTENT {
		payroll_number: $payroll_number,
		period_month: $period_month,
		period_year: $period_year,
		payment_date: $payment_date,
		status: $status,
		total_employees: $total_employees,
		total_gross_pay: $total_gross_pay,
		total_deductions: $total_deductions,
		total_net_pay: $total_net_pay,
		created_by: $created_by,
		created_at: $created_at,
		updated_at: $updated_at
	}`

	params := map[string]interface{}{
		"payroll_number":   payroll.PayrollNumber,
		"period_month":     payroll.PeriodMonth,
		"period_year":      payroll.PeriodYear,
		"payment_date":     payroll.PaymentDate,
		"status":           payroll.Status,
		"total_employees":  0,
		"total_gross_pay":  payroll.TotalGrossPay,
		"total_deductions": payroll.TotalDeductions,
		"total_net_pay":    payroll.TotalNetPay,
		"created_by":       payroll.CreatedBy,
		"created_at":       payroll.CreatedAt,
		"updated_at":       payroll.UpdatedAt,
	}

	result, err := h.db.QuerySlice(query, params)
	if err != nil {
		logger.LogError("Failed to create payroll", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create payroll"})
		return
	}

	logger.Log.Info().
		Str("payroll_number", payroll.PayrollNumber).
		Msg("Payroll created successfully")

	c.JSON(http.StatusCreated, result[0])
}

// GetPayrolls retrieves all payrolls with filters
func (h *HRMHandler) GetPayrolls(c *gin.Context) {
	year := c.Query("year")
	month := c.Query("month")
	status := c.Query("status")

	query := "SELECT * FROM payroll"
	params := make(map[string]interface{})

	var conditions []string
	if year != "" {
		conditions = append(conditions, "period_year = $year")
		params["year"] = year
	}
	if month != "" {
		conditions = append(conditions, "period_month = $month")
		params["month"] = month
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

	query += " ORDER BY period_year DESC, period_month DESC"

	payrolls, err := h.db.QuerySlice(query, params)
	if err != nil {
		logger.LogError("Failed to get payrolls", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve payrolls"})
		return
	}

	c.JSON(http.StatusOK, payrolls)
}

// GetPayroll retrieves a payroll by ID
func (h *HRMHandler) GetPayroll(c *gin.Context) {
	payrollID := c.Param("id")

	payrolls, err := h.db.QuerySlice(
		"SELECT * FROM $id",
		map[string]interface{}{"id": payrollID},
	)
	if err != nil || len(payrolls) == 0 {
		logger.LogError("Payroll not found", err, map[string]interface{}{"payroll_id": payrollID})
		c.JSON(http.StatusNotFound, gin.H{"error": "Payroll not found"})
		return
	}

	c.JSON(http.StatusOK, payrolls[0])
}

// ProcessPayroll processes payroll for all active employees
func (h *HRMHandler) ProcessPayroll(c *gin.Context) {
	payrollID := c.Param("id")

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Check if payroll is draft
	existing, err := h.db.QuerySlice(
		"SELECT * FROM $id",
		map[string]interface{}{"id": payrollID},
	)
	if err != nil || len(existing) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Payroll not found"})
		return
	}

	existingPayroll := existing[0].(map[string]interface{})
	if existingPayroll["status"] != "draft" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only draft payrolls can be processed"})
		return
	}

	// Get all active employees
	employees, err := h.db.QuerySlice(
		"SELECT * FROM employee WHERE is_active = true AND status = 'active'",
		nil,
	)
	if err != nil {
		logger.LogError("Failed to get employees", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get employees"})
		return
	}

	// Process each employee (simplified - in production, calculate allowances, deductions, taxes)
	totalGross := decimal.Zero
	totalDeductions := decimal.Zero
	totalNet := decimal.Zero

	for _, emp := range employees {
		employee := emp.(map[string]interface{})
		basicSalary := decimal.NewFromFloat(employee["basic_salary"].(float64))
		
		// Simplified calculation
		grossPay := basicSalary
		deductions := decimal.Zero
		netPay := grossPay.Sub(deductions)

		totalGross = totalGross.Add(grossPay)
		totalDeductions = totalDeductions.Add(deductions)
		totalNet = totalNet.Add(netPay)
	}

	// Update payroll
	query := `UPDATE $id SET 
		status = 'processed',
		total_employees = $total_employees,
		total_gross_pay = $total_gross_pay,
		total_deductions = $total_deductions,
		total_net_pay = $total_net_pay,
		processed_by = $processed_by,
		processed_at = $processed_at,
		updated_at = $updated_at`
	
	params := map[string]interface{}{
		"id":               payrollID,
		"total_employees":  len(employees),
		"total_gross_pay":  totalGross,
		"total_deductions": totalDeductions,
		"total_net_pay":    totalNet,
		"processed_by":     userID.(string),
		"processed_at":     time.Now(),
		"updated_at":       time.Now(),
	}

	result, err := h.db.QuerySlice(query, params)
	if err != nil || len(result) == 0 {
		logger.LogError("Failed to process payroll", err, map[string]interface{}{"payroll_id": payrollID})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process payroll"})
		return
	}

	logger.Log.Info().
		Str("payroll_id", payrollID).
		Int("total_employees", len(employees)).
		Msg("Payroll processed successfully")

	c.JSON(http.StatusOK, result[0])
}

// ApprovePayroll approves a processed payroll
func (h *HRMHandler) ApprovePayroll(c *gin.Context) {
	payrollID := c.Param("id")

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Check if payroll is processed
	existing, err := h.db.QuerySlice(
		"SELECT * FROM $id",
		map[string]interface{}{"id": payrollID},
	)
	if err != nil || len(existing) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Payroll not found"})
		return
	}

	existingPayroll := existing[0].(map[string]interface{})
	if existingPayroll["status"] != "processed" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only processed payrolls can be approved"})
		return
	}

	query := `UPDATE $id SET 
		status = 'approved',
		approved_by = $approved_by,
		approved_at = $approved_at,
		updated_at = $updated_at`
	
	params := map[string]interface{}{
		"id":          payrollID,
		"approved_by": userID.(string),
		"approved_at": time.Now(),
		"updated_at":  time.Now(),
	}

	result, err := h.db.QuerySlice(query, params)
	if err != nil || len(result) == 0 {
		logger.LogError("Failed to approve payroll", err, map[string]interface{}{"payroll_id": payrollID})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to approve payroll"})
		return
	}

	logger.Log.Info().
		Str("payroll_id", payrollID).
		Msg("Payroll approved successfully")

	c.JSON(http.StatusOK, result[0])
}

// PayPayroll marks payroll as paid
func (h *HRMHandler) PayPayroll(c *gin.Context) {
	payrollID := c.Param("id")

	// Check if payroll is approved
	existing, err := h.db.QuerySlice(
		"SELECT * FROM $id",
		map[string]interface{}{"id": payrollID},
	)
	if err != nil || len(existing) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Payroll not found"})
		return
	}

	existingPayroll := existing[0].(map[string]interface{})
	if existingPayroll["status"] != "approved" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only approved payrolls can be paid"})
		return
	}

	query := `UPDATE $id SET status = 'paid', updated_at = $updated_at`
	params := map[string]interface{}{
		"id":         payrollID,
		"updated_at": time.Now(),
	}

	result, err := h.db.QuerySlice(query, params)
	if err != nil || len(result) == 0 {
		logger.LogError("Failed to pay payroll", err, map[string]interface{}{"payroll_id": payrollID})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to pay payroll"})
		return
	}

	logger.Log.Info().
		Str("payroll_id", payrollID).
		Msg("Payroll paid successfully")

	c.JSON(http.StatusOK, result[0])
}

// GetPayrollDetails retrieves payroll details for employees
func (h *HRMHandler) GetPayrollDetails(c *gin.Context) {
	payrollID := c.Query("payroll_id")
	employeeID := c.Query("employee_id")

	query := "SELECT * FROM payroll_detail"
	params := make(map[string]interface{})

	var conditions []string
	if payrollID != "" {
		conditions = append(conditions, "payroll_id = $payroll_id")
		params["payroll_id"] = payrollID
	}
	if employeeID != "" {
		conditions = append(conditions, "employee_id = $employee_id")
		params["employee_id"] = employeeID
	}

	if len(conditions) > 0 {
		query += " WHERE " + conditions[0]
		for i := 1; i < len(conditions); i++ {
			query += " AND " + conditions[i]
		}
	}

	query += " ORDER BY employee_code ASC"

	details, err := h.db.QuerySlice(query, params)
	if err != nil {
		logger.LogError("Failed to get payroll details", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve payroll details"})
		return
	}

	c.JSON(http.StatusOK, details)
}

// GetEmployeePayslip retrieves payslip for an employee
func (h *HRMHandler) GetEmployeePayslip(c *gin.Context) {
	payrollID := c.Param("payroll_id")
	employeeID := c.Param("employee_id")

	details, err := h.db.QuerySlice(
		"SELECT * FROM payroll_detail WHERE payroll_id = $payroll_id AND employee_id = $employee_id",
		map[string]interface{}{
			"payroll_id":  payrollID,
			"employee_id": employeeID,
		},
	)
	if err != nil || len(details) == 0 {
		logger.LogError("Payslip not found", err, map[string]interface{}{
			"payroll_id":  payrollID,
			"employee_id": employeeID,
		})
		c.JSON(http.StatusNotFound, gin.H{"error": "Payslip not found"})
		return
	}

	c.JSON(http.StatusOK, details[0])
}
