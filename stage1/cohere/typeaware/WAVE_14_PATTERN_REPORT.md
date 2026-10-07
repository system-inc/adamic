Added native ASCII pattern validation, flat classes and Go-compatible syntax errors.
Rebased onto main f8013f0b; tested source head 852560ba, superseding pushed 74797828.
All owned suites PASS 505.910 s; bridge, uncached filtered Node oracle and vet pass.
Twenty semantic byte mutants, nine declaration mutants and released-handle mutants caught.
Remaining pattern grammar, shared numeric dispatch, undefined labels and JSX are incomplete.

regexp_pattern_validation.a is a new owned native helper. It validates ASCII
literals, alternation, anchors, dots, empty/negated flat classes and unescaped
character ranges. It reports the production Go rewrite's exact messages for
unclosed classes, backward ranges and a final backslash. no_invalid_regexp.a
now calls it after the existing global-constructor and flag checks. Full
constructor ranges, message IDs and empty fixes/suggestions match Go.

The order is intentional: Go finds a class's closing bracket before decoding
members, then decodes the entire body before joining ranges. The helper follows
that order within its supported domain. Escapes in closed classes, dash range
endpoints with Annex B/u differences, non-ASCII patterns, groups and quantifiers
remain explicitly unsupported. They panic NotYet instead of silently reporting
an unverified error or accepting an unverified pattern. Known v flags still skip
pattern validation, as production Go does. The supported slice is invariant
between u and non-u, so its dynamic-flags judgment matches Go's two readings.

The former Go-positive new RegExp('[') boundary now reports normally. A Go-positive
new RegExp('(') remains an explicit native NotYet boundary. This is additional
implementation of the claimed rule, not a claim of completed pattern validation.

Only owned source, the owned third-suite test and evidence changed. No shared
registration generator, harness, parser, finding model, compiler implementation
or production Go rule changed. No bridge question was added. No new rules were
claimed and no branch other than codex/typeaware-wave-14 is published.

During the initial supported-pattern run, main advanced from e8ba3d5d to
f8013f0baac41ddc340d76f83bddde38536a8f07. The branch rebased cleanly, then reran
all owned oracle suites, the complete bridge packages, filtered uncached Node
oracle and vet. Publishing uses a lease against the previously pushed
747978289db112872c02f01e7565678c4d07a1f2, as authorized by the explicit rebase/push
instruction. Earlier reports and the initial 243.155 s passing pattern run remain
historical evidence; the rebased results below are the landing gate.

Setup before the rebase completed in 27 s. Rebased setup printed ready at 0 s
for Go, clang, Node and submodules, cache warm at 106 s and done at 106 s.
nproc 5, cpu.max 400000 100000, memory 17.6 GB. Go 1.27.1, clang 20.1.8,
Node 24.19.0. Setup and correctness checks overlapped; timing measurements ran
quietly only after all builds and tests completed. Shells source
/workspace/adamic-tools/env.sh and every test command redirects to a log file.

```sh
go test ./stage1/cohere/typeaware -run '^TestWave14' -count=1 -v -timeout=30m
ADAMIC_TSGO_CORPUS=/workspace/wave-14-typescript \
  go test ./bridge/tsgo/... -count=1 -v -timeout=15m
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle \
  -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(maps_and_text|sorting|string_index|lone_surrogates|functions|closures)\.a$' \
  -count=1 -v -timeout=10m
go vet ./...
```

The owned invocation sets ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-14-typescript,
the original compiler/repository manifest variables and corresponding NEXT
variables to /workspace/wave-14-artifacts/{compiler,repository}.manifest. The four
artifact variables point to /workspace/wave-14-pattern-rebased-{original,next,
render,third}. Frozen populations remain 77 compiler roots and 287 repository
roots. No supported corpus was filtered or replaced to evade a finding.

| Rebased suite | Result |
| --- | --- |
| Numeric listener declarations | PASS 0.01 s |
| Continuation rules | PASS 155.27 s |
| Leaked-render judgments | PASS 51.12 s |
| Original three rules | PASS 67.84 s |
| Third batch and new pattern slice | PASS 231.67 s |
| Complete owned invocation | PASS 505.910 s |
| Bridge | PASS 72.162 s |
| Checker package | PASS 0.231 s |
| Filtered uncached Node oracle | PASS 10.782 s |

