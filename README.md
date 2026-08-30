# Terrable Packager

[![CI](https://github.com/terrable-hq/packager/actions/workflows/ci.yml/badge.svg)](https://github.com/terrable-hq/packager/actions/workflows/ci.yml)

Terrable Packager is an experimental Terraform provider for producing AWS
Lambda deployment artifacts from JavaScript and TypeScript entrypoints.

The first slice uses [Rolldown](https://rolldown.rs/) to create a CommonJS
bundle for Node.js and places it in a deterministic ZIP archive. Generated
artifacts are written to `.terrable/build` by default, keeping build output
away from handler source files and making the whole directory safe to ignore.

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

## Rolldown requirement

The provider currently resolves Rolldown in this order:

1. The explicit `rolldown_path` value.
2. `node_modules/.bin/rolldown` beneath `working_directory`.
3. A `rolldown` executable on `PATH`.

Install a project-local copy with:

```shell
npm install --save-dev rolldown
```

Packaging Rolldown with provider release archives is intentionally left for a
separate slice. Until that is delivered, consumers need Node.js and Rolldown
available where Terraform runs.

The data source writes its artifact when Terraform reads it, normally during
planning. If plan and apply run on different machines, preserve
`.terrable/build` between those stages or rebuild the plan on the apply runner.

## Development

Requirements:

- Go 1.25 or later.
- Node.js and npm.

Install the pinned integration-test dependency and run the suite:

```shell
npm install
make test
make test-integration
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
`go vet`, compiles the provider, and exercises the real Rolldown fixture.

The provider is not published yet. For local Terraform development, build the
binary and configure a Terraform CLI `dev_overrides` entry for
`terrable-hq/packager`.
