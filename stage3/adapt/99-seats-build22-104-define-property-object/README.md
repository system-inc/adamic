# Define-property third argument as an expression

Adamic accepts this change after review, since it changes JavaScript. It is the
`isBindableObjectDefinePropertyCall` rider of her ruling 3 on the scout pass
(#bbssdcy, Nov 24 2026), as she answered it at 18:44:43Z: the retype to
`Expression` with an `isObjectLiteralExpression` check at each of the four reads,
the predicate untouched, and every input taking stock tsc's path.

`BindableObjectDefinePropertyCall` in `src/compiler/types.ts` types its third
argument `ObjectLiteralExpression`, but `isBindableObjectDefinePropertyCall` in
`utilities.ts` never tests it, so the type claims what no check proves. The
adapter retypes that tuple element to `Expression`, which every call argument
is, and checks `isObjectLiteralExpression` at the four reads that need an object
literal:

- `getAssignedExpandoInitializer` (`utilities.ts`) passes it to
  `hasExpandoValueProperty`, which iterates its `properties`. On any other
  argument `result` is `undefined`, as stock's `forEach(undefined, ...)` returns,
  and the function falls through to its end.
- `bindPotentiallyNewExpandoMemberToNamespace` (`binder.ts`) asks twice whether a
  property is named `set` or `get`. On any other argument each test is `false`,
  as stock's `some(undefined, ...)` returns, so neither adds accessor flags.
- `makeSerializePropertySymbol` (`checker.ts`) searches each declaration for a
  property named `set`. On any other argument that declaration gives
  `undefined`, as stock's inner `forEach(undefined, ...)` returns, so the outer
  `forEach` goes on to the next declaration and `Debug.assert(!!setter)` sees what
  it sees in stock.

`forEach` and `some` in `core.ts` test `array !== undefined` and return
`undefined` and `false` when it is, and `.properties` reads `undefined` on every
node but two: `ObjectLiteralExpression` and `JsxAttributes` are the only nodes in
`types.ts` that declare `properties` and extend `Expression`, and the parser
makes a `JsxAttributes` only as a JSX opening element's attributes, never as a
call's argument. So each check skips only what stock iterated as nothing, and no
check throws or returns where stock didn't. Each check sits at its own read; the
`prototype` comparison in `getAssignedExpandoInitializer` is a property read
and comparison that only feeds `hasExpandoValueProperty`, so skipping it on a
non-literal changes nothing observable. The predicate is untouched, so its
answer is the same on every input, and the two `checkExpressionCached` reads in
`getInitializerTypeFromAssignmentDeclaration` and
`isReadonlyAssignmentDeclaration` take an `Expression` already and are
untouched.

The adapter finds each site by its nodes (the type alias, the three functions
by name, and the reads by their callee and argument) and pins each by its whole
text before and after, never by a line or a digest. It requires each of the
three files to import `isObjectLiteralExpression` already. It refuses a tree
where some sites are adapted and others aren't, and changes nothing on a tree
already adapted. On the gate map (v5), `adapter-structural-pins` holds, since
it pins by nodes and text with no digest; `adapter-output-test` holds, since
its test reads the adapter's own snapshots and runs the adapted output beside
stock on a tree `stage3/apply.sh` wrote; and `adapter-table-mentions` doesn't
apply, since the change rests on no table tsc declares.

`test.cjs` reads the adapter's own `-before` and `-after` snapshots from a tree
`stage3/apply.sh` wrote. It checks that the adapter on `-before` writes `-after`
byte for byte and then changes nothing, that `-after` differs from `-before`
only by the retype and the four checks, that the predicate is byte-identical
and holds no `isObjectLiteralExpression`, and that both `checkExpressionCached`
reads are untouched. It then runs stock tsc 6.0.3 (`stage3/api`'s bundle) beside
the same bundle carrying the adapted functions, the house's way of running stock
beside adapted code (`99-seats-build22-41-pragma-arg-regexes`). It replaces only
the predicate and the smallest function holding each read,
`getAssignedExpandoInitializer`, `bindPotentiallyNewExpandoMemberToNamespace`
and the setter search's arrow, compiled from `-after`: the rest of the bundle
stays stock, and
`makeSerializePropertySymbol` itself calls a helper an earlier adaptation
added, which the stock bundle doesn't hold. A control bundle carrying the same
functions from `-before` must match stock too, so the splice itself is shown
neutral. On JavaScript programs under `allowJs`, `checkJs` and `declaration`,
one whose define-property calls take an identifier and one whose calls take an
object literal, it asserts the same diagnostics (code, position and text), the
same emitted `.js` and `.d.ts`, the same property symbol (flags, declarations
and type) and the same expanded declaration a hover prints. A third bundle
counts what each of the four checks answered, kept in memory for the run and
never written, and the test requires the identifier program to reach every
read with a non-literal and the literal program to reach every read with a
literal.

The identifier program reaches the first three reads with any define-property
call on an expando object, function or prototype. The setter search runs only
for a property symbol with a set accessor flag whose declarations list a
define-property call before the declaration that holds the setter. One symbol
table can't hold that: a set literal declared after an identifier conflicts
with the property it finds there, since the binder mixes `Property` into every
define-property symbol. Two script files can, since the checker merges their
global symbols: `a.js` defines `p` on `F.prototype` from an identifier, `b.js`
defines it from a literal with `set`, and the expanded declaration of `F`, as a
hover prints it, serializes the merged class with `p`'s declarations in that
order.

```sh
NODE_PATH="$PWD/stage3/api/node_modules" node stage3/adapt/99-seats-build22-104-define-property-object/test.cjs <applied tree>
```
