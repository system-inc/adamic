# Inventory syntax batch 8

Implementation commit: `ca9b13a210893873d720afac3a4b188df82c025a`.
Branch: `codex/stage1-lint-batch8`, cut from origin/main
`5d4c8012a0877094134e6c6bac367ff68f9313e8`. The following evidence commit
contains this report and the raw logs. No compiler hot files were changed.

## Selection and implementation

Selection is candidates 21 through 30 of the remaining syntax-only list in
inventory commit `73ac2eb`: wave `syntax ready for AST/API adaptation`, excluding
entries already marked `ported` or `partial port`. The complete selected records
are in [selection.json](batch8_evidence/selection.json).

1. no-octal-escape
2. no-unexpected-multiline
3. no-unused-private-class-members
4. no-useless-constructor
5. prefer-template
6. react/forward-ref-uses-ref
7. react/jsx-no-comment-textnodes
8. react/no-find-dom-node
9. react/no-is-mounted
10. react/no-redundant-should-component-update

Batch 6 was checked when published at
`c5d128ab4158197a3b3a9f6452b9224e946ac462`; its selected first ten are disjoint.
Batch 7 had no published remote branch at the final check. The frozen ordinal
selection reserves its candidates 11 through 20. The final remote check is
[remote-overlap.log](batch8_evidence/remote-overlap.log).

Each rule owns a file and a single dispatch line in `batch8/registry.ts`.
The standalone ten-rule driver leaves the original five-rule driver and shared
registry untouched. It emits exact human findings, UTF-8 byte ranges, message
IDs, automatic edits, all suggestion IDs/descriptions/ordered edits, and the
converged fixed source. Parent indexes and position/line tables are built once.
Immutable discriminants distinguish structural diagnostic, suggestion and edit
records. See [runner documentation](batch8/README.md) for the protocol.

Nine listeners work on the supported TypeScript parser. JSX comment text-node
listener logic is implemented, but JSX parser integration is blocked. Forty
upstream cases use independent Go AST text boundaries, and two ordinary source
cases exercise its clean path. These are listener checks, not JSX parser ports.
Fourteen additional JSX-bearing fixtures (nine no-find-dom-node, five
no-is-mounted) are explicitly listed by source hash in
[parser_gaps.json](batch8_evidence/parser_gaps.json). The valid JSX gap probe
produces a Go finding and the same exit-70 parser refusal on Node and sanitized
native. Nothing outside this explicit ledger is silently skipped.

## Oracles and results

Cohere submodule: `715ba94f3608a6500086b1076ce5cb7e51b836db`.
TypeScript 6.0.3: `050880ce59e30b356b686bd3144efe24f875ebc8`.
Original upstream assertions run before fresh fixture capture. Capture preserves
TS/TSX/JS/JSX mode, filenames and options; exactly 714 unique cases are required.
Of these, 660 source cases plus 40 Go-AST listener cases compare byte for byte;
14 are the explicit parser gaps above. Comparison totals 12,265,076 bytes.

All 77 TypeScript compiler files and 121 stage1 TypeScript files also compare
byte for byte across Go, Node and sanitized native: 198 files, 12,356,124 bytes.
An independently compiled release binary passes the same full corpus. Fourteen
position/AST edge probes pass both native modes: Unicode byte positions, CRLF,
CR and U+2028, commented callback parameters, private writes including for-await,
constructor/default/generic boundaries, template escaping/ASI, JSX listener
boundaries and React member forms. The common comparison logger says
"sanitized native" even for the release checks; that second build explicitly
uses `Sanitize: false` and release optimization.

Every mutation has a positive control matching Go first, then a successfully
compiled Node and sanitized-native run whose output differs from Go. All twelve
mutants were caught:

| Family                               | Mutation                        | Check that caught it                   |
| ------------------------------------ | ------------------------------- | -------------------------------------- |
| no-octal-escape                      | Disable nonzero escape width    | Missing escape finding                 |
| no-unexpected-multiline              | Invert newline comparison       | Missing newline-call finding           |
| no-unused-private-class-members      | Invert read-use test            | Missing unused-private finding         |
| no-useless-constructor               | Return for every body index     | Missing constructor finding/suggestion |
| prefer-template                      | Invert literal containment      | Missing concatenation finding/fix      |
| forward-ref-uses-ref                 | Invert callback parameter count | Missing finding/suggestions            |
| jsx-no-comment-textnodes             | Invert line-start test          | Missing Go-AST text-node finding       |
| no-find-dom-node                     | Invert member name equality     | Missing quoted-member finding          |
| no-is-mounted                        | Invert member name inequality   | Missing whole-call finding             |
| no-redundant-should-component-update | Invert PureComponent test       | Missing class finding                  |
| forward-ref-uses-ref, extra          | Change second suggestion ID     | Suggestion metadata differs            |
| prefer-template, extra               | Append space to edit payload    | Edit and fixed-source bytes differ     |

