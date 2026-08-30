# Changelog

## 30/08/2026

- Added GoReleaser v2.18.0 release packaging, signed draft-release automation,
  protocol-v6 Registry metadata, and cross-platform snapshot checks in CI.
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
