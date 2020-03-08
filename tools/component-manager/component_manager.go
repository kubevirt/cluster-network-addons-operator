package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/storage/memory"
	"gopkg.in/yaml.v2"
)

type componentsConfig struct {
	Components map[string]componentConfig `json:"yaml"`
}

type componentConfig struct {
	Repository   string       `yaml:"url"`
	Branch       string       `yaml:"branch"`
	Commit       string       `yaml:"commit"`
	Description  string       `yaml:"description"`
	UpdatePolicy updatePolicy `yaml:"updatePolicy"`
}

type updatePolicy string

const (
	updatePolicyTagged = "tagged"
)

type componentRevision struct {
	Title               string
	Message             string
	RevisionID          string
	ComponentCommitHash string
	ComponentTag        string
}

func buildRevisionID(cnaoBranch string, componentName string, componentCommitHash string) string {
	return ""
}

func componentsConfigFromFile(filePath string) (componentsConfig, error) {
	var config componentsConfig

	configFile, err := ioutil.ReadFile(filePath)
	if err != nil {
		return config, err
	}

	err = yaml.Unmarshal(configFile, &config)
	if err != nil {
		return config, err
	}

	return config, nil
}

func getRepositoryOnBranch(repositoryURL string, branchName string) (*git.Repository, error) {
	return git.Clone(memory.NewStorage(), nil, &git.CloneOptions{
		URL:           repositoryURL,
		ReferenceName: plumbing.NewBranchReferenceName(branchName),
	})
}

func getTagsSinceCommit(repository *git.Repository, commitHash plumbing.Hash) ([]plumbing.Reference, error) {
	tags := []plumbing.Reference{}

	tagReferences, err := repository.Tags()
	if err != nil {
		return tags, err
	}

	tagPerCommitHash := map[string]*plumbing.Reference{}

	err = tagReferences.ForEach(func(tagReference *plumbing.Reference) error {
		tagPerCommitHash[tagReference.Hash().String()] = tagReference
		return nil
	})
	if err != nil {
		return tags, err
	}

	commitIter, err := repository.Log(&git.LogOptions{})
	if err != nil {
		return tags, err
	}

	commitIter.ForEach(func(c *object.Commit) error {
		if c.Hash.String() == commitHash.String() {
			return fmt.Errorf("stop")
		}
		if tagReference, ok := tagPerCommitHash[c.Hash.String()]; ok {
			tags = append(tags, *tagReference)
		}
		return nil
	})

	return tags, nil
}

func main() {
	githubToken, ok := os.LookupEnv("GITHUB_TOKEN")
	if !ok {
		log.Fatal("Environment variable GITHUB_TOKEN has to be set")
	}

	cnaoOrg := flag.String("cnao-org", "kubevirt", "GitHub organization of CNAO (optional)")
	cnaoRepo := flag.String("cnao-repo", "cluster-network-addons-operator", "GitHub repo of CNAO (optional)")
	componentName := flag.String("component", "", "Name of the component to be revised")
	componentsConfigPath := flag.String("config", "", "Path to components.yaml file")

	flag.Parse()

	if *componentName == "" || *componentsConfigPath == "" {
		flag.PrintDefaults()
		log.Fatal("Required attributes were not provided")
	}

	config, err := componentsConfigFromFile(*componentsConfigPath)
	if err != nil {
		log.Fatalf("Failed to read the config file: %v", err)
	}

	component := config.Components[*componentName]
	log.Printf("Component: %+v", component)

	log.Print("Cloning component's repo")
	r, err := getRepositoryOnBranch(component.Repository, component.Branch)
	if err != nil {
		log.Fatalf("Failed to read component's repository: %v", err)
	}

	revisionsByID := map[string]componentRevision{}

	switch component.UpdatePolicy {
	case updatePolicyTagged:
		log.Print("Looking for new tags")
		tags, err := getTagsSinceCommit(r, plumbing.NewHash(component.Commit))
		if err != nil {
			log.Fatalf("Failed to list tags added after the given commit: %v", err)
		}

		for _, tag := range tags {
			fmt.Printf("%v\n", tag)
		}

		// TODO: translate it to a list of revisions
	default:
		log.Fatalf("Component's update policy %q was not recognized", component.UpdatePolicy)
	}

	log.Print("Connecting to GitHub API")
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: githubToken},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	// TODO: get list of already posted revisions
	log.Print("Listing already posted revisions")

	postedRevisionIDs := map[string]bool{}

	prs, _, err := client.PullRequests.List(ctx, *cnaoOrg, *cnaoRepo, &github.PullRequestListOptions{Base: "master", State: "all"})
	if err != nil {
		log.Printf("Failed to list existing PRs: %v", err)
	}

	for _, pr := range prs {
		fmt.Printf("PR title: %v\n", *pr.Title)
		fmt.Printf("PR description: %v\n", *pr.Body)
		revisions := []componentRevision{}
		break
	}

	newRevisions := []componentRevision{}
	for revisionID, revision := range revisionsByID {
		if _, alreadyPosted := postedRevisionIDs[revisionID]; !alreadyPosted {
			newRevisions := append(newRevisions, revision)
		}
	}

	for _, newRevision := range newRevisions {
		fmt.Printf("Revision: %v", revision)
		// TODO: tag current state of the branch, defer to it
		// TODO: edit components.yaml
		// TODO: call make components
		// TODO: push new changes
		// TODO: roll back to the tag
	}
}
