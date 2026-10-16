Built: parked change chg_ws3ng9gpgkm7f18c9n12fygt1d recut onto current main as compiler/loom-chain-followers-2.
Commits: source 376febb44da1cc8018f1382f0bc488941375fcdc; base main 87ad5a4fe0b2ddf75adcde9d326e9d67edc1f2e4; final SHA in the delivery message.
Commands and outputs: 23 bounded focused commands completed, positive checks exit 0 and 14 source-mutant runs exit 1 at their named tests.
Mutants: all 21 caught, 14 source overlays plus seven native/JavaScript stdout mutants held against Node.
Not covered: full packages/full Loom gate, WASI, or dependency-blocked tasks already excluded from the authoritative source.

The branch begins at main 87ad5a4fe0b2ddf75adcde9d326e9d67edc1f2e4. Used git merge --squash 376febb44da1cc8018f1382f0bc488941375fcdc. Main was unchanged at the final fetch. One commit will be pushed to the new branch, with no force push, PR or Loom submission; the old branch is untouched.

Exactly two conflicts:

- internal/lower/exceptions.go: kept both adjacent switch cases in throwsOut. Main's c6e584809ceaf4822c96b9bf91d6ab15b6db6269, "Three lowering gaps, on main", landed #q9hz1bf's object-spread accessor throw edge. Its ir.ObjectLiteral case calls l.result.ObjectSpreadMayThrow(node), so throws from enumerable literal getters propagate through callers and cleanup. Kept that unchanged, and also kept the source's ir.PhantomMember case, found = found || !node.Optional, for a non-optional read that can throw a TypeError on undefined. Both types participate in throw propagation. The recorded landed diff is exceptions-landed.patch.
- internal/oracle/counts.md: took main with git checkout --ours, not the parked table. Regenerated once after the compiler resolution, measuring all fifteen source fixtures plus main's seven new witnesses. Main's 1022 existing rows retain identical values; exactly fifteen rows are added. TestCountsAreRecordedTable then passed. No prior failed regeneration or second regeneration occurred.

The automatic merges in internal/flow/build.go and internal/lower/expression.go retain main's object-spread flow edge, static constructor spread handling and nullish template work alongside the source's phantom brands, primitive concatenation and contextual generic function values. Existing tests, including main's, were not edited. No test was removed or ran 60 seconds or more; no split was needed.

Focused verification, all uncached with ADAMIC_GATE_UNCACHED=1, GOMAXPROCS=4, -parallel 4, -count=1 and -timeout 90s:

- counts-regeneration: exit 0, 27.365 command seconds.
- counts-table: exit 0, 3.043 command seconds.
- lower-focus: exit 0, 10.667 command seconds.
- oracle-focus: exit 0, 7.715 command seconds.
- ir-focus: exit 0, 12.283 command seconds.
- fixtures: exit 0, 17.927 command seconds.
- stage3: exit 0, 28.586 command seconds.
- readers: exit 0, 28.262 command seconds.
- vet: exit 0, 3.385 command seconds.

checks.json gives every exact argv and command wall time; focus.json names all tests and fixtures. Tests write stdout and stderr directly to per-command logs, never a pipe. Each subprocess has a 240-second hard limit, the runner a 900-second limit. The counts command measures only these 22 fixture leaves and preserves unmeasured rows. The oracle run independently verified exactly those same 22 leaves, including Node truth, JS backend, native release/sanitizers and leak checks. Source fixtures are fifteen; main witnesses are accessor_spread_throw_first/middle/last, plain_data_spread, static_constructor_spread_gap/owned and template_nullish. The seven changed stage3 status rows also passed, preserving their recorded Node bytes. The lower and IR checks include TestClassFeaturesThrowingAccessorSpread and TestObjectSpreadAccessorEffects, proving the kept main side of the exceptions conflict.

