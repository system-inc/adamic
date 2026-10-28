# Evidence: compiler/assigned-symbol-memo (#pv935yw)

The change under test is `b65d7026ecebcca80ac12499103a40aa993345d3`, one commit on main `5f0b2152`. This branch is
that sha plus this directory. Nothing in the change was edited; mutants were applied in the working tree and restored.

## The box

From `box.log` and `setup.log`:

- Linux 6.18.44 x86_64, `nproc` 4, 15.7 GB
- go1.27.1 linux/amd64
- clang 20.1.8 (from `cloud/setup.sh`, `/opt/adamic-tools/llvm/bin/clang`)
- Node v24.22.1
- Submodules: `git -c protocol.file.allow=always submodule update --init --recursive --depth 1`, exit 0
- `bash cloud/setup.sh`, exit 0, then `source /opt/adamic-tools/env.sh`, `TMPDIR=/tmp/adamic-gate`

The scripts that ran everything are `run.sh` and `run2.sh`. Their stderr is in `driver.log` and `driver2.log`.

## Runs

### 1. The two named tests

`go test ./internal/lower -run '^(TestAssignedSymbolAnswersAsAFreshLowering|TestPredicateCallbackContracts)$' -count=1 -v`
(`run1-tests.log`): **exit 0**. Both passed, along with all eight `TestPredicateCallbackContracts` subtests,
`reassigned` among them.

### 2. Mutants

Each mutant was applied alone to `assignedSymbol` in `internal/lower/predicates.go` and run with the same command as
run 1, then the file was restored with `git checkout`. Every log ends with `git diff after restore: [0 bytes]`. The
exact patch for each mutant is in `mutant-<x>.diff`.

| Mutant | Exit | Red? | Failing test and line |
|---|---|---|---|
| a. walk reads `l.program.Files()[:1]` | 1 | red | `TestAssignedSymbolAnswersAsAFreshLowering`: `predicates_proof_test.go:321: assignedThere alone: assigned=false, want true` |
| b. `if l.assignedSymbols != nil { return false }` at the top | 1 | red | `TestAssignedSymbolAnswersAsAFreshLowering`: `predicates_proof_test.go:329: assignedThere from the memo: assigned=false, want true` |
| c. `return len(l.assignedSymbols) > 0` | 1 | red | `TestAssignedSymbolAnswersAsAFreshLowering`: `predicates_proof_test.go:321: readHere alone: assigned=true, want false` |

`TestPredicateCallbackContracts` passed under all three mutants, `reassigned` included.

**My first attempt at mutant a never applied the mutant.** The text
`for _, file := range l.program.Files() {` appears twice in predicates.go (lines 518 and 648). My script refused
to edit a string that occurs more than once (the AssertionError is in `driver.log`), so it tested the unmutated code,
which passed with exit 0. That log is kept as `mutant-a-NOT-APPLIED.log`, and it is not evidence about the mutant.
`run2.sh` then applied mutant a to line 648 alone, inside `assignedSymbol`. That diff is `mutant-a.diff`, and its
results are in `mutant-a.log` and the table above.

### 3. `go test ./internal/lower -count=1`, whole

`run3-lower-whole.log`: **exit 1**, 401.9 s. Five tests failed, each with the same message:

- `--- FAIL: TestProgramRegionCensusMembership`: `program_region_census_test.go:369: ADAMIC_TYPESCRIPT_SOURCE must name the pinned upstream TypeScript checkout; Program sources never fetch`
- `--- FAIL: TestProduct_ProgramRegionCensusPairwise`: same message, same line
- `--- FAIL: TestProduct_ProgramRegionCensusIndexed`: same message, same line
- `--- FAIL: TestCycleFinderCensusEquality`: `cycles_census_equality_test.go:280: ADAMIC_TYPESCRIPT_SOURCE must name the pinned upstream TypeScript checkout; Program sources never fetch`
- `--- FAIL: TestProduct_CycleScanningCensus`: same message as the line above, same line

At main `5f0b2152` I ran the same five tests with
`go test ./internal/lower -run '^(TestProgramRegionCensusMembership|TestProduct_ProgramRegionCensusPairwise|TestProduct_ProgramRegionCensusIndexed|TestCycleFinderCensusEquality|TestProduct_CycleScanningCensus)$' -count=1 -v`
(`run3-main-5f0b2152.log`): **exit 1**, and all five failed with the same messages at the same lines.

Two caveats. The whole-package run at the change was not run with `-v`, so its log names only the failures. Following
`CLAUDE.md`, a plain `go test` without the products built first is expected to fail on product consumers. Here every
failure is the missing `ADAMIC_TYPESCRIPT_SOURCE` input; none is a product refusal.

### 4. `go test ./stage3/census/latent/... -count=1`

`run4-latent.log`: **exit 0**. `importclean`, `refusalrewrite`, `statecopy` and `statementrewrite` all passed.

## What I infer (not observed directly)

- The five whole-package failures come from the box, not the change. Main fails the same five tests the same way, and
  the message names a missing gate input (a pinned upstream TypeScript 6.0.3 source checkout) that `cloud/setup.sh`
  did not install. This box never ran those five tests on their real input, so they say nothing about the change
  either way.
- Only `TestAssignedSymbolAnswersAsAFreshLowering` catches these three mutants. `TestPredicateCallbackContracts/reassigned`
  catches none of them. The new test's comment says a missed assignment "also reddens
  TestPredicateArgumentProofs/reassigned", but that test does not exist at this sha. A grep for
  `func TestPredicateArgumentProofs` finds nothing; the only match is `TestPredicateCallbackContracts`, and under
  mutant a it stayed green. The likely reason is that its program has a single file, which the first-file walk still
  reads. The comment's claim is wrong, or names a test that was renamed.
