# Type declaration ancestry

The isolated raw question is `type-declaration-ancestry\nTYPE_ID`, implemented in
`bridge/tsgo/checker/type_declaration_ancestry.go` and decoded by
`type_declaration_ancestry.a`. The existing Go dispatcher gains one case/return
registration, two physical lines after gofmt. No existing Adamic registrations
or shared checker algorithms are changed.

TYPE_ID must be a canonical decimal, nonzero, live type identity from the same
program. The existing dispatcher validates exact file/span/kind and handle
lifetime before the question runs. The response is a version-1 framed record:

1. Version, mode, number of the type symbol's declarations (zero if no symbol).
2. For each declaration: source filename, number of nodes in its ancestry chain.
3. For each node, from declaration upward through its raw parents: syntax kind,
   byte start/end, declared name for a type alias or interface (empty otherwise),
   union member count (zero for other kinds), and presence of an alias's declared
   type. If present, the declared type's byte start/end and syntax kind follow.

The question returns raw symbol declarations and AST topology. It has no accepted
Nexus filenames, alias-name list, completeness decision, lint verdict, message or
edit. Adamic walks through parenthesized type nodes, requires a direct union arm
and a top-level alias declared by the known file, groups distinct arms by union,
and chooses the production rule's first complete recognized outcome type.

`TestTypeDeclarationAncestry` compares every serialized field against the direct
checker symbol and AST, including generic instantiated literal arms,
parenthesized unions and alias types. It also checks zero declaration framing,
malformed IDs and inexact anchors. A compiling union-count mutant finishes
normally and is killed only by independent Go production diagnostic bytes.
A released-handle probe and a compiling registry mutant prove the lifetime check.

Collection misuse and discarded pure results need no new checker questions.
`library_declarations.a` decodes the existing symbol-detail records into existing
DeclarationFact/DocTag/SymbolDetail classes without building an unnecessary local
binding index. `ancestor_node.a` and `outcome_arms.a` isolate the two helper classes.
