package pullspec

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"io/fs"

	"github.com/operator-framework/operator-manifest-tools/internal/utils"
	"github.com/operator-framework/operator-manifest-tools/pkg/imagename"
	yamlv3 "gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
)

// NamedPullSpec is an interface that allows for some elements
// of a pull spec to be generalized.
type NamedPullSpec interface {
	fmt.Stringer
	Name() string
	Image() string
	Data() map[string]interface{}
	SetImage(string)
	AsYamlObject() map[string]interface{}
}

type namedPullSpec struct {
	imageKey string
	data     map[string]interface{}
}

// Name returns the name of the pull spec data.
func (named *namedPullSpec) Name() string {
	return strings.TrimSpace(named.data["name"].(string))
}

// Image returns the image string.
func (named *namedPullSpec) Image() string {
	return named.data[named.imageKey].(string)
}

// Data returns the namedPullSpec as a json map
func (named *namedPullSpec) Data() map[string]interface{} {
	return named.data
}

// SetImage will override the image of the pull spec
func (named *namedPullSpec) SetImage(image string) {
	named.data[named.imageKey] = image
}

// AsYamlObject returns the pull spec as an object
func (named *namedPullSpec) AsYamlObject() map[string]interface{} {
	return map[string]interface{}{
		"name":  named.Name(),
		"image": named.Image(),
	}
}

// Container is a pullspec for containers in kubernetes resources like a pod
type Container struct {
	namedPullSpec
}

// String returns a string representation of the container pullSpec.
func (container *Container) String() string {
	return fmt.Sprintf("container %s", container.Name())
}

// NewContainer returns a container pullspec
func NewContainer(data interface{}) (*Container, error) {
	dataMap, ok := data.(map[string]interface{})

	if !ok {
		return nil, errors.New("expected map[string]interface{} type")
	}

	if _, ok := dataMap["image"]; !ok {
		return nil, utils.ErrImageIsARequiredProperty
	}

	return &Container{
		namedPullSpec: namedPullSpec{
			imageKey: "image",
			data:     dataMap,
		},
	}, nil
}

// InitContainer is a pull spec representing init containers
// in kubernetes objects.
type InitContainer struct {
	namedPullSpec
}

// String returns a string representation of the pullspec.
func (container *InitContainer) String() string {
	return fmt.Sprintf("initcontainer %s", container.Name())
}

// NewInitContainer returns a new init container pullspec.
func NewInitContainer(data interface{}) (*InitContainer, error) {
	dataMap, ok := data.(map[string]interface{})

	if !ok {
		return nil, errors.New("expected map[string]interface{} type")
	}

	if _, ok := dataMap["image"]; !ok {
		return nil, utils.ErrImageIsARequiredProperty
	}

	return &InitContainer{
		namedPullSpec: namedPullSpec{
			imageKey: "image",
			data:     dataMap,
		},
	}, nil
}

// RelatedImage is a pullspec representing the CSV relatedImage field.
type RelatedImage struct {
	namedPullSpec
}

// String returns a string representation of the pullspec.
func (relatedImage *RelatedImage) String() string {
	return fmt.Sprintf("relatedImage %s", relatedImage.Name())
}

// NewRelatedImage returns a new related image pullspec.
func NewRelatedImage(data interface{}) (*RelatedImage, error) {
	dataMap, ok := data.(map[string]interface{})

	if !ok {
		return nil, errors.New("expected map[string]interface{} type")
	}

	return &RelatedImage{
		namedPullSpec: namedPullSpec{
			imageKey: "image",
			data:     dataMap,
		},
	}, nil
}

// RelatedImageEnv is a pullspec representing environment variables
// that start with RELATED_IMAGE_.
type RelatedImageEnv struct {
	namedPullSpec
}

// String returns a string representation of the pullspec.
func (relatedImageEnv *RelatedImageEnv) String() string {
	return fmt.Sprintf("%s var", relatedImageEnv.Name())
}

// Name returns the name of the related image.
func (relatedImageEnv *RelatedImageEnv) Name() string {
	text := fmt.Sprintf("%v", relatedImageEnv.data["name"])
	return strings.TrimSpace(strings.ToLower(text[len("RELATED_IMAGE_"):]))
}

// AsYamlObject returns the pullspec as a map[string]interface{}.
func (relatedImageEnv *RelatedImageEnv) AsYamlObject() map[string]interface{} {
	return map[string]interface{}{
		"name":  relatedImageEnv.Name(),
		"image": relatedImageEnv.Image(),
	}
}

