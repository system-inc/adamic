# Established UTF-16 index cache fast path

Branch `codex/perf-utf16-cache`, from main
`fe3b9f236e0672e968bcb7ad53c1882cdde18f29`. The production change is ten added
lines and two replaced lines in `runtime/string_index.c`. `emit.go`, `lower.go`,
`native.go` and the oracle driver are untouched. The other code change is a
Node-backed runtime test for literal, stack and heap cache states.

## Selection from the two port profiles

The evidence is JSON's PERFORMANCE.md at `67796a5` and scanner's PERFORMANCE.md
on `codex/typescript-scanner` (fetched tip `17813f8`). Compare disjoint named
self instructions, with each port given equal weight; do not add overlapping
inclusive rows or mix their unlike corpus sizes into a raw-Ir total:

| Shared proposal family | Scanner self share | JSON self share | Mean share |
| --- | ---: | ---: | ---: |
| UTF-16 lookup and access | 30.973% | 4.584% | 17.778% |
| Ownership / RC traffic | 4.742% | 18.551% | 11.647% |
| String building / slices / joins | 6.429% | 14.377% | 10.403% |

UTF-16 includes locate, units, units-before, usable, unit_at, char-code wrappers
and code-point access. RC includes retain and release. String building uses the
same named-function set as the JSON report. These are shared proposal families,
not a complete partition of either profile. UTF-16 has the largest combined
normalized share among these common proposals. The chosen patch addresses its
repeated index-helper entry, not all of that 30.973% scanner cost.

## Change and before/after instructions

Both byte-to-unit and unit-to-byte lookup now read the existing index locally.
NULL and the literal marker still go through `usable`, which decides whether
to build an index or walk without one. An established index is used directly.
The old helper's first check has exactly this predicate; the patch avoids
calling through its larger build-capable body on every established-cache read.
ASCII handling, lazy construction, checkpoints, cursor behavior, surrogate
halves, cache invalidation, layout and ownership are unchanged.

An earlier bounded-backward-cursor experiment saved only 0.118% scanner and
0.018% JSON Ir and was discarded. Explicitly reading the cached length was also
removed: clang already inlines that path, and removing the extra source logic
preserved the measured result. Neither experiment is in the production diff.

| Port | Before Ir | Final Ir | Reduction |
| --- | ---: | ---: | ---: |
| Scanner | 2,429,650,932 | 2,345,680,271 | **3.456%** |
| JSON bounded sample | 909,698,082 | 908,183,803 | **0.166%** |

| Named self cost | Scanner before / final | JSON before / final |
| --- | ---: | ---: |
| usable | 153,766,828 / 63,250,673 | 1,278,760 / 781,385 |
| locate | 288,080,341 / 294,625,823 | 9,500,949 / 8,542,232 |
| units-before | inlined / inlined | 20,308,907 / 20,250,738 |

Some work moves into the callers; the scanner locate row rises. The whole
process total, not the disappearance of a helper row, proves the reduction.
No wall-time gain is claimed: profiles overlapped verification jobs, and wall
times would not be a controlled comparison.

## Measurement isolation and scope

The implementation branch contains no port merge. The separate worktree
`/tmp/adamic-perf-ports`, branch `codex/scratch-utf16-ports`, merges the scanner
and JSON branches only for measurement; its merge commit is
`4d9b344b1acb15b209eb6105ef9dc55668621870`. The runtime file was copied there
for after-build validation. No port source was optimized in this unit.

Before/final measurement pairs use the **same generated C**, with release
code-generation flags `-O2 -ffp-contract=off -fno-optimize-sibling-calls` plus
`-g`, clang 20.1.8, no sanitizers or RC instrumentation. Callgrind 3.24.0 records
Ir. Its self accounting is checked by the port's existing profile.py; all self
records must sum to its summary. The four raw profiles, top-twenty logs and tool
diagnostics are adjacent to this report.

Scanner: all 77 compiler files of TypeScript v6.0.3,
`050880ce59e30b356b686bd3144efe24f875ebc8`, 434,790 tokens. Count mode still
computes token values and errors; its count agrees with Go and Node. Both
before/final release and profiled snapshots additionally produced all
25,014,914 Go/Node answer bytes on the full merged-checkout scanner corpus:
77 compiler files, 84 stage1 files and 18,236 generated inputs.

JSON: the previous unit's fixed 97-text, 340,876-source-byte sample, including
all generated cases. Its input SHA256 is
`48cc2ac1454143609af5bf87a5f2b1d9a8b9c65957636dcc095d312fe0b11d18`.
Both profiled streams match the independently saved Go answers byte for byte.
This bounded sample excludes the largest JSON files; its 0.166% is not a
full-corpus instruction or throughput claim. Both binaries also run against
all JSON inputs in the merged checkout. That corpus has one extra JSON file,
scanner's measurements.json: **1,076** inputs instead of the earlier 1,075.
The original corpus is retained in full. JSON profiles emit the same nonfatal
Valgrind brk-segment warning as the previous unit and finish normally.

A symlinked cohere initially caused WalkDir to miss submodules and Go overlays
to target a different physical path; that attempted JSON run failed visibly
with missing oracle answers. The scratch checkout was repaired with physical
copies of both submodules, retaining their pins. Only the repaired full-corpus
run is used for the JSON parity claim below.

## Oracle counts and Node validation

The baseline is main's `internal/oracle/counts.md`, **173 rows**. The final
`TestCountsAreRecorded` passed without updating it: allocations, frees,
retains, releases, peak live and regions are identical on every row. **No row
gets worse; no exception is needed.** Table SHA256:
`18925c2153195e928190ff5323b5ad9a9e61c5db5f60648aeee0c5bde5517ee4`.
These are allocation/RC counts, not per-fixture CPU Ir. The existing exclusion
of stack_overflow.a remains as stated in the table; no new exclusion was added.

