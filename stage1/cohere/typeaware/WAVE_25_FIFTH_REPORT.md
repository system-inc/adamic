Built native .a ports of react-hooks/unsupported-syntax and react-hooks/use-memo; react/boolean-prop-naming is partial and explicitly refuses three unsupported paths.
Previous batch pushed in 056799e9; claim b0ad78cd pushed before code; this accompanying commit contains implementation, tests and evidence.
Rule gate PASS 66.240s: 242 controls, 141 configured findings, 364 corpus files; byte comparison, sanitizers and released-handle checks pass.
Three native rule mutants are caught only by Go bytes at offsets 61, 19207 and 43552; an annotation-question mutant fails its independent assertion.
Incomplete: arbitrary boolean regex patterns, cross-file props annotations, typed memo/forwardRef wrappers, options matrix, shared profile/emitted-JavaScript comparison and full repository gate. No more rules claimed.

## Claim and scope

All preceding claimed rules were completed, tested and pushed before selection.
An explicit fetch of every origin head refreshed 394 refs. The selection checked
77 distinct textual claim blobs on all origin branches and excluded ports on
origin/main at e011f8f60899586d6373a5ccb07335ad82cfbf3c and
origin/codex/tsgo-c-library at 5afbdb83da2ed7ad9815657cd3f6ececd5294bf6.
The first three available checker-dependent rules in the combined-volume
ranking were unsupported-syntax, use-memo and boolean-prop-naming, each with
volume zero. Claim b0ad78cd was pushed before implementation. The initial
main-only fetch was insufficient because origin's configured fetchspec only
covers main; the subsequent explicit all-heads fetch corrected that before
selection. Snapshots and scripts are in validation-wave-25-fifth.

Every new Adamic source is .a, including scratch fixtures. Scratch-only .tsx
and .d.ts symlinks retain TypeScript source identities. Shared generators,
registration, harness and protected compiler files are unchanged. The only
existing implementation edit is the two-line dispatch case in this worker's
own process_questions.go. No pull request was opened.

## Native implementation

Unsupported-syntax reproduces the compilation eligibility gate, nested function
boundaries, hook calls and reachable component positions. It reports global
library eval references, with statements and inline class declarations at the
exact production spans. Symbol provenance is a generic compiler fact.

Use-memo resolves bare, imported, aliased and React receiver hook references,
then checks missing/inline callbacks, parameters, async/generator flags,
dependency array shapes and captured assignments with production traversal
boundaries. Diagnostic spans and descriptions match Go, including cases where
callback and dependency diagnostics both occur.

Boolean-prop-naming implements nil-options decline, the default pattern and a
second fixed is-prefix pattern, propTypes declarations and assignments, local
props interfaces/type aliases, nested traversal, component recognition and
whole-property reporting. Its verified configured mode is the default options
object. Custom message, nesting, prop-type-name and alternate-pattern constructor
settings are not claimed as a verified options matrix.

The isolated declared-syntax-type Go question and .a decoder return the raw
annotation node identity for a parameter, variable, property signature, type
alias or parenthesized type. No lint verdict or component decision crosses the
bridge. Existing generic syntax and symbol projections are reused.

## Independent comparisons

The independent oracle calls the unmodified production Go cohere registry at
715ba94f3608a6500086b1076ce5cb7e51b836db with its own loader, checker and AST walk.
It serializes findings, ordered fix ranges/text and suggestions. Native ports
consume generic bridge facts and perform rule decisions in Adamic. Complete
sorted output bytes are compared, not only counts.

| Population | Files | Findings | Identical bytes | ASan/UBSan/LSan |
| --- | ---: | ---: | ---: | --- |
| Controls with default boolean options object | 242 | 141 | 58944 | Pass |
| Controls with nil options | 242 | 84 | 51616 | Normal comparison |
| Frozen repository corpus | 287 | 0 | 18485 | Pass |
| TypeScript v6.0.3 src/compiler | 77 | 0 | 5010 | Pass |

Configured controls contain 33 unsupported-syntax, 51 use-memo and 57 boolean
findings. All eleven emitted message kinds are represented. Fixes and suggestions
are zero, matching these production rules. The 242 distinct inputs are extracted
from production test source text plus three explicit edge controls; none are
parse-invalid. Production expected results are not imported. Inputs from
nondefault-option rows are compared under default options, not their original
options. This is not exhaustive component or TypeScript syntax coverage.

