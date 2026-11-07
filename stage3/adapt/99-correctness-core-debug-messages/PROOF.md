# Completed road (b) proof

Task b5w3ycg, Adamic, Nov 7 19:16Z. JavaScript-changing at the four lazy failure messages listed in README.md.

Base/main/compiler: `d15620712c459baf72b3d359c1829df8d0808fda`. TypeScript 6.0.3 source: `050880ce59e30b356b686bd3144efe24f875ebc8`. Node `v24.19.0`, linux/x64. The ownership audit preceded editing; `ownership-audit.json` pins the inspected sources. No active pin covered these four whole owners or message expressions, so this is a new directory.

## 1. Separate zero counts over the whole lane oracle

Every manifest unit from `stage3/lane/shard-plan.json` ran through the ordinary `stage3/lane/shard.py` path used to split `stage3/lane/run.sh`, following the referenced record-table measurement. The accepted inventory is exactly 398 units, once each, reconstructed from all 20 raw piece archives. The ordinary test-identity, public-API, baseline-diff and test-count guards all pass. The unchanged expectation is 106366 passes, 1 expected public-API acknowledgement failure, and 0 skips.

| Callback | Whole-oracle executions | Forced failure before | Forced failure after |
| --- | ---: | ---: | ---: |
| assertEachNode message | 0 | 1 | 1 |
| assertNode message | 0 | 1 | 1 |
| assertNotNode message | 0 | 1 | 1 |
| assertOptionalNode message | 0 | 1 | 1 |

The observer instruments the callback body itself, not the enclosing assertion or the predicate. It prepends a counter and returns the original expression. All four callbacks were located independently in each of `_tsc.js`, `typescript.js` and `run.js`. Only private measurement copies of these emitted files were changed. Normal build outputs and compiler source were not instrumented. The preload exposes the counter to Node workers and VM contexts and emits explicit zero rows on process exit. There are 797 retained Node process records. The harness/oracle worker code uses child processes; no worker_threads import is present in those sources.

The full baseline oracle is the zero-count measurement. A second whole candidate oracle is not claimed. Both normal builds, candidate type checks, completed-tree witnesses, emitted positive controls, declaration comparisons, paired replay and census comparison are separate proofs. `whole-lane-reconciled.json` is produced by reopening all archive hashes, inner manifests, raw guard reports, identities and process rows. The largest accepted piece took 143.936 seconds.

The initial whole-oracle controller reached its generic 1490-second cap after 220 accepted units. Its actual child exit was -15. All partial child groups were stopped; incomplete, unaccepted units were retained separately and rerun. The remaining accepted pieces ran in smaller batches. No test failure is hidden by this restart, and no incomplete unit contributes to the exactly-once accepted inventory. `raw/discarded-incomplete-lane.json` and command results are retained in raw-proof.tar.gz.

## 2. Only failure text changes

The exact CRLF before/after source audit changes four message expressions and no other bytes. It removes the space and quoted function name before the existing final period. Node kind interpolation remains. Each callback is still lazy and still occupies the same assert argument. Assertion conditions, assertion-level checks, arguments, stack-crawl marker, `assert`, `fail`, `new Error`, `throw e` and their ordering remain byte-identical. The helper whole-body pins guard this property before writing.

The final Node witness executes the completed before/after apply texts, not reimplementations. It covers all four assertions at levels 0, 1, 2 and 3, passing and failing, for 32 paired cases. Enabled failures throw the same Error constructor/prototype with the shorter message. Passing and disabled cases return and execute no message callback. Every failing callback counter fires exactly once, independently, before and after. `final-witness.json` retains every class, message and count. The built-artifact controls independently load each side's actual instrumented typescript.js and force all four failures.

The adapter writes nothing on rejection. Eight mutants are rejected: four partial restorations, a changed condition, a changed throw class, a changed throw statement, and unreviewed wording. The accepted adapter is idempotent. `adapter-guards.json` retains each rejection. Early guard-development failures are preserved as preliminary failures in the raw archive.

