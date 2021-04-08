package network

import (
	"time"

	"github.com/pkg/errors"

	cnao "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/shared"
)

const (
	caRotateIntervalDefault    = 48 * time.Hour
	caOverlapIntervalDefault   = 24 * time.Hour
	certRotateIntervalDefault  = 24 * time.Hour
	certOverlapIntervalDefault = 12 * time.Hour
)

// validateSelfSignConfiguration validates the following fields
// - CARotateInterval
// - CAOverlapInterval
// - CertRotateInterval
// - CertOverlapInterval
func validateSelfSignConfiguration(conf *cnao.NetworkAddonsConfigSpec) []error {
	if conf.SelfSignConfiguration == nil {
		return []error{}
	}

	selfSignConfiguration := *conf.SelfSignConfiguration

	errs := []error{}

	err := validateNotEmpty("caRotateInterval", selfSignConfiguration.CARotateInterval)
	errs = appendOnError(errs, err)

	err = validateNotEmpty("caOverlapInterval", selfSignConfiguration.CAOverlapInterval)
	errs = appendOnError(errs, err)

	err = validateNotEmpty("certRotateInterval", selfSignConfiguration.CertRotateInterval)
	errs = appendOnError(errs, err)

	err = validateNotEmpty("certOverlapInterval", selfSignConfiguration.CertOverlapInterval)
	errs = appendOnError(errs, err)

	// There are empty values don't continue
	if len(errs) > 0 {
		return errs
	}

	caRotateInterval, err := parseCertificateKnob("caRotateInterval", selfSignConfiguration.CARotateInterval)
	errs = appendOnError(errs, err)

	caOverlapInterval, err := parseCertificateKnob("caOverlapInterval", selfSignConfiguration.CAOverlapInterval)
	errs = appendOnError(errs, err)

	certRotateInterval, err := parseCertificateKnob("certRotateInterval", selfSignConfiguration.CertRotateInterval)
	errs = appendOnError(errs, err)

	certOverlapInterval, err := parseCertificateKnob("certOverlapInterval", selfSignConfiguration.CertOverlapInterval)
	errs = appendOnError(errs, err)

	// If they cannot be parsed don't continue
	if len(errs) > 0 {
		return errs
	}

	err = validateGreaterThanZero("caRotateInterval", caRotateInterval)
	errs = appendOnError(errs, err)

	err = validateGreaterThanZero("caOverlapInterval", caOverlapInterval)
	errs = appendOnError(errs, err)

	err = validateGreaterThanZero("certRotateInterval", certRotateInterval)
	errs = appendOnError(errs, err)

	err = validateGreaterThanZero("certOverlapInterval", certOverlapInterval)
	errs = appendOnError(errs, err)

	// If we have a zero value don't continue
	if len(errs) > 0 {
		return errs
	}

	err = validateGreaterThan("caRotateInterval", caRotateInterval, "caOverlapInterval", caOverlapInterval)
	errs = appendOnError(errs, err)

	err = validateGreaterThan("caRotateInterval", caRotateInterval, "certRotateInterval", certRotateInterval)
	errs = appendOnError(errs, err)

	err = validateGreaterThan("certRotateInterval", certRotateInterval, "certOverlapInterval", certOverlapInterval)
	errs = appendOnError(errs, err)
	return errs
}
func DefaultSelfSignConfiguration() *cnao.SelfSignConfiguration {
	return &cnao.SelfSignConfiguration{
		CARotateInterval:    caRotateIntervalDefault.String(),
		CAOverlapInterval:   caOverlapIntervalDefault.String(),
		CertRotateInterval:  certRotateIntervalDefault.String(),
		CertOverlapInterval: certOverlapIntervalDefault.String(),
	}
}
func fillDefaultsSelfSignConfiguration(conf, previous *cnao.NetworkAddonsConfigSpec) []error {
	isMissingSelfSignConfiguration := func(confToCheck *cnao.NetworkAddonsConfigSpec) bool {
		return confToCheck.SelfSignConfiguration == nil ||
			confToCheck.SelfSignConfiguration.CARotateInterval == "" ||
			confToCheck.SelfSignConfiguration.CAOverlapInterval == "" ||
			confToCheck.SelfSignConfiguration.CertRotateInterval == "" ||
			confToCheck.SelfSignConfiguration.CertOverlapInterval == ""
	}

	if isMissingSelfSignConfiguration(conf) {
		if previous != nil && !isMissingSelfSignConfiguration(previous) {
			conf.SelfSignConfiguration = previous.SelfSignConfiguration

		} else {
			conf.SelfSignConfiguration = DefaultSelfSignConfiguration()
		}
	}
	return []error{}
}

func parseCertificateKnob(name, value string) (time.Duration, error) {
	d, err := time.ParseDuration(value)
	if err != nil {
		return d, errors.Wrapf(err, "failed to validate selfSignConfiguration: error parsing %s", name)
	}
	return d, nil
}

func validateNotEmpty(name, value string) error {
	if value == "" {
		return errors.Errorf("failed to validate selfSignConfiguration: %s is missing", name)
	}
	return nil

}

func validateGreaterThanZero(name string, d time.Duration) error {
	if d == 0 {
		return errors.Errorf("failed to validate selfSignConfiguration: %s duration has to be > 0", name)
	}
	return nil

}

func validateGreaterThan(lhsName string, lhsValue time.Duration, rhsNamed string, rhsValue time.Duration) error {
	if rhsValue > lhsValue {
		return errors.Errorf("failed to validate selfSignConfiguration: %s(%s) has to be <= %s(%s)", rhsNamed, rhsValue, lhsName, lhsValue)
	}
	return nil
}

func appendOnError(errs []error, err error) []error {
	if err != nil {
		return append(errs, err)
	}
	return errs
}