Reader guard and vet: TestCallTargetReaders ran with -timeout 90s; go vet ran on ./internal/flow ./internal/ir ./internal/javascript ./internal/lower ./internal/native ./internal/oracle. Both passed. Changed Go files are gofmt clean. Tests/products add no build-directory or machine-path literals; there is no new compiler-consuming package, so cloud/fast-gate/compiler-dependencies.json needs no change. No cohere code was copied; generic inference retains its pinned checker/shim reference.

Every test's seconds and outcome are recorded in test-times.tsv and test-times.json, including nested leaves, family wrappers and mutant failures. Maximum measured test seconds: 25.090. Compilation and process startup are included separately in the command seconds above. The inherited top-level arrow checks passed again; their source overlays fail before C emission, so invalid C or -Werror is not the catcher.

All 21 mutants and their catchers:

- non-null-check (source overlay): TestCheckedNonNullTypeScript/undefined (0.030s), TestCheckedNonNullTypeScript (0.000s).
- non-null-narrowing (source overlay): TestCheckedNonNullTypeScript/narrowed (0.030s), TestCheckedNonNullTypeScript (0.000s).
- nested-empty (source overlay): TestNestedEmptyArrayElementKinds/strings (0.100s), TestNestedEmptyArrayElementKinds/objects (0.050s), TestNestedEmptyArrayElementKinds (0.000s).
- string-concatenation (source overlay): TestNativeAgreesWithNode_081/internal/oracle/testdata/scanner_expressions/primitive_concatenation.a (0.030s), TestNativeAgreesWithNode_081 (0.000s).
- phantom-primitive (source overlay): TestPhantomCastsAreErased (0.040s).
- proven-predicates (source overlay): TestCensusComputedRelation (0.030s).
- generic-empty (source overlay): TestNativeAgreesWithNode_059/internal/oracle/testdata/generic_empty_array.a (0.030s), TestNativeAgreesWithNode_059 (0.000s).
- debugger (source overlay): TestDebuggerNativeEmitsNothing/.ts (0.050s), TestDebuggerNativeEmitsNothing/.a (0.050s), TestDebuggerNativeEmitsNothing (0.000s).
- generic-values (source overlay): TestGenericFunctionValuesHaveSeparateInstances (0.030s).
- generic-namespace-interaction (source overlay): TestGenericFunctionValueNamespaceDirectCallsKeepIdentity (0.030s).
- phantom-overload-interaction (source overlay): TestPhantomOverloadOnlyBrandsKeepMainProof (0.030s).
- proven-cleanup (source overlay): TestCensusGenericRefusalInsideClosure (0.030s).
- arrow-constructor-frame (source overlay): TestArrowFieldCaptureFrames (0.040s).
- arrow-receiver-capture (source overlay): TestArrowSuperCaptureFrames (0.050s), TestArrowStoredCaptureFrames (0.050s).
- empty-array-falsiness (IR output mutant): TestEmptyArrayRuntimeMutants/empty_array_is_falsy (0.560s).
- fresh-array-sharing (IR output mutant): TestEmptyArrayRuntimeMutants/fresh_number_array_is_shared (0.560s).
- generic-wrong-result (IR output mutant): TestGenericFunctionValueWrongResultMutant (2.760s).
- numeric-concatenation-spelling (IR output mutant): TestScannerConcatenationMutant (2.720s).
- boolean-concatenation-spelling (IR output mutant): TestScannerPrimitiveConcatenationMutants/boolean (0.440s).
- undefined-concatenation-spelling (IR output mutant): TestScannerPrimitiveConcatenationMutants/undefined (0.410s).
- null-concatenation-spelling (IR output mutant): TestScannerPrimitiveConcatenationMutants/null (0.470s).

The fourteen overlays are reconstructed by applying only their mutation delta, comparing each stored mutant to parked source 376febb4, then transplanting the delta onto the resolved current tree. No old full-file snapshot overwrites landed main behavior. Each source mutant exits 1 with a named failed test and no build failure. The seven in-test IR mutants require stdout-only disagreement with Node in both backends; sanitizer/build failure is not accepted. Full failed-test messages and the in-test assertions' evidence are in mutants/results.json and the logs.

Runtime files changed against main: none. Compiler implementation files changed:

- internal/flow/build.go
- internal/ir/phantom_brands.go
- internal/javascript/javascript.go
- internal/lower/cast_proof.go
- internal/lower/census_small.go
- internal/lower/class_inheritance.go
- internal/lower/empty_arrays.go
- internal/lower/exceptions.go
- internal/lower/expression.go
- internal/lower/generic.go
- internal/lower/generic_function_value.go
- internal/lower/invariance.go
- internal/lower/library_string.go
- internal/lower/object.go
- internal/lower/phantom_brands.go
- internal/lower/proven_relations.go
- internal/lower/refusals.go
- internal/lower/statements.go
- internal/native/emit_expressions.go
- internal/native/phantom_brands.go

Counts additions, allocations/frees/retains/releases/peak/regions:

| Fixture | Allocations | Frees | Retains | Releases | Peak | Regions |
|---|---:|---:|---:|---:|---:|---:|
| internal/oracle/testdata/arrow_this_super.a | 5 | 5 | 8 | 8 | 5 | 0 |
| internal/oracle/testdata/arrow_this_field.a | 7 | 7 | 9 | 11 | 6 | 0 |
| internal/oracle/testdata/arrow_this_stored.a | 7 | 7 | 9 | 11 | 7 | 0 |
| internal/oracle/testdata/native-generic-empty-array.a | 3 | 3 | 1 | 5 | 3 | 0 |
| internal/oracle/testdata/generic_empty_array.a | 32 | 32 | 24 | 60 | 17 | 0 |
| internal/oracle/testdata/generic_function_value_arms.a | 18 | 18 | 72 | 99 | 7 | 0 |
| internal/oracle/testdata/generic_function_value.a | 15 | 15 | 27 | 45 | 7 | 0 |
| internal/oracle/testdata/phantom_brands.a | 15 | 15 | 29 | 42 | 6 | 0 |
| internal/oracle/testdata/phantom_undefined.a | 1 | 0 | 5 | 4 | 1 | 0 |
| internal/oracle/testdata/phantom_catch.a | 8 | 8 | 10 | 18 | 2 | 0 |
| internal/oracle/testdata/host_phantom_overload_primitive.a | 25 | 25 | 13 | 43 | 7 | 0 |
| internal/oracle/testdata/host_phantom_optional_array.a | 5 | 5 | 2 | 7 | 2 | 0 |
| internal/oracle/testdata/scanner_expressions/string-number-concatenation.a | 2 | 2 | 3 | 6 | 2 | 0 |
| internal/oracle/testdata/scanner_expressions/number_spellings.a | 85 | 85 | 2 | 88 | 3 | 0 |
| internal/oracle/testdata/scanner_expressions/primitive_concatenation.a | 33 | 33 | 2 | 38 | 3 | 0 |

counts-attribution.json records all added rows and an empty changed-existing set. counts-main.md.txt is the main table before regeneration. New evidence lives under review/compiler/loom-chain-followers-2; earlier evidence remains historical under the original branch directories.

Setup: exported GOPROXY=https://proxy.golang.org|direct, ran timeout 600 bash cloud/setup.sh, and sourced the printed /workspace/adamic-tools/env.sh. It succeeded, including the real sanitizer overflow probe. nproc=5, cpu.max=400000 100000 (four-CPU quota). Cumulative timing lines: Go ready 0.144s; submodules ready 0.156s; Node ready 0.230s; Markdown ready 0.375s (validation step 0.034s); clang ready 0.750s; shared cache ready 1.889s; Go build ready 51.627s; test binaries deferred 51.706s; build cache warm 51.708s; done 51.743s. Toolchain: Go 1.27.1, clang 20.1.8, Node v24.19.0. Full timing lines and flags are in setup.log.

Conservative scope assumption: use the parked SHA 376febb4 as authoritative, carrying its compiler support and independent arrow tests. Its #agwccbw, #p6a00tt and #k881crd exclusions remain unchanged. Loom owns the full gate; this worker did not run whole packages or submit the change. Existing historical evidence bytes were preserved.
