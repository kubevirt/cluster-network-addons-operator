package network

import (
	"crypto/rand"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/kubevirt/cluster-network-addons-operator/pkg/render"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	cnao "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/shared"
)

// ValidateMultus validates the combination of DisableMultiNetwork and AddtionalNetworks
func validateKubeMacPool(conf *cnao.NetworkAddonsConfigSpec) []error {
	if conf.KubeMacPool == nil {
		return []error{}
	}

	// If the range is not configured by the administrator we generate a random range.
	// This random range spans from 02:XX:XX:00:00:00 to 02:XX:XX:FF:FF:FF,
	// where 02 makes the address local unicast and XX:XX is a random prefix.
	if conf.KubeMacPool.RangeStart == "" && conf.KubeMacPool.RangeEnd == "" {
		return []error{}
	}

	if (conf.KubeMacPool.RangeStart == "" && conf.KubeMacPool.RangeEnd != "") ||
		(conf.KubeMacPool.RangeStart != "" && conf.KubeMacPool.RangeEnd == "") {
		return []error{errors.Errorf("both or none of the KubeMacPool ranges needs to be configured")}
	}

	rangeStart, err := net.ParseMAC(conf.KubeMacPool.RangeStart)
	if err != nil {
		return []error{errors.Errorf("failed to parse rangeStart because the mac address is invalid")}
	}

	rangeEnd, err := net.ParseMAC(conf.KubeMacPool.RangeEnd)
	if err != nil {
		return []error{errors.Errorf("failed to parse rangeEnd because the mac address is invalid")}
	}

	if err := validateRange(rangeStart, rangeEnd); err != nil {
		return []error{errors.Errorf("failed to set mac address range: %v", err)}
	}

	if err := validateUnicast(rangeStart); err != nil {
		return []error{errors.Errorf("failed to set RangeStart: %v", err)}
	}

	if err := validateUnicast(rangeEnd); err != nil {
		return []error{errors.Errorf("failed to set RangeEnd: %v", err)}
	}

	return []error{}
}

func fillDefaultsKubeMacPool(conf, previous *cnao.NetworkAddonsConfigSpec) []error {
	if conf.KubeMacPool == nil {
		return []error{}
	}

	// If user hasn't explicitly requested a range, we try to reuse previously applied range
	if conf.KubeMacPool.RangeStart == "" || conf.KubeMacPool.RangeEnd == "" {
		if previous != nil && previous.KubeMacPool != nil {
			conf.KubeMacPool = previous.KubeMacPool
			return []error{}
		}

		// If no range was specified, we generated a random prefix
		prefix, err := generateRandomMacPrefix()
		if err != nil {
			return []error{errors.Wrap(err, "failed to generate random mac address prefix")}
		}

		rangeStart := net.HardwareAddr(append(prefix, 0x00, 0x00, 0x00))
		conf.KubeMacPool.RangeStart = rangeStart.String()

		rangeEnd := net.HardwareAddr(append(prefix, 0xFF, 0xFF, 0xFF))
		conf.KubeMacPool.RangeEnd = rangeEnd.String()
	}

	return []error{}
}

func changeSafeKubeMacPool(prev, next *cnao.NetworkAddonsConfigSpec) []error {
	return []error{}
}

// renderLinuxBridge generates the manifests of Linux Bridge
func renderKubeMacPool(conf *cnao.NetworkAddonsConfigSpec, manifestDir string, clusterInfo *ClusterInfo) ([]*unstructured.Unstructured, error) {
	if conf.KubeMacPool == nil {
		return nil, nil
	}

	// render the manifests on disk
	data := render.MakeRenderData()
	data.Data["Namespace"] = os.Getenv("OPERAND_NAMESPACE")
	data.Data["KubeMacPoolImage"] = os.Getenv("KUBEMACPOOL_IMAGE")
	data.Data["KubeRbacProxyImage"] = os.Getenv("KUBE_RBAC_PROXY_IMAGE")
	data.Data["ImagePullPolicy"] = conf.ImagePullPolicy
	data.Data["RangeStart"] = conf.KubeMacPool.RangeStart
	data.Data["RangeEnd"] = conf.KubeMacPool.RangeEnd
	data.Data["Placement"] = conf.PlacementConfiguration.Infra
	data.Data["CARotateInterval"] = conf.SelfSignConfiguration.CARotateInterval
	data.Data["CAOverlapInterval"] = conf.SelfSignConfiguration.CAOverlapInterval
	data.Data["CertRotateInterval"] = conf.SelfSignConfiguration.CertRotateInterval
	data.Data["CertOverlapInterval"] = conf.SelfSignConfiguration.CertOverlapInterval

	if clusterInfo.SCCAvailable {
		data.Data["RunAsNonRoot"] = "null"
		data.Data["RunAsUser"] = "null"
	} else {
		data.Data["RunAsNonRoot"] = "true"
		data.Data["RunAsUser"] = "107"
	}

	ciphers, tlsMinVersion := SelectCipherSuitesAndMinTLSVersion(conf.TLSSecurityProfile)
	data.Data["TLSSecurityProfileCiphers"] = strings.Join(ciphers, ",")
	data.Data["TLSMinVersion"] = TLSVersionToHumanReadable(tlsMinVersion)

	objs, err := render.RenderDir(filepath.Join(manifestDir, "kubemacpool"), &data)
	if err != nil {
		return nil, errors.Wrap(err, "failed to render kubeMacPool manifests")
	}

	return objs, nil
}

func generateRandomMacPrefix() ([]byte, error) {
	suffix := make([]byte, 2)
	_, err := rand.Read(suffix)
	if err != nil {
		return []byte{}, err
	}

	prefix := append([]byte{0x02}, suffix...)

	return prefix, nil
}

func validateRange(startMac, endMac net.HardwareAddr) error {
	for idx := 0; idx <= 5; idx++ {
		if startMac[idx] < endMac[idx] {
			return nil
		}
	}
	return fmt.Errorf("invalid range. Range end is lesser than or equal to its start. start: %v end: %v", startMac, endMac)
}

func validateUnicast(mac net.HardwareAddr) error {
	// A bitwise AND between 00000001 and the mac address first octet.
	multicastBit := 1 & mac[0]

	// In case where the LSB of the first octet (the multicast bit) is on, it will return 1, and 0 otherwise.
	if multicastBit == 1 {
		return fmt.Errorf("invalid mac address. Multicast addressing is not supported. Unicast addressing must be used. The first octet is %#0X", mac[0])
	}

	return nil
}
