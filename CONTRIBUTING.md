# Contributing

Thank you for considering a contribution to `aads`. Issues and pull requests
are welcome.

This project is a command-line interface for the Apple Ads Campaign Management
API. Changes should preserve the CLI's safety rules, scriptability, and
documented behavior.

## Before You Start

For non-trivial changes, read these documents first:

- [`docs/requirements.md`](docs/requirements.md): CLI behavior, user-facing
  conventions, safety rules, and testing expectations.
- [`docs/architecture.md`](docs/architecture.md): package boundaries, code
  organization, generation pipeline, and implementation strategy.
- [`docs/coverage.md`](docs/coverage.md): endpoint coverage and known Apple Ads
  API path discrepancies.

Do not hand-edit generated files unless the change specifically requires it.
For example, regenerate `docs/commands.md` with the documented command instead
of editing it directly.

## Development Environment

You need:

- Go 1.24 or newer.
- `make`.
- Python 3 for documentation and Apple Ads OpenAPI generation scripts.
- Git.

Optional:

- `golangci-lint`. If it is not installed, `make lint` falls back to `go vet`.

Build the CLI with:

```sh
make build
```

The local binary is written to `bin/aads`.

Install the tracked git hooks with:

```sh
make install-git-hooks
```

This sets `core.hooksPath` to `.githooks` for the local clone. The pre-commit
hook runs `make format-check`, `make lint`, and `make test`, and blocks the
commit if any of them fail.

## Tests And Checks

Run the most specific tests for the code you changed. Useful targets include:

```sh
make test
make test-integration
make format-check
make lint
make commands-doc-check
make openapi-check
```

Use `make test-integration` for mocked command-integration coverage in `./cmd`.
Use `make commands-doc-check` when command help or CLI behavior changes. Use
`make openapi-check` when Apple Ads documentation artifacts, schema generation,
or endpoint coverage are affected.

Before relying on command behavior, check the current help output:

```sh
bin/aads --help
bin/aads <group> --help
bin/aads <group> <command> --help
```

## Live Tests

Live tests run against the real Apple Ads API:

```sh
make test-live
```

This target sets `AADS_INTEGRATION_TEST=1` and runs the live command-integration
tests in `./cmd`. Live tests require Apple Ads credentials and access to a test
organization.

Run live tests deliberately. They may create, update, and delete test resources.
Use an isolated configuration directory with `AADS_CONFIG_DIR` or `--config-dir`;
do not rely on, inspect, or expose real configuration or token cache contents
from `~/.aads/`.

Live tests are not expected for every pull request. They are most relevant for
changes to authentication, request routing, request payloads, mutation commands,
or behavior that cannot be validated with mocked tests.

## Before Opening A Pull Request

Before opening a PR:

- Keep the change focused.
- Run targeted tests for the code you changed.
- Run `make test` for broad or shared changes.
- Run `make format-check` and `make lint`.
- Update `docs/requirements.md` when user-facing CLI behavior changes.
- Update `docs/architecture.md` when package boundaries, generation flow, or
  implementation strategy changes.
- Update `docs/coverage.md` when endpoint coverage or API request paths change.
- Regenerate generated documentation or Apple Ads artifacts when relevant.
- Check command help output when changing CLI routing, flags, or help text.
- Do not include credentials, tokens, private keys, customer data, or real
  account identifiers.

If a check was not run, mention that in the PR and explain why.

## Pull Request Guidance

Describe the problem, the approach, and the validation performed. Link related
issues when possible. Call out user-visible behavior changes, breaking changes,
and any Apple Ads API discrepancies found during the work.

Prefer small PRs with one clear purpose. Avoid unrelated refactors, formatting
churn, or generated-file changes that are not needed for the contribution.

If a proposed change was generated mostly with the help of AI agents, consider
opening an issue with the prompts, context, and intended outcome before opening
a large generated PR. For many changes, reviewing the prompt trail and desired
behavior first may be easier than reviewing a large generated diff.

## Issues And Triage Labels

Use GitHub issues for bug reports, enhancement requests, documentation
improvements, and questions.

Issues should use one of these labels:

- `bug`: something is broken or behaves contrary to the documented
  requirements.
- `enhancement`: a new feature, an improvement to existing behavior, or an
  improvement to documentation.
- `question`: a request for clarification, usage help, or design discussion.

The preferred workflow is to use GitHub issue forms so the appropriate label is
applied automatically. If labels are unavailable, include the category in the
issue title or body.

## Security Issues

Do not open public issues for vulnerabilities or sensitive security concerns.
Use GitHub private vulnerability reporting instead.

Do not include credentials, private keys, tokens, customer data, or sensitive
Apple Ads account information in issues, PRs, logs, screenshots, or test
fixtures.

## Contribution Terms

By submitting a contribution, you agree that your contribution may be used,
modified, sublicensed, and distributed under the project license, the MIT
License, or both, without compensation. This permission is irrevocable and
continues to apply if the project license changes in the future, including for
commercial, private, or proprietary distributions.
