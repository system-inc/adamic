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
