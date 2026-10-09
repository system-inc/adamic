# Optional object-slot views

Roadmap step 09, #tzd3gjg. Candidate branch compiler/views-optional-object-slots, based on V3 8880e6bca. This candidate rides after V3, separately from compiler/step09-predicates-ahead. Evidence stays in this review directory.

Optional plain object | undefined own fields use the existing readiness-aware union snapshot and child contract. Reading never changes own-property presence. The child fields remain checked after aliasing and recursive reads. Accessors, optional receiver chaining, nominal object families and incompatible source optional declarations retain their existing boundaries. No producer metadata or construction evidence is added.

Six reduced .a fixtures cover present objects, present undefined, missing, recursive saved aliases, a wrong physical child field type and a wrong literal child field. Source Node decides the successful results; wrong child fields deliberately stop with the named exit-70 check instead of Node's unchecked output. Native runs ASan/UBSan and leak checks; both backends agree on the named stopping result. Unproven optional source relations remain refused with the path and fix.

Mutants change missing storage into present undefined and present undefined into missing; both finish cleanly but disagree with Node's hasOwnProperty output in both backends. A separate mutant drops the child literal check; both backends then reproduce Node's unchecked wrong value, contradicting the required exit-70 result. The representation misfit also remains checked.

## Checks

`ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestOptionalObject' -count=1 -parallel=1 -v`: PASS, 3.678s. `go test ./internal/lower -run '^TestOptionalObject' -count=1 -parallel=1 -v`: PASS, 0.100s. Each added top-level test includes its fixture setup. One earlier cold-cache present-object run took 12.47s; the final uncached measurements follow.

| Test | Seconds |
| --- | ---: |
| TestOptionalObjectAliasContract | 0.03 |
| TestOptionalObjectGetterRemainsRefused | 0.03 |
| TestOptionalObjectUnprovenSourceRefused | 0.03 |
| TestOptionalObjectPresent | 0.46 |
| TestOptionalObjectLiteralMisfit | 0.54 |
| TestOptionalObjectChildViewMutant | 0.39 |
| TestOptionalObjectUndefinedPresenceMutant | 0.32 |
| TestOptionalObjectMissingPresenceMutant | 0.33 |
| TestOptionalObjectRecursive | 0.42 |
| TestOptionalObjectMisfit | 0.40 |
| TestOptionalObjectMissing | 0.42 |
| TestOptionalObjectUndefined | 0.40 |

Counts refreshed with `go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -args -update-counts`: PASS, 65.502s. This pre-existing writer was not added or touched. Six allocation rows were added, with no existing row changed. The four finishing fixtures balance allocations/frees. The two checked exits retain their terminal allocations.

Setup uses GOPROXY=https://proxy.golang.org|direct and /workspace/adamic-tools/env.sh. Timing: Node 0.034s, Go 0.035s, markdown 0.094s, submodules 0.100s, clang 0.220s, build 43.128s, deferred test binaries 43.258s, warm 43.260s, done 43.289s. nproc=5, cgroup cpu.max=400000 100000 (four CPUs). Go 1.27.1, Node 24.19.0, clang 20.1.8.

The local predicates merge measurement and committed integration lane checks will be recorded before the single push. The predicates merge is scratch-only and will not be pushed. No sibling test file or protected compiler orchestration file is edited by this unit.

## Final V3 dependency and remeasurement

V3 was merged at b8dfa388f (candidate merge c88cc27ae); the conversion is 90c5001d5, and the missing-parent refusal is bdc8700d5. The predicates scratch tree starts at acc2bfbcf, merges the same V3 tip and this candidate, and is never pushed. Local compatibility resolutions retain scalar optional views, presence/order storage, tuple identity, and delegated proof bodies. The measurement overlay disables executable output and preserves all 320 checker diagnostics. The same 79 source hashes and exact 138 predicate coordinates match 8a7ab17e.

| Group | Bodies | Logical proofs | Audited full admissions before → after | Pending views before → after | Final .a refusals |
| --- | ---: | ---: | ---: | ---: | ---: |
| Direct kind | 133 | 133 | 0 → 0 | 133 → 133 | 133 |
| Delegation | 5 | 5 | 0 → 0 | 5 → 5 | 5 |

Zero additional full corpus bodies admit. The old unmodified admission probe reports 138 Proven on the V3 merge even before this conversion, but it constructs no original check: 122 target roots are unsupported representation-conversion placeholders, and 16 union targets have unavailable member roots. That is not a valid admission. The audit records the descriptors and refuses to label them pass. Final production now refuses those optional object reads with a real path and fix rather than allowing the placeholder to erase checks.

Representative final diagnostic:

```text
utilitiesPublic.ts:768:16: Adamic 0.1 refuses an optional object field read node.original without a complete BigIntLiteral view; prove or implement the representation conversion contract before reading this field through its checked child view
```

The first recorded descriptor construction failure per body is:

