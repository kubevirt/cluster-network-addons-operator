package image

import (
	"github.com/operator-framework/operator-manifest-tools/pkg/imageresolver"
	"github.com/operator-framework/operator-manifest-tools/pkg/pullspec"
)

// Pin iterates through manifests and replaces all image tags with resolved digests.
func Pin(resolver imageresolver.ImageResolver, manifests []*pullspec.OperatorCSV) error {
	imageNames, err := Extract(manifests)
	if err != nil {
		return err
	}
	replacements, err := Resolve(resolver, imageNames)
	if err != nil {
		return err
	}

	return Replace(manifests, replacements)
}
