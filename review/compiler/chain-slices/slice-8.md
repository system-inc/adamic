Built a main-only audit; neither member meets admission requirements.  
Base: `79f2067b13beb8603380fc993d188b22f10213c2`; audit commit: `58b66ce2ed80e934195ae0aa0f2498217b591da0`; audited source tips: `b6ef2cb4`, `57ea02f8`; no member commits retained.  
Commands and outputs: 23 lowering shards passed (maximum 35.540 s), reader guard passed (2.428 s), count regeneration passed (96.05 s), lane passed (3.9 s).  
Mutants: candidate results and catchers are recorded below; incomplete campaigns do not certify either dropped member.  
Not covered: complete optional mutant campaign, fresh 271-site lowering classification, full gate; no tasks closed.

Admission divergences are the first findings. These observations concern rejected compiler candidates. Evidence is under [chain-slice-8](../chain-slice-8/).

| Valid program | Node stdout / exit | JavaScript stdout / exit | Sanitized native stdout / exit | Cast, non-null assertion, declared-type lie |
| --- | --- | --- | --- | --- |
| [Catch after optional-chain read](../chain-slice-8/dependency-probe/optional-catch.a) | `"caught\n"` / 0 | `""` / 70 | `""` / 70 | None; callback clears the guarded field. |
| [80-level recursive tree](../chain-slice-8/depth-probe/source.a.txt) | `"passed\n"` / 0 | `""` / 70 | `""` / 70 | None; `any` transports actual values satisfying ConfigNode. |
| [NaN numeric field](../chain-slice-8/admission-final/observed-00/source.a.txt) | `"NaN\n"` / 0 | `""` / 70 | `""` / 70 | None; NaN is a TypeScript number. |

Exact stderr and all three observations: [catch](../chain-slice-8/dependency-probe/outputs.json), [depth](../chain-slice-8/depth-probe/outputs.json), [NaN](../chain-slice-8/admission-final/observed-00/outputs.json). Catch reports `TypeError: Cannot read properties of undefined (reading 'text')`. Depth reports non-JSON recursion at `raw` plus 64 `.next` segments plus `.value`. NaN reports non-JSON number at `raw.number`. These stop valid programs instead of refusing them before execution. They are not negative witnesses.

The initial optional extraction also stopped four valid owned fixtures because main's explicit undefined/null field metadata was not recognized as absent. All have no cast, non-null assertion or false declared type. A candidate repair recognized tags 12/13 and the owned differential fixture run then passed in 7.621 seconds. That repair was discarded with its member. Original native observations and independently rerun Node/JavaScript observations appear below and in [all-three.json](../chain-slice-8/divergences/all-three.json).

| Fixture | Node stdout / exit | JavaScript stdout / exit | Original native stdout / exit |
| --- | --- | --- | --- |
| field.a.txt | `"arg\nhost1:value2\narg\nreplacement:value2\nabsent\n"` / 0 | `"arg\nhost1:value2\narg\nreplacement:value2\nabsent\n"` / 0 | native: exit 70, stdout "arg\nhost1:value2\narg\nreplacement:value2\narg\n", stderr "adamic: panic: TypeError: optional call value is not callable\n" |
| optional-method.a.txt | `"absent\nabsent\nargument\nclass1:1\nabsent\nabsent\nargument\nliteral2:2\n2\n"` / 0 | `"absent\nabsent\nargument\nclass1:1\nabsent\nabsent\nargument\nliteral2:2\n2\n"` / 0 | native: exit 70, stdout "absent\nabsent\nargument\nclass1:1\nargument\n", stderr "adamic: panic: TypeError: optional call value is not callable\n" |
| runtime-bound-method-selection.a.txt | `"selected:argument1\nabsent\nmethod2:argument1\nmethod2:next\n"` / 0 | `"selected:argument1\nabsent\nmethod2:argument1\nmethod2:next\n"` / 0 | native: exit 70, stdout "selected:argument1\n", stderr "adamic: panic: TypeError: optional call value is not callable\n" |
| runtime-order.a.txt | `"absent\nselected2:arg1\n2 1\nselected3:arg2\nabsent\n2\nselected4:arg3\nabsent\n3\nfinished\n"` / 0 | `"absent\nselected2:arg1\n2 1\nselected3:arg2\nabsent\n2\nselected4:arg3\nabsent\n3\nfinished\n"` / 0 | native: exit 70, stdout "absent\nselected2:arg1\n2 1\nselected3:arg2\nabsent\n2\nselected4:arg3\n", stderr "adamic: panic: TypeError: optional call value is not callable\n" |

