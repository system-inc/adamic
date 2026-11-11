# assertEachNode proof and outstanding compiler handoff

Task 8n7eet3, site 2, Adamic Nov 10 22:01Z.
Final base: `337cf6917f7231422b5e168bc215888752a9f194`. Initial preparation: `e24f2600a5d8a8521c5128827cad395d47c81c81`.
The audited fast-forward changes no stage3 input. Pinned TypeScript:
`050880ce59e30b356b686bd3144efe24f875ebc8`; Node v24.19.0; Go 1.27.1.

**Compiler limitation:** base main lacks `internal/lower/overload_callback_use.go`.
The original and loop forms both stop at the existing overload-1 callback
contract. The requested causal restoring delta and predicted for-of binding stop
cannot be demonstrated on this main. They are not claimed as passing. The source
adaptation, Node behavior checks and JavaScript scope checks are complete; the
backend change must be paired with Optimizer.

## 1. Whole-lane admission count

`stage3/lane/prepare_shards.py` completed apply, locked install, normal build and
harness preparation, child exit 0, in 1126.257 seconds. Then every generated
TestShard_ inventory unit ran through the ordinary shard.py path, as road (b)'s
section 1 does. Reconciliation reopens the raw process rows and ordinary guard
reports and proves a bijection with all generated shard functions.

| Observation | Result |
| --- | ---: |
| Accepted units, exactly once | 398 |
| Node process records | 797 |
| Pass / expected failure / skip | 106366 / 1 / 0 |
| assertEachNode calls with test defined | 1,086,412 |
| Calls whose test source can see a second argument | 0 |
| Throws from inside test | 0 |
| Distinct callback sources | 26 |
| `(n, i)` positive control, before / after | 1 / 1 |
| Throw control | 1, identical thrown object |

Private copies instrument the entry of each assertEachNode in _tsc.js,
typescript.js and run.js. Every JavaScript wrapper is copied, preventing Node's
realpath resolution from escaping to an uninstrumented bundle. The preload
propagates to child processes and VM contexts and writes explicit zero rows.
It records Function.prototype.toString source and parses parameters, defaults,
rest and arguments reads; it never uses Function.length. Every observed source
has one plain node parameter and no arguments read. The wrapper counts throws
and rethrows the same object. Normal build outputs are not instrumented.

