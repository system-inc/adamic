# Mutants for the callable boxed-split words (#bjtwhg8)

## M1: the weak sentence back

Delete the `callableBoxedSplit` branch before the weak sentence in `internal/lower/expression.go`, so a callable whose slot is boxed on one side and unboxed on the other speaks main's words again, "(one keeps something weakly that the other keeps strongly)".

Expected: `TestCallableBoxedSplitNamesTheSlot` goes red on both shapes, and `TestRestElementWeakControl` stays green, since its slot differs by Weak alone.
