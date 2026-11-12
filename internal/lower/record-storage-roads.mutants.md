# Mutants for the record storage roads (#mez92e7)

Each is one edit, run on a house box, then put back. The red is internal/oracle's TestRecordStorageRoadsArePinned: the row lowers, and the failure prints its sanitized native run beside Node's 1 3.

| Mutant | File | Old text | New text | Red |
|---|---|---|---|---|
| The shorthand dropped | records.go | `} else if node.Kind == ast.KindShorthandPropertyAssignment {` | `} else if false && node.Kind == ast.KindShorthandPropertyAssignment {` | f02_shorthand.ts |
| The logical assignments dropped | records.go | `(viewSite(node) \|\| logicalAssignmentRight(node))` | `(viewSite(node) \|\| false && logicalAssignmentRight(node))` | f05_logical_assignment.ts, f05b_or_assignment.ts, f05c_and_assignment.ts |
| The binding check dropped | constraint_relation.go | `if l.sameRecordStorage(binding.argument, binding.constraint, map[[2]*checker.Type]bool{}) {` | `if true \|\| l.sameRecordStorage(binding.argument, binding.constraint, map[[2]*checker.Type]bool{}) {` | f07_constraint_binding.ts, f07b_written_type_argument.ts |
| The instantiation expression unbound | type_argument_bindings.go | `return l.instantiationTypeArgumentBindings(node, visit)` | `return nil` | f07c_instantiation_expression.ts |
