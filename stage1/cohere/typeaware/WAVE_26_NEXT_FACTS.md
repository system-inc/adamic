# Wave 26 next batch raw questions

All questions retain ABI version 1 and the length-framed fields in facts.md.
Go and Adamic implementations have matching names. The only shared bridge edit
is one switch registration line per question. No registration generator or shared
test harness is changed. The new runner and test file belong to this batch.

`node-structure` requires an exact SourceFile. It refuses parse diagnostics and
returns emit ModuleKind, a node count, then preorder syntax records. Each record
has Kind, byte Pos, End, token start, parent identity, NodeFlags, identifier/string
text, child identities, eight optional role identities (expression, name, type,
body, left, right, whenTrue, whenFalse), argument identities, parameter identities,
binary operator Kind and JSX spread presence. Zero is an absent identity. Roles
are populated for the syntax kinds documented by node_structure.go; every AST
child is included. The Adamic decoder checks bounds. These are syntax facts, with
no candidate selection, rule names, messages, findings or edits.

The existing Adamic parser has no JSX support. This batch consumes this syntax
projection for all three rules rather than changing that shared parser. This is
a different syntax source from the previous wave, and costs a complete serialized
AST per root even when a rule has no candidates. Rule judgments and traversal are
native Adamic. The Go oracle imports no bridge code and calls production rules.

`declaration-lineage\nnode` returns the unaliased symbol at the exact node.
`declaration-lineage\nsignature` returns the resolved call signature declaration.
`declaration-lineage\ntype\n<ID>` returns a type's symbol declarations. Header
fields are symbol identity, SymbolFlags, name and declaration count. Each
declaration carries source path, actual compiler default-library classification,
type-only import/export flag and the complete ancestor chain. An ancestor has
Kind, name text, name Kind, byte Pos, End and NodeFlags. Signature metadata has
zero symbol identity and flags. The bridge does not decide whether a declaration
is a global listener, event parameter, element interface, namespace or MockTracker.

`literal-string\n<ID>` accepts number and bigint literal identities, returns
whether the compiler literal value implements fmt.Stringer and its String value.
Computed values without that representation return false and empty text. Adamic
chooses which values can leak. This differs from the existing literal-value
question by preserving precisely the production rule's fmt.Stringer test.

`transformed-shape\n<ID>\nnon-nullable` calls GetNonNullableType.
`transformed-shape\n<ID>\nconstraint` calls GetBaseConstraintOfType. Both return
ordinary Types framing, with absent roots when the compiler returns nil. Only
Adamic decides whether a constraint is relevant, how far to follow it, and what
assignability means for a rule. Noncanonical, unknown and zero identities, malformed
selectors and suffixes are rejected. Handles and output ownership use the existing
bridge implementation without changes.
