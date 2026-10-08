d61b40d3cb24022a72b6730d3d02052d7adeca37: 26 raw roots have representations covered by this unit: 24 signature roots and two Map construction roots; these are not 26 completely lowered compiler functions.
Built nullable/undefined and object signatures, concrete generic binder recovery, Map key specialization/context/brands, and defaulted object binding parameters.
Commits pushed: 8e3b5daf, 2f92582e, fd0e1225, b647eebb, fe2bd0e0, a53c8958, ba33e849, 3063d1a9, af7f71ab, d61b40d3.
Focused Node oracles pass in release native, ASan/UBSan native and JavaScript, with leak checks; 21 semantic mutants and 16 isolated production-rule mutants are caught; counts refreshed.
Not covered: the 175 uninstantiated T/T|undefined census declarations, 50 value-T sites owned by representations, the six reserved clock cases, and the pending cases listed below.

The requested base was resolved to origin/area/compiler b410340dc8f889b5799c3bc519117c63def3aa24. The explicit area base instruction took precedence over the generic main-base instruction. Census replay 9a1f14c5 was merged at 8e3b5daf. The requested non-null branch c41c0e06 was merged at ba33e849; the sole counts conflict retained both sets of rows, then the generator refreshed their order. Only codex/notyet-generic-returns-t was pushed; no PR was opened and no area or main branch was modified.

The original category is 100 function-T roots, 75 function-T|undefined roots and 50 value-T roots, counted from the raw CSV. The census lowers the first two examples as declarations with no concrete instantiation. They still correctly say NotYet at checker.ts:1985:14 and checker.ts:6626:18. Generic calls in the fixtures instantiate number, boolean, built strings, objects, arrays and callbacks; they do not erase T. Recovery of the checker's resolved mapper additionally handles binders hidden inside NonNullable<T>, indexed results, indexed callbacks and default type arguments. The existing class mapper bridge is reused through its guarded reflected pointer; nothing was copied from cohere. No signature representation is invented for an unbound parameter.

To make these generic declaration roots measurable as lowered, the replay needs actual checker-resolved instantiations or callers. Authorizing an arbitrary representation for unbound T would erase its proof. This unit leaves the stop rather than weakening it.

Coverage count method: filter phase=lowering and kind=NotYet; group by kind, where, reason and text; exclude a group only when every observation has blocked_by_kind. No repeated contextual observations are counted twice. The following signature categories total 24 roots:

| Exact raw reason | Roots |
| --- | ---: |
| a function returning undefined | 11 |
| a function returning void &#124; "skip" | 2 |
| a function returning void &#124; SolutionBuilder<EmitAndSemanticDiagnosticsBuilderProgram> | 1 |
| a function returning void &#124; SolutionBuilder<EmitAndSemanticDiagnosticsBuilderProgram> &#124; WatchOfConfigFile<EmitAndSemanticDiagnosticsBuilderProgram> | 1 |
| a function returning void &#124; WatchOfConfigFile<EmitAndSemanticDiagnosticsBuilderProgram> | 1 |
| a function returning void &#124; number &#124; Symbol | 1 |
| a function returning void &#124; undefined | 1 |
| a function returning object | 3 |
| a function returning object &#124; undefined | 1 |
| a function returning string &#124; object | 1 |
| a parameter that isn't a plain name | 1 |

The two counted Map roots are core.ts:19:52 (new Map<never, never>(), now no lowering findings) and parser.ts:10599:23 (implicit new Map() construction gains the contextual/asserted Map arguments). Parser's first next stop is a BinaryExpression at 10594:25; the construction also reaches the Map value-union representation stop owned by runtime. The additional core.ts:1543:17 root advances to Refused, a cast the runtime cannot check; it is reported separately and is not added to the 26 supported roots. Two extra review sites, debug.ts:420 and utilities.ts:645, are absent from this raw root table and are not added to the count. Category coverage is an inference from the representation rule and fixtures; replay observations are recorded separately below. Some raw category sites already reached earlier body/dependency stops before this unit, so this count is not a claim of 26 newly green functions.

All production edits authored for this unit stay in functions.go, generic.go, the key branch of object.go, map_key_brands.go, map_type_arguments.go, generic_signature_arguments.go and function_pattern_defaults.go. The non-null merge also imports its upstream IR, CLI, flow, native and JavaScript changes. This unit authored no runtime C helpers and retained none of the experimental nullable-union runtime helpers. New source fixtures are .a; .ts fixtures were imported by the explicitly requested merge.

The clock reservations were pushed before 05:00 MDT in 3063d1a95f33493ffd588b813cd0c5060aca0cb6:
stage3/clock-briefs/generic-returns-t-01.md
stage3/clock-briefs/generic-returns-t-02.md
stage3/clock-briefs/generic-returns-t-03.md
stage3/clock-briefs/generic-returns-t-04.md
stage3/clock-briefs/generic-returns-t-05.md
stage3/clock-briefs/generic-returns-t-06.md

Each brief records a post-non-null replay and exact stop, a compiler-reproducing minimal program with measured Node output, the fix location on ba33e849, fixture and own registration paths, named mutant and completion checks. These six kinds were not implemented by this worker.

Node oracle fixtures, with registration in the corresponding own notyet_*_test.go file:
internal/oracle/testdata/notyet_generic_explicit_returns.a
internal/oracle/testdata/notyet_generic_optional_return.a
internal/oracle/testdata/notyet_generic_return.a
internal/oracle/testdata/notyet_map_branded.a
internal/oracle/testdata/notyet_map_context.a
internal/oracle/testdata/notyet_map_never.a
internal/oracle/testdata/notyet_map_specialization.a
internal/oracle/testdata/notyet_signature_object.a
internal/oracle/testdata/notyet_signature_pattern_default.a
internal/oracle/testdata/notyet_signature_undefined.a
internal/oracle/testdata/notyet_signature_void_references.a
internal/oracle/testdata/notyet_signature_void_union.a

