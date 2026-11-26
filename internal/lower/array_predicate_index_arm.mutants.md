# Mutants for the index-only record arm in Array.isArray's domain (#gcec0bx)

rulings.md:260's amendment (Nov 26, @system_adamic on #gcec0bx) admits an index-only arm by four conditions, in library_array_predicate.go's arrayExcludedIndexArm. Each mutant is one edit to library_array_predicate.go, planted on a house box, the named tests run, then put back.

| Mutant | Old text | New text | Red |
|---|---|---|---|
| M1: condition 2 dropped, every index value admitted | `if !l.indexValueAdmitsNoCallable(infos[0].ValueType(), map[*checker.Type]bool{}) {` | `if false && !l.indexValueAdmitsNoCallable(infos[0].ValueType(), map[*checker.Type]bool{}) {` | internal/lower's TestArrayPredicateIndexArmRefusals: unknown-value and callable-value, each lowering past isArray; TestArrayPredicateIndexArmValueType/open: Bag, Loosely, Calls and Unknowns admitted |
| M2: the array read as the record, the brand answering false at a site whose arm is index-only | `return ir.ArrayIsArray{Value: fit(value, ir.Union)}, true, nil` | the same, after `if sites := l.arrayArmSites; len(sites) > 0 && sites[len(sites)-1].node == argument && len(l.checker.GetIndexInfosOfType(sites[len(sites)-1].arms[0])) > 0 { return ir.BooleanConstant{Value: false}, true, nil }` | internal/oracle's TestArrayPredicateIndexArmAgrees: json_record_test on its array line (true where Node prints false), and json_record on its array line where it lowers |
| M3: condition 4 dropped, the index arm not recorded in arrayArmSites | `if arms != nil {` | `if arms != nil && l.arrayExcludedArm(proven) {` | internal/lower's TestArrayPredicateIndexArmViewStops, the overlap view of Named over `[number, number] \| Other` no longer stopping the isArray over AdamicJsonRecoveryObject |

M1 leaves TestArrayPredicateIndexArmRefusals/array-stored-string-index green, since condition 3 still stops it, and TestArrayPredicateIndexArmValueType/closed-world-string-index green for the same reason. M3 leaves every refusal and the fixtures green: nothing in them makes a checked view reaching the arm.
