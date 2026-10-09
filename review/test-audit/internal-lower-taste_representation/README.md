u044: internal/lower taste and representation audit

Starting commit: 2f7d81f91623c81b86e2ce5a439fdca69ef2d2b8. Branch: test-audit/internal-lower-taste_representation. All 15 requested names were present in go test -list. None moved or vanished from the six listed test files. No top-level families, helper entries, setup checks, or witnesses were identified. Callback error/unknown descriptors in the mixed-union tests are production API inputs; those tests assert transaction behavior rather than witness a harness agreement check.

CODE UNDER TEST: Adamic Lower, viewContract, internMixedUnionViewContract, UntaggedViewMembers, and the named functions listed by clean slice coverage in reached-functions.txt. The production source, tests, and CLAUDE.md were read before mutations. ORACLE: handwritten IR and refusal expectations, with the live Microsoft TypeScript-Go checker supplying the graph row's expected member names/count. No Node/native comparison runs in these 15 rows. No independently checked external-authority value.

Results: {'subsumed': 6, 'sacred': 6, 'overlapping': 3}. All 15 own-entry probes killed their rows; no vacuous row. 19 fixed-menu mutants plus supplemental M20. All subsumption/overlap findings are hints based on the exact recorded kill sets, not deletion recommendations. Sacred means bounded worthiness here: skipped checks and unfinished outside-slice rows remain unknown, and repository-wide uniqueness is for central replay. Known outside-slice catches are retained, notably EnumLimitsStayLoud and DefaultTaggedInterfaceAdmission.

Evidence: results.json contains every requested field including matrix_rows. matrix.json and matrix.tsv contain observed states. mutant-plan.json fixes the code-derived changes and origin lines. Every Mxx.diff and Pxx.diff applies to the starting origin/main and passes go vet ./internal/lower/ (24 checks in standalone-vet.json). Source switches are retained only as *.switched.txt and audit_switch.go.txt. Production source was restored byte-for-byte and the clean slice passed again in restored-slice.log. All test output went directly to files. *.log.gz are lossless copies of the raw logs.

Commands

Warm tools: source /workspace/adamic-tools/env.sh; nproc (5). npm ci in stage3/api before baseline. Discovery: timeout 90 go test -list . ./internal/lower/. Baseline: ADAMIC_BUILD_CACHE_DIR=/tmp/u044/cache/baseline timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > baseline.log 2>&1. Binary line: ok  	github.com/system-inc/adamic/internal/lower	39.194s.

Each timing: timeout 120 go test -count=1 -timeout 90s ./internal/lower/ -run '^TEST$' > time-TEST-N.log 2>&1, three runs. Medians are the binary's ok line, not go command wall time. Two outside-slice subsumers were also timed three times without matrix load; earlier loaded timings are preserved separately.

Each whole-package matrix: ADAMIC_MUTANT=Mxx ADAMIC_BUILD_CACHE_DIR=/tmp/u044/cache/Mxx timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > Mxx.log 2>&1. Two independent commands ran concurrently. Cooked: M01, M09, M10, M11, M12, M15, M17. Their timeout panics are not kills. Each of the 15 scoped rows was rerun alone for those mutants with the same command and -run '^TEST$'. Those isolated observations replace the incomplete scoped results; other incomplete rows remain unknown. Reached-function coverage establishes the bounded slice, not the complete set of outside callers.

Own-entry probes: P01 returns nil,nil from Lower for seven lowering rows; P02 returns 0,nil from internMixedUnionViewContract for five mixed-union rows; P03 returns 0,nil from viewContract for phantom void; P04 returns nil,nil from UntaggedViewMembers for its two rows. Each was run per row alone. Standalone probe diffs replace the whole entry body and clean unused imports, equivalent to the switched entry return used in the observed runs.

Survivors

M01, bounded survivor: Lower('const a = new Uint8Array([])') changes TypedArrayNew FromArray=true to false, with ArrayLiteral source, while all 15 scoped rows pass. The whole-package attempt cooked, so unfinished outside-slice results are unknown.
M02, enabled whole-package survivor: Lower('const b = new Uint8Array([1,2])') changes TypedArrayNew FromArray=true to false, with ArrayLiteral source. Whole package passed in 88.080s (JSON event 88.082s). Neither is an equivalent candidate. witness/main.go and witness-clean/M01/M02.log contain the direct IR witnesses. Commands: ADAMIC_MUTANT=ID ADAMIC_BUILD_CACHE_DIR=/tmp/u044/cache/witness-ID go run ./review/test-audit/internal-lower-taste_representation/witness > witness-ID.log 2>&1. No native runtime behavior is asserted by these witnesses.

Brief issues, limits, and time costs

