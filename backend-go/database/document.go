package database

import (
	"fmt"
	"time"

	"agileos-backend/models"

	"github.com/surrealdb/surrealdb.go"
)

// CreateDocument creates a new document
func (s *SurrealDB) CreateDocument(doc *models.Document) error {
	doc.CreatedAt = time.Now()
	doc.UpdatedAt = time.Now()
	
	if doc.Status == "" {
		doc.Status = models.DocumentStatusDraft
	}

	query := `CREATE document CONTENT $doc`

	result, err := s.client.Query(query, map[string]interface{}{
		"doc": doc,
	})
	if err != nil {
		return fmt.Errorf("failed to create document: %w", err)
	}

	// Extract ID
	if resultArray, ok := result.([]interface{}); ok && len(resultArray) > 0 {
		if outerMap, ok := resultArray[0].(map[string]interface{}); ok {
			if resultField, ok := outerMap["result"].([]interface{}); ok && len(resultField) > 0 {
				if innerMap, ok := resultField[0].(map[string]interface{}); ok {
					if id, ok := innerMap["id"].(string); ok {
						doc.ID = id
						return nil
					}
				}
			}
		}
	}

	return fmt.Errorf("document created but ID extraction failed")
}

// GetDocument retrieves a document by ID
func (s *SurrealDB) GetDocument(documentID string) (*models.Document, error) {
	query := `SELECT * FROM $document`

	result, err := s.client.Query(query, map[string]interface{}{
		"document": documentID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get document: %w", err)
	}

	var documents []models.Document
	if err := surrealdb.Unmarshal(result, &documents); err != nil {
		return nil, fmt.Errorf("failed to unmarshal document: %w", err)
	}

	if len(documents) == 0 {
		return nil, fmt.Errorf("document not found")
	}

	return &documents[0], nil
}