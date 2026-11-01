# 98-axolotl-b: required reads in es2015.ts

Five indexed reads in `src/compiler/transformers/es2015.ts` each gain a `!` at
their original evaluation point. Node erases it; Adamic checks it loudly. The
file goes from 5 own-file checker diagnostics to 0. Adaptation 33 recorded
these five occurrences as declines with unfinished invariant reviews; this head
finishes those reviews and asserts them. Nothing else changes: no read is
added, moved, hoisted, defaulted or turned into an optional chain.

[sites.json](sites.json) records each site: file, function, expression,
occurrence and total, the parent node kind, the class (as 30 and 33 use them),
the action and a one-line invariant. Line and column are documentation in
main's adapted tree, never edit offsets.

Run `node adapt.cjs <tree>` with `NODE_PATH` at stock typescript@6.0.3 (apply
supplies it). The adapter is 98-axolotl-a's, with only its site list changed:
it parses current text with stock 6.0.3 and, for each site, requires the
recorded occurrence count, the enclosing function and the parent node kind,
with or without the `!` already present. It refuses a doubled assertion, a site
that drifted or moved, and a tree where some sites are asserted and others are
not. It transpiles the file before and after with comments kept and removed and
refuses any JavaScript change, all before writing anything. Inserting one
character keeps the file's CRLF bytes. Its enclosing-function walk (nearest
named function declaration) is unchanged: each of the five sites sits directly
in a named function declaration, with no arrow, method or function expression
between, so it did not need extending.

## The invariants

- `node.declarations[0]` in visitVariableDeclarationList (2803:38),
  U-endpoint. The guard is `node.transformFlags &
  TransformFlags.ContainsBindingPattern`. createVariableDeclarationList adds no
  such flag of its own; the list's flags are its declarations NodeArray's
  aggregate, which is None for an empty array. Only the binding pattern
  factories set the flag, and it reaches the list through a declaration, so the
  list is nonempty. Upstream's own `last(node.declarations)` on the same line
  asserts the same. Dereferenced in place (`.name`), so absence already throws
  on Node.
- `declarations[0]` in convertForOfStatementHead (3046:64), U-endpoint. The
  array is flattenDestructuringBinding (transformers/destructuring.ts) of the
  first declaration, guarded by `isBindingPattern` on its name, with
  hoistTempVariables false. Both callers (convertForOfStatementForArray and
  convertForOfStatementForIterable) pass a bound value built as an element or
  property access, never an Identifier. ensureIdentifier reuses only an
  Identifier, and otherwise emits its temp through emitBindingOrAssignment,
  which records a declaration. An empty or multi-element pattern, an all-omitted
  array pattern, any array pattern under downlevelIteration, and any element
  with an initializer (createDefaultValueCheck) take such a temp; a
  single-element pattern recurses on its element with an access or helper-call
  value, never an Identifier, until an identifier target emits. So the output
  is nonempty. Upstream's `last(declarations)` on the same line asserts the
  same. Stock 6.0.3 transpiles `for (const [,] of xs)`, `[]`, `{}`, `[,,]` and
  `{a: [,]}` to ES5, with and without downlevelIteration, without throwing.
  Dereferenced in place (`.pos`), so absence already throws on Node.
- `funcStatements[classBodyStart]` in visitTypeScriptClassWrapper (4437:29),
  the second of two occurrences, U-position. funcStatements is the body of the
  down-level class IIFE that es2015 itself built for the class expression
  inside the ts transformer's wrapper: visitClassExpression creates the
  function expression with transformClassBody's block, and transformClassBody
  pushes an optional `__extends(C, _super);` expression statement
  (addExtendsHelperIfNeeded), then always the constructor function declaration
  (addConstructor), the members, and always the return statement;
  insertStatementsAfterStandardPrologue only adds hoisted declarations. So the
  body has at least two statements, and classBodyStart is 0 or 1 at this read.
  **This one is not a dereference.** The read is pushed straight into
  `statements`: if the contract broke, Node would push undefined and carry on,
  while Adamic stops loudly at the `!`. It is a loud check that changes
  behavior only on a contract break, not an erased one. The first occurrence
  (4430:41, inside `tryCast`) stays declined as 33 recorded it.
- `segments[0]` twice in transformAndSpreadElements (4691:34 `.kind`, 4693:13
  `.expression`), the second and third of three occurrences, U-endpoint. Every
  caller passes nonempty elements: visitArrayLiteralExpression under
  `some(node.elements, isSpreadElement)`;
  visitCallExpressionWithPotentialCapturedThisAssignment under
  `node.transformFlags & TransformFlags.ContainsRestOrSpread`, which only a
  SpreadElement argument can carry onto a call (createSpreadElement and a rest
  BindingElement are the only producers, and array literals, calls, news and
  binding patterns exclude it from their parents); and visitNewExpression,
  which prepends `void 0`. spanMap over a nonempty array yields at least one
  nonempty chunk, each mapped to a segment or a nonempty segment array, so
  flatten's result is nonempty. Both reads are dereferenced in place, so
  absence already throws on Node. The first occurrence (4676:34) was asserted
  by 33.

No site was declined. Every reading agreed with the invariant given.

## Before and after

Own-file diagnostics, counted the way the daily meter counts: the census tool
(`go build ./stage3/census/tool`) over each adapted tree's `src/compiler`, then
stage3/meter/report.py's `report()` over its JSONL, attributing each diagnostic
to its primary location and deduplicating repeated observations.

| File | Before | After |
| --- | ---: | ---: |
| transformers/es2015.ts | 5 | 0 |
| Own file, all of src/compiler | 58/79 | 59/79 |

The before tree is main's apply (the Nov 1 meter's 5 and its 58/79,
reproduced). Its five findings are TS2532 at 2803:38, 3046:64, 4691:34 and
4693:13, and TS2345 at 4437:29. No other file's count moved.

JavaScript: es2015.ts transpiled from the before and after trees with stock
6.0.3, comments kept and removed, is byte-identical (equal sha256 both ways).
The two trees' `src` differ in this one file, on these five lines.

Apply: `stage3/apply.sh` passes end to end with 98-axolotl-b in place; its row
is `98-axolotl-b | 1 | 5 | 5`. Adaptation 33, whose ledger still lists these
five as declines, runs before 98 and passes. A second run of `adapt.cjs` on
the applied tree reports 0 assertions and leaves the file's bytes unchanged.
Four probes are refused with nothing written: a partial apply (only 4437
asserted), a drift (an extra `segments[0]`), a moved parent (4437's read
parenthesized), and a doubled `!!` (refused by the parent-kind check, which
runs before the doubled-assertion check in 98-a's adapter).

Mutant: dropping the `!` from `funcStatements[classBodyStart]!` at 4437 raises
es2015.ts from 0 to 1 own-file diagnostic (TS2345 at 4437:29), no other file
moving. It ran on a scratch copy of the applied tree; the committed adapter is
unchanged.

lane: run as part of wave N (Loom).
