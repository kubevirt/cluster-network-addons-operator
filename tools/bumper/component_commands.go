package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/google/go-github/v32/github"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
)

type gitComponent struct {
	configParams *component

	githubInterface githubInterface

	gitRepo *gitRepo
}

type githubApi struct {
	client *github.Client

	// context needed for github api
	ctx context.Context
}

type githubInterface interface {
	listMatchingRefs(owner, repo string, opts *github.ReferenceListOptions) ([]*github.Reference, *github.Response, error)
	listCommits(owner, repo string, opts *github.CommitsListOptions) ([]*github.RepositoryCommit, *github.Response, error)
	getBranchRef(owner string, repo string, ref string) (*github.Reference, *github.Response, error)
	createBranchRef(owner string, repo string, ref *github.Reference) (*github.Reference, *github.Response, error)
	createTree(owner string, repo string, baseTree string, entries []*github.TreeEntry) (*github.Tree, *github.Response, error)
	getCommit(owner string, repo string, sha string) (*github.Commit, *github.Response, error)
	createCommit(owner string, repo string, commit *github.Commit) (*github.Commit, *github.Response, error)
	updateRef(owner string, repo string, ref *github.Reference, force bool) (*github.Reference, *github.Response, error)
	listPullRequests(owner string, repo string, branch string) ([]*github.PullRequest, *github.Response, error)
	createPullRequest(owner string, repo string, pull *github.NewPullRequest) (*github.PullRequest, *github.Response, error)
}

func (g githubApi) listMatchingRefs(owner, repo string, opts *github.ReferenceListOptions) ([]*github.Reference, *github.Response, error) {
	return g.client.Git.ListMatchingRefs(g.ctx, owner, repo, opts)
}

func (g githubApi) listCommits(owner, repo string, opts *github.CommitsListOptions) ([]*github.RepositoryCommit, *github.Response, error) {
	return g.client.Repositories.ListCommits(g.ctx, owner, repo, opts)
}

func (g githubApi) getBranchRef(owner string, repo string, ref string) (*github.Reference, *github.Response, error) {
	return g.client.Git.GetRef(g.ctx, owner, repo, ref)
}

func (g githubApi) createBranchRef(owner string, repo string, newRef *github.Reference) (*github.Reference, *github.Response, error) {
	return g.client.Git.CreateRef(g.ctx, owner, repo, newRef)
}

func (g githubApi) createTree(owner string, repo string, baseTree string, entries []*github.TreeEntry) (*github.Tree, *github.Response, error) {
	return g.client.Git.CreateTree(g.ctx, owner, repo, baseTree, entries)
}

func (g githubApi) getCommit(owner string, repo string, sha string) (*github.Commit, *github.Response, error) {
	return g.client.Git.GetCommit(g.ctx, owner, repo, sha)
}

func (g githubApi) createCommit(owner string, repo string, commit *github.Commit) (*github.Commit, *github.Response, error) {
	return g.client.Git.CreateCommit(g.ctx, owner, repo, commit)
}

func (g githubApi) updateRef(owner string, repo string, ref *github.Reference, force bool) (*github.Reference, *github.Response, error) {
	return g.client.Git.UpdateRef(g.ctx, owner, repo, ref, force)
}

func (g githubApi) listPullRequests(owner string, repo string, branch string) ([]*github.PullRequest, *github.Response, error) {
	return g.client.PullRequests.List(g.ctx, owner, repo, &github.PullRequestListOptions{State: "open", Base: branch})
}

func (g githubApi) createPullRequest(owner string, repo string, pull *github.NewPullRequest) (*github.PullRequest, *github.Response, error) {
	return g.client.PullRequests.Create(g.ctx, owner, repo, pull)
}

type gitRepo struct {
	repo *git.Repository

	localDir string
}

func newGitComponent(api *githubApi, componentName string, componentParams *component) (*gitComponent, error) {
	componentGitRepo, err := newGitRepo(componentName, componentParams)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to clone git repo for component %s", componentName)
	}

	gitComponent := &gitComponent{
		configParams:    componentParams,
		githubInterface: api,
		gitRepo:         componentGitRepo,
	}

	return gitComponent, nil
}

// newGithubApi establishes connection with the github Api server using the token
func newGithubApi(token string) (*githubApi, error) {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	githubApi := &githubApi{
		client: github.NewClient(tc),
		ctx:    ctx,
	}

	return githubApi, nil
}

