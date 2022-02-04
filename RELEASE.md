# Release process

The release process for CNAO consist of a pair of manual steps:
1. Create release PR with github action [Create Version PR](https://github.com/kubevirt/cluster-network-addons-operator/actions/workflows/prepare-version.yaml),
   it has a pair of fields: versionLevel and baseBranch.
   versionLevel is the semVer version part to bump to and
   baseBranch is the branch to release from.
   To run it:
   - From the gh UI passing the `versionLevel` or `baseBranch`
   - With the gh cli like `gh workflow run prepare-version.yaml -s versionLevel=... -s baseBranch=...`
2. Merge the version PR when continous integration is fine.

After the version PR is merged a prow job will [release-cluster-network-addons-operator](https://github.com/kubevirt/project-infra/blob/main/github/ci/prow-deploy/files/jobs/kubevirt/cluster-network-addons-operator/cluster-network-addons-operator-postsubmits.yaml#L2-L28) call the `make release` target that to do the following:

1. Tag the code the version from the version PR.
2. Generate release notes from the PRs release-notes field at description
   and the [template](hack/release-notes.tmpl)
3. Push containers to quay.io.
4. Create a new github release and upload the artifacts.


## Release notes handling

The release notes are generated using the kubernetes [release-notes](https://github.com/kubernetes/release/blob/master/cmd/release-notes/README.md).
The tool scan all the PRs between the version to release and the previous one
and look for the following block on the PR description:

````
```release-note
Set default placement to os/linux nodes.
```
````

To categorize the change use the prow command [/kind](https://prow.k8s.io/command-help) with the
category you want `/kind bug`, `/kind enhancement`, etc... the kinds understood
by the tool are listed [here](https://github.com/kubernetes/release/blob/master/pkg/notes/notes.go#L67-L78).

Sometime is needed to add some notes related actions to be done before or after
an upgrade for that adding the "action requiered" string inside the
`relase-note` block will add that line to the Urgent Upgrade Notes section on
the release, like in the example

````
```release-note
Set default placement to os/linux nodes.
[action required]: Be sure that the nodes are linux
```
````


More information at [k8s release notes workflow](https://github.com/kubernetes/release/blob/d41c86508aac7f0b6d5f701fb2f6d3ae29bf35e0/docs/release-notes.md#kubernetes-release-notes-workflow).
