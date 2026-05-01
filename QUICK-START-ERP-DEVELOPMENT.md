# 🚀 QUICK START - ERP Development Guide

**For developers continuing this ERP implementation**

---

## 📁 WHAT'S ALREADY DONE

### ✅ Complete (Ready to Use)
1. **All Data Models** - `backend-go/models/`
   - `hrm.go` - 10 models
   - `inventory.go` - 12 models
   - `crm.go` - 10 models
   - `manufacturing.go` - 12 models
   - `accounting.go` - Extended with 4 report models

2. **Sample Handler** - `backend-go/handlers/accounting_ar.go`
   - 12 endpoints (Customer + Sales Invoice)
   - Use as template for other handlers

3. **Documentation** - Root directory
   - `ERP-IMPLEMENTATION-FINAL-SUMMARY.md` - Read this first!
   - `ERP-MODELS-COMPLETE-SUMMARY.md` - All models explained
   - `BACKEND-HANDLERS-IMPLEMENTATION-STATUS.md` - What's left to do

---

## 🎯 YOUR TASK: Complete Remaining Handlers

### Step 1: Choose a Module
Pick one from:
- Accounting (2 files left)
- HRM (4 files)
- Inventory (4 files)
- CRM (4 files)
- Manufacturing (4 files)

### Step 2: Copy Template

```go
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

// Create handler struct (or use existing AccountingHandler)
type HRMHandler struct {
	db *database.SurrealDB
}

func NewHRMHandler(db *database.SurrealDB) *HRMHandler {
	return &HRMHandler{db: db}
}

// Example: Create Employee
func (h *HRMHandler) CreateEmployee(c *gin.Context) {
	var employee models.Employee
	if err := c.ShouldBindJSON(&employee); err != nil {
		logger.LogError("Failed to bind employee data", err, nil)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Auto-generate code
	employees, err := h.db.Query(
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

	// Set metadata
	employee.EmployeeCode = employeeCode
	employee.CreatedBy = userID.(string)
	employee.CreatedAt = time.Now()
	employee.UpdatedAt = time.Now()
	employee.IsActive = true

	// Create in database
	query := `CREATE employee CONTENT {
		employee_code: $employee_code,
		first_name: $first_name,
		last_name: $last_name,
		email: $email,
		phone: $phone,
		department: $department,
		position: $position,
		join_date: $join_date,
		basic_salary: $basic_salary,
		is_active: $is_active,
		created_by: $created_by,
		created_at: $created_at,
		updated_at: $updated_at
	}`

	params := map[string]interface{}{
		"employee_code": employee.EmployeeCode,
		"first_name":    employee.FirstName,
		"last_name":     employee.LastName,
		"email":         employee.Email,
		"phone":         employee.Phone,
		"department":    employee.Department,
		"position":      employee.Position,
		"join_date":     employee.JoinDate,
		"basic_salary":  employee.BasicSalary,
		"is_active":     employee.IsActive,
		"created_by":    employee.CreatedBy,
		"created_at":    employee.CreatedAt,
		"updated_at":    employee.UpdatedAt,
	}

	result, err := h.db.Query(query, params)
	if err != nil {
		logger.LogError("Failed to create employee", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create employee"})
		return
	}

	logger.Log.Info().
		Str("employee_code", employee.EmployeeCode).
		Str("employee_name", employee.FirstName+" "+employee.LastName).
		Msg("Employee created successfully")

	c.JSON(http.StatusCreated, result[0])
}

// Add other CRUD methods: Get, GetAll, Update, Delete
```

### Step 3: Register Routes in main.go

```go
// In backend-go/main.go

// Create handler
hrmHandler := handlers.NewHRMHandler(db)

// Register routes
hrm := api.Group("/hrm")
{
    hrm.Use(middleware.AuthorizeRole("admin", "hr"))
    {
        // Employees
        hrm.POST("/employees", hrmHandler.CreateEmployee)
        hrm.GET("/employees", hrmHandler.GetEmployees)
        hrm.GET("/employees/:id", hrmHandler.GetEmployee)
        hrm.PUT("/employees/:id", hrmHandler.UpdateEmployee)
        hrm.DELETE("/employees/:id", hrmHandler.DeleteEmployee)
        
        // Add other endpoints...
    }
}
```

### Step 4: Test with Thunder Client/Postman

```json
POST http://localhost:8080/api/v1/hrm/employees
Authorization: Bearer YOUR_JWT_TOKEN
Content-Type: application/json

{
  "first_name": "John",
  "last_name": "Doe",
  "email": "john.doe@company.com",
  "phone": "+6281234567890",
  "department": "IT",
  "position": "Software Engineer",
  "employment_type": "full_time",
  "join_date": "2026-05-01T00:00:00Z",
  "basic_salary": 15000000,
  "currency": "IDR"
}
```

---

## 📋 CHECKLIST FOR EACH HANDLER FILE

