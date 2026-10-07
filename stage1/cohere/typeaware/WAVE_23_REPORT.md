Built: only-throw-error, prefer-promise-reject-errors and prefer-reduce-type-parameter, each in its own .a file.
Commits: claim f6ffdf22; implementation e628ca7b; base 0d540f4 on codex/tsgo-c-library.
Commands and outputs: wave gate PASS 90.402s; bridge PASS 94.766s; checker PASS; filtered Node oracle PASS 15.017s; vet and gofmt logs empty.
Mutants: three native rule mutants and two raw fact mutants caught only by Go finding bytes; released-registry mutant caught by the required panic; foundation mutants also caught.
Not covered: configurable options, full upstream fixture matrices, the full repository test gate, suppression/edit application, JSX populations, and the pinned cohere CLI's unsupported .a paths.

The branch is `codex/typeaware-wave-23`, based on the requested bridge branch
rather than main. The three rules are positions 67, 68 and 69 in the descending
combined counts linked from VOLUME_REPORT.md, excluding the existing 26 ports
and resolving zero-count ties lexically. Each has zero compiler and repository
findings in the original selection evidence. The claim was pushed before source
implementation. The initial scan checked 268 origin refs; a final fetch and
content/claim scan checked 310 refs, deduplicating 98 stage1 trees. No competing
port or claim was found. The only matching unrelated source was the configuration
set table on two other branches. No rule was skipped. No submodule pin changed.

The runner is [wave_23_suite.a](wave_23_suite.a). It uses one checker program,
parses each file with Adamic, makes the three native decisions, and sorts full
canonical finding lines. The independent Go overlay calls the pinned registry's
three unchanged production `Run` functions with default options and its own
program loader and AST walk. It imports no bridge code. Comparisons include
headers, complete findings, ordered fix triples, complete suggestions and the
summary, retaining duplicates. Proposed edits are compared without applying them.
Only reduce produces repairs: controls include 10 reduce findings with 29 ordered
fix triples. These three rules produce no suggestions under the supported defaults.

The native files preserve Go cohere's behavior, including the permissive any and
unknown throw defaults, restrictive rejection defaults, caught-binding rethrows,
thenable rejection-handler selection, Promise executor symbol identity and
shadowing, default-library and inherited Error identities, Readonly<Error>, and
reduce's assignability condition, array/tuple intersections/unions, const-assertion
exemption, existing type arguments and parenthesized repair anchors. Go cohere's
computed-key behavior is the authority here, rather than the installed ESLint
implementation's differing constant-folding behavior.

Two raw questions were necessary. Alias symbol origins and arguments support
Readonly<Error>. Apparent first callback parameter types, with rest parameters
indexed numerically, support thenable rethrows. Adamic makes the judgments. Both
questions have named Go and Adamic files. The shared Go dispatcher changes by
one line; the Adamic decoders use the existing dynamic API and need no shared
registration change. Their framing and lifetimes are in
[wave_23_facts.md](wave_23_facts.md). No protected compiler files were changed.

Agreement, normal and ASan/UBSan/LeakSanitizer:

| Population | Roots | Findings | Identical stream bytes |
| --- | ---: | ---: | ---: |
| Targeted controls | 26 | 39 | 13,380 |
| TypeScript compiler | 77 | 0 | 5,318 |
| Frozen repository | 287 | 0 | 18,485 |

Controls include Unicode/CRLF spans, undefined versus null, never/any/unknown,
Error subclasses and constrained parameters, unions/intersections, catch bindings,
private catch members, typed arrow/function/rest rejection handlers, fake thenables,
rest callback parameters, static and executor rejections, nested closures,
shadowed reject bindings, readonly errors, spread arguments, both assertion
spellings, array/tuple/generic receivers, genuine narrowing assertions, parenthesized
calls/arguments, existing type arguments and fix trivia. Controls exercise the
native source files themselves: 15 throw findings, 14 rejection findings and 10
reduce findings. Empty corpus findings are held by these positive controls and
mutants, rather than treated as proof on their own.

| Mutant | Observation and catcher |
| --- | --- |
| Throw: substitute null's flag for undefined's | Exit 0, empty stderr; Go byte oracle differs at byte 89 |
| Rejection: change the report message | Exit 0, empty stderr; Go byte oracle differs at byte 5,351 |
| Reduce: add one to the trailing removal's start | Exit 0, empty stderr; Go byte oracle differs at byte 10,961 |
| Alias facts: omit alias arguments | Exit 0, empty stderr; Go byte oracle differs at byte 6,493 |
| Callback facts: omit a rest parameter's indexed type | Exit 0, empty stderr; Go byte oracle differs at byte 4,509 |
| Registry: retain the released program | Mutant exits 0; normal exits 70 with exactly `invalid or released checker handle` |

