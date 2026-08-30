# Changelog

## 30/08/2026

- Added the initial Terraform provider and `packager_bundle` data source.
- Added Rolldown-backed TypeScript and JavaScript bundling for AWS Lambda.
- Added deterministic ZIP artifacts under `.terrable/build` by default, with
  support for a caller-selected output directory.
- Added artifact path, deployment hash, and size outputs for Lambda resources.
