# Mutants for a new collator's compare read as a value (#9tw2b9z)

Each lowering mutant is one exact line edit in internal/lower/library_intl.go, run on a Linux box
against the test it names, then put back from a saved copy. The oracle rows run with
`ADAMIC_GATE_UNCACHED=1`. The oracle's test is internal/oracle/library_intl_compare_value_test.go, and
the lower tests are in internal/lower/library_intl_test.go.

| Mutant | Edit | Red |
|---|---|---|
| (a) The receiver check dropped | `intlCollatorCompareRead`: `if receiver.Kind != ast.KindNewExpression \|\| !l.isIntlCollator(receiver.AsNewExpression().Expression) {` made `if false && (…) {`, the condition kept so it compiles | TestIntlCollatorCompareValueStopsArePinned, p02_held_collator and p04_class_compare, .a and .ts: the read is let past the method value refusal and meets intlCollatorCompareValue's "off a collator this read didn't make" instead. TestIntlCollatorOtherRoadsAreNotYet's three held-collator rows the same |
| (b) The exact-slot check dropped | `intlCollatorCompareRead`: `return l.exactFunctionSlot(node) \|\| intlCollatorCompareConst(node)` made `return true` | TestIntlCollatorCompareValueStopsArePinned, p03_wider_slot.a and .ts: the read into `(x: string, y: string) => unknown` lowers where the refusal stood |
| (c) The runtime's guard removed (Runtime's) | internal/native's gcc invariant harness removes the guard line in adamic_intl_collator_compare | TestGCCInvariantCollatorMutant holds the mutant to UBSan's null witness, and TestGCCInvariantCollator holds the control to the loud invariant stop |
