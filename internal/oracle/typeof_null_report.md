# typeof null unit

Branch codex/typeof-null starts at origin/main 5d4c801. The compiler change is confined to
internal/native/union.go: an existing ir.Null operand of typeof emits the immortal "object"
string. ir.Null carries a reference representation, not a separate ir.Type, so the test is
on the expression before dispatch by type. JavaScript already emits null and its adamicTypeOf
helper uses JavaScript's typeof. Lowering emits ir.TypeOf for the literal; its only typeof
constant folds concern intrinsic functions. Neither JavaScript nor lowering needed an edit.
The prohibited emission, lowering and oracle files, and the other workers' functions, were
not edited.

V02 and w01 are registered oracle fixtures. V01 is retained as an unregistered probe for the nullable worker. On the baseline v02 prints
"undefined" instead of Node's "object". V01 prints "undefined true\nobject\nundefined\n"
instead of "object true\nobject\nobject\n". W01 passes on the baseline.

After the scoped fix, v02 and w01 agree with source on Node and the JavaScript backend,
including native release builds, ASan/UBSan, and LeakSanitizer. V01 still fails its native
stdout comparison. Its match and exec results are nullable array references, not ir.Null
expressions. Fixing that requires the general nullable path assigned to codex/nullable-references.
V01 is excluded from this branch's oracle registration and counts pending that dependency. This branch does not claim to have fixed nullable references.

The permanent mutant test scans all registered ordinary, lowering fixtures, changes emitted C,
builds every changed program under ASan/UBSan, and requires clean execution and no leaks.
The three uncovered-path mutants each have exactly one catcher, w01, by Node stdout alone:

| Mutant | Changed observation |
|---|---|
| Union narrowed to boolean or undefined marks absence present | `3: false false false` instead of `3: undefined false true` |
| Boxing a missing boolean pair uses the false box | `boolean:false` instead of `undefined:undefined` |
| typeof of a present pair uses undefined | first line becomes `undefined undefined undefined undefined` |

The fix mutant changes the literal's emitted typeof object to undefined. V02 catches it by
Node stdout alone. All four mutant programs compile and run without sanitizer or leak errors.
An initial, broader narrowing mutation also changed array-element unboxing and was discarded;
the final mutation requires a union heap operand.

Counts were regenerated with TestCountsAreRecorded. No existing row moved. New registered rows, in
allocations, frees, retains, releases, peak, regions order:

- v02: 0, 0, 0, 0, 0, 0.
- w01: 14, 14, 6, 27, 6, 0.

Setup: `bash cloud/setup.sh`, then `source /workspace/adamic-tools/env.sh`. Go, clang, Node,
and submodules ready at 0s; build cache warm at 77s; done in 77s. nproc: 5.
Go 1.27.1, clang 20.1.8, Node v24.19.0. No setup workaround was needed.

Commands and logs:

- Baseline and fixed probes: `go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/e4eec87' -count=1 -v`, /tmp/typeof-null-baseline.log and /tmp/typeof-null-fixed.log.
- Mutants: `go test ./internal/oracle -run TestTypeOfNullAndUncoveredMutants -count=1 -v -timeout 15m`, /tmp/typeof-null-mutants.log, PASS (27.784s).
- Counts: `go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts`, /tmp/typeof-null-counts.log, PASS (12.170s).
- Formatting: `gofmt -l cmd internal`, /tmp/typeof-null-format.log, empty output.
- Vet: `go vet ./...`, /tmp/typeof-null-vet.log, exit 0 and empty output.
- Packages: `go test ./internal/native ./internal/javascript ./internal/lower -count=1 -timeout 15m`, /tmp/typeof-null-packages.log, native PASS (74.192s), JavaScript no test files, lower PASS (8.253s).
- Uncached probes: `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/e4eec87' -count=1 -v -timeout 15m`, /tmp/typeof-null-uncached.log, v02 and w01 PASS, v01 FAIL by native stdout only (0.633s overall).

The complete repository gate was not run. General nullable references and a passing v01 are
left to the named dependency; no refusal or representation workaround is introduced here.

After v01 was handed to the nullable worker, its registration and counts row were removed.
`go test ./internal/oracle -count=1 -timeout 30m` passes, including counts and all four mutants.
The complete oracle package log is /tmp/typeof-null-oracle-final.log. V01 remains an unregistered source probe.
