//nolint:stylecheck
package kubemodel

import (
	"time"

	"github.com/google/uuid"
)

// @bingen:generate:Namespace
type Namespace struct {
	UID         uuid.UUID         `json:"uid"`             // @bingen:field[version=1]
	Name        string            `json:"name"`            // @bingen:field[version=1]
	Labels      map[string]string `json:"labels"`          // @bingen:field[version=1]
	Annotations map[string]string `json:"annotations"`     // @bingen:field[version=1]
	Start       time.Time         `json:"start,omitempty"` // @bingen:field[version=1]
	End         time.Time         `json:"end,omitempty"`   // @bingen:field[version=1]
}
