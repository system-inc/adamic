# A traced generator parks at each yield: mutants (#7wn0t67)

Each mutant is one edit, applied alone, then reverted. Run
`go test ./internal/flow -run '^TestFlowGeneratorsTraced$' -count=1 -v`.
Not run yet: written on a machine that doesn't build Go, so "goes red" below is what the code says, not what a run
printed.

## M1. The yield emitted plain, with no park or resume around it

File: internal/javascript/javascript.go, `statement`, the `ir.Evaluate` case.

Old:

```go
		if yield, ok := statement.Value.(ir.Yield); ok && e.options.Suspend != nil {
```

New:

```go
		if yield, ok := statement.Value.(ir.Yield); ok && false {
```

The activation is still declared, since the condition at the function's entry is untouched, but the yield is
`(yield value);` again, as before the fix, and the generator's frame stays on the trace's stack while the consumer
runs.

Expected: trace_in_place.a and trace_named.a go red. checkGraphPathsProgram runs first and reports walk's problem,
`a point of function 0 ran inside a call of function 2` (the consumer's point under the generator's call; the
numbers are whatever the two functions' indexes are). checkLivenessProgram then panics in frames' consumer:
`index out of range [12] with length 3`, the consumer's instruction ids filed under the generator's three
instructions, and the panic stops the binary there. Run alone, each of the other programs goes red in walk the same
way at the first point its consumer runs after a yield; frames panics for one only where a consumer's id is past the
generator's instructions.

## M2. walk's park arm back to an await alone

File: internal/flow/trace_test.go, `walk`, the `park ` case.

Old:

```go
				if !Suspends(block.Terminal) || position+1 != len(block.Instructions) {
```

New:

```go
				if _, ok := block.Terminal.(*Suspend); !ok || position+1 != len(block.Instructions) {
```

Expected: every program in the test goes red with `park bypassed a suspension terminal`, at each park of a generator,
whose block ends in a `*Yield`. frames still splits the trace right, so liveness and definitions stay green; walk is
the one check that reads the terminal a park left from.

## M3. A generator closed at a yield held to ending where the graph returns

File: internal/flow/trace_test.go, `walk`, the `leave ` case.

Old:

```go
			if top.resumed && top.last != nil {
```

New:

```go
			if false && top.resumed && top.last != nil {
```

Expected: view_body_throw.a goes red with `function N: the call ended at bbM, which doesn't lead to a return`: its
consumer's throw closes `numbers` at the yield inside its loop, and the yield's next block runs the loop's update, so
nothing but the allowance lets the call end there. trace_in_place.a and trace_named.a stay green: each yield is the
body's last statement, and its next block reaches the return through blocks that run nothing.
