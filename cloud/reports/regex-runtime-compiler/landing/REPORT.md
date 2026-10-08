Built: merged area/library d0735bda into runtime compiler ef3bbf53, preserving dynamic compilation and area intent.
Commit: this merge commit on codex/regex-runtime-compiler; no rebase or force push.
Commands/results: Linux counts regenerated in 181.247s; dynamic Node/backends/WASI gate passes in 133.333s; common-baseline WASI artifacts byte-identical.
Mutants: all 18 inherited compiler/reference/table/ownership/refusal mutants caught at their intended assertions.
Uncovered: whole repository and full native concurrency gates delegated to the fast gate; no sanitized WASI runtime available.

The ten requested conflicts are resolved together. The sole Build archive-selection hook remains internal/native/native.go:150, `library, err := runtimeLibraryForSource(source, options)`. Exported Slabs/Malloc support is retained alongside the area's FusedRuntime option, and fused flags and dynamic ownership flags both participate in the archive cache key.

Constructor lowering preserves the area's undefined and constant-RegExp clone/identity handling; nonconstant string patterns/flags reach the runtime compiler. Unsupported nonconstant RegExp-object sources remain explicitly refused. Constant SyntaxError constructors use the area's refusal classification, while V8 divergences preserve their exact refusal reason. Throw discovery combines runtime constructor errors and fromCodePoint errors without erasing an earlier throw. JavaScript preserves nullish optional calls and the concurrent-map adapter.

Captured-stack support and string well-formedness headers are retained with packed atomic slot caches and concurrent string-index publication. The optimized regular matcher dispatch is retained; its test/embedding mode uses atomic storage, like the step budget, and is listed in docs/runtime-statics.md. The fuzzer retains the isolated-environment API and the area's kernel CPU limits with a wall backstop. Its infinite, SIGXCPU and loaded-box executor witnesses pass in 5.853s.

The first contract run passed every compiler test but found new inherited storage missing from the statics inventory. Six runtime units and nine declarations were reviewed and documented; the repaired statics/layout gate passes in 0.468s, including 39 layouts. The inherited shape-type registry is explicitly documented as unsynchronized for parallel registration/dynamic reads; no concurrency certification of that area behavior is claimed. The runtime compiler adds no mutable tables or caches. A final unified contract run is recorded in contract-final.log.

The bytecode corpus remains 5,746 test262 patterns, 875 pinned cohere patterns, 4,000 seeded patterns and nine ruling probes. Linux/WASI compare 7,433 byte-identical programs and 3,197 syntax/refusal cases; V8-shape probes compare 7,677 programs and 3,354 refusals. Complete Node SyntaxError messages, including counted bytes after embedded NUL, agree. Property identity checks 1,722 aliases, 448 unique properties, 23,045 ranges and 7,906 strings. All compiler C units execute on WASI with 32-bit pointers and no atomics; Linux is the ASan/UBSan gate of record.

Seven dynamic fixtures pass source Node, emitted JavaScript, native sanitizers/release/slabs/leak checks and WASI. Refusal, ownership and counted-error mutants are rerun. The seven Go-reference mutants and five isolated table mutants also pass. No mutant is counted as killed by a compiler diagnostic.

Linux TestCountsAreRecorded -update-counts regenerated the complete merged fixture registry. It passes in 181.247s; internal/oracle/counts.md is part of this merge. The fast gate handles remaining area-wide behavior.

| Fixture | WASI raw before | Raw after | Brotli before | Brotli after |
|---|---:|---:|---:|---:|
| hello | 288690 | 288690 | 84091 | 84091 |
| request.a | 325292 | 325292 | 97236 | 97236 |

Both files are byte-identical on an identical merged common baseline. The baseline Go overlay omits only five unused compiler translation units; all other front/runtime changes are held constant. The enabled dynamic_gap fixture is 826519 raw / 161138 Brotli bytes. Brotli uses quality 11. The measurement script and numeric artifact are committed here. Ordinary-program symbol and ownership opt-in mutants independently prove the compiler stays unlinked when unused.

```sh
source /workspace/adamic-tools/env.sh
export GOPROXY='https://proxy.golang.org|direct'
export WASI_SYSROOT=/workspace/adamic-tools/wasi-sdk/share/wasi-sysroot
GOFLAGS=-buildvcs=false go test ./internal/native -run '^TestRegExpRuntime|^TestRuntimeStaticsAreListed$|^TestRuntimeFieldLayoutsAreIncluded$' -count=1 -v -timeout=30m
GOFLAGS=-buildvcs=false go test ./internal/regexp -run '^TestRuntimeReference' -count=1 -v
python3 internal/regexp/testdata/run-runtime-table-mutants.py
ADAMIC_GATE_UNCACHED=1 ADAMIC_ORACLE_WASI=1 GOFLAGS=-buildvcs=false go test ./internal/oracle -run '^(TestNativeAgreesWithNode|TestWASIAgreesWithNode)/internal/oracle/testdata/regexp_dynamic/|^TestDynamicRegExp' -count=1 -v -timeout=30m
GOFLAGS=-buildvcs=false go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout=30m -args -update-counts
python3 cloud/reports/regex-runtime-compiler/landing/measure-sizes.py
```

Touched-package vet passes. The Linux toolchain and exact cohere pin are unchanged from the established worker environment; GOFLAGS disables VCS stamping on the shared submodule symlink. Old disposable Go cache and confirmed abandoned scratch build directories were reclaimed to prevent disk exhaustion. All test outputs were written directly to logs.

Final unified compiler contract: PASS, 41.350s, including the repaired statics and field-layout scans. Full-parent whitespace checks report existing imported licenses, raw logs, diff artifacts and CSV line endings from both branches. Those evidence bytes are preserved. The ten resolved source files, runtime audit, regenerated counts and new report/script pass the scoped diff check.
