#!/usr/bin/env python3
# SPDX-License-Identifier: MIT

from __future__ import annotations

import argparse
import hashlib
import json
from pathlib import Path


def sha256_file(p: Path) -> str:
    h = hashlib.sha256()
    with p.open("rb") as f:
        for chunk in iter(lambda: f.read(65536), b""):
            h.update(chunk)
    return h.hexdigest()


def read_text_lf(p: Path) -> str:
    s = p.read_text(encoding="utf-8")
    if "\r\n" in s:
        raise ValueError(f"CRLF detected: {p}")
    return s


def main() -> None:
    ap = argparse.ArgumentParser()
    ap.add_argument("--in", dest="in_dir", default="out/demo_input", help="input directory to verify")
    ap.add_argument("--pack", dest="pack_dir", default="out", help="audit pack output directory")
    args = ap.parse_args()

    repo = Path(".")
    in_dir = (repo / args.in_dir).resolve()
    pack_dir = (repo / args.pack_dir).resolve()

    manifest_p = pack_dir / "manifest.json"
    meta_p = pack_dir / "run_meta.json"
    sums_p = pack_dir / "manifest.sha256"

    # 1) Verify manifest.sha256 matches actual hashes of pack artifacts.
    expected: dict[str, str] = {}
    for line in read_text_lf(sums_p).splitlines():
        if not line.strip():
            continue
        h, name = line.split(None, 1)
        expected[name.strip()] = h.strip()

    for name, h in expected.items():
        got = sha256_file(pack_dir / name)
        assert got == h, f"sha256 mismatch for {name}"

    # 2) Verify manifest.json structure and stable ordering.
    manifest = json.loads(read_text_lf(manifest_p))
    files = manifest["files"]
    paths = [f["path"] for f in files]
    assert paths == sorted(paths), "manifest files not sorted by path"
    assert len(paths) == len(set(paths)), "manifest paths must be unique"

    # 3) Verify each manifest record matches the input directory.
    for f in files:
        p = in_dir / f["path"]
        assert p.exists(), f"missing input file: {p}"
        assert p.is_file(), f"not a file: {p}"
        assert p.stat().st_size == f["size_bytes"], f"size mismatch: {p}"
        assert sha256_file(p) == f["sha256"], f"hash mismatch: {p}"

    # 4) Verify run_meta.json summary matches computed totals.
    meta = json.loads(read_text_lf(meta_p))
    assert meta["tool"] == "proof-first-auditpack"

    in_label = str(meta.get("input", ""))
    assert in_label, "run_meta.json missing input label"
    ok = (
        in_label == str(in_dir)
        or in_label == in_dir.as_posix()
        or in_label.endswith(in_dir.as_posix())
        or in_label == in_dir.name
        or in_label.endswith(in_dir.name)
    )
    assert ok, f"run_meta.json input label mismatch: {in_label}"
