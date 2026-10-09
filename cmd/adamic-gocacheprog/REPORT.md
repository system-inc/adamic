# Shared Go cache validation

Measured on 2026-10-09, Linux amd64, Go 1.27.1, a four-CPU cgroup quota,
`GOMAXPROCS=4`, and `go test -p=4 -count=1 -x ./internal/lower`.
Each invocation used separate, initially empty `GOCACHE` and
`ADAMIC_GOCACHE_DIR` directories. The toolchain, module dependencies, clang,
and the pinned stage3/api Node dependencies were already installed.

| Run | Wall time | Go compiler invocations | Test execution time printed by Go | Result |
| --- | ---: | ---: | ---: | --- |
| Cold, remote disabled | 257.062 s | 244 | 56.950 s | Passed |
| Prime empty shared store | 327.508 s | 244 | 62.682 s | Passed |
| Cold local caches, warm shared store | 49.646 s | 0 | 42.457 s | Passed |

The store was an isolated, filesystem-backed HTTP server on loopback, with
public GETs and bearer-authenticated PUTs implementing the blob/ref paths.
Priming used the helper's normal Go put protocol with `ADAMIC_GOCACHE_TRUST=main`.
The warm run had no write token. Neither priming nor the warm run reported
cache diagnostics. `-count=1` forced actual test execution in all runs; the
warm trace contained no `/compile ` invocation. The observed wall-time
improvement was approximately 5.18 times. These timings include downloading,
hashing and writing cached artifacts, but do not represent production HTTPS
latency. Test execution times varied, as shown separately above.

Validation:

- `GOWORK=off GOMAXPROCS=4 GORACE=atexit_sleep_ms=0 go test -race -count=1 ./cmd/adamic-gocacheprog`: passed in 0.377 s. The race detector remained enabled; its subprocess exit sleep was disabled. Tests run in parallel with isolated stores/directories; each subprocess has a 25-second deadline.
- `GOWORK=off go vet ./cmd/adamic-gocacheprog`: passed.
- Process tests covered shared bytes, empty output, output and record poisoning,
  oversized poisoned output, local poisoning, unavailable stores, missing or
  unreadable tokens, failed uploads, ref conflicts, namespace separation,
  main-ref preference, and corruption stopping namespace fallback.
- The actual Go build test built a small package twice with separate empty
  local caches; the second invocation had no compiler step.
- A temporary mutant that bypassed blob hash validation was rejected by the
  same-length poisoned-output test. Validation was restored before final tests.

Protocol details that were less obvious than a basic get/put API: the helper
must advertise commands immediately with response ID zero; put bodies are
separate base64 JSON string values, with no body value for zero-byte puts;
responses may be out of order; even puts must return an absolute local path
that remains valid until close. Go computes OutputID as the body's SHA-256,
so the record's output ID and blob hash must agree. Records omit timestamps to
remain deterministic across writers.

The current helper buffers bodies per request, limits concurrent requests to
sixteen, and retains its local cache without eviction. Only the new command
package was changed.