`TestTypeAliasInfo` checks alias name, every library declaration origin, alias
argument identity and indexed rest callback identity against direct checker
operations, and rejects malformed identities/suffixes. The bridge foundation gate
checks 100 C ABI queries, buffer survival after release, zero/stale handle refusal
and a distinct next handle. Its independent corpus checks 1,600 positions in four
compiler files, 54,982 identical bytes under ASan/UBSan/LeakSanitizer. The foundation
input/output length mutants trigger ASan; the stale-handle and link-guard mutants
fail their refusal assertions; the wrong-position mutant differs at byte 6; the
missing C free and heap-instead-of-region mutants trigger LeakSanitizer. ASan
instruments native/C memory, not the external Go heap.

Commands run from the repository after sourcing `/workspace/adamic-tools/env.sh`:

```sh
bash cloud/setup.sh > /tmp/wave-23-setup.log 2>&1

ADAMIC_WAVE23_ARTIFACTS=/workspace/wave-23/verified \
ADAMIC_WAVE23_REPOSITORY_MANIFEST=/workspace/wave-23/repository.manifest \
ADAMIC_WAVE23_COMPILER_MANIFEST=/workspace/wave-23/compiler.manifest \
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-23/typescript \
go test ./stage1/cohere/typeaware -run '^TestWave23AgreementAndMutants$' \
  -count=1 -timeout=30m -v > /workspace/wave-23/verified-test.log 2>&1

ADAMIC_TSGO_CORPUS=/workspace/wave-23/typescript \
go test ./bridge/tsgo/... -count=1 -timeout=15m -v \
  > /workspace/wave-23/bridge-test.log 2>&1

go test ./bridge/tsgo/checker -count=1 -v \
  > /workspace/wave-23/checker-final.log 2>&1

go test ./internal/oracle \
  -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(maps_and_text|functions|closures|sorting)\.a$' \
  -count=1 -timeout=10m -v > /workspace/wave-23/node-oracle.log 2>&1

go vet ./... > /workspace/wave-23/vet.log 2>&1
gofmt -l cmd internal bridge/tsgo stage1/cohere/typeaware \
  > /workspace/wave-23/gofmt.log 2>&1
```

The filtered Node gate actually selected six fixtures, including method_closures
and generic_functions, and the one-byte oracle mutant. It compares native/Node/JS
behavior and runs sanitizer/leak checks. The wave gate builds stage 0 and the
normal/sanitized C archives itself. All test outputs were logged, never piped.

Setup succeeded: Go 1.27.1 ready 0s; clang 20.1.8 ready 0s; Node v24.19.0 ready
0s; submodules ready 0s; build cache warm 83s; setup done 83s. `nproc` is 5;
`cpu.max` is `400000 100000`, a four-core quota. Memory reported 17.6 GB. The
initial broad fetch completed origin refs, then began an unnecessary large
recursive submodule fetch; that recursive fetch was stopped. Later fetches used
`--no-recurse-submodules`. The TypeScript corpus was obtained separately at
v6.0.3, verified commit `050880ce59e30b356b686bd3144efe24f875ebc8`.

Three quiet alternating whole-process rounds, each comparing full diagnostic
streams, including load, parse, lint, formatting and teardown:

| Corpus | Native median | Go median | Native / Go |
| --- | ---: | ---: | ---: |
| Compiler | 1.894 s | 0.316 s | 5.99 |
| Repository | 0.266 s | 0.124 s | 2.14 |

These are wall-time observations on this machine, not a claim of Go-speed parity.
The native query counts were 1,694 for compiler and 1,165 for repository. Timings
exclude compilation and sanitizer runs. Raw phases and all rounds are in
[measurements.json](validation-wave23/measurements.json). The reproducible script
is [measure.py](validation-wave23/measure.py); manifests are portable, while output
headers/hash evidence contain the actual absolute paths of this run.

The first diagnostic and sanitizer control run passed before the rest-callback
question was added. Its subsequent source mutant used the older archive and
refused the new question, so that interrupted run is not counted as a mutant kill.
The first alias-argument mutation failed Go compilation because its range variable
became unused; it was replaced with omission of the argument list, then compiled
and caught solely by diagnostic bytes. The first registry mutation named a
nonexistent source anchor; it was corrected to the pinned `programs.live` deletion.
The initial direct checker fixture had no config-recognized roots and was corrected
using the existing `.a` root plus prelude declaration pattern. Final full runs
pass. Initial failure logs are preserved alongside the final evidence.

A dry run of the pinned cohere CLI on the seven new `.a` sources refuses them as
unsupported paths, stating that none is in the program. It is not a passing lint
gate and is not counted as one. Adamic's loader and this bridge explicitly load
`.a` roots, and the native runner and sanitizer controls compile these exact source
files. Fixing the pinned CLI's extension support is outside this unit's territory.
The full repository `go test ./...` gate was not run; the touched bridge packages,
wave package gate, filtered Node oracle and full Go vet/gofmt checks are the
reported gate. No full per-rule configurable option matrix or all upstream
fixtures is claimed. The runner remains a default-options diagnostic comparison
suite, without suppression or edit application.
