# proof-first-auditpack

Deterministic audit pack generator (Go-first): SHA-256 manifest + run metadata + integrity verification.

![ci](https://github.com/nicholaskarlson/proof-first-auditpack/actions/workflows/ci.yml/badge.svg)
![license](https://img.shields.io/badge/license-MIT-blue.svg)

> **Book:** *The Deterministic Finance Toolkit*
> This repo is **Project 2 of 4**. The exact code referenced in the manuscript is tagged **[`book-v1`](https://github.com/nicholaskarlson/proof-first-auditpack/tree/book-v1)**.

## Toolkit navigation

- **[proof-first-recon](https://github.com/nicholaskarlson/proof-first-recon)** — deterministic CSV reconciliation (matched/unmatched + summary JSON)
- **[proof-first-auditpack](https://github.com/nicholaskarlson/proof-first-auditpack)** — deterministic audit packs (manifest.json + sha256 + verify)
- **[proof-first-normalizer](https://github.com/nicholaskarlson/proof-first-normalizer)** — deterministic CSV normalize + validate (schema → normalized.csv/errors.csv/report.json)
- **[proof-first-finance-calc](https://github.com/nicholaskarlson/proof-first-finance-calc)** — proof-first finance calc service (Amortization v1 API + demo)

## What it does

Turns an input directory tree into a verifiable “audit pack”:

- `manifest.json` — deterministic list of files (relative paths), sizes, SHA-256 hashes
- `run_meta.json` — how the pack was produced (**tool + version + input label + counts**)
- `manifest.sha256` — checksums for the *pack outputs* so anyone can validate the pack with standard tools

**Primary use case:** produce a repeatable, verifiable record of “these exact inputs existed”.

## What this is not

- No database
- No Docker requirement
- No background service / daemon
- No uploads / networking
- No vendor lock-in

This is intentionally **run-once, deterministic, and easy to hand off**.

## Quick start

Requirements:
- Go **1.22+**
- GNU Make (optional, but recommended)

```bash
# One-command proof gate
make verify

# Portable proof gate (no Makefile)
go test -count=1 ./...
go run ./cmd/auditpack demo --out ./out
```


## Core commands

### Build a pack

```bash
go run ./cmd/auditpack run --in /path/to/input_dir --out /path/to/out_dir \
  --label input_dir
```

Notes:
- `--label` is optional. Use it when `--in` is an absolute path and you want stable, portable metadata.
- If `--out` is inside `--in` (e.g. `--in . --out ./out`), auditpack excludes the `--out` subtree from hashing to avoid “self-capturing” old packs.

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

## Fixtures + proof gate

The acceptance gate is `make verify`, which runs:
- unit tests (`go test -count=1 ./...`)
- a full demo run that regenerates the canonical pack and checks it matches goldens

Fixtures:
- Canonical input tree: `fixtures/input/case01/`
- Canonical expected pack: `fixtures/expected/case01/`

## Repo layout (high level)

- `cmd/auditpack/` — CLI entrypoint
- `internal/` — build/verify/self-check engine
- `fixtures/` — canonical input + expected output packs
- `tests/` — tests + verifier invariants
- `docs/` — handoff and maintenance notes

## Determinism contract

This project is intentionally “boring” in the best way: the same inputs must produce the same outputs.

See: **[`docs/CONVENTIONS.md`](docs/CONVENTIONS.md)** (rounding, ordering, LF, atomic writes, stable JSON, etc.).


## Handoff / maintenance

See: **[`docs/HANDOFF.md`](docs/HANDOFF.md)** (acceptance gates, troubleshooting, and “what to change (and what not to)”).


## License

MIT (see `LICENSE`).

