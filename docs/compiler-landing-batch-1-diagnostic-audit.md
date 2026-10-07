# Stage 3 diagnostic audit

Compiler source audited at `39f07796474a2a28d186127d2a9bbe308c85e1b6`.

## Compile regressions first

None of these 18 records changes from Compiles to Refused or NotYet. No deliberate compile regression is being recorded.

Only two records are regenerated. Both change NotYet to Refused through codex/no-optional-widening (`226e757f`): absent optional members can conceal incompatible runtime fields. The refusal implementation is from `0a6a155ea0`, an ancestor of that exact branch tip.

Every recorded Node stdout, stderr and exit value stays byte for byte identical. Only stage0.outcome and stage0.what change; fixtures and all other status fields are untouched. The existing full stage 3 run found no Node or native runtime output disagreement.

## All 18 observations

New diagnostic means the observed compiler diagnostic, including records deliberately left unchanged.

| Fixture | Old outcome and exact diagnostic | New outcome and exact diagnostic | Decision and reason |
|---|---|---|---|
| assertions/14_structural_cache.a | NotYet: stage3/fixtures/assertions/14_structural_cache.a:11:16: stage 0 can't lower an ElementAccessExpression yet | Refused: stage3/fixtures/assertions/14_structural_cache.a:11:17: Adamic 0.1 refuses optional property iterationTypes in IterableOrIteratorType is absent from the source type Type; the value may hide an incompatible field; declare iterationTypes with a compatible type on the source type, or copy the declared fields into a new object (adamic/no-optional-widening) | Regenerated. New no-optional-widening refusal (226e757f); stronger than the previous unsupported lowering diagnostic. |
| objects/01_reference_spreads.a | Refused: stage3/fixtures/objects/01_reference_spreads.a:7:111: Adamic 0.1 refuses a spread after the first field; spread once, first: { ...source, field: value } (adamic/single-spread) | NotYet: stage3/fixtures/objects/01_reference_spreads.a:7:111: stage 0 can't lower a spread whose complete ordered own shape is not static (adamic/spread-not-first) yet | Left unchanged. spread-any-position (0c0cac93) removes the former single-spread refusal; this is outside the authorized strictness refresh. |
| objects/30_host_optional_method.a | NotYet: stage3/fixtures/objects/30_host_optional_method.a:6:12: stage 0 can't lower a call through ?. (an optional call) yet | Refused: stage3/fixtures/objects/30_host_optional_method.a:8:73: Adamic 0.1 refuses optional property getCompilerHost in ResolutionCacheHost is absent from the source type { name: string; }; the value may hide an incompatible field; declare getCompilerHost with a compatible type on the source type, or copy the declared fields into a new object (adamic/no-optional-widening) | Regenerated. New no-optional-widening refusal (226e757f); stronger than the previous unsupported lowering diagnostic. |
| objects/27_delete_substitution.a | Refused: stage3/fixtures/objects/27_delete_substitution.a:12:23: Adamic 0.1 refuses an index signature; use a Map, which keeps keys in the order they were added | NotYet: stage3/fixtures/objects/27_delete_substitution.a:12:23: stage 0 can't lower an index signature beside named members, or a non-mutable unrestricted string signature (dictionary storage cannot preserve named-property contracts) yet | Left unchanged. records-lowering (70fb62b1) removes the index-signature refusal and exposes a different downstream check or unsupported operation; not a stronger refusal from either authorized branch. |
| objects/17_container_rest.a | NotYet: stage3/fixtures/objects/17_container_rest.a:13:19: stage 0 can't lower destructuring anything but a tuple into [names] yet | NotYet: stage3/fixtures/objects/17_container_rest.a:13:42: stage 0 can't lower a structural view of library prototype member toString as an own field or method; use a wrapper object with an arrow that calls the member on its library receiver yet | Left unchanged. structural-library-view (6432f5e8) supplies a different NotYet reason; this does not establish a new or stronger refusal. |
| objects/10_build_options.a | Refused: stage3/fixtures/objects/10_build_options.a:14:54: Adamic 0.1 refuses an index signature; use a Map, which keeps keys in the order they were added | NotYet: stage3/fixtures/objects/10_build_options.a:14:54: stage 0 can't lower an index signature beside named members, or a non-mutable unrestricted string signature (dictionary storage cannot preserve named-property contracts) yet | Left unchanged. records-lowering (70fb62b1) removes the index-signature refusal and exposes a different downstream check or unsupported operation; not a stronger refusal from either authorized branch. |
| objects/05_polling_array_metadata.a | Refused: stage3/fixtures/objects/05_polling_array_metadata.a:12:23: Adamic 0.1 refuses a cast the runtime can't check; narrow it instead (===, typeof, a discriminant), or cast a discriminated union to its members (adamic/no-unchecked-cast) | NotYet: stage3/fixtures/objects/05_polling_array_metadata.a:19:7: stage 0 can't lower a value of type PollingIntervalQueue yet | Left unchanged. The unchecked-cast refusal disappears, exposing unsupported PollingIntervalQueue lowering; not a stronger refusal from either authorized branch. |
| objects/04_trace_metadata.a | Refused: stage3/fixtures/objects/04_trace_metadata.a:13:67: Adamic 0.1 refuses a spread after the first field; spread once, first: { ...source, field: value } (adamic/single-spread) | NotYet: stage3/fixtures/objects/04_trace_metadata.a:14:46: stage 0 can't lower JSON.stringify object references (structural types can hide fields and toJSON; runtime shapes need complete value metadata) yet | Left unchanged. spread-any-position (0c0cac93) removes the former single-spread refusal; this is outside the authorized strictness refresh. |
| objects/03_resolution_cache_spreads.a | Refused: stage3/fixtures/objects/03_resolution_cache_spreads.a:13:9: Adamic 0.1 refuses a spread after the first field; spread once, first: { ...source, field: value } (adamic/single-spread) | Compiles: (empty) | Left unchanged. spread-any-position (0c0cac93) removes the former single-spread refusal; this is outside the authorized strictness refresh. |
| records/01_has_property.a | Refused: stage3/fixtures/records/01_has_property.a:5:5: Adamic 0.1 refuses an index signature; use a Map, which keeps keys in the order they were added | Refused: stage3/fixtures/records/01_has_property.a:7:24: Adamic 0.1 refuses a method read as a value (hasOwnProperty would lose its object, and this with it); call it in an arrow that keeps the object: (v) => its object.hasOwnProperty(v) (unbound-method) | Left unchanged. records-lowering (70fb62b1) removes the index-signature refusal and exposes a different downstream check or unsupported operation; not a stronger refusal from either authorized branch. |
| records/16_built_strings.a | Refused: stage3/fixtures/records/16_built_strings.a:5:5: Adamic 0.1 refuses an index signature; use a Map, which keeps keys in the order they were added | Refused: stage3/fixtures/records/16_built_strings.a:7:24: Adamic 0.1 refuses a method read as a value (hasOwnProperty would lose its object, and this with it); call it in an arrow that keeps the object: (v) => its object.hasOwnProperty(v) (unbound-method) | Left unchanged. records-lowering (70fb62b1) removes the index-signature refusal and exposes a different downstream check or unsupported operation; not a stronger refusal from either authorized branch. |
| records/13_strict_option.a | Refused: stage3/fixtures/records/13_strict_option.a:4:72: Adamic 0.1 refuses an index signature; use a Map, which keeps keys in the order they were added | NotYet: stage3/fixtures/records/13_strict_option.a:4:72: stage 0 can't lower an index signature beside named members, or a non-mutable unrestricted string signature (dictionary storage cannot preserve named-property contracts) yet | Left unchanged. records-lowering (70fb62b1) removes the index-signature refusal and exposes a different downstream check or unsupported operation; not a stronger refusal from either authorized branch. |
| records/02_get_property.a | Refused: stage3/fixtures/records/02_get_property.a:5:5: Adamic 0.1 refuses an index signature; use a Map, which keeps keys in the order they were added | Refused: stage3/fixtures/records/02_get_property.a:7:24: Adamic 0.1 refuses a method read as a value (hasOwnProperty would lose its object, and this with it); call it in an arrow that keeps the object: (v) => its object.hasOwnProperty(v) (unbound-method) | Left unchanged. records-lowering (70fb62b1) removes the index-signature refusal and exposes a different downstream check or unsupported operation; not a stronger refusal from either authorized branch. |
| records/05_integer_order.a | Refused: stage3/fixtures/records/05_integer_order.a:5:5: Adamic 0.1 refuses an index signature; use a Map, which keeps keys in the order they were added | Refused: stage3/fixtures/records/05_integer_order.a:7:24: Adamic 0.1 refuses a method read as a value (hasOwnProperty would lose its object, and this with it); call it in an arrow that keeps the object: (v) => its object.hasOwnProperty(v) (unbound-method) | Left unchanged. records-lowering (70fb62b1) removes the index-signature refusal and exposes a different downstream check or unsupported operation; not a stronger refusal from either authorized branch. |
| records/06_delete_readd.a | Refused: stage3/fixtures/records/06_delete_readd.a:6:5: Adamic 0.1 refuses an index signature; use a Map, which keeps keys in the order they were added | Refused: stage3/fixtures/records/06_delete_readd.a:8:24: Adamic 0.1 refuses a method read as a value (hasOwnProperty would lose its object, and this with it); call it in an arrow that keeps the object: (v) => its object.hasOwnProperty(v) (unbound-method) | Left unchanged. records-lowering (70fb62b1) removes the index-signature refusal and exposes a different downstream check or unsupported operation; not a stronger refusal from either authorized branch. |
| records/07_optional_view.a | Refused: stage3/fixtures/records/07_optional_view.a:5:5: Adamic 0.1 refuses an index signature; use a Map, which keeps keys in the order they were added | Refused: stage3/fixtures/records/07_optional_view.a:7:24: Adamic 0.1 refuses a method read as a value (hasOwnProperty would lose its object, and this with it); call it in an arrow that keeps the object: (v) => its object.hasOwnProperty(v) (unbound-method) | Left unchanged. records-lowering (70fb62b1) removes the index-signature refusal and exposes a different downstream check or unsupported operation; not a stronger refusal from either authorized branch. |
| records/03_own_keys.a | Refused: stage3/fixtures/records/03_own_keys.a:5:5: Adamic 0.1 refuses an index signature; use a Map, which keeps keys in the order they were added | Refused: stage3/fixtures/records/03_own_keys.a:7:24: Adamic 0.1 refuses a method read as a value (hasOwnProperty would lose its object, and this with it); call it in an arrow that keeps the object: (v) => its object.hasOwnProperty(v) (unbound-method) | Left unchanged. records-lowering (70fb62b1) removes the index-signature refusal and exposes a different downstream check or unsupported operation; not a stronger refusal from either authorized branch. |
| taste/06_localized_message.a | Refused: stage3/fixtures/taste/06_localized_message.a:4:36: Adamic 0.1 refuses an index signature; use a Map, which keeps keys in the order they were added | NotYet: stage3/fixtures/taste/06_localized_message.a:6:12: stage 0 can't lower a BinaryExpression with a value and a string yet | Left unchanged. records-lowering (70fb62b1) removes the index-signature refusal and exposes a different downstream check or unsupported operation; not a stronger refusal from either authorized branch. |

