package image

import (
	"errors"
	"fmt"

	"github.com/operator-framework/operator-manifest-tools/pkg/imagename"
	"github.com/operator-framework/operator-manifest-tools/pkg/pullspec"
)

// Repleamcents contains a mapping of image names to alternative image names that should be used instead.
// Usually image names with tags to image names with digests.
type Replacements map[imagename.ImageName]imagename.ImageName

// NewReplacements takes in raw replacements and parses them to image names.
func NewReplacements(replacements map[string]string) (Replacements, error) {
	r := make(Replacements, len(replacements))
	for k, v := range replacements {
		key := imagename.Parse(k)
		value := imagename.Parse(v)
		if key == nil || value == nil {
			return nil, fmt.Errorf("invalid replacement: (%q => %q)", k, v)
		}

		r[*key] = *value
	}

	return r, nil
}

// Replace takes a list of manifests and replaces the images specified in the replacement mapping.
func Replace(manifests []*pullspec.OperatorCSV, replacements Replacements) error {
	for i := range manifests {
		manifest := manifests[i]
		if err := manifest.ReplacePullSpecsEverywhere(replacements); err != nil {
			return errors.New("failed to replace everywhere: " + err.Error())
		}

		if err := manifest.SetRelatedImages(); err != nil {
			return errors.New("failed to set related images: " + err.Error())
		}
	}

	return nil
}