- [ ] Import required packages
- [ ] Create handler struct (or use existing)
- [ ] Implement Create method
  - [ ] Bind JSON
  - [ ] Get user from context
  - [ ] Auto-generate code
  - [ ] Set metadata
  - [ ] Create in database
  - [ ] Log action
  - [ ] Return response
- [ ] Implement GetAll method (with filters)
- [ ] Implement Get by ID method
- [ ] Implement Update method
  - [ ] Check if exists
  - [ ] Validate status if needed
  - [ ] Update in database
- [ ] Implement Delete method (soft delete)
- [ ] Implement additional actions (approve, cancel, etc.)
- [ ] Register routes in main.go
- [ ] Test all endpoints

---

## 🎨 FRONTEND DEVELOPMENT

### Step 1: Create List Page

```typescript
// Example: app/[locale]/hrm/employees/page.tsx

'use client';

import { useState, useEffect } from 'react';
import { authenticatedFetch } from '@/lib/auth';
import { toast } from 'sonner';

export default function EmployeesPage() {
  const [employees, setEmployees] = useState([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchEmployees();
  }, []);

  const fetchEmployees = async () => {
    try {
      setLoading(true);
      const response = await authenticatedFetch(
        `${process.env.NEXT_PUBLIC_API_URL}/api/v1/hrm/employees`
      );
      
      if (!response.ok) throw new Error('Failed to fetch employees');
      
      const data = await response.json();
      setEmployees(data);
    } catch (error) {
      console.error('Failed to fetch employees:', error);
      toast.error('Failed to load employees');
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return <div>Loading...</div>;
  }

  return (
    <div className="p-8">
      <h1 className="text-3xl font-bold mb-6">Employees</h1>
      
      <table className="w-full">
        <thead>
          <tr>
            <th>Code</th>
            <th>Name</th>
            <th>Department</th>
            <th>Position</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          {employees.map((emp: any) => (
            <tr key={emp.id}>
              <td>{emp.employee_code}</td>
              <td>{emp.full_name}</td>
              <td>{emp.department}</td>
              <td>{emp.position}</td>
              <td>
                <button>View</button>
                <button>Edit</button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
```

### Step 2: Create Form Page

Copy pattern from `purchase-invoices/new/page.tsx` and modify for your entity.

### Step 3: Create Detail Page

Copy pattern from `purchase-invoices/[id]/page.tsx` and modify.

---

## 🔍 REFERENCE FILES

### Backend Patterns
- **Handler Template**: `backend-go/handlers/accounting_ar.go`
- **Models**: `backend-go/models/*.go`
- **Main Routes**: `backend-go/main.go`

### Frontend Patterns
- **List Page**: `frontend-next/app/[locale]/accounting/purchase-invoices/page.tsx`
- **Create Form**: `frontend-next/app/[locale]/accounting/purchase-invoices/new/page.tsx`
- **Detail View**: `frontend-next/app/[locale]/accounting/purchase-invoices/[id]/page.tsx`
- **Edit Form**: `frontend-next/app/[locale]/accounting/purchase-invoices/[id]/edit/page.tsx`

---

## 🐛 COMMON ISSUES & SOLUTIONS

### Issue: "User not authenticated"
**Solution**: Make sure JWT token is included in request headers

### Issue: "Failed to generate code"
**Solution**: Check database query syntax for SurrealDB

### Issue: "Cannot bind JSON"
**Solution**: Verify JSON structure matches model fields

### Issue: "Route not found"
**Solution**: Check if route is registered in main.go

---

## 📞 NEED HELP?

1. **Read Documentation**:
   - `ERP-IMPLEMENTATION-FINAL-SUMMARY.md`
   - `ERP-MODELS-COMPLETE-SUMMARY.md`

2. **Check Examples**:
   - `accounting_ar.go` for backend
   - `purchase-invoices/` for frontend

3. **Follow Patterns**:
   - All handlers follow same structure
   - All pages follow same layout

---

## 🎯 PRIORITY ORDER

### High Priority (Complete First)
1. ✅ Account Receivable (DONE)
2. ⏳ Financial Reports
3. ⏳ Budget Management
4. ⏳ HRM - Employee & Payroll

### Medium Priority
5. ⏳ Inventory - Product & Stock
6. ⏳ CRM - Lead & Opportunity

### Lower Priority
7. ⏳ Manufacturing modules

---

## ✅ DEFINITION OF DONE

For each module:
- [ ] All handler methods implemented
- [ ] Routes registered in main.go
- [ ] All endpoints tested and working
- [ ] Frontend list page created
- [ ] Frontend create form created
- [ ] Frontend detail view created
- [ ] Frontend edit form created (if applicable)
- [ ] CRUD operations work end-to-end
- [ ] Error handling works properly
- [ ] Loading states implemented
- [ ] Toast notifications working

---

**Good luck with the development!** 🚀

The foundation is solid. Just follow the patterns and you'll complete this ERP system successfully!
