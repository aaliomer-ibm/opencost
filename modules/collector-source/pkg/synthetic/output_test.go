package synthetic

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestOutputWriter_BingenPath(t *testing.T) {
	tmpDir := t.TempDir()
	w := NewOutputWriter(OutputConfig{
		BaseDir:    tmpDir,
		ClusterID:  "cluster-abc-123",
		Resolution: "1h",
	})

	start := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	end := time.Date(2025, 6, 15, 13, 0, 0, 0, time.UTC)

	err := w.WriteBingenJSON("allocations", start, end, map[string]string{"test": "data"})
	if err != nil {
		t.Fatalf("WriteBingenJSON failed: %v", err)
	}

	expected := filepath.Join(tmpDir,
		"federated/cluster-abc-123/etl/bingen/allocations/1h",
		fmt.Sprintf("%d-%d.json", start.Unix(), end.Unix()))
	if _, err := os.Stat(expected); os.IsNotExist(err) {
		t.Errorf("expected file not found: %s", expected)
	}
}

func TestOutputWriter_KubeModelPath(t *testing.T) {
	tmpDir := t.TempDir()
	w := NewOutputWriter(OutputConfig{
		BaseDir:    tmpDir,
		ClusterID:  "cluster-abc-123",
		Resolution: "1h",
	})

	start := time.Date(2025, 12, 15, 12, 0, 0, 0, time.UTC)

	err := w.WriteKubeModelJSON(start, map[string]string{"test": "data"})
	if err != nil {
		t.Fatalf("WriteKubeModelJSON failed: %v", err)
	}

	expected := filepath.Join(tmpDir,
		"finops-agent/cluster-abc-123/kubemodel/1h/2025/12/15/20251215120000.v2.json")
	if _, err := os.Stat(expected); os.IsNotExist(err) {
		t.Errorf("expected file not found: %s", expected)
	}
}

func TestOutputWriter_EventPath(t *testing.T) {
	tmpDir := t.TempDir()
	w := NewOutputWriter(OutputConfig{
		BaseDir:    tmpDir,
		ClusterID:  "cluster-abc-123",
		Resolution: "1h",
	})

	ts := time.Date(2025, 6, 15, 12, 40, 0, 0, time.UTC)

	err := w.WriteEventJSON("finops-agent", "pricingmodel", ts, map[string]string{"test": "data"})
	if err != nil {
		t.Fatalf("WriteEventJSON failed: %v", err)
	}

	expected := filepath.Join(tmpDir,
		"finops-agent/cluster-abc-123/pricingmodel/20250615124000.json")
	if _, err := os.Stat(expected); os.IsNotExist(err) {
		t.Errorf("expected file not found: %s", expected)
	}
}
