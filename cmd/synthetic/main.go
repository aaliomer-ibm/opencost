package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/opencost/opencost/core/pkg/log"
	"github.com/opencost/opencost/core/pkg/opencost"
	"github.com/opencost/opencost/modules/collector-source/pkg/collector"
	"github.com/opencost/opencost/modules/collector-source/pkg/metric"
	synthspec "github.com/opencost/opencost/modules/collector-source/pkg/synthetic"
	"github.com/opencost/opencost/modules/collector-source/pkg/util"
	"github.com/opencost/opencost/pkg/costmodel"
	synth "github.com/opencost/opencost/pkg/synthetic"
)

func main() {
	config := parseCommandLineArgs()

	clusterSpec := generateClusterSpecification(config)
	metricsRepo := generateTimeSeriesMetrics(clusterSpec, config)
	costModel := createCostModelFromMetrics(clusterSpec, metricsRepo, config)

	computeAndWriteOutputs(costModel, clusterSpec, config)
}

func parseCommandLineArgs() *generationConfig {
	cfg := &generationConfig{}
	now := time.Now().UTC().Truncate(time.Minute)
	defaultStart := now.Add(-36 * time.Hour)

	flag.StringVar(&cfg.startStr, "start", defaultStart.Format(time.RFC3339), "Start time (RFC3339)")
	flag.StringVar(&cfg.endStr, "end", now.Format(time.RFC3339), "End time (RFC3339)")
	flag.Uint64Var(&cfg.seed, "seed", 42, "Random seed for deterministic generation")
	flag.IntVar(&cfg.numNodes, "nodes", 2, "Number of nodes")
	flag.IntVar(&cfg.podsPerNode, "pods-per-node", 1, "Pods per node")
	flag.IntVar(&cfg.numNamespaces, "namespaces", 1, "Number of namespaces")
	flag.IntVar(&cfg.numPVs, "pvs", 1, "Number of PVs in the storage pool")
	flag.IntVar(&cfg.numLoadBalancers, "load-balancers", 1, "Number of load balancers")
	flag.StringVar(&cfg.intervalStr, "interval", "1m", "Scrape interval for metric generation")
	flag.StringVar(&cfg.outputDir, "output-dir", "./agent-output", "Output directory")
	flag.StringVar(&cfg.resolution, "resolution", "", "Output resolution (10m, 1h, 1d, comma-separated, or empty for all)")
	flag.StringVar(&cfg.minPodLifetimeStr, "min-pod-lifetime", "10m", "Minimum pod lifetime")
	flag.StringVar(&cfg.maxPodLifetimeStr, "max-pod-lifetime", "1h", "Maximum pod lifetime")
	flag.Float64Var(&cfg.pvcRatio, "pvc-ratio", 0.1, "Ratio of pods with PVCs (0.0-1.0)")
	flag.Float64Var(&cfg.podMigrationProb, "pod-migration-prob", 0.05, "Probability of pod migration (0.0-1.0)")
	flag.Parse()

	cfg.validateAndParseTimes()
	return cfg
}

func generateClusterSpecification(cfg *generationConfig) synthspec.ClusterSpec {
	log.Infof("Generating cluster spec: %d nodes, %d pods/node, %d namespaces, %d PVs, %d LBs",
		cfg.numNodes, cfg.podsPerNode, cfg.numNamespaces, cfg.numPVs, cfg.numLoadBalancers)
	log.Infof("Pod lifecycle: %s-%s, PVC ratio: %.1f%%, migration prob: %.1f%%",
		cfg.minPodLifetime, cfg.maxPodLifetime, cfg.pvcRatio*100, cfg.podMigrationProb*100)

	return synthspec.RandomClusterSpec(synthspec.RandomClusterSpecConfig{
		NumNodes: cfg.numNodes, PodsPerNode: cfg.podsPerNode, NumNamespaces: cfg.numNamespaces,
		NumPVs: cfg.numPVs, NumLoadBalancers: cfg.numLoadBalancers,
		LabelsPerPod: 3, LabelsPerNode: 5, LabelsPerNamespace: 2,
		Seed: cfg.seed, ScrapeInterval: cfg.scrapeInterval,
		TotalDuration:  cfg.timeWindow(),
		MinPodLifetime: cfg.minPodLifetime, MaxPodLifetime: cfg.maxPodLifetime,
		PVCRatio: cfg.pvcRatio, PodMigrationProb: cfg.podMigrationProb,
	})
}