Both full applies exit 0. The baseline apply restores the verified keyed product; the candidate is a fresh complete apply through every adapter and product packaging. Both normal `npm run build` executions exit 0. The candidate full apply took 1850.920 seconds while other proofs shared the machine. Its original generic 25-minute supervising parent was temporarily paused while a separate 900-second finishing controller monitored the progressing child, then resumed to reap and report that child's actual exit 0. `apply-controller-transfer.json` retains the handoff. This is not described as a sub-25-minute apply. The post-apply checkpoint is `9fe9d71cbb885469409efbf39672f73c31bae200`; current adapt.cjs and pins.json are byte-identical to that snapshot.

Only debug.ts differs among all 746 source files. All 10 normal emitted JavaScript files were compared. Exactly three files differ, each at four callback message expressions. Every other JavaScript byte is identical. Both complete emitted JavaScript sets and public declarations are in the before/after-javascript archives. All emitted hunks follow:

```diff
--- before/_tsc.js
+++ after/_tsc.js
@@ -1351,7 +1351,7 @@
       assert(
         test === void 0 || every(nodes, test),
         message || "Unexpected node.",
-        () => `Node array did not pass test '${getFunctionName(test)}'.`,
+        () => `Node array did not pass test.`,
         stackCrawlMark || assertEachNode
       );
     }
@@ -1362,7 +1362,7 @@
       assert(
         node !== void 0 && (test === void 0 || test(node)),
         message || "Unexpected node.",
-        () => `Node ${formatSyntaxKind(node == null ? void 0 : node.kind)} did not pass test '${getFunctionName(test)}'.`,
+        () => `Node ${formatSyntaxKind(node == null ? void 0 : node.kind)} did not pass test.`,
         stackCrawlMark || assertNode
       );
     }
@@ -1373,7 +1373,7 @@
       assert(
         node === void 0 || test === void 0 || !test(node),
         message || "Unexpected node.",
-        () => `Node ${formatSyntaxKind(node.kind)} should not have passed test '${getFunctionName(test)}'.`,
+        () => `Node ${formatSyntaxKind(node.kind)} should not have passed test.`,
         stackCrawlMark || assertNotNode
       );
     }
@@ -1384,7 +1384,7 @@
       assert(
         test === void 0 || node === void 0 || test(node),
         message || "Unexpected node.",
-        () => `Node ${formatSyntaxKind(node == null ? void 0 : node.kind)} did not pass test '${getFunctionName(test)}'.`,
+        () => `Node ${formatSyntaxKind(node == null ? void 0 : node.kind)} did not pass test.`,
         stackCrawlMark || assertOptionalNode
       );
     }
--- before/run.js
+++ after/run.js
@@ -1685,7 +1685,7 @@
           assert3(
             test === void 0 || every(nodes, test),
             message || "Unexpected node.",
-            () => `Node array did not pass test '${getFunctionName(test)}'.`,
+            () => `Node array did not pass test.`,
             stackCrawlMark || assertEachNode
           );
         }
@@ -1696,7 +1696,7 @@
           assert3(
             node !== void 0 && (test === void 0 || test(node)),
             message || "Unexpected node.",
-            () => `Node ${formatSyntaxKind(node == null ? void 0 : node.kind)} did not pass test '${getFunctionName(test)}'.`,
+            () => `Node ${formatSyntaxKind(node == null ? void 0 : node.kind)} did not pass test.`,
             stackCrawlMark || assertNode
           );
         }
@@ -1707,7 +1707,7 @@
           assert3(
             node === void 0 || test === void 0 || !test(node),
             message || "Unexpected node.",
-            () => `Node ${formatSyntaxKind(node.kind)} should not have passed test '${getFunctionName(test)}'.`,
+            () => `Node ${formatSyntaxKind(node.kind)} should not have passed test.`,
             stackCrawlMark || assertNotNode
           );
         }
@@ -1718,7 +1718,7 @@
           assert3(
             test === void 0 || node === void 0 || test(node),
             message || "Unexpected node.",
-            () => `Node ${formatSyntaxKind(node == null ? void 0 : node.kind)} did not pass test '${getFunctionName(test)}'.`,
+            () => `Node ${formatSyntaxKind(node == null ? void 0 : node.kind)} did not pass test.`,
             stackCrawlMark || assertOptionalNode
           );
         }
--- before/typescript.js
+++ after/typescript.js
@@ -3926,7 +3926,7 @@
       assert(
         test === void 0 || every(nodes, test),
         message || "Unexpected node.",
-        () => `Node array did not pass test '${getFunctionName(test)}'.`,
+        () => `Node array did not pass test.`,
         stackCrawlMark || assertEachNode
       );
     }
@@ -3937,7 +3937,7 @@
       assert(
         node !== void 0 && (test === void 0 || test(node)),
         message || "Unexpected node.",
-        () => `Node ${formatSyntaxKind(node == null ? void 0 : node.kind)} did not pass test '${getFunctionName(test)}'.`,
+        () => `Node ${formatSyntaxKind(node == null ? void 0 : node.kind)} did not pass test.`,
         stackCrawlMark || assertNode
       );
     }
@@ -3948,7 +3948,7 @@
       assert(
         node === void 0 || test === void 0 || !test(node),
         message || "Unexpected node.",
-        () => `Node ${formatSyntaxKind(node.kind)} should not have passed test '${getFunctionName(test)}'.`,
+        () => `Node ${formatSyntaxKind(node.kind)} should not have passed test.`,
         stackCrawlMark || assertNotNode
       );
     }
@@ -3959,7 +3959,7 @@
       assert(
         test === void 0 || node === void 0 || test(node),
         message || "Unexpected node.",
-        () => `Node ${formatSyntaxKind(node == null ? void 0 : node.kind)} did not pass test '${getFunctionName(test)}'.`,
+        () => `Node ${formatSyntaxKind(node == null ? void 0 : node.kind)} did not pass test.`,
         stackCrawlMark || assertOptionalNode
       );
     }
```

