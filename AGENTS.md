# Repository Guidelines

## Project Structure & Module Organization
Core Go logic lives in `pkg` for reusable transaction primitives and in `internal` for private coordination layers and adapters. Runtime components are split across `executor/`, `ft-executor/`, `timeoracle/`, and `ft-timeoracle/`; each module owns its protocol handlers, entrypoints, and deployment manifests. Benchmark harnesses sit in `benchmarks/`, integration scenarios in `integration/` (currently unstable and gated behind manual setup), and user-facing snippets in `examples/`. Visual assets reside in `assets/`. Additions should keep reusable code in `pkg`, hide service-only helpers in `internal`, and mirror this layout for any new executor variant.

## Build, Test, and Development Commands
- `go build ./...` — compile every Go module and surface cross-package issues.
- `go test ./...` — run unit tests; append `-race` when touching concurrency-sensitive paths.
- `golangci-lint run` — execute the lint suite that backs CI.
- `pre-commit run --all-files` — run the repo hooks locally before committing.
- `just --list` — discover available task recipes; extend `justfile` when adding repeated flows.

## Coding Style & Naming Conventions
Format Go sources with `gofmt` (tabs for indents) and keep imports ordered via `goimports`. Package names stay short, lower-case, and words should not be reused across levels (e.g., `timeoracle/client`). Exported identifiers use CamelCase and document behavior with full sentences; unexported helpers prefer concise names. Tests live in `_test.go` files and match the subject file’s package. Run `golangci-lint` after structural changes to confirm style and static checks line up with CI.

## Testing Guidelines
Write table-driven Go tests and keep fixtures alongside the code under test. Use `go test -cover ./pkg/...` to verify coverage for shared primitives and `go test ./internal/...` for coordinator logic. Integration suites under `integration/` assume external NoSQL services; gate them behind environment checks and document required endpoints in the PR. Benchmarks in `benchmarks/` run with `go test -bench . ./benchmarks/...`; record baseline numbers in the PR when performance is affected.

## Commit & Pull Request Guidelines
Follow Conventional Commits as in `feat:`, `refactor:`, and `ci:` from the existing history. Keep subject lines in the imperative mood and describe user impact in the body when the change is non-trivial. Every PR should include: a brief summary, linked issues (GitHub `#123` or equivalent), test evidence (`go test`, lint output, or benchmark deltas), and screenshots or logs for UI or tooling updates. Re-run `pre-commit` and `golangci-lint` before requesting review, and ensure reviewers can reproduce any new integration steps.
