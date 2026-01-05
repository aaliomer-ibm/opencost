package kubemodel

import (
	"time"

	"github.com/google/uuid"
)

// @bingen:generate:DiagnosticResult
type DiagnosticResult struct {
	UID         uuid.UUID         `json:"uid"`               // @bingen:field[version=1]
	Name        string            `json:"name"`              // @bingen:field[version=1]
	Description string            `json:"description"`       // @bingen:field[version=1]
	Category    string            `json:"category"`          // @bingen:field[version=1]
	Timestamp   time.Time         `json:"timestamp"`         // @bingen:field[version=1]
	Error       string            `json:"error,omitempty"`   // @bingen:field[version=1]
	Details     map[string]string `json:"details,omitempty"` // @bingen:field[version=1]
}
