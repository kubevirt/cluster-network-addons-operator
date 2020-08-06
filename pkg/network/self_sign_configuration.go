package network

import (
	"reflect"
	"time"

	"github.com/pkg/errors"

	opv1alpha1 "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/v1alpha1"
)

const (
	caRotateIntervalDefault   = 7 * 24 * time.Hour // 7 days
	caOverlapIntervalDefault  = 24 * time.Hour     // 1 day
	certRotateIntervalDefault = 24 * time.Hour     // 1 day
)

// validateSelfSignConfiguration validates the following fields
// - CARotateInterval
// - CAOverlapInterval
// - CertRotateInterval
func validateSelfSignConfiguration(conf *opv1alpha1.NetworkAddonsConfigSpec) []error {
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

	// If we have a zero value don't continue
	if len(errs) > 0 {
		return errs
	}

	err = validateGreaterThan("caRotateInterval", caRotateInterval, "caOverlapInterval", caOverlapInterval)
	errs = appendOnError(errs, err)

	err = validateGreaterThan("caRotateInterval", caRotateInterval, "certRotateInterval", certRotateInterval)
	errs = appendOnError(errs, err)
	return errs
}

func fillDefaultsSelfSignConfiguration(conf, previous *opv1alpha1.NetworkAddonsConfigSpec) []error {
	if conf.SelfSignConfiguration == nil || conf.SelfSignConfiguration.CARotateInterval == "" || conf.SelfSignConfiguration.CAOverlapInterval == "" || conf.SelfSignConfiguration.CertRotateInterval == "" {
		if previous != nil && previous.SelfSignConfiguration != nil {
			conf.SelfSignConfiguration = previous.SelfSignConfiguration
			return []error{}
		}

		conf.SelfSignConfiguration = &opv1alpha1.SelfSignConfiguration{
			CARotateInterval:   caRotateIntervalDefault.String(),
			CAOverlapInterval:  caOverlapIntervalDefault.String(),
			CertRotateInterval: certRotateIntervalDefault.String(),
		}

	}
	return []error{}
}

func changeSafeSelfSignConfiguration(prev, next *opv1alpha1.NetworkAddonsConfigSpec) []error {
	if prev.SelfSignConfiguration != nil && next.SelfSignConfiguration != nil && !reflect.DeepEqual(prev.SelfSignConfiguration, next.SelfSignConfiguration) {
		return []error{errors.Errorf("cannot modify SelfSignConfiguration configuration once it is deployed")}
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
