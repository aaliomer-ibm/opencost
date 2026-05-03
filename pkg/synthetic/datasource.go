package synthetic

import (
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/opencost/opencost/core/pkg/clusters"
	"github.com/opencost/opencost/core/pkg/diagnostics"
	"github.com/opencost/opencost/core/pkg/source"
	"github.com/opencost/opencost/modules/collector-source/pkg/collector"
	"github.com/opencost/opencost/modules/collector-source/pkg/metric"
	"github.com/opencost/opencost/modules/collector-source/pkg/util"
)

type SyntheticDataSource struct {
	metricsQuerier source.MetricsQuerier
	clusterMap     clusters.ClusterMap
	clusterInfo    clusters.ClusterInfoProvider
	resolution     time.Duration
}

func NewSyntheticDataSource(
	repo *metric.MetricRepository,
	resConfigs []util.ResolutionConfiguration,
	clusterUID, clusterName, provider, region string,
	resolution time.Duration,
) source.OpenCostDataSource {
	querier := collector.NewMetricsQuerier(repo, resConfigs)
	ci := &clusters.ClusterInfo{
		ID:       clusterUID,
		Name:     clusterName,
		Provider: provider,
		Region:   region,
	}
	return &SyntheticDataSource{
		metricsQuerier: querier,
		clusterMap:     &staticClusterMap{uid: clusterUID, info: ci},
		clusterInfo:    &staticClusterInfoProvider{info: ci},
		resolution:     resolution,
	}
}

func (s *SyntheticDataSource) RegisterEndPoints(_ *httprouter.Router)              {}
func (s *SyntheticDataSource) RegisterDiagnostics(_ diagnostics.DiagnosticService) {}
func (s *SyntheticDataSource) Metrics() source.MetricsQuerier                      { return s.metricsQuerier }
func (s *SyntheticDataSource) ClusterMap() clusters.ClusterMap                     { return s.clusterMap }
func (s *SyntheticDataSource) ClusterInfo() clusters.ClusterInfoProvider           { return s.clusterInfo }
func (s *SyntheticDataSource) BatchDuration() time.Duration                        { return 1<<63 - 1 }
func (s *SyntheticDataSource) Resolution() time.Duration                           { return s.resolution }

type staticClusterMap struct {
	uid  string
	info *clusters.ClusterInfo
}

func (m *staticClusterMap) GetClusterIDs() []string { return []string{m.uid} }
func (m *staticClusterMap) AsMap() map[string]*clusters.ClusterInfo {
	return map[string]*clusters.ClusterInfo{m.uid: m.info}
}
func (m *staticClusterMap) InfoFor(id string) *clusters.ClusterInfo {
	if id == m.uid {
		return m.info
	}
	return nil
}
func (m *staticClusterMap) NameFor(id string) string {
	if id == m.uid {
		return m.info.Name
	}
	return ""
}
func (m *staticClusterMap) NameIDFor(id string) string {
	if id == m.uid && m.info.Name != "" {
		return m.info.Name + "/" + m.uid
	}
	return id
}

type staticClusterInfoProvider struct {
	info *clusters.ClusterInfo
}

func (p *staticClusterInfoProvider) GetClusterInfo() map[string]string {
	return map[string]string{clusters.ClusterInfoIdKey: p.info.ID, clusters.ClusterInfoNameKey: p.info.Name, clusters.ClusterInfoProviderKey: p.info.Provider, clusters.ClusterInfoRegionKey: p.info.Region}
}
