# Proof of compiler/type-string-memo 05c10993 (task #wy2dyf0)

Status: **PART 1 (step 1) only.** Steps 2 to 4 follow in later commits on this branch.

## Box (measured)
- Linux 6.18.44 x86_64, 4 CPUs, 15 GiB RAM. Run as root.
- Go 1.27.1, **Node v24.22.2** (official tarball), **clang 20.1.8** (official LLVM-20.1.8-Linux-X64), wasi-sdk 27 (WASI_SYSROOT set, sha256 pin checked), TypeScript 6.0.3 as the global `tsc` (the box shipped 6.0.2; a gate test refuses that), upstream TypeScript checkout v6.0.3 at 050880ce (= stage3/source.json) as ADAMIC_TYPESCRIPT_SOURCE, `npm ci` in stage3/api, cohere + cohere/TypeScript submodules initialised. Env in `env.sh.txt`. TMPDIR=/tmp/adamic-gate (mode 1777), ADAMIC_BUILD_CACHE_DIR=/tmp/adamic-gate/cache.
- Not installed: wasmtime, workerd.

## Commands (head 05c109931001dfa58a9ed6a49bca98d4c7bc5822, main 006b368745e6a37f18b1427b1a63a78285f92694)
Same set at both (main's driver: `main-run.sh.txt`):
1. `gofmt -l cmd internal` -> no output at either.
2. `go vet ./internal/lower/` -> exit 0 at either.
3. Products first, in waves from cloud/fast-gate/product-order.json (`waves.py.txt`) for internal/native, internal/oracle, internal/lower and what they wait on: 366 products in 3 levels; `failed groups 0` at both.
4. `go test -count=1 -timeout 30m [-skip '^TestProduct_'] ./internal/lower/`
5. `go test -count=1 -timeout 30m -skip '^TestProduct_' ./internal/oracle/`
6. `go test -count=1 -timeout 30m ./stage3/census/latent/...`
7. stage3/meter per its README: `build-probes.py --prepare`, then `shards.py meter` for ADAMIC_TEST_SHARD 0/8..7/8 and `shards.py census` for 0/2, 1/2 (`meter.sh.txt`). The README's audit-shards.py was not run (it is an orchestrator, not a gate unit).

## Package lines (measured; logs/head, logs/main)
| step | head 05c10993 | main 006b3687 |
|---|---|---|
| internal/lower | `ok 523.978s` (run without -skip, so it included TestProduct_*) | `ok 556.160s` (with -skip) |
| internal/oracle | `FAIL 1214.021s` | `FAIL 1202.168s` |
| stage3/census/latent/... | 6 packages `ok`, none FAIL | 6 packages `ok`, none FAIL |
| meter shards 0..7/8 | all `Shard result: PASS` | all PASS |
| census shards 0/2, 1/2 | both FAIL | both FAIL |

A first head lower run (logs/head/lower-first-run-no-ts-source.log) FAILED 5 tests only because ADAMIC_TYPESCRIPT_SOURCE was unset; the 5 pass with it set and the whole package then passed (the row above).

## Failures (all present at BOTH heads, so not this head's)
- internal/oracle: `TestClosureMergeRefusals` subtests `optional_widening_refused/node_fs_file_utimes.a` and `node_fs_file_stat.a`: `source Node: exit 70 ... EPERM: operation not permitted, utime`. Identical 2 subtests at head and main, the only oracle failures at either.
- census shards: `replay_test.ReplayTests setUpClass` asserts the nested unit's status is `attempted`, gets `panic`; same assertion and same unit at both.
- Inferred, not measured: the EPERM is this sandbox's (root with a dropped-uid child). The census panic's cause was not investigated.

**Result of step 1: no failure appears at the head only.**
