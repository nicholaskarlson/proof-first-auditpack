# CONVENTIONS — Proof-First Repos

This file is shared across the repos in this book so the workflow feels identical.

## Deterministic ordering rules

- CSV outputs are treated as sets: do not depend on input row order.
- Every emitted table must have a documented stable sort, typically by a primary key like `id` ascending.
- If a schema defines column order, outputs must use schema order.
- JSON must be stable: prefer structs over maps, or sort keys before writing.

## Rounding and formatting conventions

- Money values use fixed-point decimal semantics (no floats).
- When formatting money, use the repo-defined scale (commonly 2) and keep trailing zeros.
- Write JSON with consistent indentation and a trailing newline.

## Line endings

- All committed fixtures and goldens use LF (\n).
- Tools may accept CRLF input, but outputs are normalized to LF.

## Atomic writes and overwrite safety

- Write files to a temp path first, then rename into place.
- Avoid leaving partially-written directories when generating multiple outputs.

## No silent ambiguity

Fail fast (do not guess) on:

- Duplicate headers (after any trim/normalization rules)
- Duplicate IDs/keys where uniqueness is required
- Schema mismatches (missing required columns; unexpected columns when not allowed)

## Fixtures and goldens

Layout:

- fixtures/input/<case>/... inputs
- fixtures/expected/<case>/... goldens

Tests and demos should run a case, write outputs to a temp or out directory, then byte-compare against fixtures/expected.

### Expected-fail fixtures

For cases that should fail, commit:

- fixtures/input/<case>/... inputs
- fixtures/expected/<case>/error.txt expected error text (end with a single newline)