## Skipped branches

Items 12, 13 and 20 were undone locally. No test or refusal was weakened.

### 12: codex/accessors f5bae9e0

The full flow fixture sweep fails on two coverage fixtures, including on the branch itself:

- `accessors_coverage_scalars.a:16:38: stage 0 can't lower a field of type boolean | undefined yet`
- `accessors_coverage_union_setter.a: stage 0 can't lower accessors sharing a name with different native representations yet`

### 13: codex/accessors-2 57c8a12e

After the small duplicate-implementation reconciliation, accessor lowering disagreed with the existing class architecture:

```text
--- FAIL: TestAccessorReadsAreVirtualCallsWithThrowEffects (0.11s)
    accessors_test.go:96: getter must be a virtual call to both targets with MayThrow: ir.Call{Function:7, Arguments:[]ir.Expression{ir.Read{Local:0, Of:4, Checked:false}}, Returns:1, Virtual:0, Accessor:"", Setter:false}
--- FAIL: TestAccessorUpdateSnapshotsReceiverBeforeGetterAndOperand (0.11s)
    accessors_test.go:113: update block: []ir.Statement{ir.Block{Body:[]ir.Statement{ir.Declare{Local:1, Value:ir.Read{Local:0, Of:4, Checked:false}}, ir.Evaluate{Value:ir.Call{Function:6, Arguments:[]ir.Expression{ir.Read{Local:1, Of:4, Checked:false}, ir.Binary{Operator:1, Left:ir.Call{Function:5, Arguments:[]ir.Expression{ir.Read{Local:1, Of:4, Checked:false}}, Returns:1, Virtual:0, Accessor:"", Setter:false}, Right:ir.NumberConstant{Value:2}}}, Returns:0, Virtual:0, Accessor:"", Setter:false}}}}}
--- FAIL: TestSuperAccessorsAreDirectCallsWithCurrentReceiver (0.10s)
    accessors_test.go:198: want both descriptor halves, got 0
--- FAIL: TestAccessorRefusals (0.23s)
    --- FAIL: TestAccessorRefusals/update_value (0.10s)
        accessors_test.go:64: want named reason accessor-update-value (NotYet=true), got /tmp/adamic-gate/TestAccessorRefusalsupdate_value3472008783/001/main.a:1:100: stage 0 can't lower a PostfixUnaryExpression yet
    --- FAIL: TestAccessorRefusals/inherited_static_receiver (0.08s)
        accessors_test.go:64: want named reason static-accessor (NotYet=true), got <nil>
    --- FAIL: TestAccessorRefusals/super_constructor_escape (0.06s)
        accessors_test.go:64: want named reason this escaping (NotYet=false), got /tmp/adamic-gate/TestAccessorRefusalssuper_constructor_escape3167066416/001/main.a:1:111: Adamic 0.1 refuses super.x escaping a constructor before every field is set (stored, passed, or a method called on it, which could read a field that holds undefined while its type says otherwise); assign every field first, then use this
    --- FAIL: TestAccessorRefusals/constructor_escape (0.04s)
        accessors_test.go:64: want named reason this escaping (NotYet=false), got /tmp/adamic-gate/TestAccessorRefusalsconstructor_escape1629907294/001/main.a:1:89: Adamic 0.1 refuses this.value escaping a constructor before every field is set (stored, passed, or a method called on it, which could read a field that holds undefined while its type says otherwise); assign every field first, then use this
    --- FAIL: TestAccessorRefusals/base_constructor_escape (0.05s)
        accessors_test.go:64: want named reason this escaping (NotYet=false), got /tmp/adamic-gate/TestAccessorRefusalsbase_constructor_escape3720515738/001/main.a:1:69: Adamic 0.1 refuses this.x escaping a base constructor before derived fields are initialized; use this only to read or write initialized base fields; call methods and publish the object after construction
    --- FAIL: TestAccessorRefusals/abstract (0.07s)
        accessors_test.go:64: want named reason abstract-accessor (NotYet=true), got <nil>
    --- FAIL: TestAccessorRefusals/computed (0.10s)
panic: Unhandled case in Node.Text: *ast.ComputedPropertyName [recovered, repanicked]

```

