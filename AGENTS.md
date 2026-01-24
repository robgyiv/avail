# Repository Guidelines

## Project Structure & Module Organization
- `cmd/avail/` hosts the CLI entrypoint and wiring.
- `internal/` contains app internals (CLI commands, calendar providers, config, API client, auth helpers). Treat these packages as non-public.
- `pkg/` holds reusable components; the availability engine lives in `pkg/engine/`.
- `bin/` is a local build output folder (ignored in CI).
- `.github/workflows/ci.yml` defines formatting and test checks.

## Build, Test, and Development Commands
- `go build -o bin/avail ./cmd/avail` builds the CLI binary locally.
- `go test ./...` runs the full unit test suite.
- `gofmt -w .` formats all Go files (CI enforces formatting).
- Cross-build example: `GOOS=linux GOARCH=amd64 go build -o bin/avail-linux-amd64 ./cmd/avail`.

## Coding Style & Naming Conventions
- Follow standard Go style; run `gofmt -w .` before committing.
- Package and file names should be short, lower-case, and descriptive (`calendar`, `config`, `engine`).
- Keep CLI commands organized under `internal/cli/` with subcommands matching the user-facing verbs (`show`, `copy`, `push`, `auth`).

## Testing Guidelines
- Tests use Go’s standard `testing` package.
- Test files follow `*_test.go` naming and live next to implementation files (see `internal/` and `pkg/engine/`).
- Run `go test ./...` before pushing; no explicit coverage threshold is enforced in CI.

## Commit & Pull Request Guidelines
- Commit messages follow a conventional style such as `feat: ...`, `fix: ...`, or `chore: ...`.
- Keep commits scoped and descriptive; prefer one logical change per commit.
- For PRs, include a short summary, testing notes (e.g., `go test ./...`), and any relevant context or issue links. Screenshots are usually unnecessary for this CLI project.

## Configuration & Secrets
- Local config lives at `~/.config/avail/config.toml`.
- Credentials are stored in the system keyring or `~/.config/avail/credentials` for API tokens; avoid committing secrets.
- Use the Go version declared in `go.mod` (currently `1.25.5`).
