# type-reference-graph

The exact-node `tsgo_inspect` question `type-reference-graph` has no suffix.
It exposes type syntax and direct symbol declarations, including declarations
from another source file. The result contains no lint verdict or repair.

Every field uses the existing decimal UTF-16 length, LF, text framing. The
outer C string remains an owned UTF-8 buffer with an explicit byte length.
The header is schema `1`, question name, root ID, record count. Each record is:

1. Syntax kind without the `Kind` prefix.
2. Program-scoped symbol identity, or zero. Identifiers use the checker's direct
   `GetSymbolAtLocation` result; aliases and shorthand values are not substituted.
3. Body record ID, or zero. This is the Type field of a type alias, index
   signature, mapped type or parenthesized type.
4. Count and IDs of syntax children. Interfaces expose their members; type
   literals, unions, intersections, conditional types, indexed access types and
   type references expose their ordered AST children.
5. Count and IDs of direct symbol declarations for an identifier.

Record IDs are one-based and local to this response. Zero means no body.
Repeated AST nodes have one record, so cyclic declaration graphs terminate.
Constructor kinds such as arrays and function types retain their kind but are
opaque in this projection. Qualified names also retain their kind. Adamic
chooses which kinds to follow and compares symbol identities to the anchor.

`stage1/cohere/typeaware/type_reference_graph.a` validates the header, root,
body and list edges before traversal. It owns copied records; record IDs do not
keep a checker program alive. Program-scoped symbol IDs remain borrowed and
expire with that program. A question on a released handle is refused before
inspection. The wire ABI and C buffer ownership contract are unchanged.

The direct checker test compares graph identities with direct symbol lookups,
checks a two-file cycle, retains an opaque array constructor and rejects suffixes.
The independent production cohere finding oracle catches both a native symbol
comparison mutant and a Go symbol-identity omission mutant with diagnostic bytes
alone. The wave report records those runs and the released-handle mutation.
