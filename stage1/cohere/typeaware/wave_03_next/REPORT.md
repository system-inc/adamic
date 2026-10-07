Built: three further claimed Nexus rules in native Adamic .a files, with five isolated raw bridge questions.
Commits: claim 83bfce4b; implementation and evidence aee9b9d0f2480d0adf663f9fa3a4b6cdfafd50b5; this report update follows.
Commands and outputs: comparison PASS 191.808s; all 364 frozen corpus roots and 106 control roots match; sanitizer, released-handle, checker, vet and filtered Node checks pass.
Mutants: race/output/blocking build and exit 0 with empty stderr, caught only at bytes 15225/56/8652; retained released handle caught by panic expectation; Node one-byte mutant caught.
Not covered: shared harness registration and emitted JavaScript comparison; the unadapted production CLI cannot reparse .a files.

## Claim and scope

The original wave-03 ports were pushed through 3882c806 before selecting these
rules. After fetching all origin heads, the selection audited 325 refs and 33
Markdown claim files. The full audit is in
../validation-wave-03/continuation-selection.json. Already ported rules on main
and the bridge branch, and all existing claims on every origin branch, were
excluded. The first three eligible rules all have zero findings in both pinned
populations:

- nexus/correctness-no-process-exit-after-output
- nexus/correctness-no-uncleared-race-timeout
- nexus/correctness-require-blocking-standard-streams

Claim 83bfce4b was pushed before implementation. These are the only additional
rules reserved in this continuation. Each rule has its own .a file in this
folder. All native helpers and the standalone driver are .a. The existing
registration generator, shared harness, protected compiler files and submodules
remain untouched. The only edit to an existing source is the fallback line in
this wave's own checker dispatcher, wave_03_questions.go, plus removal of its
now unused fmt import.

## Boundary and independence

Go supplies raw compiler facts through the existing schema-1 framed ABI. Adamic
owns declaration interpretation, platform-symbol checks, local and imported
writer recognition, lost timer-handle checks, import and callback reachability,
blocking-call classification, path state, finding locations and messages.

| Question | Input | Raw fields |
| --- | --- | --- |
| symbol-context | validated node | own and aliased identities, flags and declaration ancestry; shorthand value identity |
| resolved-call-declaration | call, new or tagged expression | signature and declaration presence, declaration ancestry and function body span, return-type flags |
| code-path-graph | code path root | reachable blocks, successor indices, every expression/statement hook with kind and byte span |
| program-module-edges | SourceFile | all program file names, declaration-file flags, raw static/load occurrences, type-only/literal flags and resolved targets |
| source-module-context | SourceFile | compiler external-module flag, used to initialize native parser await context |

Declaration ancestry includes source path, declaration/default-library/module
flags, textual names, node flags, kind/span, body span and function flags.
Binding patterns and computed names are not textual names and serialize as an
empty name; their kind and span remain present. The module-edge question keeps
the pinned production consumer's textual import/require fast path.

The raw syntax graph is an isolated MIT-licensed copy of six files from
cohere/internal/utilities/controlflow, under bridge/tsgo/code_path_graph.
Package declarations and comment punctuation were adjusted; rule policy was
not copied into Go. The original framework comes from rslint commit
44956a0c53201157aaf5bc19e1c2728896c83ec4; its attribution and license are retained.
Graph events are unclassified: there are no write, exit, blocking or diagnostic
verdicts in this bridge question.

The comparison executable imports and runs the three unchanged production Go
rules through an independent loader, walker and canonical serializer. It
imports no bridge code. Both sides use the same pinned typescript-go checker,
and the syntax-flow algorithm is shared by provenance. The oracle independently
checks native rule decisions; it is not an independent validation of that
shared graph algorithm.

## Controls and observations

Controls are extracted from all three pinned upstream test sources and extended
with Unicode, shadowed/read timer handles, multiple timers, loops, catch-binding
exits, shebangs, helper writers and a destructured-parameter regression. All 94
candidate controls parse under Go and are compared; none is excluded. A separate
10-root program tests real StandardStreams imports, aliases, before/after order,
local calls, argument evaluation, callbacks, awaits, class static blocks,
imported writer bodies and computed imports. Another two-root program proves
imported-entry exclusion.

The canonical stream compares rule names, byte ranges, message IDs and text,
every fix/replacement and every suggestion. These three production rules emit
no fixes or suggestions. Empty fix/suggestion lists are included in every
compared record; there is no positive fix or suggestion case to exercise.

Ordinary and ASan/UBSan/LSan runs compare complete streams. The sanitizer runtime
also enables leak detection, and expected successful runs have empty stderr.
The 77 compiler and 287 repository roots are exactly the frozen base manifests,
with the same pins and source hashes documented in ../WAVE_03_REPORT.md.

Three mistakes were exposed during development and corrected locally: the
shared diagnostic class's default namespace was wrong for Nexus rules; the
native never-return test used the wrong type-flag bit; declaration ancestry
called Node.Text on destructuring names. The top-level await import control
initially failed because the parser's module await context was unset. The new
source-module-context question fixes it without changing the shared parser.
Superseded logs are retained separately from the final validation evidence.

## Integration limits

codex/lint-harness-dot-a was absent from the fetched origin refs. This folder
has a standalone native entry and an owned comparison test, and has not changed
the shared registration generator. Integration must register its modules and
add the emitted-JavaScript comparison when that worker's harness lands.

The production CLI with the earlier scratch loading adapter still panics in
proposalsForText because .a has unknown ScriptKind. A scratch-only Go build overlay supplies .a loading aliases and
ScriptKindTS during reparse; the production lint rules and formatter remain
unchanged. With that adapter the source gate checks all 19 native modules.
The overlay and exact logs are retained as evidence. No cohere submodule source
was edited. A stock, unadapted full source gate is not claimed.