The branch itself passes its focused lower checks. Reconciling dispatch, descriptor representation, update evaluation order and refusal behavior was not a small clear fix.

### 20: codex/error-classes-counts 66766ab1

A small target-preservation interaction fix removed `panic: ir: virtual call has no target set`, but the existing IR call-target audit still rejects direct target reads. Exact remaining test output:

```text
ok  	github.com/system-inc/adamic/internal/flow	238.148s
ok  	github.com/system-inc/adamic/internal/fresh	90.252s
--- FAIL: TestCallTargetReaders (1.74s)
    call_targets_guard_test.go:166: unapproved call-target read internal/lower/error_library.go:checkedStringCall:Call.Function at /workspace/adamic/internal/lower/error_library.go:116:23; use CallTargets or ClosureTargets
    call_targets_guard_test.go:166: unapproved call-target read internal/lower/error_runtime_ranges.go:guardRuntimeRanges:Call.Function at /workspace/adamic/internal/lower/error_runtime_ranges.go:44:13; use CallTargets or ClosureTargets
    call_targets_guard_test.go:166: unapproved call-target read internal/lower/error_runtime_ranges.go:guardRuntimeRanges:Call.Function at /workspace/adamic/internal/lower/error_runtime_ranges.go:50:13; use CallTargets or ClosureTargets
    call_targets_guard_test.go:166: unapproved call-target read internal/lower/error_runtime_ranges.go:guardRuntimeRanges:ArraySort.Comparator at /workspace/adamic/internal/lower/error_runtime_ranges.go:59:12; use CallTargets or ClosureTargets
    call_targets_guard_test.go:166: unapproved call-target read internal/lower/error_runtime_ranges.go:stringLengthBounds:ArraySort.Comparator at /workspace/adamic/internal/lower/error_runtime_ranges.go:319:20; use CallTargets or ClosureTargets
    call_targets_guard_test.go:166: unapproved call-target read internal/lower/error_runtime_ranges.go:stringLengthBounds:Call.Function at /workspace/adamic/internal/lower/error_runtime_ranges.go:329:17; use CallTargets or ClosureTargets
    call_targets_guard_test.go:166: unapproved call-target read internal/lower/exception_bounds.go:refineExceptionBounds:Call.Function at /workspace/adamic/internal/lower/exception_bounds.go:151:52; use CallTargets or ClosureTargets
    call_targets_guard_test.go:166: unapproved call-target read internal/lower/exception_bounds.go:refineExceptionBounds:Call.Function at /workspace/adamic/internal/lower/exception_bounds.go:154:33; use CallTargets or ClosureTargets
    call_targets_guard_test.go:166: unapproved call-target read internal/lower/exception_paths.go:refineExceptionPaths:Call.Function at /workspace/adamic/internal/lower/exception_paths.go:157:87; use CallTargets or ClosureTargets
    call_targets_guard_test.go:166: unapproved call-target read internal/lower/exception_paths.go:refineExceptionPaths:Call.Function at /workspace/adamic/internal/lower/exception_paths.go:158:33; use CallTargets or ClosureTargets
    call_targets_guard_test.go:166: unapproved call-target read internal/lower/exception_paths.go:refineExceptionPaths:Call.Function at /workspace/adamic/internal/lower/exception_paths.go:204:35; use CallTargets or ClosureTargets
    call_targets_guard_test.go:166: unapproved call-target read internal/lower/exception_precision.go:preciseChecks:ArraySort.Comparator at /workspace/adamic/internal/lower/exception_precision.go:57:12; use CallTargets or ClosureTargets
    call_targets_guard_test.go:166: unapproved call-target read internal/lower/exception_precision.go:preciseChecks:Call.Function at /workspace/adamic/internal/lower/exception_precision.go:154:92; use CallTargets or ClosureTargets
    call_targets_guard_test.go:166: unapproved call-target read internal/lower/exception_precision.go:preciseChecks:Call.Function at /workspace/adamic/internal/lower/exception_precision.go:155:32; use CallTargets or ClosureTargets
    call_targets_guard_test.go:166: unapproved call-target read internal/lower/exception_precision.go:preciseChecks:Call.Function at /workspace/adamic/internal/lower/exception_precision.go:162:52; use CallTargets or ClosureTargets
    call_targets_guard_test.go:166: unapproved call-target read internal/lower/exception_precision.go:preciseChecks:Call.Function at /workspace/adamic/internal/lower/exception_precision.go:169:30; use CallTargets or ClosureTargets
    call_targets_guard_test.go:166: unapproved call-target read internal/lower/exception_precision_test.go:TestMayThrowPrecision:Call.Function at /workspace/adamic/internal/lower/exception_precision_test.go:45:59; use CallTargets or ClosureTargets
    call_targets_guard_test.go:166: unapproved call-target read internal/lower/exception_precision_test.go:TestLoopPresenceSurvivesRuntimeMutation:Call.Function at /workspace/adamic/internal/lower/exception_precision_test.go:238:58; use CallTargets or ClosureTargets
    call_targets_guard_test.go:166: unapproved call-target read internal/lower/iteration_test.go:TestCallbackTypedClassMethodCallBindsReceiver:CallClosure.Closure at /workspace/adamic/internal/lower/iteration_test.go:217:23; use CallTargets or ClosureTargets
FAIL
FAIL	github.com/system-inc/adamic/internal/ir	1.775s
?   	github.com/system-inc/adamic/internal/javascript	[no test files]
ok  	github.com/system-inc/adamic/internal/lower	47.083s
ok  	github.com/system-inc/adamic/internal/native	293.226s
ok  	github.com/system-inc/adamic/internal/oracle	319.434s
ok  	github.com/system-inc/adamic/stage1/cohere/graphql	141.540s

```

