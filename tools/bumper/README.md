# Bumper Script

The Bumper script goes over all of CNAO's components, using the components.yaml config, finds new releases and bumps them in separate PRs. The script can be run locally or via an automation such as GitHub Actions.

## Running the script manually

In order to run the script manually, you need to have a github token. To create a token in your github user, follow this [guide](https://docs.github.com/en/free-pro-team@latest/github/authenticating-to-github/creating-a-personal-access-token).

## How to run the script

```
make ARGS="-config-path=<path-to-components.yaml-relative-to-bumping-repo> -token=<git-token>" auto-bumper
```

Where:
* config-path: relative path to components.yaml from the bumping repo. In its current position we'll simply: config-path="components.yaml"
* token: personal/gitActions github-token.
