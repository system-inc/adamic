# pw720wc: TypeScript scanner slice

This is the initial unit report. The subsequent profiling and optimization
report is [PERFORMANCE.md](PERFORMANCE.md).

Branch: `codex/typescript-scanner`. Starting main: `fe3b9f2`.
All implementation, tests, gap probes and notices are confined to
`stage1/typescript/scanner/`. No core compiler files change.

Built: skip-trivia token scanner, identifiers and Unicode escapes, strings,
templates, numeric and bigint literal values, punctuators, lexical errors,
JSDoc preceding flags, shebang/conflict markers, greater/template/regex
rescans and JSX text scanning. `GAPS.md` defines the answer protocol,
representation choices, the observed push and bigint gaps and uncovered APIs.
The oracle is the original Go scanner compiled through an overlay; the
submodule is not modified. TypeScript v6.0.3 lives only in scratch.

## Setup

```sh
git fetch origin && git checkout -b codex/typescript-scanner origin/main
bash cloud/setup.sh > /tmp/typescript-scanner-setup.log 2>&1
source /workspace/adamic-tools/env.sh
nproc
```

Setup succeeded. Its timing lines: Go 0s, clang 0s, Node 0s,
submodules 1s, cache warm 14s, done 14s. `nproc`: 5.
The printed environment path is `/workspace/adamic-tools/env.sh`,
not `/opt/adamic-tools/env.sh`. Tools: Go 1.27.1, clang 20.1.8,
Node 24.19.0.

Pinned corpus clone:

```sh
git clone --depth 1 --branch v6.0.3 https://github.com/microsoft/TypeScript.git /workspace/scratch/typescript-6.0.3 > /tmp/typescript-scanner-clone.log 2>&1
```

The test verifies commit `050880ce59e30b356b686bd3144efe24f875ebc8`.
The Go submodule commit is `8d550c837c90bd1805b047b7eeccc2baac2d5e7a`.

## Validation commands

All test stdout/stderr went to log files, never through a pipe.
The environment above was sourced before these commands.

```sh
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-6.0.3 go test -count=1 -timeout 30m ./... > /tmp/typescript-scanner-gate.log 2>&1
go vet ./... > /tmp/typescript-scanner-vet.log 2>&1
gofmt -l cmd internal stage1/typescript/scanner > /tmp/typescript-scanner-gofmt.log
/workspace/scratch/cohere --no-fix --no-cache stage1/typescript/scanner/main.ts stage1/typescript/scanner/scanner.ts stage1/typescript/scanner/characters.ts stage1/typescript/scanner/tokens.ts > /tmp/typescript-scanner-cohere.log 2>&1
```

`go vet` and gofmt: exit 0, empty output. Cohere: exit 0,
276 rules, 4 checked files, 100% Adamic-ready. The deliberately refused
`gaps/1_push.ts` is excluded by the repository's established gaps rule.

Full gate: exit 0, all packages passed. The oracle took 670.699s and
this slice took 122.502s under concurrent package load. After adding the
second gap probe, the affected package was rerun in full, and its vet and
format checks repeated. No production implementation changed after the
full gate. Final affected-package run: exit 0, 61.792s, 77 compiler files,
61 stage1 files, 18,224 generated inputs, 23,832,564 identical answer bytes.

```sh
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-6.0.3 go test -count=1 -v -timeout 30m ./stage1/typescript/scanner > /tmp/typescript-scanner-test.log 2>&1
go vet ./stage1/typescript/scanner > /tmp/typescript-scanner-vet-final.log 2>&1
```

Final test output:

