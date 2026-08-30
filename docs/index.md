---
page_title: "Terrable Packager Provider"
description: |-
  Package JavaScript and TypeScript Lambda handlers with Rolldown, keeping build artifacts separate from source files.
---

# Terrable Packager Provider

The Terrable Packager provider builds JavaScript and TypeScript Lambda handlers
with Rolldown and produces deterministic ZIP artifacts. Build output defaults
to `.terrable/build`, which can be ignored with one `.terrable/` entry in your
project's `.gitignore`.

## Requirements

Node.js and Rolldown must be installed on the machine running Terraform. The
provider first uses an explicit `rolldown_path`, then checks
`node_modules/.bin/rolldown` beneath `working_directory`, then checks `PATH`.
Node.js and Rolldown are not bundled with the provider release.

```shell
npm install --save-dev rolldown
```

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
