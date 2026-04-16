# AGENTS

This file is for coding agents working in this repository.
It is intentionally narrow: product behavior, architecture, and endpoint status live in the docs under `docs/`.

## Read First

Read these documents before making non-trivial changes:

1. `docs/requirements.md`
   Use for CLI behavior, user-facing conventions, safety rules, and testing requirements.
2. `docs/architecture.md`
   Use for package boundaries, code organization, generation pipeline, and implementation strategy.
3. `docs/coverage.md`
   Use for documented-vs-implemented endpoint coverage and known API path discrepancies.

## Discovering Commands

- Before implementing or testing a command, check the current CLI help output instead of relying on memory.
- Use:
  - `bin/aads --help`
  - `bin/aads <group> --help`
  - `bin/aads <group> <command> --help`
- Treat help text as part of the user-facing contract. If command behavior changes intentionally, update docs and tests to match.
- The CLI binary lives at `bin/aads`.
- Refresh it with `make build` before checking help or running manual command verification when needed.

## Source Of Truth

- `docs/requirements.md` is the source of truth for intended CLI behavior.
- `docs/architecture.md` is the source of truth for project structure and design intent.
- `docs/coverage.md` is the source of truth for endpoint coverage status.
- `docs/apple_ads/openapi-latest.json`, versioned `docs/apple_ads/openapi-v*.json`, and `docs/apple_ads/paths.txt` are generated references from Apple docs, not hand-maintained runtime truth.
- If tests, real API behavior, and generated docs disagree, do not blindly trust the generated docs. Reconcile the discrepancy and update `docs/coverage.md`.

## When You Change X, Also Update Y

- If you intentionally change user-facing CLI behavior, update `docs/requirements.md`.
- If you change architecture, package boundaries, generation flow, or implementation strategy, update `docs/architecture.md`.
- If you add, remove, or reroute API request files or endpoint coverage, update `docs/coverage.md`.
- If you change the Apple docs generation script or the derived docs artifacts, regenerate the generated files and update any related docs that describe the workflow.

## Generated Files

- Do not hand-edit generated artifacts unless the task explicitly asks for it.
- When generated Apple docs artifacts need to change, regenerate them using the documented workflow in `docs/architecture.md`.
- Do not hand-edit `docs/commands.md`; regenerate it with the appropriate `make` command(s).

## Verification

- Prefer targeted tests for the code you changed.
- If you change command routing, help behavior, or end-to-end CLI behavior, run relevant tests under `./cmd`.
- If you change API request paths or request construction, run the relevant package tests under `internal/api/requests/...`.
- If you change shared CLI helpers, run the relevant tests under `internal/cli/shared`.
- If you update generated Apple docs artifacts, ensure `make openapi` or the equivalent script invocation succeeds.

## Testing Discipline

- Use tests for behavior changes: bugs, new flags, changed routing, and output changes.
- For CLI behavior, start with command-level tests that assert exit code, stderr/stdout, and structured output where applicable.
- For JSON output, prefer parsing the output over fragile string matching when structure matters.
- For repetitive request wiring, prefer focused table-driven tests with at least one representative high-signal assertion.
- Do not accept flags that are silently ignored. If a flag is supported, it must affect behavior or validation.

## Credential Safety

- Do not manually read, print, inspect, or modify the default user configuration or token cache under `~/.aads/`.
- If a task requires configuration inspection, use a custom directory provided via `--config-dir` or `AADS_CONFIG_DIR`, or create an isolated test directory.
- Running tests that exercise default-path behavior is allowed, as long as you do not intentionally inspect or depend on the real contents of `~/.aads/`.
- Do not use real user credentials from `~/.aads/` in development, debugging, or tests.

## Debugging

- Reproduce the issue first before changing code.
- If the user provides a concrete failing command, run that exact command first when feasible, then keep rerunning that same command after each relevant fix until it works as expected.
- Make one logical fix at a time and re-run the most specific affected test after each fix.
- When changing request paths or API routing, verify both the routed endpoint and the request payload shape.
- If generated Apple docs disagree with tests or observed API behavior, document the discrepancy in `docs/coverage.md`.

## Definition Of Done

- The code change is implemented and the relevant tests pass.
- User-facing behavior changes are reflected in `docs/requirements.md`.
- Architecture or generation-flow changes are reflected in `docs/architecture.md`.
- Endpoint coverage changes are reflected in `docs/coverage.md`.
- If generated Apple docs artifacts are affected, regenerate them rather than editing them by hand.

## Repo Gotchas

- Filtered targeting-keyword find is campaign-scoped in Apple docs:
  - `POST /campaigns/{campaignId}/adgroups/targetingkeywords/find`
  - `adGroupId` filtering is expressed in the selector body, not the path.
- Filtered ad-group negative-keyword find is campaign-scoped in Apple docs:
  - `POST /campaigns/{campaignId}/adgroups/negativekeywords/find`
  - `adGroupId` filtering is expressed in the selector body, not the path.
- `docs/coverage.md` should remain the single coverage/status document. Do not reintroduce large endpoint coverage tables into `docs/requirements.md` or `docs/architecture.md`.

## Preferred Workflow

- Inspect existing code before editing.
- Keep changes scoped to the task.
- Preserve user changes and unrelated local modifications.
- Update docs when behavior or architecture changes, not as an afterthought.
