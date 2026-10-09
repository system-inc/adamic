# Step 06e: tsc Program region witness

Base: `23a4f2f018f1a082ef8acff06dc90ba74d40a5bd`. Branch: `runtime/step06e-tsc`.
Observed on Linux x86-64, October 9, 2026, approximately 07:19–07:25 UTC.
No compiler, runtime, selector, or fixture source changed. K60 and K144 remain counted.

## Result and limits

Host fixture 14 (`stage3/fixtures/host/14_getCurrentDirectory.a`) builds with
Program regions enabled, prints exactly Node's `true\ntrue\n`, and exits cleanly
under ASan/UBSan and leak detection. Its generated C has four Program adoption
sites: the callback cell, value cell, returned memoization closure, and input
callback closure. All four allocations execute once. Its off-mode build is
refused at line 13, column 28: the captured callback can form a strong cycle.
There is no off-mode binary or count line; this is a lowering refusal, not a
runtime failure. The enabled witness has no observed early release or reuse.

The scout's exact parse-only driver builds and runs with regions enabled and
disabled. **It adopts zero Program members.** Thus this proves compatibility and
ordinary counted ownership, not Program ownership of its acyclic parse tree or
of tsc's complete CLI core graph. It does not justify marking every ParseNode as
a member. No attempt was made to change the selector or expand this workload into
the full compiler.

## Counts: host fixture 14

| Mode | Members adopted | Allocations | Frees | Peak live values | Region values at exit |
|---|---:|---:|---:|---:|---:|
| Program on | 4 | 11 | 7 | 8 | 4 |
| Program off | unavailable: lowering refusal | unavailable | unavailable | unavailable | unavailable |

Enabled count line:

```
adamic: counts: allocations 11 frees 7 retains 18 releases 34 peak 8 regions 4
```

## Counts: 77-file parse probe

| Mode | Members adopted | Allocations | Frees | Peak live values | Region values at exit | Elapsed / user seconds |
|---|---:|---:|---:|---:|---:|---|
| Program on | 0 | 3,375,891 | 3,375,891 | 729,760 | 0 | 3.983 / 3.692 |
| Program off | 0 | 3,375,891 | 3,375,891 | 729,760 | 0 | 3.910 / 3.669 |

Both count lines:

```
adamic: counts: allocations 3375891 frees 3375891 retains 41427189 releases 37478796 peak 729760 regions 0
```

Each balance is allocations = frees + regions. `regions` is the count of values
reclaimed by region teardown, not number of region arenas; `peak` counts live
values, not bytes. ASan/UBSan and leak detection emitted no diagnostics. Zero
members makes the parse's early-member-release/reuse claim vacuous; the host
actually exercises members. Timings are single measurements, not speed claims.
The full parse is below 60 seconds and every run/build had a 90-second hard kill.

The count-mode stdout is `0\n`, identical to Node in both modes. Enabled AST mode
also matches Node byte-for-byte: 44,766,682 bytes, SHA256
`5d77733457ef9e17b5707db65a3f5af230621ca74a10c8f3daa83720504346de`.
AST mode is separate from the parse-only counts above.

## Source, toolchain and flags

The driver is absent from this base. Following `8971561e:docs/scout-36-speed.md`,
extract `stage1/cohere/parse`, `stage1/typescript/parser`,
`stage1/typescript/scanner`, and `stage1/profiles/benchmarks.json` from
`9e062aac7a55e117199c1cc5fbb7a0d59bedf07a` into scratch. The input is the 77 files
listed in that manifest from TypeScript `050880ce59e30b356b686bd3144efe24f875ebc8`;
all 77 SHA256 hashes were verified before the run. The manifest used sorted
absolute paths, one per line. No extracted source is a repository change.

Setup: `ADAMIC_TOOLS=/workspace/adamic-tools bash cloud/setup.sh` ran as a
background job and was monitored; Go ready 0.024s, Node ready 0.025s, markdown
ready 0.076s, submodules ready 0.079s, clang ready 0.195s, Go build ready 36.916s,
finished 37.158s. `nproc` = 5, CPU quota = 4. Go 1.27.1, Node 24.19.0,
clang 20.1.8. Environment: `/workspace/adamic-tools/env.sh`.

All reported native counts use **sanitized -O1**, not release -O2. Exact common
clang flags (from `native.Flags` and the observed parse clang process):

```
-std=c11 -Wall -Wextra -Werror -Wcast-function-type-strict -pedantic
-Wno-unused-variable -Wno-unused-but-set-variable -Wno-unused-function
-Wno-unused-parameter -Wno-self-assign -ffp-contract=off
-fno-optimize-sibling-calls -pthread -DADAMIC_COUNT -O1 -g
-fsanitize=address,undefined -fno-sanitize-recover=all
```

Enabled builds add `-DADAMIC_PROGRAM_REGION`. Build suffix:
`-I <runtime-cache> -o <binary> <generated-main.c> -Xlinker --whole-archive
<runtime-cache>/runtime.a -Xlinker --no-whole-archive -lm`.
`ASAN_OPTIONS=detect_leaks=1` was set on native runs. No release-performance
comparison is claimed.

## Reproduction and fault proof

Build CLI at the base, then for each source:

```sh
timeout --kill-after=2s 90s go build -o /tmp/adamic ./cmd/adamic
ADAMIC_PROGRAM_REGION=1 timeout --kill-after=2s 90s /tmp/adamic build "$source" -o "$binary" --count --sanitize
ASAN_OPTIONS=detect_leaks=1 timeout --kill-after=2s 90s "$binary" # host
ASAN_OPTIONS=detect_leaks=1 timeout --kill-after=2s 90s "$binary" --manifest compiler.txt --count # parse
# Repeat build/run with ADAMIC_PROGRAM_REGION=0 for controls.
timeout --kill-after=2s 90s node --disable-warning=ExperimentalWarning oracle/node.mjs "$source" # host
# Parse Node receives --manifest compiler.txt --count, or --ast for full AST comparison.
```

`cmp` passed for host stdout, both parse summary outputs, and the full enabled
parse AST versus Node. Existing fault proof rerun:

```sh
timeout --kill-after=2s 90s go test ./internal/native -run '^TestProgramRegionCoreMutants/member-release-frees$' -count=1 -timeout=90s -v
```

PASS, 0.271s: the planted member release that frees it is caught by ASan
`heap-use-after-free`. This branch adds measurements only; no new runtime fix
or untested membership inference was introduced. Raw output is in `evidence.txt`.
No whole-package gate or counts-table regeneration was needed for this report-only
change. Neither witness failed with Program mode enabled.
