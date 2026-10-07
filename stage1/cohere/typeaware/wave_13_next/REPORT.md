Built: native .a ports of no-process-exit-after-output, no-uncleared-race-timeout and require-blocking-standard-streams, plus five dedicated raw bridge questions.
Commits: claim d37a6984 preceded code; implementation 7f0647dd; original three ports remain pushed through 7ca1f20e.
Commands and outputs: 129 controls / 254 roots / 166 findings / 120897 bytes agree; compiler 77 roots / 4933 bytes and repository 287 roots / 18485 bytes agree; normal and sanitized runs pass.
Mutants: three rule and four raw-fact mutants compile and exit normally, caught only by comparison; five suffix mutants and seven bridge foundation mutants are caught; released handle refuses access.
Not covered: full repository Go gate, JSX/JavaScript populations, suppression or fix application, whole-source cohere CLI lint, and native construction of the raw control-flow graph.

## Scope and integration

These were the next three unported/unclaimed checker rules after the original
wave 13 trio. The claim was pushed before code. The earlier BLOCKED.md describes
a scope interpretation that has now been resolved: the user explicitly permits
new dedicated questions with minimal dispatcher entries. Five switch cases (ten
Go-formatted lines) are the only shared integration edit. No shared registration
generator, existing harness, protected compiler file or submodule pin changed.
New executable Adamic sources are .a; exported upstream TypeScript programs are
fixture data, not Adamic modules.

Each rule makes its findings in native code. The Go oracle invokes the unchanged
production Nexus rules and serializes the complete findings, fixes and
suggestions. It does not call bridge helpers. All three rules have no edits or
suggestions; their zero counts are compared, not omitted. Native declaration
checks use exact compiler symbol/declaration ancestry, resolved signature bodies,
Never flags and resolved import edges, not spelling or file-name approximations.
Source-context sets top-level await parsing only for external non-declaration
modules. Imported files are parsed as their own bindings contexts.

The bridge also supplies a raw syntax graph. Its generic graph builder remains
Go, copied from pinned cohere control_flow_graph with retained MIT provenance
under bridge/tsgo/checker/outputflow. No production cohere rule is imported or
invoked there. Native code performs all path traversal, write-state antichains,
catch resets, callee following, lost-timer checks, module closure and blocking
order decisions. This graph-construction boundary is an explicit limitation.
See QUESTIONS.md for the five raw fact interfaces.

## Agreement and meaningful positive coverage

The pinned TypeScript compiler population is v6.0.3 at
050880ce59e30b356b686bd3144efe24f875ebc8, using the original frozen 77 roots.
The frozen repository population is the original 287 roots. These three rules
produce zero findings in both populations; positive controls and mutants are
therefore essential. The 121 unchanged upstream production test cases were
exported through test-file-only overlays, retaining their setup and all helper
roots, and passed Go's own assertions. Eight additional controls cover computed
and parenthesized console keys, Unicode/CRLF, shorthand timer reads, plain writes,
chained assignments and window timers.

| Population | Roots | Findings | Exact bytes | Normal / ASan+UBSan+LSan |
| --- | ---: | ---: | ---: | --- |
| 129 control programs | 254 | 166 | 120897 | equal / equal |
| TypeScript compiler | 77 | 0 | 4933 | equal / equal |
| Frozen repository | 287 | 0 | 18485 | equal / equal |

Controls report 131 process-exit, 12 race-timeout and 23 blocking-stream findings.
Canonical streams, complete fixture snapshots, per-case hashes and normal and
sanitized results are retained in validation/. Native control stderr is empty;
Go stderr contains its phase timings. Corpus stdout hashes are
`e940d25193f84b9d9a2631c1c83d0a52a5f5b90a7246811d12445a1b08ffebcf`
(compiler) and
`20dc789d505315aefcd366d819ba20b780f64bcdce0e5fbf2f26dc54ac120a2f`
(repository). Corpus sanitized stderr contains timings and no sanitizer finding.

## Mutants and lifetime checks

All seven semantic mutants build, exit 0 and have empty native stderr. Only full
Go/native output comparison catches them:

| Mutant | Change | First differing byte |
| --- | --- | ---: |
| process-state | invert recorded-write state predicate | 134 |
| race-handle | invert lost-handle predicate | 118 |
| blocking-order | disable known blocking import order | 1690 |
| platform-declaration | declare-file flag false | 134 |
| resolved-return | all return flags Never | 120 |
| graph-reachability | invert reachable blocks | 134 |
| module-resolution | erase resolved target paths | 1412 |

