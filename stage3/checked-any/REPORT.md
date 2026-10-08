Built tagged .ts any transport, scalar use checks and safe dynamic property reads toward roadmap step 09.
Delivery is on codex/checked-unknown-any, based on area tip 4885cec50290686df487b62aac47c85d871ed40c; the delivery SHA is reported with the push.
Focused oracle fixtures, lower regressions, eight compatibility witnesses and the counts refresh pass.
Property/parameter guard mutants, null receiver mutants and eight independent source mutants are caught.
Independent coverage of the 271 sites is unmeasured; the hash-pinned stock entry stops before lowering, and dynamic calls and structural contracts remain unsupported.

The implemented subset

In .ts, any uses unknown's existing tagged representation. Typed string, number and boolean uses evaluate once, test the tag, and narrow only after success. This covers contextual scalar initializers, parameters and results, property results, and scalar arithmetic. Passing an any onward unchanged, including an assertion to any, preserves its tag. Explicit any in .a remains Refused. Existing proven intrinsic and evolving-object representations remain precise; they are not permission to trust an unverified dynamic object contract.

Failures are terminal exit 70, naming the source expression, required type and actual tag. Checks use ordinary IR in both backends. A captured mutation after a typeof narrowing is checked again before extraction. Dynamic property receivers reject null and undefined with the same named boundary. Strings expose their UTF-16 length; number/boolean own-property misses are undefined. Prototype members without native semantics are refused rather than falsely reported absent.

The implementation hooks are checkedAnyContext, checkedAnyType, checkedAnyOperands and checkedAnyPropertyReceiver in internal/lower/checked_any.go. expression.go calls the context/operand hooks, maps any to Union, preserves proven evolving objects and refuses erased any container layouts. unknown.go threads property receiver/result checks and preserves view hazards, including staged fields. refusals.go protects explicit .a any, unproven evolving .a bindings and any fields in typed objects. cast_proof.go/cast.go permit tag-preserving .ts assertions to any. collections.go preserves literal tuple inference rather than trusting synthetic any placeholders. union.c adds primitive own-property behavior. counts_test.go appends checkedAnyCounts after interfaceCastCounts. No lower.go, emit.go, native.go or oracle_test.go changes; no cohere code copied.

Programs still refused or NotYet include dynamic calls, structural/callable/literal/enum contracts, typed assertions from any, erased any arrays/tuples/maps/sets, any fields in typed objects, unsafe host views, getter/prototype reads, unsupported staged/nullable slots and dynamic element operations. Two dynamic operands for + or ordering need operation-specific dispatch. The scalar arithmetic contract rejects implicit coercion when the tag does not fit. Recursive JSON-plus-recovery contracts at typed structural boundaries are not implemented. This is partial step 09 progress, not completion of the requested general any mechanism.

Source witnesses and public validation

The inventory is ea1b2359:stage3/fixtures/checked-casts/outside-stock-casts.json: 112 explicit_any records plus 159 any_declaration records. The public ledger is 4e6121d3:stage3/scouts/step09/public-any/. Stock source is Microsoft TypeScript commit 050880ce59e30b356b686bd3144efe24f875ebc8; file hashes come from 855bcfaa:stage3/step09-ledger/files.json.

The fixtures are semantic reductions, not compilations of entire upstream functions. Property/transport/validation witnesses exercise the commandLineParser.ts convertToObject/JsonConversionNotifier family (2430, 2467) and convertJsonOption (3803). Scalar parameter/result/arithmetic witnesses isolate the typed-use obligations around any-bearing config conversion, including normalizeNonListOptionValue (3835); they do not claim a complete lowering of that upstream path. Assertion witnesses isolate tsc's as-any transport shape. Structural, callable and container witnesses demonstrate the current conservative limits rather than inventing callable or alias contracts.

config_diagnostics.a reduces the boolean branch of convertJsonOption (3803) and convertCompileOnSaveOptionFromJson (3725). Its typeof validation runs as source, preserving the diagnostic text: Compiler option 'compileOnSave' requires a value of type boolean. Node, JavaScript and native print exactly true, false and that diagnostic, with empty stderr and exit 0. No inserted failure fires. The reduced slice excludes nullable/list/path options and diagnostic locations. config_validated.a separately tests object, primitive, array, null and undefined input without a blanket object-only entry guard. Public declaration text is unchanged.

All repository witnesses are .a assets under internal/load/testdata/0.1/refuse/checked_any; the focused oracle writes a temporary .ts copy to exercise the .ts boundary. The same assets are explicitly checked to remain Refused as .a. Successful native witnesses run ASan/UBSan and leak checks. Misfits compare exact stdout, stderr and exit against the pinned terminal boundary; source Node observations are recorded separately, because intentional terminal checks differ from Node's unchecked coercion. The Node oracle adapter normalizes TypeError to its panic convention, so the null-source observation is exit 70, not a raw standalone Node exit.

Measurements

measure.py verifies all 80 ledger hashes and all 271 inventory records, then runs the stock tsc entry. Canonical generated diagnostics are produced with upstream's own generator and a relative input path (an absolute path changes its generated comment and hash). The entry emits zero C bytes and exits 1 at:

