package test

import (
	"fmt"
	"runtime"
	"time"

	. "github.com/onsi/ginkgo/v2"

	. "github.com/kubevirt/cluster-network-addons-operator/test/check"
	. "github.com/kubevirt/cluster-network-addons-operator/test/operations"
	. "github.com/kubevirt/cluster-network-addons-operator/test/releases"
)

const podsDeploymentTimeout = 20 * time.Minute

// CNAO supports multiarch from this release
const multiArchStartRelease = "0.98.2"

var _ = Context("Cluster Network Addons Operator", func() {
	testUpgrade := func(oldRelease, newRelease Release) {
		Context(fmt.Sprintf("when operator in version %s is installed and supported spec configured", oldRelease.Version), func() {
			BeforeEach(func() {
				UninstallRelease(newRelease)
				oldReleaseGvk := GetCnaoV1alpha1GroupVersionKind()
				InstallRelease(oldRelease)
				CheckOperatorIsReady(podsDeploymentTimeout)
				CreateConfig(oldReleaseGvk, oldRelease.SupportedSpec)
				CheckConfigCondition(oldReleaseGvk, ConditionAvailable, ConditionTrue, 15*time.Minute, CheckDoNotRepeat)
				CheckReleaseUsesExpectedContainerImages(oldReleaseGvk, oldRelease)
				expectedOperatorVersion := oldRelease.Version
				expectedObservedVersion := oldRelease.Version
				expectedTargetVersion := oldRelease.Version
				CheckConfigVersions(oldReleaseGvk, expectedOperatorVersion, expectedObservedVersion, expectedTargetVersion, CheckImmediately, CheckDoNotRepeat)
			})

			Context("and it is upgraded to the latest release", func() {
				newReleaseGvk := GetCnaoV1GroupVersionKind()
				BeforeEach(func() {
					InstallRelease(newRelease)
					UpdateConfig(newReleaseGvk, newRelease.SupportedSpec)
					CheckOperatorIsReady(podsDeploymentTimeout)

					// Check that operator and target versions will be set to the newer.
					expectedOperatorVersion := newRelease.Version
					expectedObservedVersion := newRelease.Version
					expectedTargetVersion := newRelease.Version
					CheckConfigVersions(newReleaseGvk, expectedOperatorVersion, expectedObservedVersion, expectedTargetVersion, podsDeploymentTimeout, CheckDoNotRepeat)
				})

				It("Should be able to explain the crd object", func() {
					CheckCrdExplainable()
				})

				It("it should report expected deployed container images and leave no leftovers from the previous version", func() {
					By("Checking reported container images")
					CheckReleaseUsesExpectedContainerImages(newReleaseGvk, newRelease)

					By("Checking for leftover objects from the previous version")
					CheckForLeftoverObjects(newRelease.Version)
					CheckForLeftoverLabels()
				})
			})

			It(fmt.Sprintf("should transition reported versions while being upgraded to version %s", newRelease.Version), func() {
				newReleaseGvk := GetCnaoV1GroupVersionKind()
				// Upgrade the operator
				InstallRelease(newRelease)

				// Check that operator and target versions will be set to the newer. Ignore observed version, since it
				// might reach the target state immediately when no changes are needed between two releases
				expectedOperatorVersion := newRelease.Version
				expectedObservedVersion := CheckIgnoreVersion
				expectedTargetVersion := newRelease.Version
				CheckConfigVersions(newReleaseGvk, expectedOperatorVersion, expectedObservedVersion, expectedTargetVersion, podsDeploymentTimeout, CheckDoNotRepeat)

				// Wait until the operator finishes configuration
				CheckConfigCondition(newReleaseGvk, ConditionAvailable, ConditionTrue, podsDeploymentTimeout, CheckDoNotRepeat)

				// Validate that observed version turned to the newer
				expectedOperatorVersion = newRelease.Version
				expectedObservedVersion = newRelease.Version
				expectedTargetVersion = newRelease.Version
				CheckConfigVersions(newReleaseGvk, expectedOperatorVersion, expectedObservedVersion, expectedTargetVersion, CheckImmediately, CheckDoNotRepeat)
			})
		})
	}

	// Run tests upgrading from each released version to the latest/main
	releases := Releases()

	start := 0
	if runtime.GOARCH == "s390x" {
		for index, release := range releases {
			if release.Version == multiArchStartRelease {
				start = index
				break
			}
		}
	}

	for _, oldRelease := range releases[start : len(releases)-1] {
		testUpgrade(oldRelease, LatestRelease())
	}
})