// newGitRepo clones the repository on a local temp directory.
func newGitRepo(componentName string, componentParams *component) (*gitRepo, error) {
	repoDir, err := ioutil.TempDir("/tmp", componentName)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create temp dir for component")
	}

	logger.Printf("Cloning to temp directory: %s", repoDir)
	repo, err := git.PlainClone(repoDir, false, &git.CloneOptions{
		URL:           componentParams.Url,
		ReferenceName: plumbing.NewBranchReferenceName(componentParams.Branch),
		Progress:      os.Stdout,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to clone %s repo", componentName)
	}

	return &gitRepo{
		repo:     repo,
		localDir: repoDir,
	}, nil
}

// getCurrentReleaseTag gets the tag name of currently configured commit sha in this component.
// since some tags are represented by the tag sha and not the commit sha, we also need to convert it to commit sha.
func (componentOps *gitComponent) getCurrentReleaseTag() (string, error) {
	repo := componentOps.getComponentNameFromUrl()
	owner := componentOps.getComponentOwnerFromUrl()
	logger.Printf("Getting current tag in repo %s sha %s", repo, componentOps.configParams.Commit)

	tagRefs, _, err := componentOps.githubInterface.listMatchingRefs(owner, repo, &github.ReferenceListOptions{Ref: "refs/tags"})
	if err != nil {
		return "", errors.Wrap(err, "Failed to get release tag refs from github client API")
	}

	// look for the tag that belongs to the chosen commit, going over from last tag
	for i := len(tagRefs) - 1; i >= 0; i-- {
		tag := tagRefs[i]
		tagName := strings.Replace(tag.GetRef(), "refs/tags/", "", 1)
		commitShaOfTag, err := componentOps.getTagCommitSha(tagName)
		if err != nil {
			return "", errors.Wrap(err, "Failed to get commit-sha from tag name")
		}

		if componentOps.configParams.Commit == commitShaOfTag {
			return tagName, nil
		}
	}

	//if not found, then return the vtag of the current commit
	return componentOps.getVirtualTag(componentOps.configParams.Commit)
}

// getUpdatedReleaseInfo gets the most updated tag and associated commit-sha from component's github api
// according to the chosen Update policy.
func (componentOps *gitComponent) getUpdatedReleaseInfo() (string, string, error) {
	repo := componentOps.getComponentNameFromUrl()
	owner := componentOps.getComponentOwnerFromUrl()
	switch componentOps.configParams.Updatepolicy {
	case updatePolicyTagged:
		logger.Printf("update policy %s will updated to the latest tagged commit in the referenced branch", updatePolicyTagged)
		return componentOps.getLatestTaggedFromBranch(repo, owner, componentOps.configParams.Branch, componentOps.gitRepo.localDir)
	case updatePolicyLatest:
		logger.Printf("update policy %s will update to the latest HEAD in the referenced branch", updatePolicyTagged)
		return componentOps.getLatestFromBranch(repo, owner, componentOps.configParams.Branch, componentOps.gitRepo.localDir)
	case updatePolicyStatic:
		logger.Printf("update policy %s will disable auto update", updatePolicyStatic)
		return componentOps.configParams.Metadata, componentOps.configParams.Commit, nil
	default:
		return "", "", fmt.Errorf("Error: Update strategy %s not supported", componentOps.configParams.Updatepolicy)
	}
}

// getLatestTaggedFromBranch get the latest updated tag and associated commit-sha under a given branch, using "tagged" Update policy
// since some tags are represented by the tag sha and not the commit sha, we also need to convert it to commit sha.
func (componentOps *gitComponent) getLatestTaggedFromBranch(repo, owner, branch, repoDir string) (string, string, error) {
	logger.Printf("Getting latest tagged from branch %s in repo %s", branch, repo)
	tagRefs, _, err := componentOps.githubInterface.listMatchingRefs(owner, repo, &github.ReferenceListOptions{Ref: "refs/tags"})
	if err != nil {
		return "", "", errors.Wrap(err, "Failed to get release tag refs from github client API")
	}

	const (
		maxPagePaginate = 10
		pageSize        = 100
	)
	// look for the first tag that belongs to the chosen branch, going over from last tag
	for i := len(tagRefs) - 1; i >= 0; i-- {
		tag := tagRefs[i]
		tagName := strings.Replace(tag.GetRef(), "refs/tags/", "", 1)
		commitShaOfTag, err := componentOps.getTagCommitSha(tagName)
		if err != nil {
			return "", "", errors.Wrap(err, "Failed to get commit-sha from tag name")
		}

		for pageIdx := 1; pageIdx <= maxPagePaginate; pageIdx++ {
			branchCommits, resp, err := componentOps.githubInterface.listCommits(owner, repo, &github.CommitsListOptions{ListOptions: github.ListOptions{PerPage: pageSize, Page: pageIdx}, SHA: branch})
			if err != nil {
				return "", "", errors.Wrap(err, "Failed to get release tag refs from github client API")
			}
			if resp.NextPage == 0 {
				break
			}

			for _, commit := range branchCommits {
				if commit.GetSHA() == commitShaOfTag {
					return tagName, commit.GetSHA(), nil
				}
			}
		}
	}

	return "", "", fmt.Errorf("Error: tag not found in branch %s on last %d commits\n", branch, maxPagePaginate*pageSize)
}

