package network

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"fmt"

	osv1 "github.com/openshift/api/operator/v1"
	v1 "k8s.io/api/core/v1"

	opv1alpha1 "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/v1alpha1"
)

var _ = Describe("Testing network", func() {
	// There is no functionality in this function yet
	Describe("Canonicalize", func() {
		Context("when given an empty config", func() {
			conf := &opv1alpha1.NetworkAddonsConfigSpec{}
			It("should pass", func() {
				Canonicalize(conf)
			})
		})
	})

	Describe("Validate", func() {
		Context("when given a valid config", func() {
			conf := &opv1alpha1.NetworkAddonsConfigSpec{}
			openshiftNetworkConf := &osv1.Network{}
			It("should pass", func() {
				err := Validate(conf, openshiftNetworkConf)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when given invalid config", func() {
			conf := &opv1alpha1.NetworkAddonsConfigSpec{
				KubeMacPool: &opv1alpha1.KubeMacPool{
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
			newConf := &opv1alpha1.NetworkAddonsConfigSpec{}
			prevConfig := &opv1alpha1.NetworkAddonsConfigSpec{}

			It("should successfully pass", func() {
				err := FillDefaults(newConf, prevConfig)
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("IsChangeSafe", func() {
		Context("when current and new configuration is compatible", func() {
			newConf := &opv1alpha1.NetworkAddonsConfigSpec{}
			prevConfig := &opv1alpha1.NetworkAddonsConfigSpec{}

			It("should pass the check", func() {
				err := IsChangeSafe(prevConfig, newConf)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when current and new configuration is not compatible", func() {
			newConf := &opv1alpha1.NetworkAddonsConfigSpec{ImagePullPolicy: v1.PullAlways}
			prevConfig := &opv1alpha1.NetworkAddonsConfigSpec{ImagePullPolicy: v1.PullIfNotPresent}

			It("should fail the check", func() {
				err := IsChangeSafe(prevConfig, newConf)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("Render", func() {
		Context("when given valid arguments", func() {
			conf := &opv1alpha1.NetworkAddonsConfigSpec{ImagePullPolicy: v1.PullAlways, Multus: &opv1alpha1.Multus{}, LinuxBridge: &opv1alpha1.LinuxBridge{}}
			manifestDir := "../../data"
			openshiftNetworkConf := &osv1.Network{}
			clusterInfo := &ClusterInfo{SCCAvailable: true, OpenShift4: false}

			It("should successfully render a set of objects", func() {
				objs, _, err := Render(nil, conf, manifestDir, openshiftNetworkConf, clusterInfo)
				Expect(err).NotTo(HaveOccurred())
				Expect(objs).NotTo(BeEmpty())
			})
		})

		Context("when given manifest directory that does not contain all expected components", func() {
			conf := &opv1alpha1.NetworkAddonsConfigSpec{ImagePullPolicy: v1.PullAlways, Multus: &opv1alpha1.Multus{}, LinuxBridge: &opv1alpha1.LinuxBridge{}}
			manifestDir := "." // Test directory with this module
			openshiftNetworkConf := &osv1.Network{}
			clusterInfo := &ClusterInfo{SCCAvailable: true, OpenShift4: false}

			It("should return an error since it's unable to load templates for render", func() {
				_, _, err := Render(nil, conf, manifestDir, openshiftNetworkConf, clusterInfo)
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