The checked-any candidate census compared 1,636 inputs with an exact-main overlay compiler built from `72ad75ef`: 50 newly admitted, zero admission regressions. Every newly admitted input was run on Node and both backends: 31 agree on stdout and exit; 19 diverge. The two valid checked-any divergences appear first above. The remaining 17 involve actual declared-result mismatches. No Ahra ruling was supplied, so none is accepted as a negative witness. Exact source bytes and all three outputs including stderr are in [observations.json](../chain-slice-8/admission-final/observations.json) and its observed directories.

| Remaining divergent candidate | Node stdout / exit | JavaScript stdout / exit | Sanitized native stdout / exit | Cast / ! / other declared-result mismatch |
| --- | --- | --- | --- | --- |
| [checked .ts copy: json_literal_misfit.ts](../chain-slice-8/admission-final/observed-04/outputs.json) | `"wrong\n"` / 0 | `""` / 70 | `""` / 70 | no cast; no !; actual result violates declaration; unruled .ts |
| [checked .ts copy: json_nullable_array_misfit.ts](../chain-slice-8/admission-final/observed-07/outputs.json) | `"passed\n"` / 0 | `""` / 70 | `""` / 70 | no cast; no !; actual result violates declaration; unruled .ts |
| [checked .ts copy: json_dictionary_misfit.ts](../chain-slice-8/admission-final/observed-11/outputs.json) | `"passed\n"` / 0 | `""` / 70 | `""` / 70 | no cast; no !; actual result violates declaration; unruled .ts |
| [checked .ts copy: staged_field.ts](../chain-slice-8/admission-final/observed-12/outputs.json) | `"undefined\n"` / 0 | `""` / 70 | `""` / 70 | no cast; !; actual result violates declaration; unruled .ts |
| [checked .ts copy: property_misfit.ts](../chain-slice-8/admission-final/observed-13/outputs.json) | `"14\n"` / 0 | `""` / 70 | `""` / 70 | no cast; no !; actual result violates declaration; unruled .ts |
| [checked .ts copy: json_union_misfit.ts](../chain-slice-8/admission-final/observed-20/outputs.json) | `"wrong\n"` / 0 | `""` / 70 | `""` / 70 | no cast; no !; actual result violates declaration; unruled .ts |
| [checked .ts copy: json_nested_list_misfit.ts](../chain-slice-8/admission-final/observed-24/outputs.json) | `"14\n"` / 0 | `""` / 70 | `""` / 70 | no cast; no !; actual result violates declaration; unruled .ts |
| [checked .ts copy: json_literal_result_misfit.ts](../chain-slice-8/admission-final/observed-28/outputs.json) | `"2\n"` / 0 | `""` / 70 | `""` / 70 | no cast; no !; actual result violates declaration; unruled .ts |
| [checked .ts copy: json_nullable_misfit.ts](../chain-slice-8/admission-final/observed-29/outputs.json) | `"passed\n"` / 0 | `""` / 70 | `""` / 70 | no cast; no !; actual result violates declaration; unruled .ts |
| [checked .ts copy: arithmetic_misfit.ts](../chain-slice-8/admission-final/observed-30/outputs.json) | `"wrong1\n"` / 0 | `""` / 70 | `""` / 70 | no cast; no !; actual result violates declaration; unruled .ts |
| [checked .ts copy: parameter_misfit.ts](../chain-slice-8/admission-final/observed-34/outputs.json) | `"wrong1\n"` / 0 | `""` / 70 | `""` / 70 | no cast; no !; actual result violates declaration; unruled .ts |
| [checked .ts copy: assertion_misfit.ts](../chain-slice-8/admission-final/observed-35/outputs.json) | `"wrong1\n"` / 0 | `""` / 70 | `""` / 70 | cast; no !; actual result violates declaration; unruled .ts |
| [checked .ts copy: json_list_misfit.ts](../chain-slice-8/admission-final/observed-38/outputs.json) | `"wrong\n"` / 0 | `""` / 70 | `""` / 70 | no cast; no !; actual result violates declaration; unruled .ts |
| [checked .ts copy: json_recursive_misfit.ts](../chain-slice-8/admission-final/observed-40/outputs.json) | `"passed\n"` / 0 | `""` / 70 | `""` / 70 | no cast; no !; actual result violates declaration; unruled .ts |
| [checked .ts copy: stale_misfit.ts](../chain-slice-8/admission-final/observed-43/outputs.json) | `"14\n"` / 0 | `""` / 70 | `""` / 70 | no cast; no !; actual result violates declaration; unruled .ts |
| [checked .ts copy: json_object_misfit.ts](../chain-slice-8/admission-final/observed-46/outputs.json) | `"wrong\n"` / 0 | `""` / 70 | `""` / 70 | no cast; no !; actual result violates declaration; unruled .ts |
| [checked .ts copy: json_dictionary_order_misfit.ts](../chain-slice-8/admission-final/observed-47/outputs.json) | `"passed\n"` / 0 | `""` / 70 | `""` / 70 | no cast; no !; actual result violates declaration; unruled .ts |

