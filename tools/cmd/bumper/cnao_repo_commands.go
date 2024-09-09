package main

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/coreos/go-semver/semver"
	"github.com/go-git/go-git/v5"
	"github.com/gobwas/glob"
	"github.com/google/go-github/v32/github"
	"github.com/pkg/errors"
)

type gitCnaoRepo struct {
	configParams *component

	githubInterface githubInterface

	gitRepo *gitRepo
}

const (
	repoUrl         = "https://github.com/kubevirt/cluster-network-addons-operator"
	allowListString = "components.yaml,data/*,test/releases/99.0.0.go,pkg/components/components.go"
)

func getCnaoRepo(api *githubApi, baseBranch string) (*gitCnaoRepo, error) {
	cnaoGitRepo, err := openGitRepo(".")
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to get git repo for cnao repo")
	}

	cnaoComponentParams := &component{
		Url:    repoUrl,
		Branch: baseBranch,
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

// resetRepo is a wrapper for resetInAllowedList
func (cnaoRepoOps *gitCnaoRepo) reset() error {
	return cnaoRepoOps.resetInAllowedList(getAllowedList())
}

// resetInAllowedList resets all the files in AllowedList.
func (cnaoRepoOps *gitCnaoRepo) resetInAllowedList(allowList []string) error {
	if len(allowList) == 0 {
		return nil
	}
	logger.Printf("Cleaning untracked files on bumping repo")
	// TODO replace this when go-git adds advanced clean abilities so we could clean specific paths
	cleanArgs := append([]string{"-C", cnaoRepoOps.gitRepo.localDir, "clean", "-fd", "--"}, allowList...)
	err := runExternalGitCommand(cleanArgs)
	if err != nil {
		return errors.Wrapf(err, "Failed to clean bumping repo")
	}

	logger.Printf("Resetting modified files in allowed list on bumping repo")
	// TODO replace this when go-git adds advanced checkout/restore abilities so we could checkout specific paths
	checkoutArgs := append([]string{"-C", cnaoRepoOps.gitRepo.localDir, "checkout", "HEAD", "--"}, allowList...)
	err = runExternalGitCommand(checkoutArgs)
	if err != nil {
		return errors.Wrapf(err, "Failed to checkout bumping repo")
	}

	return nil
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

	prList, _, err := cnaoRepoOps.githubInterface.listPullRequests(cnaoRepoOps.getCnaoRepoOwnerFromUrl(), cnaoRepoOps.getCnaoRepoNameFromUrl(), cnaoRepoOps.configParams.Branch)
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

// collectBumpFile is a wrapper for collectModifiedToTreeList
func (cnaoRepoOps *gitCnaoRepo) collectBumpFile() ([]*github.TreeEntry, error) {
	return cnaoRepoOps.collectModifiedToTreeList(getAllowedList())
}

// collectModifiedToTreeList collects the modified files in the allowedList paths and returns a github tree entry list
func (cnaoRepoOps *gitCnaoRepo) collectModifiedToTreeList(allowedList []string) ([]*github.TreeEntry, error) {
	w, err := cnaoRepoOps.gitRepo.repo.Worktree()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to get bumping repo Worktree")
	}

	status, err := w.Status()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to get bumping repo status")
	} else if status.IsClean() {
		return []*github.TreeEntry{}, nil
	}

	// Create a tree with what to commit.
	var entries []*github.TreeEntry
	for localFile, status := range status {
		if status.Staging == git.Unmodified && status.Worktree == git.Unmodified {
			continue
		}

		fileNameWithPath := filepath.Join(cnaoRepoOps.gitRepo.localDir, localFile)
		var treeEntryContent *string
		if status.Worktree != git.Deleted {
			content, err := ioutil.ReadFile(fileNameWithPath)
			if err != nil {
				return nil, errors.Wrapf(err, "Failed to read local file %s", fileNameWithPath)
			}
			treeEntryContent = github.String(string(content))
		}

		if fileInGlobList(localFile, allowedList) {
			logger.Printf("File added to tree: %s", localFile)
			entries = append(entries, &github.TreeEntry{Path: github.String(localFile), Type: github.String("blob"), Content: treeEntryContent, Mode: github.String("100644")})
		} else {
			logger.Printf("Skipping file %s, not in allowed list", localFile)
		}
	}

	return entries, nil
}

