# Terrable Packager

[![CI](https://github.com/terrable-hq/terraform-provider-packager/actions/workflows/ci.yml/badge.svg)](https://github.com/terrable-hq/terraform-provider-packager/actions/workflows/ci.yml)

Terrable Packager is an experimental Terraform provider for producing AWS
Lambda deployment artifacts from JavaScript and TypeScript entrypoints.

The provider embeds [esbuild](https://esbuild.github.io/) v0.28.2 through its
Go API to create a CommonJS bundle for Node.js and places it in a deterministic
ZIP archive. No external bundler or Node.js installation is needed for bundling.
Generated artifacts are written to `.terrable/build` by default, keeping build output
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

## Self-contained bundling

`terraform init` downloads a provider binary that includes the bundler. There
is no bundler executable lookup, runtime tool download, npm install, or
JavaScript configuration/plugin execution inside the provider.

Your handler's imported packages must already be available, typically through
your application's `npm ci` step. Node.js is needed for that installation and
for running/testing the handler, but not for the provider's bundling step.
Node built-ins stay external for the Lambda Node.js runtime. The current ZIP
contract is one CommonJS `index.js` file (`index.handler` for a `handler` export).
Native addons and extra assets are not automatically packaged; builds that
emit multiple files, such as CSS, fail rather than silently dropping files.
TypeScript is transpiled, not type-checked.

### Migrating from v0.1.0

Remove `rolldown_path` from configuration. It is retained as a deprecated,
ignored compatibility field; it no longer runs an executable. Remove a
Rolldown dev dependency only if your application doesn't otherwise use it.

Input paths, `output_directory`, ZIP layout and output attributes are unchanged.
The switch to esbuild changes generated JavaScript and therefore artifact
hashes; expect a one-time Lambda code update. Repeated builds with the same
inputs and provider version remain deterministic.

The data source writes its artifact when Terraform reads it, normally during
planning. If plan and apply run on different machines, preserve
`.terrable/build` between those stages or rebuild the plan on the apply runner.

## Development

Requirements:

- Go 1.25.8 or later.
- Node.js only for the generated-handler runtime smoke test.
- Terraform CLI for acceptance tests.

Run the suite (no npm dependencies are needed for provider development):

```shell
make test
make test-integration
make test-acceptance
```

`make test` exercises the real embedded bundler with an empty `PATH`, including
TypeScript, local imports, installed-package imports, errors and reproducibility.
`make test-integration` builds the TypeScript fixture with an empty `PATH`, then
uses an explicitly located Node.js executable to invoke the generated handler
and assert its response. Node is a test oracle, not a bundling dependency.

Before opening a pull request, run the same aggregate gate used by GitHub
Actions:

```shell
make ci
```

This checks formatting, runs the Go suite with the race detector, runs
`go vet`, compiles the provider, executes the generated handler, and runs
Terraform Plugin Testing acceptance cases over protocol v6. The acceptance
suite verifies both default and custom output directories, independently
inspects each generated ZIP, checks its hash and size, and requires a repeated
plan to be empty. Acceptance cases run without external build tools on `PATH`
and also check that legacy `rolldown_path` configuration is ignored.

For local Terraform development, build the binary and configure a Terraform
CLI `dev_overrides` entry for `terrable-hq/packager`.

## Releases

Tagged releases use GoReleaser v2.18.0 to build Linux, macOS, and Windows
archives for AMD64 and ARM64, sign their checksums, and create a draft GitHub
release. CI also builds unsigned snapshots without access to release secrets.
See [RELEASING.md](RELEASING.md) for signing-key setup, checks, and the first
Registry publication.