| Member | Source SHA | Kept or dropped | Dependency evidence / blocker | Tasks closed |
| --- | --- | --- | --- | --- |
| optional-calls-main | `b6ef2cb4` | Dropped | needs optional-chain-after-call-main (slice 3) for `Defined.Throws` and catchable failed-read emission at `internal/lower/optional_chain.go:468` and `internal/native/reuse.go:374`. Main emits a terminal panic for this IR. | None; #9v3v15w and #tvq1eqm remain open. |
| checked-any | `57ea02f8` | Dropped | No project-references symbol dependency found. Independently fails valid-program admission at `internal/native/runtime/checked_json.c:62,145` and `internal/javascript/checked_json.go:12,24` through unconditional domain/depth validation. | None; #mydv4kd remains open. |

The optional dependency is a symbol/IR semantic edge. [Candidate source](../chain-slice-8/optional-chain.go.txt) constructs the failed-read ir.Defined; the [outside helper](../chain-slice-8/optional-chain-after-call-defined.go.txt) and [outside native handling](../chain-slice-8/optional-chain-after-call-reuse.go.txt) show the missing semantics. None was copied into the delivery tree. Checked-any uses main's existing load.Load. Its probe dependency entry was split from project-references without importing that member. Checked-any was withheld for observed admission failures, not an invented project-reference edge.

Extraction and conflict audit:

