Built named properties plus a mutable string index signature on dictionary storage.
Commits: base 7f98e622; records merge ee991b43; main merge 696d2a36; implementation is the commit containing this report.
Validation: six new fixtures, Node in both backends; uncached regression oracle passes; touched-package checks below.
Mutants: named/index type confusion, absent-as-present, removed member guard, unsafe writes/views, and alias readiness all caught.
Not covered: full CompilerOptions payload, object-kind union narrowing, or a rerun of the 99-row census.

## Representation and evidence

The CompilerOptions reduction is [records_named_options.a](../../internal/oracle/testdata/records_named_options.a): an optional numeric target beside a mutable string index signature. It now lowers. Named keys keep their checker-declared types; dynamic reads use the index type. All entries use existing Record dictionary storage, with no synthesized optional entries. Writes, deletion, missing reads, Object.keys integer/string ordering, explicit undefined presence, and `in` are held to Node. [node.json](node.json) records all six independent Node executions.

The heterogeneous fixture uses a number/string/boolean/undefined index payload. Narrower named primitive reads get a runtime member-kind check before unboxing. A call can invalidate a checker narrowing: records_named_invalidated prints changed1 on Node but both checked backends stop loudly with exit 70. This is an intentional check, not an oracle agreement claim. A dynamically deleted required member can still be observed with typeof or an undefined comparison without premature unboxing.

Dynamic writes must satisfy every narrower named contract they might hit. Mutable aliases cannot erase those contracts or readonly names. Tests retain named refusals for unsafe writes/views, readonly and numeric signatures, fixed-object/dictionary conversions, boxed object-kind narrowing, and unsupported nullable payload storage.

The full upstream CompilerOptions index payload contains null, undefined and object-kind values. On this base those have independent slot/tag representation boundaries. This unit lifts the named-plus-index shape stop, but does not claim the 11 original census rows now emit. The user-supplied combined area base is still needed for that measurement.

## Integration

Started from origin/area/compiler 7f98e622. Records 456c981b was absent and was merged in ee991b43. Current main 6f16a169 was merged in 696d2a36 without rebasing. The earlier enum/catch/class branch conflicts were not resolved in this unit. No main or area branch was pushed.

The records merge initially regressed three existing fixtures: library_method_values, method_coverage_object_descriptors, and method_coverage_object_statics. The corrected dispatch retains the area's proven Object.apply and non-object Object.call adapters and routes actual dictionary arguments through records operations. Const alias copies read the original binding with readiness instead of inventing an initialized marker. All three fixtures now match Node; an alias copied before initialization also stops like Node.

## Mutants actually run

- Named read through the index type: run-mutants.py bypasses recordReadType; TestNamedRecordReadTypes fails with named read used index type 10.
- Absent treated as undefined-present: TestNamedRecordAbsentEntryMutant inserts an undefined target entry into the real RecordLiteral. Native and JavaScript stdout differ from Node.
- Removed declared-member kind guard: TestNamedRecordTypeGuardMutant removes the real panic branch. JavaScript runs on as Node does, violating the exit-70 pin.
- Unsafe dynamic named writes: run-mutants.py bypasses recordNamedWrite; the named-write refusal pin fails.
- Erased named alias contracts: run-mutants.py bypasses sameRecordNamedContracts; the view refusal pin fails.
- Removed alias-copy readiness: TestNamedRecordAliasReadinessMutant clears the real Read.Checked bit. Both backends exit 0 where Node stops 70.

Source mutations were restored after each run. Reproduce the three source mutants with `source /workspace/adamic-tools/env.sh; python3 stage3/named-index-records/run-mutants.py`. Oracle mutants are part of TestNamedRecord. Logs are in [evidence](evidence/).

## Validation

All output was redirected directly to logs. The complete repository gate was not run.

- `go test ./internal/lower ./internal/ir ./internal/fresh ./internal/flow ./internal/javascript ./internal/native -count=1 -timeout 30m`: lower, ir, fresh and native passed; JavaScript has no package tests. The initial flow run found the three integration regressions above. Native passed in 1646.305s; fresh in 354.023s; ir in 30.798s.
- Final complete lower package: `go test ./internal/lower -count=1 -timeout 30m`, PASS 121.779s. Two subsequently added representation refusal probes pass separately in 0.581s; production code unchanged.
- `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNamedRecord|TestRecord|TestPartialRecord|TestDetachedOwn|TestNativeAgreesWithNode/internal/oracle/testdata/(records_|detached_own_|library_method_values|method_coverage_object_|typed_array|library_object|literal_optional|enum)' -count=1 -timeout 30m -v`: PASS 93.509s; native 225 misses/0 hits, Node 233 misses/0 hits. This includes both backend checks and mutants.
- `go test ./internal/native -run 'TestRecord|TestPartialRecord' -count=1 -timeout 10m`: PASS 180.833s.
- Flow SSA and affected path checks: PASS 64.249s; 4333 functions, 648 phis, 4572 values, 12940 uses. Liveness plus affected mutation-range checks: PASS 68.561s. Complete corrected `go test ./internal/flow -count=1 -timeout 30m`: PASS 220.326s.
- `go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 30m -args -update-counts`: PASS 137.561s.
- `go vet ./...`, gofmt check, and git diff --check passed.

Setup was rerun after the records merge: an initial overlapping attempt encountered conflict markers and was discarded. Successful bash cloud/setup.sh timings: Node 0.067s, Go 0.079s, submodules 0.235s, markdown 0.255s, clang 0.479s, build 253.405s, warm cache 253.932s, total 254.192s. nproc=5, cgroup quota=4 CPUs. Environment: /workspace/adamic-tools/env.sh; Node 24.19.0, Go 1.27.1, clang 20.1.8.

## Counts

Counts are allocation/free/retain/release/peak/live. Six new fixture rows were generated:

| Fixture | Counts |
| --- | --- |
| options | 8/8/1/6/4/0 |
| operations | 19/19/42/48/7/0 |
| union | 14/14/30/37/5/0 |
| invalidated (checked panic) | 3/1/11/7/3/0 |
| intrinsics | 14/14/13/28/7/0 |
| alias_unready (TDZ stop) | 0/0/0/0/0/0 |

Two existing rows changed: object_descriptors 47/47/88/127/7/0 to 46/46/61/98/6/0; library_method_values 137/137/154/295/42/0 to 136/136/152/291/41/0. The own-property intrinsic uses a boolean readiness marker instead of an allocated closure, and proven supported calls bypass the generic callable-token adapter. The regexp_tree row moved to generated order with unchanged counts 91/91/81/85/35/0. All other existing rows are unchanged from the merged base.
