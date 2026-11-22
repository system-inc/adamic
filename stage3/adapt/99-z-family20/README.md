# Family 20 adaptation evidence

The scratch candidate leaves the visitor and signature-copy source unchanged. It specializes the JSDoc unary and tag getters and workers per concrete kind, retains the fresh mutable draft until the final postfix write, and widens only the watcher fallback read.

| Kirk's upstream list | Existing promise | Observed result |
|---|---|---|
| `src/compiler/visitorPublic.ts:604:12` | `visitEachChild<T extends Node>(node: T, ...): T` promises the caller's full subtype | TypeScript 6.0.3 accepts a caller-added `extra: number` on a `QualifiedName`. Replacing its right child returns the same kind and a different node whose `extra` is undefined. Node prints `true newRight undefined`. The site stays Refused. |

## Public declaration lines

The two property declarations keep the optional marker and add explicit undefined. Their factories already assign to the property on every construction, so the missing child creates an own property with undefined as its value. Rule :78 distinguishes that state from an absent property.

| Owner | Adapted source | Emitted module declaration | Shipped public declaration line containing the original property |
|---|---|---|---|
| `JSDocReturnTag.typeExpression` | `src/compiler/types.ts:4063` | `compiler/types.d.ts:1925` | `lib/typescript.d.ts:5834` |
| `JSDocThrowsTag.typeExpression` | `src/compiler/types.ts:4095` | `compiler/types.d.ts:1952` | `lib/typescript.d.ts:5861` |

```diff
-    readonly typeExpression?: JSDocTypeExpression;
+    readonly typeExpression?: JSDocTypeExpression | undefined;
```

`emitted.diff` contains the emitted JavaScript and module declaration diffs. JavaScript changes are confined to `compiler/factory/nodeFactory.js`, where the memoized wrappers and their worker bodies become concrete. The wrapper count rises from eight to 36 in this experiment. The two property declarations, the private draft annotations and the watcher widening erase. TypeScript's separate bundled declaration generator was outside this module-emit comparison.

## Signature-copy claim

At original `src/compiler/factory/nodeFactory.ts:1227:23` and candidate `:1255:23` the stop is `a value of type Mutable<T> seen as T, a type parameter whose constraint SignatureDeclarationBase can be written, so it can write what Mutable<T> can't hold`. The expression `update(updated, original)` offers the mapped draft as the caller's `T` after `updated.typeArguments = original.typeArguments`. The reported relation is on the updated object entering `T`. The array itself is shared by identity. The candidate holds the helper and the copy at their existing types.

The public API gate expectation is regenerated from the pinned upstream API baseline and the two owner-specific sanctions in `api-sanctions.json`. The new public expectation changes `JSDocReturnTag.typeExpression` at line 5831 and `JSDocThrowsTag.typeExpression` at line 5858. `public-api.diff` spells out those two changes against the preceding expectation.


## Census and exposed stops

The full 81-file entry census at 0679623c2aa66cce5144ba0f894f020e3c5c6d99 removes five family-20 sites and the old readonly-flags cast at nodeFactory.ts:5081:22. It goes from 2,144 to 2,138 unique meter keys after source alignment. There are no aligned additions, no exposed stops, no changed first terminal stop, and no changed retained-stop unit, phase or occurrence count. Both scoped-tail findings match exactly.

The standalone absent-field write control reaches NotYet at optional-tag-writes.ts:7:5 on both backends. Its words are “writing a possibly absent optional own field, where a value reaching it is built as Mutable<ReturnTag> without typeExpression”. This witness concerns backend support for creating a field after allocation. It is separate from the full entry census and the approved declaration change.

## Factory replay check

On both initial application and replay, the adapter extracts and executes the generated nullable and non-nullable getters and their real workers before writing any files. The isolated run supplies an allocator, parenthesizer and metadata fixtures. It checks memoized identity, both explicit postfix values and the default, and both no-op and changed-child updates.

Run the same assertions against the full stock-emitted compiler namespace with:

```sh
node stage3/adapt/99-z-family20/check-postfix.cjs \
  /absolute/emit/compiler/_namespaces/ts.js /absolute/stage3/api/node_modules/typescript
```