- Read slice 1's PLAN, repair ledger, merge/commit histories and own-net checks, and slice 6's model report. Ledger snapshots are preserved in the evidence directory.
- Optional own range: `7a10c877..12dce3bf`: `28bf5bfb`, `05b062f9`, `6b3238d4`, `8f26a917`, `ce0e1476`, `12dce3bf`. Excluded lowering-a envelope `12132ea1`; took `b6ef2cb4`'s own flow parallelization. Candidate repairs: `aea87960` optional RegExp field/groups and own fixture, `a6cb5660` structural Map boundary. `72a2497d` closes a CSS gap whose exception prerequisite is absent on main; `1eaed07b`'s stale serial entry is already absent. These were not used to import outside work.
- Checked-any own range: `031a1259..57ea02f8`, including scalar `ba36dda2`, structural `11cbb0be`, coverage `b726ac3d` and recursive/nullable `57ea02f8`. Took only `2ca18b1b`'s checked-any probe dependency entry and `91b0148f`'s checked-any `.a` enumeration. No project selection or loader repair was copied.
- Kept main's cached object metadata preparation and reader guards, integrating candidate JSON flags into the cached walk. Repaired a nil result in the independently invoked JSON refusal pass; added parallel declarations to owning candidate tests. Resolutions are preserved in [checked-candidate-final.patch.gz](../chain-slice-8/checked-candidate-final.patch.gz); all candidate production hunks were removed.
- Started from main `72ad75ef`; fast-forwarded the audit branch to current main `59edda91`. Intervening commits add documentation and the getter census tests/fixtures/counts, with no production compiler change. The conservative choice is to withhold checked-any rather than publish its demonstrated valid-program stops. This unit closes no roadmap step; it records blockers toward steps 18 and 09.

Toolchain: set `GOPROXY='https://proxy.golang.org|direct'`, ran `bash cloud/setup.sh`, then sourced `/workspace/adamic-tools/env.sh`. Initial shared-cache/cold-build setup stalled and was stopped. Retry with `ADAMIC_GOCACHE_OFF=1` succeeded. Observed timing lines: Go 17.262 s; Node 17.368 s; markdown dependencies 17.566 s; submodules 17.575 s; clang 18.132 s; shared cache disabled 18.136 s; Go build 255.169 s; deferred 255.371 s; cachewarm 255.373 s; done 255.504 s. `nproc`: 5; tests used `GOMAXPROCS=4`. Go 1.27.1, Node 24.19.0, clang 20.1.8. `timeout 120 npm ci --prefix stage3/api` installed pinned Node types needed by baseline tests. [Setup log](../chain-slice-8/setup-retry.log).

Candidate verification, not delivery certification:

- `go test -p 4 ./internal/oracle -run '^(TestCheckedAny|TestCheckedJSON)' -count=1 -json -timeout 90s` passed in 17.903 s. Its depth/NaN pinned negative expectations violate admission agreement, so green does not establish agreement. All 23 candidate lower shards passed, longest 51.637 s.
- Source-only Step 18 passed in 0.654 s; the repaired owned optional differential fixture selector passed in 7.621 s. Native observations used ASan and UBSan. Exact commands/selectors are in the runner scripts and JSON logs.
- Fresh extraction validation passed: `verified 271 unchanged declarations and declaration-only contexts`. The fresh classification probe did not complete with a result. Historical classifications in the own-source patch are source evidence, not freshly measured coverage.
- Candidate count regeneration failed on `stage3/interface-downcasts/untagged/fixtures/class-data-good.a:1:98`, an explicit-any `.a` fixture outside this member. This is not a successful count refresh. Earlier attempts were stopped. Final main-only regeneration is recorded below.

Candidate mutants (none certifies a delivered feature):

