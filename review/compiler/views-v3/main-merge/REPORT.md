Merged main into compiler/views-v3 for V3, #6wfvfhm, retaining admissions and concurrent tests.
Merge b6778c90343d15e9294e90f7fe1eb66de7238dda joins V3 14690251f29b6f155c5d795be13eb420de00471a and main f04f6a7073c1358c9a6c0a8fd78a4822786ba865; parallelism repair 420f6420; delivery tip is in the handoff.
All 24 requested top-level tests pass without skips, including uncached Node/native/JavaScript checks, sanitizers, and the JSON zero-store bound.
Existing V2 mutants pass their expected catches; the restored-store mutant fails with 2,885 inline stores, and lane-required array mutants pass.
No production code, admission expectation, fixture or counts changed beyond the main merge; no whole package or full gate was run.

Each conflict resolution

1. internal/lower/view_unions_mixed_test.go: retain the Node-or-readonly-Node-array
   member contract and its graph checks. Remove main's now-obsolete V3 skip,
   retaining main's first Parallel call in the subtest and its parallel calls on
   the other five test roots. The array member must still have an element contract.
2. internal/oracle/checked_views_v2_migration_test.go: retain the V3 name
   TestCheckedViewV2ArrayArmAdmission and compare its true output to source Node
   in release native, sanitized native and emitted JavaScript. Retain all main
   parallel calls on the other roots and nested cases; keep all mutation assertions.
3. internal/oracle/checked_views_v2_object_primitive_test.go: keep comment-array
   good, wrong-flags and boolean cases active, with V3's exact diagnostic pins.
   Keep the existing root/subtest parallel calls. This resolution is identical
   to the V3 file because those parallel calls were already present there.
4. internal/oracle/checked_views_v2_source_test.go: retain the V3 name
   TestCheckedViewUntaggedArraySource and the good, wrong, nested, empty and mixed
   cases, with V3's success and rejection expectations. Keep main's parallel
   calls throughout the source/flow/recursive test roots and subtests. Restore
   neither the old pending name nor its Skip.

There were no other textual merge conflicts. Main's concurrency-safe oracle
cache implementation is retained unchanged: identity initialization uses sync.Once,
result fills lock per key, and publication uses a temporary file and rename.
The resolved cache file has no diff against main. Mutants continue using private
IR and runtime snapshots, so parallel execution does not modify other cases.

The first lane run found twelve inherited V3 test roots without Parallel first.
The repair adds only those calls across view_array_writes_test.go,
array_holes_test.go, array_runtime_mutants_test.go,
checked_views_array_witnesses_test.go and checked_views_arrays_mutants_test.go.
Their assertions and private-snapshot helpers are unchanged. Those twelve named
roots were then verified separately; this is the only expansion beyond the
requested test selection, required to validate the lane repair. No checks were
removed and no test was marked pending to satisfy the lane.

Requested verification

Every Test function from the four conflicted files is selected explicitly, plus
TestCheckedViewArrays and TestPortElementMetadataEmission. All commands use
GOMAXPROCS=4, ADAMIC_GATE_UNCACHED=1, -count=1, -parallel=4, -timeout=10m and
-json. Complete argument arrays are in evidence/command-results.json, and the
exact selectors are in evidence/test-selectors.json. No broad package run was used.

```
source /workspace/adamic-tools/env.sh
export GOMAXPROCS=4
ADAMIC_GATE_UNCACHED=1 go test ./internal/lower -run '^(TestMixedUnionContractGraph|TestMixedUnionContractFailureDoesNotCertifyRetry|TestMixedUnionContractUnknownMemberFails|TestMixedUnionContractRecursiveMember|TestMixedUnionContractPhantomBrandUsesPrimitiveBase|TestMixedUnionContractPhantomVoidIsUndefined)$' -count=1 -parallel=4 -timeout=10m -json > /tmp/views-v3-internal-lower.jsonl 2> /tmp/views-v3-internal-lower.stderr
ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/json -run '^(TestPortElementMetadataEmission)$' -count=1 -parallel=4 -timeout=10m -json > /tmp/views-v3-stage1-cohere-json.jsonl 2> /tmp/views-v3-stage1-cohere-json.stderr
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^(TestCheckedViewV2ReadsAfterWrites|TestCheckedViewV2ArrayArmAdmission|TestCheckedViewV2WrongFamilyMutant|TestCheckedViewV2RepresentationMutants|TestCheckedViewV2MembershipMutant|TestCheckedViewV2TupleIdentityMutant|TestCheckedViewV2CallableProducerMutant|TestCheckedViewV2MovedStage3Results|TestCheckedViewObjectPrimitiveSource|TestCheckedViewUntaggedSourceDispatch|TestCheckedViewUntaggedCallableUnion|TestCheckedViewUntaggedOptionalCallableControl|TestCheckedViewUntaggedSourceFlows|TestCheckedViewUntaggedOwnClassData|TestCheckedViewUntaggedRecursive|TestCheckedViewUntaggedArraySource|TestCheckedViewArrays)$' -count=1 -parallel=4 -timeout=10m -json > /tmp/views-v3-internal-oracle.jsonl 2> /tmp/views-v3-internal-oracle.stderr
```

