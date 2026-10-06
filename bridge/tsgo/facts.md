# Checker facts for native type-aware rules

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
program; releasing it invalidates every ID. Target IDs support identity equality and naming, but need not have records in a
particular answer.

Questions are these exact UTF-8 strings:

| Question | Checker fact |
| --- | --- |
| `raw-type` | `GetTypeAtLocation(node)` |
| `type` | That type, replaced by its base constraint when one exists |
| `base-type` | Constrained type with literals widened by `getBaseTypeOfLiteralType` |
| `signature` | A call/new/tagged-template's resolved signature, callee type, and parameter types/rest marks |
| `raw-shape`, `type-shape`, `signature-shape` | The corresponding type/signature graph, with empty name fields and no TypeToString work |
| `name` LF type-ID | TypeToString for a previously returned identity in this live program |
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
(`assignable` and `name` omit their argument suffix).

* `options` and `assignable`: one boolean, then end.
* `name`: one TypeToString text field, then end. Zero, unknown, noncanonical and
  released type identities are refused; an exact selector remains required.
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

The optimized suite requests shapes for arguments, signatures and boolean
comparisons. It requests a type name only while rendering an actual finding.
Plus operands retain full named facts because their RegExp decision needs the
rendering. No type string is cached across checker operations. Old questions
and wire schemas retain their meaning. Numeric encoding counts UTF-16 units
without allocating rune slices or using formatted printing; native numeric
fields are validated directly in the owned frame without temporary strings.

Set `ADAMIC_TSGO_PROFILE=/absolute/path/cpu.pprof` to collect an opt-in CPU
profile after load until the last program release. It includes native leaf PCs
and Go checker/collector CPU. Go reports timed inspect/parts bodies and
allocation counters; C reports input, public-call and output intervals plus
fact bytes. These profiling runs are separate from throughput measurements.
The public-call interval minus Go-body intervals estimates boundary overhead,
including copies, cgo and timer overhead; it is not a pure cgo latency measure.
CPU seconds can exceed wall time because background Go collection is concurrent.


## Questions added for the next ten rules

All questions retain schema 1, exact node selection, explicit UTF-8 byte lengths,
UTF-16 field framing, and the existing ownership and released-handle checks.
They supply compiler facts; all rule predicates and repair construction run in
Adamic. Type-ID arguments must be canonical, nonzero IDs already issued by the
same live program. A property name consumes the remaining bytes after its second
LF, so embedded LF is not an extra argument.

| Question | Compiler fact |
| --- | --- |
| `strict-this` | SourceFile's resolved `noImplicitThis` compiler option |
| `assignable-types` LF source-ID LF target-ID | `isTypeAssignableTo` on those two types |
| `widened-shape` | Raw node type passed through `getWidenedType`, with unnamed records |
| `enum-types` | Union enum-literal constituents' parent enum type identities, in checker order |
| `type-symbol` LF ID | The type's symbol name, or empty when absent |
| `scope-locals` | SourceFile's binder locals tables and enum exports tables |
| `contextual-shape` | `getContextualType(node)`, with false presence if absent |
| `property-shape` LF ID LF name | Property's type at the selected node, with false presence if absent |
| `call-returns` | Return types of every call signature of the raw node type |
| `apparent-shape` LF ID | `getApparentType` of the issued type |
| `call-parameters` LF ID | Apparent types of the first parameter of every nonempty call signature |
| `call-count` LF ID | Number of call signatures of that type |
| `symbol-origin` | Value declaration's source path for the node's symbol, or empty |
| `type-origin` LF ID | Type symbol name and its declaration files/library membership |
| `base-shapes` LF ID | Base types of a class/interface symbol's declared type |
| `property-info` LF ID LF name | Property value declaration kind, initializer kind and first parameter shape |

`widened-shape`, `contextual-shape`, `property-shape`, `call-returns`,
`apparent-shape`, `call-parameters` and `base-shapes` use the existing type-graph
wire schema with empty names, no rest marks, and the listed roots. A missing
contextual/property type has false presence and no roots. Signature/base lists
may be empty with true presence. Call-return types remain unconstrained.

`assignable-types` returns one boolean. `enum-types` returns an ID list.
`type-symbol` and `symbol-origin` return one string. `call-count` returns one
natural number. `type-origin` returns symbol-presence; if present, name, count,
then each declaration's source path, declaration-file boolean and the program's
`IsSourceFileDefaultLibrary` boolean. `property-info` returns value-declaration
presence; if present, declaration kind, initializer kind (empty if absent), first
parameter identifier name (empty if absent or not an identifier), and first
parameter annotation kind (empty if absent). These shapes deliberately do not
answer whether a method is dangerous or a global is exempt.

`scope-locals` requires a SourceFile. After the header it returns table count,
then table kind (`locals` or `exports`), container kind/start/end, symbol count,
and alphabetically sorted symbols. Each symbol is name, SymbolFlags, declaration
count and declaration kind/start/end triples in compiler order. Adamic joins
these byte spans to its own parser nodes, chooses enclosing scopes, applies the
production rule's exemptions and renders its own source positions.

