package components

import (
	ocpv1 "github.com/openshift/api/config/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Sorting TLSProfiles", func() {
	type tlsProfilesSortCase struct {
		tlsProfiles        map[ocpv1.TLSProfileType]*ocpv1.TLSProfileSpec
		expectedSortedKeys []ocpv1.TLSProfileType
	}
	DescribeTable("When sortedKeys is called", func(c tlsProfilesSortCase) {
		Expect(tlsProfiles(c.tlsProfiles).sortedKeys()).To(Equal(c.expectedSortedKeys))
	},
		Entry("with one profile", tlsProfilesSortCase{
			tlsProfiles: map[ocpv1.TLSProfileType]*ocpv1.TLSProfileSpec{
				ocpv1.TLSProfileIntermediateType: ocpv1.TLSProfiles[ocpv1.TLSProfileIntermediateType],
			},
			expectedSortedKeys: []ocpv1.TLSProfileType{ocpv1.TLSProfileIntermediateType},
		}),
		Entry("with all profiles", tlsProfilesSortCase{
			tlsProfiles: map[ocpv1.TLSProfileType]*ocpv1.TLSProfileSpec{
				ocpv1.TLSProfileIntermediateType: ocpv1.TLSProfiles[ocpv1.TLSProfileIntermediateType],
				ocpv1.TLSProfileOldType:          ocpv1.TLSProfiles[ocpv1.TLSProfileOldType],
				ocpv1.TLSProfileModernType:       ocpv1.TLSProfiles[ocpv1.TLSProfileModernType],
			},
			expectedSortedKeys: []ocpv1.TLSProfileType{ocpv1.TLSProfileIntermediateType, ocpv1.TLSProfileModernType, ocpv1.TLSProfileOldType},
		}))
})