func generateTimeSeriesMetrics(spec synthspec.ClusterSpec, cfg *generationConfig) *metric.MetricRepository {
	resolutions := createMetricResolutions(cfg.timeWindow())
	repo := metric.NewMetricRepository(resolutions, collector.NewOpenCostMetricStore)

	log.Infof("Generating metrics from %s to %s (interval=%s)",
		cfg.startTime.Format(time.RFC3339), cfg.endTime.Format(time.RFC3339), cfg.scrapeInterval)

	generator := synthspec.NewGenerator(spec, cfg.scrapeInterval)
	generator.Generate(repo, cfg.startTime, cfg.endTime)

	return repo
}

func createCostModelFromMetrics(spec synthspec.ClusterSpec, repo *metric.MetricRepository, cfg *generationConfig) *costmodel.CostModel {
	resConfigs := createResolutionConfigs(cfg.timeWindow())
	dataSource := synth.NewSyntheticDataSource(repo, resConfigs, spec.ClusterUID, spec.ClusterName, spec.Provider, spec.Region, cfg.scrapeInterval)
	provider := synth.NewSyntheticProvider(spec)

	return costmodel.NewCostModel(spec.ClusterUID, dataSource, provider, nil, dataSource.ClusterMap(), dataSource.BatchDuration())
}

func computeAndWriteOutputs(cm *costmodel.CostModel, spec synthspec.ClusterSpec, cfg *generationConfig) {
	resolutions := []string{"10m", "1h", "1d"}
	if cfg.resolution != "" {
		resolutions = parseResolutions(cfg.resolution)
	}

	log.Infof("Computing and writing outputs to %s (resolutions: %v)", cfg.outputDir, resolutions)

	var allAssetSets []*opencost.AssetSet
	for _, res := range resolutions {
		writer := synthspec.NewOutputWriter(synthspec.OutputConfig{
			BaseDir: cfg.outputDir, ClusterID: spec.ClusterUID, Resolution: res,
		})

		log.Infof("Generating %s resolution data", res)
		assetSets := processTimeWindows(cm, writer, cfg, res)
		if len(assetSets) > 0 {
			allAssetSets = assetSets
		}
	}

	writer := synthspec.NewOutputWriter(synthspec.OutputConfig{
		BaseDir: cfg.outputDir, ClusterID: spec.ClusterUID, Resolution: resolutions[len(resolutions)-1],
	})
	writePricingModel(writer, allAssetSets, spec, cfg)

	log.Infof("Done. Output written to %s", cfg.outputDir)
}

func processTimeWindows(cm *costmodel.CostModel, writer *synthspec.OutputWriter, cfg *generationConfig, resolution string) []*opencost.AssetSet {
	var allAssetSets []*opencost.AssetSet
	windowDuration := parseResolution(resolution)

	for windowStart := cfg.startTime.Truncate(windowDuration); windowStart.Before(cfg.endTime); {
		windowEnd := minTime(windowStart.Add(windowDuration), cfg.endTime)

		writeKubeModel(cm, writer, windowStart, windowEnd)
		writeAllocations(cm, writer, windowStart, windowEnd)
		assetSet := writeAssets(cm, writer, windowStart, windowEnd)

		if assetSet != nil {
			allAssetSets = append(allAssetSets, assetSet)
		}
		windowStart = windowEnd
	}

	return allAssetSets
}

func writeKubeModel(cm *costmodel.CostModel, writer *synthspec.OutputWriter, start, end time.Time) {
	kms, err := cm.ComputeKubeModelSet(start, end)
	if err != nil {
		log.Warnf("KubeModelSet error for %s: %v", start, err)
		return
	}
	if err := writer.WriteKubeModelJSON(start, kms); err != nil {
		log.Errorf("Failed to write KubeModelSet: %v", err)
	}
}

func writeAllocations(cm *costmodel.CostModel, writer *synthspec.OutputWriter, start, end time.Time) {
	allocSet, err := cm.ComputeAllocation(start, end)
	if err != nil {
		log.Warnf("AllocationSet error for %s: %v", start, err)
		return
	}
	if err := writer.WriteBingenJSON("allocations", start, end, allocSet); err != nil {
		log.Errorf("Failed to write AllocationSet: %v", err)
	}
}