| Mutant | Catcher or disposition |
| --- | --- |
| optional callable-reselect | candidate oracle runtime-order; exit 1; caught, no build failure |
| optional callable-null-guard | candidate oracle runtime-values; exit 1; caught, no build failure |
| optional callable-eager-arguments | candidate oracle runtime-arguments; exit 1; caught, no build failure |
| optional callable-result-zero | candidate oracle runtime-values; exit 1; caught, no build failure |
| optional callable-narrowing | candidate oracle runtime-order; exit 1; caught, no build failure |
| optional return-descriptor | candidate oracle runtime-return-descriptor; exit 1; caught, no build failure |
| storage receiver-lost | candidate field; logged CAUGHT in storage-mutants-final.log |
| storage field-reselected | candidate field; logged CAUGHT in storage-mutants-final.log |
| storage arguments-before-guard | candidate field; logged CAUGHT in storage-mutants-final.log |
| storage noncallable-no-loud-stop | candidate TestStep18StorageTypeLie/.ts; logged CAUGHT in storage-mutants-final.log |
| storage noncallable-stop-before-arguments | candidate TestStep18StorageTypeLie/.ts; logged CAUGHT in storage-mutants-final.log |
| storage adamic-noncallable-admitted | candidate TestStep18StorageTypeLie/.a; logged CAUGHT in storage-mutants-final.log |
| storage nullish-chain-arguments | candidate optional-method; logged CAUGHT in storage-mutants-final.log |
| storage getter-twice | candidate getter; logged CAUGHT in storage-mutants-final.log |
| checked-any alias_write | ^TestCheckedJSONNextRefusals$/json_dictionary_alias_write$; exit 1; caught |
| checked-any nul_key | ^TestCheckedJSONNextRefusals$/json_dictionary_nul_key$; exit 1; caught |
| checked-any proto_key | ^TestCheckedJSONNextRefusals$/json_dictionary_proto_key$; exit 1; caught |
| checked-any computed_key | ^TestCheckedJSONNextRefusals$/json_dictionary_computed_key$; exit 1; caught |
| checked-any symbol_field | ^TestCheckedJSONNextRefusals$/json_symbol_contract$; exit 1; caught |
| checked-any callable_contract | ^TestCheckedJSONNextRefusals$/json_callable_contract$; exit 1; caught |
| checked-any nullable_tag | ^TestCheckedJSONNext$/json_nullable$; exit 1; caught |
| checked-any optional_read | ^TestCheckedJSONNext$/json_nullable$; exit 1; caught |
| checked-any ownership_identity | ^TestCheckedJSONPreservesIdentityAndOperandWrites$; exit 1; caught |
| checked-any javascript_depth | ^TestCheckedJSONNext$/json_depth_misfit$; exit 1; caught |
| checked-any runtime_depth | ^TestCheckedJSONNext$/json_depth_misfit$; exit 1; caught |
| checked-any runtime_key_order | ^TestCheckedJSONNext$/json_dictionary_order_misfit$; exit 1; caught |
| checked-any runtime_record_guard | ^TestCheckedJSONRecordRemainsUnsupported$; exit 1; caught |
| checked-any array-layout | not executed: patch context did not apply; not credited |
| checked-any evolving-any | not executed: patch context did not apply; not credited |
| checked-any explicit-any-policy | not executed: patch context did not apply; not credited |
| checked-any literal-result | not executed: patch context did not apply; not credited |
| checked-any primitive-prototype | not executed: patch context did not apply; not credited |
| checked-any staged-field | not executed: patch context did not apply; not credited |
| checked-any string-length | not executed: patch context did not apply; not credited |
| checked-any typed-field | not executed: patch context did not apply; not credited |
| checked-any classification-source | ^TestRefusalMustBeInsideDeclaration$; exit 1; caught |
| checked-any classification-span | ^TestRefusalMustBeInsideDeclaration$; exit 1; caught |
| structural prepared-aliases | TestCheckedAnyUnsupportedContracts/json_unprepared; logged caught in checked-only-structural-mutants.log |
| structural array-method | TestCheckedAnyUnsupportedContracts/json_array_method; logged caught in checked-only-structural-mutants.log |
| structural object-fields | TestCheckedJSON/json_object_misfit; logged caught in checked-only-structural-mutants.log |
| structural array-elements | TestCheckedJSON/json_list_misfit; logged caught in checked-only-structural-mutants.log |
| structural literal-value | TestCheckedJSON/json_literal_misfit; logged caught in checked-only-structural-mutants.log |
| structural union-alternative | TestCheckedJSON/json_union_misfit; logged caught in checked-only-structural-mutants.log |
| structural finite-number | TestCheckedJSON/json_domain_misfit; logged caught in checked-only-structural-mutants.log |
| structural boolean-element | TestCheckedJSON/json_boolean_list; logged caught in checked-only-structural-mutants.log |
| structural recovery-tags | TestCheckedJSON/json_recovery; logged caught in checked-only-structural-mutants.log |
| structural alias-write | TestCheckedAnyUnsupportedContracts/json_alias_write; logged caught in checked-only-structural-mutants.log |
| structural alias-update | TestCheckedAnyUnsupportedContracts/json_alias_update; logged caught in checked-only-structural-mutants.log |
| structural reflection | TestCheckedAnyUnsupportedContracts/json_reflection; logged caught in checked-only-structural-mutants.log |
| structural iteration | TestCheckedAnyUnsupportedContracts/json_iteration; logged caught in checked-only-structural-mutants.log |
| structural callback-tags | TestCheckedJSON/stock_decode_entities; logged caught in checked-only-structural-mutants.log |

