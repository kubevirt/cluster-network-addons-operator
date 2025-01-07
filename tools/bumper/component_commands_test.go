package main

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Testing internal git component", func() {
	var (
		githubApi    *mockGithubApi
		gitComponent *gitComponent
		repoDir      string
	)
	expectedTagCommitMap := make(map[string]string)

	BeforeEach(func() {
		tempDir, err := os.MkdirTemp("/tmp", "component-commands-test")
		Expect(err).ToNot(HaveOccurred(), "Should create temp dir for component")

		repoDir = filepath.Join(tempDir, "testOwner", "testRepo")
		Expect(os.MkdirAll(repoDir, 0777)).To(Succeed())
		githubApi = newFakeGithubApi(repoDir)

		gitComponent = newFakeGitComponent(githubApi, repoDir, &component{}, expectedTagCommitMap)
	})

	type getVirtualTagParams struct {
		TagKey string
	}
	DescribeTable("getVirtualTag function",
		func(r getVirtualTagParams) {
			defer func(path string) {
				Expect(os.RemoveAll(path)).To(Succeed())
			}(gitComponent.gitRepo.localDir)

			By("Running api to get the current virtual tag")
			commitTested := expectedTagCommitMap[r.TagKey]
			currentReleaseTag, err := gitComponent.getVirtualTag(commitTested)

			By("Checking that tag is as expected")
			Expect(err).ToNot(HaveOccurred(), "should not fail to run getVirtualTag")
			expectedTag, err := describeHash(repoDir, commitTested)
			Expect(err).ToNot(HaveOccurred(), "should not fail to run describeHash")
			Expect(currentReleaseTag).To(Equal(expectedTag), "tag should be same as expected")
		},
		Entry("Should succeed getting annotated tag in main branch: v0.0.1", getVirtualTagParams{
			TagKey: "v0.0.1",
		}),
		Entry("Should succeed getting lightweight tag in main branch: v0.0.2", getVirtualTagParams{
			TagKey: "v0.0.2",
		}),
		Entry("Should succeed getting annotated tag in release branch: v1.0.0", getVirtualTagParams{
			TagKey: "v1.0.0",
		}),
		Entry("Should succeed getting annotated tag from release branch: v1.0.1", getVirtualTagParams{
			TagKey: "v1.0.1",
		}),
		Entry("Should succeed getting virtual tag from main branch", getVirtualTagParams{
			TagKey: "dummy_tag_latest_main",
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
			defer func(path string) {
				Expect(os.RemoveAll(path)).To(Succeed())
			}(gitComponent.gitRepo.localDir)

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
		Entry("Should succeed getting annotated tag from main branch: v0.0.1", currentReleaseParams{
			TagKey:          "v0.0.1",
			shouldReturnErr: false,
		}),
		Entry("Should succeed getting lightweight tag from main branch: v0.0.2", currentReleaseParams{
			TagKey:          "v0.0.2",
			shouldReturnErr: false,
		}),
		Entry("Should succeed getting current untagged commit from main branch", currentReleaseParams{
			TagKey:          "dummy_tag_latest_main",
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
			defer func(path string) {
				Expect(os.RemoveAll(path)).To(Succeed())
			}(gitComponent.gitRepo.localDir)

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
				Expect(err).ToNot(HaveOccurred(), "should not fail to run getUpdatedReleaseInfo")

				expectedCommit := expectedTagCommitMap[r.expectedTagKey]
				expectedTag, err := describeHash(repoDir, expectedCommit)
				Expect(err).ToNot(HaveOccurred(), "should not fail to run describeHash")

				Expect(updatedReleaseTag).To(Equal(expectedTag), "tag should be same as expected")
				By("Checking that commit is as expected")
				Expect(updateReleaseCommit).To(Equal(expectedCommit), "commit should be same as expected")
			}
		},
		Entry("Update-policy static: should return original tag and commit", updatedReleaseParams{
			comp:            &component{Updatepolicy: updatePolicyStatic, Branch: "main", Metadata: "v0.0.2"},
			currentTagKey:   "v0.0.2",
			expectedTagKey:  "v0.0.2",
			shouldReturnErr: false,
		}),
		Entry("Update-policy tagged: should return latest tag in main branch. current: v0.0.1", updatedReleaseParams{
			comp:            &component{Updatepolicy: updatePolicyTagged, Branch: "main", Metadata: "v0.0.1"},
			currentTagKey:   "v0.0.1",
			expectedTagKey:  "v0.0.2",
			shouldReturnErr: false,
		}),
		Entry("Update-policy tagged: should return latest tag in main branch. current: v0.0.2", updatedReleaseParams{
			comp:            &component{Updatepolicy: updatePolicyTagged, Branch: "main", Metadata: "v0.0.2"},
			currentTagKey:   "v0.0.2",
			expectedTagKey:  "v0.0.2",
			shouldReturnErr: false,
		}),
		Entry("Update-policy tagged: should return latest tag in main branch. current: v1.0.0", updatedReleaseParams{
			comp:            &component{Updatepolicy: updatePolicyTagged, Branch: "main", Metadata: "v1.0.0"},
			currentTagKey:   "v1.0.0",
			expectedTagKey:  "v0.0.2",
			shouldReturnErr: false,
		}),
		Entry("Update-policy tagged: should return latest tag in release branch. current: v0.0.2", updatedReleaseParams{
			comp:            &component{Updatepolicy: updatePolicyTagged, Branch: "release-v1.0.0", Metadata: "v0.0.2"},
			currentTagKey:   "v0.0.2",
			expectedTagKey:  "v1.0.2",
			shouldReturnErr: false,
		}),
		Entry("Update-policy tagged: should return latest tag in release branch. current: v1.0.0", updatedReleaseParams{
			comp:            &component{Updatepolicy: updatePolicyTagged, Branch: "release-v1.0.0", Metadata: "v1.0.0"},
			currentTagKey:   "v1.0.0",
			expectedTagKey:  "v1.0.2",
			shouldReturnErr: false,
		}),
		Entry("Update-policy tagged: should fail if unknown branch", updatedReleaseParams{
			comp:            &component{Updatepolicy: updatePolicyTagged, Branch: "release-v2.0.0", Metadata: "v1.0.0"},
			currentTagKey:   "",
			expectedTagKey:  "",
			shouldReturnErr: true,
		}),
		Entry("Update-policy latest: should return latest HEAD in main branch. current: v1.0.0", updatedReleaseParams{
			comp:            &component{Updatepolicy: updatePolicyLatest, Branch: "main", Metadata: "v1.0.0"},
			currentTagKey:   "v1.0.0",
			expectedTagKey:  "dummy_tag_latest_main",
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
})
