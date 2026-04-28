package models

import "time"

// Document represents a business document in the workflow
type Document struct {
	ID                string                   `json:"id,omitempty"`
	Title             string                   `json:"title"`
	Description       string                   `json:"description,omitempty"`
	Type              string                   `json:"type"` // invoice, purchase_order, contract, etc.
	Status            string                   `json:"status"` // draft, pending_signature, signed, final, rejected
	WorkflowID        string                   `json:"workflow_id,omitempty"`
	ProcessInstanceID string                   `json:"process_instance_id,omitempty"`
	Content           []byte                   `json:"content,omitempty"` // Document content (PDF, JSON, etc.)
	ContentType       string                   `json:"content_type"` // application/pdf, application/json
	ContentHash       string                   `json:"content_hash"` // SHA-256 hash of content
	Metadata          map[string]interface{}   `json:"metadata,omitempty"`
	SignatureHistory  []DocumentSignature      `json:"signature_history"`
	CreatedBy         string                   `json:"created_by"`
	CreatedAt         time.Time                `json:"created_at"`
	UpdatedAt         time.Time                `json:"updated_at"`
	FinalizedAt       *time.Time               `json:"finalized_at,omitempty"`
}

// DocumentSignature represents a single signature on a document
type DocumentSignature struct {
	SignedBy      string                 `json:"signed_by"`
	SignedByName  string                 `json:"signed_by_name"`
	SignedAt      time.Time              `json:"signed_at"`
	Action        string                 `json:"action"` // approved, rejected, reviewed
	Signature     string                 `json:"signature"` // SHA-256 hash
	TaskID        string                 `json:"task_id,omitempty"`
	IPAddress     string                 `json:"ip_address,omitempty"`
	Comments      string                 `json:"comments,omitempty"`
	AdditionalData map[string]interface{} `json:"additional_data,omitempty"`
}

// DocumentReceipt represents a proof of approval receipt
type DocumentReceipt struct {
	ID                string              `json:"id"`
	DocumentID        string              `json:"document_id"`
	DocumentTitle     string              `json:"document_title"`
	ProcessName       string              `json:"process_name"`
	WorkflowID        string              `json:"workflow_id"`
	ProcessInstanceID string              `json:"process_instance_id"`
	Signatures        []DocumentSignature `json:"signatures"`
	GeneratedAt       time.Time           `json:"generated_at"`
	VerificationURL   string              `json:"verification_url"`
	QRCodeData        string              `json:"qr_code_data"`
	Status            string              `json:"status"`
}

// SignatureVerificationRequest represents a request to verify a signature
type SignatureVerificationRequest struct {
	TaskID     string                 `json:"task_id"`
	Signature  string                 `json:"signature"`
	UserID     string                 `json:"user_id,omitempty"`
	WorkflowID string                 `json:"workflow_id,omitempty"`
	Timestamp  time.Time              `json:"timestamp,omitempty"`
	Data       map[string]interface{} `json:"data,omitempty"`
}

// DocumentStatus constants
const (
	DocumentStatusDraft            = "draft"
	DocumentStatusPendingSignature = "pending_signature"
	DocumentStatusSigned           = "signed"
	DocumentStatusFinal            = "final"
	DocumentStatusRejected         = "rejected"
)

// SignatureAction constants
const (
	SignatureActionApproved = "approved"
	SignatureActionRejected = "rejected"
	SignatureActionReviewed = "reviewed"
	SignatureActionSigned   = "signed"
)
