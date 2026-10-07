# TypeScript parser gaps and coverage

The port follows `cohere/TypeScript/tsc/internal/parser/parser.go`, using the
existing Adamic scanner. Node runs these same TS files; an independently built
Go overlay runs the unmodified typescript-go parser.

The green-step and selected-node measurement sections retain the expression
slice's history. Current whole-file scope appears in the final continuation
section and [WHOLE_REPORT.md](WHOLE_REPORT.md).

The JSX continuation is now in [JSX_REPORT.md](JSX_REPORT.md). The historical
JSX exclusions below describe the earlier slices. Current TSX and JavaScript
mode build all thirteen JSX AST kinds and compare their trees with Go, then
run the batch-8 React listeners directly. General diagnostic/recovery parity,
binder fields and full JSX application-code compilation remain outside scope.

## 1. Strong AST parent and child pointers

Go's `finishNode` gives every child a strong Parent pointer. Graph regions now
accept the proven closed object/array cycle in `gaps/1_strong_ast_parent.ts`.
`TestStrongAstParentSupported` compares its `root` output with Node in release
and ASan/UBSan builds. The former reference-counting cycle refusal is obsolete
for this fixture; open or escaping cyclic shapes still require the compiler's
ownership proof.

The parser keeps its indexed node table: every node stores child indexes, and
traversal starts at root indexes. This representation accommodates speculative
parse rollback.

## 2. Spread into Array.push

`gaps/2_push_spread.ts` prints `1,2,3` on Node; stage 0 refuses
`a SpreadElement` at the spread argument of push. The parser initially
appended an argument list with `children.push(...arguments)`, which met this
same refusal. Workaround: an explicit loop appends each numeric child index.
This does not limit parsing spread syntax in the source being parsed.

## 3. Type-only import cycles

Splitting lookahead into its own module first imported the Parser class as a
TypeScript type. Node accepts that cycle, since the type import is erased.
Adamic 0.1 refuses an import cycle, including a type-only edge.
`gaps/3_import_cycle.ts` and its peer print `1` on Node; stage 0 returns
`lower.Refused` with `What: "an import cycle"`. `TestTypeImportCycleGap`
requires both observations. Workaround: the lookahead module owns the scanner
snapshot interface and takes the concrete Scanner class. Parser imports that
interface and passes its scanner; lookahead does not import Parser.

## 4. Class methods through a structural interface

The first shared cursor interface exposed `kind`, `next`, `mark` and `rewind`
as methods. Node called them normally, but native exited 70 with
`compiler bug: a field the checker proved is there is missing`. Changing the
interface method syntax to strict function properties does not fix dispatch.
`gaps/4_class_interface_method.ts` reduces this to a class method exposed as
an interface's readonly function property. The test entered integration 14 in
`6e21425` from the parser branch; `1789dfc` changed it to require native agreement
after integration 13 preserved constructed class method tables.

The newer user-iterator rule `b5fbec4`, merged into integration 15 by `9804f60`,
keeps method calls whose prototype dispatch is represented, but explicitly
rejects function-property views that erase the class method's origin. This
fixture remains such a gap: Node prints `1`, while Adamic returns `lower.NotYet`
with `What: "a class method through a view that erases its prototype origin"`.
`TestClassMethodInterfaceGap` requires both observations and fails when the gap
closes. The compiler is unchanged. The parser keeps its concrete Scanner and
scanner-only save/restore helpers.

## Canonical answer protocol

Each expression root starts with `expression`. Nodes follow in preorder, one
per line: depth, kind, byte position, byte end, optional-chain flag, literal
flags, argument/element/property count (-1 where absent), trailing comma,
multiline, a tab, unary/meta operator kind, a tab, cooked text, a tab, template raw text. Children are in
Go's ForEachChild order. All punctuation children retained by that visitor
are included, notably binary operator tokens and conditional question/colon.

Positions include leading trivia, as Go does, rather than scanner token start.
Inside the port they use UTF-16; printing converts to source byte positions.
Text uses the scanner slice's ASCII and escaped UTF-16 output convention.
Node flags other than OptionalChain are currently excluded. These include
parser context, JSDoc, recovery and binder-only flags. Literal flags are retained.

## Green step 1

Primary identifiers, keywords and scalar/regex literals, parentheses, prefix
and postfix unary, binary precedence including right-associative exponentiation,
assignment, comma and conditional expressions. Only expression statements and
empty statements are accepted by the file driver at this step. Unsupported
syntax stops explicitly. Fourteen generated expressions compare byte for byte
on Go, Node and sanitized native. A precedence mutant must successfully run
and disagree under Node and native.

