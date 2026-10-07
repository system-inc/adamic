# Type declaration ancestry

`type-declaration-ancestry\nraw` reads GetTypeAtLocation.
`type-declaration-ancestry\nawaited` also reads GetAwaitedType.
Other suffixes are rejected. Exact AST selection and registry handle checks apply.

After the version and question header, the response holds declaration identities,
then a node count and seven fields per node: identity, parent identity, kind,
name, source filename, direct type-node identity, and union arm count.
Zero is the absent identity. Identities are local to this response, never pointers.
Every union member's own symbol declarations contribute roots; duplicates are
preserved. Ancestors and direct type nodes of aliases and parenthesized types are
included once. The type-node field is zero for other kinds; the arm count is zero
for non-unions. This question has no lint names, paths or policy predicates.

The Adamic decoder is stage1/cohere/typeaware/type_declaration_ancestry.a.
Inspect's fallback dispatch changes one shared line. The native suite imports the
decoder through its own rule file. Existing question behavior is preserved.
