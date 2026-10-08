# JSON union dispatch

Based on origin/library/l2-land 23538b4f6cb40db0dc33990279eaf39cb6aeef4f. This work predates the encoder consolidation under docs/library-json-contract.md; it is available for that owner to reuse rather than a claim of complete JSON support.

JSON.stringify now dispatches primitive and supported container union arms using runtime storage metadata. Tagged unions distinguish null from undefined. Object shapes retain actual fields and array producers retain their element schema; tuple arms serialize as arrays. The fixture covers string | number | null, undefined, nullable scalar widening, plain objects, arrays, tuples, functions, property omission, replacers and indentation. Top-level undefined remains string | undefined.

Assumption: container union dispatch is admitted only for closed, supported origins whose runtime metadata is available. Classes, accessors, toJSON, dynamically added properties and unsupported container-producing calls remain compile-time refusals. Symbols and stored heterogeneous union arrays remain language limitations. Sparse Array index-fill work was not started, per the owner's revised scope.

## Verification on Linux

Node v24.19.0 is the reference. Commands sourced /workspace/adamic-tools/env.sh and used GOPROXY=https://proxy.golang.org|direct. Test output was written to logs.

- go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/json_stringify_unions|TestJSONStringifyOracleCatchesWrongUnionArm' -count=1 -timeout=10m -v: PASS, 29.096s (/tmp/json-union-tuple.log). Fixture agrees with Node on native and JavaScript backends.
- go test ./internal/lower ./internal/native ./internal/ir ./internal/flow -count=1 -timeout=15m: PASS (105.183s, 351.484s, 8.831s, 244.267s respectively; /tmp/json-union-packages2.log).
- Nullable and Date JSON regression fixtures plus the mutant: PASS, 1.203s (/tmp/json-union-nullable2.log).
- go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout=20m -args -update-counts: PASS, 122.741s (/tmp/json-union-counts.log). Counts regenerated on Linux. New union fixture: 244 allocations, 244 frees, 102 retains, 395 releases, peak 10, zero regions.
- bash cloud/setup.sh --wasi-sdk: PASS, 75.116s. ADAMIC_ORACLE_WASI=1 go test ./internal/oracle -run 'TestWASIAgreesWithNode/internal/oracle/testdata/json_stringify_unions' -count=1 -timeout=10m -v: PASS, 16.260s; includes runtime archive build (/tmp/json-union-wasi.log).
- Initial setup total 30.041s: Go .085s, Node .093s, submodules .081s, clang .465s, build 29.861s. nproc: 5.

The wrong-arm mutant converts a number arm to a boxed string before serialization. It compiles, exits zero and has no leaks or stderr. Only comparison with Node catches stdout differs.

The whole oracle was not run; targeted fixtures, package gates and the Linux counts gate were run. No macOS execution was performed.

## JSON corpus survey, unchanged

Pinned test262 c8c798898646638cd0c24879f8e0374e847e7d74, stock tsc 6.0.3, built-ins/JSON: 48 pass, 28 refused, 26 not-typescript, 63 skipped; total 165, zero failures/crashes. The earlier 7/69 count predates l2's JSON work. No Array-group changes were made, so there is no after-count claim.

Refusals: 10 dense-prefix indexed writes; 5 var; 4 diagnostic-code mismatches (two with an additional TS2345); 2 first-class JSON identity; 2 String prototype overwrites; 2 detached method reads; 1 sparse Array length; 1 any value; 1 delete.

Standing landing rule acknowledged: task branches only, never push main or force-push.

Final lower package rerun after the origin guard change: PASS (/tmp/json-union-final-lower.log). An initial invocation had a malformed timeout flag and did not execute tests; the corrected invocation above passed.