Not yet covered at this step: compiler file statements, arrow functions,
member/call/new, arrays/objects/spread, optional chains, template substitutions,
type assertions/satisfies, parse diagnostics/recovery, and non-UTF-8 source.
These are grammar coverage limits, not language gaps.

## Green step 2

Adds call/member/element/new, new.target and import.meta, parser-directed regex
and template rescans, arrays and omitted elements, object properties/shorthand/
computed names, spread, optional-chain propagation, non-null expressions,
and cooked plus raw template parts. Thirty generated expressions include
chained optional calls/accesses, chain boundaries at parentheses, non-null
propagation, nested new, trailing commas, and Unicode/escape templates.
The lost optional-chain flag mutant is caught under Node and native.

The file driver still accepts only expression and empty statements at this step.

## Green step 3

Adds simple, parenthesized, async and generic arrows with parameter annotations,
default/rest/destructured parameters and return types, functions and generator
bodies, object methods/accessors, type assertions/as/satisfies, type references
and arguments, unions/intersections, readonly, indexed/array/tuple/function and
conditional types. Basic block/return/throw/variable/if/while statements let
these expression bodies carry their full trees. Fifty-eight generated inputs
agree under Go, Node and sanitized native. All three requested tree mutants
run successfully and disagree; the parenthesis mutant classifies an ordinary
parenthesized expression as ArrowFunction in its returned tree.

The full compiler-file corpus and its declaration/statement traversal are still
pending at this step. Recovery diagnostics and type grammar beyond the generated
cases remain outside this green step's coverage.

## Green step 4

The driver now parses whole files, including declarations, class expressions,
functions and their signatures, loops, switch, try/catch/finally, labels,
bindings, enums, namespaces, interface/type declarations, mapped and template
literal types, constructor types and infer. It compares maximal expression
roots in source traversal order. Every descendant of each selected root is
included, so arrow/function/class expression bodies retain their statements
and types. Top-level statement and declaration trees are traversal scaffolding,
not the claimed comparison target. Static import/export declarations are
structurally skipped; their binding and module-specifier fields are not checked.

All 77 `.ts` files recursively under TypeScript 6.0.3 `src/compiler` produce
28,812,163 identical canonical bytes on Go, Node and ASan/UBSan/LeakSanitizer
native. The checkout is scratch-only and its pinned commit is checked by the
test. Go's expression predicate includes boolean/null and negative literal
types as well as type-query operands; these roots are retained.

Only valid, diagnostic-free TS input is covered. The Go oracle rejects parse
diagnostics; this slice does not promise recovery trees or diagnostic parity.
JSX, decorators, JSDoc trees and non-UTF-8 source remain outside coverage.

## Green step 5

Adds an exhaustive 25 by 25 binary operator corpus, sixteen assignment forms,
1,000 recursively generated expressions using seed 720, and targeted regressions.
The 1,676 generated inputs produce 1,432,520 identical tree bytes. The original
58 inputs still hold all three mutants, and the whole compiler corpus remains
identical at 28,836,875 bytes. The additional bytes come from retaining raw text
for no-substitution templates, which matters for tags and escaped spellings.

Generated cases exercised ambiguities missing from the compiler corpus: regex
and template rescans during arrow/type lookahead, parenthesized conditionals
versus typed arrows, contextual arrow names, generic optional calls/tags,
relational `>=` versus a closing generic `>`, omitted bindings, async parameter
defaults, and `in` contexts. Go assigns `??` the same precedence as `||`; the
operator-pair corpus caught the initial lower rank even though compiler sources
never exposed the difference. Random nesting uses the first ten operators;
all twenty-five operators are covered independently in the pair matrix.
Every input must parse without a diagnostic in the Go oracle. Generated inputs
are not filtered after comparison failures.

## Measured count-only parsing

All implementations count 557,010 selected-tree nodes from the same 77 files.
Best of five after one warm-up: Go 3,636,623 nodes/s, Node 981,532 nodes/s,
native 803,210 nodes/s. Native takes 4.528 times Go's time and 1.222 times
Node's. Linux amd64, EPYC 9V74, nproc 5, four-core cgroup quota, 16 GiB cgroup
memory limit; host one-minute load 1.81 to 1.88. Native uses stage 0's normal
clang -O2 flags without sanitizers. Timings include startup, reads, scanning,
parsing and tree counting, but exclude printing, byte mapping and build time.
All samples print the same count. See `validation/performance.log` and the
opt-in `TestPerformance` for exact samples and reproduction. This comparison
is a whole-file parser workload with equivalent selected-node counts, not a
claim of identical AST representation, parser recovery or declaration work.

