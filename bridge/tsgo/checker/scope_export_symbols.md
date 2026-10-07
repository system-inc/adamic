# scope-export-symbols

The existing `tsgo_inspect` ABI accepts this question:

```
scope-export-symbols
<canonical unsigned 32-bit symbol flags>
<property name>
```

The exact queried node must be an Identifier, QualifiedName or
PropertyAccessExpression. The bridge returns raw checker resolution facts,
using the existing version 1 UTF-16-length-prefixed field protocol:

1. Header: version and question name.
2. `node-symbol-details`' symbol record for GetSymbolAtLocation(node).
3. The same symbol record for GetAliasedSymbol when the own symbol is an alias;
   otherwise an absent symbol record.
4. Opaque program-scoped identity of GetExportSymbolOfSymbol for the first
   GetSymbolsInScope(node, flags) result whose name matches the question's name,
   or zero when no symbol matches.

Each symbol record contains presence, identity, flags, name and exact declaration
metadata, as encoded by writeSymbolDetails in declaration_facts.go. The Adamic
side is scope_export_symbols.a; its focused symbol_facts_wave_22.a decoder
consumes and validates all metadata fields while retaining declaration locations.

The bridge does not compare an accessed member with the in-scope identity, test
lexical namespace containment, decide whether a qualifier is unnecessary, or
construct a finding or fix. Those decisions are in no_unnecessary_qualifier.a.
Alias declarations and local/export identities are facts rather than verdicts.
Existing ABI buffer ownership and handle lifetimes are unchanged. Identity zero
means absent; a nonzero identity is borrowed until its program is released.

TestScopeExportSymbols compares four scope/alias cases with direct checker
operations, including a shadowed member and an aliased namespace. It also rejects
noncanonical or overflowing flags, missing fields and a non-entity query node.
The wave 22 production byte oracle tests removing export normalization, while
the native released-handle probe tests this question after program release.