func writeAssets(cm *costmodel.CostModel, writer *synthspec.OutputWriter, start, end time.Time) *opencost.AssetSet {
	assetSet, err := cm.ComputeAssets(start, end)
	if err != nil {
		log.Warnf("AssetSet error for %s: %v", start, err)
		return nil
	}
	if err := writer.WriteBingenJSON("assets", start, end, assetSet); err != nil {
		log.Errorf("Failed to write AssetSet: %v", err)
	}
	return assetSet
}

func writePricingModel(writer *synthspec.OutputWriter, assetSets []*opencost.AssetSet, spec synthspec.ClusterSpec, cfg *generationConfig) {
	pricingModel := synth.CreatePricingModelFromAssets(assetSets, spec.Provider, spec.Region, cfg.startTime)
	pricingJSON := synth.ConvertPricingModelToJSON(pricingModel)

	if err := writer.WriteEventJSON("finops-agent", "pricingmodel", cfg.startTime, pricingJSON); err != nil {
		log.Errorf("Failed to write PricingModelSet: %v", err)
	}
}

type generationConfig struct {
	startStr, endStr, intervalStr        string
	minPodLifetimeStr, maxPodLifetimeStr string
	startTime, endTime                   time.Time
	scrapeInterval                       time.Duration
	minPodLifetime, maxPodLifetime       time.Duration
	seed                                 uint64
	numNodes, podsPerNode                int
	numNamespaces, numPVs                int
	numLoadBalancers                     int
	pvcRatio, podMigrationProb           float64
	outputDir, resolution                string
}

func (cfg *generationConfig) validateAndParseTimes() {
	var err error
	cfg.startTime, err = time.Parse(time.RFC3339, cfg.startStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid --start: %v\n", err)
		os.Exit(1)
	}
	cfg.endTime, err = time.Parse(time.RFC3339, cfg.endStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid --end: %v\n", err)
		os.Exit(1)
	}
	cfg.scrapeInterval, err = time.ParseDuration(cfg.intervalStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid --interval: %v\n", err)
		os.Exit(1)
	}
	cfg.minPodLifetime, err = time.ParseDuration(cfg.minPodLifetimeStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid --min-pod-lifetime: %v\n", err)
		os.Exit(1)
	}
	cfg.maxPodLifetime, err = time.ParseDuration(cfg.maxPodLifetimeStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid --max-pod-lifetime: %v\n", err)
		os.Exit(1)
	}
	if cfg.pvcRatio < 0 || cfg.pvcRatio > 1 {
		fmt.Fprintf(os.Stderr, "--pvc-ratio must be between 0.0 and 1.0\n")
		os.Exit(1)
	}
	if cfg.podMigrationProb < 0 || cfg.podMigrationProb > 1 {
		fmt.Fprintf(os.Stderr, "--pod-migration-prob must be between 0.0 and 1.0\n")
		os.Exit(1)
	}
}

func (cfg *generationConfig) timeWindow() time.Duration {
	return cfg.endTime.Sub(cfg.startTime)
}

func minTime(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}

func createMetricResolutions(totalDuration time.Duration) []*util.Resolution {
	configs := createResolutionConfigs(totalDuration)
	resolutions := make([]*util.Resolution, 0, len(configs))

	for _, rc := range configs {
		res, err := util.NewResolution(rc)
		if err != nil {
			log.Warnf("Failed to create resolution %s: %v", rc.Interval, err)
			continue
		}
		resolutions = append(resolutions, res)
	}
	return resolutions
}

func createResolutionConfigs(totalDuration time.Duration) []util.ResolutionConfiguration {
	return []util.ResolutionConfiguration{
		{Interval: "10m", Retention: int(totalDuration.Minutes()/10) + 6},
		{Interval: "1h", Retention: int(totalDuration.Hours()) + 24},
		{Interval: "1d", Retention: int(totalDuration.Hours()/24) + 7},
	}
}

func parseResolution(res string) time.Duration {
	switch res {
	case "10m":
		return 10 * time.Minute
	case "1h":
		return time.Hour
	case "1d":
		return 24 * time.Hour
	default:
		d, err := time.ParseDuration(res)
		if err != nil {
			return time.Hour
		}
		return d
	}
}

func parseResolutions(resStr string) []string {
	if resStr == "" {
		return []string{"10m", "1h", "1d"}
	}

	parts := strings.Split(resStr, ",")
	resolutions := make([]string, 0, len(parts))

	for _, part := range parts {
		res := strings.TrimSpace(part)
		if res != "" {
			resolutions = append(resolutions, res)
		}
	}

	if len(resolutions) == 0 {
		return []string{"10m", "1h", "1d"}
	}

	return resolutions
}
