# Reset the yielded item's depth (#2p41tb1)

Inside the unwrap, replace the pairwise call with
`l.containerPairGoes(l.typeArguments(value)[0], l.typeArguments(target)[0], false, 0, walk)`.
Keep the same walk state, changing only the depth charged to S -> T.

The recursive fixture lowers and remains green under this mutant:
`TestYieldResultUnwrapRecursive`. Its production compile took 0.289 seconds in
the oracle run (the standalone CLI probe took 0.069 seconds); Node prints `1`.
The S field contains IterableIterator<S>, and the busy/done handling terminates.

There is a red pin: `TestYieldResultUnwrapRefused/own_symbol`. Resetting the
built-in yield arm's depth gives the program's own wrapper fresh depth too, so
it lowers instead of keeping today's fail-closed container-fit stop. The
recursive pin alone does not distinguish this mutation.

Run the new pins and both existing families with:
`go test ./internal/oracle ./internal/lower -run '^(TestYieldResultUnwrap|TestContainerInterfaceValueIsNotYetWhereAMapIsHeld|TestContainerIsNotYetWhereAShapeIsHeld|TestOptionalIndexingMapShapeRefused)' -count=1`.
