# Releasing Terrable Packager

Provider releases are built with GoReleaser **v2.18.0**, the latest stable
release checked on 30 August 2026. Both CI and the release workflow pin that
version. GitHub Actions are pinned to full commit SHAs; Dependabot checks those
action pins weekly. Update the GoReleaser version in both workflows together
when upgrading.

## One-time setup

1. Make `terrable-hq/terraform-provider-packager` public before Registry
   publication. Select an appropriate open-source licence before making the
   source public; this repository does not currently declare one.
2. Create a GitHub environment named `publish`. Configure required reviewers
   and restrict who can create tags matching `v*`. Referencing the environment
   in the workflow does not create its approval rules automatically.
3. Put the ASCII-armored RSA release-signing private key in the environment
   secret `GPG_PRIVATE_KEY` and its passphrase in `PASSPHRASE`. Do not commit
   private keys or passphrases. Keep a secure backup outside GitHub.
4. Add the corresponding ASCII-armored public key to the Terraform Registry's
   signing keys for the `terrable-hq` namespace.

Only the signing job uses the `publish` environment. It starts after both the
full CI/acceptance gate and the unsigned cross-platform snapshot job pass.
Before approving it, verify that the tag points at the reviewed commit and
that its release workflow and GoReleaser configuration are expected.

## Local checks

Install GoReleaser v2.18.0, Go, Node.js, npm, and Terraform, then run:

```shell
npm ci
make ci
make release-check
make release-snapshot
```

`release-snapshot` builds all six platform archives and a checksum file in
`dist/`, then checks their names, embedded provider binaries, checksums, and
protocol manifest. The manifest is hashed from the source file and receives
its versioned filename during upload, so it is not copied into `dist/` by the
snapshot. This target skips signing and publication, requires no release
secrets, and does not create a Git tag or GitHub release. `dist/` is ignored by
Git.

## Create a release

1. Review the changelog, confirm CI is green, and ensure the repository is
   clean and up to date on `main`.
2. Create and push a semantic-version tag, for example:

   ```shell
   git tag -a v0.1.0 -m "Release v0.1.0"
   git push origin v0.1.0
   ```

3. The `Release` workflow reruns CI and the snapshot build, waits for any
   configured `publish` environment approval, then builds and signs a draft
   GitHub release.
4. Inspect the draft assets. They should include:
   - ZIP archives for Linux, macOS, and Windows, each on AMD64 and ARM64.
   - `terraform-provider-packager_<version>_manifest.json` with protocol `6.0`.
   - `terraform-provider-packager_<version>_SHA256SUMS`, covering every ZIP and
     the manifest.
   - `terraform-provider-packager_<version>_SHA256SUMS.sig`, a binary detached
     GPG signature of the checksum file.
5. Download the assets, verify the GPG signature and checksums, add user-facing
   release notes, and publish the draft. Do not replace assets from an already
   published version; release a new version instead.

The first publication also requires selecting this repository through
**Publish > Provider** in the Terraform Registry. Subsequent finalized GitHub
releases are ingested through the Registry webhook. The Terraform provider
source address remains `terrable-hq/packager` despite the GitHub repository's
`terraform-provider-` prefix.

Node.js and Rolldown remain consumer runtime dependencies. These release
archives package the Go provider, not those external tools.

See [HashiCorp's provider publishing guide](https://developer.hashicorp.com/terraform/registry/providers/publishing)
for signing-key and Registry setup details.
