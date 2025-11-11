package imagename

import (
	"encoding"
	"errors"
	"fmt"
	"strings"
)

// FormatOption is a the byte switch that sets the format for the image ToString func
type FormatOption byte

const (
	// Registry output option includes the registry in the string.
	Registry FormatOption = 1 << iota
	// Tag includes the tag in the output string.
	Tag
	// ExplicitTag forces a tag in the string, will use latest if no tag is avail.
	ExplicitTag
	// ExplicitNamespace forces a namespace in the repo, will use "library" if no namespace
	ExplicitNamespace
)

var (
	// DefaultGetStringOptions is the default set of options
	DefaultGetStringOptions FormatOption = Registry | Tag

	// ErrNoImageRepository returns when there is there is no image repository
	ErrNoImageRepository = errors.New("No image repository specified")
)

// Set sets the binary flag
func (b FormatOption) Set(flag FormatOption) FormatOption { return b | flag }

// Clear clears the binary flag
func (b FormatOption) Clear(flag FormatOption) FormatOption { return b &^ flag }

// Toggle toggles the flag
func (b FormatOption) Toggle(flag FormatOption) FormatOption { return b ^ flag }

// Has checks if the flag is set.
func (b FormatOption) Has(flag FormatOption) bool { return b&flag != 0 }

// ImageName represents the parts of an image reference.
type ImageName struct {
	Registry  string
	Namespace string
	Repo      string
	Tag       string
}

var _ encoding.TextMarshaler = ImageName{}

// MarshalText marhsals the image name into a byte string so it can be used as a JSON map key
func (name ImageName) MarshalText() ([]byte, error) {
	return []byte(name.String()), nil
}

// HasDigest return true if the image uses a digest.
func (imageName *ImageName) HasDigest() bool {
	return strings.HasPrefix(imageName.Tag, "sha256:")
}

// GetRepo returns the repository of the image.
func (imageName *ImageName) GetRepo(option FormatOption) string {
	result := imageName.Repo

	if imageName.Namespace != "" {
		result = fmt.Sprintf("%s/%s", imageName.Namespace, result)
	}

	if option.Has(ExplicitNamespace) {
		result = fmt.Sprintf("%s/%s", "library", result)
	}

	return result
}

// ToString will print the image using formatting options.
func (imageName *ImageName) ToString(optionSet FormatOption) (string, error) {

	if imageName.Repo == "" {
		return "", ErrNoImageRepository
	}

	result := imageName.GetRepo(optionSet)

	if optionSet.Has(Tag) && imageName.Tag != "" {
		if imageName.HasDigest() {
			result = fmt.Sprintf("%s@%s", result, imageName.Tag)
		} else {
			result = fmt.Sprintf("%s:%s", result, imageName.Tag)
		}
	} else if optionSet.Has(Tag) && optionSet.Has(ExplicitTag) {
		result = fmt.Sprintf("%s:%s", result, "latest")
	}

	if optionSet.Has(Registry) && imageName.Registry != "" {
		result = fmt.Sprintf("%s/%s", imageName.Registry, result)
	}

	return result, nil
}

// Enclose will set the organization on the image.
func (imageName *ImageName) Enclose(organization string) {
	if imageName.Namespace == organization {
		return
	}

	repoParts := []string{imageName.Repo}

	if imageName.Namespace != "" {
		repoParts = append([]string{imageName.Namespace}, repoParts...)
	}

	imageName.Namespace = organization
	imageName.Repo = strings.Join(repoParts, "-")
}

// String returns the string representation of the image using
// the registry and tag formatting.
func (imageName *ImageName) String() string {
	result, err := imageName.ToString(Registry | Tag)

	if err != nil {
		panic(err)
	}

	return result
}

// Parse parses the image from a string.
func Parse(imageName string) *ImageName {
	result := &ImageName{}

	s := strings.SplitN(imageName, "/", 3)
	if len(s) == 2 {
		if strings.ContainsAny(s[0], ".:") {
			result.Registry = s[0]
		} else {
			result.Namespace = s[0]
		}
	} else if len(s) == 3 {
		result.Registry = s[0]
		result.Namespace = s[1]
	}

	result.Repo = s[len(s)-1]
	result.Tag = "latest"

	if strings.ContainsAny(result.Repo, "@:") {
		s = strings.SplitN(result.Repo, "@", 2)

		if len(s) != 2 {
			s = strings.SplitN(result.Repo, ":", 2)
		}

		if len(s) == 2 {
			result.Repo, result.Tag = s[0], s[1]
		}
	}
	return result
}
