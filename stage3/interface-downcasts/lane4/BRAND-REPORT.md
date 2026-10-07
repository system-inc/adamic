Built: merged phantom brands and integrated primitive branded scalar field reads, preserving literals.
Commits: phantom dependency merge dfdd2467; implementation commit accompanies this report.
Checks: branded source oracle passes 1.894s; phantom/mixed lower tests pass 2.501s; checked-view oracle passes 13.859s; vet and diff checks pass.
Mutants: four backend executions bypass the helper read with a valid unchecked string; pinned refusals catch each exit-0 result.
Uncovered: tsc __String branded-void/internal-symbol union, explicit undefined, optional members, mixed primitive source selectors and lane 1 merge.

The phantom-brand dependency at d90994da merged cleanly. The shared data walker
now recognizes only phantomBase-approved primitives, and the contract builder
records the primitive rather than inventing a brand object. Literal brands keep
the base's literal constraints in both the contract and field-read metadata.
Four source .a fixtures exercise a helper receiving a checked view: valid string,
wrong number, valid branded literal and wrong same-kind literal. Source Node
prints the actual payload, while both backends pin exit 70, expression,
declared __String and found category/value on wrong reads. Good native executions
also run under ASan/UBSan. The mutation replaces just the helper's checked read
with an unchecked valid string, so both release backends execute valid code and
the refusal pin catches the missing check. An earlier attempt merely dropped
Property.View on a numeric slot and caused a native signal; it was discarded,
not counted as a semantic mutant kill.

The frozen tsc declaration at types.ts:6196 is a union of branded string,
branded void and InternalSymbolName. These primitive-intersection fixtures do
not cover that complete declaration. Therefore this push removes zero assigned
tsc pairs. Remaining by the user's split: __String 30 pairs / 543 reads;
mixed primitive unions 63 pairs / 690 reads. Family columns overlap.

Lane 1 ab4d6f90 was explicitly fetched. Its merge conflicts in the plan, cast.go,
expression.go, interface_cast.go, readiness.go, view_objects.go and native
view_fields.go. The automatic approval reviewer rejected a bulk resolution
because selecting or concatenating conflicting compiler branches without semantic
validation risks a silent miscompile. That proposed action was abandoned and the
merge was aborted. No conflict markers or partial merge remain. The conflict log
is preserved for individual reconciliation with lane 1; its lazy admission is
not merged or claimed complete here.

Targets remain October 9, 2026, 17:00 MDT for __String and October 13 at the same
time for both assigned families. The clean phantom dependency merge alone does
not justify moving the __String date earlier while its void/undefined read path
and shared dispatch reconciliation remain unverified.

Commands (test output is logged):
- go test ./internal/oracle -run '^TestCheckedViewBrandString$' -count=1 -v -timeout 10m
- go test ./internal/lower -run 'TestPhantom|TestMixedUnionContract' -count=1 -timeout 10m
- go test ./internal/oracle -run '^TestCheckedView' -count=1 -timeout 15m
- go vet ./internal/lower ./internal/oracle
- git diff --check

The wider checked-view run preceded addition of literal-brand fixtures; the
subsequent focused branded source run covers those fixtures. The full repository
gate was not run. Fixture paths are lane-owned and outside the counted native
fixture registry; no count row or completed-pair entry is claimed.
