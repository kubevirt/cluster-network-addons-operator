//go:build tools

package tools

import (
	_ "github.com/github-release/github-release"
	_ "github.com/onsi/ginkgo/v2"
	_ "github.com/operator-framework/operator-sdk/cmd/operator-sdk"
	_ "golang.org/x/tools/cmd/goimports"
	_ "sigs.k8s.io/controller-tools/cmd/controller-gen"
)
