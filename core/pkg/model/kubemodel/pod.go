package kubemodel

import "time"

type Pod struct {
	UID                  string            `json:"uid"`
	NamespaceUID         string            `json:"namespaceUid"`
	OwnerUID             string            `json:"ownerUid"` // Reference to Owner (Deployment, StatefulSet, etc.)
	NodeUID              string            `json:"nodeUid"`
	Name                 string            `json:"name"`
	Labels               map[string]string `json:"labels,omitempty"`
	Annotations          map[string]string `json:"annotations,omitempty"`
	DurationSeconds      Measurement       `json:"durationSeconds"`
	NetworkTransferBytes Measurement       `json:"networkTransferBytes"`
	NetworkReceiveBytes  Measurement       `json:"networkReceiveBytes"`
	Start                time.Time         `json:"start,omitempty"` // Pod creation/start timestamp
	End                  time.Time         `json:"end,omitempty"`   // Pod deletion/end timestamp (nil if still running)
}
