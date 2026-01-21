package kubemodel

// @bingen:generate:Unit
type Unit string

type Measurement = float64

// bytesToKiBFactor is the conversion factor from bytes to kibibytes (1 KiB = 1024 bytes)
const bytesToKiBFactor = 1024

// BytesToKiB converts bytes to kibibytes (KiB)
func BytesToKiB(bytes Measurement) Measurement {
	return bytes / bytesToKiBFactor
}

// KiBToBytes converts kibibytes (KiB) to bytes
func KiBToBytes(kib Measurement) Measurement {
	return kib * bytesToKiBFactor
}