All ordinary identity, baseline, API and test-count guards pass. The one expected
public-API failure remains. No candidate whole-lane run is claimed. The before
and after positive controls are separate from the lane's admission count.
Raw reconciliation, 20 raw shard archives and all 797 process rows are on the
[evidence branch](https://github.com/system-inc/adamic/tree/typescript/each-node-evidence/each-node-evidence); their hashes are indexed in ARCHIVES.json and that
branch's SHA256SUMS.

## 2. Static census across all src

The checker-backed scan covers 291 relevant calls, 26 concrete predicates,
six forwarding parameters, and zero unresolved expressions. It follows
visitNodes, visitArray and visitArrayWorker and NodesVisitor calls, including the
three custom forwarding wrappers. Every concrete declaration is:

| Predicate | Declared parameters | Pinned source location |
| --- | --- | --- |
| `isArrayBindingElement` | `node: Node` | `src/compiler/utilitiesPublic.ts:1849:1` |
| `isArrayBindingOrAssignmentElement` | `node: Node` | `src/compiler/utilitiesPublic.ts:1932:1` |
| `isBindingElement` | `node: Node` | `src/compiler/factory/nodeTests.ts:570:1` |
| `isCaseOrDefaultClause` | `node: Node` | `src/compiler/utilitiesPublic.ts:2504:1` |
| `isClassElement` | `node: Node` | `src/compiler/utilitiesPublic.ts:1732:1` |
| `isEnumMember` | `node: Node` | `src/compiler/factory/nodeTests.ts:1009:1` |
| `isExportSpecifier` | `node: Node` | `src/compiler/factory/nodeTests.ts:898:1` |
| `isExpression` | `node: Node` | `src/compiler/utilitiesPublic.ts:2086:1` |
| `isExpressionWithTypeArguments` | `node: Node` | `src/compiler/factory/nodeTests.ts:672:1` |
| `isHeritageClause` | `node: Node` | `src/compiler/factory/nodeTests.ts:985:1` |
| `isImportAttribute` | `node: Node` | `src/compiler/factory/nodeTests.ts:866:1` |
| `isImportSpecifier` | `node: Node` | `src/compiler/factory/nodeTests.ts:882:1` |
| `isJsxAttributeLike` | `node: Node` | `src/compiler/utilitiesPublic.ts:2477:1` |
| `isJsxChild` | `node: Node` | `src/compiler/utilitiesPublic.ts:2468:1` |
| `isModifier` | `node: Node` | `src/compiler/utilitiesPublic.ts:1647:1` |
| `isModifierLike` | `node: Node` | `src/compiler/utilitiesPublic.ts:1778:1` |
| `isObjectBindingOrAssignmentElement` | `node: Node` | `src/compiler/utilitiesPublic.ts:1905:1` |
| `isObjectLiteralElementLike` | `node: Node` | `src/compiler/utilitiesPublic.ts:1798:1` |
| `isParameter` | `node: Node` | `src/compiler/factory/nodeTests.ts:408:1` |
| `isStatement` | `node: Node` | `src/compiler/utilitiesPublic.ts:2419:1` |
| `isTemplateLiteralTypeSpan` | `node: Node` | `src/compiler/factory/nodeTests.ts:552:1` |
| `isTemplateSpan` | `node: Node` | `src/compiler/factory/nodeTests.ts:706:1` |
| `isTypeElement` | `node: Node` | `src/compiler/utilitiesPublic.ts:1782:1` |
| `isTypeNode` | `node: Node` | `src/compiler/utilitiesPublic.ts:1815:1` |
| `isTypeParameterDeclaration` | `node: Node` | `src/compiler/factory/nodeTests.ts:403:1` |
| `isVariableDeclaration` | `node: Node` | `src/compiler/factory/nodeTests.ts:792:1` |

The six forwarded test parameters are visitNodes.test, visitArray.test,
visitArrayWorker.test, checker.ts's deepCloneOrReuseNodes.test,
expressionToTypeNode.ts's visitNodesWithoutCopyingPositions.test, and
textChanges.ts's assignPositionsToNodeArray.test. Each declares
`((node: Node) => boolean) | undefined`; inbound NodesVisitor calls resolve to
the same inventory. Full call sites and declarations are in static-census.json.

## 3. Exact source, witness and guards

Only compiler/debug.ts differs among all 747 src files. The before pin is exactly
road (b)'s after pin. The gate call and key name are unchanged; the one assert
keeps its message, lazy callback and stack mark byte for byte. Source line endings
remain CRLF. Both declaration-emission modes match. every's only non-import
reference was the replaced call, so its import is removed.

The completed before/after debug.ts texts run at assertion levels 0, 1, 2 and 3.
All 32 paired cases pass: undefined nodes, empty array, a factory-created
NodeArray, all passing, first failure, middle failure, undefined test, and a throw
inside test. Recorded call sequences stop at the first failure. Error constructor,
message and actual throw point agree. A test throw loses its every stack frame.
Four additional two-parameter cases show index 0/1 before and undefined after.
Both built-artifact source counters still classify that test once.

The own adapter is idempotent, and road (b) recognizes both completed states.
Five mutants refuse: before body without import, after body with import, drifted
loop, road (b)'s drifted finished loop, and loop form with another message restored.
The explicit road (b) drift refusal is `unreviewed owner: assertEachNode`.
No rejection writes a file. Guard receipts are in guards-result.json.

Stock TypeScript with noUnusedLocals enabled adds exactly one diagnostic when
only every's import is restored: `TS6133: 'every' is declared but its value is
never read.` Removing it clears that added diagnostic. The actual pinned config
leaves noUnusedLocals unset; the explicit probe also exposes 60 pre-existing
unused diagnostics, retained raw. No claim is made that the normal build enables
that option. The normal candidate npm run build exits 0.

## 4. Main full census and exact restoring mutant

The unchanged full meter is rebuilt from final base main. Production Load/Lower
output guards remain active. Physical generated Go evidence names end in .go.txt.
Each run uses LATENT_FULL=1, LATENT_ONLY_FILE=<tree>/src/compiler/debug.ts,
LATENT_ASSERT_NO_OUTPUT=1 and a timings path, after clearing inherited LATENT flags.

The restoring source is an isolated candidate src copy with the entire original
owner and its required every import restored. Its debug.ts equals before byte for
byte; its patch is restoring-mutant.patch. The separately retained census-mutant
copy is byte-identical to that source.

| Run | Child exit | Units | Findings | Checker diagnostics | Observed stop |
| --- | ---: | ---: | ---: | ---: | --- |
| Restoring mutant | 0 | 81 | 22 | 0 | debug.ts:257:86, overload 1 test contract |
| Loop on main | 0 | 81 | 22 | 0 | debug.ts:256:86, overload 1 test contract |

Exact reason in both streams:

> overload 1 of assertEachNode parameter test cannot be served by implementation parameter test

It appears in two recording units in each run. Main short-circuits at overload 1;
there is no measured evidence for overloads 2 through 4, nor a for-of element-type
stop on this base. The old refusal is present in the restored source, but it never
left on main, so this is not presented as a successful removal/reintroduction
experiment. Optimizer's rule 1 and for-of extension remain the required handoff.

The legacy `measured on a checker-rejected program` label is retained verbatim;
both headers have an empty diagnostics array. No usable compiler output is emitted.
Raw streams, timings, command flags, binary hash and compiler base are external.

## 5. Every emitted JavaScript hunk and runtime exports

All 10 normal JavaScript files were compared. Exactly three differ, once each,
inside assertEachNode. After masking only that body, every remaining byte is
identical; identifier-token normalization also matches. There are **0** outside
hunks and **0** esbuild renames to excuse. The seven other JavaScript files are
byte-identical. Both public declaration files are byte-identical.
All **2256** runtime ts export keys match; sorted key SHA256:
`8125408be73cd4b1b940a7ac84511eaebbd209cff601fbbc8e539aa1870fe998`.

The initial symbol-binding comparison hit Node's default stack limit. Its larger
stack retry was stopped as unnecessary after inspecting the three body-only
hunks. The final comparison uses the stronger outside-byte check plus identifier
normalization and exits 0. All original child exits are retained as preliminary
measurement history, not passing results.

Every emitted hunk follows:

```diff
--- before/_tsc.js
+++ after/_tsc.js
@@ -1354,8 +1354,17 @@
   Debug2.assertNever = assertNever;
   function assertEachNode(nodes, test, message, stackCrawlMark) {
     if (shouldAssertFunction(1 /* Normal */, "assertEachNode")) {
+      let passed = true;
+      if (test !== void 0 && nodes !== void 0) {
+        for (const n of nodes) {
+          if (!test(n)) {
+            passed = false;
+            break;
+          }
+        }
+      }
       assert(
-        test === void 0 || every(nodes, test),
+        passed,
         message || "Unexpected node.",
         () => `Node array did not pass test.`,
         stackCrawlMark || assertEachNode
--- before/run.js
+++ after/run.js
@@ -1688,8 +1688,17 @@
       Debug3.assertNever = assertNever;
       function assertEachNode(nodes, test, message, stackCrawlMark) {
         if (shouldAssertFunction(1 /* Normal */, "assertEachNode")) {
+          let passed = true;
+          if (test !== void 0 && nodes !== void 0) {
+            for (const n of nodes) {
+              if (!test(n)) {
+                passed = false;
+                break;
+              }
+            }
+          }
           assert3(
-            test === void 0 || every(nodes, test),
+            passed,
             message || "Unexpected node.",
             () => `Node array did not pass test.`,
             stackCrawlMark || assertEachNode
--- before/typescript.js
+++ after/typescript.js
@@ -3931,8 +3931,17 @@
   Debug2.assertNever = assertNever;
   function assertEachNode(nodes, test, message, stackCrawlMark) {
     if (shouldAssertFunction(1 /* Normal */, "assertEachNode")) {
+      let passed = true;
+      if (test !== void 0 && nodes !== void 0) {
+        for (const n of nodes) {
+          if (!test(n)) {
+            passed = false;
+            break;
+          }
+        }
+      }
       assert(
-        test === void 0 || every(nodes, test),
+        passed,
         message || "Unexpected node.",
         () => `Node array did not pass test.`,
         stackCrawlMark || assertEachNode
```

## 6. Final composition acceptance and reproduction

The final acceptance runs `bash stage3/apply.sh <fresh-final-tree>` after this
directory's SHA256SUMS is frozen. The fresh tree is normally built and its entire
src inventory, JavaScript set, public declarations and export keys are compared
to this measured candidate. Its child-exit receipt and final verification are
retained externally as final-apply-result.json, final-build-result.json and
final-verify.json. Publishing is conditioned on those checks passing. Keeping
those logs external avoids changing the key of the product being verified.

Setup: source /workspace/adamic-tools/env.sh; bash cloud/setup.sh;
npm ci --prefix stage3/api; locked stage3/typescript dependencies; export
ADAMIC_TYPESCRIPT_SOURCE to the pinned 050880ce checkout. Evidence controllers
retain exact commands, paths and every child's own exit. prepare.py and lane.py
produce the baseline; reconcile.py independently reopens it. static.cjs records
the source census. guards.cjs and witness.cjs run against completed trees with
NODE_PATH=stage3/api/node_modules. proofs.py, resume-candidate-final.py,
compare-built.cjs and unused-import.cjs retain the normal candidate checks.
build-census.py and census-run.py produce the guarded main streams. The witness
and exact source/restore patches are retained here; raw counts, logs, census
streams, JavaScript archives and controllers are only on the evidence branch.

LEDGER.md and rulings.md record the JavaScript-changing adaptation and its
zero-count condition. No upstream issue was filed. No Go lowering success or
for-of-specific main stop is claimed.
