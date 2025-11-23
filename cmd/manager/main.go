package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strconv"

	"github.com/go-logr/logr"
	"go.uber.org/zap/zapcore"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	k8snetworkplumbingwgv1 "github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/apis/k8s.cni.cncf.io/v1"

	osv1 "github.com/openshift/api/operator/v1"
	"github.com/rhobs/operator-observability-toolkit/pkg/operatormetrics"
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

func printVersion(logger logr.Logger) {
	logger.Info("Go Version", "version", runtime.Version())
	logger.Info("Go OS/Arch", "os", runtime.GOOS, "arch", runtime.GOARCH)
	logger.Info("cluster-network-addons-operator version", "version", os.Getenv("OPERATOR_VERSION"))
}

func setupLogger() logr.Logger {
	logLevel := zapcore.InfoLevel
	if level, err := strconv.Atoi(os.Getenv("CNAO_LOG_LEVEL")); err == nil {
		logLevel = zapcore.Level(level)
	}

	opts := zap.Options{
		Development: false,
		Level:       logLevel,
	}
	opts.BindFlags(flag.CommandLine)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()

	logger := zap.New(zap.UseFlagOptions(&opts))
	ctrl.SetLogger(logger)
	return logger.WithName("manager")
}

func main() {
	logger := setupLogger()

	printVersion(logger)

	watchNamespace, err := k8s.GetWatchNamespace()
	if err != nil {
		logger.Error(err, "failed to get watch namespace")
		os.Exit(1)
	}

	// Get a config to talk to the apiserver
	cfg, err := config.GetConfig()
	if err != nil {
		logger.Error(err, "failed to get apiserver config")
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
		logger.Error(err, "failed to instantiate new operator manager")
		os.Exit(1)
	}

	// Setup Monitoring
	operatormetrics.Register = controllerruntimemetrics.Registry.Register
	err = metrics.SetupMetrics()
	if err != nil {
		logger.Error(err, "failed to setup metrics")
		os.Exit(1)
	}

	logger.Info("registering Components")

	if err := osv1.Install(mgr.GetScheme()); err != nil {
		logger.Error(err, "failed adding openshift scheme to the client")
		os.Exit(1)
	}

	logger.Info("adding readiness and liveness probes")
	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		logger.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		logger.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	// Setup all Controllers
	if err := controller.AddToManager(mgr); err != nil {
		logger.Error(err, "failed setting up operator controllers")
		os.Exit(1)
	}

	logger.Info("starting the operator manager")

	// Start the operator manager
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		logger.Error(err, "manager exited with non-zero")
		os.Exit(1)
	}
}
