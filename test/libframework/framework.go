package libframework

import "runtime"

// NamespaceTestDefault is the default namespace used for tests
const NamespaceTestDefault = "cnao-test"

// Arch represents the current architecture
var Arch = runtime.GOARCH

// IsARM64 checks if the given architecture is ARM64
func IsARM64(arch string) bool {
	return arch == "arm64"
}
