# HANDOFF — proof-first-auditpack

This document is written for a future maintainer (or client) who needs to **build, run, and trust** this tool
without contacting the original author.

## Canonical commands

```bash
# Proof gate (one command)
make verify

# Proof gates (portable, no Makefile)
go test -count=1 ./...
go run ./cmd/auditpack demo --out ./out
```

## What this tool does

`proof-first-auditpack` generates a small “audit pack” for a directory of files:

- `manifest.json` — a deterministic list of files (relative paths), sizes, and SHA-256 hashes
- `run_meta.json` — how the pack was produced (tool + version + timestamps + counts)
- `manifest.sha256` — checksums for the *outputs* so the pack can be validated with standard tools

**Primary use case:** produce a repeatable, verifiable record of “these exact inputs existed at this time”.

## What this tool does *not* do

- It does not upload anything.
- It does not require a database.
- It does not require Docker.
- It does not watch folders or run as a service.

This is intentionally a **run-once, handoff-friendly CLI**.

---

## Requirements

- Go (any reasonably recent version; CI uses `setup-go` with `stable`)
- Linux/macOS/Windows

---

## Acceptance tests (project definition of “done”)

These must pass on a clean checkout:

```bash
go test -count=1 ./...
```

Golden fixture check (included in tests):

- `fixtures/input/case01/*` is the canonical example input tree.
- `fixtures/expected/case01/*` is the canonical expected output pack.
- Tests confirm outputs match expected **byte-for-byte**.

---

## Build

From the repo root:

```bash
# default build (version=dev)
make build

# embed a release version string
make build VERSION=vX.Y.Z

# check what you built
./bin/auditpack version
```

---

## Run

### Generate an audit pack

```bash
# input: directory tree to record
# out:   directory that will receive manifest/meta/sha files
./bin/auditpack run --in /path/to/input_dir --out /path/to/out_dir \
```

Notes:
- `--label` is optional. Use it when `--in` is an absolute path and you want stable, portable metadata.
- If `--out` is inside `--in` (e.g. `--in . --out ./out`), auditpack will exclude the `--out` subtree from hashing to avoid "self-capturing" old packs.


## Optional: Python check (stdlib only)

```bash
# Go run (fixture case01)
go run ./cmd/auditpack demo --out ./out

# Python verification (pack + input tree)
python3 examples/python/verify_auditpack_case.py --in ./out/demo_input --pack ./out
```

This is an independent verification lane (no third-party deps).

### Demo mode (creates a tiny input tree for you)

```bash
./bin/auditpack demo --out ./out
ls -la ./out
```

---

## Verify

There are **two** distinct verification modes:

### 1) Verify the audit pack outputs

This checks that the *pack files themselves* were not modified after creation.

#### Option A: standard tool (no Go required)

From inside the output directory:

```bash
sha256sum -c manifest.sha256
```

Expected output should show `OK` for each line.

#### Option B: built-in verifier

```bash
./bin/auditpack verify --pack /path/to/out_dir
```

This:
- validates `manifest.sha256`
- checks `manifest.json` invariants (sorted/unique paths; stable totals)

### 2) Verify an input tree matches the manifest (optional)

If you still have the original input directory, you can confirm its file hashes match what was recorded in `manifest.json`:

```bash
./bin/auditpack verify --pack /path/to/out_dir --in /path/to/input_dir --strict
```

Notes:
- `--in` is optional. Without it, verification is “pack integrity only”.
- `--strict` fails if extra files exist under `--in` that are not in the manifest.

---

## Self-check (client-friendly smoke test)

Runs **build -> verify -> OK** in a temporary directory and prints `OK` if everything works.

```bash
./bin/auditpack self-check
```

To keep the temp directory (for inspection):

```bash
./bin/auditpack self-check --keep
```

---

## Outputs (what to hand to someone)

The output directory will contain:

- `manifest.json`
- `run_meta.json`
- `manifest.sha256`

### Determinism rules

- Only regular files are included.
- Paths are stored as **relative** paths with **forward slashes**.
- Entries are sorted by normalized path.
- Same inputs ⇒ same `manifest.json` (byte-stable), assuming file contents are unchanged.

---

## Handoff checklist

A new developer should be able to:

1. Clone the repo
2. Run `go test -count=1 ./...` successfully
3. Build `bin/auditpack`
4. Run `auditpack version` to confirm the embedded version string
5. Run `auditpack run --in ... --out ...`
6. Verify the output pack with either:
   - `sha256sum -c manifest.sha256`, or
   - `auditpack verify --pack ...`
7. (Optional) Verify an input tree matches the manifest:
   - `auditpack verify --pack ... --in ... --strict`
8. Run `auditpack self-check` and see `OK`

---

## Common pitfalls / troubleshooting

- **CI fails but local passes:** ensure all new files are committed (fixtures + internal packages + tests).
- **Release notes run commands unexpectedly:** if you see `auditpack: command not found` when running `gh release create`,
  it’s usually because backticks were evaluated by your shell. Prefer:
  - a notes file: `gh release create vX.Y.Z --notes-file RELEASE_NOTES.md`
  - or single quotes around `--notes '...no backticks...'`
- **Empty input directory:** v0 fails fast if no regular files are found under `--in`.
- **Line endings:** keep files UTF-8; this repo enforces LF via `.gitattributes`.

---

## Support policy

This repo is designed so you do *not* need ongoing support:

- No external dependencies beyond Go
- No background processess
- Deterministic, test-defined behavior

If you add features, keep the same philosophy: **simple CLI + fixtures + golden tests + clear docs**.