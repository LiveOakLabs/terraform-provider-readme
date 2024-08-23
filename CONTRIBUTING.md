# Contributing

Live Oak Bank welcomes your interest in contributing to this project in any way
you find meaningful, be it through code contributions, documentation, or bug
reporting. We greatly value and appreciate your involvement.

## Workflow

1. Fork and branch from `main`.
2. Make changes.
3. Run `make lint test` before committing.
4. Commit, push, and open a pull request into `main`.
5. Ensure GitHub workflow checks pass.

## Common Development Tasks

Refer to the [`Makefile`](Makefile) for helpful development tasks.

* `make` - Show list of available tasks.
* `make lint` - Run linters and formatters.
* `make test` - Run all tests.
* `make test-coverage` - Run tests, generate coverage report, and open in a browser.
* `make mocks` - Generate mocks for interfaces (used by external tests).

## GitHub Labels

Issues and pull requests are labeled to help organize and version changes. Feel
free to apply labels to your contributions, or project maintainers will do so.

* Use `patch`, `minor`, or `major` to indicate the [semantic version](https://semver.org/) for a
  change. If unsure, a project maintainer will set it.
* Use `feature` or `enhancement` for added features.
* Use `fix`, `bugfix` or `bug` for fixed bugs.
* Use `chore`, `ci`, and `docs` for maintenance tasks.

## For Project Maintainers

### Releases

This project uses the [Release Drafter](https://github.com/marketplace/actions/release-drafter)
action for managing releases and tags.

The [Changelog Updater](https://github.com/marketplace/actions/changelog-updater) action updates the
[`CHANGELOG.md`](https://github.com/marketplace/actions/changelog-updater) file when releases are
published.

Project maintainers can publish a new release by editing the queued draft
release, making adjustments to the release notes, and publishing.
