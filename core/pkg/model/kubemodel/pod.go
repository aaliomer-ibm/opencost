package kubemodel

import (
	"time"

	"github.com/google/uuid"
)

type Pod struct {
	UID                  uuid.UUID         `json:"uid"`                   // @bingen:field[version=1]
	NamespaceUID         uuid.UUID         `json:"namespaceUid"`          // @bingen:field[version=1]
	OwnerUID             uuid.UUID         `json:"ownerUid"`              // @bingen:field[version=1] - Reference to Owner (Deployment, StatefulSet, etc.)
	NodeUID              uuid.UUID         `json:"nodeUid"`               // @bingen:field[version=1]
	Name                 string            `json:"name"`                  // @bingen:field[version=1]
	Labels               map[string]string `json:"labels,omitempty"`      // @bingen:field[version=1]
	Annotations          map[string]string `json:"annotations,omitempty"` // @bingen:field[version=1]
	DurationSeconds      uint64            `json:"durationSeconds"`       // @bingen:field[version=1]
	NetworkTransferBytes uint64            `json:"networkTransferBytes"`  // @bingen:field[version=1]
	NetworkReceiveBytes  uint64            `json:"networkReceiveBytes"`   // @bingen:field[version=1]
	Start                time.Time         `json:"start,omitempty"`       // @bingen:field[version=1] - Pod creation/start timestamp
	End                  time.Time         `json:"end,omitempty"`         // @bingen:field[version=1] - Pod deletion/end timestamp (nil if still running)
}
