package main

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/coreos/go-semver/semver"
)

var _ = Describe("Testing internal bumper", func() {
	type canonicalizeVersionParams struct {
		version        string
		expectedResult *semver.Version
		isValid        bool
	}
	DescribeTable("canonicalizeVersion function",
		func(v canonicalizeVersionParams) {
			By("Parsing the version string")
			formattedVersion, err := canonicalizeVersion(v.version)

			By("Checking expected error received")
			if v.isValid {
				Expect(err).ToNot(HaveOccurred(), "Expect function to not return an Error")
			} else {
				Expect(err).To(HaveOccurred(), "Expect function to return an Error")
			}

			By("Checking expected semver version received")
			Expect(formattedVersion).To(Equal(v.expectedResult))
		},
		Entry("When using 2 dotted format x.y, Should return corrected semver format version and not return an error", canonicalizeVersionParams{
			version:        "3.6",
			expectedResult: &semver.Version{Major: 3, Minor: 6, Patch: 0, PreRelease: ""},
			isValid:        true,
		}),
		Entry("When using pure semver format x.y.z, Should return valid semver format version", canonicalizeVersionParams{
			version:        "1.2.3",
			expectedResult: &semver.Version{Major: 1, Minor: 2, Patch: 3, PreRelease: ""},
			isValid:        true,
		}),
		Entry("When using semver format with v prefix vx.y.z, Should return valid semver format version", canonicalizeVersionParams{
			version:        "v4.5.6",
			expectedResult: &semver.Version{Major: 4, Minor: 5, Patch: 6, PreRelease: ""},
			isValid:        true,
		}),
		Entry("When using semver format with metadata x.y.z-Metadata, Should return valid semver format version", canonicalizeVersionParams{
			version:        "v7.8.9-rc",
			expectedResult: &semver.Version{Major: 7, Minor: 8, Patch: 9, PreRelease: "rc"},
			isValid:        true,
		}),
		Entry("When using unsupported version format, Should return Error as formatting failed", canonicalizeVersionParams{
			version:        "version1.2.3",
			expectedResult: nil,
			isValid:        false,
		}),
	)

	type isVtagFormatParams struct {
		version        string
		expectedResult bool
	}
	DescribeTable("isVtagFormat function",
		func(p isVtagFormatParams) {
			By("Checking if the version string of vtag format")
			isVtag := isVtagFormat(p.version)

			By("Checking expected result received")
			Expect(isVtag).To(Equal(p.expectedResult), "Expect result to match expected")
		},
		Entry("When using 2 dotted format x.y, Should not recognize as vtag format", isVtagFormatParams{
			version:        "3.6",
			expectedResult: false,
		}),
		Entry("When using pure semver format x.y.z, Should not recognize as vtag format", isVtagFormatParams{
			version:        "1.2.3",
			expectedResult: false,
		}),
		Entry("When using semver format with v prefix vx.y.z, Should not recognize as vtag format", isVtagFormatParams{
			version:        "v4.5.6",
			expectedResult: false,
		}),
		Entry("When using semver format with metadata x.y.z-Metadata, Should not recognize as vtag format", isVtagFormatParams{
			version:        "v7.8.9-rc",
			expectedResult: false,
		}),
		Entry("When using unsupported version format, Should not recognize as vtag format", isVtagFormatParams{
			version:        "version1.2.3",
			expectedResult: false,
		}),
		Entry("When using vtag version format, Should recognize as vtag format", isVtagFormatParams{
			version:        "0.39.0-32-g1fcbe815",
			expectedResult: true,
		}),
	)

	type isBumpNeededParams struct {
		currentReleaseVersion string
		latestReleaseVersion  string
		updatePolicy          string
		isBumpExpected        bool
		isValid               bool
	}
	DescribeTable("isBumpNeeded function",
		func(b isBumpNeededParams) {
			By("Checking if bump is needed")
			isBumpNeeded, err := isBumpNeeded(b.currentReleaseVersion, b.latestReleaseVersion, b.updatePolicy)
			By("Checking expected error received")
			if b.isValid {
				Expect(err).ToNot(HaveOccurred(), "Expect function to not return an Error")
			} else {
				Expect(err).To(HaveOccurred(), "Expect function to return an Error")
			}

			By("Checking if bump result is as expected")
			Expect(isBumpNeeded).To(Equal(b.isBumpExpected), "Expect bump result to be equal to expected")
		},
		Entry("Should not bump since there is updatePolicy static", isBumpNeededParams{
			currentReleaseVersion: "v2.5.1",
			latestReleaseVersion:  "v3.6.2",
			updatePolicy:          "static",
			isBumpExpected:        false,
			isValid:               true,
		}),
		Entry("Should not bump since there is updatePolicy static (vtag-format)", isBumpNeededParams{
			currentReleaseVersion: "v0.11.0-3-g1be91ab",
			latestReleaseVersion:  "v0.11.0-4-g1ar46a5",
			updatePolicy:          "latest",
			isBumpExpected:        true,
			isValid:               true,
		}),
		Entry("Should not bump since latest version is the same as current", isBumpNeededParams{
			currentReleaseVersion: "v3.6.2",
			latestReleaseVersion:  "v3.6.2",
			updatePolicy:          "tagged",
			isBumpExpected:        false,
			isValid:               true,
		}),
		Entry("Should not bump since latest version is not bigger than current", isBumpNeededParams{
			currentReleaseVersion: "v3.6.2",
			latestReleaseVersion:  "v3.5.2",
			updatePolicy:          "tagged",
			isBumpExpected:        false,
			isValid:               true,
		}),
		Entry("Should not bump since latest is in vtag-format and is the same as the current", isBumpNeededParams{
			currentReleaseVersion: "v0.11.0-3-g1be91ab",
			latestReleaseVersion:  "v0.11.0-3-g1be91ab",
			updatePolicy:          "tagged",
			isBumpExpected:        false,
			isValid:               true,
		}),
		Entry("Should bump since latest is in bigger than current", isBumpNeededParams{
			currentReleaseVersion: "v3.6.1",
			latestReleaseVersion:  "v3.6.2",
			updatePolicy:          "tagged",
			isBumpExpected:        true,
			isValid:               true,
		}),
		Entry("Should bump since latest is in vtag-format and different than current", isBumpNeededParams{
			currentReleaseVersion: "v0.11.0-3-g1be91ab",
			latestReleaseVersion:  "v0.11.0-4-g1ar46a5",
			updatePolicy:          "latest",
			isBumpExpected:        true,
			isValid:               true,
		}),
		Entry("Should return error since current is not in correct semver version format", isBumpNeededParams{
			currentReleaseVersion: "ver1.2.3",
			latestReleaseVersion:  "v3.6.2",
			updatePolicy:          "tagged",
			isBumpExpected:        false,
			isValid:               false,
		}),
		Entry("Should return error since latest is not in correct semver version format", isBumpNeededParams{
			currentReleaseVersion: "v3.6.2",
			latestReleaseVersion:  "ver1.2.3",
			updatePolicy:          "tagged",
			isBumpExpected:        false,
			isValid:               false,
		}),
	)
})
