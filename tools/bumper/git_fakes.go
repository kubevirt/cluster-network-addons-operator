package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/google/go-github/v32/github"
	"github.com/pkg/errors"
	"github.com/thanhpk/randstr"
)

type mockGithubApi struct {
	repoDir    string
	fakePRList []*github.PullRequest
}

func (g mockGithubApi) listMatchingRefs(owner, repo string, opts *github.ReferenceListOptions) ([]*github.Reference, *github.Response, error) {
	gitCommitObjList, err := gitLogJson(g.repoDir, "")
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed performing mock git log")
	}

	return convertLogToReferenceList(gitCommitObjList, opts.Ref), nil, nil
}

func (g mockGithubApi) listCommits(owner, repo string, opts *github.CommitsListOptions) ([]*github.RepositoryCommit, *github.Response, error) {
	gitCommitObjList, err := gitLogJson(g.repoDir, opts.SHA)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed performing mock git log")
	}

	return convertLogToRepositoryCommitList(gitCommitObjList), nil, nil
}

func (g mockGithubApi) getBranchRef(owner string, repo string, ref string) (*github.Reference, *github.Response, error) {
	gitCommitObjList, err := gitLogJson(g.repoDir, "")
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed performing mock git log")
	}

	githubRef, err := getRefFromCommitObjList(gitCommitObjList, ref)
	return githubRef, nil, err
}

func (g mockGithubApi) createBranchRef(owner string, repo string, newRef *github.Reference) (*github.Reference, *github.Response, error) {
	return nil, nil, nil
}

func (g mockGithubApi) createTree(owner string, repo string, baseTree string, entries []*github.TreeEntry) (*github.Tree, *github.Response, error) {
	return nil, nil, nil
}

func (g mockGithubApi) getCommit(owner string, repo string, sha string) (*github.Commit, *github.Response, error) {
	return nil, nil, nil
}

func (g mockGithubApi) createCommit(owner string, repo string, commit *github.Commit) (*github.Commit, *github.Response, error) {
	return nil, nil, nil
}

func (g mockGithubApi) updateRef(owner string, repo string, ref *github.Reference, force bool) (*github.Reference, *github.Response, error) {
	return nil, nil, nil
}

func (g *mockGithubApi) listPullRequests(owner string, repo string) ([]*github.PullRequest, *github.Response, error) {
	return g.fakePRList, nil, nil
}

func (g *mockGithubApi) createPullRequest(owner string, repo string, pull *github.NewPullRequest) (*github.PullRequest, *github.Response, error) {
	pullRequest := &github.PullRequest{Title: pull.Title}
	g.fakePRList = append(g.fakePRList, pullRequest)

	return pullRequest, nil, nil
}

type gitCommitMock struct {
	Commit string `json:"commit"`
	Refs   string `json:"refs"`
}

var GITFORMAT = `--pretty=format:{
  "commit": "%H",
  "parent": "%P",
  "refs": "%D",
  "subject": "%s",
  "author": { "name": "%aN", "email": "%aE", "date": "%ad" },
  "commiter": { "name": "%cN", "email": "%cE", "date": "%cd" }
 },`

func gitCommand(args []string) (string, error) {

	cmd := exec.Command("git", args...)
	out, err := cmd.Output()
	if err != nil {
		return "", errors.Wrapf(err, "failed to run git command: git %s", args)
	}

	logOut := string(out)
	logOut = logOut[:len(logOut)-1] // Remove the last ","

	return logOut, err
}

func gitLogJson(repo, branchName string) ([]gitCommitMock, error) {
	args := []string{
		"-C",
		repo,
		"log",
		"--date=iso-strict",
		"--first-parent",
		"--decorate=full",
		GITFORMAT,
	}
	if branchName != "" {
		args = append(args, branchName)
	} else {
		args = append(args, "--all")
	}

	logOut, err := gitCommand(args)
	if err != nil {
		return nil, errors.Wrap(err, "failed to run git log")
	}
	logOut = fmt.Sprintf("[%s]", logOut) // Add []

	var gitCommitObjList []gitCommitMock
	json.Unmarshal([]byte(logOut), &gitCommitObjList)

	return gitCommitObjList, err
}

func describeHash(repoDir, commitHash string) (string, error) {
	args := []string{
		"-C",
		repoDir,
		"describe",
		commitHash,
		"--tags",
		"--always",
	}

	logOut, err := gitCommand(args)
	if err != nil {
		return "", errors.Wrap(err, "failed to run git describe")
	}

	return logOut, nil
}

func convertLogToRepositoryCommitList(gitCommitObjList []gitCommitMock) []*github.RepositoryCommit {
	var commitsGithubApi []*github.RepositoryCommit
	for _, commitObj := range gitCommitObjList {
		shaString := new(string)
		*shaString = commitObj.Commit

		ghCommit := github.RepositoryCommit{
			SHA: shaString,
		}

		commitsGithubApi = append(commitsGithubApi, &ghCommit)
	}

	return commitsGithubApi
}