The lower selection passes in 0.171s (9.066s including build); the JSON bound
passes in 1.288s (9.110s including build); the oracle selection passes in 32.152s
(38.587s including build). All selected cases run, with no failure or skip.
The longest selected test takes 13.87s. Every selected test's elapsed seconds,
including all roots and subcases, are recorded in evidence/test-results.json.
No test functions were added or split in this merge. Both the test elapsed times
and each requested command's complete build/setup time are below 60 seconds.

The JSON bound observes 0 direct stores and 2,885 runtime helper calls in
2,027,487 emitted C bytes. The production-code overlay restores the old inline
store and fails TestPortElementMetadataEmission at its bound in 1.23s:
`element metadata direct stores 2885 exceed bound 0`. The mutant builds and
lowers successfully before the assertion fails. Its source is
restored-metadata-stores.go.txt, not a compilable Go file under review/.
The JSON bound is an emission regression test, not a whole JSON port execution.

The existing V2 fault injections are retained and run: wrong adapter family,
membership check omission, tuple flag removal, producer certificate removal,
and old slot layouts for null, undefined, uint8, int32 and float64. The first
four are caught by the checked-view expectations in native, sanitized native
and JavaScript. Null/undefined old layouts fail under UBSan; old typed-array
layouts fail their expected runtime read result. The selected normal fixtures
continue comparing source Node output with both generated backends, while
ill-typed views require the pinned runtime rejection in both backends.

The lane-repair verification additionally catches source-slot certificate
omission, missing array slots, missing RangeError, element-kind checks, missing
receiver checks, physical-write checks for scalar/reference storage, viewed-hole
absence and eight array read-check removals (array-second, array-boolean,
array-string-literal, array-undefined, array-iteration-bad, array-map-bad,
non-array, array-missing). Controls and both-backend mutant observations remain
unchanged. The witness-extension and source Node controls also pass.

Lane-repair root seconds

TestViewArraySourceCertificateOmission 0.04; TestViewArrayWritesNeedSourceCertificate 0.07;
TestArrayHolesMilestone 1.50; TestArrayHolesAbsentSlotMutant 3.85;
TestArrayHolesRangeErrorMutant 3.84; TestArrayElementKindRuntimeMutant 3.95;
TestArrayViewHolesAbsenceMutant 3.76; TestArrayViewMissingReceiverMutant 4.04;
TestArrayViewPhysicalWriteMutant 4.31; TestOriginalArrayWitnessExtensions 0.01;
TestOriginalArrayWitnessNode 0.01; TestCheckedViewArrayReadMutants 2.07.
Parallel child cases are recorded individually in the same results JSON.
Including build and setup, the lower repair command takes 7.891s and the oracle
repair command takes 23.423s. All finish without failures or skips.

Toolchain and lane

Setup with GOPROXY='https://proxy.golang.org|direct' passes: Node 0.023s,
Go 0.025s, markdown 0.079s, submodules 0.098s, clang 0.214s, Go build 49.954s,
cache 50.162s, done 50.191s. nproc is 5; the cgroup CPU quota is four cores.
Test binaries were deferred by setup. The printed environment is
/workspace/adamic-tools/env.sh. All setup and test output is saved in logs.

The exact required repository-root lane command runs after committed changes:

```
git fetch -q origin main devtools/fast-gate cloud/merge-tree && git show origin/cloud/merge-tree:cloud/integration/lane-checks.py | python3 -
```

After the twelve parallelism repairs it passes:
`lane checks 7.2 s: gofmt and tools on 57 Go files, t.Parallel on 7 test packages; vet 7 packages`.
The final evidence commit receives the same check before push. The V3 remote is
refetched before pushing, and the pushed branch must contain its observed tip.
An ordinary push enforces fast-forward; a concurrent update is merged and
rechecked before any retry. No force push or update to another branch is used.
