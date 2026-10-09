# TypeScript printer gaps

## Tsc-driver corpus repair

The 300 git-tracked inputs under `stage3/drivers/tsc/corpus` are now a named corpus root, not
incidental files found by a filesystem walk. [TSC_CORPUS.md](TSC_CORPUS.md) records the audit,
and [results/tsc-corpus-before.json](results/tsc-corpus-before.json) preserves every original
file-labelled port/Go disagreement: two NotYet results and five text differences, with no errors.

Four causes are closed: yield lookahead before literals, a leading block in standalone expression
text, bodyless untyped function declarations, and top-level ternary grouping in statement files.
The shared parser's yield predicate now mirrors TypeScript's identifier/keyword/literal lookahead
on the same line. It does not force a generator context; ordinary calls and newline boundaries
are independently held to Go in `stage1/typescript/parser/yield_test.go`.

The surprising `{}[0]` output is a standalone parsing boundary: Go reads a block followed by an
array expression. The expression driver follows that program interpretation rather than inventing
an object receiver. The original source's object context is not present in that extracted text.
The full audit also names any upstream parser or Prettier differences separately; they never
allow the port to differ from Go.

All 1,729 selected expression fragments and 874 selected statement fragments from this root
match Go on Node, sanitized native and the JavaScript backend. This is fragment coverage, not a
claim to format 300 complete files. Go's selector refuses the full file
`239_expressionWithJSDocTypeArguments/expressionWithJSDocTypeArguments.ts` with `Type expected.`;
the selectors retain that refusal in their coverage reports. Existing unsupported families below
remain explicit gaps, and no compiler or runtime implementation changed.

Status: **partly ported**. The shared document engine, expression core, sequence expressions, assignments, conditionals and object values are held to Go cohere and
Prettier 3.9.6. This is not a claim to port all of `internal/format/javascript`, or to format whole
TypeScript compiler files yet. The composition uses the indexed TypeScript parser, including the yield lookahead repair
recorded above. The compiler and runtime are unchanged.

## Expression work remaining

[testdata/notyet.json](testdata/notyet.json) is the runnable set of proving inputs. The Go overlay
formats every input, and upstream Prettier must accept them; their one known upstream difference is
recorded below. The composed driver, natively with leak detection, on Node and through the JavaScript backend, must return these exact reasons:

| Input | Reason | Work still needed |
|---|---|---|
| `(x: number) => x` | `function-types` | Typed signatures belong to the type-printing slice |
| `({method(x:number){return x;}})` | `function-types` | Typed method signatures |
| `[a,b]=items` | `assignment-pattern` | Destructuring target patterns and their assignment layout |
| `` test.each`a\|b\n${1}\|${2}` `` | `jest-template-table` | Jest column layout |
| `(function () {return 1;})` | `FunctionExpression` | Anonymous-function spacing differs upstream; named functions and supported bodies are ported |
| `f<T>(x)` | `type-arguments` | Type printing and type-argument layout |
| `a as T` | `AsExpression` | Cast/type composition |
| `a /* comment */ + b` | `comment-attachment` | Shared comment attachment and printer hooks |
| `a +` then a blank line then `b` | `source-trivia` | Preserve source-driven empty-line decisions |
| `a; b` | `expression-file` | Statement and document composition |
| String with a backslash-newline continuation | `string-literal-layout` | Multiline raw literal layout |
| `async x=>await (x as T)` | `AsExpression` | Await composes; cast/type syntax remains a type-printer gap |
| Numeric array with a whitespace-only blank line | `source-trivia` | Preserve empty lines containing spaces |
| Hashbang followed by `f()` | `comment-attachment` | Preserve program headers rather than silently dropping them |
| ``tag<T>`a${b}``` | `tag-type-arguments` | Type-argument printing |

Comments are conservatively detected by raw marker substrings. This also declines a marker inside
a string or regex; it never pretends to attach it. Blank lines, including whitespace-only lines, and CR source are likewise declined. Parentheses that stop an active optional chain are preserved, including in flat interpolation previews.
This core does not promise arbitrary multiline literal handling, JSX, anonymous function/class expressions,
Jest tagged tables, TypeScript assertion/satisfies expressions or whole-file syntax. Ordinary tagged templates and contextual await/yield are ported.
No accepted-corpus disagreement is put on an ignore list. The document generator covers conditional
groups and suffixes already, so the missing expression layouts can use them without another doc
engine port.

