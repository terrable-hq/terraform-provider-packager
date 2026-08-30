# Changelog

## 30/08/2026

- Replaced external Rolldown execution with esbuild v0.28.2 compiled into the
  provider; bundling no longer needs Node.js, npm or a separately installed tool.
- Preserved custom output directories, deterministic ZIPs and artifact outputs;
  deprecated `rolldown_path` as an ignored compatibility field. Generated code
  and hashes change once when upgrading from the Rolldown-backed release.
- Added empty-PATH bundling and Terraform acceptance coverage, useful build-error
  diagnostics, cancellation checks and a generated-handler execution test.
- Removed provider-development npm dependencies and bundled esbuild's licence
  notice; release checks verify esbuild is embedded in every platform binary.

- Added GoReleaser v2.18.0 release packaging, signed draft-release automation,
  protocol-v6 Registry metadata, and cross-platform snapshot checks in CI.
- Ensured CI installs the pinned GoReleaser binary for the shared local and
  hosted snapshot command.
- Updated Go module imports and project links for the renamed
  `terraform-provider-packager` repository; the Terraform source address
  remains `terrable-hq/packager`.

- Added the initial Terraform provider and `packager_bundle` data source.
- Added Rolldown-backed TypeScript and JavaScript bundling for AWS Lambda.
- Added deterministic ZIP artifacts under `.terrable/build` by default, with
  support for a caller-selected output directory.
- Added artifact path, deployment hash, and size outputs for Lambda resources.
- Added unit coverage for bundle validation, failed-build cleanup, Rolldown
  executable resolution, and the Terraform data-source schema.
- Added GitHub Actions checks for formatting, race-tested unit tests, vetting,
  provider compilation, and a real Rolldown integration build.
- Added Terraform Plugin Testing acceptance coverage over protocol v6 for
  default and custom output directories, artifact contents, hashes, sizes, and
  no-drift planning.