1. The brief pins file locations at 8de93800f4 but directs starting at current origin/main. Current fetched main is 2f7d81f91623c81b86e2ce5a439fdca69ef2d2b8; all file:line references use that actual starting commit. No requested name disappeared.
2. Three mutants per row would require 45 changes, conflicting with the 20-change cap. The cap governed: 19 verdict-bearing mutations plus one supplemental append-storage change. M20 changes the append destination rather than a clearly named constant/compiler option, so no verdict rests on it.
3. Whole-package replay dominates this small slice's cost. The selected rows total little binary time, but unrelated tests compile native products. Separate ADAMIC_BUILD_CACHE_DIR values were used for every compiler mutation. Two concurrent package commands, cold native products, and a failed isolated dependency build pushed seven binaries to 90 seconds despite the clean baseline fitting. All cooked matrices were narrowed and outside-slice unknowns retained. The audit exceeded the approximate 20-minute budget.
4. The fixed-budget matrix and automatic parallel recovery still required repeated go command overhead. A single switched source compiled once for the matrix; the scoped rows themselves build no native products. Whole-package native rebuild durations were not separately instrumented, so per-product rebuild time is not available. Per-mutant whole-binary and command wall times are reported below.
5. An isolated worktree lacked cohere and TypeScript local modules. Linking only TypeScript was insufficient because go.mod also replaces cohere. Even corrected links changed module paths and triggered a cold dependency build; that attempt was stopped. All 24 standalone checks were ultimately rerun in the warm original checkout and passed. The failed attempts are preserved, not counted as validation.
6. The initial coverage filter accidentally omitted percentages ending in 0.0, including 100.0; functions-coverage.txt already contained the complete pre-mutation inventory. reached-functions.txt was corrected to include every positive percentage (380 named functions). No mutant target was selected from test expectations.
7. Initial switch construction had a return-arity error in M02. go vet caught it before any mutant matrix ran; the final switched source and standalone diffs compile. Dropping map deletion requires dropping the whole loop (M14), avoiding unused loop variables.
8. The witness rule can be confused with callback failure injection. These mixed-union rows directly exercise the production intern transaction and inspect its state; their injected callback failures are inputs, not mutations of an oracle or harness. Production rollback/commit mutants demonstrably fail them.
9. The taste enum-prototype failure message says want NotYet although its code correctly checks Refused. The graph row's live-checker name/count comparison does not establish correct scalar contract kinds: it passes M11. TypedArraysLower omits FromArray and exact-kind/runtime assertions: M02 survives. The representation-change row checks only the NotYet class and can accept a different refusal reason.
10. Still skipped: TestOriginalCycleLedger, TestOptionalWideningCensus, TestMixedUnionContractGraph/interface_Node_{readonly_ready:boolean}_type_Target=Node|readonly_Node[];. OriginalCycleLedger needs a pristine TypeScript corpus with generated diagnostics; OptionalWideningCensus needs a project config/output. Both are outside this unit and were not enabled. The scoped graph array subcase explicitly awaits unimplemented views-v3 support, rather than an installable tool. No optional corpus/SDK row within the 15 remained disabled. No other package's tests were run.
11. Go test JSON final Elapsed and the binary ok line differ by milliseconds. Row costs use the ok line as requested. The 90-second budget is the binary timeout; outer command wall time can exceed it during compilation/cleanup, as the brief's explicit timeout command permits.

Timing

Setup skipped (0s), warm env verified. npm ci 0.737s; discovery 2.402s; clean baseline command 41.452s, binary 39.194s. Switch go vet/build check 0.651s. Final standalone vet checks 17.154s total. Whole-package binary time 1728.778s summed; command time 1886.537s summed across two workers. Recovery/probe command time 462.921s summed across four workers. Elapsed from npm start through report generation 26.3min. Initial setup verification/fetch and final evidence packaging add small overhead not separately timed. No repo-wide replay, external-authority validation, native semantic witness, or per-product rebuild timing was performed.

| Mutant | Whole binary seconds | Command wall seconds | Scope |
|---|---:|---:|---|
| M02 | 88.082 | 103.220 | enabled whole package |
| M01 | 90.087 | 105.547 | cooked; 15 isolated rows |
| M03 | 76.517 | 79.600 | enabled whole package |
| M04 | 79.448 | 81.896 | enabled whole package |
| M05 | 77.016 | 80.245 | enabled whole package |
| M06 | 79.154 | 81.759 | enabled whole package |
| M07 | 80.565 | 84.092 | enabled whole package |
| M08 | 88.932 | 91.525 | enabled whole package |
| M09 | 90.617 | 102.598 | cooked; 15 isolated rows |
| M10 | 90.463 | 101.452 | cooked; 15 isolated rows |
| M11 | 90.149 | 118.280 | cooked; 15 isolated rows |
| M12 | 90.165 | 107.145 | cooked; 15 isolated rows |
| M14 | 87.423 | 91.072 | enabled whole package |
| M13 | 89.446 | 94.567 | enabled whole package |
| M16 | 87.375 | 91.541 | enabled whole package |
| M15 | 90.096 | 96.438 | cooked; 15 isolated rows |
| M17 | 90.078 | 95.322 | cooked; 15 isolated rows |
| M18 | 89.795 | 92.772 | enabled whole package |
| M19 | 86.597 | 93.724 | enabled whole package |
| M20 | 86.773 | 93.743 | enabled whole package |