The first increment's corpus report records 60,602 unsupported/context/generated candidates.
Later increments keep separate coverage reports under `results/`. These
are overlapping AST visits, not disjoint program failures or a percentage of TypeScript ported.
Rejected parents are traversed for supported children. The report preserves their kinds; it does
not label their complete source files green. No whole-file formatter result is claimed.

## One upstream difference in an unported proving input

[testdata/prettier-differences.json](testdata/prettier-differences.json) records the exact pair:

```ts
(function () {return 1;})
```

Go cohere produces `(function() {\n    return 1;\n});\n`; npm Prettier 3.9.6 produces
`(function () {\n    return 1;\n});\n`. The difference is the space before the parameter list.
Both accept the input. The port returns `NotYet: FunctionExpression`, so this is not an accepted
formatter result. The independent comparison of the unported `gaps.json` corpus checks this exact
source and pair and fails if a new difference appears or this one changes. The accepted `cases.json`
comparison remains strict, with no exceptions. All 149,852 accepted fragments agree with both.

## Stage-0 gaps encountered

The following remaining gaps are rechecked against main b6b1538b. `TestCompilerGaps` executes each program on Node and requires stage 0's recorded diagnostic.

| Program | Node output | Stage 0 | Port treatment |
|---|---|---|---|
| [prefixUpdateValue.ts](gaps/prefixUpdateValue.ts) | `2` | Closed by compiler/area-stack (views slice 1); `TestClosedPrefixUpdateValueGap` holds native and the JavaScript backend to Node | Decrement as its own statement, then read the resulting index |
| [defaultSort.ts](gaps/defaultSort.ts) | `im` | `Refused`: sort without comparator | Supply an explicit lexical comparator for regex flags |

The last is an intentional 0.1 rule, not a compiler gap to relax. Logical assignment is intentionally
refused too; array classifiers write `false` when an element fails their predicate. Array `length`
writes are intentionally forbidden; the suffix queue is cleared with `splice(0)`.

The parser's existing `push` and indexed-node workarounds remain: appends use one value per call,
and node/doc child links are indexes, so no ownership cycle is introduced. An untyped empty array
inside `??` inferred as `never[]` was replaced by an explicit missing-value panic. A conditional
branch containing `panic` inside an argument was spelled as an explicit checked branch instead.
These are implementation accommodations, not rewrites of the formatted user's program.

## Layout representation

Go docs hold child docs directly and can share a group pointer. Here an append-only node table holds
numeric links; the same index is visited once during break propagation. A conditional group's
barrier to propagating breaks is preserved. Unknown kinds and forward/cyclic links fail loudly.
The Go printer's `settled` output optimization is omitted: chunks are trimmed and joined at the end,
with the same output bytes on the independent boundary/random corpus. Unicode width code is copied
from the established JSON slice; align-string indentation counts UTF-16 units, as Go's printer does.

Earlier increment sections below record the progression; the runnable gap table above and the newest
validation section describe current support.

## Boundary audit

The initial broad corpus missed six deliberately chosen boundaries. A separate Go/Node/native
probe showed a numeric receiver losing its parentheses (`1.toString`), an optional-chain stopping
boundary disappearing (`a?.b.c`), a spaced blank line disappearing from a numeric array, redundant
postfix-update parentheses, redundant prefix-update parentheses before `in`, and excess indentation
in `Boolean` coercion calls. Numeric receivers, update distinctions and coercion indentation are
now ported; optional boundaries and all forms of blank lines return NotYet. Regression inputs are
in the generated edge set and the unported proving set. Member lookup now carries the numeric
ancestor stack through non-null wrappers and computed-member parents, as cohere's Path does.

## Later slices

Complete statement composition, declarations and types are not implemented. Supported arrow block bodies already compose the basic statement layouts described below. Full-file comment attachment, source
normalization, embedded-language printers and non-default formatter options are not covered by this
expression driver. The first green increment is a reusable layout engine plus the expression core;
remaining expression families come before claiming the entire expression slice complete.

## Sequence normalization

Comma expressions are implemented. `(a,b),c` formats as `((a, b), c);`, while `a,b,c` formats as
`(a, b, c);`. These are observations of Go cohere and npm Prettier 3.9.6, covered by generated
proving inputs in `testdata/expressions_side_test.go`. Normalization records the parent's explicit
sequence boundary before unwrapping it; it never discards a boundary merely because the comma
operator is associative. There is no new stage-0 gap in this family.

## Assignment family

Simple and compound assignments are implemented for identifier, property, computed-property and
non-null targets, composing with every existing expression family. All 16 assignment operators
are data to the printer, including the logical assignments Adamic refuses in its own source.
Short and long assignment chains, fluid wrapping, breaking after the operator, `require` calls,
short call arguments and assignment-driven member layout follow Go cohere.

