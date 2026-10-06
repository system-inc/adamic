# Regex speed benchmark

[RESULTS.md](RESULTS.md) reports before/after wall time and Callgrind instructions for every pattern. [results.csv](results.csv) preserves exact timings and checksums. Baseline is `59d82c5de172e4e53d1224553b319d076f22424a`; optimized runtime is this commit.

## Implementation

Search metadata is derived from bytecode: mandatory ASCII literal prefixes, conservative ASCII first-character masks, and non-multiline start anchors. Native prefix scanning uses `memchr`. Unknown or non-ASCII cases retain the VM. Small forward patterns emit straight-line C; an ASCII specialization supports canonicalized classes without UTF-16 conversion when the input's cached ASCII flag permits it. Production's unlimited execution uses these callbacks; budgeted execution retains the VM and fails loudly on exhaustion.

Choice points reuse aligned capture/repetition frames in an 8 KiB execution-local arena, with a reusable heap overflow pool. Frames preserve independent snapshots; capture resets and lazy branch ordering remain observable. Common initial and lookaround registers use stack storage. Singleton sets avoid temporary position arrays, UTF-16 conversion is linear, and ASCII canonicalization avoids table searches. The Go reference VM also skips impossible starts and avoids redundant repeat snapshots.

## Reproduction

Run from the repository root, with Node 24, Go, clang and Valgrind available:

```sh
source /workspace/adamic-tools/env.sh
export REGEXP_CALLGRIND_HEADER=/usr/include/valgrind/callgrind.h
go run ./bench/regex -out /tmp/regex-after -build-only > /tmp/regex-build.log 2>&1
python3 bench/regex/measure.py /tmp/regex-before /tmp/regex-after > /tmp/regex-times.csv 2> /tmp/regex-times.log
python3 bench/regex/count.py /tmp/regex-before > /tmp/regex-before-ir.csv 2> /tmp/regex-before-ir.log
python3 bench/regex/count.py /tmp/regex-after > /tmp/regex-after-ir.csv 2> /tmp/regex-after-ir.log
```

Build the baseline with the same driver and emitted bytecode, the runtime `.c` files from baseline commit, and the current descriptor header. The extra descriptor fields are ignored by the original interpreter. Both binaries use `-std=c11 -O2 -ffp-contract=off -fno-optimize-sibling-calls` and link `-lm`. `measure.py` rejects unequal inputs or checksums and rotates engine order in five rounds. `count.py` rejects missing or zero counts. `CALLGRIND` and `VALGRIND_LIB` can select a locally extracted Valgrind installation.

This worker's system package installation was unavailable because system directories were read-only; Valgrind 3.24.0 was downloaded and extracted under `/workspace/scratch/regex-speed-tools`, without changing system directories. Toolchain setup timings were Go ready 0s, clang ready 1s, Node ready 1s, submodules ready 1s, build cache warm 24s, total 24s; `nproc` is 5, with four cores of CPU quota.

## Validation and outputs

All test commands redirected output to log files. Final commands and observed results:

```sh
go test ./internal/native ./internal/regexp ./bench/regex -count=1 -timeout 15m -v > /tmp/regex-speed-gate-packages.log 2>&1
# PASS: native 208.153s; regexp 5.628s; benchmark has no test files.
go test ./internal/oracle -run 'TestNativeAgreesWithNode/.*/.*/.*/(regexp|sweeps)' -count=1 -timeout 15m -v > /tmp/regex-speed-oracle.log 2>&1
# PASS 75.457s: regex fixtures and sweeps, including 63,960 regex method probes.
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 15m > /tmp/regex-speed-counts-gate.log 2>&1
# PASS 361.332s; recorded allocation counts unchanged.
go vet ./... > /tmp/regex-speed-vet.log 2>&1
# exit 0.
gofmt -l cmd internal bench/regex > /tmp/regex-speed-gofmt.log
# empty output.
git diff --check > /tmp/regex-speed-diff-check.log 2>&1
# exit 0.
```

Go: all 127,369 test262 execution cases and 10,000 fixed-seed randomized cases agreed with Node, with no unavailable properties. Native: those same totals agreed in BOTH budgeted VM and unlimited optimized modes, checking both `test` and `exec`, captures, indices, groups and lastIndex. The corpus therefore exercised 509,476 native API operations. The 102 benchmark rows supplied another 890 inputs per mode; 29 search controls per mode cover surrogate-boundary anchors, prefix overlap, alternatives, nullable paths, folding, and workspace overflow. Native tests use ASan, UBSan and leak detection. Catastrophic backtracking still triggers the test step-limit failure rather than becoming a failed match.

The full repository test gate was not run. Touched packages, the exact filtered oracle above, the full allocation-count gate and repository-wide vet were run. Exhaustive unchanged Unicode property tables were not revalidated over every code point in this performance step.

## Mutants

Run the existing Go mutant driver, the existing native driver, and the new speed driver from the repository root; each restores its source edits. Final logs: `/tmp/regex-speed-go-mutants-final.log`, `/tmp/regex-speed-native-mutants-final.log`, `/tmp/regex-speed-mutants-release.log`.

Commands used for these 20 mutations:

```sh
python3 internal/regexp/testdata/run-mutants.py > /tmp/regex-speed-go-mutants-final.log 2>&1
python3 internal/native/testdata/run-regexp-mutants.py native-greedy-as-lazy native-captures-not-reset native-lookbehind-left-to-right native-case-fold-without-u-distinction > /tmp/regex-speed-native-mutants-final.log 2>&1
python3 internal/native/testdata/run-regexp-speed-mutants.py > /tmp/regex-speed-mutants-release.log 2>&1
```

All 20 applied mutants were caught:

| Mutant | Witness |
|---|---|
| Go greedy as lazy | Node capture disagreement |
| Go repeated captures not reset | `(a|(b))+` capture disagreement |
| Go lookbehind left to right | Node capture disagreement |
| Go folding without u distinction | Kelvin sign legacy disagreement |
| Go property strings not snapshotted | provider snapshot disagreement |
| Go step limit becomes failed match | catastrophic-backtracking control |
| Go unavailable property becomes empty | explicit missing-property control |
| Native greedy as lazy | Node execution disagreement |
| Native repeated captures not reset | Node execution disagreement |
| Native lookbehind left to right | Node execution disagreement |
| Native folding without u distinction | Node execution disagreement |
| ASCII specialization skips canonicalization | benchmark Node disagreement |
| Anchor checked before surrogate rewind | search Node disagreement |
| Workspace heap block leaked | LeakSanitizer |
| Anchor ignores multiline | search Node disagreement |
| Unicode pair truncated | search Node disagreement |
| First mask drops an alternative | search Node disagreement |
| Prefix changes a literal | search Node disagreement |
| Straight-line capture offset changed | search Node disagreement |
| Saved captures not copied | search Node disagreement |

## Limits

The inventory accounts for 86 static production Go calls and six representative dynamic instantiations. Dynamic vocabulary is not exhaustive. Six JavaScript originals were independently fetched and audited; [original-sources.json](original-sources.json) records source URLs and content hashes. Other rows are Go-to-ECMAScript syntax translations, without a claim of identical RE2 behavior over all Unicode inputs.

Measurements cover warmed `test` loops on this synthetic input mix, not end-to-end lint throughput, cold compilation or performance of every string method. Native still trails Node on all four hard cases, by approximately 8.4x to 53.7x. General backtracking remains the VM; non-ASCII searches conservatively fall back, and overflow or string-valued set paths can still allocate. These limits are visible in the per-pattern results.