Both public files are byte-identical:

| File | Bytes | SHA256 |
| --- | ---: | --- |
| typescript.d.ts | 593502 | `c4edc2ad37cf1afeb4704ccb4bacf7c76b9f538a8e5685e4b276e2e7dd52796b` |
| tsserverlibrary.d.ts | 866 | `cee9e9bf9a8f67388ee878afd5e172189421bae6625076b774d68563d036fbfd` |

Both `adamic types` commands load all 79 sorted compiler roots together, exit 0 and report zero checker diagnostics. Source changes are exactly the four pinned failure expressions.

## 3. Full census and restoring mutant

The census tool is unchanged. The same meter-generated guarded overlay, built from the pinned compiler, ran over the compiler roots with only adapted debug.ts selected. Every run uses `LATENT_FULL=1`, `LATENT_ONLY_FILE=<tree>/src/compiler/debug.ts`, `LATENT_ASSERT_NO_OUTPUT=1`, and LATENT_TIMES; all other inherited LATENT flags are removed. Ordinary Load and Lower output guards execute in the driver. Only physical generated Go filenames gain `.txt`, so every retained Go file is `.go.txt`. No first-error guard or tool change was added.

All three runs have 105 units and zero checker diagnostics. The raw legacy checker-rejected-program label is retained, not interpreted as a rejection in these runs.

| Run | Child exit | Seconds | Raw findings |
| --- | ---: | ---: | ---: |
| before | 0 | 143.434 | 112 |
| after | 0 | 102.840 | 108 |
| mutant | 0 | 59.106 | 109 |

Exactly these four refusal-scan sites disappear: `267:72`, `280:98`, `293:108`, `307:98`. The exact removed reason is:

