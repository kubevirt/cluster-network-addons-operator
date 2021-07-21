package monitoring

import (
	"io/ioutil"
	"log"
	"net/http"
	"path"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"k8s.io/client-go/util/cert"
)

func StartPrometheus() error {
	certsDirectory, err := ioutil.TempDir("", "certs")
	if err != nil {
		return errors.Wrap(err, "error creating temp dir")
	}

	return startPrometheusEndpoint(certsDirectory)
}

// startPrometheusEndpoint starts an http server providing a prometheus endpoint using the passed
// in directory to store the self signed certificates that will be generated before starting the
// http server.
func startPrometheusEndpoint(certsDirectory string) error {
	certBytes, keyBytes, err := cert.GenerateSelfSignedCertKey("cloner_target", nil, nil)
	if err != nil {
		return errors.Wrap(err, "error generating cert for prometheus")
	}

	certFile := path.Join(certsDirectory, "tls.crt")
	if err = ioutil.WriteFile(certFile, certBytes, 0600); err != nil {
		return errors.Wrap(err, "error writing cert file")
	}

	keyFile := path.Join(certsDirectory, "tls.key")
	if err = ioutil.WriteFile(keyFile, keyBytes, 0600); err != nil {
		return errors.Wrap(err, "error writing key file")
	}

	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())
		server := http.Server{
			Addr:    "0.0.0.0:8443",
			Handler: mux,
		}
		log.Printf("Starting Prometheus metrics endpoint server")
		err := server.ListenAndServeTLS(certFile, keyFile)
		if err != nil {
			log.Fatalf("Failed to start Prometheus metrics endpoint server: %v", err)
		}
	}()
	return nil
}
