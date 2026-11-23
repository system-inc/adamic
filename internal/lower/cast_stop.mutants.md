# Mutants for the cast causes (#b5w3ycg, Nov 22)

## M1: one sentence again

`castStop` returns main's refusal, "a cast the runtime can't check" with castRepair's fix, for every cause in .ts as it does in .a.

Expected: `TestCastStopNamesItsCause` goes red on every .ts probe (each want names its cause), and `TestUncheckableCastsStayRefused` stays green, since every one of its probes is still a refusal naming adamic/no-unchecked-cast.

## M2: a ruled check refused

`castStop` answers `castUntaggedUnion` with a refusal in the NotYet's words, so the kind moves and the words don't.

Expected: `TestCastStopNamesItsCause/untagged_union` goes red on the kind (want refused=false), the words alone wouldn't catch it.

## M3: an array's interface read as a dictionary

`castExtendsArray` answers false, so an interface whose base-type chain reaches Array or ReadonlyArray (tsc's NodeArray, MutableNodeArray, MissingList, SortedArray) falls through to castIndexSignature (#2v3xtgm, Compiler's ruling).

Expected: `TestCastStopNamesItsCause/missing_list` and `TestCastStopNamesItsCause/node_or_node_array` go red on the kind (a NotYet dictionary view where the array's refusal is wanted), and `TestCastStopNamesItsCause/map-like_dictionary` stays green, since a real dictionary never had an array base.

## M4: the dictionary view's test dropped

`dictionaryViewValue` returns an unchecked narrow, `ir.Narrow{Value: held, To: ir.Record}`, so neither the record test nor the references test runs at the cast (#2v3xtgm, re-ruled Nov 23). The record test can't be dropped alone: the references test reads a record's shape, which a value that isn't one doesn't have.

Expected: oracle's `TestDictionaryViewStopsAtTheCast/array_cast.ts`, `/fixed_literal.ts` and `/class_instance.ts` go red, the JavaScript backend reading on as Node does and native reading a fixed object or an array as a record's table, and `/number_record.ts` and `/boolean_record.ts` go red too. lower's `TestDictionaryViewCasts` goes red on every admitted row, whose helper must return a checked narrow.

## M5: the references test dropped

`dictionaryViewValue` sets no `ReferencesMessage`, so a record of numbers or booleans passes the cast.

Expected: oracle's `TestDictionaryViewStopsAtTheCast/number_record.ts` and `/boolean_record.ts` go red on both backends: natively a scalar's bits are read as a pointer, and on the JavaScript backend the value reads on as Node does. The array, literal and class rows stay green, since the record test still stops them.

## M6: the JavaScript mark always a record of references

javascript's `recordLiteral` marks every record literal `"reference"`, whatever its element.

Expected: oracle's `TestDictionaryViewStopsAtTheCast/number_record.ts` and `/boolean_record.ts` go red on the javascript backend alone, which reads the number and the boolean as Node does where native stops. The native rows stay green.

## M7: the use test dropped

`dictionaryViewUse` answers true for every use, its first line `return true`, so the cast's result may be returned, stored, written or handed to any callee and the view is still taken (#2v3xtgm, J-psxf2gy: the use test on the cast's own value replaced wave 2's program-wide write walk by type, which no tsc unit passed).

Expected: lower's `TestDictionaryViewCasts/returned`, `/stored_in_a_const_then_written_by_key` and `/handed_to_a_helper_that_writes` go red, admitted where the NotYet at the cast is wanted, and so do `/handed_on_to_a_helper_that_writes`, `/handed_to_a_self-recursive_helper`, `/assigned_to_a_let`, `/written_through_a_dictionary` and `/deleted_through_a_dictionary`. Every admitted row stays green, and so do the rows a type keeps stopped (`/undefined_dropped`, `/number_values`, `/union-held_values`), whose stop comes before the use test is asked.