> a value of type (node: Node) => boolean seen as AnyFunction, which can write void where boolean is read

No location/kind/exact-reason/phase site is added, including reached dependency sites. Boundary continuation records are retained raw but excluded from site counting, as in the preceding census report. Distinct non-Boundary sites are 56 before and 52 after, including three in core.ts; physically in debug.ts the counts are 53 and 49.

The restoring mutant is an isolated after-src copy. Only assertEachNode's message is restored, byte for byte. Its full census brings back the exact Refused reason at debug.ts:267:72 in the refusal scan. The other three callbacks remain adapted. The patch, all raw streams, unit timings and child command results are retained.

## 4. getFunctionName stays; handoff to Adamic

getFunctionName remains byte-identical in source and emitted JavaScript. The selected census still attempts its standalone unit, named **getFunctionName**. No other recording unit in this debug.ts-only run reaches its body. Its existing findings remain:

| Location | Phase | Kind | Exact reason |
| --- | --- | --- | --- |
| debug.ts:355:26 | refusal_scan | Refused | a method read as a value (toString would lose its object, and this with it) |
| debug.ts:352:20 | lowering | Refused | inherited library member name read as an own field |
| debug.ts:355:26 | lowering | Refused | inherited library member call read as an own field |
| debug.ts:356:60 | lowering | NotYet | reading text |
| debug.ts:357:20 | lowering | NotYet | reading match |

This remaining getFunctionName unit goes to Adamic. It is not repaired or removed here. The four failure-message callers no longer call it.

## 5. Ledger and paired replay

The single appended LEDGER.md line records the ruled JavaScript-changing adaptation. README.md labels all four source lines JavaScript-changing.

The paired replay runs all 96 adapters in order on disposable source/root-metadata copies, re-linking the same locked dependencies. Generated built output, .git and test-baseline data are omitted from these source replay copies. Each adapter's actual exit, complete combined stdout/stderr, filesystem calls and changed-file contents are retained, and execution continues after existing refusals. Every final source change is journaled. Only disposable tree prefixes in output are normalized.

All 95 other adapter records are exactly identical, including exits, refusals and writes. This adapter alone differs: exit 0 and one debug.ts write with four edits on before, exit 0 and zero edits/writes on after. `paired-replay-comparison.json` and `paired-replay.diff` retain this comparison. The raw before/after-replay archives preserve all records.

## Evidence and reproduction

`PROOF.json` records every gate. `ARCHIVES.json` lists every raw archive and its hash. `SHA256SUMS` covers the final files. All archives are below 20 MB and the directory is below 120 files. `raw-proof.tar.gz` includes complete command logs/results, 79-root type output, all three unmodified census streams and timings, original/adapted/mutant debug.ts bytes, controls, preliminary failures and controller history. Raw lane piece archives and replay archives retain their own measurements separately. There is no report of an uncompleted proof as passing.

Start with the pinned source/Node toolchain and `source /workspace/adamic-tools/env.sh`; run cloud/setup.sh and locked npm installs for stage3/api and stage3/typescript. Supply ADAMIC_TYPESCRIPT_SOURCE at the pinned source commit. Produce before with the base checkout's full stage3/apply.sh and after with this checkout's full stage3/apply.sh. The scripts retain the actual /workspace paths and commands used, with outputs under /workspace/scratch/debug-messages.

Run work.py build/types for each side; witness.cjs compares both completed sources. instrument-built.cjs creates private measurement products; preload.cjs records callback counters. lane-pieces.py before 8 can be repeated until its accepted summary is complete, then reconcile.py reopens all pieces. replay.py runs each side and compare-replay.py checks the pair. census.py builds the guarded meter overlay and runs before/after/mutant. followup.py retains the exact restoring mutant construction and final command sequence. compare.py performs build and census comparisons. finalize.py packages the completed measurements. Every supervising command records the actual child exit; timeout and stopped attempts are explicitly labeled.