`TestNodeCountCheckCatchesMutant` separately proves the timing count guard can
fail. It changes only countTree's per-node contribution from 1 to 0. The mutant
finishes successfully on Node and sanitized native and prints exactly the same
AST bytes as Go in tree mode, but reports 0 nodes instead of Go's 18 in count
mode. Only the count check catches it. See `validation/count-mutant.log`.

## 5. Function values with optional parameters

Moving declarations into a separate module exposed another stage-0 limit:
function values with optional parameters report `NotYet: a function value with
an optional parameter`. Default parameters on ordinary class methods are
already supported, but default parameters on the callback arrows were not.
`gaps/5_optional_function_value.ts` prints `1` on Node and is refused with this
exact NotYet; `TestOptionalFunctionValueGap` requires both observations.
The statement callback interface now requires every argument. Small class
method wrappers supply the defaults explicitly. Callbacks are temporary, never
stored on Parser, so their references back to Parser do not form an owning
cycle. They are explicit function properties, avoiding gap 4's class-method
structural dispatch bug. The original expression corpus remains identical.

## 6. Conditional branches with an unannotated empty array

The whole-file driver initially used `const types = docTypes ? parser.docTypes() : []`.
Stage 0 refused the empty branch with `an array of never`. The reduced
`gaps/6_conditional_empty_array.ts` prints `1` on Node, while lowering returns
`lower.NotYet` with precisely that `What`. `TestConditionalEmptyArrayGap`
requires both observations. Workaround: annotate the conditional result as
`number[]`; the empty branch then receives an element type.

## Whole-file continuation

The current driver additionally accepts `--whole`. It compares the SourceFile,
all statements and declarations, their expression and type children, and the
end-of-file token. Each file root starts with `file`. The expression protocol
above remains unchanged. Whole-file lines append a tab and a semantic field:
variable declaration lists carry the `BlockScoped` bits (let, const, using,
await using); import clauses carry their phase token kind; import/export
specifiers and declarations carry `IsTypeOnly`; export assignments carry
`IsExportEquals`. Module declaration keyword and import attribute token kind
use the operator field. Import attributes additionally carry element count,
trailing comma and multiline fields. These fields prevent identical child
sequences from hiding different declaration meanings.

`TestWholeCompilerAgrees` compares every compiler file as a whole tree.
Generated files cover every requested declaration and statement family,
including with, resource declarations, decorators, parameter decorators,
static blocks, accessors, index signatures, dotted namespaces, ambient modules,
import equals, contextual type/as, type-only and deferred imports, namespace
exports and attributes. Unicode comment prefixes and CRLF versions verify
position conversion. These are syntax probes, not checker acceptance claims.
The Go oracle refuses source parse diagnostics. Deprecated `assert` import
attributes trigger Go diagnostic 2880 and are excluded from this clean-source
corpus; a dedicated obsolete-assertion test compares their trees with an
explicit oracle option requiring that the only Go diagnostics are 2880. Current
`with` attributes are compared without that option. Error recovery, diagnostics,
JSX, full JSDoc comment/tag trees and semantic checking remain outside this
slice. SourceFile metadata, context/transform flags, parent links, and list
range/trailing metadata beyond the explicitly printed fields are not canonical
fields. Agreement concerns the documented canonical answer, not serialization
of every Go AST field.

`TestEveryTypeNodeKindAgrees` gets its inventory from Go's own
`ast.IsTypeNodeKind`, rather than a handwritten list: all 42 accepted kinds
must occur in the generated oracle trees. This includes the 24 ordinary type
node kinds, twelve keyword types, ExpressionWithTypeArguments and five JSDoc
wrappers. The optional and variadic wrappers occur only in JSDoc type grammar;
`--doc-types` compares their types from reduced `/** @type {...} */` annotations
through Go's lazy JSDoc parser. The Adamic driver extracts those reduced
annotations and reuses its type parser. This mode does not claim a JSDoc tag
or comment parser. Ordinary whole-file traversal follows Go's ForEachChild,
which omits attached JSDoc comments.
