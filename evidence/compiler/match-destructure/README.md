# compiler/match-destructure (#jy9k7a3)

Code: branch `compiler/match-destructure`, one commit on base 82b0c076395636800764ed1af7ada569681fc2ae
(origin/main when I started; main has since moved to a7b6da7f, and `git merge-tree` says the commit merges clean).
Head: 000338402943919fcd4a8d426372d91613bfaa49.

## What was built
`internal/lower/match_destructure.go` (new, 61 lines) plus a 3-line hook in `destructureFrom` (collections.go). An
array binding pattern over an `ir.Array` whose checker type is RegExpExecArray or RegExpMatchArray lowers to a held
local plus, per name, `ArrayIndex{held, i, String}` (the same node `match[i]` lowers to), wrapped in `Coalesce`
(lazy fallback) when the name has a default. Holes skip their index. Rest elements and nested patterns say NotYet in
new words; plain `T[]` is NOT covered (still NotYet "destructuring anything but a tuple into [names]"), because
the element read for other element types is not the match read and I did not prove it. lower.go and emit.go untouched.
Absent captures need nothing new: a String slot holds NULL for undefined (read in the emitted C for `match[2]`).

## Proof
- Fixture `internal/oracle/testdata/match_destructure.a` (+ `internal/oracle/match_destructure_test.go`, counts row
  `| 135 | 135 | 204 | 207 | 30 | 0 |`, allocations = frees). Covers semver.ts's two lines with their regexes, absent
  optional groups, holes at the start and middle, a pattern longer than the match, defaults that log (used and
  unused, in order), a default using an earlier name, `String.prototype.match`, `let`, and the `if (!match)` narrowing.
  `oracle_fixture_green.log`: TestNativeAgreesWithNode_028 passed this fixture (Node, native under ASan/UBSan with the leak
  check, JavaScript backend). `node_output.txt` is Node's stdout.
- `internal/lower/match_destructure_test.go`: names/hole/default lowers; rest, nested pattern and plain array stay NotYet.
- Mutants, both turned `TestNativeAgreesWithNode_028/internal/oracle/testdata/match_destructure.a` red (stdout differs on native
  and on the JavaScript backend), then restored (diff against the good file empty):
  (a) default evaluated always (an `Evaluate{fallback}` before the Coalesce): `mutant_a_default_always_evaluated.log`,
      extra "default a / default b / default d" lines.
  (b) hole not skipped (index advances only for a non-hole): `mutant_b_hole_not_skipped.log`.
- gofmt -l cmd internal: empty. `go vet ./internal/lower/`: clean.

## Box and what is NOT as the repo asks
Linux, 4 cores, Go 1.27.1, **clang 18.1.3 (not the house 20.1.8)**, Node 24.22.2 for the package runs (Node 22.23.4 for
the fixture and mutant runs above), libclang-rt-18-dev installed for ASan. The count row was recorded with clang 18; if
20.1.8 counts differently, regenerate it there. I did not run Loom or the products-first gate; I ran `go test -count=1`
per package with ADAMIC_TYPESCRIPT_SOURCE set to the submodule (its checkout lacks pinned commit 050880ce).

## Whole packages, head vs base 82b0c076 (logs: head_*.log, base_*.log; failure lists head_failures_*, base_failures_*)
Package lines:
- lower:      head FAIL 639.7s, base FAIL 711.6s
- javascript: head ok 0.27s, base ok
- native:     head FAIL 874.9s, base FAIL 605.5s
- oracle:     head FAIL 2483.7s, base FAIL 1776.0s
Failures only at head: **none** in any package (`compare_*.txt`). Every head failure is also a base failure:
lower 5 (program-source census tests: pinned TypeScript commit absent), native (clang 18 links libLLVM, so the object/unit
cache tests refuse; debug-review and TSGo product tests), oracle 52 (Product, counts shards, Stage0Emits shards,
AssignmentProof; none names match_destructure). Base had more failures than head (flaky or product-dependent tests I
did not chase), so these are environmental, not caused by this change. No failure at head names this change's files.
