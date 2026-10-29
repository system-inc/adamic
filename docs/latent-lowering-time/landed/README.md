# The four changes landed: before and after on main

Asked: three runs back to back on one box, each built as the meter builds it plus the timing lines, over the same
adapted tree: main `95a7b63f` (before), main `51981922` (after, with d7dbbb13, b4eb424b, 609383f5 and 51981922),
and `51981922` with `07b24606` (#3hxh5jm) reverted.

## The meter's overlay refuses main 51981922 (observed)

`make_overlay.py` on `51981922` stops before any build ([overlay-refusal-51981922.log](overlay-refusal-51981922.log)):

```
panic: latent statement rewrite: lowering.statement: unsupported signature func(node *ast.Node) (result []ir.Statement, err error)
```

`statementrewrite` (stage3/census/latent/statementrewrite/main.go:44) accepts `lowering.statement` only as
`func(node *ast.Node) ([]ir.Statement, error)`. `f2cf1ebb` ("Map debug builds to source lines with budgeted
cache-safe skip-free tests", 2026-10-28 23:36 UTC) gave it named results, `(result []ir.Statement, err error)`, and
left `statementrewrite` as it was. Nothing under `stage3/census/latent/statementrewrite` changed since `95a7b63f`.
origin/main is `51981922` as of this push, so the daily meter's per-ref `build_census` (its `make_overlay.py`
line, under `set -e`) will stop at `latent-overlay` on this main. The same refusal hits the revert tree.
`95a7b63f` builds.

## The trees (observed)

At Performance's word, the fix is Correctness's `55c1c38f` (typescript/statementrewrite-named-results,
chg_0wdzk9qm: "statementrewrite takes lowering.statement by its types, not its names"), cherry-picked as a
throwaway onto both 51981922 trees, never pushed. A patch of my own, written first, was dropped unused.
`95a7b63f` needs nothing.

| label | tree | how |
|---|---|---|
| before | `95a7b63f` | main |
| after | `be99a01a` | `51981922` + `55c1c38f` cherry-picked (`-x`, clean) |
| revert | `589f337c` | `51981922` + `git revert --no-edit 07b24606` (`93245360`, clean, no fallback needed) + `55c1c38f` cherry-picked |

All three pin cohere `06af2e39` and TypeScript `d92d9bfe`, and each built as the meter builds it plus the timing
lines. `go test ./stage3/census/latent/statementrewrite/` is green on the after tree. The revert of `07b24606`
removes its lowering changes and its tests and fixtures (62 files, 593 lines). The three ran back to back in that order.

## The three runs (observed)

| run | wall | user s | peak RSS | loop sum wall s (user s) |
|---|--:|--:|--:|--:|
| before `95a7b63f` | 1:02:23 (3,743 s) | 5,082.0 | 4.56 GB | 3,737.0 (5,078.8) |
| after `51981922` (+ `55c1c38f`) | 13:00.4 (780.4 s) | 845.5 | 4.14 GB | 776.5 (841.6) |
| `51981922` without `07b24606` (+ `55c1c38f`) | 12:56.3 (776.3 s) | 842.4 | 4.01 GB | 773.1 (839.4) |
| after / before | 0.21 | 0.17 | 0.91 | |
| without / with `07b24606` | 0.99 | 1.00 | 0.97 | |

Exit 0 for all three. Each run's `time -v`, load at start and end, per-file lines and decision lines are under
[before/](before), [after/](after) and [revert/](revert); the three side by side per file in
[per-file-three-runs.txt](per-file-three-runs.txt). The before run started at load `1.22 3.00 1.99`, just after
the builds.

What the decision lines hold: each stderr line is `file`, the number of units and the number of findings
(`main.go.txt` prints `record.File, len(record.Units), len(record.Findings)`), not NotYet and refused. The
counts by kind below come from `latent.jsonl` ([findings-by-kind.txt](findings-by-kind.txt)).

### Before against after: main's lowering changed between them

`cmp`: `latent.jsonl` differs (first at line 5). Before 20,131,373 bytes, sha256 `395d5157...`, the same as every
earlier run of `95a7b63f` and of the four heads; after 27,453,371 bytes, sha256 `4a942058...`.

**57 of 79 decision lines differ** ([before-vs-after-decisions.diff](before-vs-after-decisions.diff)). In all 57 only the
findings count moves; units are the same in every file (4,543 in all), as are unit statuses (4,017 attempted,
498 deferred generic instantiations, 28 skipped checker bodies). Findings rise in 55 files and fall in 2:

| run | Refused | NotYet | skipped dependency | findings |
|---|--:|--:|--:|--:|
| before | 2,609 | 1,328 | 3 | 3,940 |
| after | 12,203 | 1,198 | 1 | 13,402 |
| without `07b24606` | 12,202 | 1,249 | 1 | 13,452 |

Nearly all of the rise is one new refusal family, "a read of a void container's elements", by role: as an argument
+5,620, as a value +897, as a returned value +880, as an assertion's operand +562, as an operand of `&&` +343, and
smaller roles ([before-vs-after-reasons.txt](before-vs-after-reasons.txt) has every reason that changed). Among
the NotYets, "a checked field alias requiring an optional, accessor, or representation conversion" (340) became
"a checked field alias of an optional field or through ?." (334), and "an indirect call of a checked predicate
overload" (38) and "a checked predicate without a reifiable target" (24) are gone.

The same files in the 79-file program cost 0.21 of the before's wall: `utilities.ts` 795.2 to 194.6 s, `types.ts`
452.2 to 48.0, `visitorPublic.ts` 242.4 to 2.9, `factory/nodeTests.ts` 154.2 to 15.2, even while the after finds
3.4 times the findings.

### After against `07b24606` reverted: the price of #3hxh5jm

`cmp`: `latent.jsonl` differs (first at line 5). After sha256 `4a942058...`, 27,453,371 bytes; without `07b24606`
sha256 `72b975ac...`, 27,486,038 bytes.

**17 of 79 decision lines differ** ([after-vs-revert.txt](after-vs-revert.txt)), each in the findings count alone,
and no file's units or declarations differ ([after-vs-revert-findings.txt](after-vs-revert-findings.txt), finding by
finding). Without `07b24606`: +38 "an indirect call of a checked predicate overload" and +13 "an indirect call of a
checked predicate" (NotYet), and one refusal fewer, "an unproven predicate argument for parameter test (argument
\"createAssignmentCallback ? isIdentifier : isExpression\")" in `transformers/destructuring.ts`. So
#3hxh5jm turns 51 NotYets on qualified predicate calls into lowered calls, and refuses one.

Its time: 13:00.4 with it against 12:56.3 without (1.005, user 845.5 against 842.4 s), within this box's noise.
Per file the largest moves are `factory/nodeTests.ts` (15.2 with, 13.7 without) and `factory/utilities.ts`
(16.5 against 15.2), each a second or two. #3hxh5jm costs nothing measurable in the meter.
