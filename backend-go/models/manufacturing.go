package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// ============================================
// BILL OF MATERIALS (BOM)
// ============================================

// BOM represents a Bill of Materials
type BOM struct {
	ID              string          `json:"id,omitempty"`
	BOMCode         string          `json:"bom_code"`         // Auto-generated: BOM-0001
	BOMName         string          `json:"bom_name"`
	ProductID       string          `json:"product_id"`       // Finished product
	ProductCode     string          `json:"product_code"`
	ProductName     string          `json:"product_name"`
	BOMType         BOMType         `json:"bom_type"`         // manufacturing, kit, phantom
	Quantity        decimal.Decimal `json:"quantity"`         // Quantity produced
	UnitOfMeasure   string          `json:"unit_of_measure"`
	Version         int             `json:"version"`          // BOM version
	IsActive        bool            `json:"is_active"`
	IsDefault       bool            `json:"is_default"`       // Default BOM for product
	Components      []BOMComponent  `json:"components"`       // BOM lines
	Operations      []BOMOperation  `json:"operations,omitempty"` // Routing operations
	TotalCost       decimal.Decimal `json:"total_cost"`       // Total material cost
	LaborCost       decimal.Decimal `json:"labor_cost"`
	OverheadCost    decimal.Decimal `json:"overhead_cost"`
	TotalProductionCost decimal.Decimal `json:"total_production_cost"`
	Notes           string          `json:"notes,omitempty"`
	CreatedBy       string          `json:"created_by"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

// BOMType defines BOM type
type BOMType string

const (
	BOMTypeManufacturing BOMType = "manufacturing"
	BOMTypeKit           BOMType = "kit"
	BOMTypePhantom       BOMType = "phantom"
)

// BOMComponent represents a component in BOM
type BOMComponent struct {
	ID              string          `json:"id,omitempty"`
	Sequence        int             `json:"sequence"`
	ProductID       string          `json:"product_id"`       // Component product
	ProductCode     string          `json:"product_code"`
	ProductName     string          `json:"product_name"`
	Quantity        decimal.Decimal `json:"quantity"`         // Quantity per unit
	UnitOfMeasure   string          `json:"unit_of_measure"`
	UnitCost        decimal.Decimal `json:"unit_cost"`
	TotalCost       decimal.Decimal `json:"total_cost"`
	ScrapPercentage decimal.Decimal `json:"scrap_percentage"` // Expected scrap %
	IsOptional      bool            `json:"is_optional"`
	Notes           string          `json:"notes,omitempty"`
}

// BOMOperation represents a manufacturing operation/routing
type BOMOperation struct {
	ID              string          `json:"id,omitempty"`
	Sequence        int             `json:"sequence"`
	OperationName   string          `json:"operation_name"`
	WorkCenterID    string          `json:"work_center_id,omitempty"`
	WorkCenterName  string          `json:"work_center_name,omitempty"`
	Duration        decimal.Decimal `json:"duration"`         // in minutes
	LaborCost       decimal.Decimal `json:"labor_cost"`
	OverheadCost    decimal.Decimal `json:"overhead_cost"`
	TotalCost       decimal.Decimal `json:"total_cost"`
	Description     string          `json:"description,omitempty"`
}

// ============================================
// PRODUCTION PLANNING
// ============================================

// ProductionPlan represents a production plan
type ProductionPlan struct {
	ID              string          `json:"id,omitempty"`
	PlanNumber      string          `json:"plan_number"`      // Auto-generated: PP-2026-0001
	PlanName        string          `json:"plan_name"`
	PlanType        PlanType        `json:"plan_type"`        // monthly, weekly, daily
	StartDate       time.Time       `json:"start_date"`
	EndDate         time.Time       `json:"end_date"`
	Status          PlanStatus      `json:"status"`           // draft, approved, in_progress, completed
	Lines           []ProductionPlanLine `json:"lines"`
	TotalQuantity   decimal.Decimal `json:"total_quantity"`
	TotalCost       decimal.Decimal `json:"total_cost"`
	Notes           string          `json:"notes,omitempty"`
	ApprovedBy      string          `json:"approved_by,omitempty"`
	ApprovedAt      *time.Time      `json:"approved_at,omitempty"`
	CreatedBy       string          `json:"created_by"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

// PlanType defines production plan type
type PlanType string

const (
	PlanTypeMonthly PlanType = "monthly"
	PlanTypeWeekly  PlanType = "weekly"
	PlanTypeDaily   PlanType = "daily"
)

// PlanStatus defines production plan status
type PlanStatus string

const (
	PlanStatusDraft      PlanStatus = "draft"
	PlanStatusApproved   PlanStatus = "approved"
	PlanStatusInProgress PlanStatus = "in_progress"
	PlanStatusCompleted  PlanStatus = "completed"
)

// ProductionPlanLine represents a line in production plan
type ProductionPlanLine struct {
	ID              string          `json:"id,omitempty"`
	ProductID       string          `json:"product_id"`
	ProductCode     string          `json:"product_code"`
	ProductName     string          `json:"product_name"`
	BOMID           string          `json:"bom_id"`
	PlannedQuantity decimal.Decimal `json:"planned_quantity"`
	ProducedQuantity decimal.Decimal `json:"produced_quantity"`
	UnitOfMeasure   string          `json:"unit_of_measure"`
	StartDate       time.Time       `json:"start_date"`
	EndDate         time.Time       `json:"end_date"`
	Priority        int             `json:"priority"`         // 1-10
	Status          string          `json:"status"`           // pending, in_progress, completed
}

// ProductionOrder represents a production order
type ProductionOrder struct {
	ID              string          `json:"id,omitempty"`
	OrderNumber     string          `json:"order_number"`     // Auto-generated: MO-2026-0001
	ProductID       string          `json:"product_id"`
	ProductCode     string          `json:"product_code"`
	ProductName     string          `json:"product_name"`
	BOMID           string          `json:"bom_id"`
	BOMCode         string          `json:"bom_code"`
	Quantity        decimal.Decimal `json:"quantity"`
	QuantityProduced decimal.Decimal `json:"quantity_produced"`
	UnitOfMeasure   string          `json:"unit_of_measure"`
	ScheduledStart  time.Time       `json:"scheduled_start"`
	ScheduledEnd    time.Time       `json:"scheduled_end"`
	ActualStart     *time.Time      `json:"actual_start,omitempty"`
	ActualEnd       *time.Time      `json:"actual_end,omitempty"`
	Status          MOStatus        `json:"status"`           // draft, confirmed, in_progress, done, cancelled
	Priority        int             `json:"priority"`
	SourceDocument  string          `json:"source_document,omitempty"` // Sales order, etc.
	SourceDocumentID string         `json:"source_document_id,omitempty"`
	WarehouseID     string          `json:"warehouse_id"`
	WarehouseName   string          `json:"warehouse_name"`
	Notes           string          `json:"notes,omitempty"`
	CreatedBy       string          `json:"created_by"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

// MOStatus defines manufacturing order status
type MOStatus string

const (
	MOStatusDraft      MOStatus = "draft"
	MOStatusConfirmed  MOStatus = "confirmed"
	MOStatusInProgress MOStatus = "in_progress"
	MOStatusDone       MOStatus = "done"
	MOStatusCancelled  MOStatus = "cancelled"
)

// ============================================
// WORK ORDER
// ============================================

// WorkOrder represents a work order for production
type WorkOrder struct {
	ID              string          `json:"id,omitempty"`
	WorkOrderNumber string          `json:"work_order_number"` // Auto-generated: WO-2026-0001
	ProductionOrderID string        `json:"production_order_id"`
	ProductionOrderNumber string    `json:"production_order_number"`
	ProductID       string          `json:"product_id"`
	ProductCode     string          `json:"product_code"`
	ProductName     string          `json:"product_name"`
	BOMID           string          `json:"bom_id"`
	Quantity        decimal.Decimal `json:"quantity"`
	QuantityProduced decimal.Decimal `json:"quantity_produced"`
	UnitOfMeasure   string          `json:"unit_of_measure"`
	WorkCenterID    string          `json:"work_center_id,omitempty"`
	WorkCenterName  string          `json:"work_center_name,omitempty"`
	ScheduledStart  time.Time       `json:"scheduled_start"`
	ScheduledEnd    time.Time       `json:"scheduled_end"`
	ActualStart     *time.Time      `json:"actual_start,omitempty"`
	ActualEnd       *time.Time      `json:"actual_end,omitempty"`
	Status          WOStatus        `json:"status"`           // pending, ready, in_progress, paused, done, cancelled
	Progress        decimal.Decimal `json:"progress"`         // 0-100%
	Operations      []WorkOrderOperation `json:"operations"`
	MaterialConsumptions []MaterialConsumption `json:"material_consumptions,omitempty"`
	Notes           string          `json:"notes,omitempty"`
	AssignedTo      string          `json:"assigned_to,omitempty"`
	AssignedToName  string          `json:"assigned_to_name,omitempty"`
	CreatedBy       string          `json:"created_by"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

// WOStatus defines work order status
type WOStatus string

const (
	WOStatusPending    WOStatus = "pending"
	WOStatusReady      WOStatus = "ready"
	WOStatusInProgress WOStatus = "in_progress"
	WOStatusPaused     WOStatus = "paused"
	WOStatusDone       WOStatus = "done"
	WOStatusCancelled  WOStatus = "cancelled"
)

// WorkOrderOperation represents an operation in work order
type WorkOrderOperation struct {
	ID              string          `json:"id,omitempty"`
	Sequence        int             `json:"sequence"`
	OperationName   string          `json:"operation_name"`
	WorkCenterID    string          `json:"work_center_id,omitempty"`
	WorkCenterName  string          `json:"work_center_name,omitempty"`
	PlannedDuration decimal.Decimal `json:"planned_duration"` // minutes
	ActualDuration  decimal.Decimal `json:"actual_duration"`  // minutes
	Status          string          `json:"status"`           // pending, in_progress, done
	StartTime       *time.Time      `json:"start_time,omitempty"`
	EndTime         *time.Time      `json:"end_time,omitempty"`
	AssignedTo      string          `json:"assigned_to,omitempty"`
	Notes           string          `json:"notes,omitempty"`
}

// MaterialConsumption represents material consumed in production
type MaterialConsumption struct {
	ID              string          `json:"id,omitempty"`
	ProductID       string          `json:"product_id"`
	ProductCode     string          `json:"product_code"`
	ProductName     string          `json:"product_name"`
	PlannedQuantity decimal.Decimal `json:"planned_quantity"`
	ConsumedQuantity decimal.Decimal `json:"consumed_quantity"`
	UnitOfMeasure   string          `json:"unit_of_measure"`
	ConsumedAt      time.Time       `json:"consumed_at"`
	ConsumedBy      string          `json:"consumed_by"`
}

// ============================================
// QUALITY CONTROL
// ============================================

// QualityCheck represents a quality check/inspection
type QualityCheck struct {
	ID              string          `json:"id,omitempty"`
	CheckNumber     string          `json:"check_number"`     // Auto-generated: QC-2026-0001
	CheckType       QCType          `json:"check_type"`       // incoming, in_process, final, random
	ProductID       string          `json:"product_id"`
	ProductCode     string          `json:"product_code"`
	ProductName     string          `json:"product_name"`
	WorkOrderID     string          `json:"work_order_id,omitempty"`
	WorkOrderNumber string          `json:"work_order_number,omitempty"`
	BatchNumber     string          `json:"batch_number,omitempty"`
	QuantityChecked decimal.Decimal `json:"quantity_checked"`
	QuantityPassed  decimal.Decimal `json:"quantity_passed"`
	QuantityFailed  decimal.Decimal `json:"quantity_failed"`
	QuantityRework  decimal.Decimal `json:"quantity_rework"`
	UnitOfMeasure   string          `json:"unit_of_measure"`
	Status          QCStatus        `json:"status"`           // pending, in_progress, passed, failed, rework
	CheckDate       time.Time       `json:"check_date"`
	InspectionPoints []InspectionPoint `json:"inspection_points"`
	Notes           string          `json:"notes,omitempty"`
	FailureReason   string          `json:"failure_reason,omitempty"`
	InspectedBy     string          `json:"inspected_by"`
	InspectedByName string          `json:"inspected_by_name"`
	ApprovedBy      string          `json:"approved_by,omitempty"`
	ApprovedAt      *time.Time      `json:"approved_at,omitempty"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

// QCType defines quality check type
type QCType string

const (
	QCTypeIncoming   QCType = "incoming"
	QCTypeInProcess  QCType = "in_process"
	QCTypeFinal      QCType = "final"
	QCTypeRandom     QCType = "random"
)

// QCStatus defines quality check status
type QCStatus string

const (
	QCStatusPending    QCStatus = "pending"
	QCStatusInProgress QCStatus = "in_progress"
	QCStatusPassed     QCStatus = "passed"
	QCStatusFailed     QCStatus = "failed"
	QCStatusRework     QCStatus = "rework"
)

// InspectionPoint represents a quality inspection point
type InspectionPoint struct {
	ID              string          `json:"id,omitempty"`
	CheckName       string          `json:"check_name"`
	CheckType       string          `json:"check_type"`       // visual, measurement, functional
	Specification   string          `json:"specification"`
	MeasuredValue   string          `json:"measured_value,omitempty"`
	Result          InspectionResult `json:"result"`          // pass, fail
	Notes           string          `json:"notes,omitempty"`
}

// InspectionResult defines inspection result
type InspectionResult string

const (
	InspectionResultPass InspectionResult = "pass"
	InspectionResultFail InspectionResult = "fail"
)

// ============================================
// WORK CENTER
// ============================================

// WorkCenter represents a work center/machine/station
type WorkCenter struct {
	ID              string          `json:"id,omitempty"`
	WorkCenterCode  string          `json:"work_center_code"` // Auto-generated: WC-001
	WorkCenterName  string          `json:"work_center_name"`
	WorkCenterType  string          `json:"work_center_type"` // machine, assembly, packaging, etc.
	Description     string          `json:"description,omitempty"`
	Capacity        decimal.Decimal `json:"capacity"`         // units per hour
	CostPerHour     decimal.Decimal `json:"cost_per_hour"`
	WarehouseID     string          `json:"warehouse_id,omitempty"`
	WarehouseName   string          `json:"warehouse_name,omitempty"`
	IsActive        bool            `json:"is_active"`
	CreatedBy       string          `json:"created_by"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

// ============================================
// PRODUCTION REPORTS
// ============================================

// ProductionReport represents production summary report
type ProductionReport struct {
	FromDate        time.Time       `json:"from_date"`
	ToDate          time.Time       `json:"to_date"`
	TotalOrders     int             `json:"total_orders"`
	CompletedOrders int             `json:"completed_orders"`
	InProgressOrders int            `json:"in_progress_orders"`
	TotalQuantityPlanned decimal.Decimal `json:"total_quantity_planned"`
	TotalQuantityProduced decimal.Decimal `json:"total_quantity_produced"`
	EfficiencyRate  decimal.Decimal `json:"efficiency_rate"`  // %
	QualityRate     decimal.Decimal `json:"quality_rate"`     // %
	TotalCost       decimal.Decimal `json:"total_cost"`
	ProductionByProduct []ProductionByProduct `json:"production_by_product"`
}

// ProductionByProduct represents production grouped by product
type ProductionByProduct struct {
	ProductID       string          `json:"product_id"`
	ProductCode     string          `json:"product_code"`
	ProductName     string          `json:"product_name"`
	QuantityPlanned decimal.Decimal `json:"quantity_planned"`
	QuantityProduced decimal.Decimal `json:"quantity_produced"`
	TotalCost       decimal.Decimal `json:"total_cost"`
}