```text
=== RUN   TestGapStandsWhereGapsMdSays
--- PASS: TestGapStandsWhereGapsMdSays (0.28s)
=== RUN   TestBigintGapStandsWhereGapsMdSays
--- PASS: TestBigintGapStandsWhereGapsMdSays (0.14s)
=== RUN   TestScannerAgreesWithTypescriptGo
    scanner_test.go:339: 77 compiler files, 61 stage1 files, 18224 generated inputs
    scanner_test.go:356: 23832564 answer bytes identical
=== RUN   TestScannerAgreesWithTypescriptGo/punctuator_!=_scanned_as_==
    scanner_test.go:379: Node caught by byte comparison: line 348: port "EqualsEqualsToken 0 2 0", Go "ExclamationEqualsToken 0 2 0"
    scanner_test.go:379: native caught by byte comparison: line 348: port "EqualsEqualsToken 0 2 0", Go "ExclamationEqualsToken 0 2 0"
=== RUN   TestScannerAgreesWithTypescriptGo/invalid_decimal_separator_accepted
    scanner_test.go:379: Node caught by byte comparison: line 689: port "NumericLiteral 0 4 512\t12", Go "error 6189 2 1"
    scanner_test.go:379: native caught by byte comparison: line 689: port "NumericLiteral 0 4 512\t12", Go "error 6189 2 1"
=== RUN   TestScannerAgreesWithTypescriptGo/regex_rescan_skipped
    scanner_test.go:379: Node caught by byte comparison: line 126195: port "SlashToken 0 1 0", Go "RegularExpressionLiteral 0 6 0\t/abc/g"
    scanner_test.go:379: native caught by byte comparison: line 126195: port "SlashToken 0 1 0", Go "RegularExpressionLiteral 0 6 0\t/abc/g"
--- PASS: TestScannerAgreesWithTypescriptGo (61.36s)
    --- PASS: TestScannerAgreesWithTypescriptGo/punctuator_!=_scanned_as_== (10.66s)
    --- PASS: TestScannerAgreesWithTypescriptGo/invalid_decimal_separator_accepted (12.43s)
    --- PASS: TestScannerAgreesWithTypescriptGo/regex_rescan_skipped (11.55s)
=== RUN   TestPerformance
    scanner_test.go:388: set ADAMIC_SCANNER_BENCH=1 for best-of-five throughput
--- SKIP: TestPerformance (0.00s)
PASS
ok  	github.com/system-inc/adamic/stage1/typescript/scanner	61.792s
```

Each mutant compiled and ran successfully under both Node and sanitized
native. The original Go answers killed each by byte comparison:

| Mutant | First disagreement on both runtimes |
| --- | --- |
| `!=` scanned as `==` | `EqualsEqualsToken` vs `ExclamationEqualsToken` |
| repeated decimal separator accepted | NumericLiteral flags 512/value 12 vs diagnostic 6189 at byte 2, length 1 |
| regex rescanning skipped | SlashToken vs RegularExpressionLiteral `/abc/g` |

The bigint probe first failed at stage 0 with
`a value of type 1237940039285380274899124223n`; its exact refusal and
Node's decimal result are required by the gap test. The first port build
also refused multi-argument push, which the other gap test requires.
During development the comparison caught an EOF punctuator overrun from
clamped slices; the implementation now checks the requested slice length.
One development run raced formatting against the corpus reads and failed;
subsequent final runs used stable sources. A mechanical property rename
was rejected by type checking and fixed before the successful full gate.

## Throughput

```sh
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-6.0.3 ADAMIC_SCANNER_BENCH=1 go test -count=1 -v -run '^TestPerformance$' -timeout 30m ./stage1/typescript/scanner > /tmp/typescript-scanner-bench.log 2>&1
```

Machine: Linux x86_64 under KVM, AMD EPYC 9V74 80-Core Processor,
5 exposed logical CPUs, cgroup `cpu.max: 400000 100000` (4 CPU quota),
17.6 GB memory. UTC start: 2026-10-06 08:19:36. Full gate had exited;
no other worker workloads were observed in the container during timing. Load average before:
`4.16 6.90 5.52`, 1 runnable; after: `3.68 6.71 5.47`, 1 runnable.
The averages retain the preceding gate's load.

The 77 compiler files produce 434,790 tokens including EOF tokens.
All three drivers returned that same count in every timed run. Five
rounds are interleaved native, Go, Node, with a fresh process per round.
Timing includes process startup, source reads and the port's UTF-16-to-byte
mapping. Printing tokens is disabled, but token values and flags are still
computed. Native is the stage 0 release build (`-O2`, sanitizers disabled);
correctness runs use sanitizers. Go uses the normal `go build`; Node runs
the same `.ts` through the repository's Node shim. Builds, cloning and
corpus generation are outside timed runs.

