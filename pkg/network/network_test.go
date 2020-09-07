package network

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"fmt"

	osv1 "github.com/openshift/api/operator/v1"
	v1 "k8s.io/api/core/v1"

	cnao "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/shared"
)

var _ = Describe("Testing network", func() {
	// There is no functionality in this function yet
	Describe("Canonicalize", func() {
		Context("when given an empty config", func() {
			conf := &cnao.NetworkAddonsConfigSpec{}
			It("should pass", func() {
				Canonicalize(conf)
			})
		})
	})

	Describe("Validate", func() {
		Context("when given a valid config", func() {
			conf := &cnao.NetworkAddonsConfigSpec{}
			openshiftNetworkConf := &osv1.Network{}
			It("should pass", func() {
				err := Validate(conf, openshiftNetworkConf)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when given invalid config", func() {
			conf := &cnao.NetworkAddonsConfigSpec{
				KubeMacPool: &cnao.KubeMacPool{
					RangeStart: "foo",
				},
				ImagePullPolicy: v1.PullPolicy("bar"),
			}
			openshiftNetworkConf := &osv1.Network{}
			It("should return a compilation of errors", func() {
				err := Validate(conf, openshiftNetworkConf)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("invalid configuration"))
				Expect(err.Error()).To(ContainSubstring("both or none of the KubeMacPool ranges needs to be configured"))
				Expect(err.Error()).To(ContainSubstring("requested imagePullPolicy 'bar' is not valid"))
			})
		})
	})

	// TODO: Mock rand.Read to fail and test error handling here
	Describe("FillDefaults", func() {
		Context("when given valid configuration", func() {
			newConf := &cnao.NetworkAddonsConfigSpec{}
			prevConfig := &cnao.NetworkAddonsConfigSpec{}

			It("should successfully pass", func() {
				err := FillDefaults(newConf, prevConfig)
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("IsChangeSafe", func() {
		Context("when current and new configuration is compatible", func() {
			newConf := &cnao.NetworkAddonsConfigSpec{}
			prevConfig := &cnao.NetworkAddonsConfigSpec{}

			It("should pass the check", func() {
				err := IsChangeSafe(prevConfig, newConf)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when current and new configuration is not compatible", func() {
			newConf := &cnao.NetworkAddonsConfigSpec{ImagePullPolicy: v1.PullAlways}
			prevConfig := &cnao.NetworkAddonsConfigSpec{ImagePullPolicy: v1.PullIfNotPresent}

			It("should fail the check", func() {
				err := IsChangeSafe(prevConfig, newConf)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("Render", func() {
		Context("when given valid arguments", func() {
			conf := &cnao.NetworkAddonsConfigSpec{ImagePullPolicy: v1.PullAlways, Multus: &cnao.Multus{}, LinuxBridge: &cnao.LinuxBridge{}, PlacementConfiguration: &cnao.PlacementConfiguration{Workloads: &cnao.Placement{}}}
			manifestDir := "../../data"
			openshiftNetworkConf := &osv1.Network{}
			clusterInfo := &ClusterInfo{SCCAvailable: true, OpenShift4: false}

			It("should successfully render a set of objects", func() {
				objs, err := Render(conf, manifestDir, openshiftNetworkConf, clusterInfo)
				Expect(err).NotTo(HaveOccurred())
				Expect(objs).NotTo(BeEmpty())
			})
		})

		Context("when given manifest directory that does not contain all expected components", func() {
			conf := &cnao.NetworkAddonsConfigSpec{ImagePullPolicy: v1.PullAlways, Multus: &cnao.Multus{}, LinuxBridge: &cnao.LinuxBridge{}, PlacementConfiguration: &cnao.PlacementConfiguration{Workloads: &cnao.Placement{}}}
			manifestDir := "." // Test directory with this module
			openshiftNetworkConf := &osv1.Network{}
			clusterInfo := &ClusterInfo{SCCAvailable: true, OpenShift4: false}

			It("should return an error since it's unable to load templates for render", func() {
				_, err := Render(conf, manifestDir, openshiftNetworkConf, clusterInfo)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("errorListToMultiLineString", func() {
		Context("when given no error", func() {
			errs := []error{}
			It("should return an empty string", func() {
				out := errorListToMultiLineString(errs)
				Expect(out).To(BeEmpty())
			})
		})

		Context("when single error", func() {
			errs := []error{fmt.Errorf("foo")}
			It("should return a single line with the error", func() {
				out := errorListToMultiLineString(errs)
				Expect(out).To(Equal("foo"))
			})
		})

		Context("when given multiple errors", func() {
			errs := []error{fmt.Errorf("foo"), fmt.Errorf("bar")}
			It("should return one line per each error", func() {
				out := errorListToMultiLineString(errs)
				Expect(out).To(Equal("foo\nbar"))
			})
		})
	})
})
