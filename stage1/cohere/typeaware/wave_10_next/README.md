# Wave 10 continuation: partial, blocked

Prepared owned `.a` code and a raw compiler fact adapter for the three continuation claims.
Claim commit: `0ac7b51b2b670c1d2a0e6f9ea3c606eefcf92251`; partial work is in the commit containing this report.
Raw fact tests pass; five production `.a` files have zero configured lint findings; timeout-only native builds pass.
The default-library fact mutant fails; positive timeout execution refuses the unregistered question with exit 70.
None of the three continuation rules is complete; full findings/fixes/suggestions comparison, per-rule mutants and timings remain unverified.

## Selection

The previous three rules are complete and pushed in `8cfa5d02`. All origin heads
were fetched before this continuation claim. Selection used VOLUME_REPORT.md's
197 checker-dependent names in the combined validation-volume compiler/repository
count tables. The production source on origin/main and origin/codex/tsgo-c-library
contains 25 ranked ports (the 26th existing port is not checker-dependent).
Distinct Markdown claim blobs on all origin branches mention 96 other ranked
names. The first three of the 76 remaining candidates were claimed and pushed
before implementation; all have combined volume zero:

- nexus/correctness-no-process-exit-after-output
- nexus/correctness-no-uncleared-race-timeout
- nexus/correctness-require-blocking-standard-streams

No further rules were claimed. Original setup was successful in 88s, nproc 5,
with a four-CPU quota; the same tools environment was reused.

## Prepared code, not completed ports

| Rule | Prepared native code | Still missing |
| --- | --- | --- |
| no-process-exit-after-output | Process-member/global-console classification; direct write/exit identification; minimal write-chain and catch-state algebra | Native CFG construction, root analysis, callee following, complete rule runner and byte oracle |
| no-uncleared-race-timeout | Race/Promise qualification, executor scan, lost-handle classification and local binding reads, production message/spans | Live fact dispatch and findings validation; no equivalence claim |
| require-blocking-standard-streams | Program import index, type-only filtering, computed import seeds, reverse reachability and entry classification | Native CFG and load-time/call analysis, complete rule runner and byte oracle |

`runtime_context.go` provides raw declaration ancestry (including global
augmentation and name kind), declaration/source flags and text, resolved signature
declarations/return flags, and program files/import-resolution observations. It
returns no lint verdict, message, CFG, or fix. Its companion `.a` decodes the
frames and makes native classification decisions. The adapter takes the bridge
caller's already-leased checker, avoiding an additional checker lease. Its own
tests verify actual encoded fields, mode/suffix guards and missing anchors.

## Exact blockers

The current live bridge has no `runtime-context` dispatch arm. A positive timeout
source reaches the first `origin` question and exits 70, with exactly:

```text
adamic: panic: unsupported checker question: runtime-context
origin
```

The timeout-only runner compiles normally and with ASan/UBSan/LSan. Both executions
refuse the question as above; this is a refusal check, not a passing lint oracle.
The fact adapter itself passes direct Go tests in about 0.03s. Its library-flag
mutant fails with `library owner absent` and exit 1. `go vet` passes. All five
production `.a` files were formatted and checked through the existing direct Go
formatter/configured-lint API gate, with zero findings. Gap source is excluded.

Ahra's correction said to keep changes in owned directories and not edit shared
files to bypass blockers. The follow-up authorizes new question files but does
not explicitly resolve the restriction on the shared bridge dispatcher. A
clarification about the minimal dispatch change was requested; no answer has
arrived. This continuation therefore changes no shared dispatcher, registration
generator or test harness. The proposed registration, for review, is one import
and one switch arm in bridge/tsgo/checker/facts.go:

```go
import wave10next "github.com/system-inc/adamic/stage1/cohere/typeaware/wave_10_next"
// Inside Program.Inspect's switch, using its validated source/node and checker:
case "runtime-context":
    return wave10next.Inspect(p.Compiler, source, node, c, question)
```

A second blocker was observed while compiling the combined native component
probe. It ran for over three minutes at CPU load; SIGQUIT captured a stack in
`internal/fresh.(*analysis).argument` / `made`. A retry with distinct literal
class discriminators was bounded at 60s and exited 124. No compiler file was
edited. `gaps/native_atoms.a` preserves the combined probe. The timeout-only
probe is kept separate and builds successfully.

`testdata/cohere_atoms_test.go` is an owned, additional Go test compiled through
an overlay in the pinned cohere nexus package. It calls unchanged production
helpers for write-chain/catch-state reduction, lost handles and import reachability.
The combined native build blocker prevented its comparison from passing. An
attempt against the earlier incomplete binary produced empty output rather than
Go's five lines; the failed log is preserved and is not counted as evidence of
native equivalence.

## Checks and limits

All output went to log files; evidence is preserved here. The source gate used
`/workspace/wave-10-next-source.manifest` listing the five production `.a` files
with repository tsconfig. Commands included:

```sh
source /workspace/adamic-tools/env.sh
go test ./stage1/cohere/typeaware/wave_10_next -count=1 -timeout=30s -v > context-test.log 2>&1
go test -overlay /workspace/wave-10-next-context-mutant.json ./stage1/cohere/typeaware/wave_10_next -count=1 -timeout=30s -v > context-mutant.log 2>&1
go vet ./stage1/cohere/typeaware/wave_10_next > vet.log 2>&1
/workspace/wave-10-source-gate "$PWD/tsconfig.json" /workspace/wave-10-next-source.manifest --format > format.log 2>&1
/workspace/wave-10-source-gate "$PWD/tsconfig.json" /workspace/wave-10-next-source.manifest > source-lint.log 2>&1
/workspace/wave-10-validation/adamic build stage1/cohere/typeaware/wave_10_next/timeout_probe.a -o /workspace/wave-10-next-timeout --tsgo /workspace/wave-10-validation/checker.a > timeout-build.log 2>&1
/workspace/wave-10-next-timeout stage1/cohere/typeaware/testdata/tsconfig.json /workspace/wave-10-next-positive.a > timeout-run.log 2>&1
```

The raw fact mutant overlays only the owned Go adapter and substitutes false for
the compiler's default-library flag. The positive source was
`Promise.race([Promise.resolve(1), new Promise((resolve, reject) => { setTimeout(reject, 10); })]);`.
Sanitizer compilation used `--sanitize` and the existing sanitized checker archive;
execution enabled `ASAN_OPTIONS=detect_leaks=1`. Its only reported failure was
the unsupported question, exit 70.

No continuation corpus comparison, complete rule mutant, released-handle test
for the new question, emitted-JavaScript comparison or native-versus-Go timing
has passed or is claimed. Zero corpus volumes would not prove these rules work.
The earlier three completed ports and their evidence are unaffected. Claims
remain marked partial and blocked, so another worker does not mistake these
components for completed rules.
