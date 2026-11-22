# Readonly deprecated-suggestion declarations

Adamic task #k8ed1f2, family 3 cause 9. Performance delegated this slice
as #bbssdcy.

## Edit and premises

`addDeprecatedSuggestion` accepts `readonly Node[]`. Its worker accepts
`Node | readonly Node[]` and uses the already imported core `isArray`
predicate. Stock TypeScript narrows the false branch to `Node` and the
true branch to the readonly array. `Array.isArray` has a mutable-array
predicate, so changing only the annotations leaves a readonly-array
possibility in the false branch at `getJSDocDeprecatedTag`. That
annotation-only form was rejected.

The parameter references are the forward to the worker, the tag
selector's guard, and its two branches. No assignment, element write,
or mutating call uses either parameter. The eight entry calls and
two worker calls still accept their existing arguments. The writer
inventory in the audit names the source positions, actual and contextual
types, and the stock checker's assignability result for each argument.
Both functions are private to `createTypeChecker`.

The adapter finds the two parameters and the selector condition by
their AST owners, checks the branches and exact type text, and refuses
a partly applied batch. It also checks the existing named import and
core guard. A rerun on the adapted source writes nothing. Earlier
adapters leave the expected three slots. Later adapters address other
owners.

## JavaScript consequence

The emitted selector changes `Array.isArray(declarations)` to
`isArray(declarations)`. Core's wrapper returns
`Array.isArray(value)` directly. The same input reaches the same
built-in once, so the same branch executes with the same declaration
or array. The selected tag, related diagnostic, and suggestion write
keep their existing computation. The annotations erase and the helper
was already imported. No public signature is edited.

The audit's `deprecated-suggestion-count.sh` builds an instrumented
copy under the sealed house preamble. It keeps the old and new
selector in that copy, compares each selected tag with `Object.is`,
and adds the diagnostic only once. It requires a positive observation
count and checks that a planted mismatch reaches the failure branch.
The local smoke mode uses the stock parsed tag objects and actual core
selector helpers.

## Census and owner acceptance

Performance accepted the different lowering stop on Nov 22, 2026.
Its Loom proof, at Adamic c02b30231 on one machine per job: the lane
value count J-0795wgt, the native suite J-rqakd70 against J-b2fjhep,
and the census J-cr56xxs.

On Adamic c02b30231, the checker refusal scan goes from 675 to 668
distinct sites. These seven adapted-tree positions clear:

| Position in checker.ts |
|---|
| 19252:61 |
| 31072:43 |
| 33878:73 |
| 35005:48 |
| 42810:25 |
| 48572:51 |
| 48592:59 |

Each of the seven stops said “a value of type Declaration[] seen as
Node[], which can write Node where Declaration is read.”

Two existing refusals say “an object refinement using an open numeric
enum as a literal tag.” The same expressions move from checker.ts:2600:69
to 2600:63 and from 2600:130 to 2600:124 because the guard is six
characters shorter. Their stop reasons stay the same.

The checker statement at `checker.ts:2600` rolled back both before
and after the edit. The full checker-file census records two different
`NotYet` rules for that statement.

- Before, `checker.ts:2600:45` reported “Array.isArray on a tuple or an erased object/any/unknown view”.
- After, `checker.ts:2600:63` reports “a narrowed union member whose object tag cannot be checked with typeof; keep differently held object kinds in separately typed variables”.

The boundary at `checker.ts:2600:9` records the rollback on both sides.
Restoring the original source reproduced the full control site list.
