# Symbol provenance for wave 29

`symbol-provenance` accepts an exact Identifier or PrivateIdentifier. Adding
`\nalias` follows the checker's aliases. Other suffixes and node kinds are refused.

The existing UTF-16-length field framing and C buffer ownership contract apply.
The first fields are schema version 1, question name, and program-scoped symbol
identity. Identity 0 ends the response. A present symbol then supplies:

1. Display name and declaration count.
2. For every declaration: source path, kind, UTF-8 start and end, declaration-file
   boolean, and nearest enclosing ambient string module name (empty if absent).
3. Value-declaration-present boolean, then its path, kind, start, end and the
   compiler's const flag when present.
4. Current source-file import count. Each import supplies its text and resolved
   boolean; a resolved import additionally supplies resolved path, external-library
   boolean and package identity name.

The Go side reports raw compiler records and module resolution. Adamic tests
which module and declaration kinds belong to Node or sharp, follows native
parsed const initializers, proves counter dependencies and path equality, checks
intervening writes, and constructs all diagnostics. No lint verdict, edit or
suggestion crosses this question. Identities are borrowed from the live program;
returned strings are owned independently under the existing ABI.

`checker/symbol_provenance_test.go` checks raw and followed aliases, ambient
origins, Unicode, private fields, unresolved names and malformed requests.
`stage1/cohere/typeaware/wave_29_test.go` compares the resulting native decisions
with unchanged production Go rules, including sanitizer and lifetime checks.
