# Releasing

## Steps to release a new version of the package

### PR Merge

- Ensure that the PR is merged into the `main` branch.
- The PR should have been reviewed and approved by at least one other contributor.

### Run the release Github Action

- Navigate to the "Actions" tab in the GitHub repository.
- Select the "Release" workflow.
- Click on "Run workflow" and specify the version number for the release.

> [!IMPORTANT]
> This workflow is configured to be run only by maintainers, so it needs to be executed by a maintainer.

### Check the release

- After the workflow completes, check the "Releases" tab in the GitHub repository to ensure that the new version is listed.
- Verify that the release notes and assets are correctly generated.

### Update the custom krew index

- Navigate to the "Pull requests" tab in the GitHub repository.
- Look for a PR titled "docs(krew): Update Krew manifest to vX.Y.Z [skip ci]" (where X.Y.Z is the new version).
- Review the PR and ensure it contains the updated Krew manifest.
- Merge the PR to update the Krew index.

### Check the Krew index

- After merging the PR, check the Krew index to ensure that the new version is available.

```bash
❯❯❯ k krew remove sgmap
Uninstalled plugin: sgmap

❯❯❯ k krew install my-plugin/sgmap
Updated the local copy of plugin index.
Updated the local copy of plugin index "my-plugin".
Installing plugin: sgmap
Installed plugin: sgmap
\
 | Use this plugin:
 |      kubectl sgmap
 | Documentation:
 |      https://github.com/naka-gawa/kubectl-sgmap
/

❯❯❯ k sgmap version
kubectl-sgmap version 0.9.4, revision a3c2eee

```
