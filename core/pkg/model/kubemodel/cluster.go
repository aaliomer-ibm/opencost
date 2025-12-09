//nolint:stylecheck
package kubemodel

import "time"

// @bingen:generate:Cluster
type Cluster struct {
	UID          string    `json:"uid"`             // @bingen:field[version=1]
	Provider     Provider  `json:"provider"`        // @bingen:field[version=1]
	Account      string    `json:"account"`         // @bingen:field[version=1]
	Name         string    `json:"name"`            // @bingen:field[version=1]
	Agent        string    `json:"agent"`           // @bingen:field[version=1]
	AgentVersion string    `json:"agentVersion"`    // @bingen:field[version=1]
	K8sVersion   string    `json:"version"`         // @bingen:field[version=1]
	ClusterType  string    `json:"type"`            // @bingen:field[version=1]
	Region       string    `json:"region"`          // @bingen:field[version=1]
	Start        time.Time `json:"start,omitempty"` // @bingen:field[version=1]
	End          time.Time `json:"end,omitempty"`   // @bingen:field[version=1]
}
