package libvmi

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "kubevirt.io/api/core/v1"
)

// NewVirtualMachine creates a VirtualMachine from a VMI template
func NewVirtualMachine(vmi *v1.VirtualMachineInstance, opts ...VMOption) *v1.VirtualMachine {
	vm := &v1.VirtualMachine{
		ObjectMeta: metav1.ObjectMeta{
			Name:      vmi.Name,
			Namespace: vmi.Namespace,
		},
		Spec: v1.VirtualMachineSpec{
			Template: &v1.VirtualMachineInstanceTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: vmi.Labels,
				},
				Spec: vmi.Spec,
			},
		},
	}

	// If VMI has no name, use GenerateName
	if vm.ObjectMeta.Name == "" && vmi.GenerateName != "" {
		vm.ObjectMeta.GenerateName = vmi.GenerateName
	}

	for _, opt := range opts {
		opt(vm)
	}

	return vm
}

// VMOption is a functional option for VirtualMachine
type VMOption func(*v1.VirtualMachine)

// WithRunStrategy sets the VM run strategy
func WithRunStrategy(strategy v1.VirtualMachineRunStrategy) VMOption {
	return func(vm *v1.VirtualMachine) {
		vm.Spec.RunStrategy = &strategy
	}
}
