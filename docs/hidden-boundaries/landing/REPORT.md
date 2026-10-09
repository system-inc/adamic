Built the step 30 hidden-boundary landing candidate on main 031a1259, merging the three requested pins in order.
Merge commits: 8fc4be59, 5cca2a8c, be91f0c6; integration fixes: 28c95758, fb98f4be, ad2e8902.
Every requested sweep passed; scoped a-check covered 53 files, and the census reveals 19,270 bytes across eight intervals.
All seven landing mutants were caught; Node/output, stop, refusal, and representation witnesses are recorded below.
Remaining stops belong to checker options, checked views, brands, and record storage; no further boundary mechanism was built.

## Candidate and measurement

Branch compiler/hidden-boundaries starts at origin/main 031a1259bc7973934792dc6cb1bd4074fc2204b9.
Ordered source pins: hidden-06 702c3ecb, hidden-08 b3b32769, overload-results e9fd9bbb.
Production compiler code measured is fb98f4be, identical to ad2e8902 except recorded oracle row ordering.
The report commit adds evidence only. Source census pin is 388096e6a83a4e9d287fb827f793c599ba1bf0ad.
All 82 adapted hashes were verified. Both measurements use the same independently overlapping
units in the eight assigned intervals, with the full checker project loaded. This is a scoped
measurement, not a corpus-wide compilation or an addition of isolated branch claims.
The replay explicitly reports measurement on a checker-rejected program; body exclusions
and their first checker diagnostics remain part of the hidden-byte accounting.

| Region | Main hidden | Candidate hidden | Revealed bytes | Next stop |
|---|---:|---:|---:|---|
| hidden-01-large | 13,625 | 13,625 | 0 | declarations.ts:1678:110 TS2345; checker #k881crd |
| hidden-01-small | 6,899 | 6,899 | 0 | declarations.ts:1678:110 TS2345; checker #k881crd |
| hidden-05-large | 5,583 | 5,583 | 0 | visitorPublic.ts:123:5 visitNode parameter, from es2018.ts:828:9; checked views |
| hidden-05-small | 5,420 | 5,420 | 0 | visitorPublic.ts:123:5 visitNode parameter, from es2015.ts:3193:9; checked views |
| hidden-06 | 11,417 | 1,020 | 10,397 | visitorPublic.ts:194:1 visitNodes result covariance, from esDecorators.ts:1250:13; checked views |
| hidden-13 | 7,289 | 6,965 | 324 | es2017.ts:764:38 TS18048; checker #k881crd |
| hidden-14 | 7,102 | 5,983 | 1,119 | utilities.ts:5041:1 result._expressionBrand, from utilities.ts:11338:9; brands |
| hidden-08 | 9,790 | 2,360 | 7,430 | utilities.ts:9045:23 union or optional field read; record storage |
| Total assigned intervals | 67,125 | 47,855 | 19,270 | |

[Exact diagnostics, paths and raising functions](next-stops.json), [byte intersections](regions.json),
[main raw census](main.jsonl.gz), and [candidate raw census](stack.jsonl.gz) preserve the observations.
01's excluded checker body begins at declarations.ts:1387:5; 13's begins at es2017.ts:738:5.
05, 06 and 14 name both the region's call and the dependency that rejects it.
08's region boundary starts at utilities.ts:9045:13 and raises at its field read, 9045:23,
through lowering.property / notYet. No stop was advanced after the user's boundary ruling.

## Every textual conflict

Merge 8fc4be59: internal/oracle/counts.md kept all main function-value rows and both TNode rows.
Merge 5cca2a8c: internal/oracle/counts.md kept those rows and both never-array rows.
Merge be91f0c6 resolved four files:

- internal/lower/expression.go uses one shared constraintStorage implementation, after concrete mapper and substitution precedence, keeping main view observations. Instantiated nested calls retain overload result checks inside namespace readiness, so readiness executes before call evaluation.
- internal/lower/hidden_boundary_generic_tnode_test.go keeps parameter-storage, constraint, mapper and substitution assertions; unsupported fixture paths point to testdata/notyet.
- internal/oracle/hidden_boundary_generic_tnode_test.go keeps the exact 3:36 NotYet assertion and relocated registry paths. Duplicate root generic-value and mutation fixtures were removed; one copy remains under notyet.
- internal/oracle/counts.md keeps main inference/function-value, TNode, never-array and incoming overload-result rows, deduplicating identical TNode rows without discarding numeric values.

