# Optional calls rebuilt on main

Task #vx8qdwg, roadmap step 18. Delivery branch: compiler/optional-calls-main.
Base: origin/main 7a10c877667582a15326acb10fa22fa7a0c45fb8.
Source: codex/scout-optional-calls-next fcbbfdb849384a7e356a24c23b48f9df90eee8e0.
The imported change is 4885cec50290686df487b62aac47c85d871ed40c..fcbbfdb8,
limited to that lane's files, with counts regenerated locally. No old base or
area-next merge was imported. Earlier census artifacts remain historical evidence.

Represented free and bound optional calls now save the selected callable and
receiver before arguments. Continuous chains guard the full continuation,
skipping absent indices, spreads and callback factories. Parentheses end the
short circuit. Signature return ABI remains separate from optional result storage;
discarded reference returns are released.

There were no textual conflicts. Three semantic interactions were reconciled:

- General chain indexing intercepted main's typed buffers. The chain path now
  uses the existing typed-buffer index IR and required-value unwrap contract.
  Main's optional_indexing_typed_array.a passes against Node in both backends.
- SSA's gap 2 now lowers. Its probe requires native output 7 undefined with
  sanitizers and leak checks; renameReturns uses its original optional chain.
  The port oracle and its twelve existing output mutants pass.
- Markdown's gap 17 now lowers. The unchanged witness requires output 1 from
  source Node, native and the JavaScript backend, with sanitizers and no leaks.

Main's checked-view metadata remains attached by readObjectField. Refusal,
readiness, cycle and signature proofs remain in their existing paths.
No language policy was relaxed. Callable accessors, Weak callable slots,
checked/uninitialized callable fields, unsupported signatures and representations
remain NotYet. Detached method writes retain their existing refusal.
For pending #tvq1eqm, the proposal is to retain these NotYet boundaries until
callable selection, signature and ownership contracts are ruled. They are not
reported as accepted. Required numeric optional reads also remain pending.

The fresh entry measurement used stage3/apply.sh on the pinned TypeScript source
050880ce59e30b356b686bd3144efe24f875ebc8, then:

```sh
/tmp/optional-calls-main-adamic c /workspace/scratch/optional-calls-main-adapted/src/tsc/tsc.ts
```

Its first stop is src/compiler/builder.ts:1246:69, TS2345: Path | undefined
is not assignable to string. There are 320 checker diagnostics. The checker stop
prevents an entry-level measurement of later optional-call barriers; no whole-tsc
acceptance or fresh census retirement count is claimed.

The 25 supported lane fixtures agree with source Node in both backends, native
ASan/UBSan, release builds and leak checks. The required numeric gap and method
write refusal remain pinned. Every moved existing count row was also held to
Node; the namespace/taste registry groups cover the four moved stage3 rows.
No stage3 status or recorded Node observation changed.

Setup used GOPROXY=https://proxy.golang.org|direct and succeeded. nproc=5.
Reported setup lines: submodules 0.066s, clang 0.173s, Go build 44.745s,
cache warm 44.912s, done 44.940s. The Go test cache was cleared after disk
pressure. Further space failures were traced to the separate /tmp filesystem;
remaining runs used /workspace/scratch/optional-calls-main-tmp for TMPDIR and
GOTMPDIR. Those infrastructure failures are not credited as mutant catches.

## Commands and coverage

All test output was written directly to log files. Commands sourced
/workspace/adamic-tools/env.sh; final runs used workspace-backed scratch storage.

- go build -p 1 ./...
- go vet -p 1 ./internal/...
- a-check on the 27 changed .a files against origin/main, using the pinned
  fast-gate aCheck implementation fbac28c62493f27a788edc02a18bc8edb68de5da.
  The intentional method-write refusal has its required first-line header.
- go test -p 1 ./stage1/... -run 'Gap|Gaps|Probes' -count=1
- go test -p 1 ./stage3/fixtures -count=1, after npm ci --prefix stage3/api.
- ADAMIC_GATE_UNCACHED=1 go test -p 1 ./internal/oracle -run
  'TestStep18|TestNativeAgreesWithNode/docs/step-18/fixtures|TestNativeAgreesWithNode/internal/oracle/testdata/optional_indexing_typed_array.a' -count=1