| ID | Origin file:line | Change | Observed failing rows |
|---|---|---|---|
| M01 | internal/lower/typed_arrays.go:94 | FromArray: true -> FromArray: false |  |
| M02 | internal/lower/typed_arrays.go:113 | return ir.TypedArrayNew{Of: kind, Source: source, FromArray: true}, true, nil -> return ir.TypedArrayNew{Of: kind, Source: source, FromArray: false}, true, nil |  |
| M03 | internal/lower/typed_arrays.go:256 | Value: value, Element: ir.Number, Site: -> Value: value, Element: ir.String, Site: | TestTypedArraysLower |
| M04 | internal/lower/typed_arrays.go:208 | if value.Type() != kind { -> if value.Type() == kind { | TestTypedArrayGaps, TestTypedArraysLower |
| M05 | internal/lower/expression.go:286 | return fromKind != 0 && fromKind == toKind -> return true | TestTypedArrayViewsCannotChangeRepresentation |
| M06 | internal/lower/unknown.go:31 | strings.ContainsRune(key.Text(), 0) -> strings.ContainsRune(key.Text(), 1) | TestUnknownReflectionRefusals |
| M07 | internal/lower/unknown.go:51 | key.Text() == "stack" -> key.Text() == "stack_disabled" | TestUnknownReflectionRefusals |
| M08 | internal/lower/enums.go:183 | member.Name().Text() == "__proto__" -> member.Name().Text() == "__proto_disabled__" | TestEnumLimitsStayLoud, TestTasteRepresentationLimitsStayExplicit |
| M09 | internal/lower/interface_cast.go:77 | if of == ir.Object { -> if of != ir.Object { | TestDefaultTaggedInterfaceAdmission, TestViewObjectWritesNeedSourceCertificate |
| M10 | internal/lower/view_contracts.go:27 | Kind: ir.ViewUndefined, Name: "undefined", Undefined: true -> Kind: ir.ViewUndefined, Name: "undefined", Undefined: false | TestMixedUnionContractPhantomVoidIsUndefined |
| M11 | internal/lower/view_contracts.go:69 | contract.Kind = ir.ViewScalar -> contract.Kind = ir.ViewObject | TestMixedUnionContractPhantomBrandUsesPrimitiveBase, TestMixedUnionContractRecursiveMember, TestViewObjectContractsAreAvailableToEraser |
| M12 | internal/lower/view_contracts.go:90 | l.result.ViewContracts[int(id)-1] = contract -> (drop) | TestDefaultTaggedInterfaceAdmission, TestMixedUnionContractGraph, TestMixedUnionContractRecursiveMember, TestPredicateBodyProof, TestViewObjectContractsAreAvailableToEraser, TestViewObjectWritesNeedSourceCertificate |
| M13 | internal/lower/view_unions_mixed.go:42 | l.result.ViewContracts = l.result.ViewContracts[:start] -> (drop) | TestMixedUnionContractFailureDoesNotCertifyRetry, TestMixedUnionContractUnknownMemberFails |
| M14 | internal/lower/view_unions_mixed.go:43 | for key, value := range l.result.ViewContractTypes { 			if int(value) > start { 				delete(l.result.ViewContractTypes, key) 			} 		} -> (drop) | TestMixedUnionContractFailureDoesNotCertifyRetry, TestMixedUnionContractUnknownMemberFails |
| M15 | internal/lower/view_unions_mixed.go:56 | l.result.ViewContracts[int(child)-1].Kind == ir.ViewUnknown -> l.result.ViewContracts[int(child)-1].Kind == ir.ViewCallable | TestMixedUnionContractUnknownMemberFails |
| M16 | internal/lower/view_unions_mixed.go:61 | l.result.ViewContracts[int(id)-1] = contract -> (drop) | TestMixedUnionContractFailureDoesNotCertifyRetry, TestMixedUnionContractGraph, TestMixedUnionContractPhantomBrandUsesPrimitiveBase, TestMixedUnionContractRecursiveMember |
| M17 | internal/lower/view_unions_primitive_brands.go:69 | return flags&checker.TypeFlagsVoid != 0 \|\| optional && flags&checker.TypeFlagsUndefined != 0 -> return flags&checker.TypeFlagsVoid == 0 \|\| optional && flags&checker.TypeFlagsUndefined != 0 | TestMixedUnionContractPhantomBrandUsesPrimitiveBase, TestMixedUnionContractPhantomVoidIsUndefined |
| M18 | internal/lower/view_unions_untagged.go:51 | if field.Optional { -> if !field.Optional { | TestUntaggedViewMemberTags |
| M19 | internal/lower/view_unions_untagged.go:45 | if contract.Kind != ir.ViewObject \|\| contract.Of != ir.Object { -> if contract.Kind == ir.ViewObject \|\| contract.Of != ir.Object { | TestUntaggedViewMemberTags, TestUntaggedViewStructuralFallback |
| M20 supplemental | internal/lower/view_unions_untagged.go:66 | append([]ir.ViewLiteral(nil), tag.Allowed...) -> append(tag.Allowed[:0], tag.Allowed...) | TestUntaggedViewMemberTags |
