---
page_title: "Terrable Packager Provider"
description: |-
  Package JavaScript and TypeScript Lambda handlers with embedded esbuild, keeping build artifacts separate from source files.
---

# Terrable Packager Provider

The Terrable Packager provider builds JavaScript and TypeScript Lambda handlers
with embedded esbuild and produces deterministic ZIP artifacts. Build output defaults
to `.terrable/build`, which can be ignored with one `.terrable/` entry in your
project's `.gitignore`.

## Requirements

The bundler is compiled into the provider: no external bundler or Node.js is
required for packaging. Install your handler's imported application packages
before running Terraform (typically with `npm ci`). Package installation,
JavaScript bundler plugins/configuration and native-addon packaging are not
performed by this provider. Lambda still needs its Node.js runtime.

When upgrading from v0.1.0, remove `rolldown_path`; it is deprecated and ignored.
The engine change updates generated JavaScript and ZIP hashes, while preserving
the input/output interface and deterministic packaging for unchanged inputs.

## Example usage

```terraform
terraform {
  required_providers {
    packager = {
      source = "terrable-hq/packager"
    }
  }
}

data "packager_bundle" "handler" {
  name              = "handler"
  entrypoint        = "src/handler.ts"
  working_directory = path.root
  output_directory  = ".terrable/build"
}
```

Use `artifact_path` as a Lambda resource's `filename`, and `base64sha256` as its
`source_code_hash`. The ZIP contains a CommonJS `index.js` file; a handler
exported as `handler` uses the Lambda handler setting `index.handler`.

Packaging runs when Terraform reads the data source, normally during planning.
Preserve the output directory if plan and apply run on separate machines.

The provider has no provider-level configuration arguments.
