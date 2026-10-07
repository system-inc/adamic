Built: corrected all fifteen owned listener declarations to named ast.Kind values; active JSX metadata requests supplied nodes and prepared reporters contain only ranges.
Commits: tested code 9fe207dc0b68593347b8b7e5d9fedd2a35486fb1, based on fetched main c01907a7036a22c2ea7ee686ed5fe4c6cd4bbc06; previous published tip fa0053aa467cad15fdb126061c979c2a0ccf8271.
Checks: twelve full Go byte oracles pass on 1,126 controls and frozen 77 compiler/287 repository roots, normal and sanitized; fifteen JSX rendering controls pass all four execution modes; scoped vet passes.
Mutants: twelve source-rule plus one regex, five raw-question, four registry-retention, JSX refusal, three reporter-ID, two reporter guards and forty-eight named-metadata mutations caught.
Uncovered: native JSX source parsing, three source callbacks/verdicts/span discovery, full shared registration, Unicode quoting, custom options, inherited 26-rule/full repository gate; no new claims.

## Corrected contract and real blocker

The user's corrected contract uses the pinned typescript-go ast.Kind names, exactly as the shared registry validates them. No numeric adapter is needed. ParseNode.kind being a string is compatible with this contract. The twelve completed modules now export readonly string listenerKinds metadata. The active rule.json kinds are fragments [JsxFragment, JsxElement, JsxSelfClosingElement], constructed-context [JsxOpeningElement, JsxSelfClosingElement] and undef [JsxOpeningElement, JsxSelfClosingElement]. All three active descriptors include node: true. These descriptors remain explicitly partial-reporting, and are not presented to the full shared registry as complete source modules.

SuppliedNode now contains only prepared start/end positions. It had no meaningful use for a numeric kind; that field and all invented numeric fixture arguments were removed. Reporting functions still take their supplied range object and perform no parser refetch or relevance dispatch. This is reporting groundwork, not a completed native source callback. Existing twelve suite runners are unchanged apart from metadata; this maintenance does not claim that they have migrated into shared per-kind callback dispatch.

The fresh fetch leaves main and area/stage1-lint at c01907a70. Announced shared harness ab70f38d47de1d4974082b38f84a56af2368b7af is still not an ancestor of main. Its registry declares Kinds as strings, validates Kind.String() names, and node: true causes visit(node, index[, parent]). That interface was read, not edited. Its name contract replaces the earlier numeric metadata assumption.

The actual source blocker is current main's native JSX parser. The production-positive React.Fragment control still exits 70 with empty stdout and parser slice expected GreaterThanToken, got SlashToken. The three active JSX claims need native JSX parsing before their source discovery and callback algorithms can be integrated and compared. Their existing reporting portions are pushed, but source ports remain ACTIVE and INCOMPLETE. The earlier HIR/SSA/capture/memo React batch remains PARKED. No additional claim is taken. No shared parser, harness, registration generator, diagnostic model, protected compiler file, submodule pin, main or area branch is changed.

## Independent metadata truth and failure probes

Owned kind_names.go is an external Go oracle, not an Adamic program or checker bridge verdict. It enumerates the same pinned syntax.Kind.String() names as the shared registry: 351 names. check_kinds.py reads that output, validates all twelve string exports and three named rule.json declarations, and checks the intended listener list. It does not claim full registry discovery for the incomplete source rules. The historical numeric check_listeners.py and its reports describe the earlier contract; current validation uses check_kinds.py.

For each of fifteen declarations, three mutations are rejected: an unknown kind name, a numeric value and a valid AST name for the wrong listener. Three additional probes set node: false and lose the required supplied-node flag. Total forty-eight metadata mutations. This proves metadata checks, not source verdicts or runtime callback performance. kind-summary.json records every mutation separately.

## Commands and outputs

All toolchain commands source /workspace/adamic-tools/env.sh. bash cloud/setup.sh PASS: Go 0s, clang 0s, Node 0s, submodules 0s, cache 41s, total 41s; nproc 5, quota four cores, memory 17.6 GB. Go 1.27.1, clang 20.1.8, Node 24.19.0. Test output went directly to logs.

Root: go test -v -count=1 -timeout 30m ./stage1/cohere/typeaware -run '^TestWave12AgreementAndMutants$'. Other gates: the same go test options on wave12_next, wave12_third and wave12_fourth with '^TestAgreement$'. ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-12/corpus; per-batch COMPILER_MANIFEST and REPOSITORY_MANIFEST select the unchanged /workspace/wave-12 frozen manifests. ARTIFACTS point to /workspace/wave-12/named/<batch>. PASS durations: root 115.898s, next 145.004s, third 109.085s, fourth 147.471s. Gate durations include builds and are not performance measurements.

