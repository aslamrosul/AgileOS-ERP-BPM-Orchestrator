package database

import (
	"fmt"
	"time"

	"agileos-backend/models"
)

// CreateDocument creates a new document
func (s *SurrealDB) CreateDocument(doc *models.Document) error {
	doc.CreatedAt = time.Now()
	doc.UpdatedAt = time.Now()
	
	if doc.Status == "" {
		doc.Status = models.DocumentStatusDraft
	}

	query := `CREATE document CONTENT $doc`

	var created []models.Document
	if err := s.queryAndUnmarshal(query, map[string]interface{}{"doc": doc}, &created); err != nil {
		return fmt.Errorf("failed to create document: %w", err)
	}

	if len(created) > 0 {
		doc.ID = created[0].ID
		return nil
	}

	return fmt.Errorf("document created but ID extraction failed")
}

// GetDocument retrieves a document by ID
func (s *SurrealDB) GetDocument(documentID string) (*models.Document, error) {
	query := `SELECT * FROM $document`

	var documents []models.Document
	if err := s.queryAndUnmarshal(query, map[string]interface{}{"document": documentID}, &documents); err != nil {
		return nil, fmt.Errorf("failed to get document: %w", err)
	}

	if len(documents) == 0 {
		return nil, fmt.Errorf("document not found")
	}

	return &documents[0], nil
}
