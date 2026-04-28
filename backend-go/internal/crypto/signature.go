package crypto

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

// SignatureData represents the data used to generate a digital signature
type SignatureData struct {
	TaskID     string
	UserID     string
	Timestamp  time.Time
	WorkflowID string
	Action     string
	Data       map[string]interface{}
}

// GenerateSignature creates a SHA-256 hash for task approval
func GenerateSignature(data SignatureData) string {
	combined := fmt.Sprintf("%s|%s|%d|%s|%s",
		data.TaskID,
		data.UserID,
		data.Timestamp.Unix(),
		data.WorkflowID,
		data.Action,
	)

	if len(data.Data) > 0 {
		for key, value := range data.Data {
			combined += fmt.Sprintf("|%s:%v", key, value)
		}
	}

	hash := sha256.Sum256([]byte(combined))
	return hex.EncodeToString(hash[:])
}

// VerifySignature checks if a signature matches the provided data
func VerifySignature(signature string, data SignatureData) bool {
	expectedSignature := GenerateSignature(data)
	return signature == expectedSignature
}

// GenerateDocumentHash creates a hash for document integrity
func GenerateDocumentHash(documentID string, content []byte, metadata map[string]interface{}) string {
	contentHash := sha256.Sum256(content)
	
	combined := fmt.Sprintf("%s|%s",
		documentID,
		hex.EncodeToString(contentHash[:]),
	)

	if len(metadata) > 0 {
		for key, value := range metadata {
			combined += fmt.Sprintf("|%s:%v", key, value)
		}
	}

	hash := sha256.Sum256([]byte(combined))
	return hex.EncodeToString(hash[:])
}

// SignatureMetadata contains information about a signature
type SignatureMetadata struct {
	Signature      string                 `json:"signature"`
	SignedBy       string                 `json:"signed_by"`
	SignedAt       time.Time              `json:"signed_at"`
	Action         string                 `json:"action"`
	IPAddress      string                 `json:"ip_address,omitempty"`
	UserAgent      string                 `json:"user_agent,omitempty"`
	AdditionalData map[string]interface{} `json:"additional_data,omitempty"`
}

// VerificationResult contains the result of signature verification
type VerificationResult struct {
	Valid        bool      `json:"valid"`
	Signature    string    `json:"signature"`
	VerifiedAt   time.Time `json:"verified_at"`
	Message      string    `json:"message"`
	ExpectedHash string    `json:"expected_hash,omitempty"`
	ActualHash   string    `json:"actual_hash,omitempty"`
}

// VerifyTaskSignature performs comprehensive signature verification
func VerifyTaskSignature(taskID, userID, workflowID, action, signature string, timestamp time.Time, data map[string]interface{}) VerificationResult {
	signatureData := SignatureData{
		TaskID:     taskID,
		UserID:     userID,
		Timestamp:  timestamp,
		WorkflowID: workflowID,
		Action:     action,
		Data:       data,
	}

	expectedSignature := GenerateSignature(signatureData)
	valid := signature == expectedSignature

	result := VerificationResult{
		Valid:        valid,
		Signature:    signature,
		VerifiedAt:   time.Now(),
		ExpectedHash: expectedSignature,
		ActualHash:   signature,
	}

	if valid {
		result.Message = "Signature is valid - approval is authentic and unmodified"
	} else {
		result.Message = "Signature mismatch - data may have been tampered with"
	}

	return result
}

// GenerateQRCodeData creates data string for QR code generation
func GenerateQRCodeData(taskID, signature string, verificationURL string) string {
	return fmt.Sprintf("%s/verify?task=%s&sig=%s", verificationURL, taskID, signature)
}
