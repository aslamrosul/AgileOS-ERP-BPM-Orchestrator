package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// ============================================
// CONTACT MANAGEMENT
// ============================================

// Contact represents a contact person
type Contact struct {
	ID              string          `json:"id,omitempty"`
	ContactCode     string          `json:"contact_code"`     // Auto-generated: CON-0001
	FirstName       string          `json:"first_name"`
	LastName        string          `json:"last_name"`
	FullName        string          `json:"full_name"`
	Email           string          `json:"email"`
	Phone           string          `json:"phone"`
	Mobile          string          `json:"mobile,omitempty"`
	JobTitle        string          `json:"job_title,omitempty"`
	Department      string          `json:"department,omitempty"`
	Company         string          `json:"company,omitempty"`
	CompanyID       string          `json:"company_id,omitempty"`
	Address         string          `json:"address,omitempty"`
	City            string          `json:"city,omitempty"`
	State           string          `json:"state,omitempty"`
	Country         string          `json:"country,omitempty"`
	PostalCode      string          `json:"postal_code,omitempty"`
	Website         string          `json:"website,omitempty"`
	LinkedIn        string          `json:"linkedin,omitempty"`
	Twitter         string          `json:"twitter,omitempty"`
	ContactType     ContactType     `json:"contact_type"`     // lead, customer, partner, vendor
	Source          string          `json:"source,omitempty"` // website, referral, campaign, etc.
	Tags            []string        `json:"tags,omitempty"`
	Notes           string          `json:"notes,omitempty"`
	IsActive        bool            `json:"is_active"`
	CreatedBy       string          `json:"created_by"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

// ContactType defines contact type
type ContactType string

const (
	ContactTypeLead     ContactType = "lead"
	ContactTypeCustomer ContactType = "customer"
	ContactTypePartner  ContactType = "partner"
	ContactTypeVendor   ContactType = "vendor"
)

// ============================================
// LEAD MANAGEMENT
// ============================================

// Lead represents a sales lead
type Lead struct {
	ID              string          `json:"id,omitempty"`
	LeadNumber      string          `json:"lead_number"`      // Auto-generated: LEAD-2026-0001
	LeadName        string          `json:"lead_name"`
	Company         string          `json:"company,omitempty"`
	ContactID       string          `json:"contact_id,omitempty"`
	ContactName     string          `json:"contact_name,omitempty"`
	Email           string          `json:"email"`
	Phone           string          `json:"phone,omitempty"`
	Source          LeadSource      `json:"source"`           // website, referral, campaign, cold_call, etc.
	Status          LeadStatus      `json:"status"`           // new, contacted, qualified, unqualified, converted
	LeadScore       int             `json:"lead_score"`       // 0-100
	Industry        string          `json:"industry,omitempty"`
	EstimatedValue  decimal.Decimal `json:"estimated_value"`
	Currency        string          `json:"currency"`
	ExpectedCloseDate *time.Time    `json:"expected_close_date,omitempty"`
	Description     string          `json:"description,omitempty"`
	Notes           string          `json:"notes,omitempty"`
	AssignedTo      string          `json:"assigned_to,omitempty"`
	AssignedToName  string          `json:"assigned_to_name,omitempty"`
	ConvertedToOpportunityID string `json:"converted_to_opportunity_id,omitempty"`
	ConvertedAt     *time.Time      `json:"converted_at,omitempty"`
	CreatedBy       string          `json:"created_by"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

// LeadSource defines lead source
type LeadSource string

const (
	LeadSourceWebsite   LeadSource = "website"
	LeadSourceReferral  LeadSource = "referral"
	LeadSourceCampaign  LeadSource = "campaign"
	LeadSourceColdCall  LeadSource = "cold_call"
	LeadSourceSocialMedia LeadSource = "social_media"
	LeadSourceEvent     LeadSource = "event"
	LeadSourceOther     LeadSource = "other"
)

// LeadStatus defines lead status
type LeadStatus string

const (
	LeadStatusNew         LeadStatus = "new"
	LeadStatusContacted   LeadStatus = "contacted"
	LeadStatusQualified   LeadStatus = "qualified"
	LeadStatusUnqualified LeadStatus = "unqualified"
	LeadStatusConverted   LeadStatus = "converted"
)

// ============================================
// OPPORTUNITY/SALES PIPELINE
// ============================================

// Opportunity represents a sales opportunity
type Opportunity struct {
	ID              string          `json:"id,omitempty"`
	OpportunityNumber string        `json:"opportunity_number"` // Auto-generated: OPP-2026-0001
	OpportunityName string          `json:"opportunity_name"`
	CustomerID      string          `json:"customer_id"`
	CustomerName    string          `json:"customer_name"`
	ContactID       string          `json:"contact_id,omitempty"`
	ContactName     string          `json:"contact_name,omitempty"`
	Stage           OpportunityStage `json:"stage"`            // prospecting, qualification, proposal, negotiation, closed_won, closed_lost
	Probability     int             `json:"probability"`       // 0-100%
	ExpectedRevenue decimal.Decimal `json:"expected_revenue"`
	ActualRevenue   decimal.Decimal `json:"actual_revenue"`
	Currency        string          `json:"currency"`
	ExpectedCloseDate time.Time     `json:"expected_close_date"`
	ActualCloseDate *time.Time      `json:"actual_close_date,omitempty"`
	Source          string          `json:"source,omitempty"`
	Description     string          `json:"description,omitempty"`
	Notes           string          `json:"notes,omitempty"`
	LossReason      string          `json:"loss_reason,omitempty"`
	AssignedTo      string          `json:"assigned_to"`
	AssignedToName  string          `json:"assigned_to_name"`
	CreatedBy       string          `json:"created_by"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

// OpportunityStage defines opportunity stage
type OpportunityStage string

const (
	OpportunityStageProspecting  OpportunityStage = "prospecting"
	OpportunityStageQualification OpportunityStage = "qualification"
	OpportunityStageProposal     OpportunityStage = "proposal"
	OpportunityStageNegotiation  OpportunityStage = "negotiation"
	OpportunityStageClosedWon    OpportunityStage = "closed_won"
	OpportunityStageClosedLost   OpportunityStage = "closed_lost"
)

// ============================================
// QUOTATION & SALES ORDER
// ============================================

// Quotation represents a sales quotation
type Quotation struct {
	ID              string          `json:"id,omitempty"`
	QuotationNumber string          `json:"quotation_number"` // Auto-generated: QUO-2026-0001
	CustomerID      string          `json:"customer_id"`
	CustomerName    string          `json:"customer_name"`
	ContactID       string          `json:"contact_id,omitempty"`
	ContactName     string          `json:"contact_name,omitempty"`
	QuotationDate   time.Time       `json:"quotation_date"`
	ValidUntil      time.Time       `json:"valid_until"`
	Status          QuotationStatus `json:"status"`           // draft, sent, accepted, rejected, expired
	Lines           []QuotationLine `json:"lines"`
	SubTotal        decimal.Decimal `json:"sub_total"`
	TaxAmount       decimal.Decimal `json:"tax_amount"`
	DiscountAmount  decimal.Decimal `json:"discount_amount"`
	TotalAmount     decimal.Decimal `json:"total_amount"`
	Currency        string          `json:"currency"`
	PaymentTerms    string          `json:"payment_terms,omitempty"`
	DeliveryTerms   string          `json:"delivery_terms,omitempty"`
	Notes           string          `json:"notes,omitempty"`
	TermsConditions string          `json:"terms_conditions,omitempty"`
	SentAt          *time.Time      `json:"sent_at,omitempty"`
	AcceptedAt      *time.Time      `json:"accepted_at,omitempty"`
	CreatedBy       string          `json:"created_by"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

// QuotationStatus defines quotation status
type QuotationStatus string

const (
	QuotationStatusDraft    QuotationStatus = "draft"
	QuotationStatusSent     QuotationStatus = "sent"
	QuotationStatusAccepted QuotationStatus = "accepted"
	QuotationStatusRejected QuotationStatus = "rejected"
	QuotationStatusExpired  QuotationStatus = "expired"
)

// QuotationLine represents a line in quotation
type QuotationLine struct {
	ID              string          `json:"id,omitempty"`
	ProductID       string          `json:"product_id"`
	ProductCode     string          `json:"product_code"`
	ProductName     string          `json:"product_name"`
	Description     string          `json:"description,omitempty"`
	Quantity        decimal.Decimal `json:"quantity"`
	UnitOfMeasure   string          `json:"unit_of_measure"`
	UnitPrice       decimal.Decimal `json:"unit_price"`
	TaxRate         decimal.Decimal `json:"tax_rate"`
	DiscountRate    decimal.Decimal `json:"discount_rate"`
	LineTotal       decimal.Decimal `json:"line_total"`
}

// SalesOrder represents a sales order
type SalesOrder struct {
	ID              string          `json:"id,omitempty"`
	OrderNumber     string          `json:"order_number"`     // Auto-generated: SO-2026-0001
	QuotationID     string          `json:"quotation_id,omitempty"`
	QuotationNumber string          `json:"quotation_number,omitempty"`
	CustomerID      string          `json:"customer_id"`
	CustomerName    string          `json:"customer_name"`
	ContactID       string          `json:"contact_id,omitempty"`
	ContactName     string          `json:"contact_name,omitempty"`
	OrderDate       time.Time       `json:"order_date"`
	ExpectedDeliveryDate time.Time  `json:"expected_delivery_date"`
	ActualDeliveryDate *time.Time   `json:"actual_delivery_date,omitempty"`
	Status          SOStatus        `json:"status"`           // draft, confirmed, processing, delivered, invoiced, cancelled
	Lines           []SalesOrderLine `json:"lines"`
	SubTotal        decimal.Decimal `json:"sub_total"`
	TaxAmount       decimal.Decimal `json:"tax_amount"`
	DiscountAmount  decimal.Decimal `json:"discount_amount"`
	TotalAmount     decimal.Decimal `json:"total_amount"`
	Currency        string          `json:"currency"`
	PaymentTerms    string          `json:"payment_terms,omitempty"`
	DeliveryAddress string          `json:"delivery_address"`
	DeliveryNotes   string          `json:"delivery_notes,omitempty"`
	Notes           string          `json:"notes,omitempty"`
	ConfirmedBy     string          `json:"confirmed_by,omitempty"`
	ConfirmedAt     *time.Time      `json:"confirmed_at,omitempty"`
	InvoiceID       string          `json:"invoice_id,omitempty"`
	InvoiceNumber   string          `json:"invoice_number,omitempty"`
	CreatedBy       string          `json:"created_by"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

// SOStatus defines sales order status
type SOStatus string

const (
	SOStatusDraft      SOStatus = "draft"
	SOStatusConfirmed  SOStatus = "confirmed"
	SOStatusProcessing SOStatus = "processing"
	SOStatusDelivered  SOStatus = "delivered"
	SOStatusInvoiced   SOStatus = "invoiced"
	SOStatusCancelled  SOStatus = "cancelled"
)

// SalesOrderLine represents a line in sales order
type SalesOrderLine struct {
	ID              string          `json:"id,omitempty"`
	ProductID       string          `json:"product_id"`
	ProductCode     string          `json:"product_code"`
	ProductName     string          `json:"product_name"`
	Description     string          `json:"description,omitempty"`
	Quantity        decimal.Decimal `json:"quantity"`
	QuantityDelivered decimal.Decimal `json:"quantity_delivered"`
	UnitOfMeasure   string          `json:"unit_of_measure"`
	UnitPrice       decimal.Decimal `json:"unit_price"`
	TaxRate         decimal.Decimal `json:"tax_rate"`
	DiscountRate    decimal.Decimal `json:"discount_rate"`
	LineTotal       decimal.Decimal `json:"line_total"`
}

// ============================================
// CAMPAIGN MANAGEMENT
// ============================================

// Campaign represents a marketing campaign
type Campaign struct {
	ID              string          `json:"id,omitempty"`
	CampaignCode    string          `json:"campaign_code"`    // Auto-generated: CAM-2026-0001
	CampaignName    string          `json:"campaign_name"`
	CampaignType    CampaignType    `json:"campaign_type"`    // email, social, event, webinar, etc.
	Status          CampaignStatus  `json:"status"`           // planned, active, completed, cancelled
	StartDate       time.Time       `json:"start_date"`
	EndDate         time.Time       `json:"end_date"`
	Budget          decimal.Decimal `json:"budget"`
	ActualCost      decimal.Decimal `json:"actual_cost"`
	Currency        string          `json:"currency"`
	TargetAudience  string          `json:"target_audience,omitempty"`
	Description     string          `json:"description,omitempty"`
	
	// Metrics
	LeadsGenerated  int             `json:"leads_generated"`
	OpportunitiesCreated int        `json:"opportunities_created"`
	Revenue         decimal.Decimal `json:"revenue"`
	ROI             decimal.Decimal `json:"roi"`              // Return on Investment
	
	AssignedTo      string          `json:"assigned_to,omitempty"`
	AssignedToName  string          `json:"assigned_to_name,omitempty"`
	CreatedBy       string          `json:"created_by"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

// CampaignType defines campaign type
type CampaignType string

const (
	CampaignTypeEmail    CampaignType = "email"
	CampaignTypeSocial   CampaignType = "social"
	CampaignTypeEvent    CampaignType = "event"
	CampaignTypeWebinar  CampaignType = "webinar"
	CampaignTypeContent  CampaignType = "content"
	CampaignTypeOther    CampaignType = "other"
)

// CampaignStatus defines campaign status
type CampaignStatus string

const (
	CampaignStatusPlanned   CampaignStatus = "planned"
	CampaignStatusActive    CampaignStatus = "active"
	CampaignStatusCompleted CampaignStatus = "completed"
	CampaignStatusCancelled CampaignStatus = "cancelled"
)

// ============================================
// ACTIVITY TRACKING
// ============================================

// Activity represents a CRM activity (call, meeting, email, task)
type Activity struct {
	ID              string          `json:"id,omitempty"`
	ActivityType    ActivityType    `json:"activity_type"`    // call, meeting, email, task, note
	Subject         string          `json:"subject"`
	Description     string          `json:"description,omitempty"`
	Status          ActivityStatus  `json:"status"`           // planned, completed, cancelled
	Priority        Priority        `json:"priority"`         // low, medium, high
	DueDate         *time.Time      `json:"due_date,omitempty"`
	CompletedDate   *time.Time      `json:"completed_date,omitempty"`
	Duration        int             `json:"duration,omitempty"` // in minutes
	
	// Related To
	RelatedType     string          `json:"related_type,omitempty"`     // lead, opportunity, contact, customer
	RelatedID       string          `json:"related_id,omitempty"`
	RelatedName     string          `json:"related_name,omitempty"`
	
	AssignedTo      string          `json:"assigned_to"`
	AssignedToName  string          `json:"assigned_to_name"`
	CreatedBy       string          `json:"created_by"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

// ActivityType defines activity type
type ActivityType string

const (
	ActivityTypeCall    ActivityType = "call"
	ActivityTypeMeeting ActivityType = "meeting"
	ActivityTypeEmail   ActivityType = "email"
	ActivityTypeTask    ActivityType = "task"
	ActivityTypeNote    ActivityType = "note"
)

// ActivityStatus defines activity status
type ActivityStatus string

const (
	ActivityStatusPlanned   ActivityStatus = "planned"
	ActivityStatusCompleted ActivityStatus = "completed"
	ActivityStatusCancelled ActivityStatus = "cancelled"
)

// Priority defines priority level
type Priority string

const (
	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"
)
