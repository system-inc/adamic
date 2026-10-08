Fresh map callback returns lower without changing checker contextual types.
Commits: follows 28eeeb08 on codex/generic-function-value; nullable support merged at cc73ab53.
Checks: lower 15.787s, Node/native/JavaScript positive oracle; shared/mixed/stored Node observations and pinned refusals; vet and diff checks.
Mutant: proving only the first conditional return path loses the mixed callback refusal; TestFreshMapMixedReturnRefused fails with got <nil>.
Uncovered: named callbacks and local alias escape analysis remain conservative; full counts gate blocked by existing regexp_tree.ts project-root loading; no whole-repository gate.

Date: 2026-10-08 UTC.

The checker still infers never[][] at map and never[] at the callback return.
The view proof visits every inline callback exit, requires a directly returned
fresh array (a literal or an existing proven fresh allocation), and judges each
at the destination element type. It changes no checker type. Conditional paths
and return/if/block flow are supported. Stores, local aliases, unrecognized
control flow and shared returns retain the existing refusal. Nested shared
values inside inhabited literals still undergo invariance comparison.

The positive oracle instantiates the generic at number and an object with an
owned string field, pushes into each empty row, and also covers conditional
paths, block returns and slice allocations. Native sanitized/release builds and
the JavaScript backend match source Node. The negative source programs all
print 1 on Node, showing the shared array is mutated through the wider view;
they are refused before backend emission, with diagnostic text and site pinned.

Counts: 52 allocations, 52 frees, 33 retains, 65 releases, peak 16, regions 0.
Measured with the existing counted helper; only the new row was added, because
the full updater fails on the pre-existing internal/fresh/testdata/regexp_tree.ts
being outside its project's root files after the nullable dependency merge.

Logs in /tmp: fresh-map-{final-lower,final-refusals,final-backends,vet,counts,
counts-row,every-path-mutant,nested-control}.log. Setup from this unit:
30.366s total, clang .186s, build 30.106s, cache 30.340s; nproc 5.
The original checker-context investigation is preserved in
CONTEXTUAL_EMPTY_MAP_BLOCKER.md; the user's decision now resolves that choice.
