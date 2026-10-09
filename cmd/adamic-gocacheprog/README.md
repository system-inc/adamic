# adamic-gocacheprog

Build with `go build -o /absolute/path/adamic-gocacheprog ./cmd/adamic-gocacheprog`
and set `GOCACHEPROG=/absolute/path/adamic-gocacheprog` when running Go 1.24+.
The helper must be built before enabling it.

| Variable | Default / behavior |
| --- | --- |
| `ADAMIC_GOCACHE_DIR` | User cache directory's `adamic-gocache` |
| `ADAMIC_GOCACHE_STORE` | `https://adamic-store.kirkouimet.com`; `off` disables all remote operations |
| `ADAMIC_GOCACHE_WRITE` | `https://loom.kirkouimet.com/public` |
| `ADAMIC_GOCACHE_TOKEN` | Path to a readable bearer token file; absent/unreadable means no remote writes |
| `ADAMIC_GOCACHE_TRUST` | `main` reads/writes only `gocache`; other values write `gocache-candidate`, read `gocache` then `gocache-candidate` |

Refs at `/refs/<namespace>/<action ID hex>` contain the hex SHA-256 of a
JSON record blob with `output_id` (base64), `size`, and `sha256` (output blob
hash in hex). All blobs live at `/blobs/<sha256>`. Writes publish the output,
then record, then ref. Every blob read, including local reads, verifies its hash.
Network failures miss; remote write failures only log to stderr. A conflicting
ref never replaces the local action entry. Remote requests time out after three
seconds; up to sixteen requests run concurrently. Redirects are not followed.

Local refs use the same trust namespaces. Files are published atomically and
retained across processes; there is currently no eviction. Do not remove the
cache while a Go invocation is using it: returned paths must survive until close.
Bodies are currently buffered in memory per request.

The protocol follows Go's `cmd/go/internal/cacheprog` and
`cmd/go/internal/cache/prog.go`: a capabilities response with ID zero comes
first, byte slices are base64, put bodies are separate JSON string values only
when their size is nonzero, responses may arrive out of order, and both put
and get return absolute, extension-free disk paths. Go's output ID is itself
the SHA-256 of the body; the record validates that relationship too.

`go test ./cmd/adamic-gocacheprog` uses only local httptest servers and helper
subprocesses, including two actual Go builds with fresh local caches. The second
build must contain no compiler invocation. Each child command has a 25-second
deadline.

See [REPORT.md](REPORT.md) for validation and real-package timings.
