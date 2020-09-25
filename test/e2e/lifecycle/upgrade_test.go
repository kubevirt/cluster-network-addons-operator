package test

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"

	. "github.com/kubevirt/cluster-network-addons-operator/test/check"
	. "github.com/kubevirt/cluster-network-addons-operator/test/operations"
	. "github.com/kubevirt/cluster-network-addons-operator/test/releases"
)

const podsDeploymentTimeout = 20 * time.Minute

var _ = Context("Cluster Network Addons Operator", func() {
	testUpgrade := func(oldRelease, newRelease Release) {
		Context(fmt.Sprintf("when operator in version %s is installed and supported spec configured", oldRelease.Version), func() {
			BeforeEach(func() {
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
					UninstallRelease(oldRelease)
					InstallRelease(newRelease)
					UpdateConfig(newReleaseGvk, newRelease.SupportedSpec)
					CheckOperatorIsReady(podsDeploymentTimeout)

					// Check that operator and target versions will be set to the newer.
					expectedOperatorVersion := newRelease.Version
					expectedObservedVersion := newRelease.Version
					expectedTargetVersion := newRelease.Version
					CheckConfigVersions(newReleaseGvk, expectedOperatorVersion, expectedObservedVersion, expectedTargetVersion, podsDeploymentTimeout, CheckDoNotRepeat)
				})

				It("it should report expected deployed container images and leave no leftovers from the previous version", func() {
					By("Checking reported container images")
					CheckReleaseUsesExpectedContainerImages(newReleaseGvk, newRelease)

					By("Checking for leftover objects from the previous version")
					CheckForLeftoverObjects(newRelease.Version)
				})
			})

			It(fmt.Sprintf("should transition reported versions while being upgraded to version %s", newRelease.Version), func() {
				newReleaseGvk := GetCnaoV1GroupVersionKind()
				// Upgrade the operator
				UninstallRelease(oldRelease)
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

	// Run tests upgrading from each released version to the latest/master
	releases := Releases()
	for _, oldRelease := range releases[:len(releases)-1] {
		testUpgrade(oldRelease, LatestRelease())
	}
})
