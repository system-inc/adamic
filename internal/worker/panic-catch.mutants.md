# A panic is never caught, in a Worker too

A Worker's panic throws an AdamicPanic for the bridge to answer 500, so every
try the JavaScript backend emits asks the runtime's `isPanic` of the value it
caught: a catch throws a panic on, and a finally body is skipped while one
unwinds, by a flag local to that try that a catch around it sets. The test is
the value's identity, never module state, since a Worker's async handler
interleaves requests at every await. Node's runtime answers false for every
value, because there a panic exits the process.

`TestWorkerPanicCatchMutantsAreCaught` uses the existing Worker buildcache
mutation products and runs each unmutated control before its mutated child.
Each top-level product test builds only its own mutated worker.test.

| Mutation | Catching test | Witness in panic_catch_test.go |
| --- | --- | --- |
| panic-catch-identity: drop the catch's `adamicIsPanic` line | TestWorkerPanicIsNeverCaught/catch | `a panic reached the program catch` |
| panic-finally-guard: drop the flag from the finally's guard | TestWorkerPanicIsNeverCaught/finally | `a finally ran while a panic unwound` |
| panic-finally-unset: the catch around the try never sets the flag | TestWorkerPanicIsNeverCaught/finally_return | `a finally return swallowed a panic` |

The interleaved case starts a panicking request beside eight that throw and
catch ordinary errors, each after a different number of awaits, so some of
their catches and finallies run while the panic is still unwinding.
