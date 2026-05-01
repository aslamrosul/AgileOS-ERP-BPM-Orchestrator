package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// ============================================
// PRODUCT/ITEM MASTER
// ============================================

// Product represents a product/item in inventory
type Product struct {
	ID              string          `json:"id,omitempty"`
	ProductCode     string          `json:"product_code"`     // Auto-generated: PRD-0001
	ProductName     string          `json:"product_name"`
	Description     string          `json:"description,omitempty"`
	Category        string          `json:"category"`
	SubCategory     string          `json:"sub_category,omitempty"`
	UnitOfMeasure   string          `json:"unit_of_measure"`  // pcs, kg, liter, box, etc.
	ProductType     ProductType     `json:"product_type"`     // goods, service, consumable
	
	// Pricing
	CostPrice       decimal.Decimal `json:"cost_price"`
	SellingPrice    decimal.Decimal `json:"selling_price"`
	Currency        string          `json:"currency"`
	
	// Inventory
	TrackInventory  bool            `json:"track_inventory"`
	MinStockLevel   decimal.Decimal `json:"min_stock_level"`
	MaxStockLevel   decimal.Decimal `json:"max_stock_level"`
	ReorderLevel    decimal.Decimal `json:"reorder_level"`
	ReorderQuantity decimal.Decimal `json:"reorder_quantity"`
	
	// Physical Properties
	Weight          decimal.Decimal `json:"weight,omitempty"`
	WeightUnit      string          `json:"weight_unit,omitempty"`
	Volume          decimal.Decimal `json:"volume,omitempty"`
	VolumeUnit      string          `json:"volume_unit,omitempty"`
	
	// Identification
	Barcode         string          `json:"barcode,omitempty"`
	SKU             string          `json:"sku,omitempty"`
	
	// Status
	IsActive        bool            `json:"is_active"`
	IsSaleable      bool            `json:"is_saleable"`
	IsPurchaseable  bool            `json:"is_purchaseable"`
	
	// Metadata
	CreatedBy       string          `json:"created_by"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

// ProductType defines the type of product
type ProductType string

const (
	ProductTypeGoods      ProductType = "goods"
	ProductTypeService    ProductType = "service"
	ProductTypeConsumable ProductType = "consumable"
)

// ============================================
// STOCK MANAGEMENT
// ============================================

// Stock represents product stock in a warehouse
type Stock struct {
	ID              string          `json:"id,omitempty"`
	ProductID       string          `json:"product_id"`
	ProductCode     string          `json:"product_code"`
	ProductName     string          `json:"product_name"`
	WarehouseID     string          `json:"warehouse_id"`
	WarehouseName   string          `json:"warehouse_name"`
	LocationBin     string          `json:"location_bin,omitempty"`
	QuantityOnHand  decimal.Decimal `json:"quantity_on_hand"`
	QuantityReserved decimal.Decimal `json:"quantity_reserved"`
	QuantityAvailable decimal.Decimal `json:"quantity_available"`
	UnitOfMeasure   string          `json:"unit_of_measure"`
	ValuationMethod StockValuation  `json:"valuation_method"` // fifo, lifo, average
	AverageCost     decimal.Decimal `json:"average_cost"`
	TotalValue      decimal.Decimal `json:"total_value"`
	LastStockDate   time.Time       `json:"last_stock_date"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

// StockValuation defines stock valuation method
type StockValuation string

const (
	StockValuationFIFO    StockValuation = "fifo"
	StockValuationLIFO    StockValuation = "lifo"
	StockValuationAverage StockValuation = "average"
)

// StockMovement represents stock movement transaction
type StockMovement struct {
	ID              string          `json:"id,omitempty"`
	MovementNumber  string          `json:"movement_number"`  // Auto-generated: SM-2026-0001
	MovementType    MovementType    `json:"movement_type"`    // in, out, transfer, adjustment
	ProductID       string          `json:"product_id"`
	ProductCode     string          `json:"product_code"`
	ProductName     string          `json:"product_name"`
	WarehouseID     string          `json:"warehouse_id"`
	WarehouseName   string          `json:"warehouse_name"`
	Quantity        decimal.Decimal `json:"quantity"`
	UnitOfMeasure   string          `json:"unit_of_measure"`
	UnitCost        decimal.Decimal `json:"unit_cost"`
	TotalCost       decimal.Decimal `json:"total_cost"`
	ReferenceType   string          `json:"reference_type,omitempty"`   // purchase_order, sales_order, etc.
	ReferenceID     string          `json:"reference_id,omitempty"`
	ReferenceNumber string          `json:"reference_number,omitempty"`
	Notes           string          `json:"notes,omitempty"`
	MovementDate    time.Time       `json:"movement_date"`
	CreatedBy       string          `json:"created_by"`
	CreatedAt       time.Time       `json:"created_at"`
}

// MovementType defines stock movement type
type MovementType string

const (
	MovementTypeIn         MovementType = "in"
	MovementTypeOut        MovementType = "out"
	MovementTypeTransfer   MovementType = "transfer"
	MovementTypeAdjustment MovementType = "adjustment"
)

// StockAdjustment represents stock adjustment
type StockAdjustment struct {
	ID              string          `json:"id,omitempty"`
	AdjustmentNumber string         `json:"adjustment_number"` // Auto-generated: SA-2026-0001
	ProductID       string          `json:"product_id"`
	ProductCode     string          `json:"product_code"`
	ProductName     string          `json:"product_name"`
	WarehouseID     string          `json:"warehouse_id"`
	WarehouseName   string          `json:"warehouse_name"`
	OldQuantity     decimal.Decimal `json:"old_quantity"`
	NewQuantity     decimal.Decimal `json:"new_quantity"`
	AdjustmentQty   decimal.Decimal `json:"adjustment_qty"`
	Reason          string          `json:"reason"`
	Notes           string          `json:"notes,omitempty"`
	AdjustmentDate  time.Time       `json:"adjustment_date"`
	ApprovedBy      string          `json:"approved_by,omitempty"`
	ApprovedAt      *time.Time      `json:"approved_at,omitempty"`
	CreatedBy       string          `json:"created_by"`
	CreatedAt       time.Time       `json:"created_at"`
}

// ============================================
// WAREHOUSE MANAGEMENT
// ============================================

// Warehouse represents a warehouse/storage location
type Warehouse struct {
	ID              string          `json:"id,omitempty"`
	WarehouseCode   string          `json:"warehouse_code"`   // Auto-generated: WH-001
	WarehouseName   string          `json:"warehouse_name"`
	Description     string          `json:"description,omitempty"`
	WarehouseType   WarehouseType   `json:"warehouse_type"`   // main, branch, transit, virtual
	Address         string          `json:"address"`
	City            string          `json:"city"`
	State           string          `json:"state,omitempty"`
	Country         string          `json:"country"`
	PostalCode      string          `json:"postal_code,omitempty"`
	Phone           string          `json:"phone,omitempty"`
	Email           string          `json:"email,omitempty"`
	ManagerID       string          `json:"manager_id,omitempty"`
	ManagerName     string          `json:"manager_name,omitempty"`
	IsActive        bool            `json:"is_active"`
	CreatedBy       string          `json:"created_by"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

// WarehouseType defines warehouse type
type WarehouseType string

const (
	WarehouseTypeMain    WarehouseType = "main"
	WarehouseTypeBranch  WarehouseType = "branch"
	WarehouseTypeTransit WarehouseType = "transit"
	WarehouseTypeVirtual WarehouseType = "virtual"
)

// StockTransfer represents stock transfer between warehouses
type StockTransfer struct {
	ID              string          `json:"id,omitempty"`
	TransferNumber  string          `json:"transfer_number"`  // Auto-generated: ST-2026-0001
	FromWarehouseID string          `json:"from_warehouse_id"`
	FromWarehouseName string        `json:"from_warehouse_name"`
	ToWarehouseID   string          `json:"to_warehouse_id"`
	ToWarehouseName string          `json:"to_warehouse_name"`
	TransferDate    time.Time       `json:"transfer_date"`
	Status          TransferStatus  `json:"status"`           // draft, sent, in_transit, received, cancelled
	Lines           []StockTransferLine `json:"lines"`
	Notes           string          `json:"notes,omitempty"`
	SentBy          string          `json:"sent_by,omitempty"`
	SentAt          *time.Time      `json:"sent_at,omitempty"`
	ReceivedBy      string          `json:"received_by,omitempty"`
	ReceivedAt      *time.Time      `json:"received_at,omitempty"`
	CreatedBy       string          `json:"created_by"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

// TransferStatus defines stock transfer status
type TransferStatus string

const (
	TransferStatusDraft     TransferStatus = "draft"
	TransferStatusSent      TransferStatus = "sent"
	TransferStatusInTransit TransferStatus = "in_transit"
	TransferStatusReceived  TransferStatus = "received"
	TransferStatusCancelled TransferStatus = "cancelled"
)

// StockTransferLine represents a line in stock transfer
type StockTransferLine struct {
	ID              string          `json:"id,omitempty"`
	ProductID       string          `json:"product_id"`
	ProductCode     string          `json:"product_code"`
	ProductName     string          `json:"product_name"`
	QuantitySent    decimal.Decimal `json:"quantity_sent"`
	QuantityReceived decimal.Decimal `json:"quantity_received"`
	UnitOfMeasure   string          `json:"unit_of_measure"`
	Notes           string          `json:"notes,omitempty"`
}

// ============================================
// PURCHASING
// ============================================

// PurchaseRequisition represents a purchase requisition
type PurchaseRequisition struct {
	ID              string          `json:"id,omitempty"`
	PRNumber        string          `json:"pr_number"`        // Auto-generated: PR-2026-0001
	RequestDate     time.Time       `json:"request_date"`
	RequiredDate    time.Time       `json:"required_date"`
	Department      string          `json:"department"`
	RequestedBy     string          `json:"requested_by"`
	RequestedByName string          `json:"requested_by_name"`
	Status          PRStatus        `json:"status"`           // draft, submitted, approved, rejected, ordered
	Lines           []PRLine        `json:"lines"`
	Notes           string          `json:"notes,omitempty"`
	ApprovedBy      string          `json:"approved_by,omitempty"`
	ApprovedAt      *time.Time      `json:"approved_at,omitempty"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

// PRStatus defines purchase requisition status
type PRStatus string

const (
	PRStatusDraft     PRStatus = "draft"
	PRStatusSubmitted PRStatus = "submitted"
	PRStatusApproved  PRStatus = "approved"
	PRStatusRejected  PRStatus = "rejected"
	PRStatusOrdered   PRStatus = "ordered"
)

// PRLine represents a line in purchase requisition
type PRLine struct {
	ID              string          `json:"id,omitempty"`
	ProductID       string          `json:"product_id"`
	ProductCode     string          `json:"product_code"`
	ProductName     string          `json:"product_name"`
	Description     string          `json:"description,omitempty"`
	Quantity        decimal.Decimal `json:"quantity"`
	UnitOfMeasure   string          `json:"unit_of_measure"`
	EstimatedPrice  decimal.Decimal `json:"estimated_price"`
	TotalEstimated  decimal.Decimal `json:"total_estimated"`
}

// PurchaseOrder represents a purchase order
type PurchaseOrder struct {
	ID              string          `json:"id,omitempty"`
	PONumber        string          `json:"po_number"`        // Auto-generated: PO-2026-0001
	VendorID        string          `json:"vendor_id"`
	VendorName      string          `json:"vendor_name"`
	OrderDate       time.Time       `json:"order_date"`
	ExpectedDate    time.Time       `json:"expected_date"`
	Status          POStatus        `json:"status"`           // draft, sent, confirmed, partial, received, cancelled
	Lines           []POLine        `json:"lines"`
	SubTotal        decimal.Decimal `json:"sub_total"`
	TaxAmount       decimal.Decimal `json:"tax_amount"`
	DiscountAmount  decimal.Decimal `json:"discount_amount"`
	TotalAmount     decimal.Decimal `json:"total_amount"`
	PaymentTerms    int             `json:"payment_terms"`
	DeliveryAddress string          `json:"delivery_address"`
	Notes           string          `json:"notes,omitempty"`
	ApprovedBy      string          `json:"approved_by,omitempty"`
	ApprovedAt      *time.Time      `json:"approved_at,omitempty"`
	CreatedBy       string          `json:"created_by"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

// POStatus defines purchase order status
type POStatus string

const (
	POStatusDraft     POStatus = "draft"
	POStatusSent      POStatus = "sent"
	POStatusConfirmed POStatus = "confirmed"
	POStatusPartial   POStatus = "partial"
	POStatusReceived  POStatus = "received"
	POStatusCancelled POStatus = "cancelled"
)

// POLine represents a line in purchase order
type POLine struct {
	ID              string          `json:"id,omitempty"`
	ProductID       string          `json:"product_id"`
	ProductCode     string          `json:"product_code"`
	ProductName     string          `json:"product_name"`
	Description     string          `json:"description,omitempty"`
	Quantity        decimal.Decimal `json:"quantity"`
	QuantityReceived decimal.Decimal `json:"quantity_received"`
	UnitOfMeasure   string          `json:"unit_of_measure"`
	UnitPrice       decimal.Decimal `json:"unit_price"`
	TaxRate         decimal.Decimal `json:"tax_rate"`
	DiscountRate    decimal.Decimal `json:"discount_rate"`
	LineTotal       decimal.Decimal `json:"line_total"`
}

// GoodsReceipt represents goods receipt from purchase order
type GoodsReceipt struct {
	ID              string          `json:"id,omitempty"`
	GRNumber        string          `json:"gr_number"`        // Auto-generated: GR-2026-0001
	PONumber        string          `json:"po_number"`
	POID            string          `json:"po_id"`
	VendorID        string          `json:"vendor_id"`
	VendorName      string          `json:"vendor_name"`
	ReceiptDate     time.Time       `json:"receipt_date"`
	WarehouseID     string          `json:"warehouse_id"`
	WarehouseName   string          `json:"warehouse_name"`
	Status          GRStatus        `json:"status"`           // draft, confirmed, cancelled
	Lines           []GRLine        `json:"lines"`
	Notes           string          `json:"notes,omitempty"`
	ReceivedBy      string          `json:"received_by"`
	ConfirmedBy     string          `json:"confirmed_by,omitempty"`
	ConfirmedAt     *time.Time      `json:"confirmed_at,omitempty"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

// GRStatus defines goods receipt status
type GRStatus string

const (
	GRStatusDraft     GRStatus = "draft"
	GRStatusConfirmed GRStatus = "confirmed"
	GRStatusCancelled GRStatus = "cancelled"
)

// GRLine represents a line in goods receipt
type GRLine struct {
	ID              string          `json:"id,omitempty"`
	POLineID        string          `json:"po_line_id"`
	ProductID       string          `json:"product_id"`
	ProductCode     string          `json:"product_code"`
	ProductName     string          `json:"product_name"`
	QuantityOrdered decimal.Decimal `json:"quantity_ordered"`
	QuantityReceived decimal.Decimal `json:"quantity_received"`
	UnitOfMeasure   string          `json:"unit_of_measure"`
	Notes           string          `json:"notes,omitempty"`
}
