# Adapted-tree census, October 7, 2026

Measured the complete output of `bash stage3/apply.sh /tmp/lazy-stage3-adapted`
from origin/area/stage3 `234ab1aa5f728a5221fb6075c35b94f88a2c6437`.
Upstream is TypeScript v6.0.3, `050880ce59e30b356b686bd3144efe24f875ebc8`.
All registered adaptations ran, including 10, 20 and 75; the actual patch table
records 78 files, 5,122 added lines and 5,096 removed lines. No slice branch
or unmodified TypeScript tree was used. Installed the tree's locked dependencies
with `npm ci --ignore-scripts --no-audit --no-fund` (338 packages, 7 seconds).
The isolated installation-free trial had one extra TS2307, which is excluded
from the final measurement. Installation changed six displayed type strings,
but no cast span or classification.

Compiler under measurement includes own lazy admission and optional Error fix,
merged designated integration `4d86366105a10b39b41f466a2a3cbeafa7c5689b` in
`b8bd8392`, then verified at `df73329d`. No other lane was merged directly.

| Measure | Tagged | Untagged |
|---|---:|---:|
| Adapted interface-downcast sites | 1,758 | 1,178 |
| Lazy descriptors admitted at exact adapted spans | 1,758 | 1,178 |
| Production entry compilation | 0 | 0 |

There are 4,129 `as` casts in the non-generated compiler sources, including
1,193 outside these two selected categories. Inventory uses the pinned stock
checker and compiler tsconfig to identify interface downcasts, common finite
narrowing fields and declared ancestry. Admission is not compilation.
Production measurements use unchanged Adamic checker options and load each of
49 cast-owning files with its actual transitive imports. Every entry fails in
the checker with the same 259 diagnostic rows; the whole compiler measurement
also has these 259 rows. Thus none reaches lowering or either emitter.
No checker diagnostics were bypassed and no partial executable IR was exposed.

## Remaining diagnostic rows for the TypeScript owner

[adapted-diagnostics.csv](census/adapted-diagnostics.csv) contains every remaining
row, with exact relative file, line, column, diagnostic code and full multiline
message. These are Adamic's production checker diagnostics, not the stock
checker's diagnostic count or adaptation 75's widening-relation count.
TS1484 is absent; 25 TS2412 remain. This is the table to split into further
TypeScript source adaptations and language gaps; the census does not guess that
classification from an error code.

| Diagnostic | Rows |
|---|---:|
| TS2345 | 87 |
| TS18048 | 49 |
| TS2322 | 30 |
| TS2412 | 25 |
| TS2375 | 16 |
| TS2532 | 15 |
| TS2379 | 10 |
| TS2488 | 6 |
| TS2769 | 6 |
| TS18046 | 6 |
| TS2339 | 3 |
| TS2538 | 2 |
| TS2722 | 1 |
| TS2556 | 1 |
| TS2420 | 1 |
| TS2740 | 1 |


Remaining checked-view blockers per runtime member family and exact allocation
reachability are **unmeasured**, not zero: checker-rejected entries have no IR
or shared closed-world allocation-flow graph. Even though all descriptor targets
are admitted lazily, that does not establish which reads the real program reaches.
No syntactic read table is substituted for that proof. Rerun after these checker
rows are resolved to obtain executable shared-flow read-demand witnesses.

## Reproduction and artifacts

Set up the area/stage3 detached checkout and run its apply script, then install
the resulting tree's locked dependencies. In this branch:

```sh
NODE_PATH=/path/to/pinned/stage3/api/node_modules node stage3/interface-downcasts/lazy/adapted-casts.cjs /tmp/lazy-stage3-adapted /tmp/adapted-sites.json > /tmp/inventory.log 2>&1
LAZY_ADAPTED_ROOT=/tmp/lazy-stage3-adapted LAZY_ADAPTED_SITES=/tmp/adapted-sites.json LAZY_ADAPTED_OUTPUT=/tmp/production.json go test ./internal/lower -run '^TestLazyViewAdaptedCensus$' -v -count=1 -timeout 15m > /tmp/production.log 2>&1
```

The stock API package must be TypeScript 6.0.3. The census test passed in
125.659s. The separate exact-span descriptor audit passed in 5.131s. Compressed
cast and production ledgers retain each site and each entry's diagnostic rows.
The source-hash ledger covers all 79 compiler .ts inputs, including generated
inputs; the cast inventory excludes generated files. Patch table and apply,
installation, inventory, descriptor and production logs are alongside the ledgers.

Integrated uncached lazy/nullish oracles passed in 19.692s, including sanitized
native, release native, JavaScript and Node comparisons, plus leak checks.
Focused shared-flow lower checks passed in 0.013s. Full JavaScript package passed
in 0.925s. Full lower package failed in 43.994s in the existing overload,
predicate-marker, phantom-array and nested-function assertions; full output is
in `census/packages-integrated.log`. No fully green repository gate is claimed.
The optional Error producer fix and number/null payload mutants are described
in [REPORT.md](REPORT.md); the null-as-undefined implementation mutant was caught
and restored. Arbitrary nullable unions and a new full native gate were not covered.