The three inherited lint mutants pass too. The uncached external one-byte
oracle mutant passes with native and Node cache misses, demonstrating a fresh
outside-compiler oracle check.

## Commands, timing and throughput

Tests write to log files, not pipelines. Toolchain setup was
`bash cloud/setup.sh`; each toolchain shell sourced
`/workspace/adamic-tools/env.sh`. Go 1.27.1, clang 20.1.8, Node 24.19.0.
Setup timing lines: Go 0s, clang 0s, Node 0s, submodules 0s, build-cache warm
33s, total 33s. `nproc`: 5, cgroup quota 4 CPUs, 17.6 GB, Intel Xeon Platinum
8573C. See [setup.log](batch8_evidence/setup.log) and
[machine.log](batch8_evidence/machine.log).

```sh
export ADAMIC_LINT_BENCH=1
export ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/lint-batch2/typescript-6.0.3
TIMEFORMAT='wall %3R seconds user %3U seconds sys %3S seconds'
{ time go test ./stage1/cohere/lint -count=1 -v -timeout 30m; } > /tmp/batch8-gate-complete.log 2>&1
go vet ./... > /tmp/batch8-vet-complete.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run TestTheOracleCatchesOneByte -count=1 -v -timeout 10m > /tmp/batch8-external-oracle-complete.log 2>&1
/workspace/scratch/lint-batch2/cohere --no-fix --no-cache stage1/cohere/lint/batch8 > /tmp/batch8-cohere-gate.log 2>&1
```

Full lint package PASS: Go 428.949s; wall **437.078s (7m17s)**, user 652.362s,
sys 65.923s. This includes the inherited suite, fixture capture, sanitized and
release corpus checks, mutant builds and both throughput measurements. No
10-minute investigation threshold was crossed. Clang remains the main build
cost: ordinary sanitized batch-8 compile 23.155s, mutant compiles 18–31s,
release throughput compile 7.889s, generated C 2,281,713 bytes. The twelve
mutants run with at most four simultaneous builds. Logs:
[full lint gate](batch8_evidence/lint-gate.log), [vet](batch8_evidence/vet.log)
(empty means successful), [external oracle](batch8_evidence/external-oracle.log),
[source lint](batch8_evidence/cohere.log) (17 checked files, 100% ready).

Best of five interleaved fresh-process count runs on 77 compiler files:

| Engine         |  Seconds | Findings/s |
| -------------- | -------: | ---------: |
| Go cohere      | 0.409678 |     392.99 |
| Release native | 2.898728 |      55.54 |
| Node           | 1.723092 |      93.44 |

All three count 161 findings. All are prefer-template; the other nine families
have positive upstream/mutant controls. Per-file counts are in
[compiler_counts.json](batch8_evidence/compiler_counts.json). Measurements
include startup, reading, parsing and traversal, exclude compilation, rendering
and fix application. Native is slower than Go and Node; no speedup is claimed.
The inherited five-rule benchmark separately counts 451 findings.

## Limits and development observations

This is an isolated ten-rule runner, not integration into the original default
CLI. JSX parsing and the fourteen ledger cases remain uncovered. No binder,
configuration/suppression semantics, general malformed-source recovery or
competing-fixer/rejection protocol is claimed. Legacy-octal fixtures compare
findings and the absence of edits using Go's rule harness despite lexical
errors; when no automatic proposal exists neither oracle invokes a fix engine.
Actual fix cases use Go's converging engine and native reparsing.

Early attempts exposed wrong JS parser mode, quadratic position scans, an
incorrect mutant anchor and a concurrent source edit invalidating a running
build. These were corrected; fixed source snapshots now isolate builds and
mutants. Only the final completed gate above is evidence for the committed
implementation. A preliminary corpus run lacked the TypeScript environment
variable; the final gate sets it explicitly. Source hashes in
[sources.sha256](batch8_evidence/sources.sha256) identify the tested final files.

No full-repository `go test ./...` was run: the full touched lint package,
filtered external oracle and repository-wide `go vet` were run. No unrelated
native/lowering/oracle implementation files were edited. No pull request opened.
