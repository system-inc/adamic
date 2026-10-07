# Wave 23 continuation facts

Both questions use ABI version 1 and the existing framed `tsgoInspect` call.
Exact file/range/kind checks run before dispatch. Identities belong to one live
program; zero, unknown and noncanonical identities are refused. Release removes
the program from the C registry before a question can reach its checker. Output
buffers remain caller-owned and use the existing C-free path.

## declaration-ancestry

Request: `declaration-ancestry\n<type identity>`.

Response after the usual version/question header: declaration count, then for
each declaration its ancestor count, starting at the declaration and walking
through `Parent` to the root. Each ancestor has eight fields:

1. AST kind without the `Kind` prefix.
2. Name text, or empty when unnamed.
3. Byte position.
4. Byte end.
5. Source filename.
6. Number of direct `ForEachChild` children.
7. Written alias type's byte position, or zero for other nodes.
8. Written alias type's byte end, or zero for other nodes.

The query supplies every declaration of the type's own symbol, with no filtering
by filename, alias name, node kind or rule. Empty declaration lists are valid.
Instantiated type literals retain their original declaration locations. No AST
pointer crosses the ABI. Adamic groups distinct arm locations by union location,
checks the raw ancestry, compares the written alias type location, requires a
source-file parent and matches the five production Nexus name/path pairs. It then
chooses the first complete alias in Go cohere's specified order.

The decoder is [declaration_ancestry.a](declaration_ancestry.a); the Go provider is
[declaration_ancestry.go](../../../bridge/tsgo/checker/declaration_ancestry.go).

## awaited-shape

Request: `awaited-shape`, without a suffix, at an expression node.

Response is the existing type-graph schema for the checker's awaited type of the
expression's raw type. A missing awaited type has `present = false` and no roots.
It includes the strict-null flag, present flag, root IDs, an empty rest list and
all referenced graph records. No outcome, collection or purity judgment is made.

The decoder is [awaited_shape.a](awaited_shape.a); the Go provider is
[awaited_shape.go](../../../bridge/tsgo/checker/awaited_shape.go).

Only one line in this worker's previous `type_alias_info.go` provider changes.
Its existing supported questions run unchanged; otherwise it delegates to a
new-file dispatcher for these two questions and retains the unsupported-question
refusal. The shared `facts.go` dispatcher is unchanged from the first wave. Adamic's dynamic question strings need no shared registration edit.
The pure-result and collection rules use existing raw symbol details, constrained
and raw type graphs, literal values, property shapes, call counts and type origins.
[node_symbol_details.a](node_symbol_details.a) decodes the existing declaration
wire format; it contains no lint verdict.

The direct checker test checks all eight ancestry fields against the actual AST,
including parenthesized aliases and instantiated arms, checks the awaited opaque
identity, and rejects malformed suffixes and IDs. Both new fact mutants compile
and run normally and are caught by the independent Go diagnostic stream.
