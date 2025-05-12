package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	k8snetworkplumbingwgv1 "github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/apis/k8s.cni.cncf.io/v1"

	"github.com/machadovilaca/operator-observability/pkg/operatormetrics"
	osv1 "github.com/openshift/api/operator/v1"
	"github.com/spf13/pflag"
	apiruntime "k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
	controllerruntimemetrics "sigs.k8s.io/controller-runtime/pkg/metrics"

	cnaov1 "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/v1"
	cnaov1alpha1 "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/v1alpha1"
	"github.com/kubevirt/cluster-network-addons-operator/pkg/components"
	"github.com/kubevirt/cluster-network-addons-operator/pkg/controller"
	"github.com/kubevirt/cluster-network-addons-operator/pkg/monitoring/metrics"
	"github.com/kubevirt/cluster-network-addons-operator/pkg/util/k8s"
)

var (
	scheme = apiruntime.NewScheme()
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(cnaov1.AddToScheme(scheme))
	utilruntime.Must(cnaov1alpha1.AddToScheme(scheme))
	utilruntime.Must(k8snetworkplumbingwgv1.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme
}

func printVersion() {
	log.Printf("Go Version: %s", runtime.Version())
	log.Printf("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH)
	log.Printf("version of cluster-network-addons-operator: %v", os.Getenv("OPERATOR_VERSION"))
}

func main() {
	// Add flags registered by imported packages (e.g. controller-runtime)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()

	printVersion()

	watchNamespace, err := k8s.GetWatchNamespace()
	if err != nil {
		log.Printf("failed to get watch namespace: %v", err)
		os.Exit(1)
	}

	// Get a config to talk to the apiserver
	cfg, err := config.GetConfig()
	if err != nil {
		log.Printf("failed to get apiserver config: %v", err)
		os.Exit(1)
	}

	// Create a new Cmd to provide shared dependencies and start components
	mgr, err := manager.New(cfg, manager.Options{
		Scheme: scheme,
		Cache: cache.Options{
			DefaultNamespaces: map[string]cache.Config{
				watchNamespace: {},
			},
		},
		Metrics: metricsserver.Options{
			BindAddress: metricsserver.DefaultBindAddress,
		},
		MapperProvider: func(c *rest.Config, httpClient *http.Client) (meta.RESTMapper, error) {
			return apiutil.NewDynamicRESTMapper(c, httpClient)
		},
		HealthProbeBindAddress: fmt.Sprintf(":%d", components.HealthProbePort),
	})
	if err != nil {
		log.Printf("failed to instantiate new operator manager: %v", err)
		os.Exit(1)
	}

	// Setup Monitoring
	operatormetrics.Register = controllerruntimemetrics.Registry.Register
	err = metrics.SetupMetrics()
	if err != nil {
		log.Printf("failed to setup metrics: %v", err)
		os.Exit(1)
	}

	log.Print("registering Components")

	if err := osv1.Install(mgr.GetScheme()); err != nil {
		log.Printf("failed adding openshift scheme to the client: %v", err)
		os.Exit(1)
	}

	log.Print("Add readiness and liveness probes")
	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		log.Printf("unable to set up health check: %v", err)
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		log.Printf("unable to set up ready check: %v", err)
		os.Exit(1)
	}

	// Setup all Controllers
	if err := controller.AddToManager(mgr); err != nil {
		log.Printf("failed setting up operator controllers: %v", err)
		os.Exit(1)
	}

	log.Print("starting the operator manager")

	// Start the operator manager
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		log.Printf("manager exited with non-zero: %v", err)
		os.Exit(1)
	}
}