| Remaining contract | Bodies |
| --- | ---: |
| Unavailable union member JSDocArray | 119 |
| Unavailable callable union member (node: EmitHelperUniqueNameCallback) => string | 16 |
| Unavailable union member NodeArray<Modifier> | 1 |
| Unavailable union member NodeArray<ModifierLike> | 1 |
| Checked union representation for HasJSDoc | 1 |

These are remaining family/representation builders on compiler/views-v3 b8dfa388f, not the plain optional object slot conversion. The callback failures also recur underneath ScopedEmitHelper and EmitNode. No implementation SHA resolving these exact descriptors was found or merged, and none is credited as passing. The view builder must retain or refuse unavailable parent graphs; the new guard covers this unit's optional object reads. Other placeholder families are outside this unit. Results are admission/proof probes, not a successful native compiler build or a remeasurement of the other 442 body predicates.

To reproduce, use the unpushed local predicates/V3/candidate merge, generate the existing predicates measure-overlay.py overlay, replace its hook with measurement-hook.go.txt, and append HatchViewFailures recording to view_lazy.go's representation-conversion fallback in the scratch overlay. Build the existing measure-probe with -tags hatch_predicate_measurement. Set HATCH_BODY_ONLY=1 and HATCH_SELECTION to results.json's selection map, then run against the hash-pinned adapted tree. The before overlay substitutes the pre-candidate interface_cast.go; the after overlay uses the candidate guard. The exact descriptors, diagnostics and status corrections are in before-audited.json.gz and after-audited.json.gz.

## Final checks and test grain

Final uncached oracle selector: PASS, 3.355s. Final focused lower selector: PASS, 0.121s. Final required counts update: PASS, 52.788s; the six new rows remain the only changes. Focused vet passes for internal/ir, internal/lower, internal/native, internal/javascript and internal/oracle. The extra parent-guard omission mutant is independently caught by TestOptionalObjectIncompleteParentRefused; it is a compile-time refusal mutant, separate from the three semantic mutants held to Node in both backends.

| Added top-level test | Final seconds |
| --- | ---: |
| TestOptionalObjectAliasContract | 0.03 |
| TestOptionalObjectGetterRemainsRefused | 0.03 |
| TestOptionalObjectIncompleteParentRefused | 0.03 |
| TestOptionalObjectUnprovenSourceRefused | 0.02 |
| TestOptionalObjectPresent | 0.45 |
| TestOptionalObjectLiteralMisfit | 0.39 |
| TestOptionalObjectChildViewMutant | 0.34 |
| TestOptionalObjectUndefinedPresenceMutant | 0.32 |
| TestOptionalObjectMissingPresenceMutant | 0.33 |
| TestOptionalObjectRecursive | 0.40 |
| TestOptionalObjectMisfit | 0.37 |
| TestOptionalObjectMissing | 0.38 |
| TestOptionalObjectUndefined | 0.38 |

All 13 added test leaves run under 60 seconds including their fixture setup on the four-CPU cgroup. No pending test is added. The final oracle run is uncached. The cold post-V3 present-object leaf also passes in 23.94s. No sibling's three admission test files are edited by this unit.

The required committed repository-root lane command passed:

```text
lane checks 10.4 s: gofmt and tools on 62 Go files, t.Parallel on 6 test packages; no t.Parallel analyzer on this tree; vet 6 packages
```

This V3 base has no parallel analyzer. Integration's unchanged origin/main analyzer was also run in a scratch tree containing only this unit's added test files; the extra parent-negative leaf is included in its final rerun. The lane checks are rerun on the final committed candidate before its one push. No main, area or predicates merge is pushed.

## Delivery dependency check

V3's json-memory fix arrived before delivery. The candidate merges compiler/views-v3 14690251f29b6f155c5d795be13eb420de00471a in 1216e7e4b. Only the native array emission path changes production code from b8dfa388f; no sibling test file is edited here. The local predicates merge includes this tip too. Its final 138-body JSON is identical to the earlier final frontend measurement; results.json records the current local merge SHA and exact five delegation positions.

Final uncached optional-object oracle selector after that merge: PASS, 3.295s. Final vet of all five touched packages: PASS. The independent parallel analyzer checks 14 top-level tests including itself: zero failures, zero serial baseline entries. The corpus audit passes and catches four independent corruptions: missing body, invented pass, fabricated original check and wrong source hash.

| Oracle test after final V3 merge | Seconds |
| --- | ---: |
| TestOptionalObjectPresent | 0.44 |
| TestOptionalObjectLiteralMisfit | 0.39 |
| TestOptionalObjectChildViewMutant | 0.30 |
| TestOptionalObjectUndefinedPresenceMutant | 0.29 |
| TestOptionalObjectMissingPresenceMutant | 0.30 |
| TestOptionalObjectRecursive | 0.40 |
| TestOptionalObjectMisfit | 0.37 |
| TestOptionalObjectMissing | 0.41 |
| TestOptionalObjectUndefined | 0.39 |

Dependency: compiler/views-v3 14690251. This remains its own candidate after V3. The unpushed predicates compatibility merge is a measurement tree with executable output disabled, not a proposed integration merge.