// getNewBumpBranch creates a new bump branch.
// Since we don't want to use already open branch
// that may hold outdated/needs rebasing changes
// we concat a random string to the end of the branch name
func (cnaoRepoOps *gitCnaoRepo) getNewBumpBranch(prTitle string) (*github.Reference, string, error) {
	branchBaseName := strings.Replace(strings.ToLower(prTitle), " ", "_", -1)

	branchNameWithRandSuffix := fmt.Sprintf("%s_%s", branchBaseName, generateRandString())
	_, resp, err := cnaoRepoOps.githubInterface.getBranchRef(cnaoRepoOps.getCnaoRepoOwnerFromUrl(), cnaoRepoOps.getCnaoRepoNameFromUrl(), "refs/heads/"+branchNameWithRandSuffix)
	if err != nil {
		if resp.Response.StatusCode == http.StatusNotFound {
			return cnaoRepoOps.createNewGithubBranch(branchNameWithRandSuffix)
		} else {
			return nil, "", errors.Wrap(err, "Failed to check if branch is already open")
		}
	}
	return nil, "", fmt.Errorf("Failed to open new branch %s, already open", branchNameWithRandSuffix)
}

func (cnaoRepoOps *gitCnaoRepo) createNewGithubBranch(newBranchName string) (*github.Reference, string, error) {
	logger.Printf("Creating new branch with githubApi: %s", newBranchName)
	baseRef, _, err := cnaoRepoOps.githubInterface.getBranchRef(cnaoRepoOps.getCnaoRepoOwnerFromUrl(), cnaoRepoOps.getCnaoRepoNameFromUrl(), "refs/heads/"+cnaoRepoOps.configParams.Branch)
	if err != nil {
		return nil, "", errors.Wrapf(err, "Failed to get origin/%s github ref", cnaoRepoOps.configParams.Branch)
	}

	newRef := &github.Reference{Ref: github.String("refs/heads/" + newBranchName), Object: &github.GitObject{SHA: baseRef.Object.SHA}}
	newBranchRef, _, err := cnaoRepoOps.githubInterface.createBranchRef(cnaoRepoOps.getCnaoRepoOwnerFromUrl(), cnaoRepoOps.getCnaoRepoNameFromUrl(), newRef)
	if err != nil {
		return nil, "", errors.Wrap(err, "Failed to create new branch ref")
	}

	return newBranchRef, newBranchName, nil
}

// pushCommit creates the commit in the given reference using the given tree.
func (cnaoRepoOps *gitCnaoRepo) pushCommit(commitTitle string, branch *github.Reference, tree *github.Tree) error {
	logger.Printf("Pushing new commit")
	// Get the parent commit to attach the commit to.
	parent, _, err := cnaoRepoOps.githubInterface.getCommit(cnaoRepoOps.getCnaoRepoOwnerFromUrl(), cnaoRepoOps.getCnaoRepoNameFromUrl(), *branch.Object.SHA)
	if err != nil {
		return errors.Wrap(err, "Failed to get parent commit")
	}

	date := time.Now()
	authorName := ""
	authorEmail := ""
	repoConfig, err := cnaoRepoOps.gitRepo.repo.Config()
	// if User is not set then assume Bot is operating Bump
	if repoConfig.User.Name == "" || repoConfig.User.Email == "" {
		authorName = "CNAO Bump Bot"
		authorEmail = "noreply@github.com"
	} else {
		authorName = repoConfig.User.Name
		authorEmail = repoConfig.User.Email
	}
	commitMessage := fmt.Sprintf("%s\n\nSigned-off-by: %s <%s>", commitTitle, authorName, authorEmail)

	author := &github.CommitAuthor{Date: &date, Name: &authorName, Email: &authorEmail}
	commit := &github.Commit{Author: author, Message: &commitMessage, Tree: tree, Parents: []*github.Commit{parent}}
	newCommit, _, err := cnaoRepoOps.githubInterface.createCommit(cnaoRepoOps.getCnaoRepoOwnerFromUrl(), cnaoRepoOps.getCnaoRepoNameFromUrl(), commit)
	if err != nil {
		return errors.Wrap(err, "Failed to create new commit")
	}

	branch.Object.SHA = newCommit.SHA
	_, _, err = cnaoRepoOps.githubInterface.updateRef(cnaoRepoOps.getCnaoRepoOwnerFromUrl(), cnaoRepoOps.getCnaoRepoNameFromUrl(), branch, false)
	if err != nil {
		return errors.Wrap(err, "Failed to attach the commit to the branch.")
	}
	return nil
}

