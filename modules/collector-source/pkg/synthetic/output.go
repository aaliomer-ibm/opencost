package synthetic

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

const (
	bingenRootDir   = "federated"
	bingenStorePath = "etl/bingen"
	kubeModelRoot   = "finops-agent"
)

type OutputConfig struct {
	BaseDir    string
	ClusterID  string
	Resolution string
}

type OutputWriter struct {
	cfg OutputConfig
}

func NewOutputWriter(cfg OutputConfig) *OutputWriter {
	return &OutputWriter{cfg: cfg}
}

func (w *OutputWriter) WriteBingenJSON(pipeline string, start, end time.Time, data any) error {
	epochRange := fmt.Sprintf("%s-%s",
		strconv.FormatInt(start.Unix(), 10),
		strconv.FormatInt(end.Unix(), 10),
	)
	dir := filepath.Join(w.cfg.BaseDir, bingenRootDir, w.cfg.ClusterID, bingenStorePath, pipeline, w.cfg.Resolution)
	return writeJSON(dir, epochRange+".json", data)
}

func (w *OutputWriter) WriteKubeModelJSON(start time.Time, data any) error {
	dateDir := start.UTC().Format("2006/01/02")
	fileName := start.UTC().Format("20060102150405") + ".v2.json"
	dir := filepath.Join(w.cfg.BaseDir, kubeModelRoot, w.cfg.ClusterID, "kubemodel", w.cfg.Resolution, dateDir)
	return writeJSON(dir, fileName, data)
}

func (w *OutputWriter) WriteEventJSON(root, event string, timestamp time.Time, data any) error {
	fileName := timestamp.UTC().Format("20060102150405") + ".json"
	dir := filepath.Join(w.cfg.BaseDir, root, w.cfg.ClusterID, event)
	return writeJSON(dir, fileName, data)
}

func writeJSON(dir, fileName string, data any) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}
	fullPath := filepath.Join(dir, fileName)
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	if err := os.WriteFile(fullPath, b, 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", fullPath, err)
	}
	return nil
}
