# visitNodes checks what it returns

Task #ytagqs5. JavaScript-changing adaptation. Adamic's ruling on #ytagqs5: the four
element tests in `visitArrayWorker` run whenever `test` is given, failing in tsc's own
words, and the two loops in `visitNodes` right before its returns are covered too. The
claim they make true is the one `visitNodes`' overload 1 states: every element of the
`NodeArray<TOut>` it returns passed `test`.

## The edit

`src/compiler/visitorPublic.ts`. Upstream checks each element `visitArrayWorker` keeps
with `Debug.assertEachNode` and `Debug.assertNode`, which do nothing below
`AssertionLevel.Normal`, and tsc runs at `None`. Six edits, each made at a node the
adapter finds by the syntax tree:

| Owner | Site | Edit |
| --- | --- | --- |
| visitArrayWorker | `Debug.assertEachNode(updated, test);` after the prefix slice | the element loop over `updated` |
| visitArrayWorker | `Debug.assertNode(visitedNode, test);` before its push | the single check of `visitedNode` |
| visitArrayWorker | `Debug.assertNode(visited, test);` before its push | the single check of `visited` |
| visitArrayWorker | `Debug.assertEachNode(nodes, test);` before `return nodes` | the element loop over `nodes` |
| visitNodes | before `return updatedArray`, after `setTextRangePosEnd` | the element loop over `updatedArray`, inserted |
| visitNodes | before the final `return nodes` | the element loop over `nodes`, inserted |

The element loop, and the single check:

```ts
if (test !== undefined) {
    for (const testedNode of updatedArray) {
        if (!test(testedNode)) Debug.fail("False expression: Unexpected node.\r\nVerbose Debug Information: Node array did not pass test.");
    }
}
return updatedArray;
```

```ts
if (test !== undefined && !test(visited)) Debug.fail(`False expression: Unexpected node.\r\nVerbose Debug Information: Node ${Debug.formatSyntaxKind(visited.kind)} did not pass test.`);
```

Each loop sits as the statement right before the return of the array it checks, with
nothing between them. The `nodes === undefined` return is untouched.

## tsc's own words

The messages are the ones `Debug.assert` composes at `Normal` from the adapted
`debug.ts`: `False expression: ` and assertEachNode's and assertNode's default message
`Unexpected node.`, then `\r\nVerbose Debug Information: ` and each one's verbose text
(`Node array did not pass test.`, and `Node ${formatSyntaxKind(node?.kind)} did not pass
test.` since 99-correctness-core-debug-messages). `Debug.fail` adds `Debug Failure. `, so
a failure throws the message stock throws at `Normal`, byte for byte. The adapter reads
those words in `debug.ts` by their nodes and refuses the apply if any drifts. At this
adapter's place in apply order the verbose texts still end with upstream's
`'${getFunctionName(test!)}'.`, which 99-correctness-core-debug-messages drops later, so
the adapter reads the words up to `did not pass test`, and `test.cjs` holds the messages
to the applied tree's `debug.ts`, the one file it reads outside its snapshots.

What stays stock's: the condition at `Normal` also failed an undefined node in
`assertNode`. Here the element is never undefined (`visited` is truthy in its branch,
and `visitedNode` is an element of a visitor's `readonly Node[]`), and the ruled check is
`test !== undefined && !test(X)`. The stack crawl mark is `Debug.fail`'s own, not the
assertion's.

## The adapter

`pins.json` holds each owner whole, before and after, with CRLF bytes and SHA-256
hashes. The adapter finds the two implementations by name, and in `visitArrayWorker` the
four `Debug` call statements: their count, their order (`assertEachNode(updated)`,
`assertNode(visitedNode)`, `assertNode(visited)`, `assertEachNode(nodes)`), two name
arguments with `test` second, and that `test` is called nowhere else. In `visitNodes` it
finds the three returns, `nodes`, `updatedArray` and `nodes`, the first alone under
`nodes === undefined`, the second right after `setTextRangePosEnd(updatedArray, pos,
end);`, the third the body's last statement. Any other shape is refused before writing,
as is an owner that is neither pin, one owner adapted and the other not, or a name
`testedNode` already in either owner. A second run writes nothing.

The file's JavaScript changes only inside the two owners, with and without comments, and
its declarations don't change: `visitArrayWorker` is unexported, and an implementation
body is not `visitNodes`' declaration.

## Test

From the Adamic checkout, after a full apply:

    bash stage3/apply.sh "$TMPDIR/tree"
    NODE_PATH="$PWD/stage3/api/node_modules" node stage3/adapt/99-apple-visit-nodes-guard/test.cjs "$TMPDIR/tree"

`test.cjs` reads this adaptation's `-before` and `-after` refs, never the final tree.
Each check prints `--- PASS: <name>` or `--- FAIL: <name> (<why>)`:

- **TestSixEditsInPlace:** by the syntax tree and without the adapter's code, no
  `Debug.assert*` call is left in `visitArrayWorker`, each site holds its check beside its
  neighbours, each `visitNodes` loop is the statement right before its return, and with
  the six edits undone the file is `-before` byte for byte.
- **TestOrderOfChecks:** on six visits whose elements all pass, `test` is called in the
  order stock calls it at `Normal`, then once more over the returned array, and the
  result is stock's at `None`.
- **TestTscWords:** an element failing at each of the four worker sites throws the same
  message as stock at `Normal`, with `test` called on the same elements. Stock's side runs
  `debug.ts`'s own `assert`, `fail`, `assertEachNode` and `assertNode` from the snapshot.
- **TestVisitNodesLoops:** a test that passes a node once and fails it after is caught
  only by the two `visitNodes` loops, in tsc's words; stock at `Normal` is quiet.
- **TestAdapterReplay:** the adapter on `-before` writes `-after`, a second run writes
  nothing, and a drifted owner, a half-adapted file and drifted `debug.ts` words refuse.

A reviewer's mutation of the output turns it red: drop any one of the six edits, or
move a loop after its return, in the `-after` ref.

`node check.cjs TREE` runs stock tsc over `src/compiler` on the applied tree, with the
adapted owners and with the before owners put back: no diagnostics on either.

## The count

Adamic's premise is tsc's own suite, which runs these assertions live at `Normal`
(`stage3/oracle/run-tasks.cjs` sets `NODE_ENV: 'development'`) and passes. Over the lane
oracle at main 46a1954d, J-ezkhwzs counted the four worker sites at 1,104,013 calls and
3,830,226 elements, the `return updatedArray` loop at 366,283 and 983,167, and the final
`return nodes` loop at 708,848 and 2,816,430, all with 0 failures (J-79earng counted the
worker sites first). tsc's emit was byte-identical over 320 files, and the cost sat
inside the bench's noise. The probe's checks failed with the verbose text alone; the
message is read only on a failure, which the count never reached.