// NewRelatedImageEnv returns a new related iamge env pullspec.
func NewRelatedImageEnv(data map[string]interface{}) *RelatedImageEnv {
	return &RelatedImageEnv{
		namedPullSpec: namedPullSpec{
			imageKey: "value",
			data:     data,
		},
	}
}

// Annotation is a pullspec representing images in the annotation field of
// kubernetes objects.
type Annotation struct {
	namedPullSpec
	startI, endI int
}

// NewAnnotation returns a new annotation pullspec.
func NewAnnotation(data map[string]interface{}, key string, startI, endI int) *Annotation {
	return &Annotation{
		namedPullSpec: namedPullSpec{
			imageKey: key,
			data:     data,
		},
		startI: startI,
		endI:   endI,
	}
}

// Image returns the image string of the pullspec.
func (annotation *Annotation) Image() string {
	i, j := annotation.startI, annotation.endI
	text := fmt.Sprintf("%v", annotation.data[annotation.imageKey])
	return text[i:j]
}

// String returns a string representation of the pullspec.
func (annotation *Annotation) String() string {
	return fmt.Sprintf("annotation %s", annotation.Name())
}

// SetImage will replace the image string with the provided image string.
func (annotation *Annotation) SetImage(image string) {
	i, j := annotation.startI, annotation.endI
	text := fmt.Sprintf("%v", annotation.data[annotation.imageKey])
	annotation.data[annotation.imageKey] = fmt.Sprintf("%v%s%v", text[:i], image, text[j:])
}

// Name returns the name of the pullspec.
func (annotation *Annotation) Name() string {
	image := imagename.Parse(annotation.Image())
	tag := image.Tag

	if strings.HasPrefix(tag, "sha256") {
		tag = tag[len("sha256:"):]
	}
	return fmt.Sprintf("%s-%s-annotation", image.Repo, tag)
}

// AsYamlObject returns the annotation pullspec as a map[string]interface{}.
func (annotation *Annotation) AsYamlObject() map[string]interface{} {
	return map[string]interface{}{
		"name":  annotation.Name(),
		"image": annotation.Image(),
	}
}

// OperatorCSV represents the CSV data and holds information
// regarding how to parse the image strings.
type OperatorCSV struct {
	path              string
	data              unstructured.Unstructured
	pullspecHeuristic Heuristic
}

// NewOperatorCSV creates a OperatorCSV using the data provided via an unstructured kubernetes object.
func NewOperatorCSV(path string, data *unstructured.Unstructured, pullSpecHeuristic Heuristic) (*OperatorCSV, error) {
	if data.GetKind() != operatorCsvKind {
		return nil, utils.ErrNotClusterServiceVersion
	}

	if pullSpecHeuristic == nil {
		pullSpecHeuristic = DefaultHeuristic
	}

	return &OperatorCSV{
		data:              *data,
		path:              path,
		pullspecHeuristic: pullSpecHeuristic,
	}, nil
}

const (
	operatorCsvKind = "ClusterServiceVersion"
)

var ()

// FromDirectory creates a NewOperatorCSV from the directory path provided.
func FromDirectory(path string, pullSpecHeuristic Heuristic) ([]*OperatorCSV, error) {
	operatorCSVs := []*OperatorCSV{}

	stat, err := os.Stat(path)

	if err != nil {
		if os.IsNotExist(err) {
			return nil, utils.NewErrIsNotDirectoryOrDoesNotExist(path)
		}

		return nil, err
	}

	if !stat.IsDir() {
		return nil, utils.NewErrIsNotDirectoryOrDoesNotExist(path)
	}

	err = filepath.Walk(path, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			log.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}

		log.Println(info.Name(), info.IsDir())

		if info.IsDir() ||
			!(strings.HasSuffix(info.Name(), ".yaml") || strings.HasSuffix(info.Name(), ".yml")) {
			log.Printf("skipping non-yaml file without errors: %+v \n", info.Name())
			return nil
		}

		log.Printf("visited file or dir: %q\n", path)
		csv, err := NewOperatorCSVFromFile(path, pullSpecHeuristic)

		if err != nil && errors.Is(err, utils.ErrNotClusterServiceVersion) {
			log.Printf("skipping file because it's not a ClusterServiceVersion: %+v \n", info.Name())
			return nil
		}

		if err != nil {
			log.Printf("failure reading the file: %+v \n", info.Name())
			return err
		}

		operatorCSVs = append(operatorCSVs, csv)
		return nil
	})

	if err != nil {
		log.Printf("failure walking the directory: %+v \n", err)
		return nil, err
	}

	if len(operatorCSVs) > 1 {
		log.Printf("found too many csvs in the directory")
		return nil, utils.ErrTooManyCSVs
	}

	if len(operatorCSVs) == 0 {
		log.Printf("failure to find operator manifests in the directory")
		return nil, utils.ErrNoOperatorManifests
	}

	return operatorCSVs, nil
}

