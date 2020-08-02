package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/coreos/go-semver/semver"
	"github.com/pkg/errors"
)

var logger *log.Logger

type inputParams struct {
	componentsConfigPath string
	gitToken             string
}

const (
	updatePolicyStatic = "static"
	updatePolicyTagged = "tagged"
	updatePolicyLatest = "latest"
)

func main() {
	logger = initLog()
	logger.Printf("~~~~~~~~Bumper Script~~~~~~~~")

	inputArgs := inputParams{}
	initFlags(&inputArgs)

	githubApi, err := newGithubApi(inputArgs.gitToken)
	if err != nil {
		exitWithError(errors.Wrap(err, "Failed to create github api instance"))
	}

	logger.Printf("Parsing %s", inputArgs.componentsConfigPath)
	componentsConfig, err := parseComponentsYaml(inputArgs.componentsConfigPath)
	if err != nil {
		exitWithError(errors.Wrap(err, "Failed to parse components yaml"))
	}

	for componentName, component := range componentsConfig.Components {
		logger.Printf("~~Checking if %s needs bumping~~", componentName)

		err = printCurrentComponentParams(component)
		if err != nil {
			exitWithError(errors.Wrapf(err, "Failed to print component %s", componentName))
		}

		gitComponent, err := newGitComponent(githubApi, componentName, &component)
		if err != nil {
			exitWithError(errors.Wrapf(err, "Failed to clone %s", componentName))
		}
		defer os.RemoveAll(gitComponent.gitRepo.localDir)

		currentReleaseTag, err := gitComponent.getCurrentReleaseTag()
		if err != nil {
			exitWithError(errors.Wrapf(err, "Failed to get current release version tag from %s", componentName))
		}

		updatedReleaseTag, updatedReleaseCommit, err := gitComponent.getUpdatedReleaseInfo()
		if err != nil {
			exitWithError(errors.Wrapf(err, "Failed to get latest release version tag from %s", componentName))
		}

		bumpNeeded, err := isBumpNeeded(currentReleaseTag, updatedReleaseTag, component.Updatepolicy)
		if err != nil {
			exitWithError(errors.Wrapf(err, "Failed to discover if Bump need for %s", componentName))
		}

		if bumpNeeded {
			logger.Printf("Bumping %s from %s to %s", componentName, currentReleaseTag, updatedReleaseTag)
			// reset --hard git repo
			exitWithError(fmt.Errorf("reset --hader repo not implemented yet"))

			// PR name
			prTitle := fmt.Sprintf("bump %s to %s", componentName, updatedReleaseTag)
			logger.Printf("PR title: %s", prTitle)
			// Create PR
			exitWithError(fmt.Errorf("create PR not implemented yet"))

			// update component's entry in config yaml
			component.Commit = updatedReleaseCommit
			component.Metadata = updatedReleaseTag
			err = updateComponentsYaml(inputArgs.componentsConfigPath, componentsConfig)
			if err != nil {
				exitWithError(errors.Wrap(err, "Failed to update components yaml"))
			}

			cmd := exec.Command("make", fmt.Sprintf("bump-%s", componentName))
			if out, err := cmd.CombinedOutput(); err != nil {
				exitWithError(errors.Wrapf(err, "Failed to run bump script. StdOut = %s", string(out)))
			}

			// create a new branch name
			BranchName := strings.Replace(strings.ToLower(prTitle), " ", "_", -1)
			logger.Printf("Opening new Branch %s", BranchName)

			// push branch to PR
			exitWithError(fmt.Errorf("push branch to PR not implemented yet"))
		} else {
			logger.Printf("Bump not needed in component %s", componentName)
		}
	}
}

func isBumpNeeded(currentReleaseVersion, latestReleaseVersion, updatePolicy string) (bool, error) {
	logger.Printf("currentReleaseVersion: %s, latestReleaseVersion: %s, updatePolicy: %s\n", currentReleaseVersion, latestReleaseVersion, updatePolicy)

	if updatePolicy == updatePolicyStatic {
		logger.Printf("updatePolicy is static. avoiding auto bump")
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

func initLog() *log.Logger {
	var buf bytes.Buffer
	logger := log.New(&buf, "INFO: ", log.LstdFlags)
	logger.SetOutput(os.Stdout)
	return logger
}

func initFlags(paramArgs *inputParams) {
	flag.StringVar(&paramArgs.componentsConfigPath, "config-path", "", "Full path to components yaml")
	flag.StringVar(&paramArgs.gitToken, "token", "", "git Token")
	flag.Parse()
	if flag.NFlag() != 2 {
		exitWithError(fmt.Errorf("Wrong Number of input parameters %d, should be 2. Use --help for usage", flag.NFlag()))
	}
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

func exitWithError(err error) {
	logger.Fatal(errors.Wrap(err, "Exiting with Error"))
	os.Exit(1)
}
