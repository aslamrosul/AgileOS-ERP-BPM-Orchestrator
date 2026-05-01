package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// ============================================
// EMPLOYEE MANAGEMENT
// ============================================

// Employee represents an employee in the organization
type Employee struct {
	ID              string          `json:"id,omitempty"`
	EmployeeCode    string          `json:"employee_code"`    // Auto-generated: EMP-0001
	FirstName       string          `json:"first_name"`
	LastName        string          `json:"last_name"`
	FullName        string          `json:"full_name"`
	Email           string          `json:"email"`
	Phone           string          `json:"phone"`
	DateOfBirth     time.Time       `json:"date_of_birth"`
	Gender          Gender          `json:"gender"`
	Address         string          `json:"address"`
	City            string          `json:"city"`
	State           string          `json:"state"`
	Country         string          `json:"country"`
	PostalCode      string          `json:"postal_code"`
	
	// Employment Information
	Department      string          `json:"department"`
	Position        string          `json:"position"`
	EmploymentType  EmploymentType  `json:"employment_type"`  // full_time, part_time, contract, intern
	JoinDate        time.Time       `json:"join_date"`
	EndDate         *time.Time      `json:"end_date,omitempty"`
	ManagerID       string          `json:"manager_id,omitempty"`
	ManagerName     string          `json:"manager_name,omitempty"`
	
	// Salary Information
	BasicSalary     decimal.Decimal `json:"basic_salary"`
	Currency        string          `json:"currency"`
	PaymentMethod   PaymentMethod   `json:"payment_method"`   // bank_transfer, cash, check
	BankName        string          `json:"bank_name,omitempty"`
	BankAccount     string          `json:"bank_account,omitempty"`
	
	// Tax & Insurance
	TaxID           string          `json:"tax_id"`           // NPWP
	BPJSKesehatan   string          `json:"bpjs_kesehatan,omitempty"`
	BPJSKetenagakerjaan string      `json:"bpjs_ketenagakerjaan,omitempty"`
	
	// Status
	Status          EmployeeStatus  `json:"status"`           // active, inactive, terminated, resigned
	IsActive        bool            `json:"is_active"`
	
	// Metadata
	CreatedBy       string          `json:"created_by"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

// Gender defines employee gender
type Gender string

const (
	GenderMale   Gender = "male"
	GenderFemale Gender = "female"
	GenderOther  Gender = "other"
)

// EmploymentType defines the type of employment
type EmploymentType string

const (
	EmploymentTypeFullTime EmploymentType = "full_time"
	EmploymentTypePartTime EmploymentType = "part_time"
	EmploymentTypeContract EmploymentType = "contract"
	EmploymentTypeIntern   EmploymentType = "intern"
)

// EmployeeStatus defines employee status
type EmployeeStatus string

const (
	EmployeeStatusActive     EmployeeStatus = "active"
	EmployeeStatusInactive   EmployeeStatus = "inactive"
	EmployeeStatusTerminated EmployeeStatus = "terminated"
	EmployeeStatusResigned   EmployeeStatus = "resigned"
)

// ============================================
// PAYROLL
// ============================================

// Payroll represents a payroll period
type Payroll struct {
	ID              string          `json:"id,omitempty"`
	PayrollNumber   string          `json:"payroll_number"`   // Auto-generated: PAY-2026-01
	PeriodMonth     int             `json:"period_month"`     // 1-12
	PeriodYear      int             `json:"period_year"`
	PaymentDate     time.Time       `json:"payment_date"`
	Status          PayrollStatus   `json:"status"`           // draft, processed, approved, paid
	TotalEmployees  int             `json:"total_employees"`
	TotalGrossPay   decimal.Decimal `json:"total_gross_pay"`
	TotalDeductions decimal.Decimal `json:"total_deductions"`
	TotalNetPay     decimal.Decimal `json:"total_net_pay"`
	ProcessedBy     string          `json:"processed_by,omitempty"`
	ProcessedAt     *time.Time      `json:"processed_at,omitempty"`
	ApprovedBy      string          `json:"approved_by,omitempty"`
	ApprovedAt      *time.Time      `json:"approved_at,omitempty"`
	CreatedBy       string          `json:"created_by"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

// PayrollStatus defines payroll status
type PayrollStatus string

const (
	PayrollStatusDraft     PayrollStatus = "draft"
	PayrollStatusProcessed PayrollStatus = "processed"
	PayrollStatusApproved  PayrollStatus = "approved"
	PayrollStatusPaid      PayrollStatus = "paid"
)

// PayrollDetail represents individual employee payroll
type PayrollDetail struct {
	ID              string          `json:"id,omitempty"`
	PayrollID       string          `json:"payroll_id"`
	EmployeeID      string          `json:"employee_id"`
	EmployeeCode    string          `json:"employee_code"`
	EmployeeName    string          `json:"employee_name"`
	
	// Earnings
	BasicSalary     decimal.Decimal `json:"basic_salary"`
	Allowances      []PayrollComponent `json:"allowances"`
	TotalAllowances decimal.Decimal `json:"total_allowances"`
	Overtime        decimal.Decimal `json:"overtime"`
	Bonus           decimal.Decimal `json:"bonus"`
	GrossPay        decimal.Decimal `json:"gross_pay"`
	
	// Deductions
	Deductions      []PayrollComponent `json:"deductions"`
	TotalDeductions decimal.Decimal `json:"total_deductions"`
	TaxPPh21        decimal.Decimal `json:"tax_pph21"`
	BPJSKesehatan   decimal.Decimal `json:"bpjs_kesehatan"`
	BPJSKetenagakerjaan decimal.Decimal `json:"bpjs_ketenagakerjaan"`
	
	// Net Pay
	NetPay          decimal.Decimal `json:"net_pay"`
	
	// Payment Info
	PaymentMethod   PaymentMethod   `json:"payment_method"`
	BankName        string          `json:"bank_name,omitempty"`
	BankAccount     string          `json:"bank_account,omitempty"`
	PaymentStatus   string          `json:"payment_status"`   // pending, paid
	PaidAt          *time.Time      `json:"paid_at,omitempty"`
	
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

// PayrollComponent represents salary component (allowance or deduction)
type PayrollComponent struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Amount      decimal.Decimal `json:"amount"`
	IsTaxable   bool            `json:"is_taxable"`
}

// ============================================
// ATTENDANCE
// ============================================

// Attendance represents employee attendance record
type Attendance struct {
	ID              string          `json:"id,omitempty"`
	EmployeeID      string          `json:"employee_id"`
	EmployeeCode    string          `json:"employee_code"`
	EmployeeName    string          `json:"employee_name"`
	Date            time.Time       `json:"date"`
	CheckInTime     *time.Time      `json:"check_in_time,omitempty"`
	CheckOutTime    *time.Time      `json:"check_out_time,omitempty"`
	CheckInLocation string          `json:"check_in_location,omitempty"`
	CheckOutLocation string         `json:"check_out_location,omitempty"`
	WorkHours       float64         `json:"work_hours"`
	OvertimeHours   float64         `json:"overtime_hours"`
	Status          AttendanceStatus `json:"status"`          // present, absent, late, half_day, leave
	IsLate          bool            `json:"is_late"`
	LateMinutes     int             `json:"late_minutes"`
	Notes           string          `json:"notes,omitempty"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

// AttendanceStatus defines attendance status
type AttendanceStatus string

const (
	AttendanceStatusPresent AttendanceStatus = "present"
	AttendanceStatusAbsent  AttendanceStatus = "absent"
	AttendanceStatusLate    AttendanceStatus = "late"
	AttendanceStatusHalfDay AttendanceStatus = "half_day"
	AttendanceStatusLeave   AttendanceStatus = "leave"
)

// ============================================
// LEAVE MANAGEMENT
// ============================================

// LeaveType represents a type of leave
type LeaveType struct {
	ID              string          `json:"id,omitempty"`
	LeaveCode       string          `json:"leave_code"`       // AUTO-generated: LT-001
	LeaveName       string          `json:"leave_name"`       // Annual Leave, Sick Leave, etc.
	Description     string          `json:"description,omitempty"`
	MaxDaysPerYear  int             `json:"max_days_per_year"`
	IsPaid          bool            `json:"is_paid"`
	RequiresApproval bool           `json:"requires_approval"`
	RequiresDocument bool           `json:"requires_document"`
	IsActive        bool            `json:"is_active"`
	CreatedBy       string          `json:"created_by"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

// LeaveRequest represents an employee leave request
type LeaveRequest struct {
	ID              string          `json:"id,omitempty"`
	RequestNumber   string          `json:"request_number"`   // Auto-generated: LR-2026-0001
	EmployeeID      string          `json:"employee_id"`
	EmployeeCode    string          `json:"employee_code"`
	EmployeeName    string          `json:"employee_name"`
	LeaveTypeID     string          `json:"leave_type_id"`
	LeaveTypeName   string          `json:"leave_type_name"`
	StartDate       time.Time       `json:"start_date"`
	EndDate         time.Time       `json:"end_date"`
	TotalDays       int             `json:"total_days"`
	Reason          string          `json:"reason"`
	DocumentURL     string          `json:"document_url,omitempty"`
	Status          LeaveStatus     `json:"status"`           // pending, approved, rejected, cancelled
	ApproverID      string          `json:"approver_id,omitempty"`
	ApproverName    string          `json:"approver_name,omitempty"`
	ApprovalNotes   string          `json:"approval_notes,omitempty"`
	ApprovedAt      *time.Time      `json:"approved_at,omitempty"`
	CreatedBy       string          `json:"created_by"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

// LeaveStatus defines leave request status
type LeaveStatus string

const (
	LeaveStatusPending   LeaveStatus = "pending"
	LeaveStatusApproved  LeaveStatus = "approved"
	LeaveStatusRejected  LeaveStatus = "rejected"
	LeaveStatusCancelled LeaveStatus = "cancelled"
)

// LeaveBalance represents employee leave balance
type LeaveBalance struct {
	ID              string          `json:"id,omitempty"`
	EmployeeID      string          `json:"employee_id"`
	LeaveTypeID     string          `json:"leave_type_id"`
	LeaveTypeName   string          `json:"leave_type_name"`
	Year            int             `json:"year"`
	TotalDays       int             `json:"total_days"`
	UsedDays        int             `json:"used_days"`
	RemainingDays   int             `json:"remaining_days"`
	UpdatedAt       time.Time       `json:"updated_at"`
}
