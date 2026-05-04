package synthetic

import (
	"fmt"
	"io"
	"strconv"

	"github.com/opencost/opencost/core/pkg/clustercache"
	"github.com/opencost/opencost/pkg/cloud/models"

	synthspec "github.com/opencost/opencost/modules/collector-source/pkg/synthetic"
)

type SyntheticProvider struct {
	spec              synthspec.ClusterSpec
	nodesByProviderID map[string]synthspec.NodeSpec
	lbCostPerHr       float64
}

func NewSyntheticProvider(spec synthspec.ClusterSpec) *SyntheticProvider {
	byPID := make(map[string]synthspec.NodeSpec, len(spec.Nodes))
	for _, n := range spec.Nodes {
		byPID[n.ProviderID] = n
	}

	var lbTotal float64
	var lbCount int
	for _, svc := range spec.Services {
		if svc.Type == "LoadBalancer" && svc.CostPerHr > 0 {
			lbTotal += svc.CostPerHr
			lbCount++
		}
	}
	lbCost := 0.025
	if lbCount > 0 {
		lbCost = lbTotal / float64(lbCount)
	}

	return &SyntheticProvider{spec: spec, nodesByProviderID: byPID, lbCostPerHr: lbCost}
}

func (p *SyntheticProvider) ClusterInfo() (map[string]string, error) {
	return map[string]string{"name": p.spec.ClusterName, "id": p.spec.ClusterUID, "provider": p.spec.Provider, "region": p.spec.Region}, nil
}

// Stub methods required by Provider interface
func (p *SyntheticProvider) GetAddresses() ([]byte, error) { return []byte("[]"), nil }
func (p *SyntheticProvider) GetDisks() ([]byte, error)     { return []byte("[]"), nil }
func (p *SyntheticProvider) GetOrphanedResources() ([]models.OrphanedResource, error) {
	return nil, nil
}
func (p *SyntheticProvider) AllNodePricing() (any, error)                         { return nil, nil }
func (p *SyntheticProvider) DownloadPricingData() error                           { return nil }
func (p *SyntheticProvider) GetManagementPlatform() (string, error)               { return "", nil }
func (p *SyntheticProvider) ApplyReservedInstancePricing(map[string]*models.Node) {}
func (p *SyntheticProvider) Regions() []string                                    { return []string{p.spec.Region} }
func (p *SyntheticProvider) PricingSourceSummary() any                            { return nil }
func (p *SyntheticProvider) ServiceAccountStatus() *models.ServiceAccountStatus {
	return &models.ServiceAccountStatus{}
}
func (p *SyntheticProvider) PricingSourceStatus() map[string]*models.PricingSource {
	return map[string]*models.PricingSource{}
}
func (p *SyntheticProvider) ClusterManagementPricing() (string, float64, error)             { return "", 0, nil }
func (p *SyntheticProvider) CombinedDiscountForNode(string, bool, float64, float64) float64 { return 0 }
func (p *SyntheticProvider) GpuPricing(labels map[string]string) (string, error)            { return "", nil }

func (p *SyntheticProvider) NodePricing(key models.Key) (*models.Node, models.PricingMetadata, error) {
	n, ok := p.nodesByProviderID[key.ID()]
	if !ok {
		return nil, models.PricingMetadata{}, fmt.Errorf("no pricing for node %s", key.ID())
	}
	usageType := "on-demand"
	if n.IsSpot {
		usageType = "spot"
	}
	ramGiB := n.RAMBytes / 1024 / 1024 / 1024
	totalCost := n.CPUCostPerHr*n.CPUCores + n.RAMCostPerGiBHr*ramGiB + n.GPUCostPerHr*n.GPUCount
	node := &models.Node{
		Cost:         strconv.FormatFloat(totalCost, 'f', 6, 64),
		VCPU:         strconv.FormatFloat(n.CPUCores, 'f', 0, 64),
		VCPUCost:     strconv.FormatFloat(n.CPUCostPerHr, 'f', 6, 64),
		RAM:          strconv.FormatFloat(ramGiB, 'f', 0, 64),
		RAMBytes:     strconv.FormatFloat(n.RAMBytes, 'f', 0, 64),
		RAMCost:      strconv.FormatFloat(n.RAMCostPerGiBHr, 'f', 6, 64),
		UsageType:    usageType,
		GPU:          strconv.FormatFloat(n.GPUCount, 'f', 0, 64),
		GPUName:      n.GPUType,
		GPUCost:      strconv.FormatFloat(n.GPUCostPerHr, 'f', 6, 64),
		InstanceType: n.InstanceType,
		Region:       p.spec.Region,
		ProviderID:   n.ProviderID,
	}
	return node, models.PricingMetadata{Currency: "USD", Source: "synthetic"}, nil
}

func (p *SyntheticProvider) PVPricing(key models.PVKey) (*models.PV, error) {
	for _, pv := range p.spec.PVs {
		if pv.ProviderID == key.ID() || pv.StorageClass == key.GetStorageClass() {
			return &models.PV{
				Cost:       strconv.FormatFloat(pv.CostPerGiBHr, 'f', 8, 64),
				Class:      pv.StorageClass,
				Size:       strconv.FormatFloat(pv.CapacityBytes/1024/1024/1024, 'f', 0, 64),
				Region:     p.spec.Region,
				ProviderID: pv.ProviderID,
			}, nil
		}
	}
	return &models.PV{Cost: "0.00005"}, nil
}

func (p *SyntheticProvider) NetworkPricing() (*models.Network, error) {
	return &models.Network{ZoneNetworkEgressCost: 0.01, RegionNetworkEgressCost: 0.01, InternetNetworkEgressCost: 0.12, NatGatewayEgressCost: 0.045, NatGatewayIngressCost: 0.045}, nil
}

func (p *SyntheticProvider) LoadBalancerPricing() (*models.LoadBalancer, error) {
	return &models.LoadBalancer{Cost: p.lbCostPerHr}, nil
}

func (p *SyntheticProvider) GetKey(labels map[string]string, _ *clustercache.Node) models.Key {
	return &syntheticKey{providerID: labels["provider_id"]}
}

func (p *SyntheticProvider) GetPVKey(pv *clustercache.PersistentVolume, _ map[string]string, _ string) models.PVKey {
	providerID := pv.Name
	if pv.Spec.CSI != nil && pv.Spec.CSI.VolumeHandle != "" {
		providerID = pv.Spec.CSI.VolumeHandle
	}
	return &syntheticPVKey{providerID: providerID, storageClass: pv.Spec.StorageClassName}
}

func (p *SyntheticProvider) UpdateConfig(_ io.Reader, _ string) (*models.CustomPricing, error) {
	return p.GetConfig()
}
func (p *SyntheticProvider) UpdateConfigFromConfigMap(_ map[string]string) (*models.CustomPricing, error) {
	return p.GetConfig()
}
func (p *SyntheticProvider) GetConfig() (*models.CustomPricing, error) {
	return &models.CustomPricing{Provider: p.spec.Provider}, nil
}

type syntheticKey struct{ providerID string }

func (k *syntheticKey) ID() string       { return k.providerID }
func (k *syntheticKey) Features() string { return "" }
func (k *syntheticKey) GPUType() string  { return "" }
func (k *syntheticKey) GPUCount() int    { return 0 }

type syntheticPVKey struct{ providerID, storageClass string }

func (k *syntheticPVKey) ID() string              { return k.providerID }
func (k *syntheticPVKey) Features() string        { return "" }
func (k *syntheticPVKey) GetStorageClass() string { return k.storageClass }