Control findings 89/138/56/596 sum to 879 over 187/142/94/703 controls, totaling 1,126. Complete canonical bytes 63,342/86,333/25,557/320,890 agree with Go for findings, fixes and suggestions, normally and under ASan/UBSan/LSan. Root compiler/repository findings remain 5/1 with 8,014/18,903 bytes; other batches zero with 5,010/18,485 bytes. Artifact path length accounts for changed control-stream byte counts. The corpus roots were not expanded.

Final reporting command: PYTHONDONTWRITEBYTECODE=1 python3 stage1/cohere/typeaware/wave12_sixth/validate.py /workspace/wave-12/named/sixth-final. PASS fifteen findings (1 fragment, 1 undef, 13 constructed-context), 6,198 bytes across native, sanitized native, source Node and emitted JavaScript. Go supplies source locations and message-family selection to prepared reporting fixtures, so this establishes rendering only. Unicode quoting remains explicitly refused outside the documented ASCII/empty-name contract; nondefault options are uncovered. All final metadata probes execute in this same validation. Earlier intermediate reporting run is retained in its log, but the final summary is the evidence for this revision.

Scoped go vet passes on the root and next/third/fourth packages with empty output. No new Go checker question or handle was introduced; existing bridge and released-handle tests run in the full rule gates. The main/compiler sources and checker package are unchanged, so the previous c019 checker and uncached Node fixture results are retained as historical results rather than represented as new runs. Source/emitted Node directly revalidates the changed prepared reporting class.

## Every oracle mutation

Twelve original rule mutations remain successful compilations and exit 0 with empty stderr; only the Go byte comparison catches their output changes. They skip a redeclaration survivor; change regex g to y; increase the collection finding end by one; replace transitive write with direct write; reverse lost-timeout polarity; reverse shebang-reason polarity; reverse global binding polarity for Function, nonconstructor and wrapper rules; reverse the regex comment exception; reverse rest-symbol declaration presence; reverse optional dependency-path polarity. The additional shared-regex mutation changes Effect to EffectX and is caught at byte 8,348. All exact first differing bytes are in the gate logs and mutation-observations.json.

Five raw-question mutants erase ancestry declaration-file status, signature body presence, import target path, constructor declaration-file status and identifier declaration-file status. They also compile and exit 0 with empty stderr and fail only Go byte equality. Four registry-retention mutations change expected released-handle panic 70 into successful execution and fail the lifetime contract. Reversing raw JSX-presence detection fails the valid-control success requirement. These lifetime/refusal probes are separate from byte-only mutants.

Three final reporter-ID mutants compile, exit 0 with empty stderr and preserve finding count while changing canonical Go comparison bytes. Two reporter guards require exact panic-70 messages; bypassing each produces successful output and fails that refusal contract. The forty-eight named metadata probes are detailed above. No reporting mutation is presented as a qualifying complete JSX source-rule mutation.

## Fresh native and Go time

After all builds finished, three quiet alternating whole-process rounds per implementation and frozen corpus, with complete stdout compared every round. The command is testdata/benchmark_wave_12.py with the corresponding native/Go binary pair, compiler config/manifest and repository config/manifest. Native phase recording uses ADAMIC_TSGO_TIMING=1; Go records its own phases. Full command arguments, phase data, hashes and all rounds are retained. No prepared-rendering/full-source timing ratio is claimed.

| Batch | Corpus | Native median seconds | Go median seconds | Native / Go |
| --- | --- | ---: | ---: | ---: |
| first | compiler | 4.097 | 0.427 | 9.60 |
| first | repository | 0.617 | 0.113 | 5.47 |
| next | compiler | 2.826 | 0.412 | 6.86 |
| next | repository | 0.405 | 0.108 | 3.73 |
| third | compiler | 2.435 | 0.325 | 7.50 |
| third | repository | 0.397 | 0.099 | 4.02 |
| fourth | compiler | 2.901 | 0.407 | 7.13 |
| fourth | repository | 0.408 | 0.099 | 4.10 |

Native remains slower. This change corrects listener metadata and prepared reporting inputs; it does not establish shared callback migration or explain all overhead. The shared-table RegExp literal remains unchanged from the previous tested correction, and no hand-rolled matcher was introduced.

Evidence is under evidence/named: full compressed gate/command streams, final reporting and kind summaries, source hashes, shared-ref status, prepared driver, command exits, benchmark rounds and every mutation observation. Cleanup records name forty-two previous c019 reproducible ELF/archive artifacts (1,196,379,828 bytes) and thirty completed named artifacts (558,126,544 bytes); sources/logs and benchmark binary pairs were retained. No setup or test failed in this named-contract run. A report-only commit follows tested code 9fe207dc0 without altering executable sources. Only codex/typeaware-wave-12 is pushed.
