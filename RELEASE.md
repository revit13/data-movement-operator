# Release Process

The process of creating a release is described in this document. Replace `X.Y.Z` with the version to be released.

## 1. Create a `releases/X.Y.Z` branch 

The `releases/X.Y.Z` branch should be created from a base branch. 

For major and minor releases the base is `master` and for patch releases (fixes) the base is `releases/X.Y.(Z-1)`.
A new patch release should be created before merging pull-requests as described in the next section.

You can do that [directly from GitHub](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/proposing-changes-to-your-work-with-pull-requests/creating-and-deleting-branches-within-your-repository#creating-a-branch).

## 2. Create a [new release](https://github.com/fybrik/hello-world-module/releases/new) 

Use `vX.Y.Z` tag and set `releases/X.Y.Z` as the target.

Ensure that the release notes explicitly mention upgrade instructions and any breaking change.