- Targeted lower/flow/native step-18, receiver-origin, signature and closure
  convention checks; additional Node regressions for taste, host intrinsics,
  class interfaces, strings, Map visits and every moved count row.
- go test -p 1 ./stage1/cohere/static_single_assignment
  -run TestThePortMakesWhatGoCohereMakes -count=1
- go test -p 1 ./internal/oracle -run TestCountsAreRecorded -count=1
  -args -update-counts, on Linux.

The full repository gate and full package oracle were not run. See the historical
report for the earlier source censuses; none was reclassified as acceptance here.

## Mutation evidence

Run python3 docs/step-18/run-mutants.py after sourcing the toolchain. It writes
one Go overlay and log per compiler mutant, restores source-input mutations,
and fails if any mutant survives or only fails to build. The final ledger is
main-mutants.json. All 29 mutants were caught. Earlier receiver-barrier checks
were superseded by represented bound calls; their removed deferral is not counted
as a current check.

An array visitor release mutant escaped LeakSanitizer in one final run, despite
an earlier sanitizer catch. Its exact cause was not isolated. The new counted
balance check removes that detection dependency: the control is 42 allocations,
42 frees, zero regions; either visitor cleanup mutant gives 42/41/0. Dropping
ordinary discarded-reference cleanup gives 13/10/0. All unmutated controls pass.

| Mutant | Catcher |
|---|---|
| callable-reselect | Source Node/native/backend behavior comparison. |
| callable-null-guard | Source Node/native/backend behavior comparison. |
| callable-eager-arguments | Source Node/native/backend behavior comparison. |
| callable-result-zero | Source Node/native/backend behavior comparison. |
| callable-narrowing | Source Node/native/backend behavior comparison. |
| return-descriptor | ASan invalid release through the wrong return descriptor. |
| method-receiver | UBSan null receiver access. |
| method-presence | Source Node/native/backend behavior comparison. |
| method-reselection | Source Node/native/backend behavior comparison. |
| method-eager-spread | Source Node/native/backend behavior comparison. |
| method-retain | ASan heap use after free after callback replacement. |
| chain-short-circuit | UBSan null continuation access; backend behavior comparison. |
| chain-receiver-repeat | Source Node/native/backend behavior comparison. |
| chain-index-eager | Source Node/native/backend behavior comparison. |
| chain-method-presence | Source Node/native/backend behavior comparison. |
| chain-effects-leak | ==391076==ERROR: LeakSanitizer: detected memory leaks |
| chain-ordinary-read | UBSan null ordinary field access. |
| chain-parentheses | Native null access and backend behavior comparison. |
| chain-write-order | Source Node/native/backend behavior comparison. |
| chain-array-map-storage | ==392900==ERROR: LeakSanitizer: detected memory leaks |
| chain-callback-eager | Source Node/native/backend behavior comparison. |
| chain-weak-storage | Exact NotYet boundary assertion: mutated lowering succeeded. |
| chain-accessor | Exact NotYet boundary assertion: mutated lowering succeeded. |
| discarded-reference | step18_optional_calls_test.go:130: heap values leaked: 3 (allocations 13, frees 10, in regions 0) |
| chain-map-visitor | step18_optional_calls_test.go:130: heap values leaked: 1 (allocations 42, frees 41, in regions 0) |
| chain-array-visitor | step18_optional_calls_test.go:130: heap values leaked: 1 (allocations 42, frees 41, in regions 0) |
| source-baseline | Recorded Node stdout comparison. |
| gap-masking | Exact gap reason and location comparison. |
| census-completeness | Completeness assertion; synthetic reader probe, no compiler acceptance claim. |

## Generated counts audit

25 rows added, 42 existing rows moved, none removed. Each row is below.
A/F/R/L/P/G means allocations, frees, retains, releases, peak live and regions.
Existing rows keep their allocation/free/region counts. Saved receivers/callees
and retaining expression results before expression-local cleanup change reference
counts. Earlier local cleanup can lower peak live. NULL and immortal-reference
operations are counted too, so retain/release deltas need not balance as integers.
The three checked-stop fixtures record partial counts at Node-pinned exit 70.