Destructuring targets remain an explicit pattern-printer gap: `[a,b]=items` is formatted by Go and
Prettier, while the port reports `NotYet: assignment-pattern`. The independent Go selector excludes
such targets, and the proving corpus checks the exact refusal on every execution. No source is
passed through unchanged.

## Conditional family

Nested conditional tests receive their own parenthesized group; nested consequents retain flat
parentheses and align their broken branches; alternate chains share the outer group. Nullish
coalescing branches keep the parentheses required by Go's printer. Generated proving inputs cover
all three nesting positions and composition with assignments, sequences, calls and member access.
This family introduces no new stage-0 gap.

## Object family

Property values, shorthand properties, computed names and object spreads are implemented. The
property value uses cohere's assignment strategies, including the short-key exception. Numeric
keys stay numeric in TypeScript, and escaped string keys stay quoted when the printed spelling
does not equal the decoded value. Unicode key tables come directly from Go's letter, letter-number,
digit, mark and connector categories; `testdata/generate_keys.go` regenerates them.

Parentheses enclosing an object literal can contain optional property values: the object itself
cannot be an optional chain. Parentheses around a chain still return the recorded stopping-boundary
gap. Object methods, accessors and binding/assignment patterns remain unported. No code is passed
through unchanged.

## Call arguments

Calls and constructors with a plain identifier callee support expanded object/array arguments.
Last-argument expansion uses cohere's conditional states; homogeneous adjacent argument kinds
and numeric arrays in multi-argument calls disable expansion as in Go. Multiline arguments break
outer groups when required. `require` and top-level `define` special layouts are ported. The
independent selector now accepts these argument shapes and all previously rejected seeded random
argument cases. Member-call chains, function/arrow arguments and type arguments remain gaps.

## Member-call family

Member calls, curried calls and arbitrary supported callees are implemented. Chains are segmented
into their base, member and call groups, with cohere's factory/short-receiver merge heuristic and
conditional expansion. Constructor callees containing a call retain required parentheses.
`tracing?.pop()` has only one optional token: the parser's inherited chain flag is not an additional
`?.`. Generated cases cover optional calls separately from optional member lookups.

The placeholder parentheses branch in Go's general printer is dormant for its TypeScript adapter:
`(PRETTIER_HTML_PLACEHOLDER_0_0_IN_JS)` normalizes to the bare identifier in both Go and npm Prettier.
The proving cases retain that observed behavior rather than copying a Babel-only condition.
Comments, typed calls and chain-stopping boundaries remain the recorded gaps.

## Template family

Interpolated and multiline untagged templates are implemented, including interpolation newlines
before the closing brace, absolute indentation inherited from raw quasis, and literal-line break
propagation. Long binary interpolations do not add a second indentation level: the template owns
that indentation. The generated corpus covers leading and trailing interpolation newlines separately.
Tagged templates remain a separate proving input. No new stage-0 gap was encountered.

## Arrow and basic body layouts

Untyped arrows now compose defaults/rest, nested chains, expanded call arguments and simple block
bodies. The body printer preserves directives and string parentheses, return/throw argument groups,
object-receiver parentheses and the numeric ancestor stack. Statement composition beyond these
bodies, binding patterns and typed signatures remain separate slices.

The module split first encountered the intentional refusal of import cycles, including type-only
cycles. The final layout uses pure syntax helpers instead of a callback contract. Two further observations on this branch's
scanner/parser base are recorded with standalone programs: these are not claims about current main.

| Program | Node | Stage 0/native observation | Port accommodation |
|---|---|---|---|
| [classInterfaceMethod.ts.txt](gaps/classInterfaceMethod.ts.txt) | `17` | Lowering and clang succeed; native exits 70, `compiler bug: a field the checker proved is there is missing` | Keep stateful methods on the concrete printer; the proving workaround constructs explicit function properties |
| [optionalBooleanFunction.ts.txt](gaps/optionalBooleanFunction.ts.txt) | `absent` | Closed by `f69bf6082b6db19ec0037ef8d2a51a9a5d46a8d8` | Required-parameter proving workaround removed; the final layout has no optional callback contract |

`TestClassInterfaceMethodGap` still requires its recorded failure and checks the
explicit-function-property workaround. `TestClosedOptionalBooleanFunctionGap`
copies the unchanged boolean program to scratch `.a` and requires `absent`
on source Node, native ASan/UBSan and the JavaScript backend, with a separate
LeakSanitizer check. The required-boolean proving workaround is retired.
The pure syntax-helper layout already removed the original callback contract;
there is no stage 1 implementation workaround left to undo for this gap. The `.txt` programs are a
separate execution-gap corpus, not formatted expression fragments.