func (cnaoRepoOps *gitCnaoRepo) createPR(prTitle, branchName string) (*github.PullRequest, error) {
	logger.Printf("Creating new PR")
	prBranch := cnaoRepoOps.configParams.Branch
	prDescription := fmt.Sprintf("%s\nExecuted by Bumper script\n\n```release-note\n%s\n```", prTitle, prTitle)
	newPR := &github.NewPullRequest{
		Title:               &prTitle,
		Head:                &branchName,
		Base:                &prBranch,
		Body:                &prDescription,
		MaintainerCanModify: github.Bool(true),
	}

	pr, _, err := cnaoRepoOps.githubInterface.createPullRequest(cnaoRepoOps.getCnaoRepoOwnerFromUrl(), cnaoRepoOps.getCnaoRepoNameFromUrl(), newPR)
	if err != nil {
		return nil, err
	}

	logger.Printf("PR created: %s\n", pr.GetHTMLURL())
	return pr, nil
}

// prepareBumpPr gathers all modified files, and pushed them to a PR using github API
func (cnaoRepoOps *gitCnaoRepo) generateBumpPr(prTitle string, modifiedFilesList []*github.TreeEntry) (*github.PullRequest, error) {
	branchRef, branchName, err := cnaoRepoOps.getNewBumpBranch(prTitle)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create bump branch")
	}

	fileTree, _, err := cnaoRepoOps.githubInterface.createTree(cnaoRepoOps.getCnaoRepoOwnerFromUrl(), cnaoRepoOps.getCnaoRepoNameFromUrl(), *branchRef.Object.SHA, modifiedFilesList)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to generate github file tree")
	}

	err = cnaoRepoOps.pushCommit(prTitle, branchRef, fileTree)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create and push the commit")
	}

	pr, err := cnaoRepoOps.createPR(prTitle, branchName)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create bump PR instance")
	}

	return pr, nil
}

func (cnaoRepoOps *gitCnaoRepo) getCnaoRepoNameFromUrl() string {
	urlSlice := strings.Split(cnaoRepoOps.configParams.Url, "/")
	return urlSlice[len(urlSlice)-1]
}

func (cnaoRepoOps *gitCnaoRepo) getCnaoRepoOwnerFromUrl() string {
	urlSlice := strings.Split(cnaoRepoOps.configParams.Url, "/")
	return urlSlice[len(urlSlice)-2]
}

func fileInGlobList(fileName string, globList []string) bool {
	isAnyMatch := false
	for _, allowedGlob := range globList {
		g := glob.MustCompile(allowedGlob)

		if g.Match(fileName) {
			isAnyMatch = true
		}
	}
	return isAnyMatch
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

func runExternalGitCommand(args []string) error {
	// TODO replace this when go-git adds advanced checkout/restore/clean abilities so we could checkout specific paths
	cmd := exec.Command("git", args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	if err != nil {
		return errors.Wrapf(err, "Failed to run git command: git %s\nStdout:\n%s\nStderr:\n%s", strings.Join(args, " "), cmd.Stdout, cmd.Stderr)
	}
	return nil
}

// AllowedList is a string array of file globs, used to fine-pick the changes we want to bump.
func getAllowedList() []string {
	return strings.Split(allowListString, ",")
}

func generateRandString() string {
	charset := []rune("abcdefghijklmnopqrstuvwxyz" + "ABCDEFGHIJKLMNOPQRSTUVWXYZ" + "0123456789")
	rand.Seed(time.Now().UnixNano())
	length := 5
	var b strings.Builder
	for i := 0; i < length; i++ {
		b.WriteRune(charset[rand.Intn(len(charset))])
	}
	return b.String()
}
