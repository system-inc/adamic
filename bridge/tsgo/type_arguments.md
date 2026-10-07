# Reference and alias arguments

`type-arguments\nID` is an ABI-1 `tsgo_inspect` question. ID is a canonical,
nonzero decimal identity borrowed from the same live program. The node is still
checked by exact file, byte bounds and kind. Missing, stale or noncanonical
identities and extra fields are refused.

After the usual schema and question header, fields are:

1. Type symbol name, or an empty string.
2. Alias symbol name, or an empty string.
3. Reference argument identities, framed as a count followed by identities.
4. Alias argument identities, framed the same way.
5. The ordinary type graph frame: strict-null-checks, present, roots, rest count,
   and records. Roots are reference arguments followed by alias arguments.

The Go file `checker/type_arguments.go` exposes the checker's arguments and names.
It chooses no wrapper and makes no lint decision. Adamic decodes the result in
`stage1/cohere/typeaware/type_arguments.a`; the operation-context rule chooses
Promise, Readonly, nullable-union and array unwrapping in production Go order.
Names use the existing UTF-16 field lengths and valid UTF-8 output conversion,
including replacement of invalid bytes in internal checker symbol names.
The output buffer uses the existing owned C-string contract. Type identities
remain borrowed from the program and cannot be queried after its release.
