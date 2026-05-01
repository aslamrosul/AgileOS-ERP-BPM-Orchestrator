package handlers

import (
	"fmt"
	"net/http"
	"time"

	"agileos-backend/database"
	"agileos-backend/logger"
	"agileos-backend/models"

	"github.com/gin-gonic/gin"
)

// HRMHandler handles HRM operations
type HRMHandler struct {
	db *database.SurrealDB
}

// NewHRMHandler creates a new HRM handler
func NewHRMHandler(db *database.SurrealDB) *HRMHandler {
	return &HRMHandler{db: db}
}

// ============================================
// EMPLOYEE MANAGEMENT HANDLERS
// ============================================

// CreateEmployee creates a new employee
func (h *HRMHandler) CreateEmployee(c *gin.Context) {
	var employee models.Employee
	if err := c.ShouldBindJSON(&employee); err != nil {
		logger.LogError("Failed to bind employee data", err, nil)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Auto-generate employee code
	employees, err := h.db.QuerySlice(
		"SELECT employee_code FROM employee ORDER BY employee_code DESC LIMIT 1",
		nil,
	)
	if err != nil {
		logger.LogError("Failed to get last employee code", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate employee code"})
		return
	}

	employeeCode := "EMP-0001"
	if len(employees) > 0 {
		lastCode := employees[0].(map[string]interface{})["employee_code"].(string)
		var lastNum int
		fmt.Sscanf(lastCode, "EMP-%d", &lastNum)
		employeeCode = fmt.Sprintf("EMP-%04d", lastNum+1)
	}

	employee.EmployeeCode = employeeCode
	employee.FullName = employee.FirstName + " " + employee.LastName
	employee.CreatedBy = userID.(string)
	employee.CreatedAt = time.Now()
	employee.UpdatedAt = time.Now()
	employee.Status = models.EmployeeStatusActive
	employee.IsActive = true

	query := `CREATE employee CONTENT {
		employee_code: $employee_code,
		first_name: $first_name,
		last_name: $last_name,
		full_name: $full_name,
		email: $email,
		phone: $phone,
		date_of_birth: $date_of_birth,
		gender: $gender,
		address: $address,
		city: $city,
		state: $state,
		country: $country,
		postal_code: $postal_code,
		department: $department,
		position: $position,
		employment_type: $employment_type,
		join_date: $join_date,
		manager_id: $manager_id,
		manager_name: $manager_name,
		basic_salary: $basic_salary,
		currency: $currency,
		payment_method: $payment_method,
		bank_name: $bank_name,
		bank_account: $bank_account,
		tax_id: $tax_id,
		bpjs_kesehatan: $bpjs_kesehatan,
		bpjs_ketenagakerjaan: $bpjs_ketenagakerjaan,
		status: $status,
		is_active: $is_active,
		created_by: $created_by,
		created_at: $created_at,
		updated_at: $updated_at
	}`

	params := map[string]interface{}{
		"employee_code":         employee.EmployeeCode,
		"first_name":            employee.FirstName,
		"last_name":             employee.LastName,
		"full_name":             employee.FullName,
		"email":                 employee.Email,
		"phone":                 employee.Phone,
		"date_of_birth":         employee.DateOfBirth,
		"gender":                employee.Gender,
		"address":               employee.Address,
		"city":                  employee.City,
		"state":                 employee.State,
		"country":               employee.Country,
		"postal_code":           employee.PostalCode,
		"department":            employee.Department,
		"position":              employee.Position,
		"employment_type":       employee.EmploymentType,
		"join_date":             employee.JoinDate,
		"manager_id":            employee.ManagerID,
		"manager_name":          employee.ManagerName,
		"basic_salary":          employee.BasicSalary,
		"currency":              employee.Currency,
		"payment_method":        employee.PaymentMethod,
		"bank_name":             employee.BankName,
		"bank_account":          employee.BankAccount,
		"tax_id":                employee.TaxID,
		"bpjs_kesehatan":        employee.BPJSKesehatan,
		"bpjs_ketenagakerjaan":  employee.BPJSKetenagakerjaan,
		"status":                employee.Status,
		"is_active":             employee.IsActive,
		"created_by":            employee.CreatedBy,
		"created_at":            employee.CreatedAt,
		"updated_at":            employee.UpdatedAt,
	}

	result, err := h.db.QuerySlice(query, params)
	if err != nil {
		logger.LogError("Failed to create employee", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create employee"})
		return
	}

	logger.Log.Info().
		Str("employee_code", employee.EmployeeCode).
		Str("full_name", employee.FullName).
		Msg("Employee created successfully")

	c.JSON(http.StatusCreated, result[0])
}

// GetEmployees retrieves all employees with filters
func (h *HRMHandler) GetEmployees(c *gin.Context) {
	department := c.Query("department")
	status := c.Query("status")
	employmentType := c.Query("employment_type")

	query := "SELECT * FROM employee"
	params := make(map[string]interface{})

	var conditions []string
	if department != "" {
		conditions = append(conditions, "department = $department")
		params["department"] = department
	}
	if status != "" {
		conditions = append(conditions, "status = $status")
		params["status"] = status
	}
	if employmentType != "" {
		conditions = append(conditions, "employment_type = $employment_type")
		params["employment_type"] = employmentType
	}

	if len(conditions) > 0 {
		query += " WHERE " + conditions[0]
		for i := 1; i < len(conditions); i++ {
			query += " AND " + conditions[i]
		}
	}

	query += " ORDER BY employee_code ASC"

	employees, err := h.db.QuerySlice(query, params)
	if err != nil {
		logger.LogError("Failed to get employees", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve employees"})
		return
	}

	c.JSON(http.StatusOK, employees)
}

// GetEmployee retrieves an employee by ID
func (h *HRMHandler) GetEmployee(c *gin.Context) {
	employeeID := c.Param("id")

	employees, err := h.db.QuerySlice(
		"SELECT * FROM $id",
		map[string]interface{}{"id": employeeID},
	)
	if err != nil || len(employees) == 0 {
		logger.LogError("Employee not found", err, map[string]interface{}{"employee_id": employeeID})
		c.JSON(http.StatusNotFound, gin.H{"error": "Employee not found"})
		return
	}

	c.JSON(http.StatusOK, employees[0])
}

// UpdateEmployee updates an existing employee
func (h *HRMHandler) UpdateEmployee(c *gin.Context) {
	employeeID := c.Param("id")

	var employee models.Employee
	if err := c.ShouldBindJSON(&employee); err != nil {
		logger.LogError("Failed to bind employee data", err, nil)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	employee.FullName = employee.FirstName + " " + employee.LastName
	employee.UpdatedAt = time.Now()

	query := `UPDATE $id SET
		first_name = $first_name,
		last_name = $last_name,
		full_name = $full_name,
		email = $email,
		phone = $phone,
		date_of_birth = $date_of_birth,
		gender = $gender,
		address = $address,
		city = $city,
		state = $state,
		country = $country,
		postal_code = $postal_code,
		department = $department,
		position = $position,
		employment_type = $employment_type,
		manager_id = $manager_id,
		manager_name = $manager_name,
		basic_salary = $basic_salary,
		payment_method = $payment_method,
		bank_name = $bank_name,
		bank_account = $bank_account,
		tax_id = $tax_id,
		bpjs_kesehatan = $bpjs_kesehatan,
		bpjs_ketenagakerjaan = $bpjs_ketenagakerjaan,
		status = $status,
		is_active = $is_active,
		updated_at = $updated_at
	`

	params := map[string]interface{}{
		"id":                    employeeID,
		"first_name":            employee.FirstName,
		"last_name":             employee.LastName,
		"full_name":             employee.FullName,
		"email":                 employee.Email,
		"phone":                 employee.Phone,
		"date_of_birth":         employee.DateOfBirth,
		"gender":                employee.Gender,
		"address":               employee.Address,
		"city":                  employee.City,
		"state":                 employee.State,
		"country":               employee.Country,
		"postal_code":           employee.PostalCode,
		"department":            employee.Department,
		"position":              employee.Position,
		"employment_type":       employee.EmploymentType,
		"manager_id":            employee.ManagerID,
		"manager_name":          employee.ManagerName,
		"basic_salary":          employee.BasicSalary,
		"payment_method":        employee.PaymentMethod,
		"bank_name":             employee.BankName,
		"bank_account":          employee.BankAccount,
		"tax_id":                employee.TaxID,
		"bpjs_kesehatan":        employee.BPJSKesehatan,
		"bpjs_ketenagakerjaan":  employee.BPJSKetenagakerjaan,
		"status":                employee.Status,
		"is_active":             employee.IsActive,
		"updated_at":            employee.UpdatedAt,
	}

	result, err := h.db.QuerySlice(query, params)
	if err != nil || len(result) == 0 {
		logger.LogError("Failed to update employee", err, map[string]interface{}{"employee_id": employeeID})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update employee"})
		return
	}

	logger.Log.Info().
		Str("employee_id", employeeID).
		Str("full_name", employee.FullName).
		Msg("Employee updated successfully")

	c.JSON(http.StatusOK, result[0])
}

// DeleteEmployee soft deletes an employee
func (h *HRMHandler) DeleteEmployee(c *gin.Context) {
	employeeID := c.Param("id")

	query := `UPDATE $id SET 
		status = 'inactive',
		is_active = false, 
		updated_at = $updated_at`
	
	params := map[string]interface{}{
		"id":         employeeID,
		"updated_at": time.Now(),
	}

	result, err := h.db.QuerySlice(query, params)
	if err != nil || len(result) == 0 {
		logger.LogError("Failed to delete employee", err, map[string]interface{}{"employee_id": employeeID})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete employee"})
		return
	}

	logger.Log.Info().
		Str("employee_id", employeeID).
		Msg("Employee deleted successfully")

	c.JSON(http.StatusOK, gin.H{"message": "Employee deleted successfully"})
}
