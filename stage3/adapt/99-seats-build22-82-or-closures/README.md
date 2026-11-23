# Fixed-list or callbacks

Adamic ruled this adaptation on Nov 23 2026 (#bbssdcy, families 82 and 169); Compiler accepts the head after its review and census.

The Nov 22 callback-helper shape applies under `docs/rulings.md` rows
137, 151 and 152. The `utilities.ts` unit `usesExtensionsOnImports`
uses a two-callback closure for the default `hasExtension` initializer.
The existing parameter annotation contextually types `text`. This
removes family 82's utilities finding and the initializer's path to
family 169 at core's first `or` overload.

Both extension tests are function declarations with no reassignment.
The closure calls them in order and returns the first truthy result or
the last result. Row 152's runtime-built list still uses `or(...checks)`.

The checker site uses a six-predicate disjunction in the `find` call.
The explicit `boolean` return keeps `firstPropertyLikeDecl` at
`Declaration | undefined`. Compiler authorized keeping the existing `or`
import on Nov 23 at 08:12:14Z after the meter check. The whole-program
checker reported zero diagnostics on both base and head.

`adapt.cjs` checks both exact callback snapshots and the utilities annotation.
Run `node test.cjs <adapted tree>` with the apply tool's pinned
TypeScript on `NODE_PATH` to compare the closures with the actual core
`or` implementation using recording stubs.
