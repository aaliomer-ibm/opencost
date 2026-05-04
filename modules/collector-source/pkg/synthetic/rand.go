package synthetic

import (
	"fmt"

	"github.com/brianvoe/gofakeit/v7"
)

func randLabelMap(f *gofakeit.Faker, n int) map[string]string {
	labels := make(map[string]string, n)
	for range n {
		key := fmt.Sprintf("label_%s_%s", f.RandomString(labelDomains), f.RandomString(labelKeys))
		labels[key] = fmt.Sprintf("%s-%s", f.Adjective(), f.Noun())
	}
	return labels
}

var (
	labelDomains     = []string{"app", "team", "env", "tier", "zone", "role", "stage", "group", "dept", "cost", "sla", "org", "project", "service", "component"}
	labelKeys        = []string{"name", "version", "type", "class", "level", "priority", "owner", "region", "stack", "release", "track", "channel", "flavor", "variant"}
	instanceTypes    = []string{"m5.large", "m5.xlarge", "m5.2xlarge", "m5.4xlarge", "c5.large", "c5.xlarge", "c5.2xlarge", "c5.4xlarge", "r5.large", "r5.xlarge", "r5.2xlarge", "r5.4xlarge", "t3.medium", "t3.large", "t3.xlarge"}
	containerImages  = []string{"nginx", "redis", "postgres", "mysql", "mongo", "memcached", "rabbitmq", "kafka", "zookeeper", "consul", "vault", "envoy", "prometheus", "grafana", "elasticsearch", "kibana", "fluentd", "istio-proxy", "coredns", "etcd", "haproxy", "traefik", "api-server", "worker", "scheduler", "controller", "webhook", "ingress", "gateway", "sidecar", "init", "migration"}
	workloadPrefixes = []string{"web", "api", "worker", "batch", "cron", "cache", "queue", "stream", "auth", "gateway", "proxy", "monitor", "logger", "indexer", "scraper", "ingest", "transform", "export", "sync", "relay", "dispatch", "handler", "service", "daemon", "agent", "collector", "processor", "scheduler"}
	namespaceNames   = []string{"default", "kube-system", "monitoring", "logging", "ingress", "production", "staging", "development", "data-pipeline", "ml-training", "analytics", "payments", "notifications", "search", "media", "platform", "infra", "security", "networking", "storage"}
	ownerKinds       = []string{"ReplicaSet", "DaemonSet", "StatefulSet", "Job"}
	azSuffixes       = []string{"a", "b", "c", "d"}
	sidecarImages    = []string{"istio-proxy", "envoy", "sidecar", "fluentd", "init"}
	storageClasses   = []string{"gp3", "gp2", "io1", "standard"}
)