The class/interface panic is a compiler implementation bug. A compiler adapter could retain class method metadata through
the interface view and dispatch `reader.read()` with its original receiver. Arbitrary method
extraction needs its own Node oracle. Main b6b1538b now returns NotYet during lowering instead of reaching native with a missing
proved field, as recorded in the merge-seat repair below. Origin-preserving dispatch remains
unimplemented; this port makes no internal change.

A `.some` callback capturing the layout class was also refused because the callback interface can
reach that captured type. A loop reads the async modifier without creating that function value.
No language or runtime rule is relaxed, and no internal file is changed.

## Optional-chain stopping boundaries

The former `(a?.b).c` proving gap is implemented. A numeric boundary set preserves active-chain
parentheses through normalization; nested already-stopped receivers do not become active chains.
Ordinary member/call links and non-null assertions retain stopping parentheses. Optional outer
links remove redundant parentheses without losing the original ChainExpression call-layout choice.
The independent Go selector no longer filters these boundaries. The replacement proving input
`await value` records `AwaitExpression`: await/yield parsing depends on expression context and is
not enabled by forcing an async/generator context onto every fragment.

## Named function expressions

Named ordinary, async and generator function expressions are implemented with the existing simple
body printer. Untyped return positions and call/constructor parentheses are preserved. The name is
recognized only before the parameter list; a later identifier return type is not mistaken for a
name. Typed signatures, binding patterns and await/yield bodies still return explicit gaps.
Anonymous function expressions retain their existing proving program and exact npm spacing report.

## Object methods and accessors

Untyped methods, getters and setters are implemented, including async/generator method prefixes,
numeric and quoted names, computed names, defaults/rest and supported body statements. The former
method proving gap is now `({method(x:number){return x;}})` / `function-types`; typed signatures
remain outside this slice. The blank-line predicate is a pure helper in `syntax.ts`.

The computed-short-key audit proved a missed layout: `[x]` and `[1]` with a long string or member
chain broke after the colon on both Node and the previous native release, while Go kept the value
with the key. The fix recognizes text after cleaning text/concat-only docs, without erasing groups
or other doc commands. Sixty-four generated boundaries cover identifier, numeric, unary, literal
and width-threshold keys in four contexts. The regression is an oracle case, not a declined input.

## Optional-chain interpolation composition correction

A targeted probe after the first method gate found that the flat interpolation preview copied
sequence boundaries but omitted optional-chain boundaries. Go preserves `(x?.y).z` in
`` `a${(x?.y).z}b` ``; both source Node and native had produced `` `a${x?.y.z}b` ``.
The preview now copies both sets. The generated corpus adds 48 combinations of four stopped
chains, three raw-line layouts and four expression contexts. A mutant removes that copy and must
finish normally on Node and native, then fail the byte comparison. This is a port correction,
not a compiler or runtime change.

## Tagged-template family

Ordinary tag layout is implemented. The generator adds 504 combinations of 12 tag shapes,
six template bodies and seven contexts. Source-driven template indentation uses the tag's start;
the quasi preserves its bytes. `test.each` and related `describe`/`it`, `only`/`skip` forms have a
separate column printer in Go and return `NotYet: jest-template-table` here. Its proving program
is in `testdata/notyet.json`. Typed tag arguments likewise remain a precise `tag-type-arguments` gap.

Embedded-language formatting is outside this printer. Go's template printer explicitly has no
embed hook; CSS, HTML, GraphQL and Markdown embedding would need independent upstream reports.
The accepted ordinary-tag corpus has no upstream exceptions.

## Await and yield family

Await/yield compose in async arrows, named async functions and generators. Receiver/callee await
layout, nested await, precedence and delegated versus bare yield follow Go. A bare yield has no
argument; a delegation token is not that argument. The scratch prototype initially panicked on
`x=yield`, so the final generator includes bare and delegated assignment cases. Another scratch
probe caught extra parentheses around `yield a+b`; yield and await arguments now have their
separate precedence rules. These corrected prototypes were not counted as green native runs.

Context belongs to the parsed source. The driver does not force every `await` or `yield` token into
expression syntax: call-like uses outside those contexts retain the parser's interpretation.
Typed await operands remain the explicit `AsExpression` proving gap. No compiler file changed.

## Variable statements and composition

