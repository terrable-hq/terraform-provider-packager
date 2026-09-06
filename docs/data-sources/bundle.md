---
page_title: "packager_bundle Data Source - Terrable Packager"
subcategory: ""
description: |-
  Bundles a JavaScript or TypeScript entrypoint into a deterministic AWS Lambda ZIP artifact.
---

# packager_bundle (Data Source)

Bundles one JavaScript or TypeScript entrypoint with embedded esbuild and
creates a deterministic AWS Lambda ZIP artifact containing CommonJS `index.js`.
No external bundler or Node.js installation is required for bundling. Imported
application dependencies must already be installed. Node built-ins remain
external; native addons and extra asset files are not automatically packaged.
Builds producing additional files (such as CSS) fail instead of omitting them.

Packaging happens when Terraform reads the data source, normally during
planning. CI systems that separate plan and apply should preserve the
configured output directory between runners.

## Example usage

```terraform
data "packager_bundle" "handler" {
  name              = "handler"
  entrypoint        = "src/handler.ts"
  working_directory = path.root
  output_directory  = ".terrable/build"
}
```

## Arguments

- `name` (required): artifact filename without the `.zip` extension.
- `entrypoint` (required): JavaScript or TypeScript entrypoint.
- `working_directory` (optional): base directory for relative paths. Defaults
  to the Terraform process working directory.
- `output_directory` (optional): artifact directory. Defaults to
  `.terrable/build` beneath `working_directory`.
- `rolldown_path` (optional, deprecated): ignored compatibility field for
  v0.1.0 configurations. Remove it; esbuild is compiled into the provider.

## Attributes

- `artifact_path`: absolute path to the generated ZIP artifact.
- `base64sha256`: base64-encoded SHA-256 artifact hash.
- `size`: artifact size in bytes.