// NewOperatorCSVFromFile creates a NewOperatorCSV from a filepath.
func NewOperatorCSVFromFile(
	path string,
	pullSpecHeuristic Heuristic,
) (*OperatorCSV, error) {
	data := &unstructured.Unstructured{}

	fileData, err := os.ReadFile(path)

	if err != nil {
		return nil, err
	}

	// decode YAML into unstructured.Unstructured
	dec := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	_, _, err = dec.Decode(fileData, nil, data)

	if err != nil {
		return nil, err
	}

	csv, err := NewOperatorCSV(path, data, pullSpecHeuristic)

	if err != nil {
		return nil, err
	}

	return csv, nil
}

// ToYaml will write the OperatorCSV to yaml string and return the bytes.
func (csv *OperatorCSV) ToYaml() ([]byte, error) {
	buff := bytes.Buffer{}

	enc := yamlv3.NewEncoder(&buff)
	enc.SetIndent(2)

	err := enc.Encode(&csv.data.Object)
	if err != nil {
		return nil, err
	}

	return buff.Bytes(), nil
}

// Dump will dump the csv yaml to a writer if provided or
// the file the OperatorCSV started from if the filesystem is writable.
func (csv *OperatorCSV) Dump(writer io.Writer) error {
	if writer == nil {
		f, err := os.OpenFile(csv.path, os.O_TRUNC|os.O_WRONLY, 0755)
		if err != nil {
			return err
		}
		defer closeFile(f)

		writer = f
	}

	b, err := csv.ToYaml()

	if err != nil {
		return err
	}

	_, err = writer.Write(b)

	if err != nil {
		return err
	}

	return nil
}

// HasRelatedImages returns true with the CSV has RelatedImage pullspecs.
func (csv *OperatorCSV) HasRelatedImages() bool {
	pullSpecs, _ := csv.relatedImagePullSpecs()
	return len(pullSpecs) != 0
}

// HasRelatedImageEnvs returns true with the CSV has RelatedImageEnv pullspecs.
func (csv *OperatorCSV) HasRelatedImageEnvs() bool {
	pullSpecs, _ := csv.relatedImageEnvPullSpecs()
	return len(pullSpecs) > 0
}

// GetPullSpecs will return a list of all the images found in via pullspecs.
func (csv *OperatorCSV) GetPullSpecs() ([]*imagename.ImageName, error) {
	pullspecs := make(map[imagename.ImageName]interface{})

	namedList, err := csv.namedPullSpecs()

	if err != nil {
		return nil, err
	}

	for i := range namedList {
		ps := namedList[i]
		log.Printf("Found pullspec for %s: %s", ps.String(), ps.Image())
		image := imagename.Parse(ps.Image())
		pullspecs[*image] = nil
	}

	imageList := make([]*imagename.ImageName, 0, len(pullspecs))

	for key := range pullspecs {
		localKey := key
		imageList = append(imageList, &localKey)
	}

	return imageList, nil
}

// ReplacePullSpecs will replace each pullspec found with the provide image.
func (csv *OperatorCSV) ReplacePullSpecs(replacement map[imagename.ImageName]imagename.ImageName) error {
	pullspecs, err := csv.namedPullSpecs()
	if err != nil {
		return err
	}

	for _, pullspec := range pullspecs {
		old := imagename.Parse(pullspec.Image())
		new, ok := replacement[*old]

		if ok && *old != new {
			log.Printf("%s - Replaced pullspec for %s: %s -> %s", csv.path, pullspec.String(), *old, new)
			pullspec.SetImage(new.String())
		}
	}

	return nil
}

