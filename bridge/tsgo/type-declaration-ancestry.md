# type-declaration-ancestry

Request: `type-declaration-ancestry`, optionally LF `awaited`. Other suffixes are
rejected. The checker obtains the raw type at the exact requested node; awaited
mode applies GetAwaitedType. A nil type has zero parts. A union is split one
level into its constituents; other types are a single part. Intersections are
not recursively split and constraints are not substituted.

Version 1 inspect framing and existing UTF-8 C-buffer ownership apply. After
version and `type-declaration-ancestry` come:

1. Constituent count. For each constituent, its symbol's declaration count.
2. For each declaration: source filename, then ancestry count. The ancestry
   starts with that declaration and walks Parent through the source file.
3. For each ancestor: kind without Kind prefix, raw byte start and end, declared
   name or empty, type-child-present boolean. A TypeAliasDeclaration's Type or
   ParenthesizedType's Type supplies the child, with raw byte start/end when
   present. Other ancestors have no type child in this record.
4. Number of direct ForEachChild children of that ancestor.

This exposes unfiltered declaration locations and structural facts. The native
rule checks paths and names, walks parentheses, decides whether an alias is at
file scope, groups arm identities and checks completeness. None of those lint
judgments are performed by the bridge. Source filename and byte bounds identify
original declaration locations within the unchanged program; they are not
pointers or separately owned handles. Released program handles are rejected.
