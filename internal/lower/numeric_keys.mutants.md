# Mutants for numeric computed keys (Adamic's ruling (A), Nov 24 23:48:03Z on #d1chbm7)

Each is one edit in numeric_keys.go, run on a house box, then put back. Line numbers are the file as landed.
A row's red is its subtest of TestNumericKeysKeepTheirStop unless it names another test: each of those rows
wants NotYet "a computed field name" at the key, so a key the mutant admits reds the row whether the program
then lowers or stops somewhere else.

| Condition | Mutant | Edit | Red |
|---|---|---|---|
| 3, no key spelled twice | A key spelled twice passes | numeric_keys.go:118, `if seen[key] {` made `if false {` in distinctFieldKeys | TestDistinctFieldKeys, the `0, name, 0` row |
| 3, no key spelled twice | The duplicate test dropped at its call | numeric_keys.go:47, `if !spelled \|\| !distinctFieldKeys(keys) {` made `if !spelled {` | None: survives. tsc refuses every pair of keys that spell one string first (TS1117, by the name the key's type gives it, an enum alias's included), so no program with one reaches lowering; TestNumericKeyDuplicatesStopAtTheChecker holds that and stays green. Recorded as a survivor, not a gap in the tests: the check holds the ruling's (1) where lowering stands |
| 4, a const nothing exports | A let admitted | numeric_keys.go:140, the `\|\| list.Flags&ast.NodeFlagsBlockScoped != ast.NodeFlagsConst` term dropped | `/a_let_table` |
| 4, a const nothing exports | An export modifier admitted | numeric_keys.go:143, `if ast.GetCombinedModifierFlags(declaration)&ast.ModifierFlagsExport != 0 {` made `if false {` | `/an_exported_table` |
| 5, every reference a keyed lookup | Any reference admitted | numeric_keys.go:212, `return true` put first in numericKeyedLookup, above `object := numericKeyOuter(reference)` | `/exported_by_a_declaration`, `/passed_to_Object.keys`, `/spread`, `/written`, `/printed`, `/aliased`, `/read_by_a_string_key`, `/called_through` |
| 5, every reference a keyed lookup | A write admitted | numeric_keys.go:223, `if ast.IsAssignmentTarget(read) {` made `if false {` | `/written` (finiteElementWrite lowers the write) |
| 5, every reference a keyed lookup | A key that isn't a number admitted | numeric_keys.go:218, `if member.Flags()&checker.TypeFlagsNumberLike == 0 {` made `if false {` | `/read_by_a_string_key` (finiteElementRead reads field "0" by the string "0") |
| 5, every reference a keyed lookup | A call through the lookup admitted | numeric_keys.go:230, `return parent.AsCallExpression().Expression != read` made `return true` | `/called_through` |

The proof's memo has its own: numeric_keys.go:158, the held answer never returned (`if closed, held := ...; held && false {`),
reds TestNumericKeyTableIsWalkedOncePerBinding (three walks for one binding).

The key's enum read has its own (#kgngvxj): numericKeyRead returning nil before anything else, so (A) names the field
and never reads the ordinary enum member, reds TestEnumInitializationComputedKeyStops in internal/oracle (it prints
`initializer` and `1` again where the pin wants exit 70 with the row's ReferenceError), and the fixture runner's row for
stage3/fixtures/enum-init-reach/unknown-computed-key.a in backend_divergences.json.

## Road 2: ordinary enum keys in enumeration order

These patches are owed to Loom and have only been checked for applicability.

| Condition | Mutant | Edit | Red |
|---|---|---|---|
| Array indices increase strictly | `wrong-drops-ascending-check.patch` | numeric_keys.go:310, disable the ascending comparison in keysInEnumerationOrder | TestNumericKeyOrderedEnumStops/descending at numeric_keys_test.go:266 and the descending order table row |
| The enum must be initialized before the key is read | `wrong-key-without-readiness.patch` | object.go:98-102, omit the ordinary enum key's member read | TestNumericKeyOrderedEnumReadiness, the unresolved callback at numeric_keys_test.go:327 and checked key assertion at :374 |
