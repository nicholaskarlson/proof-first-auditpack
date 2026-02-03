# HANDOFF — proof-first-auditpack

This document is written for a future maintainer (or client) who needs to **build, run, and trust** this tool
without contacting the original author.

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
- Linux/macOS/Windows (paths are normalized in output)

---

## Build

From the repo root:

```bash
go test ./...
mkdir -p bin
go build -o bin/auditpack ./cmd/auditpack
```

---

## Run

### Generate an audit pack

```bash
# input: directory tree to record
# out:   directory that will receive manifest/meta/sha files
./bin/auditpack run --in /path/to/input_dir --out /path/to/out_dir
```

### Demo mode (creates a tiny input tree for you)

```bash
./bin/auditpack demo --out ./out
ls -la ./out
```

---

## Outputs

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

## Verify (no custom tooling required)

From inside the output directory:

```bash
sha256sum -c manifest.sha256
```

Expected output should show `OK` for each line.

> Note: `manifest.sha256` verifies the audit pack *outputs*.
> The `manifest.json` contains the per-file hashes for the original input tree.

---

## Acceptance tests (project definition of “done”)

These must pass on a clean checkout:

```bash
go test ./...
```

Golden fixture check (included in tests):

- `fixtures/input/case01/*` is the canonical example input tree.
- `fixtures/expected/case01/*` is the canonical expected output pack.
- Tests confirm outputs match expected **byte-for-byte**.

---

## Handoff checklist

A new developer should be able to:

1. Clone the repo
2. Run `go test ./...` successfully
3. Build `bin/auditpack`
4. Run `auditpack run --in ... --out ...`
5. Verify the output pack with `sha256sum -c manifest.sha256`
6. Re-run and confirm outputs are stable (optional but recommended)

---

## Common pitfalls / troubleshooting

- **CI fails but local passes:** make sure all files are committed (fixtures + internal packages).
- **Hidden Unicode warning:** avoid copy/pasting CLI code from rich text sources. Re-run `gofmt` and keep files UTF-8.
- **Empty input directory:** v0 fails fast if no regular files are found under `--in`.

---

## Support policy

This repo is designed so you do *not* need ongoing support:

- No external dependencies beyond Go
- No background processes
- Deterministic, test-defined behavior

If you add features, keep the same philosophy: **simple CLI + fixtures + golden tests + clear docs**.
