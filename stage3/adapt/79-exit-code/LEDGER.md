## 2. Truthiness treats a returned exit code as a watch handle

Status: soundness adaptation 79 under the Oct 19 language-lead ruling for
#se1zcgr; draft issue, not filed.
Upstream: TypeScript 6.0.3, commit 050880ce59e30b356b686bd3144efe24f875ebc8.
Locations: src/compiler/sys.ts System.exit; src/compiler/executeCommandLine.ts;
src/testRunner/unittests/helpers/tscWatch.ts:78 and
tscWatch/incremental.ts:77 (locations in the numbered adapted tree).

System.exit promises void, but test hosts override it with
`sys.exit = exitCode => sys.exitCode = exitCode`. executeCommandLine forwards
that value through its void arm. The harness treats the result as void or a
watch/solution object, and `if (closer) closer.close()` incorrectly excludes
all numbers using the declared void arm. A nonzero exit code is truthy.

Node 24.19.0 counterexample, using the real built TypeScript compiler:

```js
const ts = require("./built/local/typescript.js");
const sys = { ...ts.sys, write() {}, exit(_code) { return 1; } };
const closer = ts.executeCommandLine(sys, () => {}, ["--invalid-adamic-option"]);
console.log(closer); // 1
if (closer) closer.close(); // TypeError: closer.close is not a function
```

Observed: the invalid-option path calls exit with 1, executeCommandLine returns
1, and the harness's exact old close statement throws the TypeError above.
[The runnable proof](../adapt/79-exit-code/proofs/host-witness.cjs) extracts that
statement from the actual consumer, rather than maintaining a second copy.
This is a counterexample to the narrowing, not evidence that the unchanged
upstream suite exercises that invalid-option input in this particular consumer.

Our adaptation declares System.exit as number | undefined, gives the two
noop-as-exit test hosts dedicated stubs with that signature, and narrows the
two consumers using `typeof result === "object"`. Compiler JavaScript stays
byte-identical. The full unchanged upstream suite and its exact passing count
are the condition on accepting the harness changes; no test or baseline input
is added, deleted, filtered, or rewritten by adaptation 79.
