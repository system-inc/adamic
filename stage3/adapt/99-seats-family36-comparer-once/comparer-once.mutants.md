# Comparer once: the test's wrong versions

`test.cjs` prints 20 `PASS` lines on this head. Each wrong version below edits
`adapt.cjs`'s replacement text and is planted in the repository's own adapter,
and the whole tree is applied with it by `stage3/apply.sh`, so the tree's
`-after` ref holds the wrong text. The test's first two checks, its output
against that ref and its cut, then stay green, and only the behavioral drive
can turn it red. Run from the repository root:

```sh
NODE_PATH="$PWD/stage3/api/node_modules" node \
    stage3/adapt/99-seats-family36-comparer-once/test.cjs "$APPLIED_TREE"
```

1. `wrong-second-call`: the relational arm calls `comparer(next, last)` a
   second time, `const relational = comparer(next, last);`, and tests that.
   Red at `test.cjs:106` (`sameCalls`, from `check`): "false pushes and
   numeric zero continues adapted call count", 3 where the original makes 2.
2. `wrong-false-continues`: the equality arm continues on `false` as well,
   `if (comparison || !comparison) continue;`. Red at `test.cjs:105`
   (`sameArray`, from `check`): "false pushes and numeric zero continues
   adapted array length", 1 where the original keeps 2.

Both are the lane's round 3 patches, family15c-r3-wrong-second-call.patch
(sha256 `b8b9b765...`) and family15c-r3-wrong-false-continues.patch
(`2492cc33...`), attached to #8frc00b.

## Where they ran

Loom's J-kjy7jy9, at 596d93f6 on main 25f3ed7f, planted each version, applied
the whole tree with it, printed the `-after` ref's adapted loop, and ran the
test: both red as above, each after the output and cut checks passed. The
clean head was green at 20 `PASS` in the same job. Run without the whole-tree
apply, against the clean tree's ref, each version reds at `test.cjs:30`, the
output check, which proves only that its text differs (J-2mxkras).
