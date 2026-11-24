# Scoped async super helper text

Adamic ruled on family 78 at 21:49Z on Nov 22, 2026 in #bbssdcy.
Compiler is the accepter under Adamic's widened split after its review and
census. The governing row is `docs/rulings.md:398`. In the split's applied
tree the sites are `src/compiler/factory/emitHelpers.ts:1471:11` and
`:1479:11`. The two scoped async super helpers close over the printer's
`uniqueName` callback and ask for `"_superIndex"` once per rendering. The
ordinary templates retain the tagged templates' cooked head and tail bytes.

Run the test on a tree written by `stage3/apply.sh`. It reads this adapter's
before and after snapshots from its `refs/stage3/` refs, renders both helpers
twice with a stateful callback beside the original, and runs both versions
through the stock printer, in a file that already binds `_superIndex`. It
also checks the adapter's byte-exact replay and its refusals on partly
adapted input, a second substitution, and a changed `helperString` body,
which the adapter pins by its node text before any write and on replay.

```sh
NODE_PATH="$PWD/stage3/api/node_modules" node \
    stage3/adapt/99-seats-build22-78-unique-name-closures/test.cjs \
    '<applied tree>'
```
