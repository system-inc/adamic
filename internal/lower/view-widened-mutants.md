# Widened reads checked where they're made: their mutants (#zpzm9ta head 1)

checkViewWidenedReads (view_widened.go) finds a read of a field a view checks, through a receiver type that doesn't
check it, whose receiver's value may come from a view. It marks the read in Program.WidenedReads under WidenedReadKey
(its CheckedFieldKey and its place), and Program.CheckedRead answers yes for a marked read, so readiness keeps the read's
view and the read is checked: a lie stops at exit 70 with "field read failed", on native, sanitized native and the
JavaScript backend alike. A method read and a getter-named read stay NotYet. Before this, #80cs7rz made every such read
NotYet, and four recorded outcomes paid for it. They come back:

| Program | Under #80cs7rz | Now |
|---|---|---|
| stage3/interface-downcasts/lane1/objects-good.ts | NotYet at 4:92, its count row dropped | lowered, its count row back at 5\|5\|5\|11\|4\|0 |
| stage3/fixtures/predicates/03_void_zero.a | status.json `NotYet` at 23:12 | status.json `Compiles` |
| internal/oracle/testdata/review/agree/9xam9ds_predicate_closed_enum_tag.a | review/refused, NotYet at 14:43 | review/agree |
| internal/oracle/testdata/predicate_direction_uses/inferred_call.ts | NotYet at 46:66 | lowers, in TestPredicateDirectionUsesLower |

Each mutant was applied alone in the change's tree at 4486aa3a (main cee37397) on a sealed Loom job, its tests run,
then restored and checked with `cmp`. The control, the head as built, was green first in the same job. Each ran twice,
once with the gate cache and once with ADAMIC_GATE_UNCACHED=1 (J-sm5a1pq, J-ydaazgw), and the tables count what went
red in both.

M1 and M2 ran `go test ./internal/oracle -run 'TestViewWidened|TestIntersectionViewWidened|TestCheckedViewObjects|TestReviewProgramsRefuse|TestReviewProgramsAgree|TestHonestWidened|TestSourceSlotCertificate'`,
`go test ./internal/lower -run 'WidenedReadIsChecked|KindReadIsChecked|KindUnionCastRecursiveReadIsChecked|ClosedRemainderWithNodeFieldsIsChecked|ViewWidenedReadKeyCollision'`
and `go test ./stage3/fixtures -run TestFixturesPredicates`.

## M1: CheckedRead stops answering for WidenedReads

The three lines at the top of Program.CheckedRead (internal/ir/views.go) are removed, so readiness strips each marked
read's view and the read is plain. Red in both runs:

| Test | Failing |
|---|---|
| TestViewWidenedReadIsChecked | 21 roads |
| TestBrandedSupertypeWidenedReadIsChecked | 18 roads |
| TestIntersectionViewWidenedReadIsChecked | all 14 roads |
| TestOverloadResultViewWidenedReadIsChecked | 4 |
| TestBrandedSupertypeKindReadIsChecked | 3 |
| TestKindUnionCastRecursiveReadIsChecked, TestClosedRemainderWithNodeFieldsIsChecked, TestViewWidenedReadKeyCollision | each |

## M2: checkViewWidenedReads returns nil at once

No read is weighed or marked. Red in both runs: the same tests, TestViewWidenedReadIsChecked at 18 roads, and
TestViewWidenedRoadsHeld at view_widened_held_getter.ts, whose getter-named read the pass alone keeps NotYet.

## What varies, and what never goes red

A few rows went red in one run and not the other: under M1, TestViewWidenedReadIsChecked's view_widened_closure.ts
and view_widened_map.ts; under M2, its view_widened_conditional.ts, view_widened_field.ts, view_widened_map.ts and
view_widened_parameter.ts; under each, TestSourceSlotCertificate's road_destructure_assignment.ts. A row that stays green
under a mutant still stopped at the checked read on every backend, so its read was checked by another road there. The
compiler's output didn't move: `adamic c` gave one output per program in five runs, at the head
and under M2 (J-m624p37). Every widened road among them has its intersection twin, and TestIntersectionViewWidenedReadIsChecked
went red at all 14 under both mutants in both runs. road_destructure_assignment.ts has no twin: under each mutant it
went red in one run of the two, so it's the one road whose mutant evidence varies. Why the rows move under a mutant is
not traced; the head's own outcome for each held in every control run.

view_widened_generic.ts never goes red: its read is made through T at Lit, so Lit's own view checks it, with or
without this pass.

## M3: 03_void_zero's status row back to NotYet

The row is put back to main's `NotYet` at 23:12, so the head's change to stage3/fixtures/predicates/status.json is
what the fixture test reads. `go test ./stage3/fixtures -run TestFixturesPredicates`. Red: TestFixturesPredicates at
predicates/03_void_zero.a/stage0 (J-nvgasnk).

## A read dispatched to a getter

A read finishAccessors makes a getter dispatcher's call (#3jj9dzg) is weighed each round from l.dispatchedReads, and
since a getter answers its name it stops NotYet in the getter's words, never marked. TestDispatchedReadOfAViewedValueStops
holds that, and TestUndispatchedTwinOfAViewedValueIsChecked holds its twin without the getter to the checked stop.