/tmp/checked-any-stock/src/compiler/binder.ts:1109:17: error TS2412: Type 'undefined' is not assignable to type 'FlowNode' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the type of the target.

Thus 0/271 sites are verified checked by whole-entry compilation. Independent per-site lowering is not measured: the loader returned no program. Every record in evidence/coverage.json is marked project_checker_stopped_before_lowering with this project first stop. This does not mean that none of those sites can lower independently. Fixture success is not used as a substitute for site coverage. There is no claimed revealed-byte measurement.

Reproduction (source the tool environment and export GOPROXY first):

    python3 stage3/checked-any/measure.py /tmp/checked-any-stock stage3/checked-any/evidence > /tmp/checked-any-measure.log 2>&1
    go test ./internal/oracle -run '^TestCheckedAny($|UnsupportedContracts$)' -count=1 -v > /tmp/checked-any-focused.log 2>&1
    go test ./internal/lower -run '^(TestTasteRepresentationLimitsStayExplicit|TestArrayPredicateCannotInventAnElementContract|TestArrayPredicateCoexistsWithUnknownReflection|TestObjectRefusalsExplainSoundness|TestUnknownReflectionRefusals)$' -count=1 -v > /tmp/checked-any-regression.log 2>&1
    go test ./internal/oracle -run '^TestNativeAgreesWithNode$/(internal|stage3)/(oracle|fixtures)/(testdata|taste)/(gaps|09_literal_cache|host_array_unknown_predicate|nested_destructured|neighbors|unknown_narrowing|unknown_narrowing_host)[.]a$' -count=1 -v > /tmp/checked-any-compatibility.log 2>&1
    go test ./internal/oracle -run '^TestNativeAgreesWithNode$/internal/oracle/testdata/switch_case_declarations/neighbors[.]a$' -count=1 -v > /tmp/checked-any-neighbors.log 2>&1
    go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts > /tmp/checked-any-counts.log 2>&1

The first oracle selector passes 16 witnesses and seven unsupported-contract witnesses. Its fitting results match Node; its misfits hit exact boundaries. The compatibility selector selects seven fixtures (including switch_empty_neighbors via Go's segment matching); the separate selector covers switch_case_declarations/neighbors. No whole package or full gate was run. Final counts refresh passes. Logs are retained under evidence/.

Mutants

The required property mutant removes checked_any_use guards from the lowered IR: the unchecked JavaScript prints Node's 14 instead of exiting 70. The required typed-parameter mutant likewise prints Node's wrong1. Independent arithmetic, assertion and stale-narrowing variants also reproduce Node's unchecked output and differ from the pinned failure. These mutants run the shared IR through the JavaScript backend; they are not claims about native ABI behavior after removing a guard.

Removing the null-receiver guard changes the complete pinned stderr in both JavaScript and native; each observation catches it. Separately, Go overlays introduce one source mutant at a time:

- array-layout: accept erased any element storage; array_view becomes accepted and its negative test fails.
- typed-field: omit the typed-object any-field refusal; field_view becomes accepted.
- primitive-prototype: omit the unsupported primitive descriptor refusal; prototype becomes accepted.
- explicit-any-policy: omit the .a AnyKeyword refusal; property fails its .a-policy assertion.
- literal-result: accept a number tag as proof of a literal result; literal_result becomes accepted.
- staged-field: omit the uninitialized-field hazard; staged_field becomes accepted.
- evolving-any: omit the unproven evolving .a refusal; TestTasteRepresentationLimitsStayExplicit/evolving_different_objects reports got nil.
- string-length: use UTF-8 bytes instead of UTF-16 units; string_length's native stdout is 5 instead of Node's 3.

Every listed source mutant fails its intended semantic catcher, independently. The recorded patches under mutants/ describe exact changes; evidence/mutant-*.log.txt records each failure. Build errors from exploratory mutant construction are not counted as catchers. To reproduce, apply one patch to a scratch worktree, run its named focused selector, then discard the patch; the actual runs used Go overlays and did not modify production sources.

Counts and toolchain

Sixteen new checked .ts-copy rows are added through the named counts hook. nested_destructured retains/releases change 7/17 to 8/18; switch_case_declarations/neighbors change 20/35 to 21/36. Generated C compared against the area base shows one extra retain/release pair for a string carried through the tag-preserving boxed view; allocations, frees and peak live counts are unchanged. Both fixtures match Node. logical_and_reference_maybe moves row position with identical values. The stale taste/17_binder_flow row disappears because that fixture already fails lowering on the area base. host_array_unknown_predicate initially exposed redundant scalar reboxing; eliminating it restores its original counts, so there is no final change to that row. Baseline binaries/logs support these comparisons.

Setup used GOPROXY=https://proxy.golang.org|direct and /workspace/adamic-tools/env.sh. Timing lines: node ready 0.022s; submodules 0.059s; markdown validation step 0.007s and ready 0.072s; clang ready 0.157s; go build 10.065s; tests deferred 10.287s; warm 10.289s; done 10.316s. nproc=5; CPU quota is four cores. Go 1.27.1, Node 24.19.0, clang 20.1.8. Setup passed. One counts attempt exhausted the 32GB disk because the Go cache reached 28GB; deleting 6,454,226,195 bytes of old rebuildable cache entries allowed the refresh to pass. No user source data was removed.
