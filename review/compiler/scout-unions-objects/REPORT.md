Built: step 17 nullable-object boundaries and checked string mixed-field/local reads on the preserved candidate.
Commits: rebuild b3bfa1b9; base 884d907a; original residuals 7b270f77 and 3230854e.
Commands: own oracle PASS 0.620s, full lower PASS 49.673s, reader guard PASS 17.920s, counts PASS 84.099s, lanes PASS 3.2s.
Mutants: thirteen semantic failures caught; every retained mutant and its catcher are listed below.
Not covered: optional writes/delete, callable union consumption, general object-kind unions, hidden never-array lowering, or the whole gate.

`compiler/scout-unions-objects` was cut directly from `884d907a`. `compiler/scout-unions-main` remains at that SHA. This branch depends on integration accepting that candidate; no unlanded worker branch was merged. The stash bbca9651 remains preserved after applying it.

| Original slice | Delivered residual |
| --- | --- |
| 7b270f77 | Normalize legacy object-null pointers at tagged value boundaries, normalize allowed null when returning to object-pointer storage, check narrowed ordinary fields and locals, handle optional access and object coalescing. Reuse main's existing nullable representations. |
| 3230854e | Check stored string tags before consuming string fields; preserve actual tags during strict local equality and nullish local reads; explicitly stop incompatible shared nullable storage and lookups without separate tags. |

Main's nullable ABI fix 4df3a0b7 already represents the fresh array witness correctly. Restoring the old nonempty-array contextual condition did not change its Node agreement, so the old source's array-condition change is omitted. A second spread-specific guard was also unnecessary: the first guard already checks nonfresh spread storage. Both surviving mutation attempts are retained as evidence and excluded from the thirteen caught mutants. The six storage probes confirm the resulting array, field, callable, array-spread, object-spread and lookup outcomes independently against Node.

Main's existing `regexReplacementArgument` adapter has an explicit per-callback argument and return convention. The new shared-view guard exempts that established adapter. The initial counts run exposed its accidental rejection; the correction restores the existing RegExp callback fixture's Node agreement. A mutant removing this exception makes that fixture fail with a lowering stop. Stored callable views remain stopped.

New checked nullable-object field reads also include the declared Union slot when the checker says the field is exactly null. `scout_nullable_object_null_field.a` proves both a direct call and a narrowed conditional call pass null to an object-or-null consumer. Restricting this hook to Object reads produced invalid pointer arguments in the preserved work; that compiler failure is not counted as a semantic mutant.

All commands source `/workspace/adamic-tools/env.sh`; stdout and stderr go to the adjacent logs. The final source was committed before integration lane checks.

| Command | Result and log |
| --- | --- |
| `timeout 300 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/scout_(nullable_object|string_number)|TestNullableObjectStorage' -count=1 -v -timeout 90s` | PASS 0.620s, oracle-delivery.log. Six admitted fixtures compared with Node and both backends; native sanitizer, release and leak checks. The first admitted-fixture run was PASS 1.210s with native misses; the expanded witnesses also passed in 2.024s. |
| `timeout 600 go test ./internal/lower -count=1 -v -timeout 570s` | PASS 49.673s, lower-full.log. Entire package, monitored background session. |
| `timeout 300 go test ./internal/ir -run TestCallTargetReaders -count=1 -timeout 90s` | PASS 17.920s, reader-guard-final.log. Earlier run PASS 17.150s. |
| `timeout 600 go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 570s -args -update-counts` | PASS 84.099s, counts-final.log. Initial failure is in counts.log and was corrected. |
| `timeout 300 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/regexp_replace/callback.a' -count=1 -v -timeout 90s` | PASS 0.185s, regex-adapter-final.log; Go's hierarchical pattern also selects the two neighboring replace fixtures. |
| `timeout 300 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/(scout_boxed_scalar_fields|taste_stage3_representations).a' -count=1 -v -timeout 90s` | PASS 0.741s, moved-counts-node.log. |
| `git fetch -q origin main devtools/fast-gate cloud/merge-tree && git show origin/cloud/merge-tree:cloud/integration/lane-checks.py \| python3 -` | PASS 3.2s, lane-checks.log: gofmt and tools on 20 Go files, t.Parallel on 4 test packages, vet 4 packages. Fetch bounded at 120s, lane process at 300s, pipefail enabled. |

Each source mutant invokes `timeout 180 go test -overlay <name>.overlay.json <package> -run <witness> -count=1 -v -timeout 90s`, through bounded 900-second driver runs. Every retained mutant exits 1 with a semantic test failure. Sources are `.go.txt`, not compilable Go files under review. The runner's initially narrow parser missed valid missing-panic and exit-code assertions; the raw failures were inspected and the parser corrected. The rejected local-rule deletion causes a native-emitter panic and is explicitly excluded; its replacement changes an actual undefined read into a string tag and is caught by runtime output disagreement. No compiler panic, clang failure, timeout or surviving mutant is credited.

