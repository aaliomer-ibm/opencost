//nolint:stylecheck // generated code and package conventions
package kubemodel

import (
	"time"
)

// @bingen:generate:ServiceType
type ServiceType string

const (
	ServiceTypeClusterIP    ServiceType = "ClusterIP"
	ServiceTypeNodePort     ServiceType = "NodePort"
	ServiceTypeLoadBalancer ServiceType = "LoadBalancer"
	ServiceTypeExternalName ServiceType = "ExternalName"
)

// @bingen:generate:ServicePort
type ServicePort struct {
	Name       string `json:"name"`       // @bingen:field[version=1]
	Port       uint16 `json:"port"`       // @bingen:field[version=1]
	TargetPort uint16 `json:"targetPort"` // @bingen:field[version=1]
	NodePort   uint16 `json:"nodePort"`   // @bingen:field[version=1]
	Protocol   string `json:"protocol"`   // @bingen:field[version=1]
}

// @bingen:generate:Service
// Service represents a Kubernetes Service with network traffic tracking for cost allocation.
//
// Network Cost Allocation Strategy:
// Services expose applications and route traffic, incurring costs for:
// 1. Load Balancers (LoadBalancer type) - Cloud provider LB hourly cost + data transfer
// 2. Data Transfer - Egress charges based on NetworkTransferBytes
// 3. Public IPs (for LoadBalancer/NodePort with external IPs)
//
// Cost Attribution Flow:
// - LoadBalancer Services: Direct cloud resource cost (e.g., AWS ELB, GCP LB) allocated to service
// - Data Transfer: NetworkTransferBytes × cloud provider egress rate (varies by region/destination)
// - NetworkReceiveBytes: Typically free (ingress), tracked for visibility
// - Use Selector to map service costs to backing pods/containers proportionally
//
// Example: AWS Application Load Balancer
// - Fixed hourly cost: $0.0225/hour
// - LCU cost: $0.008/hour per LCU (based on connections, requests, bandwidth)
// - Data transfer: $0.09/GB for internet egress
// Total Service Cost = (LB hours × hourly rate) + (LCU hours × LCU rate) + (NetworkTransferBytes × transfer rate)
type Service struct {
	UID                  string            `json:"uid"`                   // @bingen:field[version=1]
	ClusterUID           string            `json:"clusterUid"`            // @bingen:field[version=1]
	NamespaceUID         string            `json:"namespaceUid"`          // @bingen:field[version=1]
	Name                 string            `json:"name"`                  // @bingen:field[version=1]
	Type                 ServiceType       `json:"type"`                  // @bingen:field[version=1]
	Hostname             string            `json:"hostname,omitempty"`    // @bingen:field[version=1]
	Labels               map[string]string `json:"labels,omitempty"`      // @bingen:field[version=1]
	Annotations          map[string]string `json:"annotations,omitempty"` // @bingen:field[version=1]
	Ports                []ServicePort     `json:"ports,omitempty"`       // @bingen:field[version=1]
	Start                time.Time         `json:"start"`                 // @bingen:field[version=1]
	End                  time.Time         `json:"end"`                   // @bingen:field[version=1]
	NetworkTransferBytes uint64            `json:"networkTransferBytes"`  // @bingen:field[version=1]
	NetworkReceiveBytes  uint64            `json:"networkReceiveBytes"`   // @bingen:field[version=1]
	// Label selector to identify pods/containers targeted by this service
	// Maps label keys to values (e.g., {"app": "nginx", "tier": "frontend"})
	// Pods with matching labels will receive traffic from this service
	Selector map[string]string `json:"selector,omitempty"` // @bingen:field[version=1]
	// Lifecycle tracking
	DurationSeconds uint64 `json:"durationSeconds"` // @bingen:field[version=1] - Duration service existed within measurement window
	// Version 2 fields - Network breakdown by destination type
	NetworkInternetEgressBytes uint64 `json:"networkInternetEgressBytes,omitempty"` // @bingen:field[version=1]
	NetworkCrossRegionBytes    uint64 `json:"networkCrossRegionBytes,omitempty"`    // @bingen:field[version=1]
	NetworkSameRegionBytes     uint64 `json:"networkSameRegionBytes,omitempty"`     // @bingen:field[version=1]
	NetworkIntraAZBytes        uint64 `json:"networkIntraAzBytes,omitempty"`        // @bingen:field[version=1]
}