Explicit root lists now retain configured `.d.ts` roots. An explicit `.a` root
sets `AllowNonTsExtensions`, so the checker reads its TypeScript syntax directly.
No content mapper or Adamic syntax translation is introduced.

The volume cost probe accepts `property-info:NAME` as a command-line shorthand:
one raw-shape query obtains the live ID, then N identical property-info calls
are measured. The direct Go probe performs the same bootstrap outside its
query-loop timer. This shorthand belongs only to the probes, not the C API.

`strict-this` returns one boolean. The volume runner checks it once and refuses
a program with `noImplicitThis` disabled: the special implicit-this diagnostic
variants are not part of this strict-config port. Refusal is explicit before
any finding is printed. The original six-rule runner is unaffected.


## Inventory coverage questions

The next coverage runner keeps the same ABI, ownership and framing. Symbol IDs
now join type IDs as separate program-scoped borrowed identity namespaces: they
are nonzero safe integers, never addresses, never owned independently, and die
with the program. A symbol ID cannot substitute for a type ID. Every question
still selects an exact node, even when its additional argument identifies a type.

| Question | Raw checker operation or fact |
| --- | --- |
| `binding-declarations`, `alias-declarations` | Identifier symbol declarations; shorthand values use their value symbol; the latter follows aliases |
| `resolved-name` LF name | Value-name resolution at the selected node and declaration/library origins |
| `identical-types` LF source-ID LF target-ID | Checker type identity relation |
| `literal-value` LF ID | Boolean, string or number literal value, without a lint interpretation |
| `annotation-shape` | Type obtained from the selected type annotation |
| `symbol-shape` | Global type of the selected node's symbol, not its narrowed location type |
| `type-metadata` LF ID | ObjectFlags, symbol/target symbol identities, target ObjectFlags, readonly tuple mark |
| `type-properties` LF ID | Property display names, symbol identities, SymbolFlags, readonly marks and global property types |
| `property-exists` LF type-ID LF property-symbol-ID | `GetPropertyOfType` using the original internal property key |
| `function-signatures` LF ID | Call signatures' parameter names/rest marks, return types and global parameter types |
| `contextual-argument` LF index | A call/new expression argument's contextual type |
| `annotated-return-shape` | A function-like node's annotated return type, absent when unannotated |
| `symbol-identities` | Identifier's symbol, alias, unknown-alias mark, locally resolved export symbol and shorthand-value symbol |
| `node-symbol-details` | Selected node's symbol and full declaration records |
| `declaration-details` | Selected declaration's full record |
| `type-symbol-details` LF ID | Type's own symbol and alias symbol, separately |
| `property-declarations` LF ID LF name | Property symbol and full declaration records |
| `container-bases` | Declared class/interface base and implemented types |
| `reference-shape` LF ID | Reference type graph including deferred reference arguments |

`binding-declarations` and `alias-declarations` return presence, SymbolFlags,
declaration count, then path/kind/untrimmed-byte-start/end per declaration.
`resolved-name` uses the existing type-origin symbol schema. `identical-types`
and `property-exists` return one boolean. `literal-value` returns a tag
(`missing`, `boolean`, `string`, `number`) and a value string. `type-metadata`
returns ObjectFlags, symbol ID (zero if absent), target ObjectFlags, target symbol
ID (zero if absent), and readonly-tuple boolean. `type-properties` returns count,
then display name/symbol ID/SymbolFlags/readonly boolean per property, followed
by the existing unnamed type-graph fields (without another schema header), whose
roots are the property types in the same order.

TypeScript has private internal property keys containing invalid UTF-8. Display
names replace those bytes with U+FFFD so the boundary remains UTF-8. Opaque
property symbol IDs preserve the original key for `property-exists`; list
membership and replacement display names are not substitutes for checker lookup.
`function-signatures` returns signature count, then each signature's parameter
count and name/rest-mark pairs, followed by existing unnamed graph fields. Roots
flatten each signature's return type followed by its parameter types.

`symbol-identities` returns own symbol ID, alias ID, unknown-alias boolean,
resolved local export ID, and shorthand-value ID; absent IDs are zero. A symbol
record is presence, then ID/SymbolFlags/display-name/declaration-count and its
declarations. A declaration record is path/kind/start/end/declaration-file mark/
default-library mark, parent kind/name/start/end, JSDoc tag count, then each tag's
kind/name/start/end/full raw text, then parameter count and identifier names
(empty for binding patterns). Native rules interpret tags and inherited contracts;
the bridge does not recognize exemptions or produce findings.

Shape questions use the existing graph schema; `container-bases` and
`contextual-argument` retain named records, while `*-shape` records omit names.
`reference-shape` alone includes
deferred reference arguments, leaving earlier shape questions unchanged. An
absent annotated/contextual/symbol type is an absent graph. Argument indexes and
identities must be canonical and in range. `property-declarations` consumes the
remaining bytes as its name, including embedded LF. The C ownership checks cover
these facts through the native coverage suite under ASan/UBSan/LeakSanitizer;
Go heap accesses remain outside ASan's instrumentation.