// ReplacePullSpecsEverywhere will replace image values in each pullspec throughout the entire OperatorCSV.
func (csv *OperatorCSV) ReplacePullSpecsEverywhere(replacement map[imagename.ImageName]imagename.ImageName) error {
	err := csv.ReplacePullSpecs(replacement)

	if err != nil {
		return err
	}

	pullspecs := []NamedPullSpec{}
	annotationPullSpecs, err := csv.annotationPullSpecs(knownAnnotationKeys)

	if err != nil {
		return err
	}

	guessedAnnotationPullSpecs, err := csv.annotationPullSpecs(nil)

	if err != nil {
		return err
	}

	pullspecs = append(pullspecs, annotationPullSpecs...)
	pullspecs = append(pullspecs, guessedAnnotationPullSpecs...)

	err = csv.findPotentialPullSpecsNotInAnnotations(csv.data.Object, &pullspecs)

	if err != nil {
		return err
	}

	for _, pullspec := range pullspecs {
		old := imagename.Parse(pullspec.Image())
		new, ok := replacement[*old]

		if ok && *old != new {
			log.Printf("%s - Replaced pullspec for %s: %s -> %s", csv.path, pullspec.String(), *old, new)
			pullspec.SetImage(new.String())
		}
	}

	return nil
}

// SetRelatedImages will set the related images fields based on the CSV pullspecs discovered.
func (csv *OperatorCSV) SetRelatedImages() error {
	namedPullspecs, err := csv.namedPullSpecs()

	if err != nil {
		return err
	}

	if len(namedPullspecs) == 0 {
		return nil
	}

	conflicts := []string{}
	byName := map[string]NamedPullSpec{}
	for _, newPull := range namedPullspecs {
		old, ok := byName[newPull.Name()]

		if !ok {
			byName[newPull.Name()] = newPull
			continue
		}

		if old.Image() == newPull.Image() {
			continue
		}

		conflicts = append(conflicts, fmt.Sprintf("%s: %s X %s: %s",
			old.String(), old.Image(), newPull.String(), newPull.Image()))
	}

	if len(conflicts) > 0 {
		return fmt.Errorf("%s - Found conflicts when setting relatedImages:\n%s", csv.path, strings.Join(conflicts, "\n"))
	}

	relatedImages := []map[string]interface{}{}

	for _, p := range byName {
		log.Printf("%s - Set relateImage %s (from %s): %s\n", csv.path, p.Name(), p.String(), p.Image())
		relatedImages = append(relatedImages, p.AsYamlObject())
	}

	spec, ok := csv.data.Object["spec"]
	if !ok {
		spec = map[string]interface{}{
			"relatedImages": relatedImages,
		}
		csv.data.Object["spec"] = spec
	} else {
		spec.(map[string]interface{})["relatedImages"] = relatedImages
	}

	return nil
}

var knownAnnotationKeys = stringSlice{"containerImage"}

func (csv *OperatorCSV) namedPullSpecs() ([]NamedPullSpec, error) {
	pullspecs := []NamedPullSpec{}

	relatedImages, err := csv.relatedImagePullSpecs()

	if err != nil {
		return pullspecs, err
	}

	containers, err := csv.containerPullSpecs()

	if err != nil {
		return pullspecs, err
	}

	initContainers, err := csv.initContainerPullSpecs()

	if err != nil {
		return pullspecs, err
	}

	relatedImageEnvPullSpecs, err := csv.relatedImageEnvPullSpecs()

	if err != nil {
		return pullspecs, err
	}

	annotationPullSpecs, err := csv.annotationPullSpecs(knownAnnotationKeys)

	if err != nil {
		return pullspecs, err
	}

	guessedAnnotationPullSpecs, err := csv.annotationPullSpecs(nil)

	if err != nil {
		return pullspecs, err
	}

	pullspecs = append(pullspecs, relatedImages...)
	pullspecs = append(pullspecs, containers...)
	pullspecs = append(pullspecs, initContainers...)
	pullspecs = append(pullspecs, relatedImageEnvPullSpecs...)
	pullspecs = append(pullspecs, annotationPullSpecs...)
	pullspecs = append(pullspecs, guessedAnnotationPullSpecs...)

	return pullspecs, nil
}

var relatedImagesLens = utils.Lens().M("spec").M("relatedImages").Build()

