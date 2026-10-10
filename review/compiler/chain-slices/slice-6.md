Built the step 21 per-use .ts narrowing checks through stored layout; .a writing-call refusals and exact uncaught diagnostics remain.
Merged current main 55347dce through 338052fe; member commits remain separate; refinement fix 354df3ff1a72932abf85bc8ca1f0d576c8182d2b is on top.
Node, sanitized/release native and JavaScript agreements/stops pass; current-main census: 14 .a agreements, zero .a witnesses, 1,314 files and no timeouts.
Per-use, whole-type, diagnostic and refusal mutants fail their intended checks; original runner classification failure is retained and independently verified.
Not covered: a complete native tsc binary, WASI execution, unresolved indirect effects or aggregate/accessor/nullable dynamic-property contracts, which retain NotYet boundaries.

# Chain slice 6

## Final refinement

Ruling: #bqj5drt's step 21, refined by @system_adamic Oct 9 18:5x. A property use reads the actual shape and stored slot tag before checking its demanded scalar result, including literal promises. Narrowing never changes storage representation. `.name` therefore accepts the replacement Error and `.flags` accepts every FlowNode's numeric flags, including numeric-enum bit combinations. Nullish receivers stop before dereferencing. Prototype methods/getters and unsupported aggregate contracts remain explicit compiler boundaries rather than unchecked conversions.

An argument is checked against its receiving contract. TypeError parameters require the TypeError identity; wider Error and unknown parameters accept the replacement Error. The failing parameter fixture deliberately stores the value as Error, proving that this check also works without unknown boxing. The original .ts fixture is now an agreement, exit 0 with `Error\nfinally\n`; its hash is unchanged. `narrowing-witness.json` supersedes the old 18:0x whole-type witness facts.

The original .a refusal fixture is unchanged and still refuses at :7:21 with `Adamic 0.1 refuses a narrowed read of value after change() can write it; narrow again after the call`. The non-writing .a control remains admitted and agrees with Node. Nominal identity is included when structural builtin Error/TypeError interfaces otherwise appear assignable. All fifteen main admissions removed by this policy retain exact refusal proof; see use-checks/admission-refusals.json.

The pinned TypeScript source audit found ten pairs at seven reads, 78 compiler files and zero semantic diagnostics. Every demanded result is true on each reachable normal-return path; the two Done assertions are unreachable. There is no failing consumed use to report as a tsc adaptation. The full reasoning per pair is in narrowing-sites.md, with raw final observations in use-checks/narrowing-sites-delivery.json. The finite admission CLI cannot express this all-return-path result property, so the proof is the requested source table. A reduced structural-layout and numeric-enum flags program is held to both backends; no full native tsc run is claimed.

## Witness facts for compiler/admission-delta

Standard TypeError and Error share their standard data members. The member witness therefore uses a TypeError subclass with its own readonly numeric detail member. The actual replacement Error lacks that member. No external witness-list support or entry was changed here.

| Fixture | SHA-256 | Violated declaration / narrowing / use | Node | Both backends, sanitized and release native |
| --- | --- | --- | --- | --- |
| internal/oracle/testdata/step21_narrow_member_use.ts | ef5188d026c069dc8a5f781ae4c18f2a5a589a08db3326f82c56ce0f401db0e2 | internal/oracle/testdata/step21_narrow_member_use.ts:2 / internal/oracle/testdata/step21_narrow_member_use.ts:5 / internal/oracle/testdata/step21_narrow_member_use.ts:7 | exit 0; `undefined\ncontinued\n` | exit 70; empty stdout; `adamic: panic: stale narrowing use failed: value.detail expected number\n` |
| internal/oracle/testdata/step21_narrow_parameter_use.ts | d9f55284b0b361af5566b34eb5064e2c5dcf4b9a41aa8f93607e806673650231 | internal/oracle/testdata/step21_narrow_parameter_use.ts:3 / internal/oracle/testdata/step21_narrow_parameter_use.ts:4 / internal/oracle/testdata/step21_narrow_parameter_use.ts:6 | exit 0; `Error\ncontinued\n` | exit 70; empty stdout; `adamic: panic: union member where the checker narrowed it away: a call since the narrowing put it back\n` |

Both entries use class `step 21: unsound narrowing across a writing call`, ruling 18:5x. Machine-readable routing facts are in narrowing-witnesses-18-5x.json.

## Extraction and members

