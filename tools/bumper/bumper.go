package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

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

	cnaoRepo, err := getCnaoRepo(githubApi)
	if err != nil {
		exitWithError(errors.Wrap(err, "Failed to clone cnao repo"))
	}

	logger.Printf("Parsing %s", inputArgs.componentsConfigPath)
	componentsConfig, err := cnaoRepo.getComponentsConfig(inputArgs.componentsConfigPath)
	if err != nil {
		exitWithError(errors.Wrap(err, "Failed to get components config"))
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

		proposedPrTitle := fmt.Sprintf("bump %s to %s", componentName, updatedReleaseTag)
		componentBumpNeeded, err := cnaoRepo.isComponentBumpNeeded(currentReleaseTag, updatedReleaseTag, component.Updatepolicy, proposedPrTitle)
		if err != nil {
			exitWithError(errors.Wrapf(err, "Failed to discover if Bump need for %s", componentName))
		}

		if componentBumpNeeded {
			logger.Printf("Bumping %s from %s to %s", componentName, currentReleaseTag, updatedReleaseTag)
			// reset --hard git repo
			exitWithError(fmt.Errorf("reset --hader repo not implemented yet"))

			// Create PR
			exitWithError(fmt.Errorf("create PR not implemented yet"))

			// update components yaml in the bumping repo instance
			componentsConfig, err := cnaoRepo.getComponentsConfig(inputArgs.componentsConfigPath)
			if err != nil {
				exitWithError(errors.Wrap(err, "Failed to get components config during bump"))
			}

			// update component's entry in config yaml
			component.Commit = updatedReleaseCommit
			component.Metadata = updatedReleaseTag
			componentsConfig.Components[componentName] = component
			err = cnaoRepo.updateComponentsConfig(inputArgs.componentsConfigPath, componentsConfig)
			if err != nil {
				exitWithError(errors.Wrap(err, "Failed to update components yaml"))
			}

			logger.Printf("Running bump-%s script", componentName)
			cmd := exec.Command("make", fmt.Sprintf("bump-%s", componentName))
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				exitWithError(errors.Wrapf(err, "Failed to run bump script, \nStdout:\n%s\nStderr:\n%s", cmd.Stdout, cmd.Stderr))
			}

			// create a new branch name
			branchName := strings.Replace(strings.ToLower(proposedPrTitle), " ", "_", -1)
			logger.Printf("Opening new Branch %s", branchName)

			// push branch to PR
			exitWithError(fmt.Errorf("push branch to PR not implemented yet"))
		} else {
			logger.Printf("Bump not needed in component %s", componentName)
		}
	}
}

func initLog() *log.Logger {
	var buf bytes.Buffer
	logger := log.New(&buf, "INFO: ", log.LstdFlags)
	logger.SetOutput(os.Stdout)
	return logger
}

func initFlags(paramArgs *inputParams) {
	flag.StringVar(&paramArgs.componentsConfigPath, "config-path", "", "relative path to components yaml from bumping repo")
	flag.StringVar(&paramArgs.gitToken, "token", "", "git Token")
	flag.Parse()
	if flag.NFlag() != 2 {
		exitWithError(fmt.Errorf("Wrong Number of input parameters %d, should be 2. Use --help for usage", flag.NFlag()))
	}
}

func exitWithError(err error) {
	logger.Fatal(errors.Wrap(err, "Exiting with Error"))
	os.Exit(1)
}
