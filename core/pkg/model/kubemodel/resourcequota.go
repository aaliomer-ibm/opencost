package kubemodel

import "time"

// @bingen:generate:ResourceQuota
type ResourceQuota struct {
	UID          string               `json:"uid"`          // @bingen:field[version=1]
	NamespaceUID string               `json:"namespaceUID"` // @bingen:field[version=1]
	Name         string               `json:"name"`         // @bingen:field[version=1]
	Spec         *ResourceQuotaSpec   `json:"spec"`         // @bingen:field[version=1]
	Status       *ResourceQuotaStatus `json:"status"`       // @bingen:field[version=1]
	Start        time.Time            `json:"start"`        // @bingen:field[version=1]
	End          time.Time            `json:"end"`          // @bingen:field[version=1]
}

// @bingen:generate:ResourceQuotaSpec
type ResourceQuotaSpec struct {
	Hard *ResourceQuotaSpecHard `json:"hard"` // @bingen:field[version=1]
}

// @bingen:generate:ResourceQuotaSpecHard
type ResourceQuotaSpecHard struct {
	Requests ResourceQuantities `json:"requests,omitempty"` // @bingen:field[version=1]
	Limits   ResourceQuantities `json:"limits,omitempty"`   // @bingen:field[version=1]
}

func (spec *ResourceQuotaSpecHard) SetRequest(resource Resource, unit Unit, statType StatType, value float64) {
	if spec.Requests == nil {
		spec.Requests = ResourceQuantities{}
	}

	spec.Requests.Set(resource, unit, statType, value)
}

func (spec *ResourceQuotaSpecHard) SetLimit(resource Resource, unit Unit, statType StatType, value float64) {
	if spec.Limits == nil {
		spec.Limits = ResourceQuantities{}
	}

	spec.Limits.Set(resource, unit, statType, value)
}

// @bingen:generate:ResourceQuotaStatus
type ResourceQuotaStatus struct {
	Used *ResourceQuotaStatusUsed `json:"used"` // @bingen:field[version=1]
}

// @bingen:generate:ResourceQuotaStatusUsed
type ResourceQuotaStatusUsed struct {
	Requests ResourceQuantities `json:"requests,omitempty"` // @bingen:field[version=1]
	Limits   ResourceQuantities `json:"limits,omitempty"`   // @bingen:field[version=1]
}

func (stat *ResourceQuotaStatusUsed) SetRequest(resource Resource, unit Unit, statType StatType, value float64) {
	if stat.Requests == nil {
		stat.Requests = ResourceQuantities{}
	}

	stat.Requests.Set(resource, unit, statType, value)
}

func (stat *ResourceQuotaStatusUsed) SetLimit(resource Resource, unit Unit, statType StatType, value float64) {
	if stat.Limits == nil {
		stat.Limits = ResourceQuantities{}
	}

	stat.Limits.Set(resource, unit, statType, value)
}
