# Scanner nested overload declarations

Scope: item 1 only. Nested callback values, generic sibling callback calls, and
ancestor calls belong to codex/scanner-nested-references and are unchanged here.

The branch starts at origin/main 48c05d09. That commit has no
nested_functions.go. The prerequisite merge 5ec8a706 brings the previously
published nested-functions tip 79c0361009ee5741b1dd249f6a136c02671488f4 into this
feature branch. No main or area branch was written. Merge resolutions retain
main's enum identity handling, cyclic-module readiness and Map value ownership,
and the nested branch's captured-cell readiness and counted-argument closure ABI.
The count-table resolution keeps both independent sets of rows.

The exact probe comes from codex/stage3-scanner-proof dc9f8482,
probes/nested-overload-declaration.a, copied unchanged as
internal/oracle/testdata/scanner_nested_overload.a. Node prints `x`.
The baseline prerequisite merge stops with `a function without a body` at the
signature; it does not reproduce the older scanner scratch's nil-pointer panic.
This is an observation of this baseline, not a claim that the reported panic
was absent on that scratch compiler.

Nested overload signatures are skipped only when the same enclosing statement
list contains a body-bearing declaration with the same checker symbol. The
implementation supplies the single runtime binding and body. Without an
implementation, lowering gives NotYet `a nested function declaration without
an implementation`. lowerBody also rejects any bodyless declaration before
looking at its body, with NotYet `a function without a body`. The tests use a
checked overload program and present its signature alone to the internal path,
so a checker diagnostic cannot mask the guard being exercised.

The probe matches Node in the native release build, ASan/UBSan build, JavaScript
backend, and leak check. The lower package passes in 39.004s. Focused oracle and
implementation-mutant proof pass uncached in 0.660s. Count regeneration passes
in 19.721s. Logs: /tmp/scanner-overload-{before,after,proof}.log,
/tmp/scanner-overload-lower.log, /tmp/scanner-lower-gate.log, and
/tmp/scanner-counts.log. Commands:

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestScannerNestedOverloadImplementationMutantIsCaught|TestNativeAgreesWithNode/internal/oracle/testdata/scanner_nested_overload' -count=1 -v
go test ./internal/lower -count=1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts
```

Three independent mutants were run:

- Replace the implementation's return with the empty string. The binary exits 0,
  has empty stderr and leaks nothing. Node stdout alone catches a blank line
  instead of `x`. This permanent oracle mutant also requires exactly one return
  in exactly one runtime implementation.
- Disable the missing-implementation guard in a Go overlay. Its named-refusal
  test fails because lowering succeeds. Log: /tmp/scanner-mutant-missing.log.
- Remove lowerBody's bodyless guard in a Go overlay. Its test fails with a
  nil-pointer panic at body.Kind. Log: /tmp/scanner-mutant-body.log.

The new overload count row is allocations 1, frees 1, retains 2, releases 3,
peak 1, regions 0. Regeneration also orders the prerequisite's rows by fixture
registration and brings five inherited feature rows onto main's current counting:
nested_array, nested_destructured, nested_mixed and omitted_reader_override each
lose one retain/release pair; omitted_methods loses two pairs. These are differences
from the old nested branch's table after combining with current main, not effects
of skipping overload signatures. No row already present in 48c05d09 changes.

Setup initially overlapped the prerequisite merge and its warm build failed with
missing NestedFunction and EnvironmentCell IR fields while the tree was changing.
The retry passes: Go 0.019s, Node 0.020s, submodules 0.052s, markdown dependencies
0.068s, clang 0.171s, go build 36.798s, test binaries deferred 37.198s,
cache warm 37.202s, total 37.289s. nproc is 5; cpu.max is 400000 100000.
Tool versions: Go 1.27.1, clang 20.1.8, Node 24.19.0. Logs are
/tmp/scanner-setup.log and /tmp/scanner-setup-retry.log. Every build/test shell
sources /workspace/adamic-tools/env.sh. The recursive fetch was stopped after
main arrived; probes were fetched successfully with --no-recurse-submodules.

The broader prerequisite-merge gate runs the whole uncached Node oracle and all
native tests, plus the JavaScript package, in /tmp/scanner-landing-gate.log.
The whole repository stage1 gate is not claimed. This unit does not run a native
createScanner or claim that its remaining scanner blockers are cleared.

The broader gate passed internal/native in 194.607s and reached one failure in
the whole oracle: the imported old TestRealNestedBlockers still expected fixture
08's truthiness to be refused. Current main accepts it. Fixture 08 now belongs
to the positive Node oracle instead; its exact source is unchanged and passes
both backends, sanitizers and the leak check in the focused 0.415s run. Its new
count row is allocations 6, frees 6, retains 26, releases 31, peak 6, regions 0.
Final count regeneration passes in 19.685s. The real nested fixture count is
therefore 9/11 on this branch, because of main's established truthiness policy.
01 and 09 still have their pinned non-null-policy refusals. This test update
implements no part of the separate nested-reference unit.

Vet and gofmt pass with empty /tmp/scanner-vet.log and /tmp/scanner-format.log;
git diff --check passes. A final --no-recurse-submodules main fetch still resolves
to 48c05d091f0a43c31cbe051b1d6578d99eeedf19, already a parent of this feature.
The whole uncached oracle rerun is in /tmp/scanner-oracle-final.log.

Final landing result: the whole uncached oracle passes in 123.135s after the
stale fixture expectation is corrected. The command is
`ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -count=1 -timeout 30m`, with
output saved in /tmp/scanner-oracle-final.log. The complete native package
already passed in 194.607s; lower passed in 39.004s; JavaScript has no standalone
tests and is covered by the whole oracle. Counts, vet, gofmt and whitespace are
green. The table now adds 47 rows relative to the requested main tip: the
previously published nested/omission fixtures, this overload probe, and fixture
08 newly held to Node. Existing main rows remain unchanged.