| Fixture | Before A/F/R/L/P/G | After A/F/R/L/P/G | Reason |
|---|---|---|---|
| docs/step-18/fixtures/call-result.a | new | 2/2/8/9/2/0 | Saved call-result receiver and guarded continuation. |
| docs/step-18/fixtures/cross-call.a | new | 4/4/11/16/4/0 | Saved receiver survives a call that changes the binding. |
| docs/step-18/fixtures/element.a | new | 2/2/15/15/1/0 | Guarded array index and absent result. |
| docs/step-18/fixtures/function.a | new | 1/1/7/10/1/0 | Saved optional closure; discarded invocation. |
| docs/step-18/fixtures/method.a | new | 5/5/8/12/4/0 | Saved optional closure field. |
| docs/step-18/fixtures/receiver-call.a | new | 5/5/4/8/3/0 | Saved Map receiver; lookup arguments stay inside its guard. |
| docs/step-18/fixtures/runtime-arguments.a | new | 27/27/29/55/7/0 | Lazy arguments and spreads with default/rest/count dispatch. |
| docs/step-18/fixtures/runtime-bound-method-arguments.a | new | 21/21/15/32/12/0 | Saved bound method with lazy spreads and default/rest/count dispatch. |
| docs/step-18/fixtures/runtime-bound-method-selection.a | new | 12/12/14/26/6/0 | Saved own callback remains selected across an argument write. |
| docs/step-18/fixtures/runtime-bound-methods.a | new | 15/15/12/25/5/0 | Own callback and prototype selection preserve the receiver. |
| docs/step-18/fixtures/runtime-chain-callable.a | new | 14/14/29/39/5/0 | Guarded free and bound calls with object continuations. |
| docs/step-18/fixtures/runtime-chain-index.a | new | 17/17/15/35/3/0 | Saved receivers and lazy array/string indices. |
| docs/step-18/fixtures/runtime-chain-intrinsics.a | new | 42/42/99/144/6/0 | Guarded Map/array/string intrinsics and discarded reference cleanup. |
| docs/step-18/fixtures/runtime-chain-method-continuation.a | new | 11/11/27/39/6/0 | Callable presence remains distinct from its returned undefined value. |
| docs/step-18/fixtures/runtime-chain-parentheses-write.a | new | 2/2/3/2/2/0 | Node-pinned exit 70; counts stop after right-side effects. |
| docs/step-18/fixtures/runtime-chain-parentheses.a | new | 2/2/2/2/2/0 | Node-pinned exit 70; parentheses end the short circuit. |
| docs/step-18/fixtures/runtime-chain-present-undefined.a | new | 2/1/5/3/2/0 | Node-pinned exit 70; ordinary continuation checks a returned undefined. |
| docs/step-18/fixtures/runtime-cross-chain.a | new | 8/8/13/20/5/0 | Saved receiver survives reassignment within a guarded call. |
| docs/step-18/fixtures/runtime-discarded-reference.a | new | 13/13/12/27/5/0 | Owned reference returns through void views are released. |
| docs/step-18/fixtures/runtime-methods.a | new | 17/17/11/26/6/0 | Selected method and receiver survive argument effects. |
| docs/step-18/fixtures/runtime-order.a | new | 34/34/36/74/6/0 | Saved closure snapshots and runtime presence observations. |
| docs/step-18/fixtures/runtime-return-descriptor.a | new | 3/3/14/21/3/0 | Void view retains the source signature return ABI. |
| docs/step-18/fixtures/runtime-values.a | new | 31/31/40/71/4/0 | Optional number/boolean/reference results, including null and falsy values. |
| docs/step-18/fixtures/size.a | new | 3/3/6/8/3/0 | Guarded Map size uses the nullable result representation. |
| docs/step-18/fixtures/two-guards.a | new | 4/4/9/12/4/0 | Separate receiver and callable guards. |
| internal/oracle/testdata/class_as_interface.a | 372/372/303/468/60/0 | 372/372/324/489/58/0 | Reference result retained before expression-local cleanup; NULL/immortal operations are counted. Earlier cleanup lowers peak live. |
| internal/oracle/testdata/class_inheritance_interface.a | 37/37/44/70/13/0 | 37/37/76/102/13/0 | Reference result retained before expression-local cleanup; NULL/immortal operations are counted. |
| internal/oracle/testdata/conditions_ledger78.a | 430/430/80/427/6/0 | 430/430/88/435/6/0 | Saved optional receiver/callee and expression-local cleanup. |
| internal/oracle/testdata/element_access_optional.a | 35/35/50/79/9/0 | 35/35/58/87/9/0 | Saved optional receiver/callee and expression-local cleanup. |
| internal/oracle/testdata/host_optional_intrinsic.a | 12/12/10/23/3/0 | 12/12/16/29/3/0 | Saved optional receiver/callee and expression-local cleanup. |
| internal/oracle/testdata/library_string_conversion.a | 15/15/3/20/3/0 | 15/15/4/21/3/0 | Reference result retained before expression-local cleanup; NULL/immortal operations are counted. |
| internal/oracle/testdata/library_string_raw.a | 45/45/55/89/11/0 | 45/45/57/91/11/0 | Reference result retained before expression-local cleanup; NULL/immortal operations are counted. |
| internal/oracle/testdata/method_coverage_conversions.a | 14/14/11/27/7/0 | 14/14/12/28/7/0 | Reference result retained before expression-local cleanup; NULL/immortal operations are counted. |
| internal/oracle/testdata/namespace_callable_properties.a | 12/12/10/25/4/0 | 12/12/10/27/4/0 | Reference result retained before expression-local cleanup; NULL/immortal operations are counted. |
| internal/oracle/testdata/namespace_method_receiver.a | 7/7/9/16/5/0 | 7/7/10/17/5/0 | Reference result retained before expression-local cleanup; NULL/immortal operations are counted. |
| internal/oracle/testdata/namespaces.a | 36/36/18/58/13/0 | 36/36/22/62/13/0 | Reference result retained before expression-local cleanup; NULL/immortal operations are counted. |
| internal/oracle/testdata/namespaces_debug_modules/main.a | 6/6/3/10/4/0 | 6/6/4/12/4/0 | Reference result retained before expression-local cleanup; NULL/immortal operations are counted. |
| internal/oracle/testdata/namespaces_debug_state.a | 14/14/13/28/4/0 | 14/14/17/34/4/0 | Reference result retained before expression-local cleanup; NULL/immortal operations are counted. |
| internal/oracle/testdata/namespaces_modules/main.a | 9/9/5/17/5/0 | 9/9/6/18/5/0 | Reference result retained before expression-local cleanup; NULL/immortal operations are counted. |
| internal/oracle/testdata/namespaces_nested.a | 17/17/18/43/13/0 | 17/17/21/46/13/0 | Reference result retained before expression-local cleanup; NULL/immortal operations are counted. |
| internal/oracle/testdata/namespaces_observed_narrowing.a | 6/6/9/18/3/0 | 6/6/9/20/3/0 | Reference result retained before expression-local cleanup; NULL/immortal operations are counted. |
| internal/oracle/testdata/namespaces_pair/main.a | 8/8/8/20/7/0 | 8/8/9/21/7/0 | Reference result retained before expression-local cleanup; NULL/immortal operations are counted. |
| internal/oracle/testdata/namespaces_parser_enums.a | 12/12/8/16/7/0 | 12/12/10/18/7/0 | Reference result retained before expression-local cleanup; NULL/immortal operations are counted. |
| internal/oracle/testdata/namespaces_parser_factory.a | 8/8/7/17/6/0 | 8/8/7/18/6/0 | Reference result retained before expression-local cleanup; NULL/immortal operations are counted. |
| internal/oracle/testdata/namespaces_parser_state.a | 9/9/1/11/4/0 | 9/9/3/15/4/0 | Reference result retained before expression-local cleanup; NULL/immortal operations are counted. |
| internal/oracle/testdata/optional_class_method.a | 40/40/44/71/9/0 | 40/40/71/98/9/0 | Saved optional receiver/callee and expression-local cleanup. |
| internal/oracle/testdata/optional_indexing_array.a | 43/43/61/104/8/0 | 43/43/70/113/6/0 | Saved optional receiver/callee and expression-local cleanup. Earlier cleanup lowers peak live. |
| internal/oracle/testdata/optional_indexing_chain.a | 52/52/70/107/8/0 | 52/52/112/151/8/0 | Saved optional receiver/callee and expression-local cleanup. |
| internal/oracle/testdata/optional_indexing_map.a | 35/35/35/65/8/0 | 35/35/39/69/6/0 | Saved optional receiver/callee and expression-local cleanup. Earlier cleanup lowers peak live. |
| internal/oracle/testdata/optional_indexing_string.a | 14/14/48/65/4/0 | 14/14/62/79/4/0 | Saved optional receiver/callee and expression-local cleanup. |
| internal/oracle/testdata/optional_indexing_string_gap.a | 0/0/9/10/0/0 | 0/0/12/13/0/0 | Saved optional receiver/callee and expression-local cleanup. |
| internal/oracle/testdata/params_namespaces_callbacks.a | 17/17/24/39/12/0 | 17/17/26/41/12/0 | Reference result retained before expression-local cleanup; NULL/immortal operations are counted. |
| internal/oracle/testdata/params_namespaces_dotted.a | 21/21/14/41/12/0 | 21/21/19/46/12/0 | Reference result retained before expression-local cleanup; NULL/immortal operations are counted. |
| internal/oracle/testdata/params_namespaces_order.a | 57/57/14/71/9/0 | 57/57/16/73/9/0 | Reference result retained before expression-local cleanup; NULL/immortal operations are counted. |
| internal/oracle/testdata/params_namespaces_shared/main.a | 16/16/27/46/11/0 | 16/16/28/47/11/0 | Reference result retained before expression-local cleanup; NULL/immortal operations are counted. |
| internal/oracle/testdata/params_namespaces_values.a | 43/43/85/103/23/0 | 43/43/87/105/23/0 | Reference result retained before expression-local cleanup; NULL/immortal operations are counted. |
| internal/oracle/testdata/predicate_refusals/representation_controls.a | 14/14/16/32/7/0 | 14/14/17/33/7/0 | Reference result retained before expression-local cleanup; NULL/immortal operations are counted. |
| internal/oracle/testdata/proven_satisfies.a | 28/28/30/57/13/0 | 28/28/33/60/13/0 | Saved optional receiver/callee and expression-local cleanup. |
| internal/oracle/testdata/structural_statics_optional.a | 7/7/16/24/4/0 | 7/7/23/31/4/0 | Saved optional receiver/callee and expression-local cleanup. |
| internal/oracle/testdata/taste_comma.a | 51/51/5/54/4/0 | 51/51/5/56/4/0 | Reference result retained before expression-local cleanup; NULL/immortal operations are counted. |
| internal/oracle/testdata/taste_optional_join.a | 11/11/8/22/4/0 | 11/11/11/25/3/0 | Saved optional receiver/callee and expression-local cleanup. Earlier cleanup lowers peak live. |
| internal/oracle/testdata/taste_stage3_representations.a | 24/24/11/34/10/0 | 24/24/11/35/10/0 | Reference result retained before expression-local cleanup; NULL/immortal operations are counted. |
| internal/oracle/testdata/taste_void.a | 39/39/8/44/4/0 | 39/39/8/46/4/0 | Reference result retained before expression-local cleanup; NULL/immortal operations are counted. |
| stage3/fixtures/namespaces/08_parser_jsdoc_nested.a | 10/10/4/13/6/0 | 10/10/4/14/6/0 | Reference result retained before expression-local cleanup; NULL/immortal operations are counted. |
| stage3/fixtures/taste/09_literal_cache.a | 6/6/18/26/4/0 | 6/6/20/28/4/0 | Reference result retained before expression-local cleanup; NULL/immortal operations are counted. |
| stage3/fixtures/taste/13_void_callback.a | 4/4/8/14/4/0 | 4/4/9/16/4/0 | Reference result retained before expression-local cleanup; NULL/immortal operations are counted. |
| stage3/fixtures/taste/14_relative_complement.a | 10/10/22/34/7/0 | 10/10/24/36/7/0 | Reference result retained before expression-local cleanup; NULL/immortal operations are counted. |
