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

// CRMHandler handles CRM operations
type CRMHandler struct {
	db *database.SurrealDB
}

// NewCRMHandler creates a new CRM handler
func NewCRMHandler(db *database.SurrealDB) *CRMHandler {
	return &CRMHandler{db: db}
}

// ============================================
// CONTACT MANAGEMENT HANDLERS
// ============================================

// CreateContact creates a new contact
func (h *CRMHandler) CreateContact(c *gin.Context) {
	var contact models.Contact
	if err := c.ShouldBindJSON(&contact); err != nil {
		logger.LogError("Failed to bind contact data", err, nil)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Auto-generate contact code
	contacts, err := h.db.QuerySlice(
		"SELECT contact_code FROM contact ORDER BY contact_code DESC LIMIT 1",
		nil,
	)
	if err != nil {
		logger.LogError("Failed to get last contact code", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate contact code"})
		return
	}

	contactCode := "CON-0001"
	if len(contacts) > 0 {
		lastCode := contacts[0].(map[string]interface{})["contact_code"].(string)
		var lastNum int
		fmt.Sscanf(lastCode, "CON-%d", &lastNum)
		contactCode = fmt.Sprintf("CON-%04d", lastNum+1)
	}

	contact.ContactCode = contactCode
	contact.FullName = contact.FirstName + " " + contact.LastName
	contact.CreatedBy = userID.(string)
	contact.CreatedAt = time.Now()
	contact.UpdatedAt = time.Now()
	contact.IsActive = true

	query := `CREATE contact CONTENT {
		contact_code: $contact_code,
		first_name: $first_name,
		last_name: $last_name,
		full_name: $full_name,
		email: $email,
		phone: $phone,
		mobile: $mobile,
		job_title: $job_title,
		department: $department,
		company: $company,
		company_id: $company_id,
		address: $address,
		city: $city,
		state: $state,
		country: $country,
		postal_code: $postal_code,
		website: $website,
		linkedin: $linkedin,
		twitter: $twitter,
		contact_type: $contact_type,
		source: $source,
		notes: $notes,
		is_active: $is_active,
		created_by: $created_by,
		created_at: $created_at,
		updated_at: $updated_at
	}`

	params := map[string]interface{}{
		"contact_code": contact.ContactCode,
		"first_name":   contact.FirstName,
		"last_name":    contact.LastName,
		"full_name":    contact.FullName,
		"email":        contact.Email,
		"phone":        contact.Phone,
		"mobile":       contact.Mobile,
		"job_title":    contact.JobTitle,
		"department":   contact.Department,
		"company":      contact.Company,
		"company_id":   contact.CompanyID,
		"address":      contact.Address,
		"city":         contact.City,
		"state":        contact.State,
		"country":      contact.Country,
		"postal_code":  contact.PostalCode,
		"website":      contact.Website,
		"linkedin":     contact.LinkedIn,
		"twitter":      contact.Twitter,
		"contact_type": contact.ContactType,
		"source":       contact.Source,
		"notes":        contact.Notes,
		"is_active":    contact.IsActive,
		"created_by":   contact.CreatedBy,
		"created_at":   contact.CreatedAt,
		"updated_at":   contact.UpdatedAt,
	}

	result, err := h.db.QuerySlice(query, params)
	if err != nil {
		logger.LogError("Failed to create contact", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create contact"})
		return
	}

	logger.Log.Info().
		Str("contact_code", contact.ContactCode).
		Str("full_name", contact.FullName).
		Msg("Contact created successfully")

	c.JSON(http.StatusCreated, result[0])
}

// GetContacts retrieves all contacts with filters
func (h *CRMHandler) GetContacts(c *gin.Context) {
	contactType := c.Query("contact_type")
	company := c.Query("company")
	isActive := c.Query("is_active")

	query := "SELECT * FROM contact"
	params := make(map[string]interface{})

	var conditions []string
	if contactType != "" {
		conditions = append(conditions, "contact_type = $contact_type")
		params["contact_type"] = contactType
	}
	if company != "" {
		conditions = append(conditions, "company = $company")
		params["company"] = company
	}
	if isActive != "" {
		conditions = append(conditions, "is_active = $is_active")
		params["is_active"] = isActive == "true"
	}

	if len(conditions) > 0 {
		query += " WHERE " + conditions[0]
		for i := 1; i < len(conditions); i++ {
			query += " AND " + conditions[i]
		}
	}

	query += " ORDER BY contact_code ASC"

	contacts, err := h.db.QuerySlice(query, params)
	if err != nil {
		logger.LogError("Failed to get contacts", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve contacts"})
		return
	}

	c.JSON(http.StatusOK, contacts)
}

// GetContact retrieves a contact by ID
func (h *CRMHandler) GetContact(c *gin.Context) {
	contactID := c.Param("id")

	contacts, err := h.db.QuerySlice(
		"SELECT * FROM $id",
		map[string]interface{}{"id": contactID},
	)
	if err != nil || len(contacts) == 0 {
		logger.LogError("Contact not found", err, map[string]interface{}{"contact_id": contactID})
		c.JSON(http.StatusNotFound, gin.H{"error": "Contact not found"})
		return
	}

	c.JSON(http.StatusOK, contacts[0])
}

// UpdateContact updates an existing contact
func (h *CRMHandler) UpdateContact(c *gin.Context) {
	contactID := c.Param("id")

	var contact models.Contact
	if err := c.ShouldBindJSON(&contact); err != nil {
		logger.LogError("Failed to bind contact data", err, nil)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	contact.FullName = contact.FirstName + " " + contact.LastName
	contact.UpdatedAt = time.Now()

	query := `UPDATE $id SET
		first_name = $first_name,
		last_name = $last_name,
		full_name = $full_name,
		email = $email,
		phone = $phone,
		mobile = $mobile,
		job_title = $job_title,
		department = $department,
		company = $company,
		company_id = $company_id,
		address = $address,
		city = $city,
		state = $state,
		country = $country,
		postal_code = $postal_code,
		website = $website,
		linkedin = $linkedin,
		twitter = $twitter,
		contact_type = $contact_type,
		source = $source,
		notes = $notes,
		is_active = $is_active,
		updated_at = $updated_at
	`

	params := map[string]interface{}{
		"id":           contactID,
		"first_name":   contact.FirstName,
		"last_name":    contact.LastName,
		"full_name":    contact.FullName,
		"email":        contact.Email,
		"phone":        contact.Phone,
		"mobile":       contact.Mobile,
		"job_title":    contact.JobTitle,
		"department":   contact.Department,
		"company":      contact.Company,
		"company_id":   contact.CompanyID,
		"address":      contact.Address,
		"city":         contact.City,
		"state":        contact.State,
		"country":      contact.Country,
		"postal_code":  contact.PostalCode,
		"website":      contact.Website,
		"linkedin":     contact.LinkedIn,
		"twitter":      contact.Twitter,
		"contact_type": contact.ContactType,
		"source":       contact.Source,
		"notes":        contact.Notes,
		"is_active":    contact.IsActive,
		"updated_at":   contact.UpdatedAt,
	}

	result, err := h.db.QuerySlice(query, params)
	if err != nil || len(result) == 0 {
		logger.LogError("Failed to update contact", err, map[string]interface{}{"contact_id": contactID})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update contact"})
		return
	}

	logger.Log.Info().
		Str("contact_id", contactID).
		Str("full_name", contact.FullName).
		Msg("Contact updated successfully")

	c.JSON(http.StatusOK, result[0])
}

// DeleteContact soft deletes a contact
func (h *CRMHandler) DeleteContact(c *gin.Context) {
	contactID := c.Param("id")

	query := `UPDATE $id SET is_active = false, updated_at = $updated_at`
	params := map[string]interface{}{
		"id":         contactID,
		"updated_at": time.Now(),
	}

	result, err := h.db.QuerySlice(query, params)
	if err != nil || len(result) == 0 {
		logger.LogError("Failed to delete contact", err, map[string]interface{}{"contact_id": contactID})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete contact"})
		return
	}

	logger.Log.Info().
		Str("contact_id", contactID).
		Msg("Contact deleted successfully")

	c.JSON(http.StatusOK, gin.H{"message": "Contact deleted successfully"})
}