func convertLogToReferenceList(gitCommitObjList []gitCommitMock, refsFilter string) []*github.Reference {
	var RefTagsGithubApi []*github.Reference
	for _, commitObj := range gitCommitObjList {
		if strings.Contains(commitObj.Refs, refsFilter) {
			RefTagsGithubApi = append(RefTagsGithubApi, getNewMockReference(&commitObj))
		}
	}

	sort.Slice(RefTagsGithubApi, func(i, j int) bool {
		return RefTagsGithubApi[i].GetRef() < RefTagsGithubApi[j].GetRef()
	})

	return RefTagsGithubApi
}

func getRefFromCommitObjList(gitCommitObjList []gitCommitMock, refName string) (*github.Reference, error) {
	for _, commitObj := range gitCommitObjList {
		if strings.Contains(commitObj.Refs, refName) {
			return getNewMockReference(&commitObj), nil
		}
	}

	return nil, fmt.Errorf("reference %s not found", refName)
}

func getNewMockReference(commitObj *gitCommitMock) *github.Reference {
	refString := new(string)
	shaString := new(string)

	// refactor tag name to fit githubApi format
	commitObj.Refs = strings.Replace(commitObj.Refs, "tag: ", "", 1)
	commitObj.Refs = strings.Replace(commitObj.Refs, "HEAD -> ", "", 1)

	*refString = commitObj.Refs
	*shaString = commitObj.Commit

	ghReference := &github.Reference{
		Ref: refString,
		Object: &github.GitObject{
			SHA: shaString,
		},
	}

	return ghReference
}

// newFakeGithubApi creates a fake interface
func newFakeGithubApi(repoDir string) *mockGithubApi {
	return &mockGithubApi{
		repoDir:    repoDir,
		fakePRList: []*github.PullRequest{},
	}
}

func newFakeGitComponent(api *mockGithubApi, repoDir string, componentParams *component, tagCommitMap map[string]string) *gitComponent {
	componentGitRepo := newLocalGitRepo(repoDir, tagCommitMap)

	gitComponent := &gitComponent{
		configParams:    componentParams,
		githubInterface: api,
		gitRepo:         componentGitRepo,
	}

	return gitComponent
}

func newFakeGitCnaoRepo(api *mockGithubApi, repoDir string, componentParams *component, tagCommitMap map[string]string) *gitCnaoRepo {
	componentGitRepo := newLocalGitRepo(repoDir, tagCommitMap)

	gitComponent := &gitCnaoRepo{
		configParams:    componentParams,
		githubInterface: api,
		gitRepo:         componentGitRepo,
	}

	return gitComponent
}

func newLocalGitRepo(repoDir string, tagCommitMap map[string]string) *gitRepo {
	By(fmt.Sprintf("creating new repository on directory: %s", repoDir))
	repo, err := git.PlainInit(repoDir, false)
	Expect(err).ToNot(HaveOccurred(), "Should succeed cloning git repo")

	initializeRepo(repo, repoDir, tagCommitMap)

	return &gitRepo{
		repo:     repo,
		localDir: repoDir,
	}
}

func initializeRepo(repo *git.Repository, repoDir string, tagCommitMap map[string]string) {
	w, err := repo.Worktree()
	Expect(err).ToNot(HaveOccurred(), "Should succeed getting repo Worktree")

	createCommitWithoutTag(w, tagCommitMap, repoDir, "static", "master", false)
	createCommitWithAnnotatedTag(w, repo, tagCommitMap, repoDir, "tagged_annotated", "v0.0.1", "master")
	createCommitWithLightweightTag(w, repo, tagCommitMap, repoDir, "tagged_lightweight", "v0.0.2", "master")
	createCommitWithoutTag(w, tagCommitMap, repoDir, "latest_master", "master", true)
	createBranch(repo, "release-v1.0.0")
	createCommitWithAnnotatedTag(w, repo, tagCommitMap, repoDir, "tagged_annotated_branch", "v1.0.0", "release-v1.0.0")
	createCommitWithLightweightTag(w, repo, tagCommitMap, repoDir, "tagged_lightweight_branch", "v1.0.1", "release-v1.0.0")
	createCommitWithoutTag(w, tagCommitMap, repoDir, "latest_branch", "release-v1.0.0", true)
	// adding a non-existing commit to check negative scenarios
	tagCommitMap["dummy_false_commit"] = randstr.Hex(40)
}

func createBranch(repo *git.Repository, branchName string) {
	By(fmt.Sprintf("adding a new branch %s from Head", branchName))
	headRef, err := repo.Head()
	Expect(err).ToNot(HaveOccurred(), "Should succeed getting current Head ref")

	ref := plumbing.NewHashReference(plumbing.NewBranchReferenceName(branchName), headRef.Hash())

	err = repo.Storer.SetReference(ref)
	Expect(err).ToNot(HaveOccurred(), "Should succeed setting the branch ref")
}