Removing suffix rejection separately from each of platform-symbol,
call-declaration, output-flow, program-modules and source-context is caught by
TestOutputCheckerQuestions (exit 1, accepted suffix). The dedicated checker tests
also reject wrong node kinds and inspect declaration ancestry, resolved bodies,
module edges, graph events and source context.

The full touched bridge package gate passed in 149.709s. Its independent
foundation agreement covers 162 positions across one file, 3261 identical bytes,
under ASan/UBSan/LSan. Its seven mutants are caught: input/output off-by-one by
ASan; retained released handle by stale assertion; wrong position by oracle at
byte 6; removed link opt-in by refusal; omitted C free and heap allocation replacing
region allocation by LSan. The ABI test runs 100 queries, verifies outputs survive
release, rejects zero/stale handles, distinguishes a second handle and includes
世界🌍. The new source-context native released-handle probe exits 70 with
`invalid or released checker handle` under sanitizers. All logs are retained.

## Native time against Go

Three alternating full-output runs were measured after other builds/tests ended.
All twelve streams retained their corpus hash. Medians include program loading:

| Population | Native | Production Go | Native / Go | Bridge queries |
| --- | ---: | ---: | ---: | ---: |
| Compiler | 3.020283s | 0.438155s | 6.89x | 110 |
| Repository | 0.441537s | 0.154423s | 2.86x | 319 |

This native implementation is slower on both measured populations. Phase timings
and all three rounds are in validation/measurements.json. Earlier concurrent
measurements are retained but are not the primary time comparison. Original cloud
setup reported Go 0s, clang 1s, Node 1s, submodule 1s, compiler cache 148s, done
148s; nproc was 5 (four-core cgroup). See the original wave 13 setup evidence.

## Commands and reproduction

Every test writes directly to a log file, without test-output pipelines. Use
`source /workspace/adamic-tools/env.sh` first. Build the stage0 compiler with
`go build -o /workspace/wave-13-adamic ./cmd/adamic`, the normal bridge with
`go build -buildmode=c-archive -o /workspace/wave13-next-checker.a ./bridge/tsgo/archive`,
and native rules with
`/workspace/wave-13-adamic build stage1/cohere/typeaware/wave_13_next/suite.a -o /workspace/wave13-next-native --tsgo /workspace/wave13-next-checker.a`.
The build logs are included. For the oracle, use a Go overlay mapping
`cohere/wave13_next_oracle.go` to this directory's `testdata/oracle.go`, then run
`go build -overlay <overlay.json> -o /workspace/wave13-next-oracle ./wave13_next_oracle.go`
inside cohere. No production rule edits are required.

The committed scripts accept the following arguments (all paths absolute):

```
python3 extract_controls.py /workspace/adamic /workspace/wave13-next-upstream
python3 extra_controls.py /workspace/wave13-next-upstream/controls
python3 validate.py /workspace/wave13-next-native /workspace/wave13-next-oracle /workspace/wave13-next-upstream/controls /workspace/wave13-next-validation
python3 corpora.py /workspace/wave13-next-native /workspace/wave13-next-oracle /workspace/wave13-next-corpora
python3 mutants.py /workspace/adamic /workspace/wave13-next-mutants /workspace/wave13-next-upstream/controls /workspace/wave13-next-validation /workspace/wave-13-adamic /workspace/wave13-next-checker.a
```

Use portable manifests in validation/ by replacing TYPESCRIPT and ADAMIC with
checkout paths; corpora.py expects the original /workspace paths shown above.
The corresponding sanitized runs use the `-asan` binary/archive, built with
`-asan` for Go's C archive and `--sanitize` for native Adamic. The touched gate was
`go test ./bridge/tsgo/... -count=1 -v`; the dedicated final guard check was
`go test ./bridge/tsgo/checker -run '^TestOutputCheckerQuestions$' -count=1 -v`.
Go vet, gofmt and git diff --check pass. The shared cohere CLI at the pinned
submodule does not discover .a sources, so a whole-source cohere lint pass is not
claimed. Existing shared harness files are unchanged; this directory's own .a
suite provides complete finding serialization and compilation.
