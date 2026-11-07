# Mutants for a collator's compare read as a value (#9tw2b9z)

Each mutant is one exact line edit in internal/lower/library_intl.go, run on a Linux box against the
test it names, then put back from a saved copy. The oracle rows run with `ADAMIC_GATE_UNCACHED=1`.
The oracle's test is internal/oracle/library_intl_compare_value_test.go, and the lower tests are in
internal/lower/library_intl_test.go. None has been run yet.

p03 and p04's stops are `record` until measured. Read from methodValueRefusal, each should be its
refusal, at the read:

- p03_wider_slot.a: `internal/oracle/testdata/intl_collator_compare_value/p03_wider_slot.a:6:49: Adamic 0.1 refuses a method read as a value (compare would lose its object, and this with it); call it in an arrow that keeps the object: (x, y) => collator.compare(x, y) (unbound-method)`
- p04_class_compare.a: `internal/oracle/testdata/intl_collator_compare_value/p04_class_compare.a:11:18: Adamic 0.1 refuses a method read as a value (compare would lose its object, and this with it); call it in an arrow that keeps the object: (a, b) => keyed.compare(a, b) (unbound-method)`

and the .ts twins the same at their own paths. Record what is measured, and run the mutants after.

| Mutant | Edit | Red |
|---|---|---|
| (a) The type check dropped | `intlCollatorCompareRead`: `if !l.isIntlCollatorType(l.checker.GetTypeAtLocation(node.AsPropertyAccessExpression().Expression)) {` made `if false && !l.isIntlCollatorType(l.checker.GetTypeAtLocation(node.AsPropertyAccessExpression().Expression)) {` | TestIntlCollatorCompareValueStopsArePinned, p04_class_compare.a and .ts: Keyed's compare, held in a const with no annotation, is admitted and lowers to a read of a compare field Keyed's object doesn't have, where the refusal stood. p01 to p03 stay as they were: their receivers are collators |
| (b) The exact-slot check dropped | `intlCollatorCompareRead`: `return l.exactFunctionSlot(node) \|\| intlCollatorCompareConst(node)` made `return true` | TestIntlCollatorCompareValueStopsArePinned, p03_wider_slot.a and .ts: the read into `(x: string, y: string) => unknown` lowers where the refusal stood. Also TestIntlCollatorOtherRoadsAreNotYet in internal/lower: compare into a wider slot, held in a let, as an === operand, handed to sort and through call or bind each lower or meet other words |
