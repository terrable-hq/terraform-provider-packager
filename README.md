# Terrable Packager

[![CI](https://github.com/terrable-hq/terraform-provider-packager/actions/workflows/ci.yml/badge.svg)](https://github.com/terrable-hq/terraform-provider-packager/actions/workflows/ci.yml)

Terrable Packager is an experimental Terraform provider for producing AWS
Lambda deployment artifacts from JavaScript and TypeScript entrypoints.

The [v0.2.0 prerelease](https://github.com/terrable-hq/terraform-provider-packager/releases/tag/v0.2.0)
embeds [esbuild](https://esbuild.github.io/) through its native Go API to create
a CommonJS bundle for Node.js and places it in a deterministic ZIP archive.
No external bundler or Node.js installation is needed for bundling. Generated
artifacts are written to `.terrable/build` by default, keeping build output
away from handler source files and making the whole directory safe to ignore.

**Release and branch status:** the esbuild implementation is present in the
`v0.2.0` tag, but has not yet been integrated into `main`. The `main` branch
still builds the older Rolldown implementation. Use the tagged release for
embedded esbuild; see the [v0.2.0 development instructions](https://github.com/terrable-hq/terraform-provider-packager/blob/v0.2.0/README.md#development)
when building that version from source.

## Current contract

```hcl
terraform {
  required_providers {
    packager = {
      source = "terrable-hq/packager"
    }
  }
}

data "packager_bundle" "hello" {
  name              = "hello"
  entrypoint        = "src/hello.ts"
  working_directory = path.root
}

resource "aws_lambda_function" "hello" {
  function_name    = "hello"
  filename         = data.packager_bundle.hello.artifact_path
  source_code_hash = data.packager_bundle.hello.base64sha256

  # Remaining Lambda configuration omitted.
}
```

This writes:

```text
.terrable/build/hello.zip
```

Add this to the consuming project's `.gitignore`:

```gitignore
.terrable/
```

### Choose another output directory

`output_directory` may be absolute or relative to `working_directory`:

```hcl
data "packager_bundle" "hello" {
  name              = "hello"
  entrypoint        = "src/hello.ts"
  working_directory = path.root
  output_directory  = "build/lambda"
}
```

The data source exposes:

- `artifact_path`: absolute path to the generated ZIP.
- `base64sha256`: hash suitable for
  `aws_lambda_function.source_code_hash`.
- `size`: artifact size in bytes.

## Bundling with v0.2.0

The bundler is compiled into the provider. The provider does not download or
run a bundler executable, install npm packages, or execute JavaScript plugin
configuration.

Install your handler's application dependencies before running Terraform,
typically with your application's `npm ci`. Node.js is needed to install and
run/test those dependencies, but not for the provider's bundling step.
Node built-ins stay external for the Lambda Node.js runtime. The ZIP contains
one CommonJS `index.js` file (`index.handler` for a `handler` export).
TypeScript is transpiled, not type-checked. Native addons and extra assets are
not automatically packaged; builds that emit multiple files, such as CSS,
fail rather than silently dropping files.

When migrating from v0.1.0, remove `rolldown_path`: v0.2.0 retains it only as
a deprecated, ignored compatibility field. Remove a Rolldown dev dependency
if your application does not otherwise use it. Input paths, output attributes
and ZIP layout are unchanged, but esbuild produces different JavaScript and
artifact hashes, so expect a one-time Lambda code update.

## Rolldown requirement for main and v0.1.0

The older implementation on `main` and in v0.1.0 resolves Rolldown in this order:

1. The explicit `rolldown_path` value.
2. `node_modules/.bin/rolldown` beneath `working_directory`.
3. A `rolldown` executable on `PATH`.

Install a project-local copy with:

```shell
npm install --save-dev rolldown
```

This requirement does not apply to v0.2.0, which replaces Rolldown with
embedded esbuild.

## Artifact lifecycle

The data source writes its artifact when Terraform reads it, normally during
planning. If plan and apply run on different machines, preserve
`.terrable/build` between those stages or rebuild the plan on the apply runner.

## Development

These commands describe the current `main` branch, which still uses Rolldown.
For the esbuild implementation, check out `v0.2.0` and follow its linked
development instructions above; that version does not need npm dependencies
for provider development.

Requirements:

- Go 1.25.8 or later.
- Node.js and npm.
- Terraform CLI for acceptance tests.

Install the pinned integration-test dependency and run the suite:

```shell
npm install
make test
make test-integration
make test-acceptance
```

`make test-integration` runs a real Rolldown build of the TypeScript fixture.
The ordinary test suite uses an in-process fake runner and does not require
Rolldown.

Before opening a pull request, run the same aggregate gate used by GitHub
Actions:

```shell
npm ci
make ci
```

This checks formatting, runs the Go suite with the race detector, runs
`go vet`, compiles the provider, exercises the real Rolldown fixture, and runs
Terraform Plugin Testing acceptance cases over protocol v6. The acceptance
suite verifies both default and custom output directories, independently
inspects each generated ZIP, checks its hash and size, and requires a repeated
plan to be empty.

For local Terraform development, build the binary and configure a Terraform
CLI `dev_overrides` entry for `terrable-hq/packager`.

## Releases

Published GitHub prereleases are available for
[v0.2.0 (embedded esbuild)](https://github.com/terrable-hq/terraform-provider-packager/releases/tag/v0.2.0)
and [v0.1.0 (external Rolldown)](https://github.com/terrable-hq/terraform-provider-packager/releases/tag/v0.1.0).

Tagged releases use GoReleaser v2.18.0 to build Linux, macOS, and Windows
archives for AMD64 and ARM64, sign their checksums, and create a draft GitHub
release. CI also builds unsigned snapshots without access to release secrets.
See [RELEASING.md](RELEASING.md) for signing-key setup, checks, and the first
Registry publication.