The two notyet_map_explicit_any{,_generic}.a reductions are refusal fixtures, held to their measured Node empty-map output and Adamic's explicit-any adaptation refusal. They are intentionally not registered as productive lowering fixtures.

Semantic mutants (each runs cleanly and disagrees with Node stdout in native and JavaScript, including leak checks):
Generic return: return -7 instead of the concrete numeric result; populate an absent nullable numeric result with 23; pass -7 instead of the concrete numeric callback argument.
Undefined signature: return a present object instead of undefined.
Void union: replace the implicit undefined result with skip; replace a builder result with undefined; replace a mixed result with boxed true; replace void|undefined with a present object.
Object signature: return an empty boxed object in reference, optional and textOrObject functions instead of their distinct reference kind/string/undefined results.
Map: return size 1 for never's empty map; change numeric stored values to -7 for contextual, specialized and branded-key maps.
Hidden generic binders: return -7 for NonNullable and default specializations; return present 19 from the indexed nullable result; replace the indexed callback's built string with wrong.
Object pattern default: select the default even for a supplied argument; bind prefix from suffix.
Total: 21 semantic mutants, all caught.

Isolated production-rule mutants, each using a Go source overlay and failing its specific fixture/check without a build failure:
Undefined-only signature rule removed: the fixture again says NotYet, a function returning undefined.
Map (10): unrepresented known arguments use unread cache keys; unrepresented identities collapse; contextual arguments ignored; contextual base interfaces ignored; never key handling removed; phantom string key assigned the object lane; non-void brand field accepted; collision with a boxed string member accepted; callable brand accepted; explicit-any refusal removed. The brand-field/member/callable cases fail TestMapKeyBrandGuards; the cache and any cases fail TestMapExplicitAnyRefused; the others fail their own Map fixture.
Generic mapper recovery (4): suppress recovery only in nonNull, nonNullDefault, field and watcher respectively. The fixture fails at NonNullable<T>, NonNullable<T>, Options[K]|undefined and WatchFactory<X,Y>[T] respectively.
Pattern selector: replace the incoming object with typed undefined. Both backends execute cleanly and fail stdout agreement.
Total: 16 isolated rule mutants caught.

Other experiments are not counted as successful proof: removing an earlier explicit-binding block survived stdout and specialization checks twice because the resolved mapper fallback made the block redundant; that block was removed from production. Bypassing the pattern default entirely caused UBSan null-member access and JavaScript exit 70; a separate clean selector mutant above establishes the semantic rule. An initial rule-mutant invocation hit an unconfigured Go command rather than running tests; it was discarded and rerun with the configured toolchain. The nullable CompilerOptionsValue candidate disagreed with Node in both backends on undefined equality when null was also present; all candidate production edits were removed, and this category remains NotYet.

Exact final relevant test commands (all stdout/stderr redirected to the named logs; no output piped, full package suite or full gate run):
source /workspace/adamic-tools/env.sh
go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/(notyet_|class_inheritance_generic.a$)|TestNotYetGenericReturnsMutants|TestExplicitGenericReturnMutants|TestExplicitNullableArgumentsHaveDistinctInstances|TestUndefinedSignatureMutant|TestVoidUnion(SignatureMutant|RepresentationMutants)|TestObjectSignatureMutant|TestMap(KeyMutants|ExplicitAnyRefused)' -count=1 -v > /tmp/notyet-final-oracles.log 2>&1
PASS, 1.086s on the final mapper implementation.
go test ./internal/lower -run 'TestGenericFunctionPolymorphicRecursionIsRefused|TestGenericUnionFixtureHasSeparateInstances|TestGenericJSONUnionArrayIsNotYet|TestInheritanceGeneric|TestMapKeyBrandGuards|TestNonNullAssertion' -count=1 -v > /tmp/notyet-explicit-lower-final.log 2>&1
PASS, 0.367s.
go test ./internal/lower -run 'TestParameterPropert|TestInheritanceKeepsNominalTupleDestructuring|TestIteratorDestructuringDoesNotLieAboutExhaustion|TestDestructuredMethodsCannotLoadOwnSlots' -count=1 -v > /tmp/notyet-pattern-lower.log 2>&1
PASS, 0.373s.
go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/(notyet_signature_pattern_default.a|defaults.a)$|TestPatternDefaultMutants' -count=1 -v > /tmp/notyet-pattern-tests.log 2>&1
PASS, 0.616s.
go test ./internal/oracle -run TestCountsAreRecorded -args -update-counts > /tmp/notyet-explicit-counts.log 2>&1
PASS, 23.157s after mapper finalization.
go test ./internal/oracle -run TestCountsAreRecorded -args -update-counts > /tmp/notyet-pattern-counts.log 2>&1
PASS, 25.608s after the pattern fixture.
A final stronger pattern fixture builds every captured label at runtime with join, so its ownership check does not depend on immortal string literals. All twelve productive unit fixtures and all twenty-one semantic mutants were then run together:
go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/notyet_|TestNotYetGenericReturnsMutants|TestExplicitGenericReturnMutants|TestExplicitNullableArgumentsHaveDistinctInstances|TestUndefinedSignatureMutant|TestVoidUnion(SignatureMutant|RepresentationMutants)|TestObjectSignatureMutant|TestMap(KeyMutants|ExplicitAnyRefused)|TestPatternDefaultMutants' -count=1 -v > /tmp/notyet-final-all-oracles.log 2>&1
PASS, 7.752s.
go test ./internal/oracle -run TestCountsAreRecorded -args -update-counts > /tmp/notyet-final-counts.log 2>&1
PASS, 25.233s.
The source selector mutant was rerun against these built labels: /tmp/notyet-pattern-selector-mutant-final.log, exit 1 from stdout disagreement in both backends, with no sanitizer failure.

