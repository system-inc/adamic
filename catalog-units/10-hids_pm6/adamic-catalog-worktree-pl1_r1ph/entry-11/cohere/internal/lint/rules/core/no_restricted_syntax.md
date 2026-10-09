# `no-restricted-syntax`

| | |
|---|---|
| **Recommendation** | **No — not portable to this tree** |
| Violations in ahra | none, and structurally so |
| Plugin | `eslint core` |
| Auto-fixable | no |
| Needs type information | no |
| Status | measured and declined, 2026-09-06 |

## What it checks

Disallow specified syntax, where "syntax" is an [esquery](https://github.com/estools/esquery)
selector: a CSS-like query language over an **ESTree** syntax tree.

## The decision

**Not ported.** Two sibling rules were ported in the same pass (`no-restricted-exports`,
`no-restricted-imports`); this one was measured instead, and the measurement is below.

The earlier audit recommended Yes on the strength of a zero violation count. That count is real and
it is not evidence: this rule does nothing at all until somebody configures a selector, so an
unconfigured rule and a clean tree produce the identical number. The same was true of both siblings,
and each of their doc comments says so.

## Why it is not portable

The rule itself is ten lines. Its whole body is:

```js
return context.options.reduce((result, selectorOrObject) => {
    ...
    return Object.assign(result, {
        [selector](node) { context.report({node, messageId: "restrictedSyntax", data: {message}}); },
    });
}, {});
```

Every configured selector becomes a listener key, and ESLint's own listener dispatch resolves it
through esquery. So porting the rule means porting esquery, and esquery is not a table of node
names: it is a query engine over a tree this repository does not have.

### 1. There is no selector engine here, and it is not small

`internal/lint/ecmascript/` has no esquery equivalent, and neither does anything else in the tree —
`grep -rn esquery --include='*.go'` outside the vendored compiler returns nothing. The installed
engine is esquery 1.7.0, 4,426 lines of generated parser plus matcher.

Upstream's own 34-case corpus reaches 22 distinct selectors, and between them they exercise
essentially the whole grammar, so a partial engine would not cover the corpus:

```
*  ~  *                                              sibling combinator
:is(Identifier[name='foo'], Identifier[name='bar'])  matches-any pseudo-class
:nth-child(1)                                        positional pseudo-class
ArrowFunctionExpression > BlockStatement             child combinator
FunctionDeclaration[params.length>2]                 numeric comparison on a dotted field path
ImportDeclaration[source.value=/^some\/path$/]       regex attribute match on a dotted path
Literal[regex.flags=/./]                             regex match on a nested field
VariableDeclaration[kind='using']                    attribute equality
BreakStatement[label]                                attribute presence
[optional=true]                                      bare attribute, no node type at all
Property > Literal.key                               field-position qualifier
```

### 2. The deeper problem: a selector names ESTree, and this tree is not ESTree

This is why a name-translation table cannot rescue it, and it was measured rather than argued.

Take upstream's own corpus selector `VariableDeclaration[kind='using']` against the source
`const a = () => { foo(); }; let b = 1; using c = d;`.

Through the installed esquery 1.7.0 over the typescript-eslint parser's ESTree:

```
VariableDeclaration                  3 nodes, each carrying a `kind` field
VariableDeclaration[kind='using']    matches exactly 1
```

The same source through this repository's AST, measured with a walk counting node kinds:

```
KindVariableStatement          3     the statement
KindVariableDeclarationList    3     where const/let/using lives, as a node flag
KindVariableDeclaration        3     the binding, which has no kind of its own
```

`VariableDeclaration` is not a name to translate. ESTree's one node is three nodes here, and the
`kind` the selector filters on sits on a **different node** from the one the selector selects. Every
selector in upstream's corpus that filters a declaration is in this class, and so is any selector a
project would plausibly write about `const` versus `let`.

Making selectors work would therefore mean projecting the typescript-go tree into an ESTree-shaped
one — node types, field names, field positions and all — and then running a selector engine over the
projection. That is a substrate, not a rule port. It is roughly the surface of `@typescript-eslint`'s
own AST converter, and it would have to stay faithful as the vendored compiler moves.

### 3. Nobody here is asking for it

The name does not appear in `CohereSettings.json` under any spelling, and it is not configured in any
ESLint config across the repositories checked. There is no standing decision to honour and no user
waiting on it.

## What would change this

A project that actually wants selector-based restrictions, plus an ESTree projection built for its
own reasons — most likely as part of some other compatibility work. Until then, a project wanting to
ban one specific shape is better served by a small named rule that says what it means, which is what
every other rule in this directory is.

## Reproducing the measurement

Both halves were taken with instruments, not read off documentation. The ESTree half:

```bash
cd /Users/kirkouimet/Projects/ahra
node -e "
const parser = require('@typescript-eslint/parser');
const esquery = require('esquery');
const source = 'const a = () => { foo(); }; let b = 1; using c = d;';
const ast = parser.parse(source, {sourceType:'module', ecmaVersion:2022, range:true, loc:true});
for (const selector of ['VariableDeclaration', \"VariableDeclaration[kind='using']\"]) {
    console.log(selector, esquery.match(ast, esquery.parse(selector)).length);
}
"
```

The cohere half is a walk over the same source counting `node.Kind`, run from a throwaway probe
package under `internal/`. The probe was deleted after the run, as the porting standard requires;
its numbers are the three lines quoted above, and a control assertion inside it confirmed the walk
had actually visited nodes before any of them were believed.
