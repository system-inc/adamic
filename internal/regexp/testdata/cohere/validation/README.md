# Validation observations

All commands ran after sourcing `/workspace/adamic-tools/env.sh`, which
`cloud/setup.sh` wrote using the configured `ADAMIC_TOOLS` directory.
All test output was redirected to files, then read. Nothing was piped from a
running test. These are observations from the matcher branch base
`df959ee978da7cfcac45fe0e9ff941b80a6fda58`, not claims about future matcher changes.

## Setup

`bash cloud/setup.sh > /tmp/regex-cohere-setup.log 2>&1` exited 0.
`nproc` printed `5`. The timing lines were:

```text
setup: go ready (1s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (2s)
setup: node ready (2s)
setup: submodules ready (2s)
setup: build cache warm (124s)
setup: done in 124s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

Tool versions: Go 1.27.1, clang 20.1.8, Node v24.19.0. The complete output is in
[setup.txt](setup.txt). Setup did not fail. The initial fetch brought only main;
an explicit `git fetch origin codex/regex-matcher:refs/remotes/origin/codex/regex-matcher`
made the required base available, then the branch fast-forwarded to it.

## Checks

| Command | Observed result | Log |
| --- | --- | --- |
| `gofmt -l cmd internal` | Exit 0, empty output | Empty formatting log |
| `go vet ./...` | Exit 0, empty output | Empty vet log |
| `ADAMIC_COHERE_RECORD=1 go test ./internal/regexp -run '^TestCohere' -count=1 -timeout 10m -v` | PASS, 48.897s; report and exact issues recorded | [record.txt](record.txt) |
| `go test ./internal/regexp -count=1 -timeout 10m` | PASS, 54.090s | [package.txt](package.txt) |
| `PATH=/tmp/regex-cohere-no-node:$PATH go test ./internal/regexp -run '^TestCohere' -count=1 -timeout 10m -v` | PASS, 54.768s with a Node stub that exits 99 | [no-node.txt](no-node.txt) |
| `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestNativeAgreesWithNode/internal/oracle/testdata/regexp' -count=1 -timeout 10m -v` | PASS, 12.693s; all 10 regex fixtures; cache hits 0 | [oracle.txt](oracle.txt) |
| `bash internal/regexp/testdata/cohere/regenerate.sh` | Exit 0; 875 occurrences, 17 unresolved constructors, 128,592 observations | [regenerate.txt](regenerate.txt) |
| `sha256sum -c /tmp/regex-cohere-before.sha256` | Inventory, Node corpus, selector fixture file each reproduce byte for byte | [determinism.txt](determinism.txt) |
| `node --check` on all three `.cjs` files and `bash -n` on `regenerate.sh` | Exit 0 | No output |

The full repository test gate was not run. The entire touched package, full vet
and formatting checks, and the uncached filtered regex oracle were run.
The Node-free check replaces Node with an executable that prints an error and
exits 99; passing proves no Node invocation occurs in these new CI tests.
The checks do not claim the whole regexp package runs without Node; existing
independent oracle tests invoke it.

Observed corpus totals: 13,425 source files, 875 static occurrences (868 literals,
7 constructors; 528 distinct pattern/flag pairs), 77 implementation/script
occurrences, 798 fixture/test occurrences, 17 unresolved constructors, 13
constant call-site inputs, and 62 traced calls from 171 selector fixtures.
There are 44,938 distinct-per-pattern input strings, 128,592 recorded calls,
128,542 completed Go comparisons and 385,626 completed native comparisons.
Node rejects 238 occurrences: 97 pattern/flag errors and 141 source-only lexical
errors recovered by the TypeScript parser. One Node-valid pattern is refused by
Adamic and another disagrees on four inputs in Go and each native mode. The
50 recorded calls for the refused valid pattern cannot execute in Adamic.
The report records the refusal; they are not called successful comparisons.

ASan, UBSan and LeakSanitizer passed in all native runs. There are 876 recorded
named-group match observations. All 243 failure/report rows are pinned exactly;
known failures remain visible, rather than being accepted as correct behavior.

## Mutants

`python3 internal/regexp/testdata/cohere/run-mutants.py /tmp/regex-cohere-mutants-final`
exited 0 because all six deliberately faulty tests exited 1 for the intended
reason. Every file was restored. [The runner's output](mutants.txt) and each
actual failing test log are retained:

| Mutant | The check that caught it | Evidence |
| --- | --- | --- |
| Flip expected participating capture 1 end at pattern 1, case 294 | Full Go/native comparison reports `capture indices` and an unexpected issue row | [expected-capture.txt](expected-capture.txt) |
| Remove named span/value dictionary comparison | `TestCohereNamedGroupGuard`, using pattern 425, case 92800 | [omit-named-comparison.txt](omit-named-comparison.txt) |
| Remove match index comparison | The index-only recorded-result guard | [omit-index-comparison.txt](omit-index-comparison.txt) |
| Remove capture value comparison | The values-only recorded-result guard | [omit-value-comparison.txt](omit-value-comparison.txt) |
| Remove lastIndex comparison | The lastIndex-only recorded-result guard | [omit-lastindex-comparison.txt](omit-lastindex-comparison.txt) |
| Remove only named capture value comparison, retaining span comparison | The named-values-only recorded-result guard | [omit-named-value-comparison.txt](omit-named-value-comparison.txt) |

None failed at Go compilation, clang, a sanitizer, or an unrelated check. The
mutations changed the recorded oracle or this new comparator, never the owned
matcher/compiler/runtime files. The real corpus remains unchanged afterward.

## Uncovered

Unresolved constructors and their locations are in the generated report. The
extractor does not follow constructor aliases or arbitrary imported/computed
constant expressions. Actual fixture tracing is limited to direct constant
TypeScript call relationships and the selector port, not all cohere execution.
Generated inputs sample finite alphabets and bounded witnesses, not all strings.
The exec corpus does not exercise boolean-only fast paths or method-specific
replace/split/matchAll semantics. No compiler or runtime gap was fixed.