Corpus manifests are unchanged from validation-coverage. TypeScript is pinned
at 050880ce59e30b356b686bd3144efe24f875ebc8. Boolean naming declines under nil
options in corpus comparisons and timed runs, just as Go does. Those measurements
therefore do not measure enabled boolean naming. Normal and sanitized native
comparison stderr is empty. Full bytes, hashes, inputs and logs are retained in
validation-wave-25-fifth.

## Mutants and lifetime

| Mutation | Observation | Catcher |
| --- | --- | --- |
| Library eval declaration test forced false | Builds; exits 0; empty stderr; mismatch byte 61 | Independent Go finding bytes |
| Inline callback condition inverted | Builds; exits 0; empty stderr; mismatch byte 19207 | Independent Go finding bytes |
| Boolean initial-capital range changed to lowercase | Builds; exits 0; empty stderr; mismatch byte 43552 | Independent Go finding bytes |
| Annotation question always returns node identity zero | Assertion fails with lost declared syntax type | TestDeclaredSyntaxType |

After checker release, the new question exits 70 with exactly
`adamic: panic: invalid or released checker handle`. The full bridge suite
passes existing lifetime, ABI, sanitizer and fault-injection checks. The raw
question test also checks four annotation identities and malformed requests.

## Remaining blockers and stopping point

The boolean rule is partial, not a completed port. Three paths refuse with panic
70 before output in dedicated controls:

1. An arbitrary configured regex such as `^enabled$` requires an unimplemented
native regular-expression engine. A standalone .a `new RegExp` probe is refused
by stage 0 at 1:20: `stage 0 can't lower new an Identifier yet`. The library
also documents RegExp outside 0.1. Only two fixed patterns are supported here.
2. A props declaration resolved in another source file needs a projection of
that file's annotations and declarations. Current IDs belong to the input
file's projection; using them across files would be incorrect. The native rule
explicitly refuses this path rather than silently omitting props.
3. Typed memo/forwardRef wrappers require the remaining component detector
logic. A typed wrapper control explicitly refuses before reporting; that
component path is not ported in this batch.

The three refusal controls pass in 16.141s and each emits no findings. Their
stderr evidence is retained. These are implementation/toolchain blockers, not
a claim that the shared harness caused all three. Per the instruction to stop
on other blockers without editing shared files, this tested partial work is
pushed and no additional rules are claimed. The boolean claim remains reserved.

Shared profile registration and emitted-JavaScript comparison remain with
codex/lint-harness-dot-a. This report does not claim them, CLI suppression/fix
application, every production option, previous rule suites rerun or a full
repository test gate.

## Commands and timings

Session setup: bash cloud/setup.sh, then source /workspace/adamic-tools/env.sh.
Reported setup lines: go ready 0s; clang ready 0s; node ready 0s; submodules
ready 0s; build cache warm 89s; done in 89s on 5 processors, cgroup cpu.max
400000 100000, 17.6 GB. nproc is 5. Go 1.27.1, clang 20.1.8, Node v24.19.0.
All test output was written directly to log files without piping.

```
ADAMIC_WAVE25_FIFTH_ARTIFACTS=/workspace/wave-25-fifth-final \
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-25-corpus \
go test ./stage1/cohere/typeaware -run '^TestWave25FifthAgreementAndMutants$' -count=1 -v
ADAMIC_WAVE25_FIFTH_REFUSALS=/workspace/wave-25-fifth-refusals-final \
go test ./stage1/cohere/typeaware -run '^TestWave25FifthBooleanRefusals$' -count=1 -v
go test ./bridge/tsgo/... -count=1
go test ./bridge/tsgo/checker -run '^TestDeclaredSyntaxType$' -count=1 -v
go test -overlay /workspace/wave-25-validation/fifth-annotation-mutant.json \
  ./bridge/tsgo/checker -run '^TestDeclaredSyntaxType$' -count=1 -v
go vet ./...
python3 /workspace/wave-25-validation/fifth-benchmark.py
```

Agreement gate PASS 66.240s; refusal gate PASS 16.141s; full bridge PASS
71.410s and checker 0.260s; raw annotation assertion PASS 0.056s; its mutant
fails as expected in 0.088s. Go vet passes. No new filtered Node oracle run is
claimed for this batch; the preceding batch's external Node check remains in
its own report.

Three interleaved Go/native samples were measured after all tests finished.
Every timed pair matches bytes. Wall times include load, queries, native
projection decoding/traversal and output serialization.

| Population | Native median | Go median | Native / Go |
| --- | ---: | ---: | ---: |
| Compiler | 3.959156s | 0.373697s | 10.595 |
| Repository | 0.516615s | 0.120397s | 4.291 |

Native is slower on both measured populations. Raw samples and bridge timing
counters are preserved; no causal performance claim is made.
