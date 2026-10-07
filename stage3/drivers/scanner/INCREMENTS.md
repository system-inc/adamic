# Value-used scanner increments

Value-used numeric `++` and `--` now lower for locals, captured variables,
globals, ordinary numeric fields (including private fields), and narrowed
numeric array elements. Prefix returns the new number; postfix returns the old
number. Both supplied scanner probes are unchanged from `dc9f8482`.
Base: origin/main `48c05d091f0a43c31cbe051b1d6578d99eeedf19`.

Lowering makes ordinary IR helpers: save the old number, compute and store the
new number, return the selected number. Receiver and index are arguments, each
passed once and evaluated in order. A local uses a temporary closure over its
capture cell; writes remain ordinary Assign nodes visible to the analyses.
No new backend operation or garbage collector is introduced. This favors
correctness and existing ownership proofs over eliminating helper calls and
closure allocations.

## Observations against Node

Before the change, an overlay of main's unchanged expression.go reproduces:

```text
postfix-call-argument.a:4:8: stage 0 can't lower a PostfixUnaryExpression yet
prefix-call-argument.a:4:8: stage 0 can't lower a PrefixUnaryExpression on a number yet
```

| Unchanged probe | Source Node stdout | Native stdout | Backend Node stdout |
|---|---|---|---|
| postfix-call-argument.a | `0\n1\n` | `0\n1\n` | `0\n1\n` |
| prefix-call-argument.a | `1\n1\n` | `1\n1\n` | `1\n1\n` |

All six executions exit 0 with empty stderr. Source and generated JavaScript
use oracle/node.mjs to load the repository runtime. Native was built with
`--sanitize`; the differential oracle additionally checks release native and
leaks. Directly invoking generated JavaScript without that loader initially
failed to resolve the unpublished adamic runtime; the correctly loaded rerun
above passed.

`increment_values.a` exercises all four forms on each target category, sibling
arguments, arithmetic, conditional branches, repeated closure calls, a loop
condition, private fields, signed zero, large IEEE numbers, Infinity and NaN.
The effectful field receiver prints four `receiver` lines for four updates.
The array getter prints five `array target` lines: one narrowing read and four
updates. The index is an existing constant index. A dynamic getter index did
not retain TypeScript's required narrowing and is not claimed as a source
fixture here. Lowered receiver and index arguments each occur once.

| Mutant actually run | Catcher |
|---|---|
| Postfix returns new | Source Node versus native stdout and backend stdout |
| Field receiver evaluated twice | Extra `receiver` lines, both backend comparisons |
| Array receiver evaluated twice | Extra `array target` lines, both backend comparisons |

All three mutant tests exit 1 after executable output comparisons; no lowering
refusal, compiler warning or sanitizer error is counted as the catch. Each
mutation was restored. Reproduce after sourcing the setup environment:

```sh
python3 stage3/drivers/scanner/probes/run-increment-mutants.py > /tmp/scanner-increments-mutants.log 2>&1
```

Individual logs: `/tmp/scanner-increments-mutants/<mutant>.log`.

## Prefix-site coverage and limits

One scanner prefix increment is independently identified and proven by its
unchanged reduction: `++pos`, upstream scanner.ts:3819:138 at TypeScript 6.0.3
`050880ce59e30b356b686bd3144efe24f875ebc8`. Relative to the requested 88-site
meter total, **1/88 is demonstrated**, not a claim that the other 87 are
increments or that their surrounding programs now lower. A stock-TypeScript
parse of that pinned raw scanner has 94 prefix expressions: 77 `!`, 9 unary
`-`, 6 unary `+`, 1 `~`, and 1 `++`. This change adds the increment family;
other unary operators retain their prior behavior.

The exact meter delta is unmeasured. Main's latest checked-in paired report
(20261007T062959Z.tb10Z0) has 86 `a PrefixUnaryExpression on a value` sites and
8 `on a number` sites. It does not retain their raw site list, and neither that
aggregate nor the raw scanner AST reconciles the supplied 88-site denominator.
The number above is confirmed scanner coverage, not extrapolation to the whole
compiler census. The meter artifact was requested while independent work
continued.

Accessor targets, optional-number local/field storage, tuple element storage,
and non-number slots remain NotYet. The work does not establish full scanner
native execution, close its other blockers, optimize helper allocation, or
prove effectful dynamic index evaluation with a checker-clean source fixture.
No code was copied from cohere, and the four excluded compiler files are
unchanged. Conservative assumption: this unit admits proven number slots only.

## Validation and setup

Final setup passed: Go ready 0.062s, Node ready 0.052s, submodules ready 0.141s,
markdown dependencies ready 0.147s, clang ready 0.325s, Go build ready 65.819s,
cache warm 66.073s, done 66.247s. `nproc` is 5, CPU quota is 4.
Environment: `/workspace/adamic-tools/env.sh`; Go 1.27.1, clang 20.1.8,
Node 24.19.0. GOPROXY was set to `https://proxy.golang.org|direct` before setup.
Initial setup passed in 85.289s on the original checkout. The first fetch
stalled in a nested submodule fetch; fetching without recursion and then
updating the recorded submodules aligned main. A main setup retry failed
exactly with `l.incrementValue undefined`: Go enumerated package files before
the new file existed while expression.go changed during the build. Rerunning
with all files present produced the final successful setup above.

The first filtered oracle invocation matched no fixtures and was discarded.
The corrected uncached focused run passed in 54.681s, including both supplied
probes, increment_values.a and the two existing cyclic scanner controls.
Counts regeneration passed in 75.653s and changed only the three new rows:
postfix 4/4/0/4/2/0; prefix 4/4/0/4/2/0; increment_values 118/118/32/135/8/0
(allocations/frees/retains/releases/peak/regions). All existing rows are unchanged.
Go vet passed with no output; formatting and diff whitespace checks passed.

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/(internal|stage3)/.*/.*/(increment_values.a|probes)' -count=1 -timeout 10m -v > /tmp/scanner-increments-focus.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts > /tmp/scanner-increments-counts.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./internal/lower ./internal/oracle > /tmp/scanner-increments-packages.log 2>&1
go vet ./... > /tmp/scanner-increments-vet.log 2>&1
```

Both complete touched packages passed uncached: internal/lower 52.442s and
internal/oracle 288.847s. The whole repository gate was not run. All test output
went to log files. Final fetch confirmed origin/main is still 48c05d09, already
an ancestor of this branch; no integration merge was needed.
