package imageresolver

import (
	"fmt"
	"strings"

	"github.com/operator-framework/operator-manifest-tools/pkg/imagename"
)

// ImageResolve implements a method of identifying an image reference.
type ImageResolver interface {
	// ResolveImageReference will use the image resolver to map an image reference
	// to the image's SHA256 value from the registry.
	ResolveImageReference(imageReference string) (string, error)
}

type commandRunner interface {
	CombinedOutput() ([]byte, error)
}

type commandCreator func(name string, arg ...string) commandRunner

type ResolverOption string

func (opt *ResolverOption) String() string {
	if opt == nil {
		return ""
	}

	return string(*opt)
}

const (
	ResolverCrane  ResolverOption = "crane"
	ResolverSkopeo ResolverOption = "skopeo"
	ResolverScript ResolverOption = "script"
)

var (
	validResolvers ResolverOptions = ResolverOptions{ResolverScript, ResolverSkopeo, ResolverCrane}
)

type ResolverOptions []ResolverOption

func (opts ResolverOptions) String() string {
	str := strings.Builder{}

	for i, v := range opts {
		str.WriteString(string(v))
		if i != len(opts)-1 {
			str.WriteString(", ")
		}
	}

	return str.String()
}

func GetResolverOptions() ResolverOptions {
	return validResolvers
}

func GetResolver(resolver ResolverOption, args map[string]string) (ImageResolver, error) {
	path, pathOk := args["path"]
	switch resolver {
	case ResolverSkopeo:
		if !pathOk {
			path = "skopeo"
		}

		authFile := args["authFile"]
		return NewSkopeoResolver(path, authFile)
	case ResolverScript:
		if !pathOk {
			return nil, fmt.Errorf("path is required for the script image resolver")
		}

		return &Script{path: path}, nil
	case ResolverCrane:
		opts := make([]CraneOption, 0, 1)
		username, ok := args["username"]
		if ok {
			opts = append(opts, WithUserPassAuth(username, args["password"]))
		} else {
			usedefault := args["usedefault"]
			if usedefault == "true" {
				opts = append(opts, WithDefaultKeychain())
			}
		}
		insecure := args["insecure"]
		if insecure == "true" {
			opts = append(opts, Insecure())
		}

		return NewCraneResolver(opts...), nil
	default:
		return nil, fmt.Errorf("resolver option provided isn't valid: %s", resolver)
	}
}

func getName(imageReference string) (string, error) {
	name := imagename.Parse(imageReference)
	return name.ToString(imagename.Registry)
}
