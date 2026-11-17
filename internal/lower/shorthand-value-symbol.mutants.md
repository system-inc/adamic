# Mutants for valueSymbol, the symbol a shorthand name reads or writes (#3sdynvn)

One edit to locals.go, run on a house box, then put back.

| Mutant | Old text | New text | Red |
|---|---|---|---|
| valueSymbol's shorthand arm dropped | `func (l *lowering) valueSymbol(node *ast.Node) *ast.Symbol {` | the same line, then `return l.symbol(node)` as the body's first statement | every row of internal/lower's TestShorthandValueSymbol: enumeration_binding and predicate_overload_binding lower and no longer stop at run time, and predicate_overload_callback, record_read_discarded, static_class_escape and static_outside_helper lower where they're refused, as on main 1d3efe7b (probe J-wj7maxs) |

census_overload_proof.go's scan is the one shorthand walk this change leaves alone. Its own fix, aaafbfbb on compiler/overload-proof-shorthand-write-1 (#ffkhsm0), edits the same lines, and that branch's pin and mutant hold it.
