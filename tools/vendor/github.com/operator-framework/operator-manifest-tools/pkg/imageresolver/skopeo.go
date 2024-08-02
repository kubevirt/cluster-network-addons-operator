package imageresolver

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/operator-framework/operator-manifest-tools/internal/utils"
)

// Skopeo is the default image resolver using skopeo.
type Skopeo struct {
	path     string
	authFile string

	command commandCreator
}

// NewSkopeoResolver returns the skopeo resolver setting the exec filepath
// and the authfile used by skopeo.
func NewSkopeoResolver(skopeoPath, authFile string) (*Skopeo, error) {
	if authFile != "" {
		_, err := os.Stat(authFile)

		if err != nil {
			return nil, err
		}
	}

	return &Skopeo{
		path:     skopeoPath,
		authFile: authFile,
		command: func(name string, args ...string) commandRunner {
			return exec.Command(name, args...)
		},
	}, nil
}

const (
	timeout = "300s"
)

func (skopeo *Skopeo) getSkopeoResults(args ...string) ([]byte, map[string]interface{}, error) {
	baseArgs := []string{"--command-timeout", timeout, "inspect"}
	name := "skopeo"
	if skopeo.path != "" {
		name = skopeo.path
	}
	cmd := skopeo.command(name, append(baseArgs, args...)...)

	skopeoRaw, err := cmd.CombinedOutput()
	if err != nil {
		return nil, nil, err
	}

	var skopeoJSON map[string]interface{}

	err = json.Unmarshal(skopeoRaw, &skopeoJSON)
	if err != nil {
		return nil, nil, err
	}

	return skopeoRaw, skopeoJSON, nil
}

// ResolveImageReference will use the image resolver to map an image reference
// to the image's SHA256 value from the registry.
func (skopeo *Skopeo) ResolveImageReference(imageReference string) (string, error) {
	imageName, err := getName(imageReference)
	if err != nil {
		return "", err
	}

	imageReference = fmt.Sprintf("docker://%s", imageReference)
	args := []string{imageReference}

	if skopeo.authFile != "" {
		args = append(args, "--authfile", skopeo.authFile)
	}

	retryAttempts := 3

	var skopeoRaw []byte
	var skopeoJSON map[string]interface{}

	for i := 0; i < retryAttempts; i++ {
		rawArgs := append(args, "--raw")
		log.Println("skopeo inspect raw args are ", rawArgs)
		skopeoRaw, skopeoJSON, err = skopeo.getSkopeoResults(rawArgs...)
		if err != nil {
			continue
		}

		if version, ok := skopeoJSON["schemaVersion"].(float64); ok && version == 2 {
			rawDigest := fmt.Sprintf("%x", sha256.Sum256(skopeoRaw))
			return fmt.Sprintf("%s@sha256:%s", imageName, rawDigest), nil
		}

		log.Println("skopeo inspect args are ", args)
		_, skopeoJSON, err = skopeo.getSkopeoResults(args...)
		if err != nil {
			continue
		}

		digest, ok := skopeoJSON["Digest"].(string)

		if !ok {
			return "", errors.New("Digest not on response")
		}

		return fmt.Sprintf("%s@%s", imageName, digest), nil
	}

	if err != nil {
		return "", utils.NewErrImageDoesNotExist(imageReference, err)
	}

	return "", nil
}
