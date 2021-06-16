package releases

import (
	"fmt"
	"os"
	"sort"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/blang/semver"
	"github.com/gobwas/glob"
	"k8s.io/apimachinery/pkg/runtime/schema"

	cnao "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/shared"
	. "github.com/kubevirt/cluster-network-addons-operator/test/kubectl"
	. "github.com/kubevirt/cluster-network-addons-operator/test/okd"
	. "github.com/kubevirt/cluster-network-addons-operator/test/operations"
)

type Release struct {
	// Release version
	Version string
	// Containers and their images for given release
	Containers []cnao.Container
	// SupportedSpec for given release should be upgradable
	SupportedSpec cnao.NetworkAddonsConfigSpec
	// Manifest that can be used to install the operator in given release
	Manifests []string
	// CrdCleanUp is used to uninstall CRDs between upgrade tests
	CrdCleanUp []string
}

// Releases are populated by respective release modules using init()
var releases = []Release{}
var releasesProcessed = false

// Returns list of releases sorted from oldest to newest
func Releases() []Release {
	if releasesProcessed {
		return releases
	}

	// Keep only releases matching the selector
	if releasesSelectorRaw, found := os.LookupEnv("RELEASES_SELECTOR"); found {
		releasesSelector := glob.MustCompile(releasesSelectorRaw)
		filteredReleases := []Release{}

		for _, release := range releases {
			if releasesSelector.Match(release.Version) {
				filteredReleases = append(filteredReleases, release)
			}
		}

		releases = filteredReleases
	}

	// Drop all releases matching the selector
	if releasesDeselectorRaw, found := os.LookupEnv("RELEASES_DESELECTOR"); found {
		releasesDeselector := glob.MustCompile(releasesDeselectorRaw)

		filteredReleases := []Release{}

		for _, release := range releases {
			if !releasesDeselector.Match(release.Version) {
				filteredReleases = append(filteredReleases, release)
			}
		}

		releases = filteredReleases
	}

	// Sort releases in ascending order
	sort.Slice(releases, func(a, b int) bool {
		releaseAVersion, err := semver.Make(releases[a].Version)
		if err != nil {
			panic(err)
		}
		releaseBVersion, err := semver.Make(releases[b].Version)
		if err != nil {
			panic(err)
		}
		return releaseAVersion.LT(releaseBVersion)
	})

	releasesProcessed = true

	releases = append(releases, releases...)
	releases = append(releases, releases...)
	return releases
}

// Iterates registered releases and returns the latest (master) based on semver
func LatestRelease() Release {
	r := Releases()
	return r[len(r)-1]
}

// Installs given release (CRD, RBAC and Deployment)
func InstallRelease(release Release) {
	By(fmt.Sprintf("Installing release %s", release.Version))
	for _, manifestName := range release.Manifests {
		out, err := Kubectl("apply", "-f", "_out/cluster-network-addons/"+release.Version+"/"+manifestName)
		Expect(err).NotTo(HaveOccurred(), out)
	}
}

// Removes given release from cluster
func UninstallRelease(release Release) {
	By(fmt.Sprintf("Uninstalling release %s", release.Version))
	for _, manifestName := range release.Manifests {
		out, err := Kubectl("delete", "--ignore-not-found", "-f", "_out/cluster-network-addons/"+release.Version+"/"+manifestName)
		Expect(err).NotTo(HaveOccurred(), out)
	}

	for _, crdInstance := range release.CrdCleanUp {
		out, err := Kubectl("delete", "crd", "--ignore-not-found", crdInstance)
		Expect(err).NotTo(HaveOccurred(), out)
	}

}

// Installs given release (RBAC and Deployment)
func InstallOperator(release Release) {
	manifestName := "operator.yaml"
	By(fmt.Sprintf("Installing operator %s", release.Version))
	out, err := Kubectl("apply", "-f", "_out/cluster-network-addons/"+release.Version+"/"+manifestName)
	Expect(err).NotTo(HaveOccurred(), out)
}

// Removes given release from cluster
func UninstallOperator(release Release) {
	manifestName := "operator.yaml"
	By(fmt.Sprintf("Uninstalling operator %s", release.Version))
	out, err := Kubectl("delete", "--ignore-not-found", "-f", "_out/cluster-network-addons/"+release.Version+"/"+manifestName)
	Expect(err).NotTo(HaveOccurred(), out)
}

// Make sure that container images currently used (reported in NetworkAddonsConfig)
// are matching images expected for given release
func CheckReleaseUsesExpectedContainerImages(gvk schema.GroupVersionKind, release Release) {
	By(fmt.Sprintf("Checking that all deployed images match release %s", release.Version))

	expectedContainers := sortContainers(release.Containers)
	if IsOnOKDCluster() {
		// On OpenShift 4, Multus is not owned by us and will not be reported in Status
		expectedContainers = dropMultusContainers(expectedContainers)
	}

	configStatus := GetConfigStatus(gvk)
	deployedContainers := sortContainers(configStatus.Containers)

	Expect(deployedContainers).To(ConsistOf(expectedContainers))
}

func sortContainers(containers []cnao.Container) []cnao.Container {
	sort.Slice(containers, func(a, b int) bool {
		return (sort.StringsAreSorted([]string{containers[a].ParentKind, containers[b].ParentKind}) &&
			sort.StringsAreSorted([]string{containers[a].ParentName, containers[b].ParentName}) &&
			sort.StringsAreSorted([]string{containers[a].Name, containers[b].Name}))
	})
	return containers
}

func dropMultusContainers(containers []cnao.Container) []cnao.Container {
	filteredContainers := []cnao.Container{}
	for _, container := range containers {
		if !strings.Contains(container.Name, "multus") {
			filteredContainers = append(filteredContainers, container)
		}
	}
	return filteredContainers
}
