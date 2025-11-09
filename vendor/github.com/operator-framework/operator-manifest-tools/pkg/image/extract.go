package image

import (
	"errors"

	"github.com/operator-framework/operator-manifest-tools/pkg/pullspec"
)

// Extract finds and returns all image names in a manifest
func Extract(manifests []*pullspec.OperatorCSV) ([]string, error) {
	imageNames := make([]string, 0, len(manifests))
	for _, manifest := range manifests {
		pullSpecs, err := manifest.GetPullSpecs()
		if err != nil {
			return nil, errors.New("error getting pullspec: " + err.Error())
		}

		for _, pullSpec := range pullSpecs {
			imageNames = append(imageNames, pullSpec.String())
		}
	}

	return imageNames, nil
}
