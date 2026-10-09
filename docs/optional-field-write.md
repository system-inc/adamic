# Optional presence rebuilt on main

Built roadmap step 17's optional-presence foundation on main 031a1259, carrying only this lane's net change from 7e7464e6.
Delivery branch: compiler/optional-presence; the delivery commit contains this report.
Validation: build, internal vet, changed .a checks, stage1 Gap/Gaps/Probes, stage3 fixtures, counts, and targeted uncached oracles pass.
Mutants: absent slot, copy presence, copy readiness, static keys, overlapping storage, representation tags, reservation, deletion, and admission boundaries all fail their tests.
Pending: checked-view optional writes await views-v1 integration; boolean optional writes, stricter options, and the stock fuzz driver's closure ABI repair are separate work.

## Layout decision

Presence uses insertion ranks, allocated in the same object allocation as values, readiness, and physical representation metadata. The layout is header, value slots, one size_t rank per slot, readiness bytes, representation bytes, alignment padding, and (for dynamic copies) the owned shape descriptor, names, tags, and reference flags. Rank zero means absent; nonzero ranks preserve string insertion order. Numeric keys enumerate in ECMAScript order.

A shared ready/present byte would save some storage but would not encode insertion order after deletion and reinsertion. Keeping ranks is the smallest sound adaptation of the existing design without introducing a second descriptor or copy mechanism. Main now has physical representation bytes as well as readiness bytes. Both byte tails remain separate from ranks, and dynamic descriptors begin after all three tails with explicit alignment.

Declared optional slots are reserved even when omitted, and start absent. Writes publish presence and readiness. Deletion clears presence; reinsertion assigns a new rank. Checked and dynamic copies preserve ranks, readiness, and physical representation bytes through the same field-copy path. Absent slots bypass checked reads; present unready slots retain the checked stop. Unknown views consult actual object state and dynamic shape metadata.

Key enumeration reads runtime presence, including for literal bindings. No binding-specific static key list survives alias deletion. Values, entries, spread, and assignment likewise exclude absent fields. Field writes use runtime lookup so deletion and reinsertion through another alias cannot invalidate a binding's assumed offset.

Object.hasOwn includes the small interface-annotation fix when a bounded const-alias walk proves a plain data literal origin. Getter, constructor, and unknown origins remain refused. Optional deletion also requires a proven plain data origin. Fresh spread and assignment admissions preserve main's existing representation and source-contract boundaries.

No changes were made to internal/native/emit.go, internal/lower/lower.go, internal/native/native.go, or internal/oracle/oracle_test.go. Hooks live in the existing object/library emitters and lower helpers, including internal/lower/class.go; the removed optional_write.go hook was not resurrected. No other worker's branch was merged.

## Views dependency

Read compiler/views-rehearsal's views-v1 4bb01786, views-v2 1bf89669, and views-v3 bc94c513 / 787cea7a, together with their integration notes. Views-v1 owns object contract frames and lazy reads; views-v2 covers mixed-union contracts; views-v3 covers array hooks. None supplies the optional-write source-slot certificate needed here.

The ordinary optional-presence implementation does not depend on those unlanded commits. The separate fixture optional_field_checked_view_pending.a is explicitly marked **awaits views-v1: checked-view optional-write integration**. Its Node observation is `0`; Adamic returns NotYet for the checked-view field. TestOptionalFieldCheckedViewPending verifies that boundary and skips it. It is neither a passing semantic fixture nor a counts row. The label names the object-view slice responsible for integration, not a claim that the existing views-v1 commit already implements it.

## Fixtures and mutants

Seven semantic fixtures cover optional writes, presence, construction, unknown views, checked copies, alias deletion, and reverse alias/entries/values/spread/reinsertion/hasOwn. They run against Node in JavaScript and the oracle's usual native builds, including sanitizers. The checked-copy fixture uses an explicitly unready class field rather than an inadmissible undefined assertion. Its intentional read-before-assignment stop is recorded as a checked observation; it is not claimed to equal raw Node's complete output.

Mutants executed and caught:

| Mutation | Catcher and observed failure |
| --- | --- |
| Drop omitted-slot reservation | TestOptionalFieldWriteCatchesDroppedSlot: native checked stop, missing field, instead of Node's value |
| Drop copied presence rank | TestOptionalFieldCopyState: copied presence and unknown-view state differ |
| Drop copied readiness | TestOptionalFieldCopyState: copied unready field becomes ready |
| Overlap readiness with rank storage | TestOptionalFieldCopyState: checked read of ready field stops |
| Drop copied representation bytes | TestOptionalFieldCopyState: expected physical tags become zero |
| Restore static key enumeration | TestOptionalFieldAliasCatchesStaticEnumeration: exit 0 but wrong output, in both native builds |
| Start omitted slots present, retain fixed write rank, or suppress deletion | TestOptionalFieldPresenceCatchesMutants: Node observation mismatch |
| Drop construction reservation | TestOptionalFieldConstructionCatchesDroppedReservation: missing-field stop |
| Replace optional lookup with required lookup | TestLiteralOptionalOracleCatchesMutant: missing-field stop; valid unreserved-slot control agrees with Node first |
| Remove plain-origin deletion guard | TestOptionalDeletionRequiresPlainStorage: class alias wrongly admitted |
| Replace hasOwn data-origin proof with names-only proof | TestObjectUnprovenShapesStayNotYet: getter interface alias wrongly admitted |

The static-key mutant prints exactly:

```
false
[first]
false
true
[first]
true
```

Node prints exactly:

```
false
[]
false
true
[first]
true
```

Inherited readiness, uninitialized/nullish, lazy-initializer, and import-cycle controls were also rerun. The last two admission mutants used scratch Go overlays; repository production code was not mutated in place.

## Commands and recorded results

Tool setup used `GOPROXY='https://proxy.golang.org|direct'`, `bash cloud/setup.sh`, and `/workspace/adamic-tools/env.sh`. Setup reported submodules 0.077s, Node 0.089s, Go 0.093s, markdown ready 0.214s, clang 0.467s, build 41.900s, deferred tests 42.027s, cache 42.028s, and completion 42.053s. `nproc` returned 5. Tools: Node 24.19.0, Go 1.27.1, clang 20.1.8.

Every test command wrote output directly to a log file.

| Command | Result / log |
| --- | --- |
| `go build ./...` | Pass; /tmp/optional-presence-build-complete.log |
| `go vet ./internal/...` | Pass; /tmp/optional-presence-vet-complete.log |
| `python3 /tmp/optional-presence-acheck.py` | All 9 changed/new .a files against origin/main pass; /tmp/optional-presence-acheck-complete.log |
| `go test ./stage1/... -run 'Gap\|Gaps\|Probes' -count=1 -timeout 30m` | Pass; /tmp/optional-presence-stage1.log |
| `go test ./stage3/fixtures -count=1 -timeout 30m` | Pass; /tmp/optional-presence-stage3-complete.log |
| `go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts` | Pass on Linux; /tmp/optional-presence-counts.log |
| `go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m` | Pass; /tmp/optional-presence-counts-complete.log |

The a-check harness uses the fetched devtools/fast-gate branch's cloud/fast-gate/run.py aCheck operation. It checks only paths from `git diff --name-only --diff-filter=ACMR origin/main -- '*.a'`.

Final focused runs:

```
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestOptionalField|TestLiteralOptionalOracleCatchesMutant|TestReadinessMutants|TestUninitializedIsNotNullishMutant|TestLazyInitializerIsNotEagerMutant|TestImportCycle|TestNativeAgreesWithNode/internal/oracle/testdata/(optional_field|non_null|import_cycles)' -count=1 -v -timeout 30m

go test ./internal/lower ./internal/native -run 'TestObject|TestUnknownAbsentSpreadPresence|TestOptionalDeletion|TestOptionalAssign|TestUniformFieldsMatchNode|TestRuntimeFieldLayoutsAreIncluded|TestRegexProgramsKeepCheckedFieldReads|TestOptionalWriteReservedSlotMatchesNode' -count=1 -timeout 20m
```

Both pass; logs are /tmp/optional-presence-oracle-complete.log and /tmp/optional-presence-layout-final.log. The oracle run has the one explicit views-v1 pending skip. Formatting and git diff checks pass. No full gate was run locally.

Counts add seven fixture rows and change zero existing rows. Stage3 regeneration has **no Compiles-to-Refused or Compiles-to-NotYet regression**. Only these stage0 outcomes/diagnostics change from NotYet to Compiles: objects/08_loop_state.a, objects/12_decorator_descriptor.a, and taste/17_binder_flow.a. A byte audit masks only stage0 objects and confirms every other byte, including Node observations, is unchanged. Log: /tmp/optional-presence-stage3-byte-audit.log.

