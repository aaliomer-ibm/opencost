package kubemodel

// @bingen:generate:Unit
type Unit string

// bytesToKiBFactor is the conversion factor from bytes to kibibytes (1 KiB = 1024 bytes)
const bytesToKiBFactor = 1024

// BytesToKiB converts bytes to kibibytes (KiB)
func BytesToKiB(bytes uint64) uint64 {
	return bytes / bytesToKiBFactor
}

// KiBToBytes converts kibibytes (KiB) to bytes
func KiBToBytes(kib uint64) uint64 {
	return kib * bytesToKiBFactor
}
