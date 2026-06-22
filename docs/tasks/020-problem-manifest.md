# Task 020: Problem Manifest Generation and CLI Resolution

## Goal

Generate a `.leetgo-problem.toml` manifest beside each Problem's Stub and TestSuite, and enable `leetgo test .` and `leetgo submit .` to resolve the Problem from the manifest.

## Dependencies

- None (independent of Verified/Reward Events)

## Likely Files

- `internal/workspace/manifest.go` — new file for manifest read/write
- `internal/workspace/workspace.go` — generate manifest on Start
- `internal/workspace/workspace_test.go` — manifest tests
- `cmd/leetgo/main.go` — update `runProblemTests` and `handleSubmit` to resolve from manifest
- `cmd/leetgo/main_test.go` — CLI resolution tests

## Implementation Notes

### Manifest format

File: `.leetgo-problem.toml`

```toml
problem_id = 1
slug = "two-sum"
roadmap = "from-zero-to-hero"
stage = "arrays-hashing"
language = "go"
stub_path = "two_sum.go"
testsuite_path = "two_sum_test.go"
```

### Manifest write policy

- Write when Start first generates files.
- If manifest exists for the same Problem ID, safely update generated fields.
- If manifest exists for a different Problem ID, stop with error and do not overwrite.

### CLI resolution order

For `leetgo test` and `leetgo submit`:

1. If argument is `.` or current directory: read `.leetgo-problem.toml` from current directory, then walk parents.
2. If argument is a path: read manifest in that path or parents.
3. If argument is a Problem ID or slug: resolve through selected Roadmap/catalog as today.
4. If no manifest and no ID/slug: show clear error.

### Manifest resolution function

```go
func resolveProblem(arg string, cfg *config.Config) (*roadmap.Problem, error)
```

This function handles all four resolution paths and returns the Problem or an error.

## Acceptance Criteria

- `.leetgo-problem.toml` is generated beside Stub and TestSuite on Start.
- Manifest contains problem_id, slug, roadmap, stage, language, stub_path, testsuite_path.
- Manifest is not overwritten if it belongs to a different Problem ID.
- `leetgo test .` resolves Problem from manifest in current directory.
- `leetgo submit .` resolves Problem from manifest in current directory.
- `leetgo test 1` and `leetgo test two-sum` still work as before.
- Missing manifest shows clear error message.
- Tests cover manifest generation, parent walking, ID/slug fallback, and mismatch error.

## Verification

```bash
gofmt -w internal/workspace cmd/leetgo
go vet ./...
go test -count=1 ./...
```