Earlier counts refreshes also ran after each productive fixture commit. The first explicit-binding candidate failed counts on class_inheritance_generic.a:91:22; the redundant binding block was removed, and that fixture plus the complete required counts refresh passed.

Rule-mutant commands:
python3 /tmp/notyet-map-rule-mutants.py > /tmp/notyet-map-rule-mutants.log 2>&1
Each script case runs go test -overlay=<case>.json <package> -run <specific fixture/check> -count=1 -v -timeout 10m; all ten exited 1 at the intended assertion.
go test -overlay=/tmp/notyet-generic-binding-mutants/<hidden-binder|resolved-default|indexed-binder|indexed-callback>.json ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/notyet_generic_explicit_returns.a$' -count=1 -v > /tmp/notyet-generic-binding-mutants/<case>.log 2>&1
Each of the four exited 1 with the expected lowering failure.
go test -overlay=/tmp/notyet-pattern-selector-overlay.json ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/notyet_signature_pattern_default.a$' -count=1 -v > /tmp/notyet-pattern-selector-mutant.log 2>&1
Exited 1; stdout differs in both backends, with no UBSan failure.
The undefined-rule overlay log is /tmp/notyet-signature-undefined-rule-mutant.log; it exited 1 at the undefined fixture's Lower assertion.

Toolchain preparation ran export GOPROXY='https://proxy.golang.org|direct', then bash cloud/setup.sh > /tmp/notyet-generic-setup.log 2>&1, then sourced /workspace/adamic-tools/env.sh (the printed path). No setup workaround was needed. Timing lines: Go 0.082s; Node 0.151s; clang 0.877s; markdown dependency step 1.084s, ready at 1.565s; submodules 15.951s; go build ready 302.378s; tests deferred 302.487s; cache warm 302.488s; total 302.514s. nproc=5; cgroup cpu.max=400000 100000; Go 1.27.1, Node 24.19.0, clang 20.1.8. Adapted TypeScript was prepared by bash stage3/apply.sh /tmp/notyet-generic-adapted > /tmp/notyet-generic-apply.log 2>&1.

Public original-example replay commands on d61b40d3:
go run ./stage3/census/latent/replay -project /tmp/notyet-generic-adapted/src/tsc/tsc.ts -where /tmp/notyet-generic-adapted/src/compiler/checker.ts:1985:14 -kind NotYet -reason 'a function returning T' > /tmp/notyet-final-original-examples/T.json 2> /tmp/notyet-final-original-examples/T.log
Exit 0 means the requested stop was reproduced: NotYet, src/compiler/checker.ts:1985:14, a function returning T.
go run ./stage3/census/latent/replay -project /tmp/notyet-generic-adapted/src/tsc/tsc.ts -where /tmp/notyet-generic-adapted/src/compiler/checker.ts:6626:18 -kind NotYet -reason 'a function returning T | undefined' > /tmp/notyet-final-original-examples/optional.json 2> /tmp/notyet-final-original-examples/optional.log
Exit 0 means the requested stop was reproduced: NotYet, src/compiler/checker.ts:6626:18, a function returning T | undefined.

The batched survey uses the same census overlay and worker, built once per code revision. On af7f71ab it replayed 118 samples spanning 98 signature reasons and 328 signature roots; 58 samples reproduced their original signature stop, the others had no findings or reached the next named stop. The largest-first sample survey is below, with both examples where available. This is measurement data, not a claim that every category was implemented.

Stops exposed by merging checked non-null: builder.ts:2460:25 now reaches BinaryExpression at 2460:31; core.ts:706:29 now reaches an array of T at core.ts:1030:37; core.ts:2531:17 now reaches an array of T at 2538:27. The value/array representation stops remain with that owner.

Pattern replay on d61b40d3: utilities.ts:9699:5 changes from a parameter that isn't a plain name to an ElementAccessExpression at 9699:111. Node prints given:given/custom!, missing:missing/missing!, undefined:undefined/undefined!, then 3. Defaulted constructor patterns, array-pattern defaults, nested binding defaults, and a pattern beside ordinary defaulted parameters remain NotYet; this unit only adds the proven object-pattern case.

Seven Map review replays:
src/compiler/core.ts:19:52: no remaining lowering findings
src/compiler/core.ts:605:20: NotYet, src/compiler/core.ts:605:20, a Map whose keys aren't strings, numbers, booleans, objects, arrays, maps or functions
src/compiler/core.ts:1543:17: Refused, src/compiler/core.ts:1543:17, a cast the runtime can't check
src/compiler/utilities.ts:779:5: NotYet, src/compiler/utilities.ts:779:5, a Map whose keys aren't strings, numbers, booleans, objects, arrays, maps or functions
src/compiler/parser.ts:10599:23: NotYet, src/compiler/parser.ts:10594:25, a BinaryExpression with a value and a value
src/compiler/debug.ts:420:29: no remaining lowering findings
src/compiler/utilities.ts:645:20: NotYet, src/compiler/utilities.ts:648:24, a value of type __String

core.ts:605 and utilities.ts:779 remain generic declaration stops without instantiations; concrete-key fixtures pass. core.ts:1543 preserves the runtime-uncheckable cast refusal. The adapted debug.ts:420 already uses a concrete Record key, so its absence of findings is not proof that explicit any became accepted. Both direct and generic explicit-any fixtures remain Refused. utilities.ts:645 advances to the value of type __String at 648:24, which is the representations owner's slot.

