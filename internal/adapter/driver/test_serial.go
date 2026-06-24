package driver

import "strings"

// IsTestSerial — serial без реального ADB (stub и e2e-префикс).
func IsTestSerial(serial string) bool {
	return serial == "stub" || strings.HasPrefix(serial, "E2E-")
}
