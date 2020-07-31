package main

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/google/go-github/v32/github"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
)

type gitComponent struct {
	configParams *component

	githubApi *githubApi

	gitRepo *gitRepo
}

type githubApi struct {
	client *github.Client

	// context needed for github api
	ctx context.Context
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
		configParams: componentParams,
		githubApi:    api,
		gitRepo:      componentGitRepo,
	}

	return gitComponent, nil
}

// establishes connection with the github Api server using the token
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

	_, _, err := githubApi.client.Users.Get(ctx, "")
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to get user from github client API")
	}

	return githubApi, nil
}

func newGitRepo(componentName string, componentParams *component) (*gitRepo, error) {
	repoDir, err := ioutil.TempDir("/tmp", componentName)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create temp dir for component")
	}

	logger.Printf("Cloning to temp directory: %s", repoDir)
	repo, err := git.PlainClone(repoDir, false, &git.CloneOptions{
		URL:           componentParams.Url,
		ReferenceName: plumbing.NewBranchReferenceName(componentParams.Branch),
	})
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to clone %s repo", componentName)
	}

	return &gitRepo{
		repo: repo,
		localDir:  repoDir,
	}, nil
}

func (mgr *gitComponent) getCurrentReleaseTag() (string, error) {
	return "", fmt.Errorf("getCurrentReleaseTag not implemented")
}

func (mgr *gitComponent) getUpdatedReleaseInfo() (string, string, error) {
	return "", "", fmt.Errorf("getUpdatedReleaseInfo not implemented")
}
