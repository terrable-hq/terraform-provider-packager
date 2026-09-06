# Releasing Terrable Packager

## Automatic preview releases

A push to `main` runs the release workflow against that exact commit. It reruns
CI, chooses the next patch version, builds all six platform archives, verifies
them, and publishes a signed GitHub prerelease. Changes limited to README,
changelog, security/release documentation, and Markdown under `docs/` do not
create a release. The comparison is against the latest version tag, so a
coalesced workflow still includes earlier unreleased code changes.

Versions are allocated in one serialized workflow. GitHub may replace older
pending runs with newer pending runs; it does not interrupt a running release.
A run superseded by a newer `main` commit skips publication; the next run
includes those changes. Each published version points to a tested commit from
`main`. Existing tags
and published assets are never replaced.

For an explicit version, run **Actions > Release > Run workflow** on `main`
and set `version`, for example `v0.3.0`. Leave it blank for the next patch.
An explicit new version must exceed existing versions and may release a
commit whose changes are documentation-only. Dispatches on other branches and
runs in fork repositories cannot reach the release jobs. Pushing a tag does
not trigger a release.

## Release credentials and protections

The `publish` environment must permit **only the branch `main`**, not arbitrary
protected branches or tags. Keep the signing secrets `GPG_PRIVATE_KEY` and
`PASSPHRASE` in that environment, not as repository-wide secrets.

Automatic tag creation uses a dedicated GitHub App:

1. Create an App with only repository **Contents: read and write** permission,
   no webhook requirement, and install it only on this repository.
2. Set environment variable `RELEASE_APP_ID` and environment secret
   `RELEASE_APP_PRIVATE_KEY` in `publish`. Do not put the private key in Git.
3. Add this App to the bypass list of **release tags: admin creation** only.
   Keep **release tags: immutable** without an App bypass. Do not give the App
   a bypass for `main` or its required checks/reviews. Do not grant general
   GitHub Actions a tag-protection bypass or substitute an administrator PAT.
4. Keep the Registry's public signing key matched to the environment's key.

The App token is created only in the publishing job, scoped to this repository,
and revoked by the pinned token action after the job. Without the App setup,
automatic publication cannot create protected tags. The existing GPG key is
not a substitute for GitHub authentication.

## Security boundary

PR CI runs on `pull_request` with read-only repository permissions. Fork PRs
receive no publishing secrets. There is no `pull_request_target` or privileged
`workflow_run` that checks out a contributor branch or consumes its artifacts.
Approving a fork workflow allows tests to run; it does not approve a merge.

Merge protections dismiss stale approvals, require approval of the latest
push, and require CI against an up-to-date branch. CODEOWNERS covers workflows,
release configuration, and release scripts. Do not use the maintainer-review
administrator exemption to merge an unreviewed fork revision. A workflow cannot
determine that deliberately approved malicious code is safe; review of the
actual final diff remains the trust boundary.

Release builds run without signing secrets, write tokens, persisted checkout
credentials, or restored caches. Publishing runs on a separate fresh runner
with no checkout and executes no repository scripts or built binaries. It
consumes only the immutable artifact ID from its own build job, validates the
exact filenames and SHA-256 checksums, then signs the checksums and publishes.
ZIP files are uploaded as data, never extracted or executed by the signer.
All external actions are pinned to commit SHAs.

## Validation and recovery

`make ci` includes version-allocation and hostile-artifact regression tests,
Go race tests, vet, a provider build, Node handler execution, and Terraform
protocol-v6 acceptance tests. `make release-snapshot` verifies the six platform
archives, Registry manifest, embedded esbuild, and third-party licence notice.
GoReleaser v2.18.0 is pinned in CI and the release workflow.

A failed build creates no remote tag. If publishing fails after reserving a
tag, rerun that same workflow/commit: it reuses the tag only if its SHA matches.
A completed release is a no-op on retry. If a draft already exists, the workflow
stops rather than replacing uploaded assets. A maintainer must inspect and
complete that draft, or delete only the unpublished draft before retrying.
Never move a version tag or alter a published release; make a new version.

The publisher first uploads a draft, then marks it as a published prerelease
only after uploads succeed. The Terraform provider address is
`terrable-hq/packager`. See the
[provider publishing guide](https://developer.hashicorp.com/terraform/registry/providers/publishing)
for Registry signing-key and repository setup.
