Built: refreshed landing and availability audit; no new claims or executable changes.
Commits: pushed wave-17 tip dc7f22fc remains on current main c01907a7 before this evidence-only update.
Commands and outputs: fresh fetch; 568 origin refs, 33 unique Markdown claim blobs, all 197 ranked entries checked; eight remaining candidates are React, zero non-React candidates.
Mutants: none rerun for metadata only; prior complete rule, sanitizer, fact and handle evidence remains in WAVE_17_LANDING3_REPORT.md.
Not covered: unclaimed React candidates, parked React integration and shared-harness adoption before it reaches main; no full gate rerun.

The landing base and executable sources have not changed since the last complete
validation. The remote wave-17 tip matched local HEAD before this evidence commit.
Main remains c01907a7036a22c2ea7ee686ed5fe4c6cd4bbc06. The named shared harness
ab70f38d4 was inspected on origin/lint-rules/harness; it is not an ancestor of
current main or this branch. Integration is merging that shared work into
area/stage1-lint. No shared files were copied, rewritten or reverted, and no
main or area branch is pushed here. Existing own listener declarations remain
available for that integration.

The refreshed ranking/claim/source audit is selection.json. All origin Markdown
claims are inspected, and existing source hits on main and codex/tsgo-c-library
are recorded. The only entries lacking both claims and source hits are:

- react/no-danger-with-children
- react/no-multi-comp
- react/no-namespace
- react/no-object-type-as-default-prop
- react/no-unstable-nested-components
- react/sort-default-props
- react/static-property-placement
- react/style-prop-object

No unclaimed non-React rule remains. This is not a claim that every rule in the
ranking is ported or claimed: the eight names above remain unclaimed. Under the
current React parking instruction this unit stops without adding claims. The
existing React claims retain their exact recorded blockers, not a newly invented
high-level IR dependency for every React rule. In particular, this audit does
not infer that the eight unclaimed rules all need capture or single-assignment
analysis merely because they are React rules.

No code was changed or new correctness check added. Prior byte-oracle and mutant
observations are therefore referenced, not represented as new executions.
