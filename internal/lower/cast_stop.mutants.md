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

## M7: the write walk dropped

`dictionaryViewWritten` answers false, so a write through a dictionary the target is assignable to no longer keeps the NotYet.

Expected: lower's `TestDictionaryViewCasts/written_through_a_dictionary` and `/deleted_through_a_dictionary` go red, admitted where the NotYet at the cast is wanted. Every other row stays green.