The earlier wave-03 shared TestInspectRequestRefusals blocker remains: its
unknown-question mutant expects a literal return in facts.go that this branch
moved into an isolated dispatcher. No shared test was rewritten to make it pass.
The full repository test gate is not claimed. Relevant rule comparisons,
checker tests, vet and the filtered Node oracle are recorded below.


## Final results and timings

| Population | Roots | Findings | Canonical bytes | Fixes | Suggestions |
| --- | ---: | ---: | ---: | ---: | ---: |
| Main controls | 94 | 96 | 58887 | 0 | 0 |
| Ordered/import controls | 10 | 12 | 7036 | 0 | 0 |
| Imported entry controls | 2 | 1 | 634 | 0 | 0 |
| Repository | 287 | 0 | 18485 | 0 | 0 |
| Compiler | 77 | 0 | 5857 | 0 | 0 |

All rows match complete ordinary native and production Go streams. Main and
ordered controls plus both corpora also match sanitizer streams, with empty
native sanitizer stderr. Imported-entry controls ran in the ordinary comparison.
The main controls contain 58 exit-after-output, 26 blocking-stream and 12
timeout findings. The regression for destructured parameters is included.

| Population | Native wall seconds | Go wall seconds | Native / Go |
| --- | ---: | ---: | ---: |
| Repository | 0.851238540 | 0.376977993 | 2.26 |
| Compiler | 6.301710826 | 0.994304173 | 6.34 |

These are single warm full-output process samples, with all finding bytes
compared again. They include program load, parsing, rule work, serialization
and process exit. Native is slower on both measured populations. The internal
bridge and Go phase counters are saved in validation/measurements.json; no
broader performance conclusion is asserted.

| Mutant | Result |
| --- | --- |
| race timeout: invert lost-handle condition | Builds, exit 0, empty stderr; byte oracle differs at 15225 |
| exit after output: invert reached-exit membership | Builds, exit 0, empty stderr; byte oracle differs at 56 |
| blocking streams: invert shebang reason | Builds, exit 0, empty stderr; byte oracle differs at 8652 |
| release registry: retain program entry | Builds and exits 0; required released-handle panic 70 is absent |
| filtered Node oracle: one-byte output mutation | Oracle test catches it |

The release probe queries program-module-edges after release and must print
exactly `adamic: panic: invalid or released checker handle` to stderr, with exit
70. Its ordinary successful counterpart and the archive mutant are recorded.
Rule mutants must build and exit 0 with empty stderr before differing bytes
count as a kill. Their full outputs are retained, not only finding totals.

## Commands and evidence

The cloud setup from the first tranche succeeded: Go ready 0s, clang ready 1s,
Node ready 1s, submodules 1s, warm 174s, total 174s; nproc printed 5 and cgroup
quota allows four cores. This continuation uses the same environment: Go 1.27.1,
clang 20.1.8 and Node 24.19.0. Setup was not rerun.

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_WAVE03_NEXT_ARTIFACTS=/workspace/wave-03/next-final3 \
ADAMIC_WAVE03_REPOSITORY_MANIFEST=/workspace/wave-03/repository.manifest \
ADAMIC_WAVE03_COMPILER_MANIFEST=/workspace/wave-03/compiler.manifest \
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-03/typescript-corpus \
go test ./stage1/cohere/typeaware \
  -run '^TestWave03ContinuationAgreementAndMutants$' -count=1 -timeout=30m -v \
  > /workspace/wave-03/next-final3.log 2>&1
# PASS 191.808s

go test ./bridge/tsgo/checker ./bridge/tsgo/code_path_graph \
  -count=1 -timeout=30m -v > /workspace/wave-03/next-checker-final.log 2>&1
# PASS checker 0.275s; graph has no direct Go tests

go vet ./bridge/tsgo/checker ./bridge/tsgo/code_path_graph \
  ./stage1/cohere/typeaware > /workspace/wave-03/next-vet-final.log 2>&1
# exit 0, empty log

gofmt -l bridge/tsgo/checker bridge/tsgo/code_path_graph \
  stage1/cohere/typeaware > /workspace/wave-03/next-gofmt-final.log
# empty log

ADAMIC_WAVE03_ARTIFACTS=/workspace/wave-03/next-original-regression \
ADAMIC_WAVE03_QUICK=1 go test ./stage1/cohere/typeaware \
  -run '^TestWave03AgreementAndMutants$' -count=1 -timeout=15m -v \
  > /workspace/wave-03/next-original-regression.log 2>&1
# PASS 21.086s: original three ports still match, 46 and 4 findings

go test ./internal/oracle \
  -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(maps_and_text|sorting|string_index|lone_surrogates|functions|closures)\.a$' \
  -count=1 -timeout=10m -v > /workspace/wave-03/next-node.log 2>&1
# PASS 31.822s: eight selected fixtures and the oracle mutant

/workspace/wave-03/cohere-next-a \
  --tsconfig /workspace/wave-03/next-cohere-tsconfig.json \
  --no-cache --no-fix stage1/cohere/typeaware/wave_03_next/*.a \
  > /workspace/wave-03/next-cohere-final3.log 2>&1
# PASS: 276 rules, 19 checked, 100% Adamic-ready, 0 findings; scratch adapter

git diff --check
# exit 0
```

validation contains full compressed stdout/stderr for all 40 final commands,
uncompressed stream hashes, exact control sources as JSON, manifests, counts,
phase measurements, source-gate adapter files and final logs. Original corpus
manifests and hashes are retained in ../validation-wave-03. Development failures
are explicitly separated into validation/superseded. No generated binaries or
C archives are committed. The owned Go test reproduces controls and mutants
from the pinned upstream test files.