## Generator observations and tooling limitation

The command exists on main. Its stock invocation was:

```
go run ./cmd/adamic-fuzz -with optional-field-write -seed 1 -count 200 -shrink=false -work /tmp/optional-presence-fuzz
```

It reports 200 findings, zero agreed, zero checked, zero NotYet, zero invalid, and zero unfit. The failures are ASan heap-buffer-overflows in captured closure allocation: generated C enables main's closure convention, while the driver links the default runtime archive with the smaller closure header. This is a driver/runtime selection mismatch, not a successful zero-panic run. Log: /tmp/optional-presence-fuzz.log.

A scratch-only overlay makes TryFile select `native.RuntimeLibraryForSource` from the generated C with `native.Options{Sanitize:true}`, and uses that selected archive for include paths and link flags. This is the runtime selector already used by usual native builds. No internal/fuzz source was changed in this branch. The exact corrected command was:

```
go run -overlay /tmp/optional-presence-fuzz-overlay.json ./cmd/adamic-fuzz -with optional-field-write -seed 1 -count 200 -shrink=false -work /tmp/optional-presence-fuzz-aligned
```

Result: **185 agreed, 15 checked, zero findings, zero NotYet, zero invalid, zero unfit**. There are zero unexpected panics. The 15 checked outcomes match inserted checks. Log: /tmp/optional-presence-fuzz-aligned.log.

For reproduction, the overlay maps internal/fuzz/run.go to a scratch copy. Immediately before TryFile's binary-path construction it inserts:

```go
runtimeLibrary, runtimeErr := native.RuntimeLibraryForSource(
    filepath.Join(c.Root, "internal", "native", "runtime"),
    string(lowered.Stdout), native.Options{Sanitize: true})
if runtimeErr != nil {
    return Outcome{Verdict: Finding, Key: "runtime", Detail: runtimeErr.Error()}
}
```

Within TryFile only, replace `filepath.Dir(c.runtime)` with `filepath.Dir(runtimeLibrary)` and `native.RuntimeLinkFlags(c.runtime)` with `native.RuntimeLinkFlags(runtimeLibrary)`. Prepare remains unchanged. The stock driver needs a separate integration fix before that unmodified command can serve as a green gate.

This unit does not implement native boolean/mixed-union optional writes (#p6a00tt), stricter options (#k881crd), general getter/constructor/unknown-origin hasOwn or deletion, or checked-view writes. It supplies their optional-presence foundation toward roadmap step 17. Darwin, WASI, and performance measurements were not covered.

## Pending views and binder-flow registration

On main c293674e the checked-view fixture still refuses at its optional write with `a checked field alias requiring an optional, accessor, or representation conversion`. Landed views-v1 handles required scalar writes; optional-field conversion remains unsupported. The pending test accepts that precise diagnostic and the earlier `checked view field slot of type number | undefined` boundary on this delivery branch. Neither is a passing backend fixture.

The stage3 status for taste/17_binder_flow.a already records Compiles. The stale NotYet promise was its separate oracle registration in taste_stage3_test.go; that registration now enables source Node comparisons in release native, ASan/UBSan native, JavaScript and WASI. All pass, including leak checks. Linux counts add only its row: 51 allocations, 51 frees, 0 retains, 51 releases, 7 maximum live, 0 released in regions. Node observations and stage3 status bytes are unchanged. Restoring the false registration through a scratch Go overlay reproduces `want stage 0 to refuse with where and what, got <nil>`.

Validation: `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestOptionalFieldCheckedViewPending|TestNativeAgreesWithNode/stage3/fixtures/taste/17_binder_flow' -count=1 -v -timeout 30m`; `ADAMIC_ORACLE_WASI=1 go test ./internal/oracle -run 'TestWASIAgreesWithNode/stage3/fixtures/taste/17_binder_flow' -count=1 -v -timeout 30m`; `go test ./stage3/fixtures -run 'TestFixtures/taste/17_binder_flow' -count=1 -v -timeout 30m -args -update`; and `go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts`. Logs: /tmp/optional-presence-two-reds.log, /tmp/optional-presence-binder-wasi.log, /tmp/optional-presence-binder-update.log and /tmp/optional-presence-two-reds-counts.log.
