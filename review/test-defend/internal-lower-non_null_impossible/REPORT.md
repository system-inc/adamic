# Defense of optional widening and override signatures

Base: d29d80ceb5d5d42d9b0ffb7b528a57a2272c76f4. Audit sources are copied alongside this report. Fresh test discovery and every matrix run include the current package, not only historical audit rows.

## Code under test and oracle

TestOptionalWideningAllowed checks Adamic lowering's optionalValue/optionalWidened admission, including class-instance exemptions and fresh object rebuilding. Its oracle is self: three hand-authored fixtures must return no lowering error. It does not inspect their generated IR or execution output.

The override family checks checkMemberOverrides and overrideParameterRepresentation. Its self oracle requires NotYet, the mismatched parameter's name, and repair wording. The three members share assertOverrideParameterRefusal and count as one row. They differ from TestInheritanceNativeSignatureLimits because their mismatched parameter is the second parameter, factor, after value. The subsumer's mismatch examples concern the first parameter.

## Coverage and semantic differences

Four clean per-row coverage runs use -coverpkg ./internal/lower. Commands and outcomes are preserved in *.coverage.log and profiles in *.cover. Exclusive blocks are preserved in *.exclusive.txt. Optional admission has 968 exclusive covered blocks against the refusal row, mostly downstream lowering reached after admission; optional_widening.go's target-union and class-exemption branches are relevant leads. The override family has only seven exclusive covered blocks, none in class_inheritance.go. Its defense uses the semantic second-parameter difference on shared code, not exclusive lines.

## Attempts and results

D1 changes the diagnostic's selected parameter index from index to 0. The wrong diagnostic names value instead of factor. Only the three override family members fail; 271 other top-level tests pass. TestInheritanceNativeSignatureLimits passes. This defends the family among executed rows. The complete passed list is in matrix.json.

D2 drops the class-instance exemption from the optional-property rejection condition. TestOptionalWideningAllowed fails, as do TestClassWrongOutputIteratorReceiver and TestIteratorMapperIndexHasNumberRepresentation. The historical subsumer TestOptionalWideningRefused passes, showing its single-mutant subsumption was provisional.

D3 changes the fresh object syntax-kind constant to NewExpression. D4 changes it to Identifier. Both break fresh-object admission and are caught by the allowed row plus seven other rows, listed in matrix.json and rows.json. They are distinct classification faults, but both exercise the same fresh-object admission behavior. After three attempts the allowed row is not defended. This is no recommendation to delete it.

All four mutants are standalone diffs, compiled with go vet ./internal/lower, and run against the full package under timeout 120 with the test binary limited to 90s. Each uses its own ADAMIC_BUILD_CACHE_DIR, recorded in matrix.json. No mutant run exceeded budget, aborted, or left an undiscovered top-level row unknown. Each diff also passed git apply --check against the restored starting commit. Production changes were restored; the final full-package control passes.

## Limits and friction

The /tmp filesystem totals only 8.8 GB and had 6.3 GB free, so the requested 15 GB free threshold is impossible there. No earlier named defender scratch directories were available for removal. /workspace had 16 GB free. No disk failure occurred.

The audit predates the current main commit. Current discovery and full-package matrices include newly added tests. Family names have inconsistent prefixes: the Abfe962 member belongs with the two Override members because their checker and assertions agree.

TestOriginalCycleLedger and TestOptionalWideningCensus skip by default. One MixedUnionContractGraph interface/readonly-array subcase also skips. Their ability to catch these mutations is unknown; uniqueness is established for executed rows, not skipped opt-ins. The package's JSON logs preserve these skips.

Go coverage measures Go lowering code, not the semantics of emitted native code. Exclusive blocks alone did not prove worth; D1's semantic diagnostic difference did. Diagnostic production behavior is code under test, and no test or oracle was edited.

The allowed row's name promises admission and its assertions check admission. It has no speed promise or threshold. Its limitation is that successful output content is not checked. The prior audit's empty Lower probe passing this row is historical evidence, not a probe rerun in this defense.

## Cost

Warm setup: 0s. nproc: 5. npm ci: 386ms reported by npm. Clean baseline: 34.292s test binary. Restored control: 36.244s. Mutant full-run wall times: D1 42.320s, D2 44.452s, D3 44.870s, D4 45.557s. Their separate go vet checks total approximately 2.102s. Native rebuild work is included in matrix wall time; it was not separately timed. Coverage times are recorded in their logs. No test was deleted, rewritten, or weakened. No other package's test suite was run.
