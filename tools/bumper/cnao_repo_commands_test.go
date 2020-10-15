package main

import (
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/coreos/go-semver/semver"
)

var _ = Describe("Testing internal git CNAO Repo", func() {
	var (
		githubApi   *mockGithubApi
		gitCnaoRepo *gitCnaoRepo
		repoDir     string
	)
	expectedTagCommitMap := make(map[string]string)

	BeforeEach(func() {
		tempDir, err := ioutil.TempDir("/tmp", "cnao-repo-commands-test")
		Expect(err).ToNot(HaveOccurred(), "Should create temp dir for CNAO repo")

		repoDir = filepath.Join(tempDir, "testOwner", "testRepo")
		os.MkdirAll(repoDir, 0777)
		githubApi = newFakeGithubApi(repoDir)

		gitCnaoRepo = newFakeGitCnaoRepo(githubApi, repoDir, &component{}, expectedTagCommitMap)
	})

	Context("Creating fake PRs", func() {
		type isPrAlreadyOpenedParams struct {
			prTitle      string
			expectResult bool
		}
		dummyOwner := "dummyOwner"
		dummyRepo := "dummyRepo"
		BeforeEach(func() {

			newPr := getFakePrWithTitle("CNAO test-component to 0.0.2")
			_, _, err := gitCnaoRepo.githubInterface.CreatePullRequest(dummyOwner, dummyRepo, newPr)
			Expect(err).ToNot(HaveOccurred(), "should succeed creating fake PR")

			newPr = getFakePrWithTitle("CNAO test-component to 1.0.0")
			_, _, err = gitCnaoRepo.githubInterface.CreatePullRequest(dummyOwner, dummyRepo, newPr)
			Expect(err).ToNot(HaveOccurred(), "should succeed creating fake PR")

			newPr = getFakePrWithTitle("CNAO test-component to 1.0.1")
			_, _, err = gitCnaoRepo.githubInterface.CreatePullRequest(dummyOwner, dummyRepo, newPr)
			Expect(err).ToNot(HaveOccurred(), "should succeed creating fake PR")
		})

		DescribeTable("and checking isPrAlreadyOpened function",
			func(r isPrAlreadyOpenedParams) {
				defer os.RemoveAll(gitCnaoRepo.gitRepo.localDir)
				By("Running api to check if a PR is already opened")
				isPrAlreadyOpened, err := gitCnaoRepo.isPrAlreadyOpened(dummyOwner, dummyRepo, r.prTitle)

				By("Checking that result is as expected")
				Expect(err).ToNot(HaveOccurred(), "should not fail to run isPrAlreadyOpened")
				Expect(isPrAlreadyOpened).To(Equal(r.expectResult))
			},
			Entry("should find PR that is first in the list", isPrAlreadyOpenedParams{
				prTitle:      "CNAO test-component to 0.0.2",
				expectResult: true,
			}),
			Entry("should find PR that is in the middle of the list", isPrAlreadyOpenedParams{
				prTitle:      "CNAO test-component to 1.0.0",
				expectResult: true,
			}),
			Entry("should find PR that is last in the list", isPrAlreadyOpenedParams{
				prTitle:      "CNAO test-component to 1.0.1",
				expectResult: true,
			}),
			Entry("should not find PR that has empty title string", isPrAlreadyOpenedParams{
				prTitle:      "",
				expectResult: false,
			}),
			Entry("should not find PR that is not in the list", isPrAlreadyOpenedParams{
				prTitle:      "CNAO test-component to 1.0.3",
				expectResult: false,
			}),
		)
	})

	type canonicalizeVersionParams struct {
		version        string
		expectedResult *semver.Version
		isValid        bool
	}
	DescribeTable("canonicalizeVersion function",
		func(v canonicalizeVersionParams) {
			defer os.RemoveAll(gitCnaoRepo.gitRepo.localDir)

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
			defer os.RemoveAll(gitCnaoRepo.gitRepo.localDir)

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

	type isComponentBumpNeededParams struct {
		currentReleaseVersion string
		latestReleaseVersion  string
		updatePolicy          string
		prTitle               string
		isBumpExpected        bool
		isValid               bool
	}
	dummyPRTitle := "dummy new PR title"
	DescribeTable("isComponentBumpNeeded function",
		func(b isComponentBumpNeededParams) {
			defer os.RemoveAll(gitCnaoRepo.gitRepo.localDir)
			gitCnaoRepo.configParams.Url = repoDir

			By("Checking if bump is needed")
			isComponentBumpNeeded, err := gitCnaoRepo.isComponentBumpNeeded(b.currentReleaseVersion, b.latestReleaseVersion, b.updatePolicy, b.prTitle)
			By("Checking expected error received")
			if b.isValid {
				Expect(err).ToNot(HaveOccurred(), "Expect function to not return an Error")
			} else {
				Expect(err).To(HaveOccurred(), "Expect function to return an Error")
			}

			By("Checking if bump result is as expected")
			Expect(isComponentBumpNeeded).To(Equal(b.isBumpExpected), "Expect bump result to be equal to expected")
		},
		Entry("Should not bump since there is updatePolicy static", isComponentBumpNeededParams{
			currentReleaseVersion: "v2.5.1",
			latestReleaseVersion:  "v3.6.2",
			updatePolicy:          "static",
			prTitle:               dummyPRTitle,
			isBumpExpected:        false,
			isValid:               true,
		}),
		Entry("Should not bump since there is updatePolicy static (vtag-format)", isComponentBumpNeededParams{
			currentReleaseVersion: "v0.11.0-3-g1be91ab",
			latestReleaseVersion:  "v0.11.0-4-g1ar46a5",
			updatePolicy:          "latest",
			prTitle:               dummyPRTitle,
			isBumpExpected:        true,
			isValid:               true,
		}),
		Entry("Should not bump since latest version is the same as current", isComponentBumpNeededParams{
			currentReleaseVersion: "v3.6.2",
			latestReleaseVersion:  "v3.6.2",
			updatePolicy:          "tagged",
			prTitle:               dummyPRTitle,
			isBumpExpected:        false,
			isValid:               true,
		}),
		Entry("Should not bump since latest version is not bigger than current", isComponentBumpNeededParams{
			currentReleaseVersion: "v3.6.2",
			latestReleaseVersion:  "v3.5.2",
			updatePolicy:          "tagged",
			prTitle:               dummyPRTitle,
			isBumpExpected:        false,
			isValid:               true,
		}),
		Entry("Should not bump since latest is in vtag-format and is the same as the current", isComponentBumpNeededParams{
			currentReleaseVersion: "v0.11.0-3-g1be91ab",
			latestReleaseVersion:  "v0.11.0-3-g1be91ab",
			updatePolicy:          "tagged",
			prTitle:               dummyPRTitle,
			isBumpExpected:        false,
			isValid:               true,
		}),
		Entry("Should bump since latest is in bigger than current", isComponentBumpNeededParams{
			currentReleaseVersion: "v3.6.1",
			latestReleaseVersion:  "v3.6.2",
			updatePolicy:          "tagged",
			prTitle:               dummyPRTitle,
			isBumpExpected:        true,
			isValid:               true,
		}),
		Entry("Should bump since latest is in vtag-format and different than current", isComponentBumpNeededParams{
			currentReleaseVersion: "v0.11.0-3-g1be91ab",
			latestReleaseVersion:  "v0.11.0-4-g1ar46a5",
			updatePolicy:          "latest",
			prTitle:               dummyPRTitle,
			isBumpExpected:        true,
			isValid:               true,
		}),
		Entry("Should return error since current is not in correct semver version format", isComponentBumpNeededParams{
			currentReleaseVersion: "ver1.2.3",
			latestReleaseVersion:  "v3.6.2",
			updatePolicy:          "tagged",
			prTitle:               dummyPRTitle,
			isBumpExpected:        false,
			isValid:               false,
		}),
		Entry("Should return error since latest is not in correct semver version format", isComponentBumpNeededParams{
			currentReleaseVersion: "v3.6.2",
			latestReleaseVersion:  "ver1.2.3",
			updatePolicy:          "tagged",
			prTitle:               dummyPRTitle,
			isBumpExpected:        false,
			isValid:               false,
		}),
	)
})