| Mutant | Witness and observed catcher |
| --- | --- |
| object-boundary | scout_nullable_objects: disabling pointer normalization changes native null into undefined; stdout disagrees with Node. |
| object-tag | scout_nullable_object_stale_field: allowing null as a nonnull object removes the required native panic, exit 0 versus expected 70. |
| object-optional | scout_nullable_objects: replacing the absent result with an object containing mutant text changes both backend outputs. |
| object-null-field | scout_nullable_object_null_field: removing null normalization prints object instead of null natively. |
| string-check | scout_string_number_stale_field: dropping the stored-tag failure condition removes the expected panic. |
| string-tag | scout_string_number_fields_locals: requiring number for a string read panics in both backends instead of completing as Node does. |
| local-nullish | scout_string_number_fields_locals: corrupting a narrowed undefined read into a string produces mutant text in both backend outputs. |
| regex-adapter | regexp_replace/callback: removing the existing adapter exception creates a stage 0 stop for the accepted fixture. |
| shared-view | TestNullableObjectStorageArray: disabling the view guard admits the incompatible storage view; the explicit-stop assertion fails. |
| lookup | TestNullableObjectStorageLookup: disabling the lookup guard admits a lookup without separate null/undefined tags; the explicit-stop assertion fails. |
| local-equality | scout_string_number_fields_locals: removing the equality-observation rule panics after a captured local changes from string to number. |
| object-local-runtime | scout_nullable_object_stale_local: bypassing the nullable-local checker removes the expected stale-null panic. |
| object-local | scout_nullable_objects: bypassing the nullable-local checker changes accepted nullable consumption and produces an exit-code disagreement. |

The consolidated result is mutants-summary.log; exact substitutions are in mutants.json. The controlled counts overlay is diagnostic evidence, not a runtime check mutant.

Counts add six rows for the admitted fixtures. The six separately tested storage probes stop before backend generation and therefore have no native allocation counts. Two existing rows move:

| Existing row | Previous retains/releases | New retains/releases | Cause |
| --- | --- | --- | --- |
| scout_boxed_scalar_fields.a | 95 / 150 | 114 / 169 | String reads now load the stored Union once through the checked helper; preserving and returning the reference adds matching retains/releases. |
| taste_stage3_representations.a | 11 / 34 | 13 / 36 | The consumed string field now uses the same checked helper and its reference lifetime. |

Allocations, frees, peak and region counts are unchanged in both rows. `timeout 300 go test ./internal/oracle -overlay review/compiler/scout-unions-objects/counts-cause.overlay.json -run '^TestScoutStringRouteCountsCause$' -count=1 -v -timeout 90s` passes in 0.389s and records exactly the previous rows after disabling only the new string route. This directly identifies the cause instead of attributing an unexplained count shift.

Every added top-level test and admitted fixture leaf is below 60 seconds on the four-CPU quota:

| Leaf | Seconds |
| --- | ---: |
| TestNullableObjectSharedStorageStaysNotYet | 0.38 |
| TestNullableObjectUnionKeepsStrongCyclesRefused | 0.10 |
| TestNullableObjectStorageArray | 0.34 |
| TestNullableObjectStorageField | 0.37 |
| TestNullableObjectStorageCallable | 0.40 |
| TestNullableObjectStorageArraySpread | 0.44 |
| TestNullableObjectStorageObjectSpread | 0.43 |
| TestNullableObjectStorageLookup | 0.22 |
| scout_nullable_objects fixture | 0.18 |
| scout_nullable_object_stale_field fixture | 0.08 |
| scout_nullable_object_null_field fixture | 0.09 |
| scout_nullable_object_stale_local fixture | 0.06 |
| scout_string_number_fields_locals fixture | 0.13 |
| scout_string_number_stale_field fixture | 0.08 |

Earlier uncached fixture observations were also below two seconds each. The diagnostic overlay-only count test takes 0.36s. No run timed out compiling in this unit. Every long command was bounded and background sessions were checked.

Setup used `GOPROXY='https://proxy.golang.org|direct'`, then `timeout 300 bash cloud/setup.sh`. It succeeded: Go 0.036s, Node 0.036s, markdown dependency validation 0.009s and ready 0.145s, submodules 0.144s, clang 0.273s, Go build 56.938s, test binaries deferred 57.107s, cache warm 57.110s, total 57.163s. nproc=5; cgroup CPU quota=4. setup.log and nproc.log contain the evidence.

This advances step 17 by delivering the last executable scout residuals on the candidate base. Optional writes/delete and general callable/object-kind union checks retain their existing scope. Hidden never-array work is not duplicated. No whole oracle package, full gate, corpus byte credit or PR is claimed.
