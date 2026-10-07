Built: landing refresh of all twelve complete wave 06 rules and the three partial React prepared-HIR validators on origin/main f8013f0b. No new claims or shared implementation edits.
Commit: the rebased implementation parent is 733c380c06486f74de94e0f5607e64d25d606f0; this evidence commit is pushed only to codex/typeaware-wave-06.
Checks: original, Nexus, constructors, callbacks, React cores/backends/refusals, bridge, vet and filtered uncached Node oracle all pass after rebasing.
Mutants: all twelve complete rule mutants, three raw-question mutants, stale-handle guards, four React core mutants, three reporting mutants, three refusal mutants, seven bridge mutants and Node one-byte mutant caught.
Not covered: native React source lowering and numeric parser/driver integration remain blocked. No full root gate or new claims.

Main advanced from e8ba3d5d to f8013f0baac41ddc340d76f83bddde38536a8f07. The branch rebased cleanly and every owned suite was rebuilt and rerun. A final remote read confirms main remains that SHA and the own remote before replacement is 66cc4d5bfbd6179ef0375166526a6929daa74109. User-authorized rebasing requires replacing the own branch with an exact force-with-lease against that SHA; main and area branches are never pushed.

| Suite | Controls | Findings | Identical canonical bytes |
| --- | ---: | ---: | ---: |
| Original three | 166 | 129 | 36604 |
| Nexus three | 84 | 42 | 28696 |
| Constructor three | 109 | 51 | 23338 |
| Callback/throw/regex three | 210 | 136 | 51544 |
| React prepared HIR | 91 admitted of 100 | 48 | 36855 |

Complete finding, fix and suggestion records match unmodified Go. Each suite repeats normal and ASan/UBSan comparisons over the frozen 77 compiler and 287 repository roots. Original compiler records contain 53 findings and 22060 bytes; all other compiler records contain zero findings and 5318 bytes. Repository records contain zero findings and 18485 bytes. The React corpus runs consume Go-prepared HIR and are not native source parity. The React canonical control byte total includes source-path headers; the scratch path changed since the prior report, while the finding count remains 48. All twelve prepared-HIR batches and both corpora also match source Node and emitted JavaScript.

Each complete rule semantic mutant exits zero and is caught only by independent Go bytes: nominal symbol identity, duplicate span, return fix span; three Nexus diagnostic IDs; constructor shadow rejection, callee span and parenthesis skipping; throw acceptance, regex direction and arrow replacement text. Raw promised-shape and declaration questions are separately mutated and caught by Go bytes. Stale handles panic 70, and disabling the registry guard loses that panic. React setter guard, unconditional membership, creator propagation and predecessor alias mutants all exit zero and lose Go byte equality. Three reporter mutants lose Go equality; three removed source refusals lose panic 70. The JSX parser probe still refuses.

Bridge suite passes in 76.963s and checker package in 0.160s. It checks 162 positions and 3261 bytes under ASan/UBSan/LSan, released/zero handles, durable output ownership and unlinked calls. Its seven mutants are caught by: input/output length ASan failures; stale-handle assertion; Go type oracle mismatch; link refusal; LeakSanitizer for omitted output frees and heap allocation at region entry.

The uncached Node oracle covers functions, generic functions, closures, method closures, regexp-cycle closures, maps/text, sorting, string indexing and lone surrogates, plus TestTheOracleCatchesOneByte. The exact newly landed TestLibraryMapSetIteratorCopiesRefused also passes. go vet ./... emits no output. All test streams go directly to logs.

Observed original-suite whole-process timing: compiler native 24.286391s versus Go 3.743972s (6.487x); repository native 1.050216s versus Go 0.314647s (3.338x). Other suite timings are below. These landing runs overlap the React validation, so they are observations under contention, not isolated performance benchmarks. Builds are excluded.

| Suite | Compiler native / Go seconds | Repository native / Go seconds |
| --- | --- | --- |
| next | 5.076207 / 1.878422 | 0.695374 / 0.281430 |
| constructors | 1.900735 / 0.369205 | 0.338644 / 0.175913 |
| callbacks | 3.124469 / 0.480722 | 0.424200 / 0.191238 |

The existing configured toolchain is reused; no setup reinstall was needed. Earlier setup timing was 105s total and nproc 5 (CPU quota 4), as recorded in CORE_REPORT.md. Source adapter gaps and Go two-creator phi ambiguity remain exactly as documented there. Current main and lint-harness-dot-a still expose ParseNode.kind as a string. The three numeric SourceFile listener declarations remain in their owned rule.json files, awaiting the shared numeric parser and driver contract. No per-rule string relevance adapter is introduced.

Reproduction uses source /workspace/adamic-tools/env.sh and TMPDIR=/workspace:

```sh
ADAMIC_WAVE06_CORPORA=1 ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-06-typescript ADAMIC_WAVE06_ARTIFACTS=/workspace/wave-06-landing3-original go test ./stage1/cohere/typeaware -run '^TestWave06(AgreementAndMutants|PinnedFlags)$' -count=1 -timeout=30m -v > /tmp/wave-06-landing3-original.log 2>&1
python3 stage1/cohere/typeaware/wave_06_next/validate.py --scratch /workspace/wave-06-landing3-next --compiler /workspace/wave-06-typescript > /tmp/wave-06-landing3-next.log 2>&1
python3 stage1/cohere/typeaware/wave_06_constructors/validate.py --scratch /workspace/wave-06-landing3-constructors --compiler /workspace/wave-06-typescript > /tmp/wave-06-landing3-constructors.log 2>&1
python3 stage1/cohere/typeaware/wave_06_callbacks/validate.py --scratch /workspace/wave-06-landing3-callbacks --compiler /workspace/wave-06-typescript > /tmp/wave-06-landing3-callbacks.log 2>&1
python3 stage1/cohere/typeaware/wave_06_react_state/validate_cores.py --scratch /workspace/wave-06-landing3-react --compiler /workspace/wave-06-typescript > /tmp/wave-06-landing3-react.log 2>&1
python3 stage1/cohere/typeaware/wave_06_react_state/validate_core_backends.py --scratch /workspace/wave-06-landing3-react > /tmp/wave-06-landing3-backends.log 2>&1
python3 stage1/cohere/typeaware/wave_06_react_state/validate_partial.py --scratch /workspace/wave-06-landing3-refusals > /tmp/wave-06-landing3-refusals.log 2>&1
go test ./bridge/tsgo/... -count=1 -timeout=15m -v > /tmp/wave-06-landing3-bridge.log 2>&1
go vet ./... > /tmp/wave-06-landing3-vet.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(maps_and_text|sorting|string_index|lone_surrogates|functions|closures)\.a$|^TestLibraryMapSetIterators' -count=1 -timeout=10m -v > /tmp/wave-06-landing3-node.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestLibraryMapSetIteratorCopiesRefused$' -count=1 -timeout=5m -v > /tmp/wave-06-landing3-iterators.log 2>&1
```

The first broad iterator pattern did not select the new test, so its exact name was run separately. A default-sandbox remote read failed to connect to the session proxy; the authorized network-enabled read succeeded and its output is preserved. Evidence streams, command records and uncompressed SHA-256 hashes are in landing3_evidence. Earlier source fixtures and mutant provenance remain in the existing evidence directories.
