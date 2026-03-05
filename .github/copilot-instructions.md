# Copilot Instructions

## Commands

```bash
go build -v .          # build
go test ./...          # run all tests
go test -run TestName  # run a single test by name
```

## Architecture

Two-phase duplicate detection pipeline:

1. **Size pass** (`sizemap.go`): `CreateSizeMap` walks the directory tree and builds a `SizeMap` (`map[int64]*list.List`), grouping file paths by byte size. Only files that share a size with at least one other file are candidates.

2. **Checksum pass** (`finddups.go`): `CreateDuplicationMap` takes the `SizeMap`, fans out MD5 checksum jobs only for candidate files, and builds a `DuplicateMap` (`map[EqualFile]*list.List`) keyed by `{checksum, size}`.

3. **Concurrency** (`workerqueue.go`): `RunWorkers` is a generic worker pool taking a `Provider` (sends `DataMap` jobs onto a channel) and a `Consumer` (processes one job, returns a `DataMap`). The number of goroutines is controlled by the `-threads` flag (defaults to `runtime.NumCPU()`).

`DataMap` (`map[string]interface{}`) is the generic message type passed between producer and workers.

## Key Conventions

- Tests use `github.com/test-go/testify/assert` (not the more common `github.com/stretchr/testify`).
- Test structure follows Setup / Execute / Verify comment blocks.
- `testdata/` contains small fixture files used directly by tests (e.g. `testdata/a`, `testdata/subdir/c`).
- The compiled binary `duplicatefinder` is committed to the repo root — do not delete it.
