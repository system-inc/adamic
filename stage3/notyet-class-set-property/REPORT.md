Built owned union field storage, checked narrowed reads, and concrete literal method replacement.
Commits: first group is the commit containing this report; compiler base b410340dc8f889b5799c3bc519117c63def3aa24; replay merge f943bdf03fbe492b65731714a7437a7980db228c.
Checks: focused oracle, lower and native tests pass; JavaScript package has no standalone tests; counts refreshed.
Mutants: all ten fail their fixtures; details below and in mutants-first-group.json.
Uncovered: structural/class/library method replacement and the remaining kinds are still being processed.

The branch starts from the newest origin/area/compiler SHA resolved at setup. Both adapted inputs match all 81 source hashes in stage3/meter/runs/20261008T035244Z.latent-full/tsc/source-manifest.json. Replay results measure a checker-rejected entry-root program; successful unit replay is not whole-compiler compilation.

Setup: GOPROXY='https://proxy.golang.org|direct' bash cloud/setup.sh completed in 880.607 seconds. Go .248s, Node .782s, clang 1.830s, markdown 5.354s, ready 6.252s, submodule 57.013s, build 880.305s, tests deferred 880.522s, cache 880.526s. nproc=5; CPU quota 400000/100000. Environment: /workspace/adamic-tools/env.sh. Setup log: /tmp/adamic-class-set-property-setup.log.

First group observations:

- Method replacement (14 roots): concrete literal methods now replace owned closures correctly, including replacing a capturing arrow and a second replacement. Class methods still fail the production unbound-method refusal. core.ts:1544/1545 retain the census stop after an earlier uncheckable Map cast; sys.ts:1382 retains the stop and needs structural slot support.
- true | Node | undefined (3 roots): parser.ts:1341 now has no findings in its selected unit. utilities.ts:8987/8992 no longer report union field storage; next stops are the binary expression at 8987:48 and direct case declaration at 8998:13. The assignment at 8987 was already masked by its RHS binary stop before this change.

Union slots retain owned boxes. Initial literal values use declared storage. A read through a readonly widened scalar view consults the actual shape, whose identity now includes each field's representation. Narrowed reads hold the slot value once and check the current member, including object/array/map/typed-array tags, before casting. Optional missing fields and optional receivers retain their distinct semantics. No runtime C files changed.

Commands (all output redirected to the corresponding /tmp/adamic-class-set-property-*.log file):

```
go test ./internal/oracle -run '^TestNativeAgreesWithNode/internal/oracle/testdata/class_set_property_' -count=1
go test ./internal/lower -run '^(TestOptionalWidening|TestCensusRestMutableElements|TestClassFeatures|TestCheckedCastProofAndElision)' -count=1
go test ./internal/native -run '^(TestUniformFieldsMatchNode|TestRuntimeFieldLayoutsAreIncluded|TestOptionalWriteMissingSlotRemainsChecked|TestInheritanceMemoryPlans)' -count=1
go test ./internal/javascript -run '^TestClassSetProperty' -count=1
go test ./internal/oracle -run '^Test(NativeAgreesWithNode/internal/oracle/testdata/class_set_property_|ClassSetPropertyMethodRefusal)' -count=1
go test ./internal/oracle -run TestCountsAreRecorded -args -update-counts
```

Oracle fixtures use source Node, emitted JavaScript on Node, native ASan/UBSan and release builds; terminating fixtures also check ownership counts/leaks. The alias-narrowing fixture checks an intentional exit-70 guard in both backends while source Node continues.

Each production mutant was applied temporarily, tested with its specific oracle fixture, and restored:

| Mutant | What caught it |
| --- | --- |
| Restore represented-method stop for literal slots | Lower NotYet |
| Treat assignment LHS as an unbound method read | Lower Refused |
| Restore union field storage stop | Lower NotYet |
| Read narrowed union using scalar slot layout | Backend/native disagreement |
| Disable union member check after an aliasing call | Source-like continuation versus required guard |
| Merge number and boolean shapes using reference bits alone | stdout differs |
| Borrow union field without retaining it | ASan heap-use-after-free |
| Check a plain object using the array tag | Backend/native stdout differs |
| Treat missing optional union field as required | ASan invalid field access |
| Dereference optional receiver without null guard | UBSan null member access |

The worker ownership audit fetched origin/codex/notyet-* and examined commits after the compiler base. Edits in object.go, census_small.go and refusals.go are minimal hooks or shared helpers; other workers' property, arrayMethodArguments, representation and callClosure functions were not edited.

Structural method follow-up:

Structural MethodSignature writes now carry SetProperty.OwnMethod. Both backends evaluate receiver then RHS, and require an existing own data slot before replacing its owned closure. An inherited class method would create an expando, so both backends stop with the same fixed-shape guard; the source Node fixture demonstrates the different JavaScript behavior for a ruling. Direct class method writes and inherited library methods retain their refusal.

core.ts:1544/1545 now reach `reading map`, after the same earlier uncheckable cast. Those two root sites are census echoes of that failed initializer, not evidence that Map expandos have been implemented. sys.ts:1382 now reaches the structural call at 1387:49, owned by the statics worker. The representative method-storage stop is removed.

The two additional oracle fixtures pass in both backends, sanitizers and release. All four additional mutants fail: rejecting MethodSignature storage stops lowering; dropping OwnMethod from the IR produces a backend/native disagreement; disabling either backend guard produces a disagreement. Focused tests also passed in internal/ir (call targets), internal/native (fields and inheritance), internal/lower (class features and optional widening), internal/javascript (no standalone tests), and oracle. TestCountsAreRecorded -args -update-counts passed in 46.121s. No runtime C files changed.

Array and any follow-up (after the requested checked non-null merge):

Merged c41c0e062e99da37820f822968d4df1b48cdaee7 at b6415ddab8eeca895d29284d13e1da332fac3ecb. The merge completed without conflicts; pending unit work was stashed and restored. Focused lower non-null controls and the new TypeScript checked-non-null oracle controls pass. All 14 representative replays were repeated against the merged compiler.

Array length decrement removes exactly the last element and releases it using existing ArrayPop. Decrementing an empty array throws a catchable RangeError with `Invalid array length`; assignment of literal zero clears all elements while preserving aliases. A minimal shared hook in class.go updateProperty routes only array `.length--` to the new helper. Other length assignments, growth, holes, and general array properties remain stopped. tracing.ts:164 no longer reports assigning a field of a value; its unit retains only debug.ts:213:28, a dependency value of type unknown.

any (2 roots) remains refused for a language ruling: docs require proven types, and an any field has no storage proof. A reduced source fixture runs on Node and pins the exact existing `storing any in a field` stop. tsbuildPublic.ts:2080 retains that stop; 2068 reaches earlier any and structural-call stops. This does not claim to lower any.

Five mutants fail: force any to a number slot (boundary assertion fails); remove the empty-array RangeError (stdout differs); omit decrement (stdout differs); omit clearing (stdout differs); evaluate the receiver twice (stdout differs). Array fixtures agree with Node in JavaScript/native, sanitizers/release and ownership checks. Focused lower class/rest checks pass; merged counts refresh passed in 52.567s. This commit adds no runtime C changes.
