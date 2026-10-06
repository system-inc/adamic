# TypeScript printer gaps

Status: **partly ported**. The shared document engine and expression core are held to Go cohere and
Prettier 3.9.6. This is not a claim to port all of `internal/format/javascript`, or to format whole
TypeScript compiler files yet. The composition uses the existing indexed TypeScript parser without
editing it, the compiler or the runtime.

## Expression work remaining

[testdata/notyet.json](testdata/notyet.json) is the runnable set of proving inputs. The Go overlay
formats every input, and upstream Prettier must accept them; their one known upstream difference is
recorded below. The composed driver, natively with leak detection, on Node and through the JavaScript backend, must return these exact reasons:

| Input | Reason | Work still needed |
|---|---|---|
| `x => x + 1` | `ArrowFunction` | Parameter and arrow-chain layout; block bodies need statements |
| `({x:1})` | `ObjectLiteralExpression` | Properties, methods, comments and object wrapping |
| `a=b` | `EqualsToken` | Assignment strategies and chains |
| `a?b:c` | `ConditionalExpression` | Nested ternary layout and conditional indentation |
| `` `a${b}` `` | `TemplateExpression` | Substitution alignment, multiline strings and embeds |
| `(function () {return 1;})` | `FunctionExpression` | Function parameters and statement bodies |
| `a.b()` | `call-callee-layout` | Member-call chain segmentation and grouping |
| `f([1,2])` | `expanded-call-argument` | Expanded and conditional argument-group states |
| `f<T>(x)` | `type-arguments` | Type printing and type-argument layout |
| `a as T` | `AsExpression` | Cast/type composition |
| `a /* comment */ + b` | `comment-attachment` | Shared comment attachment and printer hooks |
| `a +` then a blank line then `b` | `source-trivia` | Preserve source-driven empty-line decisions |
| `a; b` | `expression-file` | Statement and document composition |
| Multiline template without substitutions | `template-literal-layout` | Literal-line and indentation alignment |
| String with a backslash-newline continuation | `string-literal-layout` | Multiline raw literal layout |
| `(a?.b).c` | `optional-chain-parentheses` | Preserve optional-chain stopping boundaries |
| Numeric array with a whitespace-only blank line | `source-trivia` | Preserve empty lines containing spaces |

Comments are conservatively detected by raw marker substrings. This also declines a marker inside
a string or regex; it never pretends to attach it. Blank lines, including whitespace-only lines, and CR source are likewise declined. Parenthesized
subtrees containing optional chains are conservatively declined; removing those parentheses can
change whether a later property access throws.
This core does not promise arbitrary multiline literal handling, JSX, function/class expressions,
tagged templates, yield/await, TypeScript assertion/satisfies expressions or whole-file syntax.
No accepted-corpus disagreement is put on an ignore list. The document generator covers conditional
groups and suffixes already, so the missing expression layouts can use them without another doc
engine port.

The corpus report records 60,602 unsupported/context/generated candidates at this revision. These
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

The following are observations on the scanner/parser base of this branch, not claims about newer
main. `TestCompilerGaps` executes each program on Node and requires stage 0's recorded diagnostic.

| Program | Node output | Stage 0 | Port treatment |
|---|---|---|---|
| [numberConstructor.ts](gaps/numberConstructor.ts) | `17` | `NotYet`: reading `Number` | Protocol fields are numeric text already; use `Number.parseFloat`, not a general substitute for `Number` |
| [prefixUpdateValue.ts](gaps/prefixUpdateValue.ts) | `2` | `NotYet`: numeric `PrefixUnaryExpression` | Decrement as its own statement, then read the resulting index |
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

Statements, declarations and types are not implemented. Full-file comment attachment, source
normalization, embedded-language printers and non-default formatter options are not covered by this
expression driver. The first green increment is a reusable layout engine plus the expression core;
remaining expression families come before claiming the entire expression slice complete.