| Driver | Best seconds | Tokens/s |
| --- | ---: | ---: |
| Native Adamic | 0.717630 | 605,869 |
| typescript-go | 0.034183 | 12,719,384 |
| Node | 0.460514 | 944,141 |

Observation: native is about 21 times slower than Go and 1.56 times slower
than Node for this driver. This is a baseline port, not a speedup claim.
The causes were not profiled in this unit.

Raw rounds:

```text
=== RUN   TestPerformance
    scanner_test.go:390: 77 compiler files, 61 stage1 files, 18224 generated inputs
    scanner_test.go:416: round 1 native 0.717630s
    scanner_test.go:416: round 1 Go 0.034183s
    scanner_test.go:416: round 1 Node 0.471054s
    scanner_test.go:416: round 2 native 0.727175s
    scanner_test.go:416: round 2 Go 0.037502s
    scanner_test.go:416: round 2 Node 0.460514s
    scanner_test.go:416: round 3 native 0.727907s
    scanner_test.go:416: round 3 Go 0.034473s
    scanner_test.go:416: round 3 Node 0.494319s
    scanner_test.go:416: round 4 native 0.744479s
    scanner_test.go:416: round 4 Go 0.034953s
    scanner_test.go:416: round 4 Node 0.479617s
    scanner_test.go:416: round 5 native 0.774765s
    scanner_test.go:416: round 5 Go 0.040974s
    scanner_test.go:416: round 5 Node 0.489038s
    scanner_test.go:420: native best of 5: 434790 tokens / 0.717630s = 605869 tokens/s
    scanner_test.go:420: Go best of 5: 434790 tokens / 0.034183s = 12719384 tokens/s
    scanner_test.go:420: Node best of 5: 434790 tokens / 0.460514s = 944141 tokens/s
--- PASS: TestPerformance (9.65s)
PASS
ok  	github.com/system-inc/adamic/stage1/typescript/scanner	9.651s
```

## Full gate output

```text
?   	github.com/system-inc/adamic/bench	[no test files]
?   	github.com/system-inc/adamic/cmd/adamic	[no test files]
?   	github.com/system-inc/adamic/cmd/adamic-fuzz	[no test files]
ok  	github.com/system-inc/adamic/internal/flow	148.257s
ok  	github.com/system-inc/adamic/internal/fresh	36.942s
ok  	github.com/system-inc/adamic/internal/fuzz	15.970s
?   	github.com/system-inc/adamic/internal/ir	[no test files]
?   	github.com/system-inc/adamic/internal/javascript	[no test files]
ok  	github.com/system-inc/adamic/internal/load	1.353s
ok  	github.com/system-inc/adamic/internal/lower	7.355s
ok  	github.com/system-inc/adamic/internal/native	252.502s
ok  	github.com/system-inc/adamic/internal/oracle	670.699s
ok  	github.com/system-inc/adamic/stage1/cohere/formatfiles	108.981s
ok  	github.com/system-inc/adamic/stage1/cohere/gitignore	168.208s
ok  	github.com/system-inc/adamic/stage1/cohere/graphql	179.111s
ok  	github.com/system-inc/adamic/stage1/cohere/mediaquery	92.468s
ok  	github.com/system-inc/adamic/stage1/cohere/suppression	120.735s
ok  	github.com/system-inc/adamic/stage1/cohere/values	111.671s
ok  	github.com/system-inc/adamic/stage1/typescript/scanner	122.502s
```

## Uncovered

See GAPS.md for the explicit API boundary. In particular, regex scanning
covers `ReScanSlashToken(false)` and unterminated recovery, not the separate
regexp syntax validator. No parser context inference, alternate scanner
language settings, trivia-returning scanner, tagged continuation, JSDoc
specialized scanner, state/lookahead API, or malformed UTF-8 corpus is
claimed. No PR was opened.
