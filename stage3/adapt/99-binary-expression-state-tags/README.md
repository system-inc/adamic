# Private binary-expression traversal states

The Nov 22 family 14 ruling on #bbssdcy accepts a JavaScript-changing adaptation
inside `src/compiler/factory/utilities.ts`. The state stack holds six private
string tags. `nextState` receives and returns those tags, and the trampoline
dispatches directly to each original handler. The original transition switch,
its fallthrough, callback order, child stack operations, fold-side condition,
terminal test, and all assertions stay in the same positions in the traversal.
The handlers do not read `this`. The local state stack is never exposed to a
callback. A valid stack slot therefore identifies exactly one original handler.
For ordinary arrays, every write to a live slot is the walker's own (the root's
enter, the push's enter, and `nextState`'s closed set), and no callback receives
the state array.
If any producer inside tsc ever writes a non-tag into that array, the
`default` becomes owed that day.
[Review-o's reported prototype probe](proofs/review-o-prototype.md) witnesses
the excluded inherited accessor case and the hang when it substitutes an
unknown state tag.

The exact AST-owned before and after snapshots cover the state alias, handler
namespace and trampoline implementation. The class declaration is also pinned
to guard callback types and ordering. A mixed or unknown snapshot is rejected
before a write. The exported overloads are outside the replacements.

| Shape | Site coordinates in this adapter's pinned before snapshot |
| --- | --- |
| Compare a stack state with a named state | 1282:27, 1295:27, 1313:27, 1327:27, 1345:27, 1366:27 |
| Pass a state to `nextState` | 1284:53, 1297:53, 1315:53, 1329:53, 1346:53 |
| Write the result of `nextState` | 1284:34, 1297:34, 1315:34, 1329:34, 1346:34 |
| Store an enter state directly | 1394:34, 1477:54 |

All coordinates refer to `src/compiler/factory/utilities.ts` in this adapter's
`-before` tree `7a13290fa12616d8699d8d962c2cc56d49920b2b`, shared by Adamic
`ef741d4f0860597ff8d4df5dd042973b2f20819d`, `0e7517acfe9e7a3631d12408766e880c76bd4e9d`,
and `0147047509c7f8bc5343efdd8a749ec335b92622`.
The quoted ruling refuses a binder-context compiler change for all 18 listed
sites. The source adaptation follows `docs/rulings.md:96`, `:100`, and `:109`
for invariant nominal class views and `:131` for higher-rank generic function
values. It uses the qualified namespace calls of `:276` and leaves the
namespace identity contract of `:279` intact. The public declaration boundary
in `:406` applies to the emitted declarations.

The accepter is Adamic because this adapter changes JavaScript.

Run the test on the tree from `stage3/apply.sh`:

```sh
timeout --signal=KILL 60 env NODE_PATH="$PWD/stage3/api/node_modules" node stage3/adapt/99-binary-expression-state-tags/test.cjs "$APPLIED_TREE"
```

The test reads this adapter's own before and after refs, replays the edit,
checks four structural refusals, and compares 20 paired walker traces
for 27 rows in all. At each callback and once after the walk, the walker rows
record each registered object's identity, each own key's full descriptor,
extensibility, and prototype by label. The record is one level deep.
Mutations inside a child object the caller owns, such as the checker's
`WorkArea.typeStack`, are outside this proof's scope: a limit of the proof,
not a gap in it. That child belongs to its owner's own contract, and the
adapter neither names nor touches it.
