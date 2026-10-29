package network

import (
	"slices"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"

	cnao "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/shared"
	"github.com/kubevirt/cluster-network-addons-operator/pkg/network/cni"
)

func int32Ptr(v int32) *int32 { return &v }

var _ = Describe("validateLinuxBridge", func() {
	DescribeTable("port validation",
		func(port *int32, expectError bool) {
			conf := &cnao.NetworkAddonsConfigSpec{
				LinuxBridge: &cnao.LinuxBridge{BridgeMarkerHealthPort: port},
			}
			errs := validateLinuxBridge(conf)
			if expectError {
				Expect(errs).NotTo(BeEmpty())
			} else {
				Expect(errs).To(BeEmpty())
			}
		},
		Entry("valid port", int32Ptr(9090), false),
		Entry("minimum valid port", int32Ptr(1), false),
		Entry("maximum valid port", int32Ptr(65535), false),
		Entry("default port", int32Ptr(8081), false),
		Entry("port zero", int32Ptr(0), true),
		Entry("port too large", int32Ptr(65536), true),
		Entry("negative port", int32Ptr(-1), true),
	)

	It("passes when LinuxBridge is nil", func() {
		conf := &cnao.NetworkAddonsConfigSpec{}
		Expect(validateLinuxBridge(conf)).To(BeEmpty())
	})

	It("passes when BridgeMarkerHealthPort is nil", func() {
		conf := &cnao.NetworkAddonsConfigSpec{
			LinuxBridge: &cnao.LinuxBridge{},
		}
		Expect(validateLinuxBridge(conf)).To(BeEmpty())
	})
})

var _ = Describe("renderLinuxBridge", func() {
	const manifestDir = "../../data"

	var clusterInfo *ClusterInfo

	BeforeEach(func() {
		clusterInfo = &ClusterInfo{SCCAvailable: false, OpenShift4: false}
	})

	Context("when LinuxBridge is nil", func() {
		It("returns nil without error", func() {
			conf := &cnao.NetworkAddonsConfigSpec{}
			objs, err := renderLinuxBridge(conf, manifestDir, clusterInfo)
			Expect(err).NotTo(HaveOccurred())
			Expect(objs).To(BeNil())
		})
	})

	Context("when LinuxBridge is set without a custom health port", func() {
		It("uses the default health port in the rendered DaemonSet", func() {
			conf := &cnao.NetworkAddonsConfigSpec{
				ImagePullPolicy: v1.PullIfNotPresent,
				LinuxBridge:     &cnao.LinuxBridge{},
				PlacementConfiguration: &cnao.PlacementConfiguration{
					Workloads: &cnao.Placement{},
				},
			}

			objs, err := renderLinuxBridge(conf, manifestDir, clusterInfo)
			Expect(err).NotTo(HaveOccurred())
			Expect(objs).NotTo(BeEmpty())

			ds, err := getBridgeMarkerDaemonSet(objs)
			Expect(err).NotTo(HaveOccurred())

			Expect(healthPortFromDaemonSet(ds)).To(Equal(int32(8081)),
				"expected default health port %s", cni.DefaultBridgeMarkerHealthPort)
			Expect(healthPortArgFromDaemonSet(ds)).To(Equal(cni.DefaultBridgeMarkerHealthPort))
		})
	})

	Context("when LinuxBridge is set with a custom health port", func() {
		It("uses the custom health port in the rendered DaemonSet", func() {
			conf := &cnao.NetworkAddonsConfigSpec{
				ImagePullPolicy: v1.PullIfNotPresent,
				LinuxBridge:     &cnao.LinuxBridge{BridgeMarkerHealthPort: int32Ptr(9090)},
				PlacementConfiguration: &cnao.PlacementConfiguration{
					Workloads: &cnao.Placement{},
				},
			}

			objs, err := renderLinuxBridge(conf, manifestDir, clusterInfo)
			Expect(err).NotTo(HaveOccurred())
			Expect(objs).NotTo(BeEmpty())

			ds, err := getBridgeMarkerDaemonSet(objs)
			Expect(err).NotTo(HaveOccurred())

			Expect(healthPortFromDaemonSet(ds)).To(Equal(int32(9090)))
			Expect(healthPortArgFromDaemonSet(ds)).To(Equal("9090"))
		})
	})
})

func getBridgeMarkerDaemonSet(objs []*unstructured.Unstructured) (*appsv1.DaemonSet, error) {
	idx := slices.IndexFunc(objs, func(obj *unstructured.Unstructured) bool {
		return obj.GetKind() == "DaemonSet" && obj.GetName() == "bridge-marker"
	})
	if idx == -1 {
		return nil, nil
	}
	ds := &appsv1.DaemonSet{}
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(objs[idx].Object, ds)
	if err != nil {
		return nil, err
	}
	return ds, nil
}

// healthPortFromDaemonSet returns the containerPort named "healthport" from the bridge-marker container.
func healthPortFromDaemonSet(ds *appsv1.DaemonSet) int32 {
	for _, c := range ds.Spec.Template.Spec.Containers {
		if c.Name != "bridge-marker" {
			continue
		}
		for _, p := range c.Ports {
			if p.Name == "healthport" {
				return p.ContainerPort
			}
		}
	}
	return 0
}

// healthPortArgFromDaemonSet returns the value passed to -health-probe-port in the bridge-marker container args.
func healthPortArgFromDaemonSet(ds *appsv1.DaemonSet) string {
	for _, c := range ds.Spec.Template.Spec.Containers {
		if c.Name != "bridge-marker" {
			continue
		}
		for i, arg := range c.Args {
			if arg == "-health-probe-port" && i+1 < len(c.Args) {
				return c.Args[i+1]
			}
		}
	}
	return ""
}
