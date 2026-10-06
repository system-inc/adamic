# Checker facts for the six-rule pilot

`tsgo_inspect` adds one function to ABI 1. Adamic calls it through the explicitly
linked `tsgoInspect` prelude declaration. It borrows file/kind/question bytes,
copies them before entering Go, and returns exactly the stated UTF-8 byte count
in caller-owned C malloc memory. Free facts and errors with `tsgo_buffer_free`.
The native adapter decodes a fresh owned Adamic string and frees the C buffer.
No Go pointer crosses the boundary. All questions reject a released handle.

Selection is an exact TypeScript kind and UTF-8 byte span, including leading
trivia. `SourceFile` can have the empty span `0:0` only for an empty file. Other
empty spans, inexact nodes, unsupported questions and inappropriate node kinds
are errors. Type IDs are nonzero, stable identities within one program, not
addresses or separately owned handles. Compare IDs only from that same live
program; releasing it invalidates every ID. Target IDs support equality only and
need not have records in a particular answer.

Questions are these exact UTF-8 strings:

| Question | Checker fact |
| --- | --- |
| `raw-type` | `GetTypeAtLocation(node)` |
| `type` | That type, replaced by its base constraint when one exists |
| `base-type` | Constrained type with literals widened by `getBaseTypeOfLiteralType` |
| `signature` | A call/new/tagged-template's resolved signature, callee type, and parameter types/rest marks |
| `declarations` | A class/interface's named symbol declarations and its local symbol declarations, separately |
| `options` | A source file's program `strictNullChecks`, resolved against `strict` |
| `assignable` LF start LF end LF kind | Whether the selected node's type is assignable to the exact target node's type in the same file |

The bridge supplies these checker facts. It contains no finding predicate,
message, repair, or production cohere rule import.

Every wire field is `decimal-UTF-16-unit-count LF text`, with no extra separator.
Text can contain LF and non-BMP characters. This inner character count serves
Adamic's TypeScript string indexing; the surrounding C buffer always has a UTF-8
**byte** length. Numeric field texts are canonical decimal integers, booleans
are `0`/`1`, and list fields are a count followed by that many numeric IDs.
Every answer starts with numeric schema version `1`, then its question name
(`assignable` omits its target suffix).

* `options` and `assignable`: one boolean, then end.
* `declarations`: symbol-present boolean, declaration count and records; then
  local-symbol-present boolean, declaration count and records. Each declaration
  is kind, untrimmed byte start, byte end, source path, token-trimmed name start,
  name end. Name trivia is trimmed against the queried source, matching cohere's
  reporting context even for a declaration in another file. An absent name has
  `0:0`. Neither list is reordered or filtered.
* Type questions: strict-null-checks boolean, signature-present boolean (always
  true for type questions), root-ID list, rest-mark count and booleans, record
  count, records. A signature's first root is its callee; the rest are parameters.
  A signature with no resolution has only its callee and false presence.
  Rest roots are constrained, as required when reading a rest parameter; their
  number-index type is supplied too. Other parameter roots remain unconstrained.

A type record is: ID, TypeFlags, TypeToString, intrinsic-error boolean,
nondeferred-reference target ID (zero if absent), array boolean, tuple combined
ElementFlags (`-1` for a nontuple), base-constraint ID for a type parameter
(zero if absent), number-index element ID for a rest parameter (zero if absent),
union/intersection constituent ID list, reference type-argument ID list. Records
include the reachable constituent/argument/constraint/element graph, reserving
identities before recursion so cycles terminate. No target record is needed for
an equality question. Flags are from the pinned checker, held by numeric tests.

Adamic validates framing, canonical safe integers, natural counts/IDs, booleans,
schema, record identities, tuple sentinel, graph links and complete consumption.
Malformed facts panic; they cannot turn into a plausible finding.

Create now initializes the single-threaded checker pool in the load interval.
It does not request semantic diagnostics or resolve every type eagerly.
`ADAMIC_TSGO_TIMING=1` adds `run_ns`: elapsed wall time after successful create
until release begins. `query_ns` sums native adapter intervals, including exact
lookup, Go checker work/serialization, C copies, UTF-8 decoding and output frees.
Type resolution remains lazy and its cost belongs to the query that requests it.

`bridge/tsgo/cost` is a direct Go baseline for the identical `Inspect` work. It
intentionally shares that implementation to measure adapter costs. It is not
the findings oracle. The production cohere oracle has its own loader and invokes
the six unchanged rules; it imports no bridge code.
