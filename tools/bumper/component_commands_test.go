package main

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/coreos/go-semver/semver"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Testing internal git", func() {
	var (
		githubApi    *mockGithubApi
		gitComponent *gitComponent
		repoDir      string
	)
	expectedTagCommitMap := make(map[string]string)

	BeforeEach(func() {
		tempDir, err := ioutil.TempDir("/tmp", "component-commands-test")
		Expect(err).ToNot(HaveOccurred(), "Should create temp dir for component")

		repoDir = filepath.Join(tempDir, "testOwner", "testRepo")
		os.MkdirAll(repoDir, 0777)
		githubApi = newFakeGithubApi(repoDir)

		gitComponent = newFakeGitComponent(githubApi, repoDir, &component{}, expectedTagCommitMap)
	})

	type getVirtualTagParams struct {
		TagKey string
	}
	DescribeTable("getVirtualTag function",
		func(r getVirtualTagParams) {
			defer os.RemoveAll(gitComponent.gitRepo.localDir)

			By("Running api to get the current virtual tag")
			commitTested := expectedTagCommitMap[r.TagKey]
			currentReleaseTag, err := gitComponent.getVirtualTag(commitTested)

			By("Checking that tag is as expected")
			Expect(err).ToNot(HaveOccurred(), "should not fail to run getVirtualTag")
			expectedTag, err := describeHash(repoDir, commitTested)
			Expect(err).ToNot(HaveOccurred(), "should not fail to run describeHash")
			Expect(currentReleaseTag).To(Equal(expectedTag), "tag should be same as expected")
		},
		Entry("Should succeed getting annotated tag in master branch: v0.0.1", getVirtualTagParams{
			TagKey: "v0.0.1",
		}),
		Entry("Should succeed getting lightweight tag in master branch: v0.0.2", getVirtualTagParams{
			TagKey: "v0.0.2",
		}),
		Entry("Should succeed getting annotated tag in release branch: v1.0.0", getVirtualTagParams{
			TagKey: "v1.0.0",
		}),
		Entry("Should succeed getting annotated tag from release branch: v1.0.1", getVirtualTagParams{
			TagKey: "v1.0.1",
		}),
		Entry("Should succeed getting virtual tag from master branch", getVirtualTagParams{
			TagKey: "dummy_tag_latest_master",
		}),
		Entry("Should succeed getting virtual tag from release branch", getVirtualTagParams{
			TagKey: "dummy_tag_latest_release-v1.0.0",
		}),
	)

	type currentReleaseParams struct {
		TagKey          string
		shouldReturnErr bool
	}
	DescribeTable("getCurrentReleaseTag function",
		func(r currentReleaseParams) {
			defer os.RemoveAll(gitComponent.gitRepo.localDir)

			// update test params since you cant do it in the Entry context
			gitComponent.configParams.Url = repoDir
			gitComponent.configParams.Commit = expectedTagCommitMap[r.TagKey]

			By("Running api to get the current release tag")
			currentReleaseTag, err := gitComponent.getCurrentReleaseTag()

			By("Checking that tag is as expected")
			if r.shouldReturnErr {
				Expect(err).To(HaveOccurred(), "should fail to run getCurrentReleaseTag because tag not found")
			} else {
				Expect(err).ToNot(HaveOccurred(), "should not fail to run getCurrentReleaseTag")

				vtag, err := describeHash(repoDir, gitComponent.configParams.Commit)
				Expect(err).ToNot(HaveOccurred(), "should succeed finding virtual tag of dummy hash %s", gitComponent.configParams.Commit)
				Expect(currentReleaseTag).To(Equal(vtag), "Should return virtual tag in case tag was not found")
			}
		},
		Entry("Should succeed getting annotated tag from master branch: v0.0.1", currentReleaseParams{
			TagKey:          "v0.0.1",
			shouldReturnErr: false,
		}),
		Entry("Should succeed getting lightweight tag from master branch: v0.0.2", currentReleaseParams{
			TagKey:          "v0.0.2",
			shouldReturnErr: false,
		}),
		Entry("Should succeed getting current untagged commit from master branch", currentReleaseParams{
			TagKey:          "dummy_tag_latest_master",
			shouldReturnErr: false,
		}),
		Entry("Should succeed getting annotated tag from release branch: v1.0.0", currentReleaseParams{
			TagKey:          "v1.0.0",
			shouldReturnErr: false,
		}),
		Entry("Should succeed getting lightweight tag from release branch: v1.0.1", currentReleaseParams{
			TagKey:          "v1.0.1",
			shouldReturnErr: false,
		}),
		Entry("Should succeed getting current vtag of untagged commit from release branch", currentReleaseParams{
			TagKey:          "dummy_tag_latest_release-v1.0.0",
			shouldReturnErr: false,
		}),
		Entry("Should fail getting current vtag of unknown commit", currentReleaseParams{
			TagKey:          "dummy_false_commit",
			shouldReturnErr: true,
		}),
	)

	type updatedReleaseParams struct {
		comp            *component
		currentTagKey   string
		expectedTagKey  string
		shouldReturnErr bool
	}
	DescribeTable("getUpdatedReleaseInfo function",
		func(r updatedReleaseParams) {
			defer os.RemoveAll(gitComponent.gitRepo.localDir)

			// update test params since you cant do it in the Entry context
			gitComponent.configParams = r.comp
			gitComponent.configParams.Url = repoDir
			gitComponent.configParams.Commit = expectedTagCommitMap[r.currentTagKey]

			By("Running api to get the latest release tag")
			updatedReleaseTag, updateReleaseCommit, err := gitComponent.getUpdatedReleaseInfo()

			By("Checking that tag is as expected")
			if r.shouldReturnErr {
				Expect(err).To(HaveOccurred(), "should fail to run getUpdatedReleaseInfo")
			} else {
				expectedCommit := expectedTagCommitMap[r.expectedTagKey]
				expectedTag, err := describeHash(repoDir, expectedCommit)
				Expect(err).ToNot(HaveOccurred(), "should not fail to run describeHash")

				Expect(updatedReleaseTag).To(Equal(expectedTag), "tag should be same as expected")
				By("Checking that commit is as expected")
				Expect(updateReleaseCommit).To(Equal(expectedCommit), "commit should be same as expected")
			}
		},
		Entry("Update-policy static: should return original tag and commit", updatedReleaseParams{
			comp:            &component{Updatepolicy: updatePolicyStatic, Branch: "master", Metadata: "v0.0.2"},
			currentTagKey:   "v0.0.2",
			expectedTagKey:  "v0.0.2",
			shouldReturnErr: false,
		}),
		Entry("Update-policy tagged: should return latest tag in master branch. current: v0.0.1", updatedReleaseParams{
			comp:            &component{Updatepolicy: updatePolicyTagged, Branch: "master", Metadata: "v0.0.1"},
			currentTagKey:   "v0.0.1",
			expectedTagKey:  "v0.0.2",
			shouldReturnErr: false,
		}),
		Entry("Update-policy tagged: should return latest tag in master branch. current: v0.0.2", updatedReleaseParams{
			comp:            &component{Updatepolicy: updatePolicyTagged, Branch: "master", Metadata: "v0.0.2"},
			currentTagKey:   "v0.0.2",
			expectedTagKey:  "v0.0.2",
			shouldReturnErr: false,
		}),
		Entry("Update-policy tagged: should return latest tag in master branch. current: v1.0.0", updatedReleaseParams{
			comp:            &component{Updatepolicy: updatePolicyTagged, Branch: "master", Metadata: "v1.0.0"},
			currentTagKey:   "v1.0.0",
			expectedTagKey:  "v0.0.2",
			shouldReturnErr: false,
		}),
		Entry("Update-policy tagged: should return latest tag in release branch. current: v0.0.2", updatedReleaseParams{
			comp:            &component{Updatepolicy: updatePolicyTagged, Branch: "release-v1.0.0", Metadata: "v0.0.2"},
			currentTagKey:   "v0.0.2",
			expectedTagKey:  "v1.0.1",
			shouldReturnErr: false,
		}),
		Entry("Update-policy tagged: should return latest tag in release branch. current: v1.0.0", updatedReleaseParams{
			comp:            &component{Updatepolicy: updatePolicyTagged, Branch: "release-v1.0.0", Metadata: "v1.0.0"},
			currentTagKey:   "v1.0.0",
			expectedTagKey:  "v1.0.1",
			shouldReturnErr: false,
		}),
		Entry("Update-policy tagged: should fail if unknown branch", updatedReleaseParams{
			comp:            &component{Updatepolicy: updatePolicyTagged, Branch: "release-v2.0.0", Metadata: "v1.0.0"},
			currentTagKey:   "",
			expectedTagKey:  "",
			shouldReturnErr: true,
		}),
		Entry("Update-policy latest: should return latest HEAD in master branch. current: v1.0.0", updatedReleaseParams{
			comp:            &component{Updatepolicy: updatePolicyLatest, Branch: "master", Metadata: "v1.0.0"},
			currentTagKey:   "v1.0.0",
			expectedTagKey:  "dummy_tag_latest_master",
			shouldReturnErr: false,
		}),
		Entry("Update-policy latest: should return latest HEAD in release branch. current: v0.0.1", updatedReleaseParams{
			comp:            &component{Updatepolicy: updatePolicyLatest, Branch: "release-v1.0.0", Metadata: "v0.0.1"},
			currentTagKey:   "v0.0.1",
			expectedTagKey:  "dummy_tag_latest_release-v1.0.0",
			shouldReturnErr: false,
		}),
		Entry("Update-policy latest: should fail if unknown branch. current: v0.0.1", updatedReleaseParams{
			comp:            &component{Updatepolicy: updatePolicyLatest, Branch: "release-v2.0.0", Metadata: "v0.0.1"},
			currentTagKey:   "",
			expectedTagKey:  "",
			shouldReturnErr: true,
		}),
	)
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
