# Fixture search notes

These exploratory runs preceded the final retained fixture census. They are not
added to its counts. Runtime code was changed only in memory by observe.cjs.
No instrumented compiler was committed.

1. Literal duplicate assignments (`exports.value = 1; exports.value = 'two'`)
   do not ask getTypeOfAlias for the duplicated export. Importing identifier
   assignments, retained in commonjs.js.txt and common-consumer.ts.txt, does.
2. The first API walk used only forEachChild and missed JSDoc. Adding explicit
   node.jsDoc traversal exposed the eight excess-argument absences. CommonJS
   module and Node10 resolution were then fixed explicitly in the host.
3. Index types, unions, intersections, mapped types, tuples and a generic
   constrained to an indexed interface are retained in index.ts.txt. Their
   symbol paths did not reach the final aliasless/symbolless read.
4. The following definitions and export expressions were tried separately:

```ts
export interface Dict extends Record<string, number> { [key: symbol]: number }
declare const ctor: { new(): {} } & Record<string, number>;
export class StaticDict extends ctor { static [key: symbol]: number }
```

```js
/** @type {import('./defs').Dict} */
const dictionary = {};
module.exports = dictionary;
module.exports.extra = 2;
```

The identifier export retained Dict's symbol. Replacing the expression with
`module.exports = /** @type {import('./defs').Dict} */ ({});` removed the symbol,
but the reference's resolved members did not contain an index symbol. Neither
reached the read. A typeof StaticDict cast similarly removed the symbol and
retained the index symbol, but its inherited mapped string index was not
inherited on the static side, so xyz had no applicable index.

5. `module.exports = class { static [key] = 1; };` with a broad string key and
   `module.exports.extra = 2` creates a declaration-free index but retains the
   assigned class symbol, because the assignment attaches the augmentation to
   that class. Separating the Computed class and casting an otherwise empty
   class expression to its typeof is the retained exported.js.txt fixture.
   It removes both aliasSymbol and symbol and reaches the empty declaration
   list. The final declared-index positive control uses the same export shape.

The first combined final run had an unused exploratory StaticDict definition,
which emitted TS2417 in the positive-control dependency. It was removed from
the final defs.ts.txt; the final committed fixtures and measurements are clean.
