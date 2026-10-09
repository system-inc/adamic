# Error undefined-message correction

The optional-message NotYet is removed. The stored `census_never_rest_marker.a` is restored to `new Error(message)`, with no difference from the original main program. Editing that fixture to avoid the new stop was wrong because it changed the shape under test.

Lowering now preserves omitted and explicit undefined messages instead of converting omission to an empty argument. The native constructor supplies an immortal empty string for reads, stored in a private fallback slot, so undefined creates no public own `message` field. Explicit `''` uses the ordinary own field. Public and checked lookups resolve the fallback; it participates in the shared readiness and string-representation checks. The JavaScript backend receives undefined and uses JavaScript's Error constructor. No Error reflection admission changes.

The supplied `4ddd17f_opt_3.a` now compiles and prints `[] [m]` in both backends, matching Node. The original census fixture compiles unchanged and retains its pinned Node output. Its normal execution passes a defined message, so an additional driver appends a call to its existing `fail()` entry point, catching and printing its undefined-message result. The stored corpus program remains untouched; this driver exercises the previously unexecuted optional branch.

A new `.a` fixture covers omitted, explicit undefined, explicit empty and nonempty strings, and a message expression with an observable evaluation count. Both backends match Node under the native sanitizer and leak checks. A runtime probe separately holds message reads and both own-property APIs to Node, including alternating present and absent shapes and cached, optional and checked reads. Direct Error.hasOwnProperty and Object.hasOwn(Error, ...) remain NotYet at the language boundary.

The old-native-read mutant restores null message storage and removes the unused fallback constant to keep the mutant valid C. It must fail both the supplied optional-message witness and the census omitted-argument driver at the native null read. A second mutant always creates the own public message field; it must fail Node parity while its empty string remains valid. The five other containment mutants remain unchanged.

Commands, outputs and per-leaf seconds are recorded in error-fix-fixtures.log, error-property.log, error-fix-mutants.log, mutant-optional-error.log, mutant-error-own-message.log, error-fix-counts.log, error-fix-vet.log, error-fix-lane-checks.log and error-fix-seconds.json. The original containment report and before/after JSON are historical evidence; their optional-message refusal and edited-marker descriptions are superseded by this correction.

Final validation on main acd029d9 plus this fix:

- `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestMiscompile2A' -count=1 -timeout 90s -v`: thirteen leaves pass, package 4.184 s, both backends, ASan/UBSan and leaks.
- `go test ./internal/native -run '^Test(ErrorUndefinedMessageHasNoOwnProperty|RuntimeFieldLayoutsAreIncluded)$' -count=1 -timeout 90s -v`: passes, package 1.503 s; the new runtime probe takes 1.42 s.
- `go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 30m -v -args -update-counts`: passes, package 267.289 s. The marker row returns to main's original 10 allocations, 10 frees, 14 retains, 34 releases, 8 regions, zero checks. Only the two newly admitted Error witnesses add rows; the controls row is unchanged.
- `go vet ./internal/lower ./internal/native ./internal/oracle`: passes.
- `python3 review/compiler/miscompile-fxspptb-2a/mutants.py` and the final `--optional-only` run: all five stops and both runtime semantics mutants caught. The final old-native-read runs are separate, uncached commands, with UBSan symbolization disabled but sanitizer checks enabled. Each final negative leaf is below 60 s. The earlier parallel attempt took 70.84 s for the census driver and is retained in parallel-error-mutant-attempt.log; the separate final driver took 4.19 s.

Every new or changed test's final seconds are in error-fix-seconds.json. No evidence source under this branch's review directory is compilable Go. Full packages and the full gate were not run.

Integration lane checks pass in 4.0 s: gofmt and tools on six Go files, t.Parallel on two test packages, vet on two packages. The final main merge is 6d53fa22 and changes only unrelated stage 1 test product declarations.