Untyped identifier declarations now compose inside supported function bodies: var, let, const,
using and await using, multiple initialized or uninitialized declarations, and long initializers.
Binding patterns, annotated declarations and declaration modifiers remain loud gaps. The next
file driver will provide separate proving inputs for these declaration boundaries.

An independent complete-file preview found two layout differences in receiver/conditional
composition and five assignment chains whose expression-statement wrapper changed their layout.
Those inputs are retained in the expression corpus. The assignment-chain boundary mutant proves
that the layout distinction matters. A hashbang audit also found a real silent header loss in
both Node and native execution; hashbang and BOM-plus-hashbang are now refused, with a mutant
that removes that refusal and compares the proving corpus rather than accepted cases.

## Complete-file statement boundary

The statement driver composes supported program sequences and named untyped function declarations.
Five complete repository files pass its independent selector; this does not include an entire
TypeScript compiler file. Remaining statements include if/else, loops, labels, try/catch/finally
and switch. Imports/exports and typed declarations belong to later declaration/type families.
[testdata/statements-notyet.json](testdata/statements-notyet.json) records the five exact refusal
inputs used by the statement gate on Node, native and the JavaScript backend: annotated variables,
binding patterns, modifiers, if-statements and hashbang headers.

There is a real upstream contract boundary for control-flow statements: the pinned Go printer's
`print_statements.go` deliberately prints `if(`, `while(`, `for(`, `switch(` and `catch(`, and places
else/catch/finally on separate lines. npm Prettier 3.9.6 uses spaces and cuddles block boundaries.
These are cohere customizations, not a compiler gap. Before adding the tsc-driver corpus, the accepted expression and statement
corpora had no upstream exceptions. The new corpus pins its external outcomes separately in
[testdata/tsc-upstream-differences.json](testdata/tsc-upstream-differences.json); the port still
matches Go without exceptions. A future control-flow family must report these exact
upstream differences while retaining Go cohere as the primary contract, as previously decided.

The five fork/npm pairs are pinned in
[testdata/statement-upstream-differences.json](testdata/statement-upstream-differences.json).
`TestStatementUpstreamDifferences` verifies Go and embedded-fork bytes against each Go field,
and independently verifies npm bytes against each Prettier field. It does not grant exceptions
to the supported expression or statement corpus.

## Merge-seat repair against main b6b1538b

The Number constructor gap is closed. Its program moved to
[testdata/numberConstructor.ts](testdata/numberConstructor.ts) and now runs successfully on
Node, native and the JavaScript backend. Protocol fields and CLI widths use Number again.
The independent Go document protocol now encodes widths in hexadecimal: Number parses them
as intended, while parseFloat would silently produce zero. Ten explicit Number-call inputs
join the strict Go, embedded Prettier and npm Prettier expression corpus.

The class/interface gap moved from a native missing-field panic to a compile-time NotYet:
`gap.ts:6:28: stage 0 can't lower a class method through a view that erases its prototype origin yet`.
The original proving program is unchanged. The test checks the typed diagnostic, location and
reason before C emission. This closes the bad native path, but does not implement dispatch through
this property-style interface. Preserving the concrete method's receiver through the interface
remains a compiler implementation gap for @system_adamic; no language change is proposed here.

The working callback proof now names the concrete method readValue, exposing read as an explicit
function property. Main's origin guard conservatively considers compatible class shapes in the
module; renaming avoids confusing that concrete method with the interface's read slot. The
callback still returns the same captured Box value. Successful lowering rejects the gap-check
mutant, and Node, native and backend must all print 17 without leaks.

macOS leak-check integration is waiting for internal/leakcheck to land on main from
origin/devtools/stage1-leaks (f6eef5df). This branch does not merge that development branch or
replace the helper with local ASan option switching. Named oracle skips now cite
#xq2ecw6 (setup --gate-inputs), the work that installs the pins and removes the skips.

## Optional boolean closure validation

On compiler/area-gaps, `go test -v ./stage1/cohere/tsprinter -run
'^(TestClosedOptionalBooleanFunctionGap|TestClassInterfaceMethodGap)$' -count=1
-timeout 30m` passed (2.291s). The closed fixture agrees with Node in both backends
and is leak-free. No implementation callback workaround remains in the final
printer layout; the obsolete required-boolean proving workaround was removed
from the gap test. No broader printer suite was run for this test-only retirement.
Changing the fixture call from `show(undefined)` to `show(true)` made the closed
regression fail on Node's normal `present` output; the fixture was restored.
The existing class-interface successful-lowering mutant was also rejected.
