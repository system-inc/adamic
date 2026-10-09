### conformance tests conformance tests for tests/cases/conformance/parser/ecmascript5/Protected/Protected1.ts Correct type/symbol baselines for tests/cases/conformance/parser/ecmascript5/Protected/Protected1.ts (baseline-content)

```text
Error: New baseline created at tests/baselines/local/Protected1.types
    at writeComparison (/workspace/scratch/verdict-gate/lane-serial/adapted-tree/src/harness/harnessIO.ts:1490:27)
    at Object.runBaseline2 [as runBaseline] (/workspace/scratch/verdict-gate/lane-serial/adapted-tree/src/harness/harnessIO.ts:1509:9)
    at checkBaseLines (/workspace/scratch/verdict-gate/lane-serial/adapted-tree/src/harness/harnessIO.ts:732:26)
    at Object.doTypeAndSymbolBaseline (/workspace/scratch/verdict-gate/lane-serial/adapted-tree/src/harness/harnessIO.ts:695:13)
    at _CompilerTest.verifyTypesAndSymbols (/workspace/scratch/verdict-gate/lane-serial/adapted-tree/src/testRunner/compilerRunner.ts:327:18)
    at Context.<anonymous> (/workspace/scratch/verdict-gate/lane-serial/adapted-tree/src/testRunner/compilerRunner.ts:104:80)
    at callFn (/workspace/scratch/verdict-gate/lane-serial/adapted-tree/node_modules/mocha/lib/runnable.js:364:21)
    at Test.Runnable.run (/workspace/scratch/verdict-gate/lane-serial/adapted-tree/node_modules/mocha/lib/runnable.js:352:5)
    at Test.Mocha.Runnable.run (/workspace/adamic/stage3/oracle/observe-errors.cjs:11:24)
    at Runner.runTest (/workspace/scratch/verdict-gate/lane-serial/adapted-tree/node_modules/mocha/lib/runner.js:677:10)
    at /workspace/scratch/verdict-gate/lane-serial/adapted-tree/node_modules/mocha/lib/runner.js:800:12
    at next (/workspace/scratch/verdict-gate/lane-serial/adapted-tree/node_modules/mocha/lib/runner.js:592:14)
    at /workspace/scratch/verdict-gate/lane-serial/adapted-tree/node_modules/mocha/lib/runner.js:602:7
    at next (/workspace/scratch/verdict-gate/lane-serial/adapted-tree/node_modules/mocha/lib/runner.js:485:14)
    at Immediate._onImmediate (/workspace/scratch/verdict-gate/lane-serial/adapted-tree/node_modules/mocha/lib/runner.js:570:5)
    at processImmediate (node:internal/timers:534:21)
Error: New baseline created at tests/baselines/local/Protected1.symbols
    at writeComparison (/workspace/scratch/verdict-gate/lane-serial/adapted-tree/src/harness/harnessIO.ts:1490:27)
    at Object.runBaseline2 [as runBaseline] (/workspace/scratch/verdict-gate/lane-serial/adapted-tree/src/harness/harnessIO.ts:1509:9)
    at checkBaseLines (/workspace/scratch/verdict-gate/lane-serial/adapted-tree/src/harness/harnessIO.ts:732:26)
    at Object.doTypeAndSymbolBaseline (/workspace/scratch/verdict-gate/lane-serial/adapted-tree/src/harness/harnessIO.ts:702:13)
    at _CompilerTest.verifyTypesAndSymbols (/workspace/scratch/verdict-gate/lane-serial/adapted-tree/src/testRunner/compilerRunner.ts:327:18)
    at Context.<anonymous> (/workspace/scratch/verdict-gate/lane-serial/adapted-tree/src/testRunner/compilerRunner.ts:104:80)
    at callFn (/workspace/scratch/verdict-gate/lane-serial/adapted-tree/node_modules/mocha/lib/runnable.js:364:21)
    at Test.Runnable.run (/workspace/scratch/verdict-gate/lane-serial/adapted-tree/node_modules/mocha/lib/runnable.js:352:5)
    at Test.Mocha.Runnable.run (/workspace/adamic/stage3/oracle/observe-errors.cjs:11:24)
    at Runner.runTest (/workspace/scratch/verdict-gate/lane-serial/adapted-tree/node_modules/mocha/lib/runner.js:677:10)
    at /workspace/scratch/verdict-gate/lane-serial/adapted-tree/node_modules/mocha/lib/runner.js:800:12
    at next (/workspace/scratch/verdict-gate/lane-serial/adapted-tree/node_modules/mocha/lib/runner.js:592:14)
    at /workspace/scratch/verdict-gate/lane-serial/adapted-tree/node_modules/mocha/lib/runner.js:602:7
    at next (/workspace/scratch/verdict-gate/lane-serial/adapted-tree/node_modules/mocha/lib/runner.js:485:14)
    at Immediate._onImmediate (/workspace/scratch/verdict-gate/lane-serial/adapted-tree/node_modules/mocha/lib/runner.js:570:5)
    at processImmediate (node:internal/timers:534:21)
```

Elapsed: 967.657 ms; timeout: 40000 ms.

Stack top:     at writeComparison (src/harness/harnessIO.ts:1490:27)

Protected1.symbols (first 40 lines)

```diff
--- reference/Protected1.symbols
+++ local/Protected1.symbols
@@ -4,3 +4,6 @@
 protected class C {
 >C : Symbol(C, Decl(Protected1.ts, 0, 0))
 }
+const stage3ForcedMismatch = 123;
+>stage3ForcedMismatch : Symbol(stage3ForcedMismatch, Decl(Protected1.ts, 2, 5))
+

```

Protected1.types (first 40 lines)

```diff
--- reference/Protected1.types
+++ local/Protected1.types
@@ -5,3 +5,9 @@
 >C : C
 >  : ^
 }
+const stage3ForcedMismatch = 123;
+>stage3ForcedMismatch : 123
+>                     : ^^^
+>123 : 123
+>    : ^^^
+

```