Every supported population compares complete findings, fixes and suggestions
against unchanged production Go in normal and ASan/UBSan/LeakSanitizer builds.
All native supported stderr is empty. Original controls have 33 findings, with
three compiler findings. Continuation controls have 15 findings plus four module
settings. Synthetic render controls have 18 findings. Existing third controls
have 60 findings and constructor controls 58. Both frozen corpora agree in every
suite. Complete output and input evidence is retained in validation-wave-14-patterns.

The new pattern population has 23 patterns under five flag readings: empty,
u, imsu, v and a dynamic string. Its 115 files produce 36 findings and 12,802
identical bytes under normal and sanitized native. It covers valid literals,
empty and negated classes, forward/backward ranges, leading/trailing dashes,
class closure errors, a final backslash and competing class failures.

A second matrix checks every supported individual ASCII character and all
forward/backward/negated ranges among 09AZaz:@, under empty, u and imsu flags.
The 741 constructors produce 168 findings and 24,997 identical bytes in Go,
rebuilt native and sanitized native. SHA256:
0ca4e3e715a2f617412d56fe6650e2dc63861b7df3f02112dec8e1c13a7474c7.
The exact matrix .a input, manifest and three complete streams are retained.

Twenty semantic mutants compile, exit 0 with empty stderr, and fail only full
diagnostic-byte comparison. First differing bytes on rebased main:

| Mutant | Byte |
| --- | ---: |
| Array delete judgment | 63 |
| Stringification | 1576 |
| Extraneous class | 4460 |
| Listener assertion | 111 |
| Module namespace | 4719 |
| Leaked render | 61 |
| Cooked mapping | 65 |
| Constant writes | 8800 |
| Constructor flags | 475 |
| Overridden literal checked twice | 10418 |
| Constant deduplication | 6750 |
| Alias tracking | 13003 |
| Global writes | 22309 |
| Label membership | 60 |
| Flag precedence | 5296 |
| Unicode quoting | 5917 |
| Surrogate decoding | 18425 |
| Combining class | 7091 |
| Scope meaning Value to Variable | 1398 |
| Pattern range direction | 4970 |

The new range mutant reverses the actual native range comparison. It changes
both good and bad range input and is caught by the independent byte oracle.
The nine wrong-kind declaration mutants are metadata assertions rather than
claims of shared-driver execution. Original, continuation and third suites also
prove the released-handle panic 70 and catch a registry-retention mutant exiting 0.
The complete bridge gate repeats its seven ABI/ownership/type/link/region mutants:
ASan catches both lengths, the stale assertion catches retention, byte comparison
catches the wrong type position, the refusal assertion catches missing link opt-in,
and LeakSanitizer catches missing output frees and an unowned heap region result.
It compares 1,600 positions across four files, 54,982 bytes under sanitizers.
The filtered Node gate includes its one-byte oracle mutant and nine matched fixtures.

Final quiet alternating three-round whole-process count medians:

| Population | Native seconds | Go seconds | Native / Go |
| --- | ---: | ---: | ---: |
| Compiler | 1.708461 | 0.322694 | 5.29x |
| Repository | 0.227980 | 0.125613 | 1.81x |
| Existing controls | 0.020106 | 0.026980 | 0.75x |
| Constructors | 0.026143 | 0.029188 | 0.90x |
| Upstream constructors | 0.023016 | 0.031568 | 0.73x |
| New patterns | 0.027874 | 0.038090 | 0.73x |

Compilation and sanitizers are excluded. Constructor/upstream rows use
--class-only in both implementations; the upstream row checks its count of 89,
with its prior full-byte evidence retained in the landing report. Timings and
the reproduction script are retained. The current driver still ignores numeric
listener declarations, so no shared-dispatch speedup is claimed.

Remaining work is substantive: full ECMAScript rewriting and regexp2-compatible
pattern validation/errors are not yet ported. Real JSX and undefined labels
remain shared-parser gaps. ParseNode still has only kind: string and the current
driver calls run(), so handed-node numeric dispatch remains a shared API gap.
The latest rule.json contract applies to newly claimed rules; this unit claimed
none and did not install incompatible manifests into the syntax-only registry.
No batch-8 finding-model SHA was supplied and that model is not on this main.
The full repository gate, nondefault options and new-rule emitted JavaScript
comparison are not claimed covered. No further batch is taken while the owned
pattern paths remain incomplete.