func createCommit(w *git.Worktree, repoDir, fileName, branchName string) plumbing.Hash {
	By(fmt.Sprintf("committing a new file %s on %s branch", fileName, branchName))
	w.Checkout(&git.CheckoutOptions{Branch: plumbing.NewBranchReferenceName(branchName)})

	fileWithPath := filepath.Join(repoDir, fileName)
	err := ioutil.WriteFile(fileWithPath, []byte(""), 0644)
	Expect(err).ToNot(HaveOccurred(), fmt.Sprintf("Should succeed creating file %s", fileName))

	_, err = w.Add(fileName)
	Expect(err).ToNot(HaveOccurred(), fmt.Sprintf("Should succeed adding %s file to repo tree", fileName))

	commitHash, err := w.Commit(fmt.Sprintf("adding file %s", fileName), &git.CommitOptions{
		Author: &object.Signature{
			Name:  "John Doe",
			Email: "john@doe.org",
			When:  time.Now(),
		},
	})
	Expect(err).ToNot(HaveOccurred(), "Should succeed committing to repo tree")

	return commitHash
}

func createCommitWithoutTag(w *git.Worktree, tagCommitMap map[string]string, repoDir, fileName, branchName string, addDummy bool) {
	By(fmt.Sprintf("committing a new file on %s branch with any tag", branchName))
	commitHash := createCommit(w, repoDir, fileName, branchName)

	if addDummy {
		fakeDummyTag := "dummy_tag_latest_" + branchName
		By(fmt.Sprintf("Adding a dummy tag: %s for commit %s", fakeDummyTag, commitHash.String()))
		tagCommitMap[fakeDummyTag] = commitHash.String()
	}
}

// createCommitWithAnnotatedTag created commit and tags it in the repo.
// In order to simulate output of githubApi methods like listMatchingRefs
// that return the tags list in chronological order, we should set the tag key
// to be sorted by alphabetically+chronologically order. we do this by setting
// tag keys to be the semver version tagged (which is alphabetically+chronologically
// ordered by default). For example, chronologically tagging v0.0.1 and then v0.0.2
// will produce an ordered tag list: ["v0.0.1", "v0.0.2"].
func createCommitWithAnnotatedTag(w *git.Worktree, repo *git.Repository, tagCommitMap map[string]string, repoDir, fileName, tagName, branchName string) {
	By(fmt.Sprintf("committing a new file on %s branch with annotated tag", branchName))
	commitHash := createCommit(w, repoDir, fileName, branchName)

	_, err := repo.CreateTag(tagName, commitHash, &git.CreateTagOptions{
		Tagger: &object.Signature{
			Name:  "John Doe",
			Email: "john@doe.org",
			When:  time.Now(),
		},
		Message: fileName,
	})
	Expect(err).ToNot(HaveOccurred(), fmt.Sprintf("Should succeed adding %s tag to commit Hash %s", tagName, commitHash))

	tagCommitMap[tagName] = commitHash.String()
}

func createCommitWithLightweightTag(w *git.Worktree, repo *git.Repository, tagCommitMap map[string]string, repoDir, fileName, tagName, branchName string) {
	By(fmt.Sprintf("committing a new file on %s branch with lightweight tag", branchName))
	commitHash := createCommit(w, repoDir, fileName, branchName)

	_, err := repo.CreateTag(tagName, commitHash, nil)
	Expect(err).ToNot(HaveOccurred(), fmt.Sprintf("Should succeed adding %s tag to commit Hash %s", tagName, commitHash))

	tagCommitMap[tagName] = commitHash.String()
}

func getFakePrWithTitle(prTitle string) *github.NewPullRequest {
	return &github.NewPullRequest{Title: &prTitle}
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	Expect(err).ToNot(HaveOccurred(), fmt.Sprintf("Should succeed checking if file %s exists", filename))
	return !info.IsDir()
}

func modifyFiles(repoDir string, files []string) {
	for _, fileName := range files {
		var err error
		var f *os.File
		defer f.Close()

		fileNameWithPath := filepath.Join(repoDir, fileName)
		if fileExists(fileNameWithPath) {
			f, err = os.OpenFile(fileNameWithPath, os.O_RDWR, os.ModeAppend)
			Expect(err).ToNot(HaveOccurred(), fmt.Sprintf("should not fail to open file %s", fileNameWithPath))
		} else {
			f, err = os.Create(fileNameWithPath)
			Expect(err).ToNot(HaveOccurred(), fmt.Sprintf("should not fail to create file %s", fileNameWithPath))
		}

		_, err = f.WriteString(fileName)
		Expect(err).ToNot(HaveOccurred(), fmt.Sprintf("should not fail to write generic string (file name) to file %s", fileNameWithPath))
	}
}

func deleteFiles(repoDir string, files []string) {
	for _, fileName := range files {
		fileNameWithPath := filepath.Join(repoDir, fileName)
		if !fileExists(fileNameWithPath) {
			Fail(fmt.Sprintf("file %s cannot be deleted since it doesn't exist", fileName))
		}

		err := os.Remove(fileNameWithPath)
		Expect(err).ToNot(HaveOccurred(), fmt.Sprintf("should not fail to deleting file %s", fileNameWithPath))
	}
}
