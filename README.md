# proof-first-auditpack

Deterministic audit pack generator (Go-first): SHA-256 manifest + run metadata + integrity verification.

![ci](https://github.com/nicholaskarlson/proof-first-auditpack/actions/workflows/ci.yml/badge.svg)
![license](https://img.shields.io/badge/license-MIT-blue.svg)

## What this is

`proof-first-auditpack` is a small, handoff-friendly CLI that turns an input directory tree into a verifiable “audit pack”:

- `manifest.json` — deterministic list of files (relative paths), sizes, SHA-256 hashes
- `run_meta.json` — how the pack was produced (tool version, timestamps, counts)
- `manifest.sha256` — checksums for the *pack outputs* so anyone can validate the pack with standard tools

**Primary use case:** produce a repeatable, verifiable record of “these exact inputs existed at this time”.

## What this is not

- No database
- No Docker requirement
- No background service / daemon
- No uploads / networking
- No vendor lock-in

This is intentionally **run-once, deterministic, and easy to hand off**.

## Canonical commands

```bash
# Proof gate (one command)
make verify

# Proof gates (portable, no Makefile)
go test -count=1 ./...
go run ./cmd/auditpack demo --out ./out
```

## Quick start

```bash
# Demo: recomputes fixture case01 and verifies pack outputs match goldens
go run ./cmd/auditpack demo --out ./out
# outputs land in ./out/case01

# Optional: synthetic quick demo (tiny in-memory input tree)
go run ./cmd/auditpack demo --out ./out --quick

# Build a pack from your own input tree
go run ./cmd/auditpack run --in /path/to/input_dir --out /path/to/out_dir

# Verify a pack
go run ./cmd/auditpack verify --pack /path/to/out_dir
```

### Optional convenience

```bash
make test
make demo
make self-check
make build VERSION=vX.Y.Z
```

## Core commands

### Build a pack

```bash
go run ./cmd/auditpack run --in /path/to/input_dir --out /path/to/out_dir \
  --label fixtures/input/case01
```

`--label` is optional. Use it when `--in` is an absolute path and you want stable, portable metadata.

If `--out` is inside `--in` (e.g. `--in . --out ./out`), auditpack will exclude the `--out` subtree from hashing to avoid "self-capturing" old packs.

### Verify a pack

Verifies `manifest.sha256` (pack output integrity) and basic invariants on `manifest.json` (sorted paths, uniqueness, stable totals).

```bash
go run ./cmd/auditpack verify --pack /path/to/out_dir
```

### Verify the original input tree (optional)

If you still have the input tree, you can validate it matches the recorded hashes:

```bash
go run ./cmd/auditpack verify --pack /path/to/out_dir --in /path/to/input_dir --strict
```

## Fixtures + tests

- Canonical input tree: `fixtures/input/case01/`
- Canonical expected pack: `fixtures/expected/case01/`

Golden tests verify outputs byte-for-byte:

```bash
go test -count=1 ./...
```

## Repo layout (high level)

- `cmd/auditpack/` — CLI entrypoint
- `internal/` — build/verify/self-check engine
- `fixtures/` — canonical input + expected output packs
- `tests/` — golden tests and verifier tests
- `docs/` — handoff and maintenance notes

## Handoff

See: `docs/HANDOFF.md` (build/run/verify, acceptance tests, troubleshooting).

## License

MIT (see `LICENSE`).
