Built: merged integration ba59427 and replaced the old queue with the lazy candidate inventory.
Commits: integration ba59427; this measurement checkpoint on codex/views-mixed-unions.
Checks: supplied gzip inventory parses; exact family counts assert 14/511 and 23/538.
Mutants: no compiler check is claimed by this measurement checkpoint; read fixtures are in progress.
Uncovered: full __String integration, mixed primitive source selection and production tsc reachability.

The integration REPORT.md was read, including each hunk choice and inherited gate
limits. Its tip contains our 9a385606 and merged without conflict. Only the
integration branch is consumed for other-lane dependencies.

The queue now comes from lazy/census/read-demand-pairs.json.gz, with its exact
compressed source hash preserved in lazy-pair-progress.json. This supersedes the
old lane-4 213/1111 ledger for current scheduling. The supplied lazy REPORT calls
these static candidate reads and explicitly leaves production allocation-flow
reachability unmeasured. The table must not be presented as certified runtime
reachability. Whole-tsc checker errors do not prevent individual source fixtures.

| Family | Remaining pairs | Remaining candidate reads |
| --- | ---: | ---: |
| __String | 14 | 511 |
| Mixed primitive union | 23 | 538 |

Columns overlap: all fourteen branded pairs occur in the mixed-primitive family.
The leading Identifier.escapedText and Symbol.escapedName pairs account for
224 + 223 = 447 candidate reads in this inventory. Completion still requires
source lowering on both backends, Node controls, wrong-value pins and semantic
mutants; the previously pushed plain intersection fixtures remove no full union.

Revised working targets: __String October 8, 2026, 17:00 MDT (23:00 UTC), and
both families October 10, 2026, 17:00 MDT (23:00 UTC). The remaining intersection
receiver and callable/array consumer entries may need other-lane hooks; their
counts will remain pending until source fixtures establish completion. No date
is contingent on making the whole tsc program checker-clean. Report actual
integration blockers rather than silently removing those entries.

The in-progress implementation uses the existing registry's Undefined bit to
separate a required string field containing undefined from optional absence.
It uses the shared readiness/optional slot reader, with no second flow graph or
readiness bitmap. Source evidence and revised remaining totals follow in the
first green compiler group; no in-progress code is included in this table push.
