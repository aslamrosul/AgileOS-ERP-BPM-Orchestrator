package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type AccountingSettings struct {
	ID                    string    `json:"id"`
	CompanyName           string    `json:"company_name"`
	CompanyAddress        string    `json:"company_address"`
	CompanyPhone          string    `json:"company_phone"`
	CompanyEmail          string    `json:"company_email"`
	TaxID                 string    `json:"tax_id"`
	FiscalYearStart       string    `json:"fiscal_year_start"`
	FiscalYearEnd         string    `json:"fiscal_year_end"`
	BaseCurrency          string    `json:"base_currency"`
	DateFormat            string    `json:"date_format"`
	NumberFormat          string    `json:"number_format"`
	EnableMultiCurrency   bool      `json:"enable_multi_currency"`
	EnableInventory       bool      `json:"enable_inventory"`
	EnableProjects        bool      `json:"enable_projects"`
	EnableTimeTracking    bool      `json:"enable_time_tracking"`
	EnableExpenseTracking bool      `json:"enable_expense_tracking"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

func (h *AccountingHandler) GetSettings(c *gin.Context) {
	// Query settings from database
	query := `SELECT * FROM accounting_settings LIMIT 1`
	result, err := h.db.Query(query, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch settings"})
		return
	}

	var settings []AccountingSettings
	if err := result.Unmarshal(&settings); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse settings"})
		return
	}

	// If no settings exist, return defaults
	if len(settings) == 0 {
		defaultSettings := AccountingSettings{
			ID:                    "settings:default",
			CompanyName:           "Your Company Name",
			BaseCurrency:          "IDR",
			DateFormat:            "DD/MM/YYYY",
			NumberFormat:          "1.234.567,89",
			FiscalYearStart:       "01-01",
			FiscalYearEnd:         "12-31",
			EnableMultiCurrency:   false,
			EnableInventory:       false,
			EnableProjects:        false,
			EnableTimeTracking:    false,
			EnableExpenseTracking: false,
			CreatedAt:             time.Now(),
			UpdatedAt:             time.Now(),
		}
		c.JSON(http.StatusOK, defaultSettings)
		return
	}

	c.JSON(http.StatusOK, settings[0])
}

func (h *AccountingHandler) SaveSettings(c *gin.Context) {
	var settings AccountingSettings
	if err := c.ShouldBindJSON(&settings); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if settings exist
	checkQuery := `SELECT id FROM accounting_settings LIMIT 1`
	checkResult, err := h.db.Query(checkQuery, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check settings"})
		return
	}

	var existing []map[string]interface{}
	checkResult.Unmarshal(&existing)

	now := time.Now()
	settings.UpdatedAt = now

	var query string
	var params map[string]interface{}

	if len(existing) == 0 {
		// Create new settings
		settings.ID = "settings:default"
		settings.CreatedAt = now
		query = `CREATE accounting_settings CONTENT $settings`
		params = map[string]interface{}{
			"settings": settings,
		}
	} else {
		// Update existing settings
		settings.ID = existing[0]["id"].(string)
		query = `UPDATE $id MERGE $settings`
		params = map[string]interface{}{
			"id":       settings.ID,
			"settings": map[string]interface{}{
				"company_name":            settings.CompanyName,
				"company_address":         settings.CompanyAddress,
				"company_phone":           settings.CompanyPhone,
				"company_email":           settings.CompanyEmail,
				"tax_id":                  settings.TaxID,
				"fiscal_year_start":       settings.FiscalYearStart,
				"fiscal_year_end":         settings.FiscalYearEnd,
				"base_currency":           settings.BaseCurrency,
				"date_format":             settings.DateFormat,
				"number_format":           settings.NumberFormat,
				"enable_multi_currency":   settings.EnableMultiCurrency,
				"enable_inventory":        settings.EnableInventory,
				"enable_projects":         settings.EnableProjects,
				"enable_time_tracking":    settings.EnableTimeTracking,
				"enable_expense_tracking": settings.EnableExpenseTracking,
				"updated_at":              now,
			},
		}
	}

	result, err := h.db.Query(query, params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save settings"})
		return
	}

	var savedSettings []AccountingSettings
	if err := result.Unmarshal(&savedSettings); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse saved settings"})
		return
	}

	c.JSON(http.StatusOK, savedSettings[0])
}
