# Wave 23 compiler facts

These questions use the existing ABI-v1 `tsgo_inspect` entry point. Inputs select
an exact node by file, UTF-8 span and kind. Type identities belong to the live
program. All fields use the existing decimal UTF-16 length framing; the C-owned
outer buffer still uses an explicit UTF-8 byte length and is freed after copying.
No question returns a lint verdict, diagnostic, fix or suggestion.

`type-alias-info\n<type identity>` returns the schema header, alias-symbol origin
record, then the existing type-graph encoding. The origin record is a presence
boolean followed, when present, by the symbol name and declaration count; each
declaration has its source path, declaration-file boolean and default-library
boolean. The graph roots are the alias's type arguments in checker order. Missing
aliases have no origin and no roots. Malformed, noncanonical, zero and foreign
identity values are refused. `type_alias_info.a` validates and decodes this record.

`then-callback-parameters` returns the schema header and existing type graph.
It gets the node's apparent receiver type, its union constituents' `then`
properties, those property types' union constituents and call signatures, and the
apparent type of each signature's first parameter. A rest parameter contributes
its numeric index type; a missing index type contributes nothing. The graph roots
preserve this order. Suffixes are refused. `then_callback_parameters.a` decides
whether any root's union constituent has a call signature. This preserves Go
cohere's thenable rethrow decision, including rest callback parameters.

Both Go implementations are in files named for their question. The shared
`facts.go` change is one line delegating previously unsupported questions to the
new extension. Unsupported questions retain the prior refusal. There is no
shared Adamic dispatcher edit: the new decoders call the existing dynamic inspect
entry point directly. No protected compiler file or ABI declaration was changed.
