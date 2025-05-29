# Releasing

## Steps to release a new version of the package

1. PR Merge

- Ensure that the PR is merged into the `main` branch.
- The PR should have been reviewed and approved by at least one other contributor.

1. Run the release Github Action

- Navigate to the "Actions" tab in the GitHub repository.
- Select the "Release" workflow.
- Click on "Run workflow" and specify the version number for the release.

> [!IMPORTANT]
> This workflow is configured to be run only by maintainers, so it needs to be executed by a maintainer.

1. Check the release

- After the workflow completes, check the "Releases" tab in the GitHub repository to ensure that the new version is listed.
- Verify that the release notes and assets are correctly generated.

1. Update the custom krew index

- Navigate to the "Pull requests" tab in the GitHub repository.
- Look for a PR titled "docs(krew): Update Krew manifest to vX.Y.Z [skip ci]" (where X.Y.Z is the new version).
- Review the PR and ensure it contains the updated Krew manifest.
- Merge the PR to update the Krew index.