Each resolution is also in its merge commit message. No runtime representation or native emitter was changed.

## Integration reds resolved

The automatic merge admitted an extra implementation binder that main's result guard refuses.
The restored guard rejects that unproved relation while preserving main's nullable result boundary;
an initial overbroad guard was caught by overload_original_node.a and corrected in fb98f4be.
The binder refusal and original-node native/JavaScript oracle now pass.

The first scoped a-check found 12 deliberate refusal fixtures without their expected-refusal headers.
Only exact a-check headers were added; [header changes](a-check-header-fixes.json) and
[initial results](a-check-initial.json) show the refusals unchanged. Final a-check reports
23 checked and 30 expected refused. Its standing semantics classify pinned NotYet programs as
checked; the separate exact NotYet test verifies their actual stop.

Two existing stage3 cycle records became Compiles after the never-array fix:
05_safe_load_read and 06_import_order_mutant. The targeted record update required Node and
sanitized native agreement before updating status.json. The safe fixture prints 0 0 true;
the import-order mutant exits 70 with the source initialization ReferenceError.
04_directory_callback remains pinned NotYet at 9:71 for a generic function value.

hidden_boundary_generic_tnode_value.a remains under internal/oracle/testdata/notyet and
stops at 3:36 with "stage 0 can't lower a generic function as a value yet". The exact NotYet
test requires no IR. It is not in the flow lowering corpus, and full internal/flow passes.
Counts refresh moved five rows to registry order; [numeric changes](count-changes.json) is empty.
The final pure TestCountsAreRecorded passes. Initial failure logs are retained separately.

No requested check remains red. A nonrequired whole historical diff whitespace check reports
old imported markdown hard breaks and patch context whitespace; those raw evidence artifacts
are preserved. Changed compiler source and new landing documentation pass whitespace checks.

## Verification

[Commands and exit codes](checks.json) link every final log. Tests wrote directly to log files.

- go build ./... and go build ./cmd/adamic passed; go vet ./internal/... passed.
- Scoped a-check used only added or changed .a files against origin/main: 53 files passed.
- go test ./stage1/... -run 'Gap|Gaps|Probes' passed.
- go test ./stage3/fixtures passed (63.239s).
- go test ./internal/flow passed (218.489s).
- go test ./internal/oracle -run TestCountsAreRecorded passed (22.119s), after the recorded-count update.
- Focused Node/native/JavaScript oracle tests passed (10.382s), including sanitized native builds, checked stops, refusal pairs, and main overload fixtures.
- Focused lower tests passed (10.404s). Checked-site CLI records passed (1.480s).

Toolchain setup passed: go 0.019s, node 0.020s, submodules 0.060s, markdown 0.064s,
clang 0.145s, build 73.615s, deferred tests 73.709s, cache 73.710s, done 73.736s.
nproc is 5; CPU quota is 4. Machine is Linux x86_64; Go 1.27.1, Node 24.19, clang 20.1.8.
GOPROXY was https://proxy.golang.org|direct; environment was sourced from /workspace/adamic-tools/env.sh.

## Mutants

[Seven outcomes](mutants.json) and individual logs show each mutant failing its selected test:

- drop-indirect-check: a returned liar's wrong result escapes with exit 0 instead of the required 70.
- skip-single-signature-check: the narrow-signature call lets the liar escape with exit 0.
- erase-TIn-to-Node: helper forwarding admits a wrong visitor input instead of stopping.
- admit-unproven-invocation: the overloaded helper's fabricated input escapes its invocation check.
- nonempty-never-array: Node prints length 0; mutated native prints 1.
- drop-extra-binder-guard: main's required extra-binder refusal disappears.
- erase-TNode-object-brand: the constrained object loses its tagged representation; the kind assertion fails.

The last two are static refusal/representation witnesses, not Node output claims.
No check failure was inferred from a successful mutant build alone.

## Replay artifacts

measure.py uses the pinned census union/subtraction tool; record-stops.py keeps checker exclusions
separate from lowering boundaries. main-overlay and stack-overlay archive the exact measurement
files; overlay JSON contains the original absolute scratch paths. Measurements disable production
lowering only in scratch overlays and select the same eight intervals. No cohere source was copied.
The a-check runner uses archived fast-gate.py.txt from tooling pin 89cbe74a; run-mutants.py uses
isolated Go overlays. The production compiler receives no measurement modifications.