Remaining work and rulings:
CompilerOptionsValue needs a distinct null sentinel plus correct null/undefined union comparisons. The attempted candidate exposed the existing binary equality rule forcing false for an undefined comparison when null is present. internal/lower/expression.go's binary function belongs to codex/notyet-binary (ownership checked); its typeOf/representation function belongs to notyet-representations. No edits to either were retained. Supporting only a signature while those value/comparison rules disagree with Node would silently miscompile.
ResolvedConfigFileName, ResolvedConfigFilePath and other brand aliases need their underlying brand adaptation/proof; the existing phantom-brand ruling permits void markers. This worker does not authorize runtime-erasure of a never-valued brand field, collisions with primitive members, callable brands or indexers. Map's branded string key helper proves the primitive lane narrowly and does not claim to solve ordinary branded value slots.
Readonly-to-writable views, casts the runtime cannot check, method destructuring losing this, unknown/any adaptation and invalid overload implementation results keep their existing refusals. No ruling to relax them was assumed. Signature-only acceptance does not authorize those bodies.
NodeArray and other generic container signatures frequently already advance to arrays of T/never, mutable/readonly refusals or closure stops; this unit does not count them as newly lowered. The remaining record, intersection, project/cache, closure and nonprimitive value cases below remain pending or with their named next stop. The six clock cases are reserved, even where a broader future intersection helper could handle several at once.

