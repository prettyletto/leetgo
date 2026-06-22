# Pure-Go SQLite (modernc.org/sqlite)

We use `modernc.org/sqlite` instead of `mattn/go-sqlite3` to avoid CGO entirely. This keeps the binary fully static and cross-compilable without a C toolchain.

**Considered Options**: mattn/go-sqlite3 (CGO, faster), flat JSON files, pure-Go sqlite.
**Consequences**: Slightly slower than CGO sqlite, but the dataset is small (< 10k rows) so performance is not a concern. Simplifies CI, cross-compilation, and distribution.
