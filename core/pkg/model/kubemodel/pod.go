package kubemodel

import "time"

type Pod struct {
	UID                  string            `json:"uid"`                   // @bingen:field[version=1]
	NamespaceUID         string            `json:"namespaceUid"`          // @bingen:field[version=1]
	OwnerUID             string            `json:"ownerUid"`              // @bingen:field[version=1] - Reference to Owner (Deployment, StatefulSet, etc.)
	NodeUID              string            `json:"nodeUid"`               // @bingen:field[version=1]
	Name                 string            `json:"name"`                  // @bingen:field[version=1]
	Labels               map[string]string `json:"labels,omitempty"`      // @bingen:field[version=1]
	Annotations          map[string]string `json:"annotations,omitempty"` // @bingen:field[version=1]
	DurationSeconds      Measurement       `json:"durationSeconds"`       // @bingen:field[version=1]
	NetworkTransferBytes Measurement       `json:"networkTransferBytes"`  // @bingen:field[version=1]
	NetworkReceiveBytes  Measurement       `json:"networkReceiveBytes"`   // @bingen:field[version=1]
	Start                time.Time         `json:"start,omitempty"`       // @bingen:field[version=1] - Pod creation/start timestamp
	End                  time.Time         `json:"end,omitempty"`         // @bingen:field[version=1] - Pod deletion/end timestamp (nil if still running)
}
