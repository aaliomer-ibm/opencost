package synthetic

import (
	"time"

	"github.com/opencost/opencost/core/pkg/model/pricingmodel"
	"github.com/opencost/opencost/core/pkg/model/shared"
	"github.com/opencost/opencost/core/pkg/opencost"
)

func CreatePricingModelFromAssets(
	assetSets []*opencost.AssetSet,
	providerStr string,
	region string,
	timestamp time.Time,
) *pricingmodel.PricingModelSet {
	provider := shared.ParseProvider(providerStr)
	pms := pricingmodel.NewPricingModelSet(timestamp, "synthetic", "synthetic")

	type nodeAccumulator struct {
		cpuCoreHours, ramByteHours, gpuHours float64
		cpuCost, ramCost, gpuCost, gpuCount  float64
		preemptible                          bool
	}

	byNodeType := make(map[string]*nodeAccumulator)

	for _, as := range assetSets {
		for _, asset := range as.Assets {
			node, ok := asset.(*opencost.Node)
			if !ok {
				continue
			}
			nodeType := node.NodeType
			if nodeType == "" {
				nodeType = "unknown"
			}
			acc, exists := byNodeType[nodeType]
			if !exists {
				acc = &nodeAccumulator{}
				byNodeType[nodeType] = acc
			}
			acc.cpuCoreHours += node.CPUCoreHours
			acc.ramByteHours += node.RAMByteHours
			acc.gpuHours += node.GPUHours
			acc.cpuCost += node.CPUCost
			acc.ramCost += node.RAMCost
			acc.gpuCost += node.GPUCost
			acc.gpuCount += node.GPUCount
			if node.Preemptible > 0 {
				acc.preemptible = true
			}
		}
	}

	for nodeType, acc := range byNodeType {
		usageType := shared.UsageTypeOnDemand
		if acc.preemptible {
			usageType = shared.UsageTypeSpot
		}

		if acc.cpuCoreHours > 0 {
			pms.NodePricing[pricingmodel.NodeKey{
				Provider: provider, PricingType: pricingmodel.NodePricingTypeCPUCore,
				UsageType: usageType, Region: region, NodeType: nodeType,
			}] = pricingmodel.NodePricing{HourlyRate: acc.cpuCost / acc.cpuCoreHours}
		}

		ramGiBHours := acc.ramByteHours / (1024 * 1024 * 1024)
		if ramGiBHours > 0 {
			pms.NodePricing[pricingmodel.NodeKey{
				Provider: provider, PricingType: pricingmodel.NodePricingTypeRamGB,
				UsageType: usageType, Region: region, NodeType: nodeType,
			}] = pricingmodel.NodePricing{HourlyRate: acc.ramCost / ramGiBHours}
		}

		if acc.gpuHours > 0 {
			pms.NodePricing[pricingmodel.NodeKey{
				Provider: provider, PricingType: pricingmodel.NodePricingTypeDevice,
				UsageType: usageType, Region: region, NodeType: nodeType,
			}] = pricingmodel.NodePricing{HourlyRate: acc.gpuCost / acc.gpuHours}
		}

		totalCost := acc.cpuCost + acc.ramCost + acc.gpuCost
		if acc.cpuCoreHours > 0 {
			pms.NodePricing[pricingmodel.NodeKey{
				Provider: provider, PricingType: pricingmodel.NodePricingTypeTotal,
				UsageType: usageType, Region: region, NodeType: nodeType,
			}] = pricingmodel.NodePricing{HourlyRate: totalCost}
		}
	}

	return pms
}

type PricingModelSetJSON struct {
	TimeStamp   time.Time              `json:"timestamp"`
	SourceType  string                 `json:"sourceType"`
	SourceKey   string                 `json:"sourceKey"`
	NodePricing []NodePricingEntryJSON `json:"nodePricing"`
}

type NodePricingEntryJSON struct {
	Provider    string  `json:"provider"`
	PricingType string  `json:"pricingType"`
	UsageType   string  `json:"usageType"`
	Region      string  `json:"region"`
	NodeType    string  `json:"nodeType"`
	Family      string  `json:"family,omitempty"`
	DeviceType  string  `json:"deviceType,omitempty"`
	HourlyRate  float64 `json:"hourlyRate"`
}

func ConvertPricingModelToJSON(pms *pricingmodel.PricingModelSet) *PricingModelSetJSON {
	entries := make([]NodePricingEntryJSON, 0, len(pms.NodePricing))
	for k, v := range pms.NodePricing {
		entries = append(entries, NodePricingEntryJSON{
			Provider: string(k.Provider), PricingType: string(k.PricingType),
			UsageType: string(k.UsageType), Region: k.Region, NodeType: k.NodeType,
			Family: k.Family, DeviceType: k.DeviceType, HourlyRate: v.HourlyRate,
		})
	}
	return &PricingModelSetJSON{
		TimeStamp: pms.TimeStamp, SourceType: string(pms.SourceType),
		SourceKey: pms.SourceKey, NodePricing: entries,
	}
}
