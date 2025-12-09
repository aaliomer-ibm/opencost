//nolint:stylecheck // generated code and package conventions
package kubemodel

import "time"

// @bingen:generate:Pod
type Pod struct {
	UID                  string            `json:"uid"`                   // @bingen:field[version=1]
	NamespaceUID         string            `json:"namespaceUid"`          // @bingen:field[version=1]
	OwnerUID             string            `json:"ownerUid"`              // @bingen:field[version=1] - Reference to Owner (Deployment, StatefulSet, etc.)
	NodeUID              string            `json:"nodeUid"`               // @bingen:field[version=1]
	Name                 string            `json:"name"`                  // @bingen:field[version=1]
	Labels               map[string]string `json:"labels,omitempty"`      // @bingen:field[version=1]
	Annotations          map[string]string `json:"annotations,omitempty"` // @bingen:field[version=1]
	DurationSeconds      uint64            `json:"durationSeconds"`       // @bingen:field[version=1]
	NetworkTransferBytes uint64            `json:"networkTransferBytes"`  // @bingen:field[version=1]
	NetworkReceiveBytes  uint64            `json:"networkReceiveBytes"`   // @bingen:field[version=1]
	Start                time.Time         `json:"start,omitempty"`       // @bingen:field[version=1] - Pod creation/start timestamp
	End                  time.Time         `json:"end,omitempty"`         // @bingen:field[version=1] - Pod deletion/end timestamp (nil if still running)
}
