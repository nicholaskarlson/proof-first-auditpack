# Optional Python checks (stdlib only)

Python is **optional** in this repo. The source-of-truth proof gates are in Go:

```bash
go test -count=1 ./...
```

This folder provides an *independent verification lane* (standard library only).

## Auditpack run + Python verify (fixture case01)

From repo root:

```bash
# Go run (writes manifest.json + run_meta.json + manifest.sha256)
go run ./cmd/auditpack demo --out ./out

# Python verification (pack + input tree)
python3 examples/python/verify_auditpack_case.py --in ./out/demo_input --pack ./out
```
