package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/coreos/go-semver/semver"
	"github.com/go-git/go-git/v5"
	"github.com/pkg/errors"
)

type gitCnaoRepo struct {
	configParams *component

	githubInterface githubInterface

	gitRepo *gitRepo
}

const (
	repoUrl = "https://github.com/kubevirt/cluster-network-addons-operator"
)

func getCnaoRepo(api *githubApi) (*gitCnaoRepo, error) {
	cnaoGitRepo, err := openGitRepo(".")
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to get git repo for cnao repo")
	}

	cnaoComponentParams := &component{
		Url: repoUrl,
	}

	gitCnaoRepo := &gitCnaoRepo{
		configParams:    cnaoComponentParams,
		githubInterface: api,
		gitRepo:         cnaoGitRepo,
	}

	return gitCnaoRepo, nil
}

// openGitRepo opens an existing repository
func openGitRepo(repoPath string) (*gitRepo, error) {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to open existing repo in path %s", repoPath)
	}

	return &gitRepo{
		repo:     repo,
		localDir: repoPath,
	}, nil
}

func (cnaoRepoOps *gitCnaoRepo) getComponentsConfig(relativeConfigPath string) (componentsConfig, error) {
	configPath := filepath.Join(cnaoRepoOps.gitRepo.localDir, relativeConfigPath)
	return parseComponentsYaml(configPath)
}

func (cnaoRepoOps *gitCnaoRepo) updateComponentsConfig(relativeConfigPath string, componentsConfig componentsConfig) error {
	configPath := filepath.Join(cnaoRepoOps.gitRepo.localDir, relativeConfigPath)
	return updateComponentsYaml(configPath, componentsConfig)
}

func (cnaoRepoOps *gitCnaoRepo) bumpComponent(componentName string) error {
	logger.Printf("Running bump-%s script", componentName)
	cmd := exec.Command("make", "-C", cnaoRepoOps.gitRepo.localDir, fmt.Sprintf("bump-%s", componentName))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return errors.Wrapf(err, "Failed to run bump script, \nStdout:\n%s\nStderr:\n%s", cmd.Stdout, cmd.Stderr)
	}
	return nil
}

func (cnaoRepoOps *gitCnaoRepo) isComponentBumpNeeded(currentReleaseVersion, latestReleaseVersion, updatePolicy, proposedPrTitle string) (bool, error) {
	logger.Printf("currentReleaseVersion: %s, latestReleaseVersion: %s, updatePolicy: %s\n", currentReleaseVersion, latestReleaseVersion, updatePolicy)

	if updatePolicy == updatePolicyStatic {
		logger.Printf("updatePolicy is static. avoiding auto bump")
		return false, nil
	}

	// check if PR not already opened
	isAlreadyOpened, err := cnaoRepoOps.isPrAlreadyOpened(proposedPrTitle)
	if err != nil {
		return false, errors.Wrapf(err, "Failed to check if PR already open")
	}
	if isAlreadyOpened {
		logger.Printf("Bump PR for the latest version already exist. Aborting auto bump")
		return false, nil
	}

	// if one of the tags is in vtag format (e.g 0.39.0-32-g1fcbe815), and not equal, then always bump
	if isVtagFormat(currentReleaseVersion) || isVtagFormat(latestReleaseVersion) {
		return currentReleaseVersion == latestReleaseVersion, nil
	}

	currentVersion, err := canonicalizeVersion(currentReleaseVersion)
	if err != nil {
		return false, errors.Wrapf(err, "Failed to digest current Version %s to semver", currentVersion)
	}
	latestVersion, err := canonicalizeVersion(latestReleaseVersion)
	if err != nil {
		return false, errors.Wrapf(err, "Failed to digest latest Version %s to semver", latestVersion)
	}

	return currentVersion.LessThan(*latestVersion), nil
}

// check if there is an already open PR with the same title in the repo.
func (cnaoRepoOps *gitCnaoRepo) isPrAlreadyOpened(proposedPrTitle string) (bool, error) {
	logger.Printf("checking if there is an already open bump PR for this release")

	prList, _, err := cnaoRepoOps.githubInterface.ListPullRequests(cnaoRepoOps.getCnaoRepoOwnerFromUrl(), cnaoRepoOps.getCnaoRepoNameFromUrl())
	if err != nil {
		return false, errors.Wrapf(err, "Failed to get list of PRs from %s/%s repo", cnaoRepoOps.getCnaoRepoOwnerFromUrl(), cnaoRepoOps.getCnaoRepoNameFromUrl())
	}

	for _, pr := range prList {
		if pr.GetTitle() == proposedPrTitle {
			return true, nil
		}
	}

	return false, nil
}

func (cnaoRepoOps *gitCnaoRepo) getCnaoRepoNameFromUrl() string {
	urlSlice := strings.Split(cnaoRepoOps.configParams.Url, "/")
	return urlSlice[len(urlSlice)-1]
}

func (cnaoRepoOps *gitCnaoRepo) getCnaoRepoOwnerFromUrl() string {
	urlSlice := strings.Split(cnaoRepoOps.configParams.Url, "/")
	return urlSlice[len(urlSlice)-2]
}

// since versioning of components can sometimes divert from semver standard, we do some refactoring
func canonicalizeVersion(version string) (*semver.Version, error) {
	// remove trailing "v" if exists
	version = strings.TrimPrefix(version, "v")

	// expand to 2 dotted format
	versionSectionsNum := len(strings.Split(version, "."))
	switch versionSectionsNum {
	case 2:
		version = version + ".0"
	case 3:
		break
	default:
		return nil, fmt.Errorf("Failed to refactor version string %s", version)
	}

	return semver.NewVersion(version)
}

// check vtag format (example: 0.39.0-32-g1fcbe815)
func isVtagFormat(tagVersion string) bool {
	var vtagSyntax = regexp.MustCompile(`^[0-9]\.[0-9]+\.*[0-9]*-[0-9]+-g[0-9,a-f]{7}`)
	return vtagSyntax.MatchString(tagVersion)
}