Migrating these readers to the shared target APIs needs a broader change, so the merge and attempted fix were undone.


## Validation and limits

- Fresh `git fetch origin` followed by `git merge --no-ff origin/area/compiler`: already up to date. No wait for pending area repairs.
- `go vet ./stage3/fixtures`: passed. No Go files changed, so gofmt has no input.
- `go test ./stage3/fixtures -run 'TestFixtures/(assertions|objects)/(14_structural_cache.a|30_host_optional_method.a)' -count=1 -timeout 40m`: passed, 1.851s.
- `go test ./stage3/fixtures -count=1 -timeout 40m -json`: exactly 16 unchanged stage0 diagnostic failures; 173 Node comparisons and 7 native comparisons passed. No Compiles-to-Refused/NotYet transition among the 18 changes.
- Diagnostic-record rollback mutant, run against a scratch fixture root: both newly recorded stricter refusals failed the stage0 comparison, exit 1. The repository records remained unchanged.
- Compiler source is unchanged since the final landing gate: `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -count=1 -timeout 40m` passed, 103.938s; `go test ./internal/flow ./internal/lower ./internal/native -count=1 -timeout 40m` passed, respectively 103.761s, 31.757s and 129.045s.
- Retained oracle mutant checks passed all 19 groups and 67 named subcases. The counts-row mutant was caught by TestCountsAreRecorded.
- Toolchain setup previously passed in 107s on this resumed workspace; nproc is 5.

The full stage3 fixture package remains red by explicit instruction: the 16 ineligible records are left intact. Mac-only area issues have not been tested on this Linux worker. Item 9 remains omitted under the dedicated typeof-null-2 worker instruction.