Final checks, all output redirected to files:

- Complete `internal/native` suite: **104.249s**; complete `internal/oracle`,
  including every Node comparison, input fixture and counts row: **611.618s**.
- New cache-state check and the existing 8,796-line mixed-Unicode index sweep:
  **16.800s**, zero mismatches. The new test queries indexOf before length,
  then repeats reverse charCodeAt reads on a long literal, an unindexed stack
  value and a built heap string; expected output comes from Node.
- Complete final scanner suite, including before/final release/profiled/Node
  snapshots, sanitizer/leak checks and its three comparison mutants:
  **117.572s**, 25,014,914 answer bytes identical.
- Complete final JSON suite: **452.862s**, all **1,076** inputs identical on
  Go, Node, native and the JS backend, including both before/final profiled
  binaries. Sanitizers, 72 boundaries, driver checks, comparison mutants and
  gap refusals passed. Prettier 3.9.6 reports exactly the nine known upstream
  differences; its separate comparison and mutants passed.
- Vet, gofmt and diff whitespace checks passed with no output.

The earlier full `go test ./...` passed for the discarded cursor experiment
(oracle 849.006s). After selecting the cache-only patch, the final gate used
all touched-package tests plus the **complete** oracle and both complete port
suites. A final whole-repository ./... run was not repeated. The final new
cache-state test ran separately because it was added after the complete native
package run started. No claim of unprofiled Go/Node instruction costs or
arbitrary-input proof is made.

## Unsafe mutant

Remove `|| index == ADAMIC_LITERAL_INDEX` from both new guards. It compiles
successfully, reaches the real runtime lookup, and dereferences the literal
marker as an index. `TestStringIndexCacheStatesMatchNode` catches an **ASan
global-buffer-overflow** in units_before on the first literal indexOf, rather
than a compile warning. Node has already produced the valid expected answer.
The original source is untouched: the mutation is supplied by a Go overlay.
The exact mutation and full sanitizer diagnostic are adjacent to this report.

## Commands and reproduction

```sh
source /workspace/adamic-tools/env.sh
go test -v -count=1 -timeout 30m ./internal/native ./internal/oracle > /tmp/adamic-utf16-perf/final-core.log 2>&1
go test -v -count=1 ./internal/native -run '^TestStringIndex(CacheStatesMatchNode|MatchesNode)$' > /tmp/adamic-utf16-perf/final-index.log 2>&1
go test -overlay=/tmp/adamic-utf16-perf/cache-mutant-overlay.json -v -count=1 ./internal/native -run '^TestStringIndexCacheStatesMatchNode$' > /tmp/adamic-utf16-perf/cache-mutant.log 2>&1
go vet ./... > /tmp/adamic-utf16-perf/vet.log 2>&1
gofmt -l cmd internal > /tmp/adamic-utf16-perf/gofmt.log 2>&1
```

To rebuild profiles, merge only the two port branches into a separate worktree,
with physical, initialized cohere/TypeScript submodules. Use scanner's
TestProfileArtifacts to save main.c and its pinned compiler manifest. Save
JSON's generated C once, and prepare its manifest/expected stream using the
JSON profile.py --prepare-to command documented in its PERFORMANCE.md. Compile
each saved C twice: one runtime copy from the main baseline, one with only this
runtime patch. Actual compile/profile commands (replace PORT and its arguments):

```sh
clang -std=c11 -O2 -g -ffp-contract=off -fno-optimize-sibling-calls -Iinternal/native/runtime /tmp/adamic-utf16-perf/PORT.c internal/native/runtime/*.c -lm -o /tmp/adamic-utf16-perf/PORT-final > /tmp/adamic-utf16-perf/build.log 2>&1
VALGRIND_LIB=/tmp/adamic-json-perf/valgrind/usr/libexec/valgrind /tmp/adamic-json-perf/valgrind/usr/bin/valgrind --tool=callgrind --callgrind-out-file=/tmp/adamic-utf16-perf/PORT-final.callgrind /tmp/adamic-utf16-perf/PORT-final ARGS > /tmp/adamic-utf16-perf/PORT-final.stdout 2>/tmp/adamic-utf16-perf/PORT-final.stderr
python3 /tmp/adamic-perf-ports/stage1/cohere/json/profile.py /tmp/adamic-utf16-perf/PORT-final.callgrind > /tmp/adamic-utf16-perf/PORT-final-top.txt
```

Scanner arguments: `--manifest /tmp/adamic-utf16-perf/scanner-before/compiler.txt
--count`; JSON: `--cases /tmp/adamic-json-perf/profile-cases.txt`. Compare sample
outputs to Go, and enable ADAMIC_SCANNER_PROFILE_SNAPSHOTS and
ADAMIC_JSON_PROFILE_BINARIES for entire-protocol verification. Complete port
suite commands run in the scratch worktree with
ADAMIC_TYPESCRIPT_SOURCE=/tmp/adamic-utf16-perf/typescript and
ADAMIC_JSON_PRETTIER=/tmp/adamic-json-prettier (Prettier 3.9.6).

Setup succeeded: Go 0s, clang 0s, Node 0s, submodules 0s, cache warm **9s**,
total **9s**. `nproc` **5**, CPU quota 4, memory 17.6 GB; Go 1.27.1, clang
20.1.8, Node 24.19.0. Valgrind extracted during the previous unit was reused.
