package env

import (
	k8snetworkplumbingwgv1 "github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/apis/k8s.cni.cncf.io/v1"
	. "github.com/onsi/gomega"

	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	kubevirtv1 "kubevirt.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"

	cnaov1 "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/v1"
	cnaov1alpha1 "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/v1alpha1"
)

var (
	cfg        *rest.Config
	Client     client.Client         // You'll be using this client in your tests.
	KubeClient *kubernetes.Clientset // You'll be using this client in your tests.
	testEnv    *envtest.Environment
)

func Start() {
	useExistingCluster := true
	testEnv = &envtest.Environment{
		UseExistingCluster: &useExistingCluster,
	}

	var err error
	cfg, err = testEnv.Start()
	ExpectWithOffset(1, err).ToNot(HaveOccurred())
	ExpectWithOffset(1, cfg).ToNot(BeNil())

	err = cnaov1.AddToScheme(scheme.Scheme)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	err = cnaov1alpha1.AddToScheme(scheme.Scheme)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	err = monitoringv1.AddToScheme(scheme.Scheme)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	err = k8snetworkplumbingwgv1.AddToScheme(scheme.Scheme)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	err = kubevirtv1.AddToScheme(scheme.Scheme)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())

	// +kubebuilder:scaffold:scheme

	Client, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	ExpectWithOffset(1, err).ToNot(HaveOccurred())
	ExpectWithOffset(1, Client).ToNot(BeNil())

	KubeClient, err = kubernetes.NewForConfig(cfg)
	ExpectWithOffset(1, err).ToNot(HaveOccurred())
}