func (csv *OperatorCSV) relatedImagePullSpecs() ([]NamedPullSpec, error) {
	lookupResultSlice, err := relatedImagesLens.L(csv.data.Object)

	if err != nil {
		if errors.Is(err, utils.ErrNotFound) {
			return []NamedPullSpec{}, nil
		}

		return nil, err
	}

	pullspecs := make([]NamedPullSpec, 0, len(lookupResultSlice))

	for i := range lookupResultSlice {
		data := lookupResultSlice[i]

		pullspec, err := NewRelatedImage(data)

		if err != nil {
			return nil, err
		}

		pullspecs = append(pullspecs, pullspec)
	}

	return pullspecs, nil
}

func (csv *OperatorCSV) relatedImageEnvPullspecs() ([][]int, error) {
	return nil, nil
}

var deploymentLens = utils.Lens().M("spec").M("install").M("spec").M("deployments").Build()

func (csv *OperatorCSV) deployments() ([]interface{}, error) {
	return deploymentLens.L(csv.data.Object)
}

var initContainerLens = utils.Lens().M("spec").M("template").M("spec").M("initContainers").Build()

func (csv *OperatorCSV) initContainerPullSpecs() ([]NamedPullSpec, error) {
	deployments, err := csv.deployments()

	if err != nil {
		return nil, err
	}

	pullspecs := make([]NamedPullSpec, 0)

	for i := range deployments {
		lookupResultSlice, err := initContainerLens.L(deployments[i])
		if err != nil {
			if errors.Is(err, utils.ErrNotFound) {
				continue
			}

			return nil, err
		}

		for i := range lookupResultSlice {
			data := lookupResultSlice[i]

			pullspec, err := NewInitContainer(data)
			if err != nil {
				return nil, err
			}

			pullspecs = append(pullspecs, pullspec)
		}
	}

	return pullspecs, nil
}

var containerLens = utils.Lens().M("spec").M("template").M("spec").M("containers").Build()

func (csv *OperatorCSV) containerPullSpecs() ([]NamedPullSpec, error) {
	deployments, err := csv.deployments()

	if err != nil {
		return nil, err
	}

	pullspecs := make([]NamedPullSpec, 0)

	for i := range deployments {
		lookupResultSlice, err := containerLens.L(deployments[i])

		if err != nil {
			if errors.Is(err, utils.ErrNotFound) {
				continue
			}

			return nil, err
		}

		for i := range lookupResultSlice {
			data := lookupResultSlice[i]

			pullspec, err := NewContainer(data)

			if err != nil {
				return nil, err
			}

			pullspecs = append(pullspecs, pullspec)
		}
	}

	return pullspecs, nil
}

func (csv *OperatorCSV) relatedImageEnvPullSpecs() ([]NamedPullSpec, error) {
	containers, err := csv.containerPullSpecs()

	if err != nil {
		return nil, err
	}

	initContainers, err := csv.initContainerPullSpecs()

	if err != nil {
		return nil, err
	}

	allContainers := append(containers, initContainers...)

	relatedImageEnvs := []NamedPullSpec{}

	for i := range allContainers {
		c := allContainers[i].Data()

		env, ok := c["env"]

		if !ok {
			continue
		}

		envMaps, ok := env.([]interface{})
		if !ok {
			return nil, errors.New("expected type slice")
		}

		for j := range envMaps {
			envMap, ok := envMaps[j].(map[string]interface{})

			if !ok {
				return nil, errors.New("expected type map")
			}

			// only look at RELATED_IMAGE env vars
			if name, ok := envMap["name"]; !(ok && strings.HasPrefix(name.(string), "RELATED_IMAGE_")) {
				continue
			}

			if _, hasValueFrom := envMap["valueFrom"]; hasValueFrom {
				return nil, utils.NewError(nil, `%s: "valueFrom" references are not supported`, envMap["name"])
			}

			ps := NewRelatedImageEnv(envMap)
			relatedImageEnvs = append(relatedImageEnvs, ps)
		}
	}

	return relatedImageEnvs, nil
}

func (csv *OperatorCSV) annotationPullSpecs(keyFilter stringSlice) ([]NamedPullSpec, error) {
	pullSpecs := []NamedPullSpec{}

	annotationObjects, err := csv.findAllAnnotations()

	if err != nil {
		return nil, err
	}

	for i := range annotationObjects {
		obj := annotationObjects[i]
		for rKey := range obj {
			key := rKey
			val := obj[key]

			if keyFilter != nil && !keyFilter.Contains(key) {
				continue
			}

			valStr := fmt.Sprintf("%v", val)
			results := csv.pullspecHeuristic(valStr)

			for j := range results {
				ii, jj := results[j][0], results[j][1]
				pullSpecs = append(pullSpecs, NewAnnotation(obj, key, ii, jj))
			}
		}
	}

	return namedPullSpecSlice(pullSpecs).Reverse(), nil
}

