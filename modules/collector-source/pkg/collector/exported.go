package collector

import (
	"github.com/opencost/opencost/core/pkg/source"
	"github.com/opencost/opencost/modules/collector-source/pkg/metric"
	"github.com/opencost/opencost/modules/collector-source/pkg/util"
)

func NewMetricsQuerier(repo *metric.MetricRepository, resolutionConfigs []util.ResolutionConfiguration) source.MetricsQuerier {
	return newCollectorMetricsQuerier(repo, resolutionConfigs)
}
