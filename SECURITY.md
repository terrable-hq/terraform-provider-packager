# Security

## Reporting a vulnerability

Please report suspected vulnerabilities through
[GitHub private vulnerability reporting](https://github.com/terrable-hq/terraform-provider-packager/security/advisories/new).
Avoid public issues for vulnerabilities or credentials. Include the affected
version, impact, and a minimal reproduction without real secrets.

## Automated checks

Dependabot checks GitHub Actions and Go modules
weekly. Minor and patch Go updates are grouped; major updates remain separate.
Dependabot security updates are enabled independently of that weekly schedule.
Dependency updates go through pull requests and the normal CI requirements;
they are not automatically merged.

CodeQL scans Go and GitHub Actions with the extended query suite on pull
requests (including forks), pushes to `main`, and a weekly schedule. Maintainer
approval is required before external contributors' workflows run. Dependency
review rejects PRs that introduce dependencies with known high or critical
vulnerabilities. Repository settings also enable Dependabot alerts, secret
scanning, and push protection for supported secret patterns.

The Dependabot update schedule is owned by `.github/dependabot.yml`.
`.github/workflows/codeql.yml` owns code scanning and
`.github/workflows/dependency-review.yml` owns dependency review. Keep CodeQL
default setup disabled so it does not reject results from the explicit workflow.
Secret protection, private reporting, and merge rules are managed in GitHub
repository settings. Review changes to CI and security configuration carefully.

## Automated release trust boundary

Release jobs run only for this repository's `main` branch. The `publish`
environment restricts signing and App credentials to that branch. Builds have
no release credentials; a fresh publishing job validates artifacts from its
own build and does not check out or run repository code. PR and fork artifacts
are never promoted to releases.

New commits dismiss PR approvals, and the latest push requires review.
Workflow and release-code changes also require CODEOWNERS review. Inspect the
final revision before merging; workflow-run approval is not merge approval.
See [RELEASING.md](RELEASING.md) for the App's limited tag-creation authority
and the retained administrator review exemption.