var (
	csvAnnotations         = utils.Lens().M("metadata").M("annotations").Build()
	deploymentAnnotations  = utils.Lens().M("spec").M("template").M("metadata").M("annotations").Build()
	deploymentsAnnotations = utils.Lens().
				M("spec").M("install").M("spec").M("deployments").
				Apply(deploymentAnnotations).
				Build()
)

func (csv *OperatorCSV) findAllAnnotations() ([]map[string]interface{}, error) {
	findAnnotationMaps := []func() (map[string]interface{}, error){
		csvAnnotations.MFunc(csv.data.Object),
	}

	findAnnotationSlices := []func() ([]interface{}, error){
		deploymentsAnnotations.LFunc(csv.data.Object),
		func() ([]interface{}, error) {
			results := []interface{}{}
			err := csv.findRandomCSVAnnotations(csv.data.Object, &results, false)
			return results, err
		},
	}

	annotations := []map[string]interface{}{}

	for _, findAnnotation := range findAnnotationMaps {
		result, err := findAnnotation()

		if err != nil {
			if errors.Is(err, utils.ErrNotFound) {
				continue
			}
			return nil, err
		}

		annotations = append(annotations, result)
	}

	for _, findAnnotation := range findAnnotationSlices {
		results, err := findAnnotation()

		if err != nil {
			if errors.Is(err, utils.ErrNotFound) {
				continue
			}
			return nil, err
		}

		for _, result := range results {
			annotationResult := result.(map[string]interface{})
			annotations = append(annotations, annotationResult)
		}
	}

	return annotations, nil
}

var annotations = utils.Lens().M("metadata").M("annotations").Build()

func (csv *OperatorCSV) findRandomCSVAnnotations(root map[string]interface{}, results *[]interface{}, underMetadata bool) error {
	annos, err := annotations.M(root)

	if err != nil && !errors.Is(err, utils.ErrNotFound) {
		return err
	}

	if err == nil && len(annos) != 0 {
		*results = append(*results, annos)
	}

	for key := range root {
		isUnderMetadata := false

		if key == "metadata" {
			if underMetadata {
				continue
			}

			isUnderMetadata = true
		}

		if slicev, ok := root[key].([]interface{}); ok {

			for i := range slicev {
				if datav, ok := slicev[i].(map[string]interface{}); ok {
					err := csv.findRandomCSVAnnotations(datav, results, isUnderMetadata)

					if err != nil {
						return err
					}
				}
			}
		}

		if datav, ok := root[key].(map[string]interface{}); ok {
			err := csv.findRandomCSVAnnotations(datav, results, isUnderMetadata)

			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (csv *OperatorCSV) findPotentialPullSpecsNotInAnnotations(root map[string]interface{}, specs *[]NamedPullSpec) error {
	for rKey := range root {
		key := rKey
		val := root[key]

		valStr := fmt.Sprintf("%v", val)
		results := csv.pullspecHeuristic(valStr)

		for j := range results {
			ii, jj := results[j][0], results[j][1]
			*specs = append(*specs, NewAnnotation(root, key, ii, jj))
		}
	}

	for key := range root {
		if key == "metadata" {
			continue
		}

		if slicev, ok := root[key].([]interface{}); ok {

			for i := range slicev {
				if datav, ok := slicev[i].(map[string]interface{}); ok {
					err := csv.findPotentialPullSpecsNotInAnnotations(datav, specs)

					if err != nil {
						return err
					}
				}
			}
		}

		if datav, ok := root[key].(map[string]interface{}); ok {
			err := csv.findPotentialPullSpecsNotInAnnotations(datav, specs)

			if err != nil {
				return err
			}
		}
	}

	return nil
}

type stringSlice []string

func (l stringSlice) Contains(in string) bool {
	for _, key := range l {
		if key == in {
			return true
		}
	}
	return false
}

type namedPullSpecSlice []NamedPullSpec

func (n namedPullSpecSlice) Reverse() namedPullSpecSlice {
	for i := 0; i < len(n)/2; i++ {
		j := len(n) - i - 1
		n[i], n[j] = n[j], n[i]
	}
	return n
}

func closeFile(f *os.File) {
	err := f.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
