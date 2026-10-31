# The widened-read refusal: its mutants (#80cs7rz)

checkViewWidenedReads (view_widened.go) makes a read of a field a view checks, through a receiver type that doesn't
check it, NotYet when the receiver's value may come from a view. Four recorded outcomes move with it, ruled the price
of failing closed:

| Program | Was | Now | Why |
|---|---|---|---|
| stage3/interface-downcasts/lane1/objects-good.ts | lowered, counted | NotYet at 4:92, its count row and file dropped | alias is view.child, a viewed value a checked read took out |
| stage3/fixtures/predicates/03_void_zero.a | status.json `Compiles` | status.json `NotYet` at 23:12 | isNumericLiteral reads kind of node.expression, which a checked read took out |
| internal/oracle/testdata/review/agree/9xam9ds_predicate_closed_enum_tag.a | review/agree | review/refused, NotYet at 14:43 | found.value reads through ir.Defined, which the allocation graph doesn't follow: an unknown reach |
| internal/oracle/testdata/predicate_direction_uses/inferred_call.ts | TestPredicateDirectionUsesLower | NotYet at 46:66 | symbol.valueDeclaration.name reads through ir.Defined too: an unknown reach |

Each mutant was applied alone to view_widened.go in the change's tree and run on a Linux cloud worker (Cloud), then
restored and checked with `cmp`.

## Mutant 1: the pass returns nil at once

`go test ./internal/oracle -count=1 -run 'TestViewWidened|TestCheckedViewObjects|TestCheckedViewV2MovedStage3NotYet|TestReviewProgramsRefuse'`
and `go test ./stage3/fixtures -count=1 -run TestFixturesPredicates`. Red:

| Test | Failing |
|---|---|
| TestViewWidenedReadNotYet | all 23 live roads |
| TestViewWidenedPriceAccepted | 9xam9ds_predicate_closed_enum_tag.a, inferred_call.ts |
| TestCheckedViewObjects | objects-good |
| TestCheckedViewV2MovedStage3NotYet | 03_void_zero.a |
| TestReviewProgramsRefuse | 9xam9ds_predicate_closed_enum_tag.a (an accepted program in refused/) |
| TestFixturesPredicates | predicates/03_void_zero.a/stage0: `got {Outcome:Compiles}; recorded {Outcome:NotYet ...}` |

With `ADAMIC_VIEW_WIDENED_MUTANT=1 ADAMIC_GATE_UNCACHED=1`, TestViewWidenedReadNotYet holds the mutant's admitted
programs to Node and passes: each road disagrees with Node natively, release segfaulting and sanitized native dying in
AddressSanitizer (`SEGV on unknown address ... #0 in adamic_write_line /adamic-runtime/adamic.c:323:29` for
view_widened.ts).

## Mutant 2: entries reduced to CheckedCast alone

The structural cast and predicate local entries and the checked-read entry return no site
(`if false && entryLocals[value.Local]`, `if false && l.result.CheckedFields[...] && viewAggregate(value)`).
`go test ./internal/oracle -count=1 -run 'TestViewWidenedReadNotYet|TestCheckedViewObjects|TestCheckedViewV2MovedStage3NotYet'`.
Red: TestViewWidenedReadNotYet, at the predicate's local, parameter, closure, field and inline type roads, which
enter by narrowed reads alone.

The rest stay caught under it. The predicate's array road has an element whose reach is unknown. The values a checked
read takes out (view_widened_inner*.ts, objects-good.ts, 03_void_zero.a, so TestCheckedViewObjects and
TestCheckedViewV2MovedStage3NotYet stay green) are caught again without the checked-read entry: inner's and
objects-good's receivers are fields projected through a CheckedCast, which have no allocation behind them, so their
reach is unknown, and 03_void_zero.a also has a predicate view the allocation fallback weighs. No pin is decided by
the checked-read entry alone; it marks those values directly, where the projection would otherwise only fail closed.
