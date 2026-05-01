package handlers

import (
	"agileos-backend/database"
)

// AccountingHandler handles accounting-related requests
type AccountingHandler struct {
	db *database.SurrealDB
}

// NewAccountingHandler creates a new accounting handler
func NewAccountingHandler(db *database.SurrealDB) *AccountingHandler {
	return &AccountingHandler{db: db}
}
