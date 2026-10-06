# Expression parser gaps and coverage

The port follows `cohere/TypeScript/tsc/internal/parser/parser.go`, using the
existing Adamic scanner. Node runs these same TS files; an independently built
Go overlay runs the unmodified typescript-go parser.

## 1. Strong AST parent and child pointers are cycle-capable

Go's `finishNode` gives every child a strong Parent pointer. Go's tracing
collector can collect the resulting cycles; Adamic's reference counting cannot.
`gaps/1_strong_ast_parent.ts` prints `root` on Node. Stage 0 refuses the push:

```
Adamic 0.1 refuses Tree[], an array whose elements can reach back to an array like it: a cycle reference counting can't free
```

`TestStrongAstParentGap` requires the Node answer and the `lower.Refused`
cycle diagnostic. This is a deliberate language refusal, not a silent compile
or a claim that immutable trees are forbidden.

Workaround: a parse owns an indexed node table. Every node stores child indexes;
no parent pointer or child pointer owns another node. Traversal starts at root
indexes. The representation accommodates speculative parse rollback as well.

## 2. Spread into Array.push

`gaps/2_push_spread.ts` prints `1,2,3` on Node; stage 0 refuses
`a SpreadElement` at the spread argument of push. The parser initially
appended an argument list with `children.push(...arguments)`, which met this
same refusal. Workaround: an explicit loop appends each numeric child index.
This does not limit parsing spread syntax in the source being parsed.

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
