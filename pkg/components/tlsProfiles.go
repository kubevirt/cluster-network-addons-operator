package components

import (
	"sort"

	ocpv1 "github.com/openshift/api/config/v1"
)

type byTLSProfileName []ocpv1.TLSProfileType

func (a byTLSProfileName) Len() int           { return len(a) }
func (a byTLSProfileName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byTLSProfileName) Less(i, j int) bool { return a[i] < a[j] }

type tlsProfiles map[ocpv1.TLSProfileType]*ocpv1.TLSProfileSpec

func (p tlsProfiles) sortedKeys() []ocpv1.TLSProfileType {
	tlsProfiles := []ocpv1.TLSProfileType{}
	for k := range p {
		tlsProfiles = append(tlsProfiles, k)
	}
	sort.Sort(byTLSProfileName(tlsProfiles))
	return tlsProfiles
}
