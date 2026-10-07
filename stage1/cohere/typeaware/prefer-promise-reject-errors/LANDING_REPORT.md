Rebased the existing wave 20 branch and re-greened all nine implemented rules.
Validated source commit: `e3c8e71688862aa014b539cde8b4d79560e22e90`; main base: `e8ba3d5d81de4d3773c723914fccd4c76248b965`.
Commands: all three rule gates PASS, checker PASS 0.411s, vet empty, filtered Node PASS 1.592s.
Mutants: nine rule decisions/repairs, five fact answers, five registry retentions, and Node one-byte check caught.
Not covered: three unfinished React claims, full repository gate, own-rule emitted JavaScript comparison.

Only one branch was pushed by this worker: `codex/typeaware-wave-20`.
The user explicitly required a rebase onto current origin/main before more work.
This takes precedence over CLAUDE.md's usual prohibition on rewriting history;
the publication uses an exact lease on this worker's previously pushed remote
tip `cb87c3c48e9146a919a90ba266d96f4444291e8f`. No other branch is rewritten.
The final report/evidence commit contains no implementation changes.

The first rebase, onto `e011f8f6`, also passed all three rule gates. During
validation main advanced to `e8ba3d5d` with devirtualization and runtime call-target
changes. Because those affect native code, a second conflict-free rebase and
complete fresh rule validation were performed. Only the second run is the
landing evidence. No rule, bridge, shared harness, generator, parser, compiler,
or oracle source needed a repair during either rebase.

## Reproduction

All test output was redirected to files. Toolchain setup succeeded: Go ready
0s, clang ready 1s, Node ready 1s, submodules ready 1s, build cache warm 164s,
done 164s. `nproc` is 5; quota is four cores. Go 1.27.1, clang 20.1.8,
Node 24.19.0. Commands source `/workspace/adamic-tools/env.sh`.

```sh
ADAMIC_WAVE20_ARTIFACTS=/workspace/wave20-validation/landing-current-first \
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave20-typescript \
go test ./stage1/cohere/typeaware -run '^TestWave20AgreementAndMutants$' -count=1 -v -timeout 30m > /workspace/wave20-validation/landing-current-first.log 2>&1

python3 stage1/cohere/typeaware/no-process-exit-after-output/verify.py \
 --repository /workspace/adamic --artifacts /workspace/wave20-validation/landing-current-second \
 --compiler /workspace/wave20-typescript --cases /workspace/wave20-validation/next/restored \
 > /workspace/wave20-validation/landing-current-second.log 2>&1

python3 stage1/cohere/typeaware/prefer-promise-reject-errors/verify.py \
 --repository /workspace/adamic --artifacts /workspace/wave20-validation/landing-current-third \
 --compiler /workspace/wave20-typescript --cases /workspace/wave20-validation/third/valid-cases \
 > /workspace/wave20-validation/landing-current-third.log 2>&1

go test ./bridge/tsgo/checker -count=1 -v > /workspace/wave20-validation/landing-current-checker.log 2>&1
go vet ./... > /workspace/wave20-validation/landing-current-vet.log 2>&1
go test ./internal/oracle -run 'TestTheOracleCatchesOneByte|TestNativeAgreesWithNode/internal/oracle/testdata/(strings|closures|generic_functions)[.]a$' -count=1 -v -timeout 5m > /workspace/wave20-validation/landing-current-node.log 2>&1
```

First batch: PASS 159.214s; controls have 48 findings and 13,072 identical
bytes in normal and sanitizer runs. Second batch: PASS all 121 projects with
160 findings. Third batch: PASS all 682 projects (662 valid upstream and 20
added controls), with 502 findings and 365 suggestions. The pinned invalid
TSX fixture exclusion remains as documented in REPORT.md.

Each batch separately compares both pinned populations: the frozen 287-root
repository manifest and TypeScript v6.0.3's 77 compiler roots. Each comparison
has zero findings for these batches, 18,485 repository bytes and 5,241 compiler
bytes, equal in Go, native, and ASan/UBSan/LSan builds. Positive controls prevent
zero-finding corpora from making agreement vacuous. Complete messages, byte
ranges, fixes and suggestions are compared, including duplicates and every
suggestion repair. Native normal and sanitizer stderr is empty on these runs.

Each rule mutant compiles and exits normally with changed bytes; compiler or
sanitizer failure is not counted as a kill. First-batch rule differences occur
at bytes 70, 6,477 and 7,768; its accessed-property fact mutation at byte 470.
Second-batch mutation table and third-batch table are in the command archive.
The regex mutation changes an edit endpoint and is caught at byte 615.
Released handles panic 70; all five registry-retention probes instead exit 0
and are rejected by the required-panic assertions. Node's own one-byte check
also passes. Node gate reports native cache hits 7, misses 15, and Node cache
hits 0, misses 15; these cache observations are preserved rather than claiming
all selected fixture compilations were uncached.

## Native time compared with Go

Whole-process seconds, complete diagnostic output, quiet alternating medians
for the second and third batches. First batch records a single paired run.
Compilation and sanitizers are outside these intervals.

| Batch | Corpus | Native seconds | Go seconds |
| --- | --- | ---: | ---: |
| First | Repository | 0.738187 | 0.215048 |
| First | Compiler | 4.668412 | 0.805892 |
| Second | Repository | 0.417116 | 0.265605 |
| Second | Compiler | 2.120235 | 0.415590 |
| Third | Repository | 0.366432 | 0.217889 |
| Third | Compiler | 2.224747 | 0.469012 |

Native remains slower. Measurements and output hashes are preserved in
`validation/landing-current/bench.json`; these are observations, not a
performance improvement claim. The setup build completed before measurements.

## Outstanding scope

The three React claims remain unfinished: set-state-in-effect,
set-state-in-render and static-components. Native JSX support exists on
origin/codex/stage1-jsx-lint at `e715ef4a`, but is not in this main base.
Applying it changes shared parser/scanner files forbidden by Ahra's worker
scope instruction. React HIR lowering and value/control analyses are additional
prerequisites. Their BLOCKED.md files describe the prior parser probe and
upstream dependency. None of these three is represented as implemented or green.
No additional rules were claimed during this landing unit.

The full `go test ./...` gate and emitted-JavaScript comparison of these own
rule runners were not run. Required targeted rule, checker, vet and filtered
Node checks are the evidence. Source fixtures, reproduction drivers and prior
portable input tapes remain committed. Fresh process stdout/stderr, mutant
results, manifests, timings and logs are archived in
`validation/landing-current/command-logs.tar.gz`; binaries are reproducible and
are not committed.