| Member | Source SHA | Kept / dropped | Commit | Dependency and task evidence |
| --- | --- | --- | --- | --- |
| exceptions-21-main | 205586a0899488595200f3fca470874ef9282b37 | kept | df93806d | Main lexical readiness plus this slice's Read.Throws/Assign.Throws; no outside member symbol. Step 21, #dv99xzy and #agwccbw. |
| inherit-guards | 16a0b626b14b8fd6b541edc6cbb92e63b8ac55f5 | kept | bab6155e | Main already has void/never and inherited-layout guards; preserved with regression proof for #a898y40. |
| feature-set-link-main | e1efb527a314d834fc6fb282108e992558048eca | kept | d10e3e00 | Main RuntimeLibraryForSource/sourceFlags; this slice owns feature relocation/definition. #hmab710, Outcome 32 prerequisite. |
| lint-features | 3535387095d50f05d80e2717aed468843e59bc18 | kept | cd0b8e5f | Needs feature-set-link-main, kept in this same slice; matching program features reach lint's profile runtime. |

No member was dropped and no slice 1-4 change was copied. Source ranges, fork points, repairs and original dependency audits remain in this directory's extraction evidence and in the report at delivery commit 02deb61a. Repairs include 71972e18 lexical initialization throws, d8e6f59c payload/pending ownership, a6cb5660 cycle exit comparisons, c775dc61 readiness observations and aea87960 WASI exit ruling. Outside Read.Unset, Defined.Throws, namespace readiness and absent-message-shape changes were excluded; main's lexical model does not need them. The outside 2b2b1095 WASI outcome hunk needs ruled-backend-outcomes for backendDisagreement at internal/oracle/wasi_test.go:76 and was not copied. The member tasks have local delivery proof and await integration acceptance; Outcome 32 itself is not claimed complete.

## Commands and evidence

All final evidence paths below are under review/compiler/chain-slice-6/use-checks/. Test output went to logs. Commands sourced /workspace/adamic-tools/env.sh, unset GOCACHEPROG, and set GOMAXPROCS=4 and GOFLAGS='-buildvcs=false -trimpath -p=4'. Own oracle proofs used ADAMIC_GATE_UNCACHED=1. No whole-package test or full gate ran.

Setup used GOPROXY=https://proxy.golang.org|direct. Elapsed timing lines: go 0.051s; node 0.053s; submodules 0.113s; markdown ready 0.131s; clang 0.400s; shared cache 3.507s; build cache warm 159.041s; done 159.127s. nproc=5, cgroup quota four CPUs; Go 1.27.1, Node 24.19.0, clang 20.1.8. See setup.log.

| Proof | Command / result | Evidence |
| --- | --- | --- |
| Lower | run-lower.py partitions all 283 top-level tests into 15 exact selectors, -count=1 -json -timeout 90s; 281 pass, two pre-existing skips; largest shard 13.061s. | lower/lower-results.json, lower/lower-union.json, lower-delivery-run.log |
| Step 21 | go test ./internal/oracle -run '^TestStep21\|^TestNarrowedUnion' -count=1 -json -timeout 90s; pass 7.242s. | exceptions-delivery.jsonl |
| Member and flat admission fixtures | Exact TestNativeAgreesWithNode selector for step21_, lowering_chain_tdz_catch, inherited_fields_guard, arguments_length_value_count and the newly merged throwing getter; pass 19.543s. | admission-flat.jsonl |
| Nested admission agreements | Exact review/refused three-fixture and statements_small_stopped/structural_error selectors; pass. | admission-review.jsonl, admission-structural.jsonl |
| Current-main admission delta | Separate current-main and final-slice CLI builds; bounded 20s per file, four workers; 1,314 .a files, 14 agreements, zero witnesses, no timeouts; same fifteen exact refusals. | admission-census.py, admission-census-fix.json, admission-delta-fix.json, admission-refusals.json |
| Counts | One whole-table -update-counts pass, 97.461s. The subsequently expanded wider-parameter control was measured directly and its single row corrected; final non-writing TestCountsAreRecorded verification passed 77.015s. Six new rows and one existing getter row are fully attributed. | counts-update.log, wider-count.stderr, counts-verify.log, counts-attribution.json, getter-counts-c.patch |
| Call target reader guard | go test ./internal/ir -run '^TestCallTargetReaders$' -count=1 -timeout 90s; pass. | readers-delivery.log |
| Runtime feature link | Four exact feature selectors with the actual checker archive; 32 matching masks and mismatch controls; pass 15.845s. | features-delivery.jsonl |
| Lint profile | TestProfileRuntimeFeaturesMatchNode; pass 1.252s. | lint-delivery.jsonl |
| Lane | Required post-commit fetch/show lane-check pipeline: lane checks 10.7s; gofmt/tools on 77 Go files, t.Parallel on 13 test packages, a-check 5 .a files, vet 13 packages; all green. | lane-delivery.log |