| Raw signature reason | Root sites | Example | Observation after non-null merge | Reservation/status |
| --- | ---: | --- | --- | --- |
| a function returning T | 100 | src/compiler/checker.ts:10284:30 | NotYet src/compiler/checker.ts:10284:30: a function returning T | not claimed as newly covered |
| a function returning T | 100 | src/compiler/checker.ts:1985:14 | NotYet src/compiler/checker.ts:1985:14: a function returning T | not claimed as newly covered |
| a function returning T &#124; undefined | 75 | src/compiler/checker.ts:25177:14 | NotYet src/compiler/checker.ts:25177:14: a function returning T &#124; undefined | not claimed as newly covered |
| a function returning T &#124; undefined | 75 | src/compiler/checker.ts:25222:14 | NotYet src/compiler/checker.ts:25222:14: a function returning T &#124; undefined | not claimed as newly covered |
| a function returning undefined | 11 | src/compiler/binder.ts:2458:14 | NotYet src/compiler/binder.ts:2468:19: a value of type HasJSDoc | signature representation covered |
| a function returning undefined | 11 | src/compiler/builder.ts:2460:25 | NotYet src/compiler/builder.ts:2460:31: a BinaryExpression with a value and a value | signature representation covered |
| a function returning U &#124; undefined | 10 | src/compiler/core.ts:33:17 | NotYet src/compiler/core.ts:33:17: a function returning U &#124; undefined | not claimed as newly covered |
| a function returning U &#124; undefined | 10 | src/compiler/core.ts:50:17 | NotYet src/compiler/core.ts:50:17: a function returning U &#124; undefined | not claimed as newly covered |
| a function returning any | 8 | src/compiler/commandLineParser.ts:2437:10 | NotYet src/compiler/commandLineParser.ts:2437:10: a function returning any | not claimed as newly covered |
| a function returning any | 8 | src/compiler/commandLineParser.ts:2467:17 | NotYet src/compiler/commandLineParser.ts:2467:17: a function returning any | not claimed as newly covered |
| a function returning T[] | 7 | src/compiler/core.ts:1018:17 | NotYet src/compiler/core.ts:1018:59: a value of type T | not claimed as newly covered |
| a function returning T[] | 7 | src/compiler/core.ts:1312:17 | NotYet src/compiler/core.ts:1313:25: an array of T | not claimed as newly covered |
| a function returning CompilerOptionsValue | 4 | src/compiler/builder.ts:1456:14 | NotYet src/compiler/builder.ts:1456:14: a function returning CompilerOptionsValue | not claimed as newly covered |
| a function returning CompilerOptionsValue | 4 | src/compiler/commandLineParser.ts:2990:10 | NotYet src/compiler/commandLineParser.ts:2990:10: a function returning CompilerOptionsValue | not claimed as newly covered |
| a function returning NodeArray<T> | 4 | src/compiler/factory/nodeFactory.ts:1169:14 | NotYet src/compiler/factory/nodeFactory.ts:1171:24: an array of never | not claimed as newly covered |
| a function returning NodeArray<T> | 4 | src/compiler/parser.ts:2594:14 | Refused src/compiler/utilities.ts:10655:6: a value of type T seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | not claimed as newly covered |
| a function returning U | 4 | src/compiler/core.ts:93:17 | NotYet src/compiler/core.ts:93:17: a function returning U | not claimed as newly covered |
| a function returning U | 4 | src/compiler/transformers/classFields.ts:869:14 | NotYet src/compiler/transformers/classFields.ts:869:14: a function returning U | not claimed as newly covered |
| a function returning readonly T[] &#124; undefined | 4 | src/compiler/checker.ts:20540:14 | NotYet src/compiler/checker.ts:20541:13: a BinaryExpression with a value and a number | not claimed as newly covered |
| a function returning readonly T[] &#124; undefined | 4 | src/compiler/commandLineParser.ts:3220:14 | NotYet src/compiler/core.ts:1750:25: a value of type unknown | not claimed as newly covered |
| a function returning ResolvedConfigFileName | 3 | src/compiler/program.ts:5128:17 | NotYet src/compiler/program.ts:5128:17: a function returning ResolvedConfigFileName | not claimed as newly covered |
| a function returning ResolvedConfigFileName | 3 | src/compiler/tsbuild.ts:177:17 | NotYet src/compiler/tsbuild.ts:177:17: a function returning ResolvedConfigFileName | not claimed as newly covered |
| a function returning T[] &#124; undefined | 3 | src/compiler/core.ts:926:17 | NotYet src/compiler/core.ts:926:59: a value of type T &#124; undefined | not claimed as newly covered |
| a function returning T[] &#124; undefined | 3 | src/compiler/core.ts:985:17 | NotYet src/compiler/core.ts:987:34: an array of T | not claimed as newly covered |
| a function returning object | 3 | src/compiler/checker.ts:25300:14 | NotYet src/compiler/checker.ts:25302:13: a BinaryExpression with a number and a boolean | signature representation covered |
| a function returning object | 3 | src/compiler/commandLineParser.ts:2714:17 | Refused src/compiler/commandLineParser.ts:2715:12: Object.fromEntries | signature representation covered |
| a function returning NonNullable<T> | 2 | src/compiler/core.ts:1909:12 | NotYet src/compiler/core.ts:1908:17: a Map of T | not claimed as newly covered |
| a function returning NonNullable<T> | 2 | src/compiler/core.ts:706:29 | NotYet src/compiler/core.ts:1030:37: an array of T | not claimed as newly covered |
| a function returning PackageJson[K] &#124; undefined | 2 | src/compiler/moduleNameResolver.ts:361:10 | NotYet src/compiler/moduleNameResolver.ts:361:10: a function returning PackageJson[K] &#124; undefined | not claimed as newly covered |
| a function returning PackageJson[K] &#124; undefined | 2 | src/compiler/moduleNameResolver.ts:379:10 | NotYet src/compiler/moduleNameResolver.ts:379:10: a function returning PackageJson[K] &#124; undefined | not claimed as newly covered |
| a function returning SortedReadonlyArray<T> | 2 | src/compiler/core.ts:1038:17 | Refused src/compiler/core.ts:1039:12: a cast the runtime can't check | not claimed as newly covered |
| a function returning SortedReadonlyArray<T> | 2 | src/compiler/utilitiesPublic.ts:303:17 | NotYet src/compiler/core.ts:805:1: overload 1 of sortAndDeduplicate with additional implementation type parameters | not claimed as newly covered |
| a function returning V | 2 | src/compiler/core.ts:519:17 | NotYet src/compiler/core.ts:519:17: a function returning V | not claimed as newly covered |
| a function returning V | 2 | src/compiler/moduleNameResolver.ts:1068:10 | NotYet src/compiler/moduleNameResolver.ts:1068:10: a function returning V | not claimed as newly covered |
| a function returning WatchFactory<X, Y>[T] | 2 | src/compiler/watchUtilities.ts:727:14 | NotYet src/compiler/watchUtilities.ts:727:14: a function returning WatchFactory<X, Y>[T] | not claimed as newly covered |
| a function returning WatchFactory<X, Y>[T] | 2 | src/compiler/watchUtilities.ts:803:14 | NotYet src/compiler/watchUtilities.ts:803:14: a function returning WatchFactory<X, Y>[T] | not claimed as newly covered |
| a function returning readonly T[] | 2 | src/compiler/program.ts:2778:14 | NotYet src/compiler/core.ts:805:1: overload 1 of sortAndDeduplicate with additional implementation type parameters | not claimed as newly covered |
| a function returning readonly T[] | 2 | src/compiler/utilitiesPublic.ts:1284:17 | no lowering findings | not claimed as newly covered |
| a function returning void &#124; "skip" | 2 | src/compiler/utilities.ts:10727:14 | NotYet src/compiler/utilities.ts:10704:9: a BinaryExpression with a value and a value | signature representation covered |
| a function returning void &#124; "skip" | 2 | src/compiler/utilities.ts:10743:14 | NotYet src/compiler/utilities.ts:10744:65: a void call used as a value | signature representation covered |
| a function returning ((...args: A) => R) &#124; undefined | 1 | src/compiler/core.ts:1522:17 | NotYet src/compiler/core.ts:1522:50: a value of type T | not claimed as newly covered |
| a function returning () => T | 1 | src/compiler/core.ts:1891:17 | NotYet src/compiler/core.ts:1892:9: a value of type T | not claimed as newly covered |
| a function returning (AssignmentExpression<EqualsToken> & { readonly left: GeneratedIdentifier; }) &#124; undefined | 1 | src/compiler/factory/utilities.ts:1673:17 | NotYet src/compiler/factory/utilities.ts:1673:17: a function returning (AssignmentExpression<EqualsToken> & { readonly left: GeneratedIdentifier; }) &#124; undefined | reserved clock case |
| a function returning (ConstructorDeclaration & { body: Block; }) &#124; undefined | 1 | src/compiler/utilities.ts:6759:17 | NotYet src/compiler/utilities.ts:6759:17: a function returning (ConstructorDeclaration & { body: Block; }) &#124; undefined | reserved clock case |
| a function returning (arg: A) => T | 1 | src/compiler/core.ts:1907:17 | NotYet src/compiler/core.ts:1908:17: a Map of T | not claimed as newly covered |
| a function returning (arg: T) => boolean | 1 | src/compiler/core.ts:2454:17 | NotYet src/compiler/core.ts:2455:13: a value of type T | not claimed as newly covered |
| a function returning A | 1 | src/compiler/debug.ts:268:21 | NotYet src/compiler/debug.ts:268:21: a function returning A | not claimed as newly covered |
| a function returning AnyValidImportOrReExport | 1 | src/compiler/utilities.ts:4376:17 | NotYet src/compiler/utilities.ts:4376:17: a function returning AnyValidImportOrReExport | not claimed as newly covered |
| a function returning AnyValidImportOrReExport &#124; undefined | 1 | src/compiler/utilities.ts:4381:17 | NotYet src/compiler/utilities.ts:4381:17: a function returning AnyValidImportOrReExport &#124; undefined | not claimed as newly covered |
| a function returning BuildInvalidedProject<T> | 1 | src/compiler/tsbuildPublic.ts:932:10 | NotYet src/compiler/tsbuildPublic.ts:934:5: a value of type ResolvedConfigFileName | not claimed as newly covered |
| a function returning BuildInvalidedProject<T> &#124; UpdateOutputFileStampsProject | 1 | src/compiler/tsbuildPublic.ts:1316:10 | NotYet src/compiler/tsbuildPublic.ts:2287:109: a rest parameter outside a nongeneric named function | not claimed as newly covered |
| a function returning CanonicalKey | 1 | src/compiler/commandLineParser.ts:4170:10 | NotYet src/compiler/commandLineParser.ts:4170:10: a function returning CanonicalKey | not claimed as newly covered |
| a function returning CapturedThis | 1 | src/compiler/transformers/es2015.ts:825:14 | NotYet src/compiler/transformers/es2015.ts:825:14: a function returning CapturedThis | not claimed as newly covered |
| a function returning ClassExpression &#124; ImmediatelyInvokedArrowFunction | 1 | src/compiler/transformers/esDecorators.ts:1125:14 | NotYet src/compiler/transformers/esDecorators.ts:1125:14: a function returning ClassExpression &#124; ImmediatelyInvokedArrowFunction | not claimed as newly covered |
| a function returning ClassNamedEvaluationHelperBlock | 1 | src/compiler/transformers/namedEvaluation.ts:101:10 | NotYet src/compiler/transformers/namedEvaluation.ts:101:10: a function returning ClassNamedEvaluationHelperBlock | not claimed as newly covered |
| a function returning ClassStaticBlockDeclaration &#124; Decorator &#124; PrivateIdentifierGetAccessorDeclaration &#124; ... 5 more ... &#124; undefined | 1 | src/compiler/checker.ts:47028:14 | NotYet src/compiler/checker.ts:47028:14: a function returning ClassStaticBlockDeclaration &#124; Decorator &#124; PrivateIdentifierGetAccessorDeclaration &#124; ... 5 more ... &#124; undefined | not claimed as newly covered |
| a function returning ClassThisAssignmentBlock | 1 | src/compiler/transformers/classThis.ts:33:10 | NotYet src/compiler/transformers/classThis.ts:33:10: a function returning ClassThisAssignmentBlock | not claimed as newly covered |
| a function returning Declaration & HasModifiers | 1 | src/compiler/checker.ts:6600:18 | NotYet src/compiler/checker.ts:6600:18: a function returning Declaration & HasModifiers | reserved clock case |
| a function returning EvaluatorResult<T> | 1 | src/compiler/utilities.ts:11311:17 | NotYet src/compiler/utilities.ts:11311:72: a value of type T | not claimed as newly covered |
| a function returning ExpressionWithTypeArguments & { expression: Identifier &#124; PropertyAccessEntityNameExpression; } | 1 | src/compiler/parser.ts:9568:22 | NotYet src/compiler/parser.ts:9568:22: a function returning ExpressionWithTypeArguments & { expression: Identifier &#124; PropertyAccessEntityNameExpression; } | reserved clock case |
| a function returning ExpressionWithTypeArguments & { readonly expression: Identifier &#124; PropertyAccessEntityNameExpression; } | 1 | src/compiler/utilities.ts:5216:49 | NotYet src/compiler/utilities.ts:5216:49: a function returning ExpressionWithTypeArguments & { readonly expression: Identifier &#124; PropertyAccessEntityNameExpression; } | reserved clock case |
| a function returning HasJSDoc &#124; undefined | 1 | src/compiler/utilities.ts:4812:17 | NotYet src/compiler/utilities.ts:4812:17: a function returning HasJSDoc &#124; undefined | not claimed as newly covered |
| a function returning ImmediatelyInvokedArrowFunction | 1 | src/compiler/transformers/esDecorators.ts:670:14 | NotYet src/compiler/transformers/esDecorators.ts:670:14: a function returning ImmediatelyInvokedArrowFunction | not claimed as newly covered |
| a function returning InferenceContext &#124; (T & undefined) | 1 | src/compiler/checker.ts:26263:14 | NotYet src/compiler/checker.ts:26263:14: a function returning InferenceContext &#124; (T & undefined) | not claimed as newly covered |
| a function returning InvalidatedProject<T> &#124; undefined | 1 | src/compiler/tsbuildPublic.ts:1341:10 | NotYet src/compiler/tsbuildPublic.ts:1212:47: a narrowed union member whose object tag cannot be checked with typeof; keep differently held object kinds in separately typed variables | not claimed as newly covered |
| a function returning MemberName &#124; (Expression & (NumericLiteral &#124; StringLiteralLike)) | 1 | src/compiler/utilities.ts:4171:17 | NotYet src/compiler/utilities.ts:4171:17: a function returning MemberName &#124; (Expression & (NumericLiteral &#124; StringLiteralLike)) | reserved clock case |
| a function returning MissingList<T> | 1 | src/compiler/parser.ts:3569:14 | Refused src/compiler/parser.ts:3570:22: a cast the runtime can't check | not claimed as newly covered |
| a function returning ModeAwareCache<T> | 1 | src/compiler/moduleNameResolver.ts:1119:17 | NotYet src/compiler/moduleNameResolver.ts:1120:24: a Map of T | not claimed as newly covered |
| a function returning ModifierToken<T> | 1 | src/compiler/factory/nodeFactory.ts:1550:14 | NotYet src/compiler/factory/nodeFactory.ts:1550:59: a value of type T | not claimed as newly covered |
| a function returning ModuleOrTypeReferenceResolutionCache<T> | 1 | src/compiler/moduleNameResolver.ts:1276:10 | NotYet src/compiler/moduleNameResolver.ts:1284:5: a BinaryExpression as a statement | not claimed as newly covered |
| a function returning MultiMap<K, V> | 1 | src/compiler/core.ts:1542:17 | Refused src/compiler/core.ts:1543:17: a cast the runtime can't check | not claimed as newly covered |
| a function returning NodeArray<NonNullable<T>> &#124; undefined | 1 | src/compiler/parser.ts:3492:14 | NotYet src/compiler/parser.ts:3495:40: an array of NonNullable<T> | not claimed as newly covered |
| a function returning NonRelativeNameResolutionCache<T> | 1 | src/compiler/moduleNameResolver.ts:1166:10 | NotYet src/compiler/moduleNameResolver.ts:990:5: a function inside a function (a closure) | not claimed as newly covered |
| a function returning PerDirectoryResolutionCache<T> | 1 | src/compiler/moduleNameResolver.ts:1078:10 | NotYet src/compiler/moduleNameResolver.ts:990:5: a function inside a function (a closure) | not claimed as newly covered |
| a function returning Queue<T> | 1 | src/compiler/core.ts:1569:17 | NotYet src/compiler/core.ts:1570:41: a call through ?. (an optional call) | not claimed as newly covered |
| a function returning R | 1 | src/compiler/expressionToTypeNode.ts:880:14 | NotYet src/compiler/expressionToTypeNode.ts:880:14: a function returning R | not claimed as newly covered |
| a function returning ResolvedConfigFilePath | 1 | src/compiler/tsbuildPublic.ts:559:10 | NotYet src/compiler/tsbuildPublic.ts:559:10: a function returning ResolvedConfigFilePath | not claimed as newly covered |
| a function returning ReusableDiagnosticMessageChain | 1 | src/compiler/builder.ts:1533:14 | NotYet src/compiler/builder.ts:1533:14: a function returning ReusableDiagnosticMessageChain | not claimed as newly covered |
| a function returning SolutionBuilder<T> | 1 | src/compiler/tsbuildPublic.ts:2261:10 | SkippedDependency src/compiler/tsbuildPublic.ts:429:1: a dependency function whose body has checker diagnostics (measurement skipped) | not claimed as newly covered |
| a function returning SyntheticSuper | 1 | src/compiler/transformers/es2015.ts:4823:14 | NotYet src/compiler/transformers/es2015.ts:4823:14: a function returning SyntheticSuper | not claimed as newly covered |
| a function returning T &#124; EmptyStatement &#124; undefined | 1 | src/compiler/factory/nodeFactory.ts:7163:14 | NotYet src/compiler/factory/nodeFactory.ts:7163:14: a function returning T &#124; EmptyStatement &#124; undefined | not claimed as newly covered |
| a function returning T &#124; Identifier | 1 | src/compiler/factory/nodeFactory.ts:7141:14 | NotYet src/compiler/factory/nodeFactory.ts:7141:14: a function returning T &#124; Identifier | not claimed as newly covered |
| a function returning T &#124; NumericLiteral &#124; StringLiteral &#124; BooleanLiteral | 1 | src/compiler/factory/nodeFactory.ts:7146:14 | NotYet src/compiler/factory/nodeFactory.ts:7146:14: a function returning T &#124; NumericLiteral &#124; StringLiteral &#124; BooleanLiteral | not claimed as newly covered |
| a function returning T &#124; StringLiteral | 1 | src/compiler/transformers/declarations.ts:835:14 | NotYet src/compiler/transformers/declarations.ts:835:14: a function returning T &#124; StringLiteral | not claimed as newly covered |
| a function returning T &#124; readonly T[] &#124; undefined | 1 | src/compiler/core.ts:1160:17 | Refused src/compiler/core.ts:1152:1: overload 1 of singleOrMany result T &#124; T[] cannot be served by implementation result T &#124; readonly T[] &#124; undefined | not claimed as newly covered |
| a function returning T1 & T2 | 1 | src/compiler/core.ts:1495:17 | NotYet src/compiler/core.ts:1495:17: a function returning T1 & T2 | not claimed as newly covered |
| a function returning TEntry &#124; undefined | 1 | src/compiler/transformers/utilities.ts:829:17 | NotYet src/compiler/transformers/utilities.ts:829:17: a function returning TEntry &#124; undefined | not claimed as newly covered |
| a function returning TOut | 1 | src/compiler/core.ts:1783:17 | NotYet src/compiler/core.ts:1783:17: a function returning TOut | not claimed as newly covered |
| a function returning TOut &#124; undefined | 1 | src/compiler/core.ts:1778:17 | NotYet src/compiler/core.ts:1778:17: a function returning TOut &#124; undefined | not claimed as newly covered |
| a function returning TPrivateEntry &#124; undefined | 1 | src/compiler/transformers/utilities.ts:855:17 | NotYet src/compiler/transformers/utilities.ts:855:17: a function returning TPrivateEntry &#124; undefined | not claimed as newly covered |
| a function returning TResult | 1 | src/compiler/factory/utilities.ts:1475:14 | NotYet src/compiler/factory/utilities.ts:1475:14: a function returning TResult | not claimed as newly covered |
| a function returning T[][] | 1 | src/compiler/core.ts:2531:17 | NotYet src/compiler/core.ts:2538:27: an array of T | not claimed as newly covered |
| a function returning Token<TKind> | 1 | src/compiler/factory/nodeFactory.ts:7157:14 | NotYet src/compiler/factory/nodeFactory.ts:7157:48: a value of type TKind &#124; Token<TKind> | not claimed as newly covered |
| a function returning TransformationResult<T> | 1 | src/compiler/transformer.ts:248:17 | NotYet src/compiler/transformer.ts:249:39: new an Identifier | not claimed as newly covered |
| a function returning TypeMapper &#124; (T & undefined) | 1 | src/compiler/checker.ts:26378:14 | NotYet src/compiler/checker.ts:26378:14: a function returning TypeMapper &#124; (T & undefined) | not claimed as newly covered |
| a function returning TypeOnlyAliasDeclaration &#124; undefined | 1 | src/compiler/checker.ts:4455:14 | NotYet src/compiler/checker.ts:4455:14: a function returning TypeOnlyAliasDeclaration &#124; undefined | not claimed as newly covered |
| a function returning U[] | 1 | src/compiler/core.ts:494:17 | NotYet src/compiler/core.ts:495:25: an array of U | not claimed as newly covered |
| a function returning U[] &#124; undefined | 1 | src/compiler/core.ts:320:17 | NotYet src/compiler/core.ts:323:18: an array of never | not claimed as newly covered |
| a function returning V &#124; undefined | 1 | src/compiler/transformers/utilities.ts:400:5 | NotYet src/compiler/transformers/utilities.ts:423:34: a value of type (EmitNode & { autoGenerate: AutoGenerateInfo; }) &#124; (EmitNode & { autoGenerate: AutoGenerateInfo; }) | not claimed as newly covered |
| a function returning V[] | 1 | src/compiler/transformers/utilities.ts:377:10 | NotYet src/compiler/transformers/utilities.ts:377:61: a value of type V | not claimed as newly covered |
| a function returning VisitResult<T> | 1 | src/compiler/checker.ts:2501:14 | NotYet src/compiler/checker.ts:2501:14: a function returning VisitResult<T> | not claimed as newly covered |
| a function returning WatchCompilerHost<T> | 1 | src/compiler/watch.ts:871:10 | NotYet src/compiler/watch.ts:872:34: a method call through a structural signature in a program with statics; use typeof the declaring class | not claimed as newly covered |
| a function returning WatchCompilerHostOfConfigFile<T> | 1 | src/compiler/watch.ts:923:17 | Refused src/compiler/watch.ts:934:18: a cast the runtime can't check | not claimed as newly covered |
| a function returning WatchCompilerHostOfFilesAndCompilerOptions<T> | 1 | src/compiler/watch.ts:955:17 | Refused src/compiler/watch.ts:965:18: a cast the runtime can't check | not claimed as newly covered |
| a function returning object &#124; undefined | 1 | src/compiler/utilities.ts:7786:17 | NotYet src/compiler/core.ts:1769:26: a value of type unknown | signature representation covered |
| a function returning readonly Resolution[] | 1 | src/compiler/program.ts:2231:14 | NotYet src/compiler/program.ts:2233:9: a value of type SourceFileOrString | not claimed as newly covered |
| a function returning readonly U[] | 1 | src/compiler/core.ts:399:17 | NotYet src/compiler/core.ts:403:19: a value of type U &#124; readonly U[] &#124; undefined | not claimed as newly covered |
| a function returning readonly U[] &#124; undefined | 1 | src/compiler/core.ts:350:17 | Refused src/compiler/core.ts:342:1: overload 1 of sameMap result U[] cannot be served by implementation result readonly U[] &#124; undefined | not claimed as newly covered |
| a function returning string &#124; object | 1 | src/compiler/factory/nodeFactory.ts:7238:10 | NotYet src/compiler/factory/nodeFactory.ts:7244:13: a method call through a structural signature in a program with statics; use typeof the declaring class | signature representation covered |
| a function returning unknown | 1 | src/compiler/utilities.ts:9399:17 | NotYet src/compiler/utilities.ts:9399:17: a function returning unknown | not claimed as newly covered |
| a function returning void &#124; SolutionBuilder<EmitAndSemanticDiagnosticsBuilderProgram> | 1 | src/compiler/executeCommandLine.ts:812:10 | NotYet src/compiler/executeCommandLine.ts:833:16: a method call through a structural signature in a program with statics; use typeof the declaring class | signature representation covered |
| a function returning void &#124; SolutionBuilder<EmitAndSemanticDiagnosticsBuilderProgram> &#124; WatchOfConfigFile<EmitAndSemanticDiagnosticsBuilderProgram> | 1 | src/compiler/executeCommandLine.ts:756:17 | NotYet src/compiler/executeCommandLine.ts:763:13: a BinaryExpression with a string and a value | signature representation covered |
| a function returning void &#124; WatchOfConfigFile<EmitAndSemanticDiagnosticsBuilderProgram> | 1 | src/compiler/executeCommandLine.ts:563:10 | NotYet src/compiler/executeCommandLine.ts:579:16: a method call through a structural signature in a program with statics; use typeof the declaring class | signature representation covered |
| a function returning void &#124; number &#124; Symbol | 1 | src/compiler/binder.ts:2846:14 | NotYet src/compiler/binder.ts:2855:28: a BinaryExpression with a value and a boolean | signature representation covered |
| a function returning void &#124; undefined | 1 | src/compiler/sys.ts:517:21 | NotYet src/compiler/sys.ts:503:5: a value of type T | signature representation covered |
| a function returning { [P in K as `${P}`]?: T[]; } | 1 | src/compiler/core.ts:1464:17 | Refused src/compiler/core.ts:1461:28: overload 1 of groupBy type parameter U cannot satisfy implementation parameter K | not claimed as newly covered |
| a function returning { modifiers: NodeArray<Modifier> &#124; undefined; referencedName: Expression &#124; undefined; name: PropertyName; initializersName: Identifier &#124; undefined; descriptorName: Identifier &#124; undefined; thisArg: Identifier &#124; undefined; extraInitializersName?: never; } &#124; ... | 1 | src/compiler/transformers/esDecorators.ts:1236:14 | NotYet src/compiler/transformers/esDecorators.ts:1239:9: a value of type TNode | not claimed as newly covered |
| a function returning { readonly min: number; readonly max: number; } | 1 | src/compiler/utilities.ts:10362:17 | NotYet src/compiler/debug.ts:213:28: a value of type unknown | not claimed as newly covered |

All source-rule and semantic-mutant logs are scratch artifacts under /tmp as named above; fixture registrations and the semantic mutants are committed tests. A broad untracked-file cleanup was rejected by automatic review during the discarded nullable experiment; cleanup was retried with only the explicitly named worker-owned scratch files and succeeded. No approval remains pending.