// getLatestFromBranch get the latest HEAD commit-sha under a given branch, using "latest" Update policy
// since a this commit-sha is not necessarily tagged, we use the v-tag format in case needed.
func (componentOps *gitComponent) getLatestFromBranch(repo, owner, branch, repoDir string) (string, string, error) {
	logger.Printf("Getting Latest HEAD from branch %s in repo %s", branch, repo)
	branchRef, _, err := componentOps.githubInterface.getBranchRef(owner, repo, "refs/heads/"+branch)
	if err != nil {
		return "", "", errors.Wrapf(err, "Failed to get latest HEAD ref of branch %s from github client API", branch)
	}
	updatedCommit := branchRef.GetObject().GetSHA()

	// get virtual tag using git api
	vtag, err := componentOps.getVirtualTag(updatedCommit)
	if err != nil {
		return "", "", errors.Wrapf(err, "Failed to get virtual tag of commit sha %s", updatedCommit)
	}

	return vtag, updatedCommit, nil
}

// getTagCommitSha gets the commit sha from tag name
func (componentOps *gitComponent) getTagCommitSha(tagName string) (string, error) {
	tagRef, err := componentOps.gitRepo.repo.Tag(tagName)
	if err != nil {
		return "", errors.Wrapf(err, "Failed to get tag %s from git Repo", tagName)
	}

	// if annotated tag then the sha is pointing to the tag sha, and we need to fetch the commit sha from it
	// TagObject() returns plumbing.ErrObjectNotFound if the tag is not annotated
	tagObj, err := componentOps.gitRepo.repo.TagObject(tagRef.Hash())
	if err == plumbing.ErrObjectNotFound {
		return tagRef.Hash().String(), nil
	} else if err != nil {
		return "", errors.Wrapf(err, "Failed to get tag object %s from git Repo", tagName)
	}

	tagCommit, err := tagObj.Commit()
	if err != nil {
		return "", errors.Wrapf(err, "Failed to get tag commit of annotated tag %s:\n%v", tagName, tagObj)
	}
	return tagCommit.Hash.String(), nil
}

// getVirtualTag mimics  "git describe --tags" command
func (componentOps *gitComponent) getVirtualTag(updatedCommit string) (string, error) {
	commitHash := plumbing.NewHash(updatedCommit)
	cIter, err := componentOps.gitRepo.repo.Log(&git.LogOptions{
		From:  commitHash,
		Order: git.LogOrderCommitterTime,
	})
	if err != nil {
		return "", errors.Wrapf(err, "Failed to get commit log from hash %s", updatedCommit)
	}

	tagsMap, err := componentOps.getTagMap()
	if err != nil {
		return "", errors.Wrap(err, "Failed to get tags map")
	}

	// Search the tag
	var tag plumbing.ReferenceName
	var count int
	err = cIter.ForEach(func(c *object.Commit) error {
		if t, ok := tagsMap[c.Hash]; ok {
			tag = t
		}
		if tag != "" {
			return errors.New("stop iter")
		}
		count++
		return nil
	})

	if tag == "" {
		return "", fmt.Errorf("there is no tag behind commit %s", updatedCommit)
	}

	if count == 0 {
		return fmt.Sprint(tag.Short()), nil
	} else {
		return fmt.Sprintf("%v-%v-g%v",
			tag.Short(),
			count,
			commitHash.String()[0:7],
		), nil
	}
}

func (componentOps *gitComponent) getTagMap() (map[plumbing.Hash]plumbing.ReferenceName, error) {
	TagsMap := make(map[plumbing.Hash]plumbing.ReferenceName)
	tags, err := componentOps.gitRepo.repo.Tags()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to get tags from git repo")
	}

	err = tags.ForEach(func(t *plumbing.Reference) error {
		commitSha, err := componentOps.getTagCommitSha(t.Name().Short())
		if err != nil {
			return errors.Wrapf(err, "Failed to get commit sha from tag %v", t)
		}
		TagsMap[plumbing.NewHash(commitSha)] = t.Name()
		return nil
	})

	return TagsMap, err
}

func (componentOps *gitComponent) getComponentNameFromUrl() string {
	urlSlice := strings.Split(componentOps.configParams.Url, "/")
	return urlSlice[len(urlSlice)-1]
}

func (componentOps *gitComponent) getComponentOwnerFromUrl() string {
	urlSlice := strings.Split(componentOps.configParams.Url, "/")
	return urlSlice[len(urlSlice)-2]
}
