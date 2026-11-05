# A Program region with generators: its mutants (#8xcz03h)

lower.go no longer refuses a Program region in a program with generator functions. Three invariants make that safe,
each pinned (lower.go names them where the guard stood):

| Invariant | Pin |
|---|---|
| A generator's frame has no captured cell | TestProgramRegionGeneratorCapturedLocalStops, TestProgramRegionGeneratorNestedFunctionStops |
| No generator frame is adopted into the region | internal/oracle's TestProgramRegionGeneratorsNotAdopted |
| No generator value reaches a slot | TestProgramRegionGeneratorStoredInFieldStops, TestProgramRegionGeneratorPushedIntoArrayStops, TestProgramRegionGeneratorHeldInLocalStops, and on the callee side of an Iterable<T> parameter, TestProgramRegionGeneratorParameterStoredInFieldStops, TestProgramRegionGeneratorParameterCapturedByClosureStops |

TestProgramRegionGeneratorProgramEngages is the control: the shared tree and generator lower with the region on and
its Tree literals in it, so each pin stops on its own road.

The frozen scanning reference (cycles_scanning_reference_test.go) keeps the guard, so cycleCompareProgramModes skips the
ProgramRegion=true comparison where the reference's error is exactly that guard's NotYet (cycleFrozenGeneratorGuard),
and logs why. ProgramRegion=false still compares in full.

Each mutant was applied alone to the change's tree (32dbff6f21) in the box's proof checkout on Golem, by a script
that saved the file first, then restored it and checked it with `cmp`.

## Mutant 1: the guard restored in lower.go

The why-comment replaced by the old `if lowering.result.HasGenerators()` returning the NotYet. Then
`go test ./internal/oracle -count=1 -run '^TestProgramRegionGeneratorsNotAdopted$' -timeout 30m` and
`go test ./internal/lower -count=1 -run '^TestProgramRegionGenerator' -v`. Red:

| Test | Failing |
|---|---|
| TestProgramRegionGeneratorsNotAdopted | program_region_generators.a:7:1: stage 0 can't lower a Program region in a program with generator functions (#qk8rztp) yet, at load |
| TestProgramRegionGeneratorProgramEngages | main.a:1:1, the same NotYet |

The seven stop pins stay green under it: each road stops while the program is lowered, before the guard's place at
its end, so none of them is decided by the guard.

## Mutant 2: the comparison's skip removed

The skip's condition made `if false && region && ...` in cycleCompareProgramModes. Then
`go test ./internal/lower -count=1 -run '^TestCycleFinderEquality_(008|009|013|023|024|026|039|040|042|046|047|049|058|061|063|067|071|072|073|080|083|084|086|090|091|099|104|106|107|108|112|121|123|125)$' -timeout 30m`,
the 34 shards holding a program with `function*` (cycleComparisonShard's FNV hash). Red: 29 shards, 008 009 013 023
024 026 039 042 046 047 049 058 061 067 071 073 083 084 086 090 091 099 106 107 108 112 121 123 125, each at the
first generator program it reaches with ProgramRegion=true, the scanning reference's NotYet against the real
pipeline's lowering, for example split.a:

    scanning: *lower.NotYet: .../generators/split.a:3:1: stage 0 can't lower a Program region in a program with generator functions (#qk8rztp) yet (cycle=false)
    indexed:  (cycle=false)

Shards 040, 063, 072, 080 and 104 stay green: their programs with `function*` (a regex corpus whose text only
names it, internal/conformance's async generator.a and async_generator.a, stage3's 25_map_generator.a,
void_widening_refused/shadowed_array_from.a) end the same way in both pipelines without reaching the guard's place.

Unmutated, the same 34 shards pass, logging 37 programs whose ProgramRegion=true comparison is skipped.

## Unmutated, at 32dbff6f21 on Golem

`gofmt -l` on the touched Go files is empty, `go vet ./internal/lower ./internal/oracle` is clean,
`go test ./internal/oracle -count=1 -run 'ProgramRegion' -timeout 30m` passes (72 s), and
`go test ./internal/lower -count=1 -timeout 40m` passes (515 s).

## Re-cut onto main 28897e4da0, on Pupa

The four commits cherry-picked onto main at 28897e4da0, run at e117950dee on Pupa. Main's new fixtures make 44
programs with `function*` across 35 shards, shard 030 joining the 34 above. `gofmt -l` is empty, `go vet` is clean,
every pin passes with its message unchanged, `go test ./internal/oracle -count=1 -run 'ProgramRegion' -timeout 30m`
passes (112 s), the 35 shards pass unmutated, logging 39 skipped ProgramRegion=true comparisons, and
`go test ./internal/lower -count=1 -timeout 40m` passes (522 s). The mutants weren't rerun, since no pin's message
and not the skip changed. After the pin of a generator held in a local was renamed (afa1553364), the same runs passed
again on Golem (oracle 103 s, the 35 shards with 39 skipped, internal/lower 455 s).
