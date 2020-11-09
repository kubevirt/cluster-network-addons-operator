package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/pkg/errors"
)

var logger *log.Logger

type inputParams struct {
	componentsConfigPath string
	gitToken             string
	baseBranch           string
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

	cnaoRepo, err := getCnaoRepo(githubApi, inputArgs.baseBranch)
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

			err = handleBump(cnaoRepo, component, componentName, inputArgs.componentsConfigPath, updatedReleaseTag, updatedReleaseCommit, proposedPrTitle)
			if err != nil {
				exitWithError(errors.Wrapf(err, "Failed to bump component %s", componentName))
			}
		} else {
			logger.Printf("Bump not needed in component %s", componentName)
		}
	}
}

func handleBump(cnaoRepo *gitCnaoRepo, component component, componentName, componentsConfigPath, updatedReleaseTag, updatedReleaseCommit, proposedPrTitle string) error {
	defer func() {
		err := cnaoRepo.reset()
		if err != nil {
			exitWithError(errors.Wrapf(err, "Failed to bump component %s: reset repo failed", componentName))
		}
	}()

	// update components yaml in the bumping repo instance
	componentsConfig, err := cnaoRepo.getComponentsConfig(componentsConfigPath)
	if err != nil {
		return errors.Wrap(err, "Failed to get components config during bump")
	}

	// update component's entry in config yaml
	component.Commit = updatedReleaseCommit
	component.Metadata = updatedReleaseTag
	componentsConfig.Components[componentName] = component
	err = cnaoRepo.updateComponentsConfig(componentsConfigPath, componentsConfig)
	if err != nil {
		return errors.Wrap(err, "Failed to update components yaml")
	}

	err = cnaoRepo.bumpComponent(componentName)
	if err != nil {
		return errors.Wrap(err, "Failed to bump component")
	}

	logger.Printf("Gather bump output files to list")
	bumpFilesList, err := cnaoRepo.collectBumpFile()
	if err != nil {
		return errors.Wrap(err, "Failed to collect bump output files")
	}
	if len(bumpFilesList) == 0 {
		logger.Printf("No modified/untracked files to bump. Aborting bump.")
		return nil
	}

	logger.Printf("Generate Bump PR using GithubAPI")
	_, err = cnaoRepo.generateBumpPr(proposedPrTitle, bumpFilesList)
	if err != nil {
		exitWithError(errors.Wrap(err, "Failed to generate Bump PR"))
	}

	return nil
}

func initLog() *log.Logger {
	var buf bytes.Buffer
	logger := log.New(&buf, "INFO: ", log.LstdFlags)
	logger.SetOutput(os.Stdout)
	return logger
}

func initFlags(paramArgs *inputParams) {
	flag.StringVar(&paramArgs.componentsConfigPath, "config-path", "", "relative path to components yaml from CNAO repo")
	flag.StringVar(&paramArgs.gitToken, "token", "", "git Token")
	flag.StringVar(&paramArgs.baseBranch, "base-branch", "master", "the branch CNAO is running the bumper script on, and on which the PRs will be opened")
	flag.Parse()
	if paramArgs.componentsConfigPath == "" {
		exitWithError(fmt.Errorf("config-path mandatory input paramter not entered. Use --help for usage"))
	}
	if paramArgs.gitToken == "" {
		exitWithError(fmt.Errorf("github token mandatory input paramter not entered. Use --help for usage"))
	}
}

func exitWithError(err error) {
	logger.Fatal(errors.Wrap(err, "Exiting with Error"))
	os.Exit(1)
}