Test leaf seconds are recorded in test-seconds.json; the new/touched per-use leaves are under 60s. Counts is the pre-existing aggregate writer. Existing lower skips are TestOriginalCycleLedger and TestOptionalWideningCensus. The public analyzer's unresolved indirect/bound-flow limitations remain documented in narrowing-sites.md.

## Mutants

| Mutant | Catcher |
| --- | --- |
| Drop narrowed-member result check | TestStep21NarrowedMemberUseStops: JavaScript admits undefined and continues; sanitized native reports invalid null-box access; release crashes. |
| Drop TypeError parameter check | TestStep21NarrowedParameterUseStops: all three executions admit Error, print Error then continued and exit 0. |
| Check whole narrowed type | TestStep21WritingCallTypeScriptAgreement rejects exit 70 instead of the valid .name agreement. |
| Drop wider/unknown receiving contract | TestStep21WiderParameterUseAgrees rejects the valid replacement Error at the unknown parameter. |
| Drop null receiver check | TestStep21NullPropertyUseStops: native's different dynamic-read panic and JavaScript's catchable TypeError differ from the exact inserted stop. |
| Drop literal result check | TestStep21LiteralPropertyUseStops: all backends admit 22 where the declaration requires 21. |
| Treat numeric enum as finite named values | TestStep21NarrowedFlagsUseAgrees rejects valid 12 bit-combination flags. |
| Drop native string length / scalar absence handling | StoredPrimitiveLengthUseAgrees and NarrowedUnionMemberCheck fail against Node / the exact owned stop. |
| backend-renders-stack / backend-drops-line | TestStep21Uncaught exact one-line diagnostic fails. |
| native-renders-stack / native-drops-line | TestStep21Uncaught exact one-line diagnostic fails. |
| Drop write-set test | TestStep21NonWritingCallControl wrongly refuses the control. |
| Admit writing .a call | TestStep21WritingCallRefusal receives nil instead of its exact refusal. |

Embedded per-use mutants are logged in exceptions-delivery.jsonl. Independent sources/logs are in mutants/, numeric-enum-mutant/, wider-call-mutant/ and scalar-mutant-verification.log. All original campaign metadata is preserved, including a caught=false marker-classification result for scalar-properties: its test actually failed the owned stderr assertion, independently verified without a build failure. mutant-proof-review.md explains both observations. Automatic approval review rejected a proposed evidence filtering/marker edit; it did not execute. Nothing was hidden or replaced. Earlier member campaigns remain historical evidence under exception-mutants/, feature-mutants/ and mutants/ at the parent delivery; changed whole-type witness conclusions are explicitly superseded here.

## Uncaught diagnostic and runtime re-look

The earlier runtime condition remains fulfilled: exactly `Uncaught Name: message` for actual Error instances through their shape, otherwise `Uncaught ` followed by undefined, null, number, boolean, string, function or object (arrays/maps included). CR/LF escaping and lone-surrogate U+FFFD encoding retain exactly one line. Native flushes stdout, writes/flushes stderr, releases the payload and exits 1; JavaScript synchronously writes identical bytes and exits 1. Node's own stack renderer remains excluded. docs/memory.md and the tab indentation fixes remain. The renderer itself moves no counts; the merged getter's extra pair comes from the exception member's checked Error read, not diagnostic rendering.

This refinement adds runtime/union.c primitive data-property handling. For @system_adamic_runtime, all runtime files changed against current main 55347dce are:

- internal/native/runtime/adamic.h
- internal/native/runtime/class_inheritance.c
- internal/native/runtime/exceptions.c
- internal/native/runtime/features.c
- internal/native/runtime/node_crypto.c
- internal/native/runtime/node_fs_file.c
- internal/native/runtime/node_host.c
- internal/native/runtime/regexp_replace.c
- internal/native/runtime/sort.c
- internal/native/runtime/sort_undefined.c
- internal/native/runtime/union.c

Runtime cleared 66ad4a67 subject to the supplied stderr condition; union.c is new since that clearance and needs the requested re-look. Delivery stays on compiler/chain-slice-6; no PR, main/area push or external witness-list edit.