The depth and finite-number mutants catch the candidate's rejection policy, which itself disagrees with Node. Remaining optional/source/gap mutants were not completed after dropping the member; scripts and partial logs identify their inputs. No new check is delivered. Mutant sources under review use `.go.txt`, `.c.txt` or patches; no compilable Go consumer is added there.

Final proof results are recorded below. Runtime files changed in the delivered branch: none. @system_adamic_runtime has no runtime patch to clear.

Limits: no full package suite/full gate. The required first-green push within 30 minutes was missed during extraction and rejection. No PR was opened; this delivery is an audit rather than a completed feature slice.

Final lowering: `timeout 900 python3 review/compiler/chain-slice-8/main-proof/run-lower-shards.py`; 23 shards, all exit 0, longest 35.540 s. Each shard runs `go test -p 4 ./internal/lower -run <12 exact top-level Test names> -count=1 -json -timeout 90s`; selectors and seconds are in [lower-results.json](../chain-slice-8/main-proof/lower-results.json). Main refresh left the lower tree hash unchanged (`96039efde7f45738a59c67f4998e42827d16ae37`).

Reader guard after main refresh: `timeout 110 go test -p 4 ./internal/ir -run '^TestCallTargetReaders$' -count=1 -json -timeout 90s`; PASS, package 2.428 s. [Log](../chain-slice-8/main-proof/readers-current-main.jsonl). No Go test is added or edited by this delivery.

Counts: `timeout 360 go test -p 4 ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -json -timeout 300s -args -update-counts`; PASS in 96.05 s (existing aggregate test). One successful final refresh; the stale pre-main-refresh writer was stopped. No row changed. Every unchanged row belongs to current main; this slice introduces no fixtures/count rows. Maximum fixture leaf 6.83 s. [Attribution](../chain-slice-8/main-proof/count-attribution.json), [log](../chain-slice-8/main-proof/counts.jsonl).

Final admission delta: zero. The delivery modifies only review evidence and this report; all production source, runtime, fixtures, counts and compiler dependency declarations equal current main. Consequently the admission function and emitted outputs are unchanged for every input, including inputs outside the candidate census. Runtime files changed: none.

Lane command after committing: `git fetch -q origin main devtools/fast-gate cloud/merge-tree && git show origin/cloud/merge-tree:cloud/integration/lane-checks.py | python3 -`. Result is recorded with the delivery evidence.

Main refreshed again to `79f2067b` after the audit commit: speculative-census evidence and class-order test grain landed. This changes no production compiler/runtime or internal/oracle count registry/table, and leaves the lower tree unchanged. Main was merged only into this delivery branch. Initial lane output: `lane checks 3.9 s: gofmt and tools on 0 Go files, t.Parallel on 0 test packages; a-check 1 .a files`. [Lane log](../chain-slice-8/main-proof/lane-checks.log). The final evidence commit is checked again before push.

Final main reader guard: PASS, package 0.691 s; [log](../chain-slice-8/main-proof/readers-final-main.jsonl). Every fixture table row has an explicit [owner entry](../chain-slice-8/main-proof/count-row-owners.json). [Delivery identity](../chain-slice-8/main-proof/delivery-identity.json) records the zero non-review delta against current main. Main-refresh commit: `7ddf2199850698b52773af7289acd0f0c99fe894`.
