package libvmi

import (
	k8sv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "kubevirt.io/api/core/v1"
)

// New creates a basic VirtualMachineInstance with functional options
func New(opts ...Option) *v1.VirtualMachineInstance {
	vmi := &v1.VirtualMachineInstance{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "test-vmi-",
		},
		Spec: v1.VirtualMachineInstanceSpec{
			Domain: v1.DomainSpec{
				Devices: v1.Devices{
					Interfaces: []v1.Interface{},
				},
				Resources: v1.ResourceRequirements{
					Requests: k8sv1.ResourceList{
						k8sv1.ResourceMemory: resource.MustParse("128Mi"),
					},
				},
			},
			Networks: []v1.Network{},
		},
	}

	for _, opt := range opts {
		opt(vmi)
	}

	return vmi
}

// Option is a functional option for VirtualMachineInstance
type Option func(*v1.VirtualMachineInstance)

// WithInterface adds a network interface to the VMI
func WithInterface(iface v1.Interface) Option {
	return func(vmi *v1.VirtualMachineInstance) {
		vmi.Spec.Domain.Devices.Interfaces = append(
			vmi.Spec.Domain.Devices.Interfaces, iface)
	}
}

// WithNetwork adds a network to the VMI
func WithNetwork(network v1.Network) Option {
	return func(vmi *v1.VirtualMachineInstance) {
		vmi.Spec.Networks = append(vmi.Spec.Networks, network)
	}
}

// WithMemory sets the memory request
func WithMemory(memory string) Option {
	return func(vmi *v1.VirtualMachineInstance) {
		vmi.Spec.Domain.Resources.Requests[k8sv1.ResourceMemory] =
			resource.MustParse(memory)
	}
}

// WithNamespace sets the namespace for the VMI
func WithNamespace(namespace string) Option {
	return func(vmi *v1.VirtualMachineInstance) {
		vmi.ObjectMeta.Namespace = namespace
	}
}

// WithName sets the name for the VMI
func WithName(name string) Option {
	return func(vmi *v1.VirtualMachineInstance) {
		vmi.ObjectMeta.Name = name
		vmi.ObjectMeta.GenerateName = ""
	}
}

// InterfaceDeviceWithMasqueradeBinding creates a masquerade interface
func InterfaceDeviceWithMasqueradeBinding() v1.Interface {
	return v1.Interface{
		Name: "default",
		InterfaceBindingMethod: v1.InterfaceBindingMethod{
			Masquerade: &v1.InterfaceMasquerade{},
		},
	}
}
