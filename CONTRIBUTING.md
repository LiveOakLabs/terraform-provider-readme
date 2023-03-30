# Contributing

Merge requests should be opened to merge into the `main` branch.

## Workflow

1. Fork and branch from `main`.
2. Make changes.
3. Run `pre-commit install` or `pre-commit run -a` before committing.
4. Run `make test` before committing.
5. Commit, push, and open a pull request into `main`.
6. Ensure GitLab workflow checks pass.
7. Set a GitLab label if you know what fits best.

## GitHub Labels

* Use `patch`, `minor`, or `major` to indicate the [semantic version](https://semver.org/) for a
  change. If unsure, a project maintainer will set it.
* Use `feature` or `enhancement` for added features.
* Use `fix`, `bugfix` or `bug` for fixed bugs.
* Use `chore`, `ci`, and `docs` for maintenance tasks.

## Releases

This project uses the [Release Drafter](https://github.com/marketplace/actions/release-drafter)
action for managing releases and tags.

The [Changelog Updater](https://github.com/marketplace/actions/changelog-updater) action updates the
[`CHANGELOG.md`](https://github.com/marketplace/actions/changelog-updater) file when releases are
published.

## Tools and Tests

This project uses a few [tools](readme/tools.go) for validating code quality and functionality:

* [pre-commit](https://pre-commit.com/) for ensuring consistency and code quality before committing (external dependency).
* [golangci-lint](https://golangci-lint.run/) for linting and formatting.
* [gofumpt](https://github.com/mvdan/gofumpt) (is included with golangci-lint).
* [gocover-cobertura](https://github.com/boumenot/gocover-cobertura) for code test coverage reporting.
* [govulncheck](https://github.com/golang/vuln) for detecting vulnerabilities in Go packages.

Refer to the [`Makefile`](Makefile) for helpful development tasks.
