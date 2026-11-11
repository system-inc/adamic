# Captured Union compiler proof — #n7sx9gz

HEAD: `54397e8f7ff1c896343d8e62c8c94ef87e5b6991`; base/main: `efe887763f26e96653546fae904d57539139f7ae`. Proof executed on 2026-11-11 UTC. No compiler fixes were made. Only this report is committed.

## Decisive results

HEAD does **not** pass its lowering pins: `cycle.ts` and `cycle.a` fail the closure-count assertion, although both cells are graph cells. All 64 HEAD oracle backend subtests pass. M1 and M2 are detected at lowering; M3 is detected by both leak checks and counted builds for by_value, released, and alternating in both dialects. The full lower suite has only the same two cycle failures. The full IR suite passes.

- Full lower: exit 1; wall 2135.449 seconds.
- Full ir: exit 0; wall 171.380 seconds.
- Full oracle: exit 1; wall 5296.929 seconds.
- Admission delta: exit 3, verdict void (19 native compilation timeouts); diff_count=8, admitted=8, newly refused=0. Full reasons and both-revision isolated reruns appear in step 6.

The full oracle run had 15 environment/probe-budget failures: twelve missing-module cases passed on both SHAs after exposing the external pinned install, two probe wrappers passed alone on HEAD (counted also passed alone on main; determinism still timed out on main), and the WASI refusal test failed for missing WASI_SYSROOT on both. These do not turn the requested full-suite run into a pass.

## Setup and step 1

Loaded `/workspace/adamic-tools/env.sh` before work. Fetched `origin compiler/captured-union-3`, checked out the exact HEAD detached, and ran `git submodule update --init --recursive`; all exited 0. Initial checkout was clean. Cohere pin: `5360753e19e4417675a90e92a1823bea66b13918`; its TypeScript submodule: `d92d9bfee114c80be2c375d72edae966176e3a4f`. No AGENTS.md was present in the initial workspace.

Created `/tmp/adamic-gate` outside the repository for env.sh’s TMPDIR. Provisioned a separate upstream checkout at `/workspace/adamic-typescript-upstream`, commit `050880ce59e30b356b686bd3144efe24f875ebc8`, and exported `ADAMIC_TYPESCRIPT_SOURCE` to it. Copied stage3/api’s package.json and lockfile to `/workspace/adamic-proof-api` and ran npm ci there: TypeScript 6.0.3, @types/node 25.3.3, undici-types 7.18.2. Its node_modules/.bin precedes PATH; `tsc --version` printed `Version 6.0.3`.

All subsequent proof commands, npm, and admission-gate children inherited `NODE_EXTRA_CA_CERTS=/etc/ssl/certs/ca-certificates.crt`, `npm_config_cafile=/etc/ssl/certs/ca-certificates.crt`, and `npm_config_strict_ssl=true`. Heavy commands ran sequentially. Tests’ own internal scheduling and the gate’s requested two workers were left intact.

`gofmt -l internal/lower internal/oracle`: exit 0, empty stdout. `go vet ./internal/lower ./internal/oracle`: exit 0, empty output. `/usr/bin/time` was absent; an initial timing-wrapper attempt exited 127 without starting either pin command. Actual runs below used shell or monotonic wall timing.

## Step 2: HEAD pins

```text
go test ./internal/lower -run 'TestCapturedUnionLowersToOneCell|TestCapturedMaybeBooleanStillStops' -count=1 -v -timeout 30m
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestCapturedUnionAgrees' -count=1 -v -timeout 40m
```


Lower: exit 1, wall 48 seconds. Oracle: exit 0, wall 660 seconds.

| Pin | Lower | Graph-cell log (verbatim) | native | sanitized | counted | javascript |
|---|---|---|---|---|---|---|
| class.ts | PASS | captured_union_test.go:48: active: graph cell true | PASS | PASS | PASS | PASS |
| class.a | PASS | captured_union_test.go:48: active: graph cell true | PASS | PASS | PASS | PASS |
| object.ts | PASS | captured_union_test.go:48: active: graph cell true | PASS | PASS | PASS | PASS |
| object.a | PASS | captured_union_test.go:48: active: graph cell true | PASS | PASS | PASS | PASS |
| by_value.ts | PASS | captured_union_test.go:48: active: graph cell false | PASS | PASS | PASS | PASS |
| by_value.a | PASS | captured_union_test.go:48: active: graph cell false | PASS | PASS | PASS | PASS |
| released.ts | PASS | captured_union_test.go:48: active: graph cell false | PASS | PASS | PASS | PASS |
| released.a | PASS | captured_union_test.go:48: active: graph cell false | PASS | PASS | PASS | PASS |
| nested.ts | PASS | captured_union_test.go:48: extended: graph cell true | PASS | PASS | PASS | PASS |
| nested.a | PASS | captured_union_test.go:48: extended: graph cell true | PASS | PASS | PASS | PASS |
| alternating.ts | PASS | captured_union_test.go:48: active: graph cell false | PASS | PASS | PASS | PASS |
| alternating.a | PASS | captured_union_test.go:48: active: graph cell false | PASS | PASS | PASS | PASS |
| graph.ts | PASS | captured_union_test.go:48: active: graph cell true | PASS | PASS | PASS | PASS |
| graph.a | PASS | captured_union_test.go:48: active: graph cell true | PASS | PASS | PASS | PASS |
| cycle.ts | FAIL | captured_union_test.go:48: active: graph cell true | PASS | PASS | PASS | PASS |
| cycle.a | FAIL | captured_union_test.go:48: active: graph cell true | PASS | PASS | PASS | PASS |

`TestCapturedMaybeBooleanStillStops`: PASS.

Each failing cycle pin’s whole actual and expected assertion is, verbatim: `active: 1 closures carry its cell, want every one that reads or writes it`. Actual closure count = 1; the source asserts at least 2. Both logged `active: graph cell true`. There was no compiler lowering stop for these failures.

### Every HEAD oracle counts line

class.ts:
```text
captured_union_test.go:116: counts: allocations 21, frees 21, in regions 0
```


class.a:
```text
captured_union_test.go:116: counts: allocations 21, frees 21, in regions 0
```


object.ts:
```text
captured_union_test.go:116: counts: allocations 21, frees 21, in regions 0
```


object.a:
```text
captured_union_test.go:116: counts: allocations 21, frees 21, in regions 0
```


by_value.ts:
```text
captured_union_test.go:116: counts: allocations 35, frees 35, in regions 0
```


by_value.a:
```text
captured_union_test.go:116: counts: allocations 35, frees 35, in regions 0
```


released.ts:
```text
captured_union_test.go:116: counts: allocations 96, frees 96, in regions 0
```


released.a:
```text
captured_union_test.go:116: counts: allocations 96, frees 96, in regions 0
```


nested.ts:
```text
captured_union_test.go:116: counts: allocations 97, frees 97, in regions 0
```


nested.a:
```text
captured_union_test.go:116: counts: allocations 97, frees 97, in regions 0
```


alternating.ts:
```text
captured_union_test.go:116: counts: allocations 54, frees 54, in regions 0
```


alternating.a:
```text
captured_union_test.go:116: counts: allocations 54, frees 54, in regions 0
```


graph.ts:
```text
captured_union_test.go:116: counts: allocations 42, frees 42, in regions 0
```


graph.a:
```text
captured_union_test.go:116: counts: allocations 42, frees 42, in regions 0
```


cycle.ts:
```text
captured_union_test.go:116: counts: allocations 32, frees 32, in regions 0
```


cycle.a:
```text
captured_union_test.go:116: counts: allocations 32, frees 32, in regions 0
```


## Step 3: mutants

For each mutant the exact old text occurred once (asserted before the edit). Each ran the two step-2 commands unchanged, with ADAMIC_GATE_UNCACHED=1 for oracle. Each edit was restored, followed by a successful `git diff --exit-code` (exit 0). Original stderr and complete logged got values are preserved in the appendices below; aggregate parent failures are identified separately from leaf failures.

### M1: The guard restored

Lower: exit 1, wall 42.21891390199994 seconds. Oracle: exit 1, wall 35.530805670999825 seconds.

```text
diff --git a/internal/lower/locals.go b/internal/lower/locals.go
index a6a057992..b1452f72d 100644
--- a/internal/lower/locals.go
+++ b/internal/lower/locals.go
@@ -201,7 +201,7 @@ func (l *lowering) localOf(identifier *ast.Node, symbol *ast.Symbol) (int, bool)
 		if declared := l.result.Locals[local]; declared.Placeholder != "" && declared.Captured && declared.Preallocated && declared.Ready == 0 && l.unlowerable == nil {
 			l.unlowerable = l.notYet(identifier, "an unset-capable capture before its declaration without a separate declaration-readiness cell")
 		}
-		if declared := l.result.Locals[local]; declared.Captured && !boxedSlot(declared.Type) && l.unlowerable == nil {
+		if declared := l.result.Locals[local]; declared.Captured && slotless(declared.Type) && !(declared.Placeholder != "" && declared.Type == ir.Union) && l.unlowerable == nil {
 			// A Union is one counted heap reference whatever initializes it, a placeholder or a
 			// value (#n7sx9gz), and a cell holds one: each write fits its member into the Union, the
 			// cell releases what it held, and a read is the Union every closure shares. boolean |
```




| Red lower test/subtest | First error line |
|---|---|
| TestCapturedUnionLowersToOneCell | Aggregate failure; see failing child rows. |
| TestCapturedUnionLowersToOneCell/class.ts | captured_union_test.go:37: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellclass.ts4270929907/001/main.ts:13:7: stage 0 can't lower a union of differently held members variable a function value captures yet |
| TestCapturedUnionLowersToOneCell/released.a | captured_union_test.go:37: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellreleased.a3476599601/001/main.a:17:4: stage 0 can't lower a union of differently held members variable a function value captures yet |
| TestCapturedUnionLowersToOneCell/released.ts | captured_union_test.go:37: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellreleased.ts1751637487/001/main.ts:17:4: stage 0 can't lower a union of differently held members variable a function value captures yet |
| TestCapturedUnionLowersToOneCell/by_value.a | captured_union_test.go:37: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellby_value.a1674039208/001/main.a:20:4: stage 0 can't lower a union of differently held members variable a function value captures yet |
| TestCapturedUnionLowersToOneCell/nested.ts | captured_union_test.go:37: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellnested.ts3304963122/001/main.ts:8:5: stage 0 can't lower a union of differently held members variable a function value captures yet |
| TestCapturedUnionLowersToOneCell/cycle.a | captured_union_test.go:37: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellcycle.a2898997350/001/main.a:15:7: stage 0 can't lower a union of differently held members variable a function value captures yet |
| TestCapturedUnionLowersToOneCell/alternating.a | captured_union_test.go:37: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellalternating.a3778674745/001/main.a:17:4: stage 0 can't lower a union of differently held members variable a function value captures yet |
| TestCapturedUnionLowersToOneCell/class.a | captured_union_test.go:37: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellclass.a2067780351/001/main.a:13:7: stage 0 can't lower a union of differently held members variable a function value captures yet |
| TestCapturedUnionLowersToOneCell/by_value.ts | captured_union_test.go:37: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellby_value.ts447987668/001/main.ts:20:4: stage 0 can't lower a union of differently held members variable a function value captures yet |
| TestCapturedUnionLowersToOneCell/graph.a | captured_union_test.go:37: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellgraph.a1974754399/001/main.a:19:5: stage 0 can't lower a union of differently held members variable a function value captures yet |
| TestCapturedUnionLowersToOneCell/object.ts | captured_union_test.go:37: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellobject.ts222104710/001/main.ts:7:7: stage 0 can't lower a union of differently held members variable a function value captures yet |
| TestCapturedUnionLowersToOneCell/cycle.ts | captured_union_test.go:37: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellcycle.ts2530527128/001/main.ts:15:7: stage 0 can't lower a union of differently held members variable a function value captures yet |
| TestCapturedUnionLowersToOneCell/object.a | captured_union_test.go:37: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellobject.a497294124/001/main.a:7:7: stage 0 can't lower a union of differently held members variable a function value captures yet |
| TestCapturedUnionLowersToOneCell/graph.ts | captured_union_test.go:37: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellgraph.ts596596665/001/main.ts:19:5: stage 0 can't lower a union of differently held members variable a function value captures yet |
| TestCapturedUnionLowersToOneCell/nested.a | captured_union_test.go:37: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellnested.a1215474141/001/main.a:8:5: stage 0 can't lower a union of differently held members variable a function value captures yet |
| TestCapturedUnionLowersToOneCell/alternating.ts | captured_union_test.go:37: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellalternating.ts2570205261/001/main.ts:17:4: stage 0 can't lower a union of differently held members variable a function value captures yet |



| Red oracle test/subtest | First error line |
|---|---|
| TestCapturedUnionAgrees | Aggregate failure; see failing child rows. |
| TestCapturedUnionAgrees/by_value.ts | captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesby_value.ts496171518/001/by_value.ts:20:4: stage 0 can't lower a union of differently held members variable a function value captures yet |
| TestCapturedUnionAgrees/nested.ts | captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesnested.ts3612537006/001/nested.ts:8:5: stage 0 can't lower a union of differently held members variable a function value captures yet |
| TestCapturedUnionAgrees/released.ts | captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesreleased.ts3562262458/001/released.ts:17:4: stage 0 can't lower a union of differently held members variable a function value captures yet |
| TestCapturedUnionAgrees/graph.ts | captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesgraph.ts2505154369/001/graph.ts:19:5: stage 0 can't lower a union of differently held members variable a function value captures yet |
| TestCapturedUnionAgrees/class.ts | captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesclass.ts2093953919/001/class.ts:13:7: stage 0 can't lower a union of differently held members variable a function value captures yet |
| TestCapturedUnionAgrees/object.ts | captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesobject.ts2148970366/001/object.ts:7:7: stage 0 can't lower a union of differently held members variable a function value captures yet |
| TestCapturedUnionAgrees/by_value.a | captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesby_value.a1770380039/001/by_value.a:20:4: stage 0 can't lower a union of differently held members variable a function value captures yet |
| TestCapturedUnionAgrees/released.a | captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesreleased.a2049496842/001/released.a:17:4: stage 0 can't lower a union of differently held members variable a function value captures yet |
| TestCapturedUnionAgrees/alternating.a | captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesalternating.a4119439711/001/alternating.a:17:4: stage 0 can't lower a union of differently held members variable a function value captures yet |
| TestCapturedUnionAgrees/cycle.ts | captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreescycle.ts2353261059/001/cycle.ts:15:7: stage 0 can't lower a union of differently held members variable a function value captures yet |
| TestCapturedUnionAgrees/object.a | captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesobject.a2652515123/001/object.a:7:7: stage 0 can't lower a union of differently held members variable a function value captures yet |
| TestCapturedUnionAgrees/alternating.ts | captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesalternating.ts2740834732/001/alternating.ts:17:4: stage 0 can't lower a union of differently held members variable a function value captures yet |
| TestCapturedUnionAgrees/graph.a | captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesgraph.a1602122324/001/graph.a:19:5: stage 0 can't lower a union of differently held members variable a function value captures yet |
| TestCapturedUnionAgrees/cycle.a | captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreescycle.a3363162970/001/cycle.a:15:7: stage 0 can't lower a union of differently held members variable a function value captures yet |
| TestCapturedUnionAgrees/nested.a | captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesnested.a215508565/001/nested.a:8:5: stage 0 can't lower a union of differently held members variable a function value captures yet |
| TestCapturedUnionAgrees/class.a | captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesclass.a3835253819/001/class.a:13:7: stage 0 can't lower a union of differently held members variable a function value captures yet |

### M2: The cell held as a plain reference

Lower: exit 1, wall 25.26475540299998 seconds. Oracle: exit 1, wall 39.59347280299994 seconds.

```text
diff --git a/internal/lower/locals.go b/internal/lower/locals.go
index a6a057992..c6c466fdc 100644
--- a/internal/lower/locals.go
+++ b/internal/lower/locals.go
@@ -201,6 +201,9 @@ func (l *lowering) localOf(identifier *ast.Node, symbol *ast.Symbol) (int, bool)
 		if declared := l.result.Locals[local]; declared.Placeholder != "" && declared.Captured && declared.Preallocated && declared.Ready == 0 && l.unlowerable == nil {
 			l.unlowerable = l.notYet(identifier, "an unset-capable capture before its declaration without a separate declaration-readiness cell")
 		}
+		if declared := l.result.Locals[local]; declared.Captured && declared.Type == ir.Union && declared.Placeholder == "" {
+			l.result.Locals[local].Type = ir.Object
+		}
 		if declared := l.result.Locals[local]; declared.Captured && !boxedSlot(declared.Type) && l.unlowerable == nil {
 			// A Union is one counted heap reference whatever initializes it, a placeholder or a
 			// value (#n7sx9gz), and a cell holds one: each write fits its member into the Union, the
```




| Red lower test/subtest | First error line |
|---|---|
| TestCapturedUnionLowersToOneCell | Aggregate failure; see failing child rows. |
| TestCapturedUnionLowersToOneCell/by_value.ts | captured_union_test.go:37: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellby_value.ts876155215/001/main.ts:30:35: stage 0 can't lower a BinaryExpression with a value and a string yet |
| TestCapturedUnionLowersToOneCell/class.ts | captured_union_test.go:37: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellclass.ts2147513667/001/main.ts:16:7: stage 0 can't lower a BinaryExpression with a value and a string yet |
| TestCapturedUnionLowersToOneCell/graph.ts | captured_union_test.go:37: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellgraph.ts1315285737/001/main.ts:33:8: stage 0 can't lower a BinaryExpression with a value and a string yet |
| TestCapturedUnionLowersToOneCell/nested.ts | captured_union_test.go:37: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellnested.ts3373742230/001/main.ts:25:22: stage 0 can't lower a template interpolating an object, an array, a map, a function or undefined yet |
| TestCapturedUnionLowersToOneCell/object.a | captured_union_test.go:37: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellobject.a1240869268/001/main.a:10:7: stage 0 can't lower a BinaryExpression with a value and a string yet |
| TestCapturedUnionLowersToOneCell/object.ts | captured_union_test.go:37: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellobject.ts2793319395/001/main.ts:10:7: stage 0 can't lower a BinaryExpression with a value and a string yet |
| TestCapturedUnionLowersToOneCell/cycle.ts | captured_union_test.go:37: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellcycle.ts3792324500/001/main.ts:18:7: stage 0 can't lower a BinaryExpression with a value and a string yet |
| TestCapturedUnionLowersToOneCell/class.a | captured_union_test.go:37: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellclass.a474699678/001/main.a:16:7: stage 0 can't lower a BinaryExpression with a value and a string yet |
| TestCapturedUnionLowersToOneCell/cycle.a | captured_union_test.go:37: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellcycle.a4147608424/001/main.a:18:7: stage 0 can't lower a BinaryExpression with a value and a string yet |
| TestCapturedUnionLowersToOneCell/released.ts | captured_union_test.go:37: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellreleased.ts3930721773/001/main.ts:39:7: stage 0 can't lower a BinaryExpression with a value and a string yet |
| TestCapturedUnionLowersToOneCell/released.a | captured_union_test.go:37: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellreleased.a1595996122/001/main.a:39:7: stage 0 can't lower a BinaryExpression with a value and a string yet |
| TestCapturedUnionLowersToOneCell/by_value.a | captured_union_test.go:37: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellby_value.a3765384823/001/main.a:30:35: stage 0 can't lower a BinaryExpression with a value and a string yet |
| TestCapturedUnionLowersToOneCell/graph.a | captured_union_test.go:37: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellgraph.a1709038358/001/main.a:33:8: stage 0 can't lower a BinaryExpression with a value and a string yet |
| TestCapturedUnionLowersToOneCell/alternating.ts | captured_union_test.go:37: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellalternating.ts2835279116/001/main.ts:28:7: stage 0 can't lower a BinaryExpression with a value and a string yet |
| TestCapturedUnionLowersToOneCell/alternating.a | captured_union_test.go:37: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellalternating.a3897242562/001/main.a:28:7: stage 0 can't lower a BinaryExpression with a value and a string yet |
| TestCapturedUnionLowersToOneCell/nested.a | captured_union_test.go:37: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellnested.a443927910/001/main.a:25:22: stage 0 can't lower a template interpolating an object, an array, a map, a function or undefined yet |



| Red oracle test/subtest | First error line |
|---|---|
| TestCapturedUnionAgrees | Aggregate failure; see failing child rows. |
| TestCapturedUnionAgrees/class.ts | captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesclass.ts581344860/001/class.ts:16:7: stage 0 can't lower a BinaryExpression with a value and a string yet |
| TestCapturedUnionAgrees/nested.ts | captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesnested.ts2262350609/001/nested.ts:25:22: stage 0 can't lower a template interpolating an object, an array, a map, a function or undefined yet |
| TestCapturedUnionAgrees/by_value.ts | captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesby_value.ts3955907946/001/by_value.ts:30:35: stage 0 can't lower a BinaryExpression with a value and a string yet |
| TestCapturedUnionAgrees/released.ts | captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesreleased.ts31105505/001/released.ts:39:7: stage 0 can't lower a BinaryExpression with a value and a string yet |
| TestCapturedUnionAgrees/graph.ts | captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesgraph.ts2544236627/001/graph.ts:33:8: stage 0 can't lower a BinaryExpression with a value and a string yet |
| TestCapturedUnionAgrees/object.ts | captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesobject.ts2854214631/001/object.ts:10:7: stage 0 can't lower a BinaryExpression with a value and a string yet |
| TestCapturedUnionAgrees/object.a | captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesobject.a1864731276/001/object.a:10:7: stage 0 can't lower a BinaryExpression with a value and a string yet |
| TestCapturedUnionAgrees/class.a | captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesclass.a2187697731/001/class.a:16:7: stage 0 can't lower a BinaryExpression with a value and a string yet |
| TestCapturedUnionAgrees/alternating.a | captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesalternating.a2537336030/001/alternating.a:28:7: stage 0 can't lower a BinaryExpression with a value and a string yet |
| TestCapturedUnionAgrees/alternating.ts | captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesalternating.ts4265678876/001/alternating.ts:28:7: stage 0 can't lower a BinaryExpression with a value and a string yet |
| TestCapturedUnionAgrees/cycle.a | captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreescycle.a29200094/001/cycle.a:18:7: stage 0 can't lower a BinaryExpression with a value and a string yet |
| TestCapturedUnionAgrees/released.a | captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesreleased.a1461874531/001/released.a:39:7: stage 0 can't lower a BinaryExpression with a value and a string yet |
| TestCapturedUnionAgrees/nested.a | captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesnested.a2925901490/001/nested.a:25:22: stage 0 can't lower a template interpolating an object, an array, a map, a function or undefined yet |
| TestCapturedUnionAgrees/graph.a | captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesgraph.a1774914312/001/graph.a:33:8: stage 0 can't lower a BinaryExpression with a value and a string yet |
| TestCapturedUnionAgrees/cycle.ts | captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreescycle.ts496804564/001/cycle.ts:18:7: stage 0 can't lower a BinaryExpression with a value and a string yet |
| TestCapturedUnionAgrees/by_value.a | captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesby_value.a2804373078/001/by_value.a:30:35: stage 0 can't lower a BinaryExpression with a value and a string yet |

### M3: The overwritten value kept

Lower: exit 1, wall 43.28693612400002 seconds. Oracle: exit 1, wall 52.30952770199997 seconds.

```text
diff --git a/internal/native/emit_locals.go b/internal/native/emit_locals.go
index 775b738bd..e9c8e39b2 100644
--- a/internal/native/emit_locals.go
+++ b/internal/native/emit_locals.go
@@ -60,7 +60,7 @@ func (e *emitter) store(local int, value string, owned bool) {
 		if _, framed := e.asyncSlots[local]; framed {
 			e.line("if (%s != NULL) adamic_release(%s);", old, old)
 		} else {
-			e.line("adamic_release(%s);", old)
+			e.line("(void)%s;", old)
 		}
 		return
 	}
```




| Red lower test/subtest | First error line |
|---|---|
| TestCapturedUnionLowersToOneCell | Aggregate failure; see failing child rows. |
| TestCapturedUnionLowersToOneCell/cycle.ts | captured_union_test.go:59: active: 1 closures carry its cell, want every one that reads or writes it |
| TestCapturedUnionLowersToOneCell/cycle.a | captured_union_test.go:59: active: 1 closures carry its cell, want every one that reads or writes it |



| Red oracle test/subtest | First error line |
|---|---|
| TestCapturedUnionAgrees | Aggregate failure; see failing child rows. |
| TestCapturedUnionAgrees/by_value.ts | Aggregate failure; see failing child rows. |
| TestCapturedUnionAgrees/by_value.ts/sanitized | captured_union_test.go:94: leaks: LeakSanitizer ended the run, so its output wasn't compared: exit 1 |
| TestCapturedUnionAgrees/by_value.ts/counted | captured_union_test.go:110: counts: heap values leaked: 3 (allocations 35, frees 32, in regions 0) |
| TestCapturedUnionAgrees/alternating.a | Aggregate failure; see failing child rows. |
| TestCapturedUnionAgrees/alternating.a/sanitized | captured_union_test.go:94: leaks: LeakSanitizer ended the run, so its output wasn't compared: exit 1 |
| TestCapturedUnionAgrees/alternating.a/counted | captured_union_test.go:110: counts: heap values leaked: 18 (allocations 54, frees 36, in regions 0) |
| TestCapturedUnionAgrees/alternating.ts | Aggregate failure; see failing child rows. |
| TestCapturedUnionAgrees/alternating.ts/sanitized | captured_union_test.go:94: leaks: LeakSanitizer ended the run, so its output wasn't compared: exit 1 |
| TestCapturedUnionAgrees/alternating.ts/counted | captured_union_test.go:110: counts: heap values leaked: 18 (allocations 54, frees 36, in regions 0) |
| TestCapturedUnionAgrees/released.a | Aggregate failure; see failing child rows. |
| TestCapturedUnionAgrees/released.a/sanitized | captured_union_test.go:94: leaks: LeakSanitizer ended the run, so its output wasn't compared: exit 1 |
| TestCapturedUnionAgrees/released.a/counted | captured_union_test.go:110: counts: heap values leaked: 30 (allocations 96, frees 66, in regions 0) |
| TestCapturedUnionAgrees/released.ts | Aggregate failure; see failing child rows. |
| TestCapturedUnionAgrees/released.ts/sanitized | captured_union_test.go:94: leaks: LeakSanitizer ended the run, so its output wasn't compared: exit 1 |
| TestCapturedUnionAgrees/released.ts/counted | captured_union_test.go:110: counts: heap values leaked: 30 (allocations 96, frees 66, in regions 0) |
| TestCapturedUnionAgrees/by_value.a | Aggregate failure; see failing child rows. |
| TestCapturedUnionAgrees/by_value.a/sanitized | captured_union_test.go:94: leaks: LeakSanitizer ended the run, so its output wasn't compared: exit 1 |
| TestCapturedUnionAgrees/by_value.a/counted | captured_union_test.go:110: counts: heap values leaked: 3 (allocations 35, frees 32, in regions 0) |



| Pin (both .ts and .a) | Sanitized leak check | Counted build | Detection |
|---|---|---|---|
| class | PASS | PASS | neither |
| object | PASS | PASS | neither |
| by_value | FAIL | FAIL | both |
| released | FAIL | FAIL | both |
| nested | PASS | PASS | neither |
| alternating | FAIL | FAIL | both |
| graph | PASS | PASS | neither |
| cycle | PASS | PASS | neither |

Native and JavaScript subtests all pass. Counted leaks: by_value = 3 (35 allocations, 32 frees), released = 30 (96, 66), alternating = 18 (54, 36), zero in regions. M3 lower failures are the pre-existing HEAD cycle assertions, not an additional M3 regression.

A separately labelled diagnostic rerun temporarily added only full-result logging to the sanitized and counted checks, because their original failure messages omit stdout. It reapplied the same M3 emitter edit, ran the oracle pins uncached, then restored both files and passed git diff --exit-code. These full got records include exit code, complete stdout, complete stderr, and expected stdout; they are in the diagnostic appendix. Original requested runs above remain the primary results.

## Steps 4, 4b, and 5: full suites and main comparisons

```text
go test ./internal/lower -count=1 -timeout 90m > /tmp/lower.log 2>&1
go test ./internal/ir -count=1 -timeout 30m
go test ./internal/oracle -count=1 -timeout 120m > /tmp/oracle.log 2>&1
```


### Full lower

Exit 1; wall 2135.449 seconds; command timeout 90m.

| Failing test | First error line | Environment? | Main comparison |
|---|---|---|---|
| TestCapturedUnionLowersToOneCell | Aggregate failure; see child rows. | No | TestCapturedUnionLowersToOneCell: exit 0 (main-lower-0.log); TestCapturedUnionLowersToOneCell/cycle.(ts\|a): exit 1 (main-lower-0-transplanted.log) |
| TestCapturedUnionLowersToOneCell/cycle.a | captured_union_test.go:59: active: 1 closures carry its cell, want every one that reads or writes it | No | TestCapturedUnionLowersToOneCell: exit 0 (main-lower-0.log); TestCapturedUnionLowersToOneCell/cycle.(ts\|a): exit 1 (main-lower-0-transplanted.log) |
| TestCapturedUnionLowersToOneCell/cycle.ts | captured_union_test.go:59: active: 1 closures carry its cell, want every one that reads or writes it | No | TestCapturedUnionLowersToOneCell: exit 0 (main-lower-0.log); TestCapturedUnionLowersToOneCell/cycle.(ts\|a): exit 1 (main-lower-0-transplanted.log) |

```text
--- FAIL: TestCapturedUnionLowersToOneCell (0.00s)
    --- FAIL: TestCapturedUnionLowersToOneCell/cycle.a (0.11s)
        captured_union_test.go:48: active: graph cell true
        captured_union_test.go:59: active: 1 closures carry its cell, want every one that reads or writes it
    --- FAIL: TestCapturedUnionLowersToOneCell/cycle.ts (0.11s)
        captured_union_test.go:48: active: graph cell true
        captured_union_test.go:59: active: 1 closures carry its cell, want every one that reads or writes it
FAIL
FAIL	github.com/system-inc/adamic/internal/lower	2130.048s
FAIL
```


### Full ir

Exit 0; wall 171.380 seconds; command timeout 30m.

No failures.

```text
ok  	github.com/system-inc/adamic/internal/ir	138.375s
```


### Full oracle

Exit 1; wall 5296.929 seconds; command timeout 120m.

| Failing test | First error line | Environment? | Main comparison |
|---|---|---|---|
| TestAssignmentProofWiderTargetMutant | assignment_proofs_test.go:202: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520 | Yes: missing TypeScript resolution | TestAssignmentProofWiderTargetMutant: exit 1 (main-oracle-11.log) |
| TestAssignmentProofWiderTarget | assignment_proofs_test.go:197: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520 | Yes: missing TypeScript resolution | TestAssignmentProofWiderTarget: exit 1 (main-oracle-10.log) |
| TestAssignmentProofCompoundMutant | assignment_proofs_test.go:192: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520 | Yes: missing TypeScript resolution | TestAssignmentProofCompoundMutant: exit 1 (main-oracle-3.log) |
| TestAssignmentProofCompound | assignment_proofs_test.go:187: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520 | Yes: missing TypeScript resolution | TestAssignmentProofCompound: exit 1 (main-oracle-2.log) |
| TestAssignmentProofChainedMutant | assignment_proofs_test.go:182: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520 | Yes: missing TypeScript resolution | TestAssignmentProofChainedMutant: exit 1 (main-oracle-1.log) |
| TestAssignmentProofChained | assignment_proofs_test.go:177: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520 | Yes: missing TypeScript resolution | TestAssignmentProofChained: exit 1 (main-oracle-0.log) |
| TestAssignmentProofReturnDefinedMutant | assignment_proofs_test.go:172: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520 | Yes: missing TypeScript resolution | TestAssignmentProofReturnDefinedMutant: exit 1 (main-oracle-7.log) |
| TestAssignmentProofReturnDefined | assignment_proofs_test.go:167: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520 | Yes: missing TypeScript resolution | TestAssignmentProofReturnDefined: exit 1 (main-oracle-6.log) |
| TestAssignmentProofIfNarrowingMutant | assignment_proofs_test.go:162: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520 | Yes: missing TypeScript resolution | TestAssignmentProofIfNarrowingMutant: exit 1 (main-oracle-5.log) |
| TestAssignmentProofIfNarrowing | assignment_proofs_test.go:157: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520 | Yes: missing TypeScript resolution | TestAssignmentProofIfNarrowing: exit 1 (main-oracle-4.log) |
| TestAssignmentProofScannerKeywordMutant | assignment_proofs_test.go:152: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520 | Yes: missing TypeScript resolution | TestAssignmentProofScannerKeywordMutant: exit 1 (main-oracle-9.log) |
| TestAssignmentProofScannerKeyword | assignment_proofs_test.go:147: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520 | Yes: missing TypeScript resolution | TestAssignmentProofScannerKeyword: exit 1 (main-oracle-8.log) |
| TestStdoutColumnsRefusesWASI | library_node_stdout_columns_test.go:246: want typed, named WASI refusal, got native: wasm32-wasi requires WASI_SYSROOT | Yes: WASI_SYSROOT unset | TestStdoutColumnsRefusesWASI: exit 1 (main-oracle-14.log) |
| TestFamiliesCatchAPlantedCase_Stage0EmitsDeterministically | families_generated_test.go:5425: the probe of TestStage0EmitsDeterministically was killed at 90 s: | Yes: 90 s probe budget; see isolated reruns | TestFamiliesCatchAPlantedCase_Stage0EmitsDeterministically: exit 1 (main-oracle-13.log) |
| TestFamiliesCatchAPlantedCase_CountsAreRecorded | families_generated_test.go:5190: the probe of TestCountsAreRecorded was killed at 90 s: | Yes: 90 s probe budget; see isolated reruns | TestFamiliesCatchAPlantedCase_CountsAreRecorded: exit 0 (main-oracle-12.log) |



Oracle environment classification: the twelve assignment-proof failures cannot resolve TypeScript because those tests replace NODE_PATH with the repository’s stage3/api/node_modules path; the requested npm install lives outside the repository. The WASI refusal test encounters the missing WASI_SYSROOT before reaching its expected typed refusal. The two family-wrapper failures report their internal 90-second probe limit; isolated main/HEAD results below determine reproducibility. These are separate from the captured-Union pins, which all pass at runtime.

### Main reruns

`lower` / `TestCapturedUnionLowersToOneCell`: exit 0, wall 38.836 seconds.

```text
testing: warning: no tests to run
PASS
ok  	github.com/system-inc/adamic/internal/lower	0.012s [no tests to run]
```


`lower` / `TestCapturedUnionLowersToOneCell/cycle.(ts|a)`: exit 1, wall 30.232 seconds. HEAD test and fixtures temporarily transplanted; main compiler unchanged

```text
=== RUN   TestCapturedUnionLowersToOneCell
=== PAUSE TestCapturedUnionLowersToOneCell
=== CONT  TestCapturedUnionLowersToOneCell
=== RUN   TestCapturedUnionLowersToOneCell/cycle.ts
=== PAUSE TestCapturedUnionLowersToOneCell/cycle.ts
=== RUN   TestCapturedUnionLowersToOneCell/cycle.a
=== PAUSE TestCapturedUnionLowersToOneCell/cycle.a
=== CONT  TestCapturedUnionLowersToOneCell/cycle.ts
=== CONT  TestCapturedUnionLowersToOneCell/cycle.a
    captured_union_test.go:37: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellcycle.a708020591/001/main.a:15:7: stage 0 can't lower a union of differently held members variable a function value captures yet
=== NAME  TestCapturedUnionLowersToOneCell/cycle.ts
    captured_union_test.go:37: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellcycle.ts3898901391/001/main.ts:15:7: stage 0 can't lower a union of differently held members variable a function value captures yet
--- FAIL: TestCapturedUnionLowersToOneCell (0.00s)
    --- FAIL: TestCapturedUnionLowersToOneCell/cycle.a (0.16s)
    --- FAIL: TestCapturedUnionLowersToOneCell/cycle.ts (0.16s)
FAIL
FAIL	github.com/system-inc/adamic/internal/lower	0.199s
FAIL
```


`oracle` / `TestAssignmentProofChained`: exit 1, wall 58.069 seconds.

```text
=== RUN   TestAssignmentProofChained
=== PAUSE TestAssignmentProofChained
=== CONT  TestAssignmentProofChained
    assignment_proofs_test.go:177: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofChained (0.08s)
FAIL
FAIL	github.com/system-inc/adamic/internal/oracle	0.099s
FAIL
```


`oracle` / `TestAssignmentProofChainedMutant`: exit 1, wall 5.075 seconds.

```text
=== RUN   TestAssignmentProofChainedMutant
=== PAUSE TestAssignmentProofChainedMutant
=== CONT  TestAssignmentProofChainedMutant
    assignment_proofs_test.go:182: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofChainedMutant (0.08s)
FAIL
FAIL	github.com/system-inc/adamic/internal/oracle	0.090s
FAIL
```


`oracle` / `TestAssignmentProofCompound`: exit 1, wall 6.651 seconds.

```text
=== RUN   TestAssignmentProofCompound
=== PAUSE TestAssignmentProofCompound
=== CONT  TestAssignmentProofCompound
    assignment_proofs_test.go:187: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofCompound (0.08s)
FAIL
FAIL	github.com/system-inc/adamic/internal/oracle	0.100s
FAIL
```


`oracle` / `TestAssignmentProofCompoundMutant`: exit 1, wall 4.577 seconds.

```text
=== RUN   TestAssignmentProofCompoundMutant
=== PAUSE TestAssignmentProofCompoundMutant
=== CONT  TestAssignmentProofCompoundMutant
    assignment_proofs_test.go:192: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofCompoundMutant (0.07s)
FAIL
FAIL	github.com/system-inc/adamic/internal/oracle	0.091s
FAIL
```


`oracle` / `TestAssignmentProofIfNarrowing`: exit 1, wall 4.409 seconds.

```text
=== RUN   TestAssignmentProofIfNarrowing
=== PAUSE TestAssignmentProofIfNarrowing
=== CONT  TestAssignmentProofIfNarrowing
    assignment_proofs_test.go:157: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofIfNarrowing (0.12s)
FAIL
FAIL	github.com/system-inc/adamic/internal/oracle	0.206s
FAIL
```


`oracle` / `TestAssignmentProofIfNarrowingMutant`: exit 1, wall 4.441 seconds.

```text
=== RUN   TestAssignmentProofIfNarrowingMutant
=== PAUSE TestAssignmentProofIfNarrowingMutant
=== CONT  TestAssignmentProofIfNarrowingMutant
    assignment_proofs_test.go:162: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofIfNarrowingMutant (0.13s)
FAIL
FAIL	github.com/system-inc/adamic/internal/oracle	0.159s
FAIL
```


`oracle` / `TestAssignmentProofReturnDefined`: exit 1, wall 3.746 seconds.

```text
=== RUN   TestAssignmentProofReturnDefined
=== PAUSE TestAssignmentProofReturnDefined
=== CONT  TestAssignmentProofReturnDefined
    assignment_proofs_test.go:167: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofReturnDefined (0.07s)
FAIL
FAIL	github.com/system-inc/adamic/internal/oracle	0.079s
FAIL
```


`oracle` / `TestAssignmentProofReturnDefinedMutant`: exit 1, wall 3.717 seconds.

```text
=== RUN   TestAssignmentProofReturnDefinedMutant
=== PAUSE TestAssignmentProofReturnDefinedMutant
=== CONT  TestAssignmentProofReturnDefinedMutant
    assignment_proofs_test.go:172: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofReturnDefinedMutant (0.10s)
FAIL
FAIL	github.com/system-inc/adamic/internal/oracle	0.112s
FAIL
```


`oracle` / `TestAssignmentProofScannerKeyword`: exit 1, wall 3.596 seconds.

```text
=== RUN   TestAssignmentProofScannerKeyword
=== PAUSE TestAssignmentProofScannerKeyword
=== CONT  TestAssignmentProofScannerKeyword
    assignment_proofs_test.go:147: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofScannerKeyword (0.05s)
FAIL
FAIL	github.com/system-inc/adamic/internal/oracle	0.075s
FAIL
```


`oracle` / `TestAssignmentProofScannerKeywordMutant`: exit 1, wall 4.113 seconds.

```text
=== RUN   TestAssignmentProofScannerKeywordMutant
=== PAUSE TestAssignmentProofScannerKeywordMutant
=== CONT  TestAssignmentProofScannerKeywordMutant
    assignment_proofs_test.go:152: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofScannerKeywordMutant (0.07s)
FAIL
FAIL	github.com/system-inc/adamic/internal/oracle	0.091s
FAIL
```


`oracle` / `TestAssignmentProofWiderTarget`: exit 1, wall 5.363 seconds.

```text
=== RUN   TestAssignmentProofWiderTarget
=== PAUSE TestAssignmentProofWiderTarget
=== CONT  TestAssignmentProofWiderTarget
    assignment_proofs_test.go:197: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofWiderTarget (0.13s)
FAIL
FAIL	github.com/system-inc/adamic/internal/oracle	0.152s
FAIL
```


`oracle` / `TestAssignmentProofWiderTargetMutant`: exit 1, wall 4.026 seconds.

```text
=== RUN   TestAssignmentProofWiderTargetMutant
=== PAUSE TestAssignmentProofWiderTargetMutant
=== CONT  TestAssignmentProofWiderTargetMutant
    assignment_proofs_test.go:202: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
          throw err;
          ^
        
        Error: Cannot find module 'typescript'
        Require stack:
        - /workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs
            at Module._resolveFilename (node:internal/modules/cjs/loader:1517:15)
            at wrapResolveFilename (node:internal/modules/cjs/loader:1071:27)
            at defaultResolveImplForCJSLoading (node:internal/modules/cjs/loader:1095:10)
            at resolveForCJSWithHooks (node:internal/modules/cjs/loader:1122:12)
            at Module._load (node:internal/modules/cjs/loader:1294:5)
            at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
            at Module.require (node:internal/modules/cjs/loader:1617:12)
            at require (node:internal/modules/helpers:153:16)
            at Object.<anonymous> (/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs:5:12)
            at Module._compile (node:internal/modules/cjs/loader:1872:14) {
          code: 'MODULE_NOT_FOUND',
          requireStack: [
            '/workspace/adamic/stage3/fixtures/assignment-proofs/check-types.cjs'
          ]
        }
        
        Node.js v24.19.0
--- FAIL: TestAssignmentProofWiderTargetMutant (0.07s)
FAIL
FAIL	github.com/system-inc/adamic/internal/oracle	0.080s
FAIL
```


`oracle` / `TestFamiliesCatchAPlantedCase_CountsAreRecorded`: exit 0, wall 33.794 seconds.

```text
=== RUN   TestFamiliesCatchAPlantedCase_CountsAreRecorded
=== PAUSE TestFamiliesCatchAPlantedCase_CountsAreRecorded
=== CONT  TestFamiliesCatchAPlantedCase_CountsAreRecorded
    families_generated_test.go:5190: the mutant planted in dedication/dedication.a failed exactly TestCountsAreRecorded_003, with "the counts changed"
--- PASS: TestFamiliesCatchAPlantedCase_CountsAreRecorded (30.19s)
PASS
ok  	github.com/system-inc/adamic/internal/oracle	30.235s
```


`oracle` / `TestFamiliesCatchAPlantedCase_Stage0EmitsDeterministically`: exit 1, wall 98.799 seconds.

```text
=== RUN   TestFamiliesCatchAPlantedCase_Stage0EmitsDeterministically
=== PAUSE TestFamiliesCatchAPlantedCase_Stage0EmitsDeterministically
=== CONT  TestFamiliesCatchAPlantedCase_Stage0EmitsDeterministically
    families_generated_test.go:5425: the probe of TestStage0EmitsDeterministically was killed at 90 s:
        === RUN   TestStage0EmitsDeterministically_000
        === PAUSE TestStage0EmitsDeterministically_000
        === RUN   TestStage0EmitsDeterministically_001
        === PAUSE TestStage0EmitsDeterministically_001
        === RUN   TestStage0EmitsDeterministically_002
        === PAUSE TestStage0EmitsDeterministically_002
        === RUN   TestStage0EmitsDeterministically_003
        === PAUSE TestStage0EmitsDeterministically_003
        === RUN   TestStage0EmitsDeterministically_004
        === PAUSE TestStage0EmitsDeterministically_004
        === RUN   TestStage0EmitsDeterministically_005
        === PAUSE TestStage0EmitsDeterministically_005
        === RUN   TestStage0EmitsDeterministically_006
        === PAUSE TestStage0EmitsDeterministically_006
        === RUN   TestStage0EmitsDeterministically_007
        === PAUSE TestStage0EmitsDeterministically_007
        === RUN   TestStage0EmitsDeterministically_008
        === PAUSE TestStage0EmitsDeterministically_008
        === RUN   TestStage0EmitsDeterministically_009
        === PAUSE TestStage0EmitsDeterministically_009
        === RUN   TestStage0EmitsDeterministically_010
        === PAUSE TestStage0EmitsDeterministically_010
        === RUN   TestStage0EmitsDeterministically_011
        === PAUSE TestStage0EmitsDeterministically_011
        === RUN   TestStage0EmitsDeterministically_012
        === PAUSE TestStage0EmitsDeterministically_012
        === RUN   TestStage0EmitsDeterministically_013
        === PAUSE TestStage0EmitsDeterministically_013
        === RUN   TestStage0EmitsDeterministically_014
        === PAUSE TestStage0EmitsDeterministically_014
        === RUN   TestStage0EmitsDeterministically_015
        === PAUSE TestStage0EmitsDeterministically_015
        === RUN   TestStage0EmitsDeterministically_016
        === PAUSE TestStage0EmitsDeterministically_016
        === RUN   TestStage0EmitsDeterministically_017
        === PAUSE TestStage0EmitsDeterministically_017
        === RUN   TestStage0EmitsDeterministically_018
        === PAUSE TestStage0EmitsDeterministically_018
        === RUN   TestStage0EmitsDeterministically_019
        === PAUSE TestStage0EmitsDeterministically_019
        === RUN   TestStage0EmitsDeterministically_020
        === PAUSE TestStage0EmitsDeterministically_020
        === RUN   TestStage0EmitsDeterministically_021
        === PAUSE TestStage0EmitsDeterministically_021
        === RUN   TestStage0EmitsDeterministically_022
        === PAUSE TestStage0EmitsDeterministically_022
        === RUN   TestStage0EmitsDeterministically_023
        === PAUSE TestStage0EmitsDeterministically_023
        === RUN   TestStage0EmitsDeterministically_024
        === PAUSE TestStage0EmitsDeterministically_024
        === RUN   TestStage0EmitsDeterministically_025
        === PAUSE TestStage0EmitsDeterministically_025
        === RUN   TestStage0EmitsDeterministically_026
        === PAUSE TestStage0EmitsDeterministically_026
        === RUN   TestStage0EmitsDeterministically_027
        === PAUSE TestStage0EmitsDeterministically_027
        === RUN   TestStage0EmitsDeterministically_028
        === PAUSE TestStage0EmitsDeterministically_028
        === RUN   TestStage0EmitsDeterministically_029
        === PAUSE TestStage0EmitsDeterministically_029
        === RUN   TestStage0EmitsDeterministically_030
        === PAUSE TestStage0EmitsDeterministically_030
        === RUN   TestStage0EmitsDeterministically_031
        === PAUSE TestStage0EmitsDeterministically_031
        === CONT  TestStage0EmitsDeterministically_000
        === CONT  TestStage0EmitsDeterministically_016
        === CONT  TestStage0EmitsDeterministically_024
        === CONT  TestStage0EmitsDeterministically_031
        === NAME  TestStage0EmitsDeterministically_000
            families_generated_test.go:5265: build go_build_./cmd/adamic_adamic e8c1fca58916 hit 43.47
        === NAME  TestStage0EmitsDeterministically_031
            families_generated_test.go:5420: build go_build_./cmd/adamic_adamic e8c1fca58916 hit 43.47
        === NAME  TestStage0EmitsDeterministically_016
            families_generated_test.go:5345: build go_build_./cmd/adamic_adamic e8c1fca58916 hit 43.47
        === NAME  TestStage0EmitsDeterministically_024
            families_generated_test.go:5385: build go_build_./cmd/adamic_adamic e8c1fca58916 miss 43.47
--- FAIL: TestFamiliesCatchAPlantedCase_Stage0EmitsDeterministically (90.01s)
FAIL
FAIL	github.com/system-inc/adamic/internal/oracle	90.032s
FAIL
```


`oracle` / `TestStdoutColumnsRefusesWASI`: exit 1, wall 8.534 seconds.

```text
=== RUN   TestStdoutColumnsRefusesWASI
=== PAUSE TestStdoutColumnsRefusesWASI
=== CONT  TestStdoutColumnsRefusesWASI
    library_node_stdout_columns_test.go:246: want typed, named WASI refusal, got native: wasm32-wasi requires WASI_SYSROOT
--- FAIL: TestStdoutColumnsRefusesWASI (0.54s)
FAIL
FAIL	github.com/system-inc/adamic/internal/oracle	0.571s
FAIL
```


The new cycle test is absent on main: the ordinary named rerun reports no tests. The explicitly labelled transplanted probe runs only cycle.ts and cycle.a against main’s unchanged compiler. Both fail there at main’s captured-Union lowering stop, before the closure-count assertion can be reached; therefore main does not exhibit the same reached assertion failure, and does not support these programs. No test/compiler edits from this comparison remain.

## Supplemental isolated and provisioning checks

These do not replace the requested full-suite observations. The two HEAD family probes were rerun individually with no other heavy command running. The assignment-proof reruns temporarily exposed the external pinned install through a stage3/api/node_modules symlink, ran the twelve tests at HEAD and main, then removed the symlink and restored HEAD. No package was installed inside the repository.

### head-isolated-CountsAreRecorded
Exit 0; wall 29.946 seconds.

```text
=== RUN   TestFamiliesCatchAPlantedCase_CountsAreRecorded
=== PAUSE TestFamiliesCatchAPlantedCase_CountsAreRecorded
=== CONT  TestFamiliesCatchAPlantedCase_CountsAreRecorded
    families_generated_test.go:5190: the mutant planted in dedication/dedication.a failed exactly TestCountsAreRecorded_003, with "the counts changed"
--- PASS: TestFamiliesCatchAPlantedCase_CountsAreRecorded (24.25s)
PASS
ok  	github.com/system-inc/adamic/internal/oracle	24.267s

```

### head-isolated-Stage0EmitsDeterministically
Exit 0; wall 10.685 seconds.

```text
=== RUN   TestFamiliesCatchAPlantedCase_Stage0EmitsDeterministically
=== PAUSE TestFamiliesCatchAPlantedCase_Stage0EmitsDeterministically
=== CONT  TestFamiliesCatchAPlantedCase_Stage0EmitsDeterministically
    families_generated_test.go:5425: the mutant planted in internal/oracle/testdata/determinism_dispatch.a failed exactly TestStage0EmitsDeterministically_020, with "stage 0 emitted different C"
--- PASS: TestFamiliesCatchAPlantedCase_Stage0EmitsDeterministically (3.86s)
PASS
ok  	github.com/system-inc/adamic/internal/oracle	3.894s

```

### head-assignment-provisioned
Exit 0; wall 25.598 seconds.

```text
=== RUN   TestAssignmentProofScannerKeyword
=== PAUSE TestAssignmentProofScannerKeyword
=== RUN   TestAssignmentProofScannerKeywordMutant
=== PAUSE TestAssignmentProofScannerKeywordMutant
=== RUN   TestAssignmentProofIfNarrowing
=== PAUSE TestAssignmentProofIfNarrowing
=== RUN   TestAssignmentProofIfNarrowingMutant
=== PAUSE TestAssignmentProofIfNarrowingMutant
=== RUN   TestAssignmentProofReturnDefined
=== PAUSE TestAssignmentProofReturnDefined
=== RUN   TestAssignmentProofReturnDefinedMutant
=== PAUSE TestAssignmentProofReturnDefinedMutant
=== RUN   TestAssignmentProofChained
=== PAUSE TestAssignmentProofChained
=== RUN   TestAssignmentProofChainedMutant
=== PAUSE TestAssignmentProofChainedMutant
=== RUN   TestAssignmentProofCompound
=== PAUSE TestAssignmentProofCompound
=== RUN   TestAssignmentProofCompoundMutant
=== PAUSE TestAssignmentProofCompoundMutant
=== RUN   TestAssignmentProofWiderTarget
=== PAUSE TestAssignmentProofWiderTarget
=== RUN   TestAssignmentProofWiderTargetMutant
=== PAUSE TestAssignmentProofWiderTargetMutant
=== CONT  TestAssignmentProofScannerKeyword
=== CONT  TestAssignmentProofChained
=== CONT  TestAssignmentProofChainedMutant
=== CONT  TestAssignmentProofWiderTargetMutant
=== CONT  TestAssignmentProofIfNarrowingMutant
=== NAME  TestAssignmentProofScannerKeyword
    assignment_proofs_test.go:147: gate cache node miss
=== NAME  TestAssignmentProofChainedMutant
    assignment_proofs_test.go:182: gate cache node miss
=== NAME  TestAssignmentProofIfNarrowingMutant
    assignment_proofs_test.go:162: gate cache node miss
=== NAME  TestAssignmentProofChained
    assignment_proofs_test.go:177: gate cache node miss
=== NAME  TestAssignmentProofWiderTargetMutant
    assignment_proofs_test.go:202: gate cache node miss
=== NAME  TestAssignmentProofIfNarrowingMutant
    assignment_proofs_test.go:162: gate cache node miss
=== NAME  TestAssignmentProofChained
    assignment_proofs_test.go:177: gate cache node miss
=== NAME  TestAssignmentProofChainedMutant
    assignment_proofs_test.go:182: gate cache node miss
=== NAME  TestAssignmentProofWiderTargetMutant
    assignment_proofs_test.go:202: gate cache node miss
=== NAME  TestAssignmentProofScannerKeyword
    assignment_proofs_test.go:147: gate cache node miss
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 1.01
=== NAME  TestAssignmentProofChained
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 1.18
=== NAME  TestAssignmentProofScannerKeyword
    oracle_test.go:688: gate cache native miss
=== NAME  TestAssignmentProofChainedMutant
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 1.23
=== NAME  TestAssignmentProofIfNarrowingMutant
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 1.26
=== NAME  TestAssignmentProofChained
    oracle_test.go:688: gate cache native miss
=== NAME  TestAssignmentProofWiderTargetMutant
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 1.32
=== NAME  TestAssignmentProofChainedMutant
    oracle_test.go:688: gate cache native miss
=== NAME  TestAssignmentProofIfNarrowingMutant
    oracle_test.go:688: gate cache native miss
=== NAME  TestAssignmentProofWiderTargetMutant
    oracle_test.go:688: gate cache native miss
=== NAME  TestAssignmentProofIfNarrowingMutant
    oracle_test.go:681: gate cache native miss
=== NAME  TestAssignmentProofWiderTargetMutant
    oracle_test.go:681: gate cache native miss
=== NAME  TestAssignmentProofScannerKeyword
    oracle_test.go:681: gate cache native miss
=== NAME  TestAssignmentProofChainedMutant
    oracle_test.go:681: gate cache native miss
=== NAME  TestAssignmentProofChained
    oracle_test.go:681: gate cache native miss
=== NAME  TestAssignmentProofIfNarrowingMutant
    oracle_test.go:695: gate cache native hit
    assignment_proofs_test.go:162: caught by Node stdout byte comparison after zero stock TypeScript diagnostics: golden "4 1\n-1 2\n", mutant "4 1\n7 2\n"
--- PASS: TestAssignmentProofIfNarrowingMutant (15.21s)
=== CONT  TestAssignmentProofReturnDefinedMutant
=== NAME  TestAssignmentProofWiderTargetMutant
    oracle_test.go:695: gate cache native hit
    assignment_proofs_test.go:202: caught by Node stdout byte comparison after zero stock TypeScript diagnostics: golden "128 128 1\n", mutant "128 0 1\n"
--- PASS: TestAssignmentProofWiderTargetMutant (15.23s)
=== CONT  TestAssignmentProofReturnDefined
=== NAME  TestAssignmentProofScannerKeyword
    oracle_test.go:695: gate cache native hit
--- PASS: TestAssignmentProofScannerKeyword (15.26s)
=== CONT  TestAssignmentProofScannerKeywordMutant
=== NAME  TestAssignmentProofChainedMutant
    oracle_test.go:695: gate cache native hit
    assignment_proofs_test.go:182: caught by Node stdout byte comparison after zero stock TypeScript diagnostics: golden "value1 value1 value1 1\n", mutant "value1 value1 undefined 1\n"
--- PASS: TestAssignmentProofChainedMutant (15.37s)
=== CONT  TestAssignmentProofIfNarrowing
=== NAME  TestAssignmentProofChained
    oracle_test.go:695: gate cache native hit
--- PASS: TestAssignmentProofChained (15.46s)
=== CONT  TestAssignmentProofCompoundMutant
=== NAME  TestAssignmentProofIfNarrowing
    assignment_proofs_test.go:157: gate cache node miss
=== NAME  TestAssignmentProofReturnDefinedMutant
    assignment_proofs_test.go:172: gate cache node miss
=== NAME  TestAssignmentProofCompoundMutant
    assignment_proofs_test.go:192: gate cache node miss
=== NAME  TestAssignmentProofReturnDefined
    assignment_proofs_test.go:167: gate cache node miss
=== NAME  TestAssignmentProofIfNarrowing
    assignment_proofs_test.go:157: gate cache node miss
=== NAME  TestAssignmentProofScannerKeywordMutant
    assignment_proofs_test.go:152: gate cache node miss
=== NAME  TestAssignmentProofIfNarrowing
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 0.03
=== NAME  TestAssignmentProofReturnDefinedMutant
    assignment_proofs_test.go:172: gate cache node miss
=== NAME  TestAssignmentProofCompoundMutant
    assignment_proofs_test.go:192: gate cache node miss
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 0.01
=== NAME  TestAssignmentProofReturnDefinedMutant
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 0.02
=== NAME  TestAssignmentProofIfNarrowing
    oracle_test.go:688: gate cache native miss
=== NAME  TestAssignmentProofReturnDefined
    assignment_proofs_test.go:167: gate cache node miss
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 0.02
=== NAME  TestAssignmentProofCompoundMutant
    oracle_test.go:688: gate cache native miss
=== NAME  TestAssignmentProofReturnDefinedMutant
    oracle_test.go:688: gate cache native miss
=== NAME  TestAssignmentProofReturnDefined
    oracle_test.go:688: gate cache native miss
=== NAME  TestAssignmentProofScannerKeywordMutant
    assignment_proofs_test.go:152: gate cache node miss
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 0.10
    oracle_test.go:688: gate cache native miss
=== NAME  TestAssignmentProofCompoundMutant
    oracle_test.go:681: gate cache native miss
=== NAME  TestAssignmentProofIfNarrowing
    oracle_test.go:681: gate cache native miss
=== NAME  TestAssignmentProofReturnDefined
    oracle_test.go:681: gate cache native miss
=== NAME  TestAssignmentProofReturnDefinedMutant
    oracle_test.go:681: gate cache native miss
=== NAME  TestAssignmentProofCompoundMutant
    oracle_test.go:695: gate cache native hit
    assignment_proofs_test.go:192: caught by Node stdout byte comparison after zero stock TypeScript diagnostics: golden "12 12 1\n14 14 2\n", mutant "2 2 1\n2 2 2\n"
=== CONT  TestAssignmentProofWiderTarget
--- PASS: TestAssignmentProofCompoundMutant (3.87s)
=== NAME  TestAssignmentProofScannerKeywordMutant
    oracle_test.go:681: gate cache native miss
=== NAME  TestAssignmentProofIfNarrowing
    oracle_test.go:695: gate cache native hit
--- PASS: TestAssignmentProofIfNarrowing (4.18s)
=== CONT  TestAssignmentProofCompound
=== NAME  TestAssignmentProofReturnDefined
    oracle_test.go:695: gate cache native hit
--- PASS: TestAssignmentProofReturnDefined (4.41s)
=== NAME  TestAssignmentProofReturnDefinedMutant
    oracle_test.go:695: gate cache native hit
    assignment_proofs_test.go:172: caught by Node stdout byte comparison after zero stock TypeScript diagnostics: golden "value1 value1 1\nvalue2 value2 2\n", mutant "value1 undefined 1\nvalue2 undefined 2\n"
--- PASS: TestAssignmentProofReturnDefinedMutant (4.45s)
=== NAME  TestAssignmentProofScannerKeywordMutant
    oracle_test.go:695: gate cache native hit
    assignment_proofs_test.go:152: caught by Node stdout byte comparison after zero stock TypeScript diagnostics: golden "128 128\n83 83\n80 80\n80 80\n80 80\n", mutant "80 80\n80 80\n80 80\n80 80\n80 80\n"
--- PASS: TestAssignmentProofScannerKeywordMutant (4.53s)
=== NAME  TestAssignmentProofCompound
    assignment_proofs_test.go:187: gate cache node miss
=== NAME  TestAssignmentProofWiderTarget
    assignment_proofs_test.go:197: gate cache node miss
=== NAME  TestAssignmentProofCompound
    assignment_proofs_test.go:187: gate cache node miss
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 0.01
=== NAME  TestAssignmentProofWiderTarget
    assignment_proofs_test.go:197: gate cache node miss
=== NAME  TestAssignmentProofCompound
    oracle_test.go:688: gate cache native miss
=== NAME  TestAssignmentProofWiderTarget
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 0.01
    oracle_test.go:688: gate cache native miss
=== NAME  TestAssignmentProofCompound
    oracle_test.go:681: gate cache native miss
=== NAME  TestAssignmentProofWiderTarget
    oracle_test.go:681: gate cache native miss
=== NAME  TestAssignmentProofCompound
    oracle_test.go:695: gate cache native hit
--- PASS: TestAssignmentProofCompound (1.85s)
=== NAME  TestAssignmentProofWiderTarget
    oracle_test.go:695: gate cache native hit
--- PASS: TestAssignmentProofWiderTarget (2.10s)
PASS
gate cache: native hits=12 misses=24
gate cache: node hits=0 misses=24
gate cache: probe hits=0 misses=0
ok  	github.com/system-inc/adamic/internal/oracle	21.457s

```

### main-assignment-provisioned
Exit 0; wall 44.883 seconds.

```text
=== RUN   TestAssignmentProofScannerKeyword
=== PAUSE TestAssignmentProofScannerKeyword
=== RUN   TestAssignmentProofScannerKeywordMutant
=== PAUSE TestAssignmentProofScannerKeywordMutant
=== RUN   TestAssignmentProofIfNarrowing
=== PAUSE TestAssignmentProofIfNarrowing
=== RUN   TestAssignmentProofIfNarrowingMutant
=== PAUSE TestAssignmentProofIfNarrowingMutant
=== RUN   TestAssignmentProofReturnDefined
=== PAUSE TestAssignmentProofReturnDefined
=== RUN   TestAssignmentProofReturnDefinedMutant
=== PAUSE TestAssignmentProofReturnDefinedMutant
=== RUN   TestAssignmentProofChained
=== PAUSE TestAssignmentProofChained
=== RUN   TestAssignmentProofChainedMutant
=== PAUSE TestAssignmentProofChainedMutant
=== RUN   TestAssignmentProofCompound
=== PAUSE TestAssignmentProofCompound
=== RUN   TestAssignmentProofCompoundMutant
=== PAUSE TestAssignmentProofCompoundMutant
=== RUN   TestAssignmentProofWiderTarget
=== PAUSE TestAssignmentProofWiderTarget
=== RUN   TestAssignmentProofWiderTargetMutant
=== PAUSE TestAssignmentProofWiderTargetMutant
=== CONT  TestAssignmentProofScannerKeyword
=== CONT  TestAssignmentProofChained
=== CONT  TestAssignmentProofCompoundMutant
=== CONT  TestAssignmentProofIfNarrowingMutant
=== CONT  TestAssignmentProofReturnDefinedMutant
=== NAME  TestAssignmentProofIfNarrowingMutant
    assignment_proofs_test.go:162: gate cache node miss
=== NAME  TestAssignmentProofScannerKeyword
    assignment_proofs_test.go:147: gate cache node miss
=== NAME  TestAssignmentProofReturnDefinedMutant
    assignment_proofs_test.go:172: gate cache node miss
=== NAME  TestAssignmentProofCompoundMutant
    assignment_proofs_test.go:192: gate cache node miss
=== NAME  TestAssignmentProofChained
    assignment_proofs_test.go:177: gate cache node miss
=== NAME  TestAssignmentProofIfNarrowingMutant
    assignment_proofs_test.go:162: gate cache node miss
=== NAME  TestAssignmentProofChained
    assignment_proofs_test.go:177: gate cache node miss
=== NAME  TestAssignmentProofReturnDefinedMutant
    assignment_proofs_test.go:172: gate cache node miss
=== NAME  TestAssignmentProofCompoundMutant
    assignment_proofs_test.go:192: gate cache node miss
=== NAME  TestAssignmentProofScannerKeyword
    assignment_proofs_test.go:147: gate cache node miss
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 2565158e7096 miss 17.25
=== NAME  TestAssignmentProofIfNarrowingMutant
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 2565158e7096 hit 17.25
=== NAME  TestAssignmentProofCompoundMutant
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 2565158e7096 hit 17.25
=== NAME  TestAssignmentProofReturnDefinedMutant
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 2565158e7096 hit 17.25
=== NAME  TestAssignmentProofChained
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 2565158e7096 hit 17.25
=== NAME  TestAssignmentProofIfNarrowingMutant
    oracle_test.go:688: gate cache native miss
=== NAME  TestAssignmentProofChained
    oracle_test.go:688: gate cache native miss
=== NAME  TestAssignmentProofReturnDefinedMutant
    oracle_test.go:688: gate cache native miss
=== NAME  TestAssignmentProofScannerKeyword
    oracle_test.go:688: gate cache native miss
=== NAME  TestAssignmentProofCompoundMutant
    oracle_test.go:688: gate cache native miss
=== NAME  TestAssignmentProofChained
    oracle_test.go:681: gate cache native miss
=== NAME  TestAssignmentProofIfNarrowingMutant
    oracle_test.go:681: gate cache native miss
=== NAME  TestAssignmentProofReturnDefinedMutant
    oracle_test.go:681: gate cache native miss
=== NAME  TestAssignmentProofScannerKeyword
    oracle_test.go:681: gate cache native miss
=== NAME  TestAssignmentProofCompoundMutant
    oracle_test.go:681: gate cache native miss
=== NAME  TestAssignmentProofIfNarrowingMutant
    oracle_test.go:695: gate cache native hit
    assignment_proofs_test.go:162: caught by Node stdout byte comparison after zero stock TypeScript diagnostics: golden "4 1\n-1 2\n", mutant "4 1\n7 2\n"
--- PASS: TestAssignmentProofIfNarrowingMutant (32.34s)
=== CONT  TestAssignmentProofIfNarrowing
=== NAME  TestAssignmentProofChained
    oracle_test.go:695: gate cache native hit
--- PASS: TestAssignmentProofChained (32.39s)
=== CONT  TestAssignmentProofScannerKeywordMutant
=== NAME  TestAssignmentProofReturnDefinedMutant
    oracle_test.go:695: gate cache native hit
    assignment_proofs_test.go:172: caught by Node stdout byte comparison after zero stock TypeScript diagnostics: golden "value1 value1 1\nvalue2 value2 2\n", mutant "value1 undefined 1\nvalue2 undefined 2\n"
--- PASS: TestAssignmentProofReturnDefinedMutant (32.43s)
=== CONT  TestAssignmentProofWiderTargetMutant
=== NAME  TestAssignmentProofScannerKeyword
    oracle_test.go:695: gate cache native hit
--- PASS: TestAssignmentProofScannerKeyword (32.54s)
=== CONT  TestAssignmentProofWiderTarget
=== NAME  TestAssignmentProofCompoundMutant
    oracle_test.go:695: gate cache native hit
    assignment_proofs_test.go:192: caught by Node stdout byte comparison after zero stock TypeScript diagnostics: golden "12 12 1\n14 14 2\n", mutant "2 2 1\n2 2 2\n"
--- PASS: TestAssignmentProofCompoundMutant (32.65s)
=== CONT  TestAssignmentProofCompound
=== NAME  TestAssignmentProofScannerKeywordMutant
    assignment_proofs_test.go:152: gate cache node miss
    assignment_proofs_test.go:152: gate cache node miss
=== NAME  TestAssignmentProofWiderTargetMutant
    assignment_proofs_test.go:202: gate cache node miss
=== NAME  TestAssignmentProofScannerKeywordMutant
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 2565158e7096 hit 0.03
    oracle_test.go:688: gate cache native miss
=== NAME  TestAssignmentProofIfNarrowing
    assignment_proofs_test.go:157: gate cache node miss
=== NAME  TestAssignmentProofCompound
    assignment_proofs_test.go:187: gate cache node miss
=== NAME  TestAssignmentProofWiderTargetMutant
    assignment_proofs_test.go:202: gate cache node miss
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 2565158e7096 hit 0.21
    oracle_test.go:688: gate cache native miss
=== NAME  TestAssignmentProofWiderTarget
    assignment_proofs_test.go:197: gate cache node miss
=== NAME  TestAssignmentProofIfNarrowing
    assignment_proofs_test.go:157: gate cache node miss
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 2565158e7096 hit 0.05
=== NAME  TestAssignmentProofCompound
    assignment_proofs_test.go:187: gate cache node miss
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 2565158e7096 hit 0.03
=== NAME  TestAssignmentProofScannerKeywordMutant
    oracle_test.go:681: gate cache native miss
=== NAME  TestAssignmentProofIfNarrowing
    oracle_test.go:688: gate cache native miss
=== NAME  TestAssignmentProofCompound
    oracle_test.go:688: gate cache native miss
=== NAME  TestAssignmentProofWiderTarget
    assignment_proofs_test.go:197: gate cache node miss
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 2565158e7096 hit 0.02
=== NAME  TestAssignmentProofScannerKeywordMutant
    oracle_test.go:695: gate cache native hit
    assignment_proofs_test.go:152: caught by Node stdout byte comparison after zero stock TypeScript diagnostics: golden "128 128\n83 83\n80 80\n80 80\n80 80\n", mutant "80 80\n80 80\n80 80\n80 80\n80 80\n"
--- PASS: TestAssignmentProofScannerKeywordMutant (5.04s)
=== CONT  TestAssignmentProofChainedMutant
=== NAME  TestAssignmentProofWiderTarget
    oracle_test.go:688: gate cache native miss
=== NAME  TestAssignmentProofWiderTargetMutant
    oracle_test.go:681: gate cache native miss
=== NAME  TestAssignmentProofIfNarrowing
    oracle_test.go:681: gate cache native miss
=== NAME  TestAssignmentProofWiderTargetMutant
    oracle_test.go:695: gate cache native hit
    assignment_proofs_test.go:202: caught by Node stdout byte comparison after zero stock TypeScript diagnostics: golden "128 128 1\n", mutant "128 0 1\n"
--- PASS: TestAssignmentProofWiderTargetMutant (5.41s)
=== CONT  TestAssignmentProofReturnDefined
=== NAME  TestAssignmentProofCompound
    oracle_test.go:681: gate cache native miss
=== NAME  TestAssignmentProofIfNarrowing
    oracle_test.go:695: gate cache native hit
--- PASS: TestAssignmentProofIfNarrowing (5.99s)
=== NAME  TestAssignmentProofWiderTarget
    oracle_test.go:681: gate cache native miss
=== NAME  TestAssignmentProofCompound
    oracle_test.go:695: gate cache native hit
--- PASS: TestAssignmentProofCompound (6.20s)
=== NAME  TestAssignmentProofChainedMutant
    assignment_proofs_test.go:182: gate cache node miss
=== NAME  TestAssignmentProofWiderTarget
    oracle_test.go:695: gate cache native hit
--- PASS: TestAssignmentProofWiderTarget (6.40s)
=== NAME  TestAssignmentProofChainedMutant
    assignment_proofs_test.go:182: gate cache node miss
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 2565158e7096 hit 0.01
    oracle_test.go:688: gate cache native miss
=== NAME  TestAssignmentProofReturnDefined
    assignment_proofs_test.go:167: gate cache node miss
=== NAME  TestAssignmentProofChainedMutant
    oracle_test.go:681: gate cache native miss
=== NAME  TestAssignmentProofReturnDefined
    assignment_proofs_test.go:167: gate cache node miss
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 2565158e7096 hit 0.01
    oracle_test.go:688: gate cache native miss
=== NAME  TestAssignmentProofChainedMutant
    oracle_test.go:695: gate cache native hit
    assignment_proofs_test.go:182: caught by Node stdout byte comparison after zero stock TypeScript diagnostics: golden "value1 value1 value1 1\n", mutant "value1 value1 undefined 1\n"
--- PASS: TestAssignmentProofChainedMutant (2.71s)
=== NAME  TestAssignmentProofReturnDefined
    oracle_test.go:681: gate cache native miss
    oracle_test.go:695: gate cache native hit
--- PASS: TestAssignmentProofReturnDefined (2.86s)
PASS
gate cache: native hits=12 misses=24
gate cache: node hits=0 misses=24
gate cache: probe hits=0 misses=0
ok  	github.com/system-inc/adamic/internal/oracle	40.742s

```


## Step 6: admission delta

The first attempt was interrupted during the environment transition/tool cancellation: no gate process remained and gate-attempt1.json was empty, so it established no verdict. Its stderr and build trace were preserved outside the repository. The completed run below restarted the same command alone and reused both completed compiler products. The admission tool itself was built outside the repository before the measured gate command, so exit 3 is the tool’s actual exit, not a go-run wrapper exit.

Command and process result:
```text
/tmp/n7s-proof/adamic-admission-delta --base efe887763f26e96653546fae904d57539139f7ae --head 54397e8f7ff1c896343d8e62c8c94ef87e5b6991 --manifest /tmp/n7s-proof/manifest.json --manifest-generator cloud/admission-corpus/manifest.py --workers 2 --json
```


Exit: 3. Wall seconds: 2077.179847525.

- `base`: `"efe887763f26e96653546fae904d57539139f7ae"`
- `head`: `"54397e8f7ff1c896343d8e62c8c94ef87e5b6991"`
- `verdict`: `"void"`
- `void`: `true`
- `void_reason`: `"internal/oracle/testdata/captured_union/alternating.ts: native compile timeout; internal/oracle/testdata/captured_union/by_value.ts: native compile timeout; internal/oracle/testdata/captured_union/class.ts: native compile timeout; internal/oracle/testdata/captured_union/cycle.ts: native compile timeout; internal/oracle/testdata/captured_union/graph.ts: native compile timeout; internal/oracle/testdata/captured_union/nested.ts: native compile timeout; internal/oracle/testdata/captured_union/object.ts: native compile timeout; internal/oracle/testdata/captured_union/released.ts: native compile timeout; internal/oracle/testdata/intersection_views_stops/nested_boolean.ts: native compile timeout; internal/oracle/testdata/intersection_views_stops/nested_boolean_repaired.ts: native compile timeout; internal/oracle/testdata/intersection_views_stops/nested_literal.ts: native compile timeout; internal/oracle/testdata/intersection_views_stops/nested_literal_repaired.ts: native compile timeout; internal/oracle/testdata/intersection_views_stops/nested_token.ts: native compile timeout; internal/oracle/testdata/intersection_views_stops/nested_token_repaired.ts: native compile timeout; internal/oracle/testdata/intersection_views_stops/nested_write.ts: native compile timeout; internal/oracle/testdata/intersection_views_stops/nested_write_repaired.ts: native compile timeout; internal/oracle/testdata/intl_collator/in_compare.a: native compile timeout; internal/oracle/testdata/intl_collator/locale_en.a: native compile timeout; internal/oracle/testdata/reflect_get/stop_literal_getter.a: native compile timeout"`
- `diff_count`: `8`
- `admitted`: `8`
- `newly_refused`: `0`

### Classification stderr differences (8 programs)

#### internal/oracle/testdata/captured_union/alternating.ts

Base stderr:
```text
adamic: /tmp/adamic-gate/admission-delta-2249853319/head/internal/oracle/testdata/captured_union/alternating.ts:17:4: stage 0 can't lower a union of differently held members variable a function value captures yet
```


HEAD stderr:
```text

```


#### internal/oracle/testdata/captured_union/by_value.ts

Base stderr:
```text
adamic: /tmp/adamic-gate/admission-delta-2249853319/head/internal/oracle/testdata/captured_union/by_value.ts:20:4: stage 0 can't lower a union of differently held members variable a function value captures yet
```


HEAD stderr:
```text

```


#### internal/oracle/testdata/captured_union/class.ts

Base stderr:
```text
adamic: /tmp/adamic-gate/admission-delta-2249853319/head/internal/oracle/testdata/captured_union/class.ts:13:7: stage 0 can't lower a union of differently held members variable a function value captures yet
```


HEAD stderr:
```text

```


#### internal/oracle/testdata/captured_union/cycle.ts

Base stderr:
```text
adamic: /tmp/adamic-gate/admission-delta-2249853319/head/internal/oracle/testdata/captured_union/cycle.ts:15:7: stage 0 can't lower a union of differently held members variable a function value captures yet
```


HEAD stderr:
```text

```


#### internal/oracle/testdata/captured_union/graph.ts

Base stderr:
```text
adamic: /tmp/adamic-gate/admission-delta-2249853319/head/internal/oracle/testdata/captured_union/graph.ts:19:5: stage 0 can't lower a union of differently held members variable a function value captures yet
```


HEAD stderr:
```text

```


#### internal/oracle/testdata/captured_union/nested.ts

Base stderr:
```text
adamic: /tmp/adamic-gate/admission-delta-2249853319/head/internal/oracle/testdata/captured_union/nested.ts:8:5: stage 0 can't lower a union of differently held members variable a function value captures yet
```


HEAD stderr:
```text

```


#### internal/oracle/testdata/captured_union/object.ts

Base stderr:
```text
adamic: /tmp/adamic-gate/admission-delta-2249853319/head/internal/oracle/testdata/captured_union/object.ts:7:7: stage 0 can't lower a union of differently held members variable a function value captures yet
```


HEAD stderr:
```text

```


#### internal/oracle/testdata/captured_union/released.ts

Base stderr:
```text
adamic: /tmp/adamic-gate/admission-delta-2249853319/head/internal/oracle/testdata/captured_union/released.ts:17:4: stage 0 can't lower a union of differently held members variable a function value captures yet
```


HEAD stderr:
```text

```


### Original gate timeouts and void reasons

The gate classified 5241 programs: 3,282 accepted by both, 1,951 refused by both, 8 newly accepted, 0 newly refused. All 8 stderr differences are the captured_union .ts pins: main reports its capture stop and HEAD stderr is empty. There were no classification timeouts. Runtime selection size was 123; omitted 0; budget 0. The gate recorded 19 runtime timeout runs, retried all 19 alone, and 19 still timed out. Every one was a native compilation timeout at the original 10-second limit. This is a void result, not a green admission proof.

| Program | Void reason | Retried by gate |
|---|---|---|
| internal/oracle/testdata/captured_union/alternating.ts | native compile timeout | True |
| internal/oracle/testdata/captured_union/by_value.ts | native compile timeout | True |
| internal/oracle/testdata/captured_union/class.ts | native compile timeout | True |
| internal/oracle/testdata/captured_union/cycle.ts | native compile timeout | True |
| internal/oracle/testdata/captured_union/graph.ts | native compile timeout | True |
| internal/oracle/testdata/captured_union/nested.ts | native compile timeout | True |
| internal/oracle/testdata/captured_union/object.ts | native compile timeout | True |
| internal/oracle/testdata/captured_union/released.ts | native compile timeout | True |
| internal/oracle/testdata/intersection_views_stops/nested_boolean.ts | native compile timeout | True |
| internal/oracle/testdata/intersection_views_stops/nested_boolean_repaired.ts | native compile timeout | True |
| internal/oracle/testdata/intersection_views_stops/nested_literal.ts | native compile timeout | True |
| internal/oracle/testdata/intersection_views_stops/nested_literal_repaired.ts | native compile timeout | True |
| internal/oracle/testdata/intersection_views_stops/nested_token.ts | native compile timeout | True |
| internal/oracle/testdata/intersection_views_stops/nested_token_repaired.ts | native compile timeout | True |
| internal/oracle/testdata/intersection_views_stops/nested_write.ts | native compile timeout | True |
| internal/oracle/testdata/intersection_views_stops/nested_write_repaired.ts | native compile timeout | True |
| internal/oracle/testdata/intl_collator/in_compare.a | native compile timeout | True |
| internal/oracle/testdata/intl_collator/locale_en.a | native compile timeout | True |
| internal/oracle/testdata/reflect_get/stop_literal_getter.a | native compile timeout | True |

### Gate stderr

```text
negative witnesses: 50
ruled run-time stops: 49
ruled backend divergences: 0
runtime timeouts: 19 runs timed out, 19 retried alone, 19 timed out again
changed programs: 8
admission delta is void: internal/oracle/testdata/captured_union/alternating.ts: native compile timeout; internal/oracle/testdata/captured_union/by_value.ts: native compile timeout; internal/oracle/testdata/captured_union/class.ts: native compile timeout; internal/oracle/testdata/captured_union/cycle.ts: native compile timeout; internal/oracle/testdata/captured_union/graph.ts: native compile timeout; internal/oracle/testdata/captured_union/nested.ts: native compile timeout; internal/oracle/testdata/captured_union/object.ts: native compile timeout; internal/oracle/testdata/captured_union/released.ts: native compile timeout; internal/oracle/testdata/intersection_views_stops/nested_boolean.ts: native compile timeout; internal/oracle/testdata/intersection_views_stops/nested_boolean_repaired.ts: native compile timeout; internal/oracle/testdata/intersection_views_stops/nested_literal.ts: native compile timeout; internal/oracle/testdata/intersection_views_stops/nested_literal_repaired.ts: native compile timeout; internal/oracle/testdata/intersection_views_stops/nested_token.ts: native compile timeout; internal/oracle/testdata/intersection_views_stops/nested_token_repaired.ts: native compile timeout; internal/oracle/testdata/intersection_views_stops/nested_write.ts: native compile timeout; internal/oracle/testdata/intersection_views_stops/nested_write_repaired.ts: native compile timeout; internal/oracle/testdata/intl_collator/in_compare.a: native compile timeout; internal/oracle/testdata/intl_collator/locale_en.a: native compile timeout; internal/oracle/testdata/reflect_get/stop_literal_getter.a: native compile timeout
```


The isolated cross-revision reruns below retain that same 10-second native compilation limit. Seven of the eight new pins build and run on HEAD in isolation; graph still times out. Main refuses all eight before emission. All eight intersection-view cases and both Intl.Collator cases build on both compilers in isolation. reflect_get/stop_literal_getter.a still times out on both. Native runtime exits of 70 for ruled stop/witness cases are recorded verbatim and are not ordinary success exits. The original void verdict is retained.

## Isolated timeout reruns at both SHAs

All 19 timed-out programs were rerun serially, one compiler command at a time, using the exact cached admission compiler binaries. Both compilers consumed the same HEAD corpus in the original clean checkout. The command was `adamic build <path> -o <external binary>`, the same native-build command as the gate, with its 10-second limit; a timeout kills the entire process group. Successful builds were executed with a 10-second limit. Main and HEAD compiler binary keys are ab134ca2c57d… and 653a603ac159… respectively. These reruns do not change the original full gate’s void verdict.

| Program | Revision | Compile exit | Compile error | Wall seconds | Native exit if built |
|---|---|---|---|---|---|
| internal/oracle/testdata/captured_union/alternating.ts | main | 1 | none | 0.255 | not run |
| internal/oracle/testdata/captured_union/alternating.ts | HEAD | 0 | none | 4.416 | 0 |
| internal/oracle/testdata/captured_union/by_value.ts | main | 1 | none | 0.110 | not run |
| internal/oracle/testdata/captured_union/by_value.ts | HEAD | 0 | none | 3.874 | 0 |
| internal/oracle/testdata/captured_union/class.ts | main | 1 | none | 0.106 | not run |
| internal/oracle/testdata/captured_union/class.ts | HEAD | 0 | none | 4.691 | 0 |
| internal/oracle/testdata/captured_union/cycle.ts | main | 1 | none | 0.196 | not run |
| internal/oracle/testdata/captured_union/cycle.ts | HEAD | 0 | none | 3.019 | 0 |
| internal/oracle/testdata/captured_union/graph.ts | main | 1 | none | 0.093 | not run |
| internal/oracle/testdata/captured_union/graph.ts | HEAD | -9 | timeout | 10.014 | not run |
| internal/oracle/testdata/captured_union/nested.ts | main | 1 | none | 0.095 | not run |
| internal/oracle/testdata/captured_union/nested.ts | HEAD | 0 | none | 4.226 | 0 |
| internal/oracle/testdata/captured_union/object.ts | main | 1 | none | 0.096 | not run |
| internal/oracle/testdata/captured_union/object.ts | HEAD | 0 | none | 4.177 | 0 |
| internal/oracle/testdata/captured_union/released.ts | main | 1 | none | 0.136 | not run |
| internal/oracle/testdata/captured_union/released.ts | HEAD | 0 | none | 5.159 | 0 |
| internal/oracle/testdata/intersection_views_stops/nested_boolean.ts | main | 0 | none | 2.957 | 70 |
| internal/oracle/testdata/intersection_views_stops/nested_boolean.ts | HEAD | 0 | none | 6.231 | 70 |
| internal/oracle/testdata/intersection_views_stops/nested_boolean_repaired.ts | main | 0 | none | 4.629 | 0 |
| internal/oracle/testdata/intersection_views_stops/nested_boolean_repaired.ts | HEAD | 0 | none | 4.426 | 0 |
| internal/oracle/testdata/intersection_views_stops/nested_literal.ts | main | 0 | none | 3.007 | 70 |
| internal/oracle/testdata/intersection_views_stops/nested_literal.ts | HEAD | 0 | none | 3.138 | 70 |
| internal/oracle/testdata/intersection_views_stops/nested_literal_repaired.ts | main | 0 | none | 3.891 | 0 |
| internal/oracle/testdata/intersection_views_stops/nested_literal_repaired.ts | HEAD | 0 | none | 2.778 | 0 |
| internal/oracle/testdata/intersection_views_stops/nested_token.ts | main | 0 | none | 2.877 | 70 |
| internal/oracle/testdata/intersection_views_stops/nested_token.ts | HEAD | 0 | none | 2.761 | 70 |
| internal/oracle/testdata/intersection_views_stops/nested_token_repaired.ts | main | 0 | none | 2.680 | 0 |
| internal/oracle/testdata/intersection_views_stops/nested_token_repaired.ts | HEAD | 0 | none | 3.470 | 0 |
| internal/oracle/testdata/intersection_views_stops/nested_write.ts | main | 0 | none | 3.593 | 70 |
| internal/oracle/testdata/intersection_views_stops/nested_write.ts | HEAD | 0 | none | 3.029 | 70 |
| internal/oracle/testdata/intersection_views_stops/nested_write_repaired.ts | main | 0 | none | 3.701 | 0 |
| internal/oracle/testdata/intersection_views_stops/nested_write_repaired.ts | HEAD | 0 | none | 2.828 | 0 |
| internal/oracle/testdata/intl_collator/in_compare.a | main | 0 | none | 2.943 | 70 |
| internal/oracle/testdata/intl_collator/in_compare.a | HEAD | 0 | none | 3.332 | 70 |
| internal/oracle/testdata/intl_collator/locale_en.a | main | 0 | none | 2.945 | 70 |
| internal/oracle/testdata/intl_collator/locale_en.a | HEAD | 0 | none | 2.927 | 70 |
| internal/oracle/testdata/reflect_get/stop_literal_getter.a | main | -9 | timeout | 10.019 | not run |
| internal/oracle/testdata/reflect_get/stop_literal_getter.a | HEAD | -9 | timeout | 10.013 | not run |

### Complete isolated observations

```json
[
  {
    "path": "internal/oracle/testdata/captured_union/alternating.ts",
    "revision": "main",
    "command": [
      "/home/agent/.cache/adamic-build/local/ab134ca2c57d0efe0a737cb7da57b7dc172b33c85cb39ee8ea262711b6e7edfb/adamic",
      "build",
      "internal/oracle/testdata/captured_union/alternating.ts",
      "-o",
      "/tmp/n7s-proof/isolated-native-0-main"
    ],
    "native_compile": {
      "exit": 1,
      "wall_seconds": 0.25509275299918954,
      "stdout": "",
      "stderr": "adamic: /workspace/adamic/internal/oracle/testdata/captured_union/alternating.ts:17:4: stage 0 can't lower a union of differently held members variable a function value captures yet\n",
      "error": ""
    }
  },
  {
    "path": "internal/oracle/testdata/captured_union/alternating.ts",
    "revision": "HEAD",
    "command": [
      "/home/agent/.cache/adamic-build/local/653a603ac15988bca5637f1b28210f3c4a193ce1f760c9730fe405bc89de4c33/adamic",
      "build",
      "internal/oracle/testdata/captured_union/alternating.ts",
      "-o",
      "/tmp/n7s-proof/isolated-native-0-HEAD"
    ],
    "native_compile": {
      "exit": 0,
      "wall_seconds": 4.415678121999008,
      "stdout": "",
      "stderr": "",
      "error": ""
    },
    "native": {
      "exit": 0,
      "wall_seconds": 0.0038445689988293452,
      "stdout": "session 0.0 with 2, stopping, undefined, session 0.3 with 2, stopping, undefined, session 0.6 with 2, stopping\nsession 1.0 with 2, stopping, undefined, session 1.3 with 2, stopping, undefined, session 1.6 with 2, stopping\n",
      "stderr": "",
      "error": ""
    }
  },
  {
    "path": "internal/oracle/testdata/captured_union/by_value.ts",
    "revision": "main",
    "command": [
      "/home/agent/.cache/adamic-build/local/ab134ca2c57d0efe0a737cb7da57b7dc172b33c85cb39ee8ea262711b6e7edfb/adamic",
      "build",
      "internal/oracle/testdata/captured_union/by_value.ts",
      "-o",
      "/tmp/n7s-proof/isolated-native-1-main"
    ],
    "native_compile": {
      "exit": 1,
      "wall_seconds": 0.11030315300013172,
      "stdout": "",
      "stderr": "adamic: /workspace/adamic/internal/oracle/testdata/captured_union/by_value.ts:20:4: stage 0 can't lower a union of differently held members variable a function value captures yet\n",
      "error": ""
    }
  },
  {
    "path": "internal/oracle/testdata/captured_union/by_value.ts",
    "revision": "HEAD",
    "command": [
      "/home/agent/.cache/adamic-build/local/653a603ac15988bca5637f1b28210f3c4a193ce1f760c9730fe405bc89de4c33/adamic",
      "build",
      "internal/oracle/testdata/captured_union/by_value.ts",
      "-o",
      "/tmp/n7s-proof/isolated-native-1-HEAD"
    ],
    "native_compile": {
      "exit": 0,
      "wall_seconds": 3.874003167999035,
      "stdout": "",
      "stderr": "",
      "error": ""
    },
    "native": {
      "exit": 0,
      "wall_seconds": 0.0033966809987759916,
      "stdout": "step 0 false false false false true\nstep 1 true true true true false\nstep 2 false false false false true\nstep 3 true true true true false\nthe cell holds the word\nthe cell holds something else\nthe cell holds the word\n",
      "stderr": "",
      "error": ""
    }
  },
  {
    "path": "internal/oracle/testdata/captured_union/class.ts",
    "revision": "main",
    "command": [
      "/home/agent/.cache/adamic-build/local/ab134ca2c57d0efe0a737cb7da57b7dc172b33c85cb39ee8ea262711b6e7edfb/adamic",
      "build",
      "internal/oracle/testdata/captured_union/class.ts",
      "-o",
      "/tmp/n7s-proof/isolated-native-2-main"
    ],
    "native_compile": {
      "exit": 1,
      "wall_seconds": 0.10600376399997913,
      "stdout": "",
      "stderr": "adamic: /workspace/adamic/internal/oracle/testdata/captured_union/class.ts:13:7: stage 0 can't lower a union of differently held members variable a function value captures yet\n",
      "error": ""
    }
  },
  {
    "path": "internal/oracle/testdata/captured_union/class.ts",
    "revision": "HEAD",
    "command": [
      "/home/agent/.cache/adamic-build/local/653a603ac15988bca5637f1b28210f3c4a193ce1f760c9730fe405bc89de4c33/adamic",
      "build",
      "internal/oracle/testdata/captured_union/class.ts",
      "-o",
      "/tmp/n7s-proof/isolated-native-2-HEAD"
    ],
    "native_compile": {
      "exit": 0,
      "wall_seconds": 4.691440386999602,
      "stdout": "",
      "stderr": "",
      "error": ""
    },
    "native": {
      "exit": 0,
      "wall_seconds": 0.003583327001251746,
      "stdout": "start undefined false\nstep 0 session first true object\nstep 1 stopping true string\nstep 2 undefined false undefined\nenabled\nenable while profiled\ndisabled profiled during stopping\nnothing to disable\nenable while stopping\nenabled\nafter a write the cell says string\nstopping\n",
      "stderr": "",
      "error": ""
    }
  },
  {
    "path": "internal/oracle/testdata/captured_union/cycle.ts",
    "revision": "main",
    "command": [
      "/home/agent/.cache/adamic-build/local/ab134ca2c57d0efe0a737cb7da57b7dc172b33c85cb39ee8ea262711b6e7edfb/adamic",
      "build",
      "internal/oracle/testdata/captured_union/cycle.ts",
      "-o",
      "/tmp/n7s-proof/isolated-native-3-main"
    ],
    "native_compile": {
      "exit": 1,
      "wall_seconds": 0.19628457500039076,
      "stdout": "",
      "stderr": "adamic: /workspace/adamic/internal/oracle/testdata/captured_union/cycle.ts:15:7: stage 0 can't lower a union of differently held members variable a function value captures yet\n",
      "error": ""
    }
  },
  {
    "path": "internal/oracle/testdata/captured_union/cycle.ts",
    "revision": "HEAD",
    "command": [
      "/home/agent/.cache/adamic-build/local/653a603ac15988bca5637f1b28210f3c4a193ce1f760c9730fe405bc89de4c33/adamic",
      "build",
      "internal/oracle/testdata/captured_union/cycle.ts",
      "-o",
      "/tmp/n7s-proof/isolated-native-3-HEAD"
    ],
    "native_compile": {
      "exit": 0,
      "wall_seconds": 3.019073801999184,
      "stdout": "",
      "stderr": "",
      "error": ""
    },
    "native": {
      "exit": 0,
      "wall_seconds": 0.0025949160008167382,
      "stdout": "undefined, session first 0, stopping, session second 0, undefined, session last 0\nundefined, session first 1, stopping, session second 1, undefined, session last 1\n",
      "stderr": "",
      "error": ""
    }
  },
  {
    "path": "internal/oracle/testdata/captured_union/graph.ts",
    "revision": "main",
    "command": [
      "/home/agent/.cache/adamic-build/local/ab134ca2c57d0efe0a737cb7da57b7dc172b33c85cb39ee8ea262711b6e7edfb/adamic",
      "build",
      "internal/oracle/testdata/captured_union/graph.ts",
      "-o",
      "/tmp/n7s-proof/isolated-native-4-main"
    ],
    "native_compile": {
      "exit": 1,
      "wall_seconds": 0.09345787100028247,
      "stdout": "",
      "stderr": "adamic: /workspace/adamic/internal/oracle/testdata/captured_union/graph.ts:19:5: stage 0 can't lower a union of differently held members variable a function value captures yet\n",
      "error": ""
    }
  },
  {
    "path": "internal/oracle/testdata/captured_union/graph.ts",
    "revision": "HEAD",
    "command": [
      "/home/agent/.cache/adamic-build/local/653a603ac15988bca5637f1b28210f3c4a193ce1f760c9730fe405bc89de4c33/adamic",
      "build",
      "internal/oracle/testdata/captured_union/graph.ts",
      "-o",
      "/tmp/n7s-proof/isolated-native-4-HEAD"
    ],
    "native_compile": {
      "exit": -9,
      "wall_seconds": 10.013721731998885,
      "stdout": "",
      "stderr": "",
      "error": "timeout"
    }
  },
  {
    "path": "internal/oracle/testdata/captured_union/nested.ts",
    "revision": "main",
    "command": [
      "/home/agent/.cache/adamic-build/local/ab134ca2c57d0efe0a737cb7da57b7dc172b33c85cb39ee8ea262711b6e7edfb/adamic",
      "build",
      "internal/oracle/testdata/captured_union/nested.ts",
      "-o",
      "/tmp/n7s-proof/isolated-native-5-main"
    ],
    "native_compile": {
      "exit": 1,
      "wall_seconds": 0.09517727500133333,
      "stdout": "",
      "stderr": "adamic: /workspace/adamic/internal/oracle/testdata/captured_union/nested.ts:8:5: stage 0 can't lower a union of differently held members variable a function value captures yet\n",
      "error": ""
    }
  },
  {
    "path": "internal/oracle/testdata/captured_union/nested.ts",
    "revision": "HEAD",
    "command": [
      "/home/agent/.cache/adamic-build/local/653a603ac15988bca5637f1b28210f3c4a193ce1f760c9730fe405bc89de4c33/adamic",
      "build",
      "internal/oracle/testdata/captured_union/nested.ts",
      "-o",
      "/tmp/n7s-proof/isolated-native-5-HEAD"
    ],
    "native_compile": {
      "exit": 0,
      "wall_seconds": 4.225810186999297,
      "stdout": "",
      "stderr": "",
      "error": ""
    },
    "native": {
      "exit": 0,
      "wall_seconds": 0.0039006970000627916,
      "stdout": "no base\none base ./base.json\n3 bases ./a.json+./b.json+./c.json\n2 bases ./a.json+./b.json\none base ./base.json\nno base\nno base\n",
      "stderr": "",
      "error": ""
    }
  },
  {
    "path": "internal/oracle/testdata/captured_union/object.ts",
    "revision": "main",
    "command": [
      "/home/agent/.cache/adamic-build/local/ab134ca2c57d0efe0a737cb7da57b7dc172b33c85cb39ee8ea262711b6e7edfb/adamic",
      "build",
      "internal/oracle/testdata/captured_union/object.ts",
      "-o",
      "/tmp/n7s-proof/isolated-native-6-main"
    ],
    "native_compile": {
      "exit": 1,
      "wall_seconds": 0.09647745200163627,
      "stdout": "",
      "stderr": "adamic: /workspace/adamic/internal/oracle/testdata/captured_union/object.ts:7:7: stage 0 can't lower a union of differently held members variable a function value captures yet\n",
      "error": ""
    }
  },
  {
    "path": "internal/oracle/testdata/captured_union/object.ts",
    "revision": "HEAD",
    "command": [
      "/home/agent/.cache/adamic-build/local/653a603ac15988bca5637f1b28210f3c4a193ce1f760c9730fe405bc89de4c33/adamic",
      "build",
      "internal/oracle/testdata/captured_union/object.ts",
      "-o",
      "/tmp/n7s-proof/isolated-native-6-HEAD"
    ],
    "native_compile": {
      "exit": 0,
      "wall_seconds": 4.17722573500032,
      "stdout": "",
      "stderr": "",
      "error": ""
    },
    "native": {
      "exit": 0,
      "wall_seconds": 0.004846235999139026,
      "stdout": "start undefined false\nstep 0 session first true object\nstep 1 stopping true string\nstep 2 undefined false undefined\nenabled\nenable while profiled\ndisabled profiled during stopping\nnothing to disable\nenable while stopping\nenabled\nafter a write the cell says string\nstopping\n",
      "stderr": "",
      "error": ""
    }
  },
  {
    "path": "internal/oracle/testdata/captured_union/released.ts",
    "revision": "main",
    "command": [
      "/home/agent/.cache/adamic-build/local/ab134ca2c57d0efe0a737cb7da57b7dc172b33c85cb39ee8ea262711b6e7edfb/adamic",
      "build",
      "internal/oracle/testdata/captured_union/released.ts",
      "-o",
      "/tmp/n7s-proof/isolated-native-7-main"
    ],
    "native_compile": {
      "exit": 1,
      "wall_seconds": 0.1361335269994015,
      "stdout": "",
      "stderr": "adamic: /workspace/adamic/internal/oracle/testdata/captured_union/released.ts:17:4: stage 0 can't lower a union of differently held members variable a function value captures yet\n",
      "error": ""
    }
  },
  {
    "path": "internal/oracle/testdata/captured_union/released.ts",
    "revision": "HEAD",
    "command": [
      "/home/agent/.cache/adamic-build/local/653a603ac15988bca5637f1b28210f3c4a193ce1f760c9730fe405bc89de4c33/adamic",
      "build",
      "internal/oracle/testdata/captured_union/released.ts",
      "-o",
      "/tmp/n7s-proof/isolated-native-7-HEAD"
    ],
    "native_compile": {
      "exit": 0,
      "wall_seconds": 5.159215805999338,
      "stdout": "",
      "stderr": "",
      "error": ""
    },
    "native": {
      "exit": 0,
      "wall_seconds": 0.0037185359997238265,
      "stdout": "first 0 of 3, stopping, undefined, second 0 of 3, third 0 of 3, stopping, last 0 of 3\nfirst 1 of 3, stopping, undefined, second 1 of 3, third 1 of 3, stopping, last 1 of 3\nfirst 2 of 3, stopping, undefined, second 2 of 3, third 2 of 3, stopping, last 2 of 3\n",
      "stderr": "",
      "error": ""
    }
  },
  {
    "path": "internal/oracle/testdata/intersection_views_stops/nested_boolean.ts",
    "revision": "main",
    "command": [
      "/home/agent/.cache/adamic-build/local/ab134ca2c57d0efe0a737cb7da57b7dc172b33c85cb39ee8ea262711b6e7edfb/adamic",
      "build",
      "internal/oracle/testdata/intersection_views_stops/nested_boolean.ts",
      "-o",
      "/tmp/n7s-proof/isolated-native-8-main"
    ],
    "native_compile": {
      "exit": 0,
      "wall_seconds": 2.9573645899999974,
      "stdout": "",
      "stderr": "",
      "error": ""
    },
    "native": {
      "exit": 70,
      "wall_seconds": 0.0026056269998662174,
      "stdout": "other\n",
      "stderr": "adamic: panic: field read failed: declaration.clause.isTypeOnly expected true, found boolean false\n",
      "error": ""
    }
  },
  {
    "path": "internal/oracle/testdata/intersection_views_stops/nested_boolean.ts",
    "revision": "HEAD",
    "command": [
      "/home/agent/.cache/adamic-build/local/653a603ac15988bca5637f1b28210f3c4a193ce1f760c9730fe405bc89de4c33/adamic",
      "build",
      "internal/oracle/testdata/intersection_views_stops/nested_boolean.ts",
      "-o",
      "/tmp/n7s-proof/isolated-native-8-HEAD"
    ],
    "native_compile": {
      "exit": 0,
      "wall_seconds": 6.23078068399991,
      "stdout": "",
      "stderr": "",
      "error": ""
    },
    "native": {
      "exit": 70,
      "wall_seconds": 0.002947634999145521,
      "stdout": "other\n",
      "stderr": "adamic: panic: field read failed: declaration.clause.isTypeOnly expected true, found boolean false\n",
      "error": ""
    }
  },
  {
    "path": "internal/oracle/testdata/intersection_views_stops/nested_boolean_repaired.ts",
    "revision": "main",
    "command": [
      "/home/agent/.cache/adamic-build/local/ab134ca2c57d0efe0a737cb7da57b7dc172b33c85cb39ee8ea262711b6e7edfb/adamic",
      "build",
      "internal/oracle/testdata/intersection_views_stops/nested_boolean_repaired.ts",
      "-o",
      "/tmp/n7s-proof/isolated-native-9-main"
    ],
    "native_compile": {
      "exit": 0,
      "wall_seconds": 4.62862474800022,
      "stdout": "",
      "stderr": "",
      "error": ""
    },
    "native": {
      "exit": 0,
      "wall_seconds": 0.0034215030009363545,
      "stdout": "other\n4 true\n",
      "stderr": "",
      "error": ""
    }
  },
  {
    "path": "internal/oracle/testdata/intersection_views_stops/nested_boolean_repaired.ts",
    "revision": "HEAD",
    "command": [
      "/home/agent/.cache/adamic-build/local/653a603ac15988bca5637f1b28210f3c4a193ce1f760c9730fe405bc89de4c33/adamic",
      "build",
      "internal/oracle/testdata/intersection_views_stops/nested_boolean_repaired.ts",
      "-o",
      "/tmp/n7s-proof/isolated-native-9-HEAD"
    ],
    "native_compile": {
      "exit": 0,
      "wall_seconds": 4.426144702998499,
      "stdout": "",
      "stderr": "",
      "error": ""
    },
    "native": {
      "exit": 0,
      "wall_seconds": 0.004938633001074777,
      "stdout": "other\n4 true\n",
      "stderr": "",
      "error": ""
    }
  },
  {
    "path": "internal/oracle/testdata/intersection_views_stops/nested_literal.ts",
    "revision": "main",
    "command": [
      "/home/agent/.cache/adamic-build/local/ab134ca2c57d0efe0a737cb7da57b7dc172b33c85cb39ee8ea262711b6e7edfb/adamic",
      "build",
      "internal/oracle/testdata/intersection_views_stops/nested_literal.ts",
      "-o",
      "/tmp/n7s-proof/isolated-native-10-main"
    ],
    "native_compile": {
      "exit": 0,
      "wall_seconds": 3.0068420770003286,
      "stdout": "",
      "stderr": "",
      "error": ""
    },
    "native": {
      "exit": 70,
      "wall_seconds": 0.00339679500029888,
      "stdout": "other\n",
      "stderr": "adamic: panic: field read failed: imported.argument.literal expected 11, found number 12\n",
      "error": ""
    }
  },
  {
    "path": "internal/oracle/testdata/intersection_views_stops/nested_literal.ts",
    "revision": "HEAD",
    "command": [
      "/home/agent/.cache/adamic-build/local/653a603ac15988bca5637f1b28210f3c4a193ce1f760c9730fe405bc89de4c33/adamic",
      "build",
      "internal/oracle/testdata/intersection_views_stops/nested_literal.ts",
      "-o",
      "/tmp/n7s-proof/isolated-native-10-HEAD"
    ],
    "native_compile": {
      "exit": 0,
      "wall_seconds": 3.1382455859984475,
      "stdout": "",
      "stderr": "",
      "error": ""
    },
    "native": {
      "exit": 70,
      "wall_seconds": 0.0027569409994612215,
      "stdout": "other\n",
      "stderr": "adamic: panic: field read failed: imported.argument.literal expected 11, found number 12\n",
      "error": ""
    }
  },
  {
    "path": "internal/oracle/testdata/intersection_views_stops/nested_literal_repaired.ts",
    "revision": "main",
    "command": [
      "/home/agent/.cache/adamic-build/local/ab134ca2c57d0efe0a737cb7da57b7dc172b33c85cb39ee8ea262711b6e7edfb/adamic",
      "build",
      "internal/oracle/testdata/intersection_views_stops/nested_literal_repaired.ts",
      "-o",
      "/tmp/n7s-proof/isolated-native-11-main"
    ],
    "native_compile": {
      "exit": 0,
      "wall_seconds": 3.8905664260000776,
      "stdout": "",
      "stderr": "",
      "error": ""
    },
    "native": {
      "exit": 0,
      "wall_seconds": 0.003382635999514605,
      "stdout": "other\nnode:fs 11\n",
      "stderr": "",
      "error": ""
    }
  },
  {
    "path": "internal/oracle/testdata/intersection_views_stops/nested_literal_repaired.ts",
    "revision": "HEAD",
    "command": [
      "/home/agent/.cache/adamic-build/local/653a603ac15988bca5637f1b28210f3c4a193ce1f760c9730fe405bc89de4c33/adamic",
      "build",
      "internal/oracle/testdata/intersection_views_stops/nested_literal_repaired.ts",
      "-o",
      "/tmp/n7s-proof/isolated-native-11-HEAD"
    ],
    "native_compile": {
      "exit": 0,
      "wall_seconds": 2.7776196809991234,
      "stdout": "",
      "stderr": "",
      "error": ""
    },
    "native": {
      "exit": 0,
      "wall_seconds": 0.0022150139993755147,
      "stdout": "other\nnode:fs 11\n",
      "stderr": "",
      "error": ""
    }
  },
  {
    "path": "internal/oracle/testdata/intersection_views_stops/nested_token.ts",
    "revision": "main",
    "command": [
      "/home/agent/.cache/adamic-build/local/ab134ca2c57d0efe0a737cb7da57b7dc172b33c85cb39ee8ea262711b6e7edfb/adamic",
      "build",
      "internal/oracle/testdata/intersection_views_stops/nested_token.ts",
      "-o",
      "/tmp/n7s-proof/isolated-native-12-main"
    ],
    "native_compile": {
      "exit": 0,
      "wall_seconds": 2.8767134259996965,
      "stdout": "",
      "stderr": "",
      "error": ""
    },
    "native": {
      "exit": 70,
      "wall_seconds": 0.002305668000190053,
      "stdout": "other\n",
      "stderr": "adamic: panic: field read failed: imported.argument.token is not a string; expected string, found number\n",
      "error": ""
    }
  },
  {
    "path": "internal/oracle/testdata/intersection_views_stops/nested_token.ts",
    "revision": "HEAD",
    "command": [
      "/home/agent/.cache/adamic-build/local/653a603ac15988bca5637f1b28210f3c4a193ce1f760c9730fe405bc89de4c33/adamic",
      "build",
      "internal/oracle/testdata/intersection_views_stops/nested_token.ts",
      "-o",
      "/tmp/n7s-proof/isolated-native-12-HEAD"
    ],
    "native_compile": {
      "exit": 0,
      "wall_seconds": 2.761070253000071,
      "stdout": "",
      "stderr": "",
      "error": ""
    },
    "native": {
      "exit": 70,
      "wall_seconds": 0.0028022149999742396,
      "stdout": "other\n",
      "stderr": "adamic: panic: field read failed: imported.argument.token is not a string; expected string, found number\n",
      "error": ""
    }
  },
  {
    "path": "internal/oracle/testdata/intersection_views_stops/nested_token_repaired.ts",
    "revision": "main",
    "command": [
      "/home/agent/.cache/adamic-build/local/ab134ca2c57d0efe0a737cb7da57b7dc172b33c85cb39ee8ea262711b6e7edfb/adamic",
      "build",
      "internal/oracle/testdata/intersection_views_stops/nested_token_repaired.ts",
      "-o",
      "/tmp/n7s-proof/isolated-native-13-main"
    ],
    "native_compile": {
      "exit": 0,
      "wall_seconds": 2.6796453629995085,
      "stdout": "",
      "stderr": "",
      "error": ""
    },
    "native": {
      "exit": 0,
      "wall_seconds": 0.0022113850009191083,
      "stdout": "other\nnode:fs 2\n",
      "stderr": "",
      "error": ""
    }
  },
  {
    "path": "internal/oracle/testdata/intersection_views_stops/nested_token_repaired.ts",
    "revision": "HEAD",
    "command": [
      "/home/agent/.cache/adamic-build/local/653a603ac15988bca5637f1b28210f3c4a193ce1f760c9730fe405bc89de4c33/adamic",
      "build",
      "internal/oracle/testdata/intersection_views_stops/nested_token_repaired.ts",
      "-o",
      "/tmp/n7s-proof/isolated-native-13-HEAD"
    ],
    "native_compile": {
      "exit": 0,
      "wall_seconds": 3.470476607000819,
      "stdout": "",
      "stderr": "",
      "error": ""
    },
    "native": {
      "exit": 0,
      "wall_seconds": 0.004098912000699784,
      "stdout": "other\nnode:fs 2\n",
      "stderr": "",
      "error": ""
    }
  },
  {
    "path": "internal/oracle/testdata/intersection_views_stops/nested_write.ts",
    "revision": "main",
    "command": [
      "/home/agent/.cache/adamic-build/local/ab134ca2c57d0efe0a737cb7da57b7dc172b33c85cb39ee8ea262711b6e7edfb/adamic",
      "build",
      "internal/oracle/testdata/intersection_views_stops/nested_write.ts",
      "-o",
      "/tmp/n7s-proof/isolated-native-14-main"
    ],
    "native_compile": {
      "exit": 0,
      "wall_seconds": 3.593395526999302,
      "stdout": "",
      "stderr": "",
      "error": ""
    },
    "native": {
      "exit": 70,
      "wall_seconds": 0.004157137000220246,
      "stdout": "other\n",
      "stderr": "adamic: panic: field read failed: <write>.flags is not a number; expected number, found string\n",
      "error": ""
    }
  },
  {
    "path": "internal/oracle/testdata/intersection_views_stops/nested_write.ts",
    "revision": "HEAD",
    "command": [
      "/home/agent/.cache/adamic-build/local/653a603ac15988bca5637f1b28210f3c4a193ce1f760c9730fe405bc89de4c33/adamic",
      "build",
      "internal/oracle/testdata/intersection_views_stops/nested_write.ts",
      "-o",
      "/tmp/n7s-proof/isolated-native-14-HEAD"
    ],
    "native_compile": {
      "exit": 0,
      "wall_seconds": 3.029172655000366,
      "stdout": "",
      "stderr": "",
      "error": ""
    },
    "native": {
      "exit": 70,
      "wall_seconds": 0.003284321999672102,
      "stdout": "other\n",
      "stderr": "adamic: panic: field read failed: <write>.flags is not a number; expected number, found string\n",
      "error": ""
    }
  },
  {
    "path": "internal/oracle/testdata/intersection_views_stops/nested_write_repaired.ts",
    "revision": "main",
    "command": [
      "/home/agent/.cache/adamic-build/local/ab134ca2c57d0efe0a737cb7da57b7dc172b33c85cb39ee8ea262711b6e7edfb/adamic",
      "build",
      "internal/oracle/testdata/intersection_views_stops/nested_write_repaired.ts",
      "-o",
      "/tmp/n7s-proof/isolated-native-15-main"
    ],
    "native_compile": {
      "exit": 0,
      "wall_seconds": 3.7010851300001377,
      "stdout": "",
      "stderr": "",
      "error": ""
    },
    "native": {
      "exit": 0,
      "wall_seconds": 0.0031188330012810184,
      "stdout": "other\nstamped\nnumber\n",
      "stderr": "",
      "error": ""
    }
  },
  {
    "path": "internal/oracle/testdata/intersection_views_stops/nested_write_repaired.ts",
    "revision": "HEAD",
    "command": [
      "/home/agent/.cache/adamic-build/local/653a603ac15988bca5637f1b28210f3c4a193ce1f760c9730fe405bc89de4c33/adamic",
      "build",
      "internal/oracle/testdata/intersection_views_stops/nested_write_repaired.ts",
      "-o",
      "/tmp/n7s-proof/isolated-native-15-HEAD"
    ],
    "native_compile": {
      "exit": 0,
      "wall_seconds": 2.8284298710004805,
      "stdout": "",
      "stderr": "",
      "error": ""
    },
    "native": {
      "exit": 0,
      "wall_seconds": 0.0031257890004781075,
      "stdout": "other\nstamped\nnumber\n",
      "stderr": "",
      "error": ""
    }
  },
  {
    "path": "internal/oracle/testdata/intl_collator/in_compare.a",
    "revision": "main",
    "command": [
      "/home/agent/.cache/adamic-build/local/ab134ca2c57d0efe0a737cb7da57b7dc172b33c85cb39ee8ea262711b6e7edfb/adamic",
      "build",
      "internal/oracle/testdata/intl_collator/in_compare.a",
      "-o",
      "/tmp/n7s-proof/isolated-native-16-main"
    ],
    "native_compile": {
      "exit": 0,
      "wall_seconds": 2.9431272380006703,
      "stdout": "",
      "stderr": "",
      "error": ""
    },
    "native": {
      "exit": 70,
      "wall_seconds": 0.005685378000634955,
      "stdout": "collator\n",
      "stderr": "adamic: panic: this runtime value's properties aren't read by a run-time key yet (reading 'compare')\n",
      "error": ""
    }
  },
  {
    "path": "internal/oracle/testdata/intl_collator/in_compare.a",
    "revision": "HEAD",
    "command": [
      "/home/agent/.cache/adamic-build/local/653a603ac15988bca5637f1b28210f3c4a193ce1f760c9730fe405bc89de4c33/adamic",
      "build",
      "internal/oracle/testdata/intl_collator/in_compare.a",
      "-o",
      "/tmp/n7s-proof/isolated-native-16-HEAD"
    ],
    "native_compile": {
      "exit": 0,
      "wall_seconds": 3.3324476100005995,
      "stdout": "",
      "stderr": "",
      "error": ""
    },
    "native": {
      "exit": 70,
      "wall_seconds": 0.0025213829994754633,
      "stdout": "collator\n",
      "stderr": "adamic: panic: this runtime value's properties aren't read by a run-time key yet (reading 'compare')\n",
      "error": ""
    }
  },
  {
    "path": "internal/oracle/testdata/intl_collator/locale_en.a",
    "revision": "main",
    "command": [
      "/home/agent/.cache/adamic-build/local/ab134ca2c57d0efe0a737cb7da57b7dc172b33c85cb39ee8ea262711b6e7edfb/adamic",
      "build",
      "internal/oracle/testdata/intl_collator/locale_en.a",
      "-o",
      "/tmp/n7s-proof/isolated-native-17-main"
    ],
    "native_compile": {
      "exit": 0,
      "wall_seconds": 2.94467000899931,
      "stdout": "",
      "stderr": "",
      "error": ""
    },
    "native": {
      "exit": 70,
      "wall_seconds": 0.0031395790010719793,
      "stdout": "before\n",
      "stderr": "adamic: panic: Intl.Collator with a defined locale isn't built yet\n",
      "error": ""
    }
  },
  {
    "path": "internal/oracle/testdata/intl_collator/locale_en.a",
    "revision": "HEAD",
    "command": [
      "/home/agent/.cache/adamic-build/local/653a603ac15988bca5637f1b28210f3c4a193ce1f760c9730fe405bc89de4c33/adamic",
      "build",
      "internal/oracle/testdata/intl_collator/locale_en.a",
      "-o",
      "/tmp/n7s-proof/isolated-native-17-HEAD"
    ],
    "native_compile": {
      "exit": 0,
      "wall_seconds": 2.9273867170013546,
      "stdout": "",
      "stderr": "",
      "error": ""
    },
    "native": {
      "exit": 70,
      "wall_seconds": 0.004627203999916674,
      "stdout": "before\n",
      "stderr": "adamic: panic: Intl.Collator with a defined locale isn't built yet\n",
      "error": ""
    }
  },
  {
    "path": "internal/oracle/testdata/reflect_get/stop_literal_getter.a",
    "revision": "main",
    "command": [
      "/home/agent/.cache/adamic-build/local/ab134ca2c57d0efe0a737cb7da57b7dc172b33c85cb39ee8ea262711b6e7edfb/adamic",
      "build",
      "internal/oracle/testdata/reflect_get/stop_literal_getter.a",
      "-o",
      "/tmp/n7s-proof/isolated-native-18-main"
    ],
    "native_compile": {
      "exit": -9,
      "wall_seconds": 10.018813663000401,
      "stdout": "",
      "stderr": "",
      "error": "timeout"
    }
  },
  {
    "path": "internal/oracle/testdata/reflect_get/stop_literal_getter.a",
    "revision": "HEAD",
    "command": [
      "/home/agent/.cache/adamic-build/local/653a603ac15988bca5637f1b28210f3c4a193ce1f760c9730fe405bc89de4c33/adamic",
      "build",
      "internal/oracle/testdata/reflect_get/stop_literal_getter.a",
      "-o",
      "/tmp/n7s-proof/isolated-native-18-HEAD"
    ],
    "native_compile": {
      "exit": -9,
      "wall_seconds": 10.013136592000592,
      "stdout": "",
      "stderr": "",
      "error": "timeout"
    }
  }
]
```


## Verbatim pin and mutant logs

These logs preserve every graph-cell line, counts line, red first diagnostic, and complete logged got/error payload without truncation. Paths, sanitizer addresses, and process IDs are left as observed.

### head-lower.log

```text
=== RUN   TestCapturedUnionLowersToOneCell
=== PAUSE TestCapturedUnionLowersToOneCell
=== RUN   TestCapturedMaybeBooleanStillStops
=== PAUSE TestCapturedMaybeBooleanStillStops
=== CONT  TestCapturedUnionLowersToOneCell
=== CONT  TestCapturedMaybeBooleanStillStops
=== RUN   TestCapturedUnionLowersToOneCell/class.ts
=== PAUSE TestCapturedUnionLowersToOneCell/class.ts
=== RUN   TestCapturedUnionLowersToOneCell/class.a
=== PAUSE TestCapturedUnionLowersToOneCell/class.a
=== RUN   TestCapturedUnionLowersToOneCell/object.ts
=== PAUSE TestCapturedUnionLowersToOneCell/object.ts
=== RUN   TestCapturedUnionLowersToOneCell/object.a
=== PAUSE TestCapturedUnionLowersToOneCell/object.a
=== RUN   TestCapturedUnionLowersToOneCell/by_value.ts
=== PAUSE TestCapturedUnionLowersToOneCell/by_value.ts
=== RUN   TestCapturedUnionLowersToOneCell/by_value.a
=== PAUSE TestCapturedUnionLowersToOneCell/by_value.a
=== RUN   TestCapturedUnionLowersToOneCell/released.ts
=== PAUSE TestCapturedUnionLowersToOneCell/released.ts
=== RUN   TestCapturedUnionLowersToOneCell/released.a
=== PAUSE TestCapturedUnionLowersToOneCell/released.a
=== RUN   TestCapturedUnionLowersToOneCell/nested.ts
=== PAUSE TestCapturedUnionLowersToOneCell/nested.ts
=== RUN   TestCapturedUnionLowersToOneCell/nested.a
=== PAUSE TestCapturedUnionLowersToOneCell/nested.a
=== RUN   TestCapturedUnionLowersToOneCell/alternating.ts
=== PAUSE TestCapturedUnionLowersToOneCell/alternating.ts
=== RUN   TestCapturedUnionLowersToOneCell/alternating.a
=== PAUSE TestCapturedUnionLowersToOneCell/alternating.a
=== RUN   TestCapturedUnionLowersToOneCell/graph.ts
=== PAUSE TestCapturedUnionLowersToOneCell/graph.ts
=== RUN   TestCapturedUnionLowersToOneCell/graph.a
=== PAUSE TestCapturedUnionLowersToOneCell/graph.a
=== RUN   TestCapturedUnionLowersToOneCell/cycle.ts
=== PAUSE TestCapturedUnionLowersToOneCell/cycle.ts
=== RUN   TestCapturedUnionLowersToOneCell/cycle.a
=== PAUSE TestCapturedUnionLowersToOneCell/cycle.a
=== CONT  TestCapturedUnionLowersToOneCell/class.ts
=== CONT  TestCapturedUnionLowersToOneCell/nested.ts
=== CONT  TestCapturedUnionLowersToOneCell/by_value.ts
=== CONT  TestCapturedUnionLowersToOneCell/graph.ts
=== NAME  TestCapturedUnionLowersToOneCell/nested.ts
    captured_union_test.go:48: extended: graph cell true
=== CONT  TestCapturedUnionLowersToOneCell/alternating.a
=== NAME  TestCapturedUnionLowersToOneCell/class.ts
    captured_union_test.go:48: active: graph cell true
=== CONT  TestCapturedUnionLowersToOneCell/object.a
--- PASS: TestCapturedMaybeBooleanStillStops (0.69s)
=== CONT  TestCapturedUnionLowersToOneCell/alternating.ts
=== NAME  TestCapturedUnionLowersToOneCell/by_value.ts
    captured_union_test.go:48: active: graph cell false
=== CONT  TestCapturedUnionLowersToOneCell/object.ts
=== NAME  TestCapturedUnionLowersToOneCell/alternating.a
    captured_union_test.go:48: active: graph cell false
=== CONT  TestCapturedUnionLowersToOneCell/nested.a
=== NAME  TestCapturedUnionLowersToOneCell/object.ts
    captured_union_test.go:48: active: graph cell true
=== CONT  TestCapturedUnionLowersToOneCell/class.a
=== NAME  TestCapturedUnionLowersToOneCell/graph.ts
    captured_union_test.go:48: active: graph cell true
=== CONT  TestCapturedUnionLowersToOneCell/released.ts
=== NAME  TestCapturedUnionLowersToOneCell/alternating.ts
    captured_union_test.go:48: active: graph cell false
=== CONT  TestCapturedUnionLowersToOneCell/released.a
=== NAME  TestCapturedUnionLowersToOneCell/object.a
    captured_union_test.go:48: active: graph cell true
=== CONT  TestCapturedUnionLowersToOneCell/cycle.ts
=== NAME  TestCapturedUnionLowersToOneCell/released.ts
    captured_union_test.go:48: active: graph cell false
=== CONT  TestCapturedUnionLowersToOneCell/cycle.a
=== NAME  TestCapturedUnionLowersToOneCell/nested.a
    captured_union_test.go:48: extended: graph cell true
=== CONT  TestCapturedUnionLowersToOneCell/by_value.a
=== NAME  TestCapturedUnionLowersToOneCell/released.a
    captured_union_test.go:48: active: graph cell false
=== CONT  TestCapturedUnionLowersToOneCell/graph.a
=== NAME  TestCapturedUnionLowersToOneCell/class.a
    captured_union_test.go:48: active: graph cell true
=== NAME  TestCapturedUnionLowersToOneCell/cycle.ts
    captured_union_test.go:48: active: graph cell true
    captured_union_test.go:59: active: 1 closures carry its cell, want every one that reads or writes it
=== NAME  TestCapturedUnionLowersToOneCell/cycle.a
    captured_union_test.go:48: active: graph cell true
    captured_union_test.go:59: active: 1 closures carry its cell, want every one that reads or writes it
=== NAME  TestCapturedUnionLowersToOneCell/by_value.a
    captured_union_test.go:48: active: graph cell false
=== NAME  TestCapturedUnionLowersToOneCell/graph.a
    captured_union_test.go:48: active: graph cell true
--- FAIL: TestCapturedUnionLowersToOneCell (0.00s)
    --- PASS: TestCapturedUnionLowersToOneCell/nested.ts (0.55s)
    --- PASS: TestCapturedUnionLowersToOneCell/class.ts (0.65s)
    --- PASS: TestCapturedUnionLowersToOneCell/by_value.ts (1.01s)
    --- PASS: TestCapturedUnionLowersToOneCell/alternating.a (0.57s)
    --- PASS: TestCapturedUnionLowersToOneCell/object.ts (0.13s)
    --- PASS: TestCapturedUnionLowersToOneCell/graph.ts (1.14s)
    --- PASS: TestCapturedUnionLowersToOneCell/alternating.ts (0.50s)
    --- PASS: TestCapturedUnionLowersToOneCell/object.a (0.57s)
    --- PASS: TestCapturedUnionLowersToOneCell/released.ts (0.14s)
    --- PASS: TestCapturedUnionLowersToOneCell/nested.a (0.21s)
    --- PASS: TestCapturedUnionLowersToOneCell/released.a (0.15s)
    --- PASS: TestCapturedUnionLowersToOneCell/class.a (0.20s)
    --- FAIL: TestCapturedUnionLowersToOneCell/cycle.ts (0.14s)
    --- FAIL: TestCapturedUnionLowersToOneCell/cycle.a (0.11s)
    --- PASS: TestCapturedUnionLowersToOneCell/by_value.a (0.07s)
    --- PASS: TestCapturedUnionLowersToOneCell/graph.a (0.07s)
FAIL
FAIL	github.com/system-inc/adamic/internal/lower	1.551s
FAIL
```


### head-oracle.log

```text
=== RUN   TestCapturedUnionAgrees
=== PAUSE TestCapturedUnionAgrees
=== CONT  TestCapturedUnionAgrees
=== RUN   TestCapturedUnionAgrees/class.ts
=== PAUSE TestCapturedUnionAgrees/class.ts
=== RUN   TestCapturedUnionAgrees/class.a
=== PAUSE TestCapturedUnionAgrees/class.a
=== RUN   TestCapturedUnionAgrees/object.ts
=== PAUSE TestCapturedUnionAgrees/object.ts
=== RUN   TestCapturedUnionAgrees/object.a
=== PAUSE TestCapturedUnionAgrees/object.a
=== RUN   TestCapturedUnionAgrees/by_value.ts
=== PAUSE TestCapturedUnionAgrees/by_value.ts
=== RUN   TestCapturedUnionAgrees/by_value.a
=== PAUSE TestCapturedUnionAgrees/by_value.a
=== RUN   TestCapturedUnionAgrees/released.ts
=== PAUSE TestCapturedUnionAgrees/released.ts
=== RUN   TestCapturedUnionAgrees/released.a
=== PAUSE TestCapturedUnionAgrees/released.a
=== RUN   TestCapturedUnionAgrees/nested.ts
=== PAUSE TestCapturedUnionAgrees/nested.ts
=== RUN   TestCapturedUnionAgrees/nested.a
=== PAUSE TestCapturedUnionAgrees/nested.a
=== RUN   TestCapturedUnionAgrees/alternating.ts
=== PAUSE TestCapturedUnionAgrees/alternating.ts
=== RUN   TestCapturedUnionAgrees/alternating.a
=== PAUSE TestCapturedUnionAgrees/alternating.a
=== RUN   TestCapturedUnionAgrees/graph.ts
=== PAUSE TestCapturedUnionAgrees/graph.ts
=== RUN   TestCapturedUnionAgrees/graph.a
=== PAUSE TestCapturedUnionAgrees/graph.a
=== RUN   TestCapturedUnionAgrees/cycle.ts
=== PAUSE TestCapturedUnionAgrees/cycle.ts
=== RUN   TestCapturedUnionAgrees/cycle.a
=== PAUSE TestCapturedUnionAgrees/cycle.a
=== CONT  TestCapturedUnionAgrees/class.ts
=== CONT  TestCapturedUnionAgrees/nested.ts
=== CONT  TestCapturedUnionAgrees/by_value.ts
=== CONT  TestCapturedUnionAgrees/object.ts
=== CONT  TestCapturedUnionAgrees/released.ts
=== RUN   TestCapturedUnionAgrees/by_value.ts/native
=== RUN   TestCapturedUnionAgrees/object.ts/native
=== RUN   TestCapturedUnionAgrees/nested.ts/native
=== RUN   TestCapturedUnionAgrees/released.ts/native
=== RUN   TestCapturedUnionAgrees/class.ts/native
=== RUN   TestCapturedUnionAgrees/released.ts/sanitized
=== RUN   TestCapturedUnionAgrees/nested.ts/sanitized
=== RUN   TestCapturedUnionAgrees/by_value.ts/sanitized
=== RUN   TestCapturedUnionAgrees/object.ts/sanitized
=== RUN   TestCapturedUnionAgrees/class.ts/sanitized
=== RUN   TestCapturedUnionAgrees/released.ts/counted
=== RUN   TestCapturedUnionAgrees/nested.ts/counted
=== RUN   TestCapturedUnionAgrees/object.ts/counted
=== RUN   TestCapturedUnionAgrees/by_value.ts/counted
=== NAME  TestCapturedUnionAgrees/nested.ts/counted
    captured_union_test.go:116: counts: allocations 97, frees 97, in regions 0
=== RUN   TestCapturedUnionAgrees/nested.ts/javascript
=== NAME  TestCapturedUnionAgrees/released.ts/counted
    captured_union_test.go:116: counts: allocations 96, frees 96, in regions 0
=== RUN   TestCapturedUnionAgrees/released.ts/javascript
=== RUN   TestCapturedUnionAgrees/class.ts/counted
=== NAME  TestCapturedUnionAgrees/by_value.ts/counted
    captured_union_test.go:116: counts: allocations 35, frees 35, in regions 0
=== RUN   TestCapturedUnionAgrees/by_value.ts/javascript
=== NAME  TestCapturedUnionAgrees/object.ts/counted
    captured_union_test.go:116: counts: allocations 21, frees 21, in regions 0
=== RUN   TestCapturedUnionAgrees/object.ts/javascript
=== NAME  TestCapturedUnionAgrees/class.ts/counted
    captured_union_test.go:116: counts: allocations 21, frees 21, in regions 0
=== RUN   TestCapturedUnionAgrees/class.ts/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a miss 464.06
=== NAME  TestCapturedUnionAgrees/by_value.ts/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 464.06
=== NAME  TestCapturedUnionAgrees/object.ts/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 464.06
=== NAME  TestCapturedUnionAgrees/nested.ts/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 464.06
=== NAME  TestCapturedUnionAgrees/released.ts/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 464.06
=== CONT  TestCapturedUnionAgrees/released.a
=== CONT  TestCapturedUnionAgrees/by_value.a
=== CONT  TestCapturedUnionAgrees/class.a
=== CONT  TestCapturedUnionAgrees/object.a
=== CONT  TestCapturedUnionAgrees/graph.ts
=== RUN   TestCapturedUnionAgrees/released.a/native
=== RUN   TestCapturedUnionAgrees/class.a/native
=== RUN   TestCapturedUnionAgrees/by_value.a/native
=== RUN   TestCapturedUnionAgrees/graph.ts/native
=== RUN   TestCapturedUnionAgrees/object.a/native
=== RUN   TestCapturedUnionAgrees/released.a/sanitized
=== RUN   TestCapturedUnionAgrees/by_value.a/sanitized
=== RUN   TestCapturedUnionAgrees/object.a/sanitized
=== RUN   TestCapturedUnionAgrees/class.a/sanitized
=== RUN   TestCapturedUnionAgrees/released.a/counted
=== RUN   TestCapturedUnionAgrees/by_value.a/counted
=== RUN   TestCapturedUnionAgrees/object.a/counted
=== NAME  TestCapturedUnionAgrees/released.a/counted
    captured_union_test.go:116: counts: allocations 96, frees 96, in regions 0
=== RUN   TestCapturedUnionAgrees/class.a/counted
=== RUN   TestCapturedUnionAgrees/released.a/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 0.03
=== NAME  TestCapturedUnionAgrees/by_value.a/counted
    captured_union_test.go:116: counts: allocations 35, frees 35, in regions 0
=== RUN   TestCapturedUnionAgrees/by_value.a/javascript
=== CONT  TestCapturedUnionAgrees/cycle.a
=== NAME  TestCapturedUnionAgrees/by_value.a/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 0.02
=== NAME  TestCapturedUnionAgrees/object.a/counted
    captured_union_test.go:116: counts: allocations 21, frees 21, in regions 0
=== RUN   TestCapturedUnionAgrees/object.a/javascript
=== CONT  TestCapturedUnionAgrees/alternating.ts
=== NAME  TestCapturedUnionAgrees/object.a/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 0.01
=== CONT  TestCapturedUnionAgrees/cycle.ts
=== RUN   TestCapturedUnionAgrees/cycle.a/native
=== NAME  TestCapturedUnionAgrees/class.a/counted
    captured_union_test.go:116: counts: allocations 21, frees 21, in regions 0
=== RUN   TestCapturedUnionAgrees/class.a/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 0.02
=== RUN   TestCapturedUnionAgrees/alternating.ts/native
=== CONT  TestCapturedUnionAgrees/alternating.a
=== RUN   TestCapturedUnionAgrees/cycle.ts/native
=== RUN   TestCapturedUnionAgrees/cycle.a/sanitized
=== RUN   TestCapturedUnionAgrees/alternating.a/native
=== RUN   TestCapturedUnionAgrees/alternating.ts/sanitized
=== RUN   TestCapturedUnionAgrees/cycle.ts/sanitized
=== RUN   TestCapturedUnionAgrees/alternating.a/sanitized
=== RUN   TestCapturedUnionAgrees/cycle.a/counted
=== RUN   TestCapturedUnionAgrees/alternating.ts/counted
=== RUN   TestCapturedUnionAgrees/cycle.ts/counted
=== NAME  TestCapturedUnionAgrees/cycle.a/counted
    captured_union_test.go:116: counts: allocations 32, frees 32, in regions 0
=== RUN   TestCapturedUnionAgrees/cycle.a/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 0.01
=== RUN   TestCapturedUnionAgrees/alternating.a/counted
=== CONT  TestCapturedUnionAgrees/graph.a
=== NAME  TestCapturedUnionAgrees/alternating.a/counted
    captured_union_test.go:116: counts: allocations 54, frees 54, in regions 0
=== RUN   TestCapturedUnionAgrees/alternating.a/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 0.05
=== CONT  TestCapturedUnionAgrees/nested.a
=== RUN   TestCapturedUnionAgrees/graph.a/native
=== RUN   TestCapturedUnionAgrees/nested.a/native
=== RUN   TestCapturedUnionAgrees/nested.a/sanitized
=== RUN   TestCapturedUnionAgrees/nested.a/counted
    captured_union_test.go:116: counts: allocations 97, frees 97, in regions 0
=== RUN   TestCapturedUnionAgrees/nested.a/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 0.01
=== NAME  TestCapturedUnionAgrees/cycle.ts/counted
    captured_union_test.go:116: counts: allocations 32, frees 32, in regions 0
=== RUN   TestCapturedUnionAgrees/cycle.ts/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 0.02
=== NAME  TestCapturedUnionAgrees/alternating.ts/counted
    captured_union_test.go:116: counts: allocations 54, frees 54, in regions 0
=== RUN   TestCapturedUnionAgrees/alternating.ts/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 0.01
=== RUN   TestCapturedUnionAgrees/graph.a/sanitized
=== RUN   TestCapturedUnionAgrees/graph.ts/sanitized
=== RUN   TestCapturedUnionAgrees/graph.a/counted
=== RUN   TestCapturedUnionAgrees/graph.ts/counted
    captured_union_test.go:116: counts: allocations 42, frees 42, in regions 0
=== RUN   TestCapturedUnionAgrees/graph.ts/javascript
=== NAME  TestCapturedUnionAgrees/graph.a/counted
    captured_union_test.go:116: counts: allocations 42, frees 42, in regions 0
=== RUN   TestCapturedUnionAgrees/graph.a/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 0.01
=== NAME  TestCapturedUnionAgrees/graph.ts/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 0.01
--- PASS: TestCapturedUnionAgrees (0.00s)
    --- PASS: TestCapturedUnionAgrees/class.ts (595.09s)
        --- PASS: TestCapturedUnionAgrees/class.ts/native (0.48s)
        --- PASS: TestCapturedUnionAgrees/class.ts/sanitized (1.26s)
        --- PASS: TestCapturedUnionAgrees/class.ts/counted (0.34s)
        --- PASS: TestCapturedUnionAgrees/class.ts/javascript (464.22s)
    --- PASS: TestCapturedUnionAgrees/object.ts (595.13s)
        --- PASS: TestCapturedUnionAgrees/object.ts/native (0.61s)
        --- PASS: TestCapturedUnionAgrees/object.ts/sanitized (0.70s)
        --- PASS: TestCapturedUnionAgrees/object.ts/counted (0.61s)
        --- PASS: TestCapturedUnionAgrees/object.ts/javascript (464.55s)
    --- PASS: TestCapturedUnionAgrees/by_value.ts (595.13s)
        --- PASS: TestCapturedUnionAgrees/by_value.ts/native (0.45s)
        --- PASS: TestCapturedUnionAgrees/by_value.ts/sanitized (0.97s)
        --- PASS: TestCapturedUnionAgrees/by_value.ts/counted (0.50s)
        --- PASS: TestCapturedUnionAgrees/by_value.ts/javascript (464.57s)
    --- PASS: TestCapturedUnionAgrees/released.ts (595.14s)
        --- PASS: TestCapturedUnionAgrees/released.ts/native (0.29s)
        --- PASS: TestCapturedUnionAgrees/released.ts/sanitized (0.86s)
        --- PASS: TestCapturedUnionAgrees/released.ts/counted (0.50s)
        --- PASS: TestCapturedUnionAgrees/released.ts/javascript (464.72s)
    --- PASS: TestCapturedUnionAgrees/nested.ts (595.14s)
        --- PASS: TestCapturedUnionAgrees/nested.ts/native (0.34s)
        --- PASS: TestCapturedUnionAgrees/nested.ts/sanitized (0.86s)
        --- PASS: TestCapturedUnionAgrees/nested.ts/counted (0.29s)
        --- PASS: TestCapturedUnionAgrees/nested.ts/javascript (464.93s)
    --- PASS: TestCapturedUnionAgrees/released.a (2.06s)
        --- PASS: TestCapturedUnionAgrees/released.a/native (0.48s)
        --- PASS: TestCapturedUnionAgrees/released.a/sanitized (0.56s)
        --- PASS: TestCapturedUnionAgrees/released.a/counted (0.43s)
        --- PASS: TestCapturedUnionAgrees/released.a/javascript (0.22s)
    --- PASS: TestCapturedUnionAgrees/by_value.a (2.17s)
        --- PASS: TestCapturedUnionAgrees/by_value.a/native (0.39s)
        --- PASS: TestCapturedUnionAgrees/by_value.a/sanitized (0.67s)
        --- PASS: TestCapturedUnionAgrees/by_value.a/counted (0.51s)
        --- PASS: TestCapturedUnionAgrees/by_value.a/javascript (0.16s)
    --- PASS: TestCapturedUnionAgrees/object.a (2.28s)
        --- PASS: TestCapturedUnionAgrees/object.a/native (0.41s)
        --- PASS: TestCapturedUnionAgrees/object.a/sanitized (0.72s)
        --- PASS: TestCapturedUnionAgrees/object.a/counted (0.55s)
        --- PASS: TestCapturedUnionAgrees/object.a/javascript (0.12s)
    --- PASS: TestCapturedUnionAgrees/class.a (2.56s)
        --- PASS: TestCapturedUnionAgrees/class.a/native (0.54s)
        --- PASS: TestCapturedUnionAgrees/class.a/sanitized (0.84s)
        --- PASS: TestCapturedUnionAgrees/class.a/counted (0.63s)
        --- PASS: TestCapturedUnionAgrees/class.a/javascript (0.13s)
    --- PASS: TestCapturedUnionAgrees/cycle.a (1.76s)
        --- PASS: TestCapturedUnionAgrees/cycle.a/native (0.34s)
        --- PASS: TestCapturedUnionAgrees/cycle.a/sanitized (0.64s)
        --- PASS: TestCapturedUnionAgrees/cycle.a/counted (0.39s)
        --- PASS: TestCapturedUnionAgrees/cycle.a/javascript (0.11s)
    --- PASS: TestCapturedUnionAgrees/alternating.a (1.72s)
        --- PASS: TestCapturedUnionAgrees/alternating.a/native (0.35s)
        --- PASS: TestCapturedUnionAgrees/alternating.a/sanitized (0.57s)
        --- PASS: TestCapturedUnionAgrees/alternating.a/counted (0.44s)
        --- PASS: TestCapturedUnionAgrees/alternating.a/javascript (0.14s)
    --- PASS: TestCapturedUnionAgrees/nested.a (1.67s)
        --- PASS: TestCapturedUnionAgrees/nested.a/native (0.24s)
        --- PASS: TestCapturedUnionAgrees/nested.a/sanitized (0.62s)
        --- PASS: TestCapturedUnionAgrees/nested.a/counted (0.28s)
        --- PASS: TestCapturedUnionAgrees/nested.a/javascript (0.13s)
    --- PASS: TestCapturedUnionAgrees/alternating.ts (4.70s)
        --- PASS: TestCapturedUnionAgrees/alternating.ts/native (0.32s)
        --- PASS: TestCapturedUnionAgrees/alternating.ts/sanitized (0.68s)
        --- PASS: TestCapturedUnionAgrees/alternating.ts/counted (3.17s)
        --- PASS: TestCapturedUnionAgrees/alternating.ts/javascript (0.22s)
    --- PASS: TestCapturedUnionAgrees/cycle.ts (4.92s)
        --- PASS: TestCapturedUnionAgrees/cycle.ts/native (0.38s)
        --- PASS: TestCapturedUnionAgrees/cycle.ts/sanitized (0.66s)
        --- PASS: TestCapturedUnionAgrees/cycle.ts/counted (2.93s)
        --- PASS: TestCapturedUnionAgrees/cycle.ts/javascript (0.61s)
    --- PASS: TestCapturedUnionAgrees/graph.a (42.16s)
        --- PASS: TestCapturedUnionAgrees/graph.a/native (9.26s)
        --- PASS: TestCapturedUnionAgrees/graph.a/sanitized (18.57s)
        --- PASS: TestCapturedUnionAgrees/graph.a/counted (13.55s)
        --- PASS: TestCapturedUnionAgrees/graph.a/javascript (0.09s)
    --- PASS: TestCapturedUnionAgrees/graph.ts (45.94s)
        --- PASS: TestCapturedUnionAgrees/graph.ts/native (13.31s)
        --- PASS: TestCapturedUnionAgrees/graph.ts/sanitized (21.11s)
        --- PASS: TestCapturedUnionAgrees/graph.ts/counted (11.00s)
        --- PASS: TestCapturedUnionAgrees/graph.ts/javascript (0.09s)
PASS
gate cache: native hits=0 misses=32
gate cache: node hits=0 misses=32
gate cache: probe hits=0 misses=0
ok  	github.com/system-inc/adamic/internal/oracle	641.106s
```


### M1-lower.log

```text
=== RUN   TestCapturedUnionLowersToOneCell
=== PAUSE TestCapturedUnionLowersToOneCell
=== RUN   TestCapturedMaybeBooleanStillStops
=== PAUSE TestCapturedMaybeBooleanStillStops
=== CONT  TestCapturedUnionLowersToOneCell
=== RUN   TestCapturedUnionLowersToOneCell/class.ts
=== PAUSE TestCapturedUnionLowersToOneCell/class.ts
=== RUN   TestCapturedUnionLowersToOneCell/class.a
=== PAUSE TestCapturedUnionLowersToOneCell/class.a
=== RUN   TestCapturedUnionLowersToOneCell/object.ts
=== PAUSE TestCapturedUnionLowersToOneCell/object.ts
=== RUN   TestCapturedUnionLowersToOneCell/object.a
=== PAUSE TestCapturedUnionLowersToOneCell/object.a
=== RUN   TestCapturedUnionLowersToOneCell/by_value.ts
=== CONT  TestCapturedMaybeBooleanStillStops
=== PAUSE TestCapturedUnionLowersToOneCell/by_value.ts
=== RUN   TestCapturedUnionLowersToOneCell/by_value.a
=== PAUSE TestCapturedUnionLowersToOneCell/by_value.a
=== RUN   TestCapturedUnionLowersToOneCell/released.ts
=== PAUSE TestCapturedUnionLowersToOneCell/released.ts
=== RUN   TestCapturedUnionLowersToOneCell/released.a
=== PAUSE TestCapturedUnionLowersToOneCell/released.a
=== RUN   TestCapturedUnionLowersToOneCell/nested.ts
=== PAUSE TestCapturedUnionLowersToOneCell/nested.ts
=== RUN   TestCapturedUnionLowersToOneCell/nested.a
=== PAUSE TestCapturedUnionLowersToOneCell/nested.a
=== RUN   TestCapturedUnionLowersToOneCell/alternating.ts
=== PAUSE TestCapturedUnionLowersToOneCell/alternating.ts
=== RUN   TestCapturedUnionLowersToOneCell/alternating.a
=== PAUSE TestCapturedUnionLowersToOneCell/alternating.a
=== RUN   TestCapturedUnionLowersToOneCell/graph.ts
=== PAUSE TestCapturedUnionLowersToOneCell/graph.ts
=== RUN   TestCapturedUnionLowersToOneCell/graph.a
=== PAUSE TestCapturedUnionLowersToOneCell/graph.a
=== RUN   TestCapturedUnionLowersToOneCell/cycle.ts
=== PAUSE TestCapturedUnionLowersToOneCell/cycle.ts
=== RUN   TestCapturedUnionLowersToOneCell/cycle.a
=== PAUSE TestCapturedUnionLowersToOneCell/cycle.a
=== CONT  TestCapturedUnionLowersToOneCell/class.ts
=== CONT  TestCapturedUnionLowersToOneCell/nested.ts
=== CONT  TestCapturedUnionLowersToOneCell/released.a
=== CONT  TestCapturedUnionLowersToOneCell/cycle.a
--- PASS: TestCapturedMaybeBooleanStillStops (0.31s)
=== CONT  TestCapturedUnionLowersToOneCell/released.ts
=== NAME  TestCapturedUnionLowersToOneCell/class.ts
    captured_union_test.go:37: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellclass.ts4270929907/001/main.ts:13:7: stage 0 can't lower a union of differently held members variable a function value captures yet
=== CONT  TestCapturedUnionLowersToOneCell/by_value.a
=== NAME  TestCapturedUnionLowersToOneCell/released.a
    captured_union_test.go:37: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellreleased.a3476599601/001/main.a:17:4: stage 0 can't lower a union of differently held members variable a function value captures yet
=== CONT  TestCapturedUnionLowersToOneCell/by_value.ts
=== NAME  TestCapturedUnionLowersToOneCell/released.ts
    captured_union_test.go:37: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellreleased.ts1751637487/001/main.ts:17:4: stage 0 can't lower a union of differently held members variable a function value captures yet
=== CONT  TestCapturedUnionLowersToOneCell/object.a
=== NAME  TestCapturedUnionLowersToOneCell/by_value.a
    captured_union_test.go:37: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellby_value.a1674039208/001/main.a:20:4: stage 0 can't lower a union of differently held members variable a function value captures yet
=== CONT  TestCapturedUnionLowersToOneCell/object.ts
=== NAME  TestCapturedUnionLowersToOneCell/nested.ts
    captured_union_test.go:37: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellnested.ts3304963122/001/main.ts:8:5: stage 0 can't lower a union of differently held members variable a function value captures yet
=== CONT  TestCapturedUnionLowersToOneCell/class.a
=== NAME  TestCapturedUnionLowersToOneCell/cycle.a
    captured_union_test.go:37: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellcycle.a2898997350/001/main.a:15:7: stage 0 can't lower a union of differently held members variable a function value captures yet
=== CONT  TestCapturedUnionLowersToOneCell/alternating.a
    captured_union_test.go:37: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellalternating.a3778674745/001/main.a:17:4: stage 0 can't lower a union of differently held members variable a function value captures yet
=== CONT  TestCapturedUnionLowersToOneCell/cycle.ts
=== NAME  TestCapturedUnionLowersToOneCell/class.a
    captured_union_test.go:37: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellclass.a2067780351/001/main.a:13:7: stage 0 can't lower a union of differently held members variable a function value captures yet
=== CONT  TestCapturedUnionLowersToOneCell/graph.a
=== NAME  TestCapturedUnionLowersToOneCell/by_value.ts
    captured_union_test.go:37: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellby_value.ts447987668/001/main.ts:20:4: stage 0 can't lower a union of differently held members variable a function value captures yet
=== CONT  TestCapturedUnionLowersToOneCell/graph.ts
=== NAME  TestCapturedUnionLowersToOneCell/graph.a
    captured_union_test.go:37: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellgraph.a1974754399/001/main.a:19:5: stage 0 can't lower a union of differently held members variable a function value captures yet
=== CONT  TestCapturedUnionLowersToOneCell/nested.a
=== NAME  TestCapturedUnionLowersToOneCell/object.ts
    captured_union_test.go:37: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellobject.ts222104710/001/main.ts:7:7: stage 0 can't lower a union of differently held members variable a function value captures yet
=== CONT  TestCapturedUnionLowersToOneCell/alternating.ts
=== NAME  TestCapturedUnionLowersToOneCell/cycle.ts
    captured_union_test.go:37: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellcycle.ts2530527128/001/main.ts:15:7: stage 0 can't lower a union of differently held members variable a function value captures yet
=== NAME  TestCapturedUnionLowersToOneCell/object.a
    captured_union_test.go:37: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellobject.a497294124/001/main.a:7:7: stage 0 can't lower a union of differently held members variable a function value captures yet
=== NAME  TestCapturedUnionLowersToOneCell/graph.ts
    captured_union_test.go:37: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellgraph.ts596596665/001/main.ts:19:5: stage 0 can't lower a union of differently held members variable a function value captures yet
=== NAME  TestCapturedUnionLowersToOneCell/nested.a
    captured_union_test.go:37: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellnested.a1215474141/001/main.a:8:5: stage 0 can't lower a union of differently held members variable a function value captures yet
=== NAME  TestCapturedUnionLowersToOneCell/alternating.ts
    captured_union_test.go:37: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellalternating.ts2570205261/001/main.ts:17:4: stage 0 can't lower a union of differently held members variable a function value captures yet
--- FAIL: TestCapturedUnionLowersToOneCell (0.00s)
    --- FAIL: TestCapturedUnionLowersToOneCell/class.ts (0.37s)
    --- FAIL: TestCapturedUnionLowersToOneCell/released.a (0.39s)
    --- FAIL: TestCapturedUnionLowersToOneCell/released.ts (0.23s)
    --- FAIL: TestCapturedUnionLowersToOneCell/by_value.a (0.24s)
    --- FAIL: TestCapturedUnionLowersToOneCell/nested.ts (0.83s)
    --- FAIL: TestCapturedUnionLowersToOneCell/cycle.a (0.84s)
    --- FAIL: TestCapturedUnionLowersToOneCell/alternating.a (0.06s)
    --- FAIL: TestCapturedUnionLowersToOneCell/class.a (0.08s)
    --- FAIL: TestCapturedUnionLowersToOneCell/by_value.ts (0.99s)
    --- FAIL: TestCapturedUnionLowersToOneCell/graph.a (0.51s)
    --- FAIL: TestCapturedUnionLowersToOneCell/object.ts (0.84s)
    --- FAIL: TestCapturedUnionLowersToOneCell/cycle.ts (0.72s)
    --- FAIL: TestCapturedUnionLowersToOneCell/object.a (1.12s)
    --- FAIL: TestCapturedUnionLowersToOneCell/graph.ts (0.29s)
    --- FAIL: TestCapturedUnionLowersToOneCell/nested.a (0.32s)
    --- FAIL: TestCapturedUnionLowersToOneCell/alternating.ts (0.27s)
FAIL
FAIL	github.com/system-inc/adamic/internal/lower	1.778s
FAIL
```


### M1-oracle.log

```text
=== RUN   TestCapturedUnionAgrees
=== PAUSE TestCapturedUnionAgrees
=== CONT  TestCapturedUnionAgrees
=== RUN   TestCapturedUnionAgrees/class.ts
=== PAUSE TestCapturedUnionAgrees/class.ts
=== RUN   TestCapturedUnionAgrees/class.a
=== PAUSE TestCapturedUnionAgrees/class.a
=== RUN   TestCapturedUnionAgrees/object.ts
=== PAUSE TestCapturedUnionAgrees/object.ts
=== RUN   TestCapturedUnionAgrees/object.a
=== PAUSE TestCapturedUnionAgrees/object.a
=== RUN   TestCapturedUnionAgrees/by_value.ts
=== PAUSE TestCapturedUnionAgrees/by_value.ts
=== RUN   TestCapturedUnionAgrees/by_value.a
=== PAUSE TestCapturedUnionAgrees/by_value.a
=== RUN   TestCapturedUnionAgrees/released.ts
=== PAUSE TestCapturedUnionAgrees/released.ts
=== RUN   TestCapturedUnionAgrees/released.a
=== PAUSE TestCapturedUnionAgrees/released.a
=== RUN   TestCapturedUnionAgrees/nested.ts
=== PAUSE TestCapturedUnionAgrees/nested.ts
=== RUN   TestCapturedUnionAgrees/nested.a
=== PAUSE TestCapturedUnionAgrees/nested.a
=== RUN   TestCapturedUnionAgrees/alternating.ts
=== PAUSE TestCapturedUnionAgrees/alternating.ts
=== RUN   TestCapturedUnionAgrees/alternating.a
=== PAUSE TestCapturedUnionAgrees/alternating.a
=== RUN   TestCapturedUnionAgrees/graph.ts
=== PAUSE TestCapturedUnionAgrees/graph.ts
=== RUN   TestCapturedUnionAgrees/graph.a
=== PAUSE TestCapturedUnionAgrees/graph.a
=== RUN   TestCapturedUnionAgrees/cycle.ts
=== PAUSE TestCapturedUnionAgrees/cycle.ts
=== RUN   TestCapturedUnionAgrees/cycle.a
=== PAUSE TestCapturedUnionAgrees/cycle.a
=== CONT  TestCapturedUnionAgrees/class.ts
=== CONT  TestCapturedUnionAgrees/nested.ts
=== CONT  TestCapturedUnionAgrees/by_value.ts
=== CONT  TestCapturedUnionAgrees/graph.ts
=== CONT  TestCapturedUnionAgrees/released.ts
=== NAME  TestCapturedUnionAgrees/by_value.ts
    captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesby_value.ts496171518/001/by_value.ts:20:4: stage 0 can't lower a union of differently held members variable a function value captures yet
=== CONT  TestCapturedUnionAgrees/by_value.a
=== NAME  TestCapturedUnionAgrees/nested.ts
    captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesnested.ts3612537006/001/nested.ts:8:5: stage 0 can't lower a union of differently held members variable a function value captures yet
=== CONT  TestCapturedUnionAgrees/released.a
=== NAME  TestCapturedUnionAgrees/released.ts
    captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesreleased.ts3562262458/001/released.ts:17:4: stage 0 can't lower a union of differently held members variable a function value captures yet
=== CONT  TestCapturedUnionAgrees/alternating.a
=== NAME  TestCapturedUnionAgrees/graph.ts
    captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesgraph.ts2505154369/001/graph.ts:19:5: stage 0 can't lower a union of differently held members variable a function value captures yet
=== CONT  TestCapturedUnionAgrees/object.ts
=== NAME  TestCapturedUnionAgrees/class.ts
    captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesclass.ts2093953919/001/class.ts:13:7: stage 0 can't lower a union of differently held members variable a function value captures yet
=== CONT  TestCapturedUnionAgrees/cycle.ts
=== NAME  TestCapturedUnionAgrees/object.ts
    captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesobject.ts2148970366/001/object.ts:7:7: stage 0 can't lower a union of differently held members variable a function value captures yet
=== CONT  TestCapturedUnionAgrees/object.a
=== NAME  TestCapturedUnionAgrees/by_value.a
    captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesby_value.a1770380039/001/by_value.a:20:4: stage 0 can't lower a union of differently held members variable a function value captures yet
=== CONT  TestCapturedUnionAgrees/alternating.ts
=== NAME  TestCapturedUnionAgrees/released.a
    captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesreleased.a2049496842/001/released.a:17:4: stage 0 can't lower a union of differently held members variable a function value captures yet
=== CONT  TestCapturedUnionAgrees/cycle.a
=== NAME  TestCapturedUnionAgrees/alternating.a
    captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesalternating.a4119439711/001/alternating.a:17:4: stage 0 can't lower a union of differently held members variable a function value captures yet
=== CONT  TestCapturedUnionAgrees/graph.a
=== NAME  TestCapturedUnionAgrees/cycle.ts
    captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreescycle.ts2353261059/001/cycle.ts:15:7: stage 0 can't lower a union of differently held members variable a function value captures yet
=== CONT  TestCapturedUnionAgrees/nested.a
=== NAME  TestCapturedUnionAgrees/object.a
    captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesobject.a2652515123/001/object.a:7:7: stage 0 can't lower a union of differently held members variable a function value captures yet
=== CONT  TestCapturedUnionAgrees/class.a
=== NAME  TestCapturedUnionAgrees/alternating.ts
    captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesalternating.ts2740834732/001/alternating.ts:17:4: stage 0 can't lower a union of differently held members variable a function value captures yet
=== NAME  TestCapturedUnionAgrees/graph.a
    captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesgraph.a1602122324/001/graph.a:19:5: stage 0 can't lower a union of differently held members variable a function value captures yet
=== NAME  TestCapturedUnionAgrees/cycle.a
    captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreescycle.a3363162970/001/cycle.a:15:7: stage 0 can't lower a union of differently held members variable a function value captures yet
=== NAME  TestCapturedUnionAgrees/nested.a
    captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesnested.a215508565/001/nested.a:8:5: stage 0 can't lower a union of differently held members variable a function value captures yet
=== NAME  TestCapturedUnionAgrees/class.a
    captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesclass.a3835253819/001/class.a:13:7: stage 0 can't lower a union of differently held members variable a function value captures yet
--- FAIL: TestCapturedUnionAgrees (0.00s)
    --- FAIL: TestCapturedUnionAgrees/by_value.ts (0.21s)
    --- FAIL: TestCapturedUnionAgrees/nested.ts (0.21s)
    --- FAIL: TestCapturedUnionAgrees/released.ts (0.22s)
    --- FAIL: TestCapturedUnionAgrees/graph.ts (0.23s)
    --- FAIL: TestCapturedUnionAgrees/class.ts (0.23s)
    --- FAIL: TestCapturedUnionAgrees/object.ts (0.14s)
    --- FAIL: TestCapturedUnionAgrees/by_value.a (0.18s)
    --- FAIL: TestCapturedUnionAgrees/released.a (0.18s)
    --- FAIL: TestCapturedUnionAgrees/alternating.a (0.19s)
    --- FAIL: TestCapturedUnionAgrees/cycle.ts (0.18s)
    --- FAIL: TestCapturedUnionAgrees/object.a (0.08s)
    --- FAIL: TestCapturedUnionAgrees/alternating.ts (0.11s)
    --- FAIL: TestCapturedUnionAgrees/graph.a (0.12s)
    --- FAIL: TestCapturedUnionAgrees/cycle.a (0.13s)
    --- FAIL: TestCapturedUnionAgrees/nested.a (0.12s)
    --- FAIL: TestCapturedUnionAgrees/class.a (0.09s)
FAIL
FAIL	github.com/system-inc/adamic/internal/oracle	0.562s
FAIL
```


### M2-lower.log

```text
=== RUN   TestCapturedUnionLowersToOneCell
=== PAUSE TestCapturedUnionLowersToOneCell
=== RUN   TestCapturedMaybeBooleanStillStops
=== PAUSE TestCapturedMaybeBooleanStillStops
=== CONT  TestCapturedUnionLowersToOneCell
=== RUN   TestCapturedUnionLowersToOneCell/class.ts
=== CONT  TestCapturedMaybeBooleanStillStops
=== PAUSE TestCapturedUnionLowersToOneCell/class.ts
=== RUN   TestCapturedUnionLowersToOneCell/class.a
=== PAUSE TestCapturedUnionLowersToOneCell/class.a
=== RUN   TestCapturedUnionLowersToOneCell/object.ts
=== PAUSE TestCapturedUnionLowersToOneCell/object.ts
=== RUN   TestCapturedUnionLowersToOneCell/object.a
=== PAUSE TestCapturedUnionLowersToOneCell/object.a
=== RUN   TestCapturedUnionLowersToOneCell/by_value.ts
=== PAUSE TestCapturedUnionLowersToOneCell/by_value.ts
=== RUN   TestCapturedUnionLowersToOneCell/by_value.a
=== PAUSE TestCapturedUnionLowersToOneCell/by_value.a
=== RUN   TestCapturedUnionLowersToOneCell/released.ts
=== PAUSE TestCapturedUnionLowersToOneCell/released.ts
=== RUN   TestCapturedUnionLowersToOneCell/released.a
=== PAUSE TestCapturedUnionLowersToOneCell/released.a
=== RUN   TestCapturedUnionLowersToOneCell/nested.ts
=== PAUSE TestCapturedUnionLowersToOneCell/nested.ts
=== RUN   TestCapturedUnionLowersToOneCell/nested.a
=== PAUSE TestCapturedUnionLowersToOneCell/nested.a
=== RUN   TestCapturedUnionLowersToOneCell/alternating.ts
=== PAUSE TestCapturedUnionLowersToOneCell/alternating.ts
=== RUN   TestCapturedUnionLowersToOneCell/alternating.a
=== PAUSE TestCapturedUnionLowersToOneCell/alternating.a
=== RUN   TestCapturedUnionLowersToOneCell/graph.ts
=== PAUSE TestCapturedUnionLowersToOneCell/graph.ts
=== RUN   TestCapturedUnionLowersToOneCell/graph.a
=== PAUSE TestCapturedUnionLowersToOneCell/graph.a
=== RUN   TestCapturedUnionLowersToOneCell/cycle.ts
=== PAUSE TestCapturedUnionLowersToOneCell/cycle.ts
=== RUN   TestCapturedUnionLowersToOneCell/cycle.a
=== PAUSE TestCapturedUnionLowersToOneCell/cycle.a
=== CONT  TestCapturedUnionLowersToOneCell/class.ts
=== CONT  TestCapturedUnionLowersToOneCell/nested.ts
=== CONT  TestCapturedUnionLowersToOneCell/graph.ts
=== CONT  TestCapturedUnionLowersToOneCell/by_value.ts
--- PASS: TestCapturedMaybeBooleanStillStops (0.16s)
=== CONT  TestCapturedUnionLowersToOneCell/object.a
=== NAME  TestCapturedUnionLowersToOneCell/class.ts
    captured_union_test.go:37: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellclass.ts2147513667/001/main.ts:16:7: stage 0 can't lower a BinaryExpression with a value and a string yet
=== NAME  TestCapturedUnionLowersToOneCell/by_value.ts
    captured_union_test.go:37: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellby_value.ts876155215/001/main.ts:30:35: stage 0 can't lower a BinaryExpression with a value and a string yet
=== CONT  TestCapturedUnionLowersToOneCell/object.ts
=== CONT  TestCapturedUnionLowersToOneCell/class.a
=== NAME  TestCapturedUnionLowersToOneCell/graph.ts
    captured_union_test.go:37: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellgraph.ts1315285737/001/main.ts:33:8: stage 0 can't lower a BinaryExpression with a value and a string yet
=== CONT  TestCapturedUnionLowersToOneCell/cycle.ts
=== NAME  TestCapturedUnionLowersToOneCell/nested.ts
    captured_union_test.go:37: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellnested.ts3373742230/001/main.ts:25:22: stage 0 can't lower a template interpolating an object, an array, a map, a function or undefined yet
=== CONT  TestCapturedUnionLowersToOneCell/cycle.a
=== NAME  TestCapturedUnionLowersToOneCell/object.a
    captured_union_test.go:37: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellobject.a1240869268/001/main.a:10:7: stage 0 can't lower a BinaryExpression with a value and a string yet
=== NAME  TestCapturedUnionLowersToOneCell/object.ts
    captured_union_test.go:37: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellobject.ts2793319395/001/main.ts:10:7: stage 0 can't lower a BinaryExpression with a value and a string yet
=== CONT  TestCapturedUnionLowersToOneCell/graph.a
=== CONT  TestCapturedUnionLowersToOneCell/released.ts
=== NAME  TestCapturedUnionLowersToOneCell/cycle.ts
    captured_union_test.go:37: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellcycle.ts3792324500/001/main.ts:18:7: stage 0 can't lower a BinaryExpression with a value and a string yet
=== CONT  TestCapturedUnionLowersToOneCell/released.a
=== NAME  TestCapturedUnionLowersToOneCell/class.a
    captured_union_test.go:37: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellclass.a474699678/001/main.a:16:7: stage 0 can't lower a BinaryExpression with a value and a string yet
=== CONT  TestCapturedUnionLowersToOneCell/by_value.a
=== NAME  TestCapturedUnionLowersToOneCell/cycle.a
    captured_union_test.go:37: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellcycle.a4147608424/001/main.a:18:7: stage 0 can't lower a BinaryExpression with a value and a string yet
=== CONT  TestCapturedUnionLowersToOneCell/alternating.ts
=== NAME  TestCapturedUnionLowersToOneCell/released.ts
    captured_union_test.go:37: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellreleased.ts3930721773/001/main.ts:39:7: stage 0 can't lower a BinaryExpression with a value and a string yet
=== CONT  TestCapturedUnionLowersToOneCell/alternating.a
=== NAME  TestCapturedUnionLowersToOneCell/released.a
    captured_union_test.go:37: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellreleased.a1595996122/001/main.a:39:7: stage 0 can't lower a BinaryExpression with a value and a string yet
=== CONT  TestCapturedUnionLowersToOneCell/nested.a
=== NAME  TestCapturedUnionLowersToOneCell/by_value.a
    captured_union_test.go:37: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellby_value.a3765384823/001/main.a:30:35: stage 0 can't lower a BinaryExpression with a value and a string yet
=== NAME  TestCapturedUnionLowersToOneCell/graph.a
    captured_union_test.go:37: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellgraph.a1709038358/001/main.a:33:8: stage 0 can't lower a BinaryExpression with a value and a string yet
=== NAME  TestCapturedUnionLowersToOneCell/alternating.ts
    captured_union_test.go:37: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellalternating.ts2835279116/001/main.ts:28:7: stage 0 can't lower a BinaryExpression with a value and a string yet
=== NAME  TestCapturedUnionLowersToOneCell/alternating.a
    captured_union_test.go:37: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellalternating.a3897242562/001/main.a:28:7: stage 0 can't lower a BinaryExpression with a value and a string yet
=== NAME  TestCapturedUnionLowersToOneCell/nested.a
    captured_union_test.go:37: /tmp/adamic-gate/TestCapturedUnionLowersToOneCellnested.a443927910/001/main.a:25:22: stage 0 can't lower a template interpolating an object, an array, a map, a function or undefined yet
--- FAIL: TestCapturedUnionLowersToOneCell (0.00s)
    --- FAIL: TestCapturedUnionLowersToOneCell/by_value.ts (0.16s)
    --- FAIL: TestCapturedUnionLowersToOneCell/class.ts (0.18s)
    --- FAIL: TestCapturedUnionLowersToOneCell/graph.ts (0.18s)
    --- FAIL: TestCapturedUnionLowersToOneCell/nested.ts (0.21s)
    --- FAIL: TestCapturedUnionLowersToOneCell/object.a (0.15s)
    --- FAIL: TestCapturedUnionLowersToOneCell/object.ts (0.15s)
    --- FAIL: TestCapturedUnionLowersToOneCell/cycle.ts (0.13s)
    --- FAIL: TestCapturedUnionLowersToOneCell/class.a (0.13s)
    --- FAIL: TestCapturedUnionLowersToOneCell/cycle.a (0.17s)
    --- FAIL: TestCapturedUnionLowersToOneCell/released.ts (0.08s)
    --- FAIL: TestCapturedUnionLowersToOneCell/released.a (0.13s)
    --- FAIL: TestCapturedUnionLowersToOneCell/by_value.a (0.14s)
    --- FAIL: TestCapturedUnionLowersToOneCell/graph.a (0.14s)
    --- FAIL: TestCapturedUnionLowersToOneCell/alternating.ts (0.11s)
    --- FAIL: TestCapturedUnionLowersToOneCell/alternating.a (0.11s)
    --- FAIL: TestCapturedUnionLowersToOneCell/nested.a (0.07s)
FAIL
FAIL	github.com/system-inc/adamic/internal/lower	0.554s
FAIL
```


### M2-oracle.log

```text
=== RUN   TestCapturedUnionAgrees
=== PAUSE TestCapturedUnionAgrees
=== CONT  TestCapturedUnionAgrees
=== RUN   TestCapturedUnionAgrees/class.ts
=== PAUSE TestCapturedUnionAgrees/class.ts
=== RUN   TestCapturedUnionAgrees/class.a
=== PAUSE TestCapturedUnionAgrees/class.a
=== RUN   TestCapturedUnionAgrees/object.ts
=== PAUSE TestCapturedUnionAgrees/object.ts
=== RUN   TestCapturedUnionAgrees/object.a
=== PAUSE TestCapturedUnionAgrees/object.a
=== RUN   TestCapturedUnionAgrees/by_value.ts
=== PAUSE TestCapturedUnionAgrees/by_value.ts
=== RUN   TestCapturedUnionAgrees/by_value.a
=== PAUSE TestCapturedUnionAgrees/by_value.a
=== RUN   TestCapturedUnionAgrees/released.ts
=== PAUSE TestCapturedUnionAgrees/released.ts
=== RUN   TestCapturedUnionAgrees/released.a
=== PAUSE TestCapturedUnionAgrees/released.a
=== RUN   TestCapturedUnionAgrees/nested.ts
=== PAUSE TestCapturedUnionAgrees/nested.ts
=== RUN   TestCapturedUnionAgrees/nested.a
=== PAUSE TestCapturedUnionAgrees/nested.a
=== RUN   TestCapturedUnionAgrees/alternating.ts
=== PAUSE TestCapturedUnionAgrees/alternating.ts
=== RUN   TestCapturedUnionAgrees/alternating.a
=== PAUSE TestCapturedUnionAgrees/alternating.a
=== RUN   TestCapturedUnionAgrees/graph.ts
=== PAUSE TestCapturedUnionAgrees/graph.ts
=== RUN   TestCapturedUnionAgrees/graph.a
=== PAUSE TestCapturedUnionAgrees/graph.a
=== RUN   TestCapturedUnionAgrees/cycle.ts
=== PAUSE TestCapturedUnionAgrees/cycle.ts
=== RUN   TestCapturedUnionAgrees/cycle.a
=== PAUSE TestCapturedUnionAgrees/cycle.a
=== CONT  TestCapturedUnionAgrees/class.ts
=== CONT  TestCapturedUnionAgrees/nested.ts
=== CONT  TestCapturedUnionAgrees/by_value.ts
=== CONT  TestCapturedUnionAgrees/graph.ts
=== CONT  TestCapturedUnionAgrees/released.ts
=== NAME  TestCapturedUnionAgrees/class.ts
    captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesclass.ts581344860/001/class.ts:16:7: stage 0 can't lower a BinaryExpression with a value and a string yet
=== CONT  TestCapturedUnionAgrees/object.a
=== NAME  TestCapturedUnionAgrees/nested.ts
    captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesnested.ts2262350609/001/nested.ts:25:22: stage 0 can't lower a template interpolating an object, an array, a map, a function or undefined yet
=== CONT  TestCapturedUnionAgrees/alternating.a
=== NAME  TestCapturedUnionAgrees/by_value.ts
    captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesby_value.ts3955907946/001/by_value.ts:30:35: stage 0 can't lower a BinaryExpression with a value and a string yet
=== CONT  TestCapturedUnionAgrees/object.ts
=== NAME  TestCapturedUnionAgrees/released.ts
    captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesreleased.ts31105505/001/released.ts:39:7: stage 0 can't lower a BinaryExpression with a value and a string yet
=== CONT  TestCapturedUnionAgrees/alternating.ts
=== NAME  TestCapturedUnionAgrees/graph.ts
    captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesgraph.ts2544236627/001/graph.ts:33:8: stage 0 can't lower a BinaryExpression with a value and a string yet
=== CONT  TestCapturedUnionAgrees/class.a
=== NAME  TestCapturedUnionAgrees/object.ts
    captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesobject.ts2854214631/001/object.ts:10:7: stage 0 can't lower a BinaryExpression with a value and a string yet
=== CONT  TestCapturedUnionAgrees/nested.a
=== NAME  TestCapturedUnionAgrees/object.a
    captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesobject.a1864731276/001/object.a:10:7: stage 0 can't lower a BinaryExpression with a value and a string yet
=== CONT  TestCapturedUnionAgrees/released.a
=== NAME  TestCapturedUnionAgrees/class.a
    captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesclass.a2187697731/001/class.a:16:7: stage 0 can't lower a BinaryExpression with a value and a string yet
=== CONT  TestCapturedUnionAgrees/cycle.ts
=== NAME  TestCapturedUnionAgrees/alternating.a
    captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesalternating.a2537336030/001/alternating.a:28:7: stage 0 can't lower a BinaryExpression with a value and a string yet
=== CONT  TestCapturedUnionAgrees/cycle.a
=== NAME  TestCapturedUnionAgrees/alternating.ts
    captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesalternating.ts4265678876/001/alternating.ts:28:7: stage 0 can't lower a BinaryExpression with a value and a string yet
=== CONT  TestCapturedUnionAgrees/graph.a
=== NAME  TestCapturedUnionAgrees/cycle.a
    captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreescycle.a29200094/001/cycle.a:18:7: stage 0 can't lower a BinaryExpression with a value and a string yet
=== CONT  TestCapturedUnionAgrees/by_value.a
=== NAME  TestCapturedUnionAgrees/released.a
    captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesreleased.a1461874531/001/released.a:39:7: stage 0 can't lower a BinaryExpression with a value and a string yet
=== NAME  TestCapturedUnionAgrees/nested.a
    captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesnested.a2925901490/001/nested.a:25:22: stage 0 can't lower a template interpolating an object, an array, a map, a function or undefined yet
=== NAME  TestCapturedUnionAgrees/graph.a
    captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesgraph.a1774914312/001/graph.a:33:8: stage 0 can't lower a BinaryExpression with a value and a string yet
=== NAME  TestCapturedUnionAgrees/cycle.ts
    captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreescycle.ts496804564/001/cycle.ts:18:7: stage 0 can't lower a BinaryExpression with a value and a string yet
=== NAME  TestCapturedUnionAgrees/by_value.a
    captured_union_test.go:40: /tmp/adamic-gate/TestCapturedUnionAgreesby_value.a2804373078/001/by_value.a:30:35: stage 0 can't lower a BinaryExpression with a value and a string yet
--- FAIL: TestCapturedUnionAgrees (0.00s)
    --- FAIL: TestCapturedUnionAgrees/class.ts (0.30s)
    --- FAIL: TestCapturedUnionAgrees/nested.ts (0.32s)
    --- FAIL: TestCapturedUnionAgrees/by_value.ts (0.33s)
    --- FAIL: TestCapturedUnionAgrees/released.ts (0.35s)
    --- FAIL: TestCapturedUnionAgrees/graph.ts (0.37s)
    --- FAIL: TestCapturedUnionAgrees/object.ts (0.17s)
    --- FAIL: TestCapturedUnionAgrees/object.a (0.23s)
    --- FAIL: TestCapturedUnionAgrees/class.a (0.18s)
    --- FAIL: TestCapturedUnionAgrees/alternating.a (0.23s)
    --- FAIL: TestCapturedUnionAgrees/alternating.ts (0.25s)
    --- FAIL: TestCapturedUnionAgrees/cycle.a (0.11s)
    --- FAIL: TestCapturedUnionAgrees/released.a (0.17s)
    --- FAIL: TestCapturedUnionAgrees/nested.a (0.19s)
    --- FAIL: TestCapturedUnionAgrees/graph.a (0.11s)
    --- FAIL: TestCapturedUnionAgrees/cycle.ts (0.16s)
    --- FAIL: TestCapturedUnionAgrees/by_value.a (0.08s)
FAIL
FAIL	github.com/system-inc/adamic/internal/oracle	0.780s
FAIL
```


### M3-lower.log

```text
=== RUN   TestCapturedUnionLowersToOneCell
=== PAUSE TestCapturedUnionLowersToOneCell
=== RUN   TestCapturedMaybeBooleanStillStops
=== PAUSE TestCapturedMaybeBooleanStillStops
=== CONT  TestCapturedUnionLowersToOneCell
=== CONT  TestCapturedMaybeBooleanStillStops
=== RUN   TestCapturedUnionLowersToOneCell/class.ts
=== PAUSE TestCapturedUnionLowersToOneCell/class.ts
=== RUN   TestCapturedUnionLowersToOneCell/class.a
=== PAUSE TestCapturedUnionLowersToOneCell/class.a
=== RUN   TestCapturedUnionLowersToOneCell/object.ts
=== PAUSE TestCapturedUnionLowersToOneCell/object.ts
=== RUN   TestCapturedUnionLowersToOneCell/object.a
=== PAUSE TestCapturedUnionLowersToOneCell/object.a
=== RUN   TestCapturedUnionLowersToOneCell/by_value.ts
=== PAUSE TestCapturedUnionLowersToOneCell/by_value.ts
=== RUN   TestCapturedUnionLowersToOneCell/by_value.a
=== PAUSE TestCapturedUnionLowersToOneCell/by_value.a
=== RUN   TestCapturedUnionLowersToOneCell/released.ts
=== PAUSE TestCapturedUnionLowersToOneCell/released.ts
=== RUN   TestCapturedUnionLowersToOneCell/released.a
=== PAUSE TestCapturedUnionLowersToOneCell/released.a
=== RUN   TestCapturedUnionLowersToOneCell/nested.ts
=== PAUSE TestCapturedUnionLowersToOneCell/nested.ts
=== RUN   TestCapturedUnionLowersToOneCell/nested.a
=== PAUSE TestCapturedUnionLowersToOneCell/nested.a
=== RUN   TestCapturedUnionLowersToOneCell/alternating.ts
=== PAUSE TestCapturedUnionLowersToOneCell/alternating.ts
=== RUN   TestCapturedUnionLowersToOneCell/alternating.a
=== PAUSE TestCapturedUnionLowersToOneCell/alternating.a
=== RUN   TestCapturedUnionLowersToOneCell/graph.ts
=== PAUSE TestCapturedUnionLowersToOneCell/graph.ts
=== RUN   TestCapturedUnionLowersToOneCell/graph.a
=== PAUSE TestCapturedUnionLowersToOneCell/graph.a
=== RUN   TestCapturedUnionLowersToOneCell/cycle.ts
=== PAUSE TestCapturedUnionLowersToOneCell/cycle.ts
=== RUN   TestCapturedUnionLowersToOneCell/cycle.a
=== PAUSE TestCapturedUnionLowersToOneCell/cycle.a
=== CONT  TestCapturedUnionLowersToOneCell/class.ts
=== CONT  TestCapturedUnionLowersToOneCell/nested.ts
=== CONT  TestCapturedUnionLowersToOneCell/graph.ts
=== CONT  TestCapturedUnionLowersToOneCell/released.a
=== NAME  TestCapturedUnionLowersToOneCell/nested.ts
    captured_union_test.go:48: extended: graph cell true
=== CONT  TestCapturedUnionLowersToOneCell/released.ts
=== NAME  TestCapturedUnionLowersToOneCell/released.a
    captured_union_test.go:48: active: graph cell false
=== CONT  TestCapturedUnionLowersToOneCell/by_value.a
--- PASS: TestCapturedMaybeBooleanStillStops (0.97s)
=== CONT  TestCapturedUnionLowersToOneCell/by_value.ts
=== NAME  TestCapturedUnionLowersToOneCell/class.ts
    captured_union_test.go:48: active: graph cell true
=== CONT  TestCapturedUnionLowersToOneCell/object.a
=== NAME  TestCapturedUnionLowersToOneCell/graph.ts
    captured_union_test.go:48: active: graph cell true
=== CONT  TestCapturedUnionLowersToOneCell/object.ts
=== NAME  TestCapturedUnionLowersToOneCell/object.a
    captured_union_test.go:48: active: graph cell true
=== CONT  TestCapturedUnionLowersToOneCell/class.a
=== NAME  TestCapturedUnionLowersToOneCell/released.ts
    captured_union_test.go:48: active: graph cell false
=== CONT  TestCapturedUnionLowersToOneCell/cycle.ts
=== NAME  TestCapturedUnionLowersToOneCell/by_value.a
    captured_union_test.go:48: active: graph cell false
=== CONT  TestCapturedUnionLowersToOneCell/cycle.a
=== NAME  TestCapturedUnionLowersToOneCell/class.a
    captured_union_test.go:48: active: graph cell true
=== CONT  TestCapturedUnionLowersToOneCell/alternating.ts
=== NAME  TestCapturedUnionLowersToOneCell/object.ts
    captured_union_test.go:48: active: graph cell true
=== CONT  TestCapturedUnionLowersToOneCell/alternating.a
=== NAME  TestCapturedUnionLowersToOneCell/cycle.ts
    captured_union_test.go:48: active: graph cell true
    captured_union_test.go:59: active: 1 closures carry its cell, want every one that reads or writes it
=== CONT  TestCapturedUnionLowersToOneCell/graph.a
=== NAME  TestCapturedUnionLowersToOneCell/cycle.a
    captured_union_test.go:48: active: graph cell true
    captured_union_test.go:59: active: 1 closures carry its cell, want every one that reads or writes it
=== CONT  TestCapturedUnionLowersToOneCell/nested.a
=== NAME  TestCapturedUnionLowersToOneCell/graph.a
    captured_union_test.go:48: active: graph cell true
=== NAME  TestCapturedUnionLowersToOneCell/alternating.ts
    captured_union_test.go:48: active: graph cell false
=== NAME  TestCapturedUnionLowersToOneCell/nested.a
    captured_union_test.go:48: extended: graph cell true
=== NAME  TestCapturedUnionLowersToOneCell/by_value.ts
    captured_union_test.go:48: active: graph cell false
=== NAME  TestCapturedUnionLowersToOneCell/alternating.a
    captured_union_test.go:48: active: graph cell false
--- FAIL: TestCapturedUnionLowersToOneCell (0.00s)
    --- PASS: TestCapturedUnionLowersToOneCell/nested.ts (0.83s)
    --- PASS: TestCapturedUnionLowersToOneCell/released.a (0.96s)
    --- PASS: TestCapturedUnionLowersToOneCell/class.ts (0.98s)
    --- PASS: TestCapturedUnionLowersToOneCell/graph.ts (1.02s)
    --- PASS: TestCapturedUnionLowersToOneCell/object.a (0.39s)
    --- PASS: TestCapturedUnionLowersToOneCell/released.ts (0.56s)
    --- PASS: TestCapturedUnionLowersToOneCell/by_value.a (0.47s)
    --- PASS: TestCapturedUnionLowersToOneCell/class.a (0.54s)
    --- PASS: TestCapturedUnionLowersToOneCell/object.ts (1.02s)
    --- FAIL: TestCapturedUnionLowersToOneCell/cycle.ts (0.65s)
    --- FAIL: TestCapturedUnionLowersToOneCell/cycle.a (0.70s)
    --- PASS: TestCapturedUnionLowersToOneCell/graph.a (0.13s)
    --- PASS: TestCapturedUnionLowersToOneCell/alternating.ts (0.27s)
    --- PASS: TestCapturedUnionLowersToOneCell/nested.a (0.07s)
    --- PASS: TestCapturedUnionLowersToOneCell/by_value.ts (1.31s)
    --- PASS: TestCapturedUnionLowersToOneCell/alternating.a (0.27s)
FAIL
FAIL	github.com/system-inc/adamic/internal/lower	2.412s
FAIL
```


### M3-oracle.log

```text
=== RUN   TestCapturedUnionAgrees
=== PAUSE TestCapturedUnionAgrees
=== CONT  TestCapturedUnionAgrees
=== RUN   TestCapturedUnionAgrees/class.ts
=== PAUSE TestCapturedUnionAgrees/class.ts
=== RUN   TestCapturedUnionAgrees/class.a
=== PAUSE TestCapturedUnionAgrees/class.a
=== RUN   TestCapturedUnionAgrees/object.ts
=== PAUSE TestCapturedUnionAgrees/object.ts
=== RUN   TestCapturedUnionAgrees/object.a
=== PAUSE TestCapturedUnionAgrees/object.a
=== RUN   TestCapturedUnionAgrees/by_value.ts
=== PAUSE TestCapturedUnionAgrees/by_value.ts
=== RUN   TestCapturedUnionAgrees/by_value.a
=== PAUSE TestCapturedUnionAgrees/by_value.a
=== RUN   TestCapturedUnionAgrees/released.ts
=== PAUSE TestCapturedUnionAgrees/released.ts
=== RUN   TestCapturedUnionAgrees/released.a
=== PAUSE TestCapturedUnionAgrees/released.a
=== RUN   TestCapturedUnionAgrees/nested.ts
=== PAUSE TestCapturedUnionAgrees/nested.ts
=== RUN   TestCapturedUnionAgrees/nested.a
=== PAUSE TestCapturedUnionAgrees/nested.a
=== RUN   TestCapturedUnionAgrees/alternating.ts
=== PAUSE TestCapturedUnionAgrees/alternating.ts
=== RUN   TestCapturedUnionAgrees/alternating.a
=== PAUSE TestCapturedUnionAgrees/alternating.a
=== RUN   TestCapturedUnionAgrees/graph.ts
=== PAUSE TestCapturedUnionAgrees/graph.ts
=== RUN   TestCapturedUnionAgrees/graph.a
=== PAUSE TestCapturedUnionAgrees/graph.a
=== RUN   TestCapturedUnionAgrees/cycle.ts
=== PAUSE TestCapturedUnionAgrees/cycle.ts
=== RUN   TestCapturedUnionAgrees/cycle.a
=== PAUSE TestCapturedUnionAgrees/cycle.a
=== CONT  TestCapturedUnionAgrees/class.ts
=== CONT  TestCapturedUnionAgrees/nested.ts
=== CONT  TestCapturedUnionAgrees/by_value.ts
=== CONT  TestCapturedUnionAgrees/object.ts
=== CONT  TestCapturedUnionAgrees/class.a
=== RUN   TestCapturedUnionAgrees/by_value.ts/native
=== RUN   TestCapturedUnionAgrees/class.a/native
=== RUN   TestCapturedUnionAgrees/nested.ts/native
=== RUN   TestCapturedUnionAgrees/nested.ts/sanitized
=== RUN   TestCapturedUnionAgrees/by_value.ts/sanitized
=== RUN   TestCapturedUnionAgrees/class.a/sanitized
=== RUN   TestCapturedUnionAgrees/nested.ts/counted
=== RUN   TestCapturedUnionAgrees/class.ts/native
=== NAME  TestCapturedUnionAgrees/by_value.ts/sanitized
    captured_union_test.go:94: leaks: LeakSanitizer ended the run, so its output wasn't compared: exit 1
        
        =================================================================
        ==10811==ERROR: LeakSanitizer: detected memory leaks
        
        Direct leak of 72 byte(s) in 1 object(s) allocated from:
            #0 0x565110e053a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x565110e71d2f in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x565110e71d2f in adamic_allocate /adamic-runtime/heap.c:267:9
        
        Direct leak of 66 byte(s) in 1 object(s) allocated from:
            #0 0x565110e053a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x565110e71d2f in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x565110e71d2f in adamic_allocate /adamic-runtime/heap.c:267:9
            #3 0x565110e4902d in adamic_function_by_value_ts_compare_86144a167155d3c4b9fb1cc42659a1f0 /adamic-build/main.c:136:2
            #4 0x565110e4902d in adamic_initialize_by_value_ts_bcb45c73d87e7954302beb5e57028536 /adamic-build/main.c:367:2
            #5 0x565110e48a6d in adamic_initialize_by_value_ts_a2576a49d678491017dd8307a246d519 /adamic-build/main.c:377:2
        
        Direct leak of 66 byte(s) in 1 object(s) allocated from:
            #0 0x565110e053a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x565110e71d2f in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x565110e71d2f in adamic_allocate /adamic-runtime/heap.c:267:9
            #3 0x565110e48def in adamic_function_by_value_ts_compare_86144a167155d3c4b9fb1cc42659a1f0 /adamic-build/main.c:91:6
            #4 0x565110e48def in adamic_initialize_by_value_ts_bcb45c73d87e7954302beb5e57028536 /adamic-build/main.c:367:2
            #5 0x565110e48a6d in adamic_initialize_by_value_ts_a2576a49d678491017dd8307a246d519 /adamic-build/main.c:377:2
        
        SUMMARY: AddressSanitizer: 204 byte(s) leaked in 3 allocation(s).
=== RUN   TestCapturedUnionAgrees/by_value.ts/counted
=== RUN   TestCapturedUnionAgrees/object.ts/native
=== NAME  TestCapturedUnionAgrees/nested.ts/counted
    captured_union_test.go:116: counts: allocations 97, frees 97, in regions 0
=== RUN   TestCapturedUnionAgrees/nested.ts/javascript
=== RUN   TestCapturedUnionAgrees/class.a/counted
=== NAME  TestCapturedUnionAgrees/by_value.ts/counted
    captured_union_test.go:110: counts: heap values leaked: 3 (allocations 35, frees 32, in regions 0)
    captured_union_test.go:116: counts: allocations 35, frees 32, in regions 0
=== RUN   TestCapturedUnionAgrees/by_value.ts/javascript
=== RUN   TestCapturedUnionAgrees/class.ts/sanitized
=== RUN   TestCapturedUnionAgrees/object.ts/sanitized
=== NAME  TestCapturedUnionAgrees/class.a/counted
    captured_union_test.go:116: counts: allocations 21, frees 21, in regions 0
=== RUN   TestCapturedUnionAgrees/class.a/javascript
=== NAME  TestCapturedUnionAgrees/nested.ts/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 1.01
=== RUN   TestCapturedUnionAgrees/class.ts/counted
=== NAME  TestCapturedUnionAgrees/by_value.ts/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 1.21
=== NAME  TestCapturedUnionAgrees/class.a/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 1.15
=== RUN   TestCapturedUnionAgrees/object.ts/counted
=== CONT  TestCapturedUnionAgrees/cycle.a
=== CONT  TestCapturedUnionAgrees/cycle.ts
=== CONT  TestCapturedUnionAgrees/graph.a
=== RUN   TestCapturedUnionAgrees/cycle.ts/native
=== RUN   TestCapturedUnionAgrees/cycle.a/native
=== RUN   TestCapturedUnionAgrees/graph.a/native
=== NAME  TestCapturedUnionAgrees/class.ts/counted
    captured_union_test.go:116: counts: allocations 21, frees 21, in regions 0
=== RUN   TestCapturedUnionAgrees/class.ts/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 0.01
=== NAME  TestCapturedUnionAgrees/object.ts/counted
    captured_union_test.go:116: counts: allocations 21, frees 21, in regions 0
=== RUN   TestCapturedUnionAgrees/object.ts/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 0.02
=== CONT  TestCapturedUnionAgrees/graph.ts
=== CONT  TestCapturedUnionAgrees/alternating.a
=== RUN   TestCapturedUnionAgrees/cycle.ts/sanitized
=== RUN   TestCapturedUnionAgrees/cycle.a/sanitized
=== RUN   TestCapturedUnionAgrees/alternating.a/native
=== RUN   TestCapturedUnionAgrees/graph.ts/native
=== RUN   TestCapturedUnionAgrees/cycle.ts/counted
=== RUN   TestCapturedUnionAgrees/alternating.a/sanitized
=== RUN   TestCapturedUnionAgrees/cycle.a/counted
=== NAME  TestCapturedUnionAgrees/cycle.ts/counted
    captured_union_test.go:116: counts: allocations 32, frees 32, in regions 0
=== RUN   TestCapturedUnionAgrees/cycle.ts/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 0.01
=== CONT  TestCapturedUnionAgrees/alternating.ts
=== NAME  TestCapturedUnionAgrees/cycle.a/counted
    captured_union_test.go:116: counts: allocations 32, frees 32, in regions 0
=== RUN   TestCapturedUnionAgrees/cycle.a/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 0.01
=== NAME  TestCapturedUnionAgrees/alternating.a/sanitized
    captured_union_test.go:94: leaks: LeakSanitizer ended the run, so its output wasn't compared: exit 1
        
        =================================================================
        ==11211==ERROR: LeakSanitizer: detected memory leaks
        
        Direct leak of 252 byte(s) in 3 object(s) allocated from:
            #0 0x55b5ec72f3a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x55b5ec79bdcf in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x55b5ec79bdcf in adamic_allocate /adamic-runtime/heap.c:267:9
            #3 0x55b5ec772e1d in adamic_function_alternating_a_alternate_01e607951fdf653584ac75f3f81a1146 /adamic-build/main.c:75:6
            #4 0x55b5ec772ac0 in adamic_initialize_alternating_a_8aabf12e8e980896d6070e8d5697d4ba /adamic-build/main.c:404:2
        
        Direct leak of 252 byte(s) in 3 object(s) allocated from:
            #0 0x55b5ec72f3a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x55b5ec79bdcf in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x55b5ec79bdcf in adamic_allocate /adamic-runtime/heap.c:267:9
            #3 0x55b5ec772e1d in adamic_function_alternating_a_alternate_01e607951fdf653584ac75f3f81a1146 /adamic-build/main.c:75:6
            #4 0x55b5ec772a84 in adamic_initialize_alternating_a_8aabf12e8e980896d6070e8d5697d4ba /adamic-build/main.c:400:2
        
        Indirect leak of 834 byte(s) in 12 object(s) allocated from:
            #0 0x55b5ec72f3a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x55b5ec79bdcf in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x55b5ec79bdcf in adamic_allocate /adamic-runtime/heap.c:267:9
        
        Indirect leak of 96 byte(s) in 6 object(s) allocated from:
            #0 0x55b5ec72f3a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x55b5ec77777f in adamic_array_make /adamic-runtime/array.c:22:21
        
        SUMMARY: AddressSanitizer: 1434 byte(s) leaked in 24 allocation(s).
=== RUN   TestCapturedUnionAgrees/alternating.a/counted
=== CONT  TestCapturedUnionAgrees/nested.a
=== RUN   TestCapturedUnionAgrees/alternating.ts/native
=== NAME  TestCapturedUnionAgrees/alternating.a/counted
    captured_union_test.go:110: counts: heap values leaked: 18 (allocations 54, frees 36, in regions 0)
    captured_union_test.go:116: counts: allocations 54, frees 36, in regions 0
=== RUN   TestCapturedUnionAgrees/alternating.a/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 0.02
=== RUN   TestCapturedUnionAgrees/nested.a/native
=== CONT  TestCapturedUnionAgrees/released.a
=== RUN   TestCapturedUnionAgrees/alternating.ts/sanitized
=== RUN   TestCapturedUnionAgrees/released.a/native
=== RUN   TestCapturedUnionAgrees/nested.a/sanitized
=== NAME  TestCapturedUnionAgrees/alternating.ts/sanitized
    captured_union_test.go:94: leaks: LeakSanitizer ended the run, so its output wasn't compared: exit 1
        
        =================================================================
        ==11299==ERROR: LeakSanitizer: detected memory leaks
        
        Direct leak of 252 byte(s) in 3 object(s) allocated from:
            #0 0x561d3468b3a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x561d346f7dcf in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x561d346f7dcf in adamic_allocate /adamic-runtime/heap.c:267:9
            #3 0x561d346cee1d in adamic_function_alternating_ts_alternate_565c9a204ed396dab50febfce9b6094f /adamic-build/main.c:75:6
            #4 0x561d346ceac0 in adamic_initialize_alternating_ts_798a1070fca02ff27cf4b4600fa7ac9f /adamic-build/main.c:404:2
        
        Direct leak of 252 byte(s) in 3 object(s) allocated from:
            #0 0x561d3468b3a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x561d346f7dcf in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x561d346f7dcf in adamic_allocate /adamic-runtime/heap.c:267:9
            #3 0x561d346cee1d in adamic_function_alternating_ts_alternate_565c9a204ed396dab50febfce9b6094f /adamic-build/main.c:75:6
            #4 0x561d346cea84 in adamic_initialize_alternating_ts_798a1070fca02ff27cf4b4600fa7ac9f /adamic-build/main.c:400:2
        
        Indirect leak of 834 byte(s) in 12 object(s) allocated from:
            #0 0x561d3468b3a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x561d346f7dcf in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x561d346f7dcf in adamic_allocate /adamic-runtime/heap.c:267:9
        
        Indirect leak of 96 byte(s) in 6 object(s) allocated from:
            #0 0x561d3468b3a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x561d346d377f in adamic_array_make /adamic-runtime/array.c:22:21
        
        SUMMARY: AddressSanitizer: 1434 byte(s) leaked in 24 allocation(s).
=== RUN   TestCapturedUnionAgrees/alternating.ts/counted
=== RUN   TestCapturedUnionAgrees/released.a/sanitized
=== RUN   TestCapturedUnionAgrees/nested.a/counted
=== NAME  TestCapturedUnionAgrees/alternating.ts/counted
    captured_union_test.go:110: counts: heap values leaked: 18 (allocations 54, frees 36, in regions 0)
    captured_union_test.go:116: counts: allocations 54, frees 36, in regions 0
=== RUN   TestCapturedUnionAgrees/alternating.ts/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 0.01
=== CONT  TestCapturedUnionAgrees/released.ts
=== NAME  TestCapturedUnionAgrees/nested.a/counted
    captured_union_test.go:116: counts: allocations 97, frees 97, in regions 0
=== RUN   TestCapturedUnionAgrees/nested.a/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 0.01
=== NAME  TestCapturedUnionAgrees/released.a/sanitized
    captured_union_test.go:94: leaks: LeakSanitizer ended the run, so its output wasn't compared: exit 1
        
        =================================================================
        ==11353==ERROR: LeakSanitizer: detected memory leaks
        
        Direct leak of 252 byte(s) in 3 object(s) allocated from:
            #0 0x563351b7c3a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x563351be93ff in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x563351be93ff in adamic_allocate /adamic-runtime/heap.c:267:9
            #3 0x563351bc08cc in adamic_function_released_a_cycle_write_anonymous_1_f66d096b0481ccf1bd7e006bc2a986dd /adamic-build/main.c:197:85
            #4 0x563351bbfd4d in adamic_function_released_a_cycle_bed10a1973e0de82b2d589920ae3ed09 /adamic-build/main.c:80:6
            #5 0x563351bbfd4d in adamic_initialize_released_a_0d43d726d956cc76c866f06cfbcdbdac /adamic-build/main.c:466:94
            #6 0x563351bbfa6d in adamic_initialize_released_a_3493427bc7441308aba02ae6046df03f /adamic-build/main.c:484:2
        
        Direct leak of 252 byte(s) in 3 object(s) allocated from:
            #0 0x563351b7c3a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x563351be93ff in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x563351be93ff in adamic_allocate /adamic-runtime/heap.c:267:9
            #3 0x563351bc06a6 in adamic_function_released_a_cycle_write_anonymous_1_f66d096b0481ccf1bd7e006bc2a986dd /adamic-build/main.c:178:84
            #4 0x563351bbfd4d in adamic_function_released_a_cycle_bed10a1973e0de82b2d589920ae3ed09 /adamic-build/main.c:80:6
            #5 0x563351bbfd4d in adamic_initialize_released_a_0d43d726d956cc76c866f06cfbcdbdac /adamic-build/main.c:466:94
            #6 0x563351bbfa6d in adamic_initialize_released_a_3493427bc7441308aba02ae6046df03f /adamic-build/main.c:484:2
        
        Direct leak of 252 byte(s) in 3 object(s) allocated from:
            #0 0x563351b7c3a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x563351be93ff in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x563351be93ff in adamic_allocate /adamic-runtime/heap.c:267:9
            #3 0x563351bc0431 in adamic_function_released_a_cycle_write_anonymous_1_f66d096b0481ccf1bd7e006bc2a986dd /adamic-build/main.c:145:80
            #4 0x563351bbfd4d in adamic_function_released_a_cycle_bed10a1973e0de82b2d589920ae3ed09 /adamic-build/main.c:80:6
            #5 0x563351bbfd4d in adamic_initialize_released_a_0d43d726d956cc76c866f06cfbcdbdac /adamic-build/main.c:466:94
            #6 0x563351bbfa6d in adamic_initialize_released_a_3493427bc7441308aba02ae6046df03f /adamic-build/main.c:484:2
        
        Direct leak of 216 byte(s) in 3 object(s) allocated from:
            #0 0x563351b7c3a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x563351be93ff in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x563351be93ff in adamic_allocate /adamic-runtime/heap.c:267:9
        
        Indirect leak of 1218 byte(s) in 18 object(s) allocated from:
            #0 0x563351b7c3a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x563351be93ff in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x563351be93ff in adamic_allocate /adamic-runtime/heap.c:267:9
        
        Indirect leak of 216 byte(s) in 9 object(s) allocated from:
            #0 0x563351b7c3a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x563351bc4daf in adamic_array_make /adamic-runtime/array.c:22:21
        
        SUMMARY: AddressSanitizer: 2406 byte(s) leaked in 39 allocation(s).
=== RUN   TestCapturedUnionAgrees/released.a/counted
=== RUN   TestCapturedUnionAgrees/released.ts/native
=== CONT  TestCapturedUnionAgrees/by_value.a
=== RUN   TestCapturedUnionAgrees/graph.ts/sanitized
=== RUN   TestCapturedUnionAgrees/graph.a/sanitized
=== RUN   TestCapturedUnionAgrees/by_value.a/native
=== NAME  TestCapturedUnionAgrees/released.a/counted
    captured_union_test.go:110: counts: heap values leaked: 30 (allocations 96, frees 66, in regions 0)
    captured_union_test.go:116: counts: allocations 96, frees 66, in regions 0
=== RUN   TestCapturedUnionAgrees/released.a/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 0.01
=== RUN   TestCapturedUnionAgrees/released.ts/sanitized
=== CONT  TestCapturedUnionAgrees/object.a
=== RUN   TestCapturedUnionAgrees/by_value.a/sanitized
=== RUN   TestCapturedUnionAgrees/object.a/native
=== NAME  TestCapturedUnionAgrees/released.ts/sanitized
    captured_union_test.go:94: leaks: LeakSanitizer ended the run, so its output wasn't compared: exit 1
        
        =================================================================
        ==11450==ERROR: LeakSanitizer: detected memory leaks
        
        Direct leak of 252 byte(s) in 3 object(s) allocated from:
            #0 0x560a89dce3a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x560a89e3b3ff in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x560a89e3b3ff in adamic_allocate /adamic-runtime/heap.c:267:9
            #3 0x560a89e128cc in adamic_function_released_ts_cycle_write_anonymous_1_ed447f88c000910f55ed34387dbd1f8f /adamic-build/main.c:197:85
            #4 0x560a89e11d4d in adamic_function_released_ts_cycle_d9ea6f0e042b91093191a3c465bc1d8b /adamic-build/main.c:80:6
            #5 0x560a89e11d4d in adamic_initialize_released_ts_bde04b1d7cf3d0450b6f2cb4dc674d34 /adamic-build/main.c:466:95
            #6 0x560a89e11a6d in adamic_initialize_released_ts_2e08b23fbedfe14b22b962cdf28991f1 /adamic-build/main.c:484:2
        
        Direct leak of 252 byte(s) in 3 object(s) allocated from:
            #0 0x560a89dce3a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x560a89e3b3ff in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x560a89e3b3ff in adamic_allocate /adamic-runtime/heap.c:267:9
            #3 0x560a89e126a6 in adamic_function_released_ts_cycle_write_anonymous_1_ed447f88c000910f55ed34387dbd1f8f /adamic-build/main.c:178:84
            #4 0x560a89e11d4d in adamic_function_released_ts_cycle_d9ea6f0e042b91093191a3c465bc1d8b /adamic-build/main.c:80:6
            #5 0x560a89e11d4d in adamic_initialize_released_ts_bde04b1d7cf3d0450b6f2cb4dc674d34 /adamic-build/main.c:466:95
            #6 0x560a89e11a6d in adamic_initialize_released_ts_2e08b23fbedfe14b22b962cdf28991f1 /adamic-build/main.c:484:2
        
        Direct leak of 252 byte(s) in 3 object(s) allocated from:
            #0 0x560a89dce3a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x560a89e3b3ff in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x560a89e3b3ff in adamic_allocate /adamic-runtime/heap.c:267:9
            #3 0x560a89e12431 in adamic_function_released_ts_cycle_write_anonymous_1_ed447f88c000910f55ed34387dbd1f8f /adamic-build/main.c:145:80
            #4 0x560a89e11d4d in adamic_function_released_ts_cycle_d9ea6f0e042b91093191a3c465bc1d8b /adamic-build/main.c:80:6
            #5 0x560a89e11d4d in adamic_initialize_released_ts_bde04b1d7cf3d0450b6f2cb4dc674d34 /adamic-build/main.c:466:95
            #6 0x560a89e11a6d in adamic_initialize_released_ts_2e08b23fbedfe14b22b962cdf28991f1 /adamic-build/main.c:484:2
        
        Direct leak of 216 byte(s) in 3 object(s) allocated from:
            #0 0x560a89dce3a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x560a89e3b3ff in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x560a89e3b3ff in adamic_allocate /adamic-runtime/heap.c:267:9
        
        Indirect leak of 1218 byte(s) in 18 object(s) allocated from:
            #0 0x560a89dce3a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x560a89e3b3ff in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x560a89e3b3ff in adamic_allocate /adamic-runtime/heap.c:267:9
        
        Indirect leak of 216 byte(s) in 9 object(s) allocated from:
            #0 0x560a89dce3a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x560a89e16daf in adamic_array_make /adamic-runtime/array.c:22:21
        
        SUMMARY: AddressSanitizer: 2406 byte(s) leaked in 39 allocation(s).
=== RUN   TestCapturedUnionAgrees/released.ts/counted
=== NAME  TestCapturedUnionAgrees/by_value.a/sanitized
    captured_union_test.go:94: leaks: LeakSanitizer ended the run, so its output wasn't compared: exit 1
        
        =================================================================
        ==11462==ERROR: LeakSanitizer: detected memory leaks
        
        Direct leak of 72 byte(s) in 1 object(s) allocated from:
            #0 0x561d034393a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x561d034a5d2f in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x561d034a5d2f in adamic_allocate /adamic-runtime/heap.c:267:9
        
        Direct leak of 66 byte(s) in 1 object(s) allocated from:
            #0 0x561d034393a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x561d034a5d2f in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x561d034a5d2f in adamic_allocate /adamic-runtime/heap.c:267:9
            #3 0x561d0347d02d in adamic_function_by_value_a_compare_1954618ce2d9e07b1980a9ac4e4160ae /adamic-build/main.c:136:2
            #4 0x561d0347d02d in adamic_initialize_by_value_a_4c44bd7515c4580a5a4f6d72acde192b /adamic-build/main.c:367:2
            #5 0x561d0347ca6d in adamic_initialize_by_value_a_e96b858367cf1e592bc28bb5116317f3 /adamic-build/main.c:377:2
        
        Direct leak of 66 byte(s) in 1 object(s) allocated from:
            #0 0x561d034393a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x561d034a5d2f in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x561d034a5d2f in adamic_allocate /adamic-runtime/heap.c:267:9
            #3 0x561d0347cdef in adamic_function_by_value_a_compare_1954618ce2d9e07b1980a9ac4e4160ae /adamic-build/main.c:91:6
            #4 0x561d0347cdef in adamic_initialize_by_value_a_4c44bd7515c4580a5a4f6d72acde192b /adamic-build/main.c:367:2
            #5 0x561d0347ca6d in adamic_initialize_by_value_a_e96b858367cf1e592bc28bb5116317f3 /adamic-build/main.c:377:2
        
        SUMMARY: AddressSanitizer: 204 byte(s) leaked in 3 allocation(s).
=== RUN   TestCapturedUnionAgrees/by_value.a/counted
=== RUN   TestCapturedUnionAgrees/object.a/sanitized
=== NAME  TestCapturedUnionAgrees/released.ts/counted
    captured_union_test.go:110: counts: heap values leaked: 30 (allocations 96, frees 66, in regions 0)
    captured_union_test.go:116: counts: allocations 96, frees 66, in regions 0
=== RUN   TestCapturedUnionAgrees/released.ts/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 0.01
=== NAME  TestCapturedUnionAgrees/by_value.a/counted
    captured_union_test.go:110: counts: heap values leaked: 3 (allocations 35, frees 32, in regions 0)
    captured_union_test.go:116: counts: allocations 35, frees 32, in regions 0
=== RUN   TestCapturedUnionAgrees/by_value.a/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 0.01
=== RUN   TestCapturedUnionAgrees/object.a/counted
    captured_union_test.go:116: counts: allocations 21, frees 21, in regions 0
=== RUN   TestCapturedUnionAgrees/object.a/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 0.01
=== RUN   TestCapturedUnionAgrees/graph.a/counted
=== RUN   TestCapturedUnionAgrees/graph.ts/counted
    captured_union_test.go:116: counts: allocations 42, frees 42, in regions 0
=== RUN   TestCapturedUnionAgrees/graph.ts/javascript
=== NAME  TestCapturedUnionAgrees/graph.a/counted
    captured_union_test.go:116: counts: allocations 42, frees 42, in regions 0
=== RUN   TestCapturedUnionAgrees/graph.a/javascript
=== NAME  TestCapturedUnionAgrees/graph.ts/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 0.01
=== NAME  TestCapturedUnionAgrees/graph.a/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 0.01
--- FAIL: TestCapturedUnionAgrees (0.00s)
    --- PASS: TestCapturedUnionAgrees/nested.ts (15.21s)
        --- PASS: TestCapturedUnionAgrees/nested.ts/native (0.32s)
        --- PASS: TestCapturedUnionAgrees/nested.ts/sanitized (1.01s)
        --- PASS: TestCapturedUnionAgrees/nested.ts/counted (0.58s)
        --- PASS: TestCapturedUnionAgrees/nested.ts/javascript (1.95s)
    --- FAIL: TestCapturedUnionAgrees/by_value.ts (15.22s)
        --- PASS: TestCapturedUnionAgrees/by_value.ts/native (0.56s)
        --- FAIL: TestCapturedUnionAgrees/by_value.ts/sanitized (1.28s)
        --- FAIL: TestCapturedUnionAgrees/by_value.ts/counted (0.41s)
        --- PASS: TestCapturedUnionAgrees/by_value.ts/javascript (1.61s)
    --- PASS: TestCapturedUnionAgrees/class.a (15.27s)
        --- PASS: TestCapturedUnionAgrees/class.a/native (0.70s)
        --- PASS: TestCapturedUnionAgrees/class.a/sanitized (1.22s)
        --- PASS: TestCapturedUnionAgrees/class.a/counted (0.74s)
        --- PASS: TestCapturedUnionAgrees/class.a/javascript (1.26s)
    --- PASS: TestCapturedUnionAgrees/class.ts (15.99s)
        --- PASS: TestCapturedUnionAgrees/class.ts/native (0.64s)
        --- PASS: TestCapturedUnionAgrees/class.ts/sanitized (1.13s)
        --- PASS: TestCapturedUnionAgrees/class.ts/counted (0.91s)
        --- PASS: TestCapturedUnionAgrees/class.ts/javascript (0.12s)
    --- PASS: TestCapturedUnionAgrees/object.ts (16.00s)
        --- PASS: TestCapturedUnionAgrees/object.ts/native (0.66s)
        --- PASS: TestCapturedUnionAgrees/object.ts/sanitized (1.35s)
        --- PASS: TestCapturedUnionAgrees/object.ts/counted (0.69s)
        --- PASS: TestCapturedUnionAgrees/object.ts/javascript (0.11s)
    --- PASS: TestCapturedUnionAgrees/cycle.ts (1.78s)
        --- PASS: TestCapturedUnionAgrees/cycle.ts/native (0.33s)
        --- PASS: TestCapturedUnionAgrees/cycle.ts/sanitized (0.54s)
        --- PASS: TestCapturedUnionAgrees/cycle.ts/counted (0.36s)
        --- PASS: TestCapturedUnionAgrees/cycle.ts/javascript (0.08s)
    --- PASS: TestCapturedUnionAgrees/cycle.a (1.96s)
        --- PASS: TestCapturedUnionAgrees/cycle.a/native (0.40s)
        --- PASS: TestCapturedUnionAgrees/cycle.a/sanitized (0.65s)
        --- PASS: TestCapturedUnionAgrees/cycle.a/counted (0.27s)
        --- PASS: TestCapturedUnionAgrees/cycle.a/javascript (0.09s)
    --- FAIL: TestCapturedUnionAgrees/alternating.a (1.53s)
        --- PASS: TestCapturedUnionAgrees/alternating.a/native (0.25s)
        --- FAIL: TestCapturedUnionAgrees/alternating.a/sanitized (0.54s)
        --- FAIL: TestCapturedUnionAgrees/alternating.a/counted (0.28s)
        --- PASS: TestCapturedUnionAgrees/alternating.a/javascript (0.11s)
    --- FAIL: TestCapturedUnionAgrees/alternating.ts (1.45s)
        --- PASS: TestCapturedUnionAgrees/alternating.ts/native (0.33s)
        --- FAIL: TestCapturedUnionAgrees/alternating.ts/sanitized (0.49s)
        --- FAIL: TestCapturedUnionAgrees/alternating.ts/counted (0.29s)
        --- PASS: TestCapturedUnionAgrees/alternating.ts/javascript (0.12s)
    --- PASS: TestCapturedUnionAgrees/nested.a (1.57s)
        --- PASS: TestCapturedUnionAgrees/nested.a/native (0.30s)
        --- PASS: TestCapturedUnionAgrees/nested.a/sanitized (0.52s)
        --- PASS: TestCapturedUnionAgrees/nested.a/counted (0.31s)
        --- PASS: TestCapturedUnionAgrees/nested.a/javascript (0.15s)
    --- FAIL: TestCapturedUnionAgrees/released.a (1.68s)
        --- PASS: TestCapturedUnionAgrees/released.a/native (0.30s)
        --- FAIL: TestCapturedUnionAgrees/released.a/sanitized (0.60s)
        --- FAIL: TestCapturedUnionAgrees/released.a/counted (0.46s)
        --- PASS: TestCapturedUnionAgrees/released.a/javascript (0.10s)
    --- FAIL: TestCapturedUnionAgrees/released.ts (1.64s)
        --- PASS: TestCapturedUnionAgrees/released.ts/native (0.44s)
        --- FAIL: TestCapturedUnionAgrees/released.ts/sanitized (0.57s)
        --- FAIL: TestCapturedUnionAgrees/released.ts/counted (0.30s)
        --- PASS: TestCapturedUnionAgrees/released.ts/javascript (0.07s)
    --- FAIL: TestCapturedUnionAgrees/by_value.a (1.52s)
        --- PASS: TestCapturedUnionAgrees/by_value.a/native (0.32s)
        --- FAIL: TestCapturedUnionAgrees/by_value.a/sanitized (0.51s)
        --- FAIL: TestCapturedUnionAgrees/by_value.a/counted (0.29s)
        --- PASS: TestCapturedUnionAgrees/by_value.a/javascript (0.09s)
    --- PASS: TestCapturedUnionAgrees/object.a (1.51s)
        --- PASS: TestCapturedUnionAgrees/object.a/native (0.42s)
        --- PASS: TestCapturedUnionAgrees/object.a/sanitized (0.46s)
        --- PASS: TestCapturedUnionAgrees/object.a/counted (0.29s)
        --- PASS: TestCapturedUnionAgrees/object.a/javascript (0.07s)
    --- PASS: TestCapturedUnionAgrees/graph.ts (7.66s)
        --- PASS: TestCapturedUnionAgrees/graph.ts/native (2.43s)
        --- PASS: TestCapturedUnionAgrees/graph.ts/sanitized (2.68s)
        --- PASS: TestCapturedUnionAgrees/graph.ts/counted (2.05s)
        --- PASS: TestCapturedUnionAgrees/graph.ts/javascript (0.09s)
    --- PASS: TestCapturedUnionAgrees/graph.a (8.37s)
        --- PASS: TestCapturedUnionAgrees/graph.a/native (3.02s)
        --- PASS: TestCapturedUnionAgrees/graph.a/sanitized (2.62s)
        --- PASS: TestCapturedUnionAgrees/graph.a/counted (2.06s)
        --- PASS: TestCapturedUnionAgrees/graph.a/javascript (0.08s)
FAIL
gate cache: native hits=0 misses=32
gate cache: node hits=0 misses=32
gate cache: probe hits=0 misses=0
FAIL	github.com/system-inc/adamic/internal/oracle	23.709s
FAIL
```


### M3-diagnostic.log

```text
=== RUN   TestCapturedUnionAgrees
=== PAUSE TestCapturedUnionAgrees
=== CONT  TestCapturedUnionAgrees
=== RUN   TestCapturedUnionAgrees/class.ts
=== PAUSE TestCapturedUnionAgrees/class.ts
=== RUN   TestCapturedUnionAgrees/class.a
=== PAUSE TestCapturedUnionAgrees/class.a
=== RUN   TestCapturedUnionAgrees/object.ts
=== PAUSE TestCapturedUnionAgrees/object.ts
=== RUN   TestCapturedUnionAgrees/object.a
=== PAUSE TestCapturedUnionAgrees/object.a
=== RUN   TestCapturedUnionAgrees/by_value.ts
=== PAUSE TestCapturedUnionAgrees/by_value.ts
=== RUN   TestCapturedUnionAgrees/by_value.a
=== PAUSE TestCapturedUnionAgrees/by_value.a
=== RUN   TestCapturedUnionAgrees/released.ts
=== PAUSE TestCapturedUnionAgrees/released.ts
=== RUN   TestCapturedUnionAgrees/released.a
=== PAUSE TestCapturedUnionAgrees/released.a
=== RUN   TestCapturedUnionAgrees/nested.ts
=== PAUSE TestCapturedUnionAgrees/nested.ts
=== RUN   TestCapturedUnionAgrees/nested.a
=== PAUSE TestCapturedUnionAgrees/nested.a
=== RUN   TestCapturedUnionAgrees/alternating.ts
=== PAUSE TestCapturedUnionAgrees/alternating.ts
=== RUN   TestCapturedUnionAgrees/alternating.a
=== PAUSE TestCapturedUnionAgrees/alternating.a
=== RUN   TestCapturedUnionAgrees/graph.ts
=== PAUSE TestCapturedUnionAgrees/graph.ts
=== RUN   TestCapturedUnionAgrees/graph.a
=== PAUSE TestCapturedUnionAgrees/graph.a
=== RUN   TestCapturedUnionAgrees/cycle.ts
=== PAUSE TestCapturedUnionAgrees/cycle.ts
=== RUN   TestCapturedUnionAgrees/cycle.a
=== PAUSE TestCapturedUnionAgrees/cycle.a
=== CONT  TestCapturedUnionAgrees/class.ts
=== CONT  TestCapturedUnionAgrees/by_value.ts
=== CONT  TestCapturedUnionAgrees/nested.ts
=== CONT  TestCapturedUnionAgrees/graph.ts
=== CONT  TestCapturedUnionAgrees/released.ts
=== RUN   TestCapturedUnionAgrees/nested.ts/native
=== RUN   TestCapturedUnionAgrees/released.ts/native
=== RUN   TestCapturedUnionAgrees/by_value.ts/native
=== RUN   TestCapturedUnionAgrees/class.ts/native
=== RUN   TestCapturedUnionAgrees/graph.ts/native
=== RUN   TestCapturedUnionAgrees/released.ts/sanitized
=== RUN   TestCapturedUnionAgrees/nested.ts/sanitized
=== RUN   TestCapturedUnionAgrees/by_value.ts/sanitized
=== RUN   TestCapturedUnionAgrees/class.ts/sanitized
=== NAME  TestCapturedUnionAgrees/nested.ts/sanitized
    captured_union_test.go:93: full got: exit 0, stdout "no base\none base ./base.json\n3 bases ./a.json+./b.json+./c.json\n2 bases ./a.json+./b.json\none base ./base.json\nno base\nno base\n", stderr ""; want stdout "no base\none base ./base.json\n3 bases ./a.json+./b.json+./c.json\n2 bases ./a.json+./b.json\none base ./base.json\nno base\nno base\n"
=== RUN   TestCapturedUnionAgrees/nested.ts/counted
=== NAME  TestCapturedUnionAgrees/released.ts/sanitized
    captured_union_test.go:93: full got: exit 1, stdout "first 0 of 3, stopping, undefined, second 0 of 3, third 0 of 3, stopping, last 0 of 3\nfirst 1 of 3, stopping, undefined, second 1 of 3, third 1 of 3, stopping, last 1 of 3\nfirst 2 of 3, stopping, undefined, second 2 of 3, third 2 of 3, stopping, last 2 of 3\n", stderr "\n=================================================================\n==212061==ERROR: LeakSanitizer: detected memory leaks\n\nDirect leak of 252 byte(s) in 3 object(s) allocated from:\n    #0 0x559046ac03a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3\n    #1 0x559046b2d3ff in allocate_storage /adamic-runtime/heap.c:253:10\n    #2 0x559046b2d3ff in adamic_allocate /adamic-runtime/heap.c:267:9\n    #3 0x559046b048cc in adamic_function_released_ts_cycle_write_anonymous_1_ed447f88c000910f55ed34387dbd1f8f /adamic-build/main.c:197:85\n    #4 0x559046b03d4d in adamic_function_released_ts_cycle_d9ea6f0e042b91093191a3c465bc1d8b /adamic-build/main.c:80:6\n    #5 0x559046b03d4d in adamic_initialize_released_ts_bde04b1d7cf3d0450b6f2cb4dc674d34 /adamic-build/main.c:466:95\n    #6 0x559046b03a6d in adamic_initialize_released_ts_2e08b23fbedfe14b22b962cdf28991f1 /adamic-build/main.c:484:2\n\nDirect leak of 252 byte(s) in 3 object(s) allocated from:\n    #0 0x559046ac03a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3\n    #1 0x559046b2d3ff in allocate_storage /adamic-runtime/heap.c:253:10\n    #2 0x559046b2d3ff in adamic_allocate /adamic-runtime/heap.c:267:9\n    #3 0x559046b046a6 in adamic_function_released_ts_cycle_write_anonymous_1_ed447f88c000910f55ed34387dbd1f8f /adamic-build/main.c:178:84\n    #4 0x559046b03d4d in adamic_function_released_ts_cycle_d9ea6f0e042b91093191a3c465bc1d8b /adamic-build/main.c:80:6\n    #5 0x559046b03d4d in adamic_initialize_released_ts_bde04b1d7cf3d0450b6f2cb4dc674d34 /adamic-build/main.c:466:95\n    #6 0x559046b03a6d in adamic_initialize_released_ts_2e08b23fbedfe14b22b962cdf28991f1 /adamic-build/main.c:484:2\n\nDirect leak of 252 byte(s) in 3 object(s) allocated from:\n    #0 0x559046ac03a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3\n    #1 0x559046b2d3ff in allocate_storage /adamic-runtime/heap.c:253:10\n    #2 0x559046b2d3ff in adamic_allocate /adamic-runtime/heap.c:267:9\n    #3 0x559046b04431 in adamic_function_released_ts_cycle_write_anonymous_1_ed447f88c000910f55ed34387dbd1f8f /adamic-build/main.c:145:80\n    #4 0x559046b03d4d in adamic_function_released_ts_cycle_d9ea6f0e042b91093191a3c465bc1d8b /adamic-build/main.c:80:6\n    #5 0x559046b03d4d in adamic_initialize_released_ts_bde04b1d7cf3d0450b6f2cb4dc674d34 /adamic-build/main.c:466:95\n    #6 0x559046b03a6d in adamic_initialize_released_ts_2e08b23fbedfe14b22b962cdf28991f1 /adamic-build/main.c:484:2\n\nDirect leak of 216 byte(s) in 3 object(s) allocated from:\n    #0 0x559046ac03a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3\n    #1 0x559046b2d3ff in allocate_storage /adamic-runtime/heap.c:253:10\n    #2 0x559046b2d3ff in adamic_allocate /adamic-runtime/heap.c:267:9\n\nIndirect leak of 1218 byte(s) in 18 object(s) allocated from:\n    #0 0x559046ac03a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3\n    #1 0x559046b2d3ff in allocate_storage /adamic-runtime/heap.c:253:10\n    #2 0x559046b2d3ff in adamic_allocate /adamic-runtime/heap.c:267:9\n\nIndirect leak of 216 byte(s) in 9 object(s) allocated from:\n    #0 0x559046ac03a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3\n    #1 0x559046b08daf in adamic_array_make /adamic-runtime/array.c:22:21\n\nSUMMARY: AddressSanitizer: 2406 byte(s) leaked in 39 allocation(s).\n"; want stdout "first 0 of 3, stopping, undefined, second 0 of 3, third 0 of 3, stopping, last 0 of 3\nfirst 1 of 3, stopping, undefined, second 1 of 3, third 1 of 3, stopping, last 1 of 3\nfirst 2 of 3, stopping, undefined, second 2 of 3, third 2 of 3, stopping, last 2 of 3\n"
    captured_union_test.go:95: leaks: LeakSanitizer ended the run, so its output wasn't compared: exit 1
        
        =================================================================
        ==212061==ERROR: LeakSanitizer: detected memory leaks
        
        Direct leak of 252 byte(s) in 3 object(s) allocated from:
            #0 0x559046ac03a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x559046b2d3ff in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x559046b2d3ff in adamic_allocate /adamic-runtime/heap.c:267:9
            #3 0x559046b048cc in adamic_function_released_ts_cycle_write_anonymous_1_ed447f88c000910f55ed34387dbd1f8f /adamic-build/main.c:197:85
            #4 0x559046b03d4d in adamic_function_released_ts_cycle_d9ea6f0e042b91093191a3c465bc1d8b /adamic-build/main.c:80:6
            #5 0x559046b03d4d in adamic_initialize_released_ts_bde04b1d7cf3d0450b6f2cb4dc674d34 /adamic-build/main.c:466:95
            #6 0x559046b03a6d in adamic_initialize_released_ts_2e08b23fbedfe14b22b962cdf28991f1 /adamic-build/main.c:484:2
        
        Direct leak of 252 byte(s) in 3 object(s) allocated from:
            #0 0x559046ac03a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x559046b2d3ff in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x559046b2d3ff in adamic_allocate /adamic-runtime/heap.c:267:9
            #3 0x559046b046a6 in adamic_function_released_ts_cycle_write_anonymous_1_ed447f88c000910f55ed34387dbd1f8f /adamic-build/main.c:178:84
            #4 0x559046b03d4d in adamic_function_released_ts_cycle_d9ea6f0e042b91093191a3c465bc1d8b /adamic-build/main.c:80:6
            #5 0x559046b03d4d in adamic_initialize_released_ts_bde04b1d7cf3d0450b6f2cb4dc674d34 /adamic-build/main.c:466:95
            #6 0x559046b03a6d in adamic_initialize_released_ts_2e08b23fbedfe14b22b962cdf28991f1 /adamic-build/main.c:484:2
        
        Direct leak of 252 byte(s) in 3 object(s) allocated from:
            #0 0x559046ac03a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x559046b2d3ff in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x559046b2d3ff in adamic_allocate /adamic-runtime/heap.c:267:9
            #3 0x559046b04431 in adamic_function_released_ts_cycle_write_anonymous_1_ed447f88c000910f55ed34387dbd1f8f /adamic-build/main.c:145:80
            #4 0x559046b03d4d in adamic_function_released_ts_cycle_d9ea6f0e042b91093191a3c465bc1d8b /adamic-build/main.c:80:6
            #5 0x559046b03d4d in adamic_initialize_released_ts_bde04b1d7cf3d0450b6f2cb4dc674d34 /adamic-build/main.c:466:95
            #6 0x559046b03a6d in adamic_initialize_released_ts_2e08b23fbedfe14b22b962cdf28991f1 /adamic-build/main.c:484:2
        
        Direct leak of 216 byte(s) in 3 object(s) allocated from:
            #0 0x559046ac03a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x559046b2d3ff in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x559046b2d3ff in adamic_allocate /adamic-runtime/heap.c:267:9
        
        Indirect leak of 1218 byte(s) in 18 object(s) allocated from:
            #0 0x559046ac03a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x559046b2d3ff in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x559046b2d3ff in adamic_allocate /adamic-runtime/heap.c:267:9
        
        Indirect leak of 216 byte(s) in 9 object(s) allocated from:
            #0 0x559046ac03a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x559046b08daf in adamic_array_make /adamic-runtime/array.c:22:21
        
        SUMMARY: AddressSanitizer: 2406 byte(s) leaked in 39 allocation(s).
=== RUN   TestCapturedUnionAgrees/released.ts/counted
=== NAME  TestCapturedUnionAgrees/by_value.ts/sanitized
    captured_union_test.go:93: full got: exit 1, stdout "step 0 false false false false true\nstep 1 true true true true false\nstep 2 false false false false true\nstep 3 true true true true false\nthe cell holds the word\nthe cell holds something else\nthe cell holds the word\n", stderr "\n=================================================================\n==212065==ERROR: LeakSanitizer: detected memory leaks\n\nDirect leak of 72 byte(s) in 1 object(s) allocated from:\n    #0 0x55c9023a93a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3\n    #1 0x55c902415d2f in allocate_storage /adamic-runtime/heap.c:253:10\n    #2 0x55c902415d2f in adamic_allocate /adamic-runtime/heap.c:267:9\n\nDirect leak of 66 byte(s) in 1 object(s) allocated from:\n    #0 0x55c9023a93a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3\n    #1 0x55c902415d2f in allocate_storage /adamic-runtime/heap.c:253:10\n    #2 0x55c902415d2f in adamic_allocate /adamic-runtime/heap.c:267:9\n    #3 0x55c9023ed02d in adamic_function_by_value_ts_compare_86144a167155d3c4b9fb1cc42659a1f0 /adamic-build/main.c:136:2\n    #4 0x55c9023ed02d in adamic_initialize_by_value_ts_bcb45c73d87e7954302beb5e57028536 /adamic-build/main.c:367:2\n    #5 0x55c9023eca6d in adamic_initialize_by_value_ts_a2576a49d678491017dd8307a246d519 /adamic-build/main.c:377:2\n\nDirect leak of 66 byte(s) in 1 object(s) allocated from:\n    #0 0x55c9023a93a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3\n    #1 0x55c902415d2f in allocate_storage /adamic-runtime/heap.c:253:10\n    #2 0x55c902415d2f in adamic_allocate /adamic-runtime/heap.c:267:9\n    #3 0x55c9023ecdef in adamic_function_by_value_ts_compare_86144a167155d3c4b9fb1cc42659a1f0 /adamic-build/main.c:91:6\n    #4 0x55c9023ecdef in adamic_initialize_by_value_ts_bcb45c73d87e7954302beb5e57028536 /adamic-build/main.c:367:2\n    #5 0x55c9023eca6d in adamic_initialize_by_value_ts_a2576a49d678491017dd8307a246d519 /adamic-build/main.c:377:2\n\nSUMMARY: AddressSanitizer: 204 byte(s) leaked in 3 allocation(s).\n"; want stdout "step 0 false false false false true\nstep 1 true true true true false\nstep 2 false false false false true\nstep 3 true true true true false\nthe cell holds the word\nthe cell holds something else\nthe cell holds the word\n"
    captured_union_test.go:95: leaks: LeakSanitizer ended the run, so its output wasn't compared: exit 1
        
        =================================================================
        ==212065==ERROR: LeakSanitizer: detected memory leaks
        
        Direct leak of 72 byte(s) in 1 object(s) allocated from:
            #0 0x55c9023a93a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x55c902415d2f in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x55c902415d2f in adamic_allocate /adamic-runtime/heap.c:267:9
        
        Direct leak of 66 byte(s) in 1 object(s) allocated from:
            #0 0x55c9023a93a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x55c902415d2f in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x55c902415d2f in adamic_allocate /adamic-runtime/heap.c:267:9
            #3 0x55c9023ed02d in adamic_function_by_value_ts_compare_86144a167155d3c4b9fb1cc42659a1f0 /adamic-build/main.c:136:2
            #4 0x55c9023ed02d in adamic_initialize_by_value_ts_bcb45c73d87e7954302beb5e57028536 /adamic-build/main.c:367:2
            #5 0x55c9023eca6d in adamic_initialize_by_value_ts_a2576a49d678491017dd8307a246d519 /adamic-build/main.c:377:2
        
        Direct leak of 66 byte(s) in 1 object(s) allocated from:
            #0 0x55c9023a93a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x55c902415d2f in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x55c902415d2f in adamic_allocate /adamic-runtime/heap.c:267:9
            #3 0x55c9023ecdef in adamic_function_by_value_ts_compare_86144a167155d3c4b9fb1cc42659a1f0 /adamic-build/main.c:91:6
            #4 0x55c9023ecdef in adamic_initialize_by_value_ts_bcb45c73d87e7954302beb5e57028536 /adamic-build/main.c:367:2
            #5 0x55c9023eca6d in adamic_initialize_by_value_ts_a2576a49d678491017dd8307a246d519 /adamic-build/main.c:377:2
        
        SUMMARY: AddressSanitizer: 204 byte(s) leaked in 3 allocation(s).
=== RUN   TestCapturedUnionAgrees/by_value.ts/counted
=== NAME  TestCapturedUnionAgrees/class.ts/sanitized
    captured_union_test.go:93: full got: exit 0, stdout "start undefined false\nstep 0 session first true object\nstep 1 stopping true string\nstep 2 undefined false undefined\nenabled\nenable while profiled\ndisabled profiled during stopping\nnothing to disable\nenable while stopping\nenabled\nafter a write the cell says string\nstopping\n", stderr ""; want stdout "start undefined false\nstep 0 session first true object\nstep 1 stopping true string\nstep 2 undefined false undefined\nenabled\nenable while profiled\ndisabled profiled during stopping\nnothing to disable\nenable while stopping\nenabled\nafter a write the cell says string\nstopping\n"
=== RUN   TestCapturedUnionAgrees/class.ts/counted
=== NAME  TestCapturedUnionAgrees/nested.ts/counted
    captured_union_test.go:110: full got: exit 0, stdout "no base\none base ./base.json\n3 bases ./a.json+./b.json+./c.json\n2 bases ./a.json+./b.json\none base ./base.json\nno base\nno base\n", stderr "adamic: graph region: live 4 bytes 208 reachable 1 bytes 32 unreachable 3 bytes 176 metadata 208\nadamic: graph region: live 7 bytes 360 reachable 1 bytes 32 unreachable 6 bytes 328 metadata 400\nadamic: graph region: live 7 bytes 360 reachable 1 bytes 32 unreachable 6 bytes 328 metadata 400\nadamic: graph region: live 10 bytes 512 reachable 1 bytes 32 unreachable 9 bytes 480 metadata 592\nadamic: graph region: live 10 bytes 512 reachable 1 bytes 32 unreachable 9 bytes 480 metadata 592\nadamic: graph region: live 10 bytes 512 reachable 1 bytes 32 unreachable 9 bytes 480 metadata 592\nadamic: graph region: live 7 bytes 360 reachable 1 bytes 32 unreachable 6 bytes 328 metadata 400\nadamic: graph counts: regions 16 merges 48\nadamic: counts: allocations 97 frees 97 retains 99 releases 151 peak 21 regions 0\n"; want stdout "no base\none base ./base.json\n3 bases ./a.json+./b.json+./c.json\n2 bases ./a.json+./b.json\none base ./base.json\nno base\nno base\n"
    captured_union_test.go:118: counts: allocations 97, frees 97, in regions 0
=== RUN   TestCapturedUnionAgrees/nested.ts/javascript
=== NAME  TestCapturedUnionAgrees/released.ts/counted
    captured_union_test.go:110: full got: exit 0, stdout "first 0 of 3, stopping, undefined, second 0 of 3, third 0 of 3, stopping, last 0 of 3\nfirst 1 of 3, stopping, undefined, second 1 of 3, third 1 of 3, stopping, last 1 of 3\nfirst 2 of 3, stopping, undefined, second 2 of 3, third 2 of 3, stopping, last 2 of 3\n", stderr "adamic: counts: allocations 96 frees 66 retains 198 releases 216 peak 43 regions 0\n"; want stdout "first 0 of 3, stopping, undefined, second 0 of 3, third 0 of 3, stopping, last 0 of 3\nfirst 1 of 3, stopping, undefined, second 1 of 3, third 1 of 3, stopping, last 1 of 3\nfirst 2 of 3, stopping, undefined, second 2 of 3, third 2 of 3, stopping, last 2 of 3\n"
    captured_union_test.go:112: counts: heap values leaked: 30 (allocations 96, frees 66, in regions 0)
    captured_union_test.go:118: counts: allocations 96, frees 66, in regions 0
=== RUN   TestCapturedUnionAgrees/released.ts/javascript
=== NAME  TestCapturedUnionAgrees/by_value.ts/counted
    captured_union_test.go:110: full got: exit 0, stdout "step 0 false false false false true\nstep 1 true true true true false\nstep 2 false false false false true\nstep 3 true true true true false\nthe cell holds the word\nthe cell holds something else\nthe cell holds the word\n", stderr "adamic: counts: allocations 35 frees 32 retains 65 releases 71 peak 10 regions 0\n"; want stdout "step 0 false false false false true\nstep 1 true true true true false\nstep 2 false false false false true\nstep 3 true true true true false\nthe cell holds the word\nthe cell holds something else\nthe cell holds the word\n"
    captured_union_test.go:112: counts: heap values leaked: 3 (allocations 35, frees 32, in regions 0)
    captured_union_test.go:118: counts: allocations 35, frees 32, in regions 0
=== RUN   TestCapturedUnionAgrees/by_value.ts/javascript
=== NAME  TestCapturedUnionAgrees/class.ts/counted
    captured_union_test.go:110: full got: exit 0, stdout "start undefined false\nstep 0 session first true object\nstep 1 stopping true string\nstep 2 undefined false undefined\nenabled\nenable while profiled\ndisabled profiled during stopping\nnothing to disable\nenable while stopping\nenabled\nafter a write the cell says string\nstopping\n", stderr "adamic: graph region: live 7 bytes 432 reachable 1 bytes 32 unreachable 6 bytes 400 metadata 256\nadamic: graph counts: regions 1 merges 6\nadamic: counts: allocations 21 frees 21 retains 66 releases 83 peak 9 regions 0\n"; want stdout "start undefined false\nstep 0 session first true object\nstep 1 stopping true string\nstep 2 undefined false undefined\nenabled\nenable while profiled\ndisabled profiled during stopping\nnothing to disable\nenable while stopping\nenabled\nafter a write the cell says string\nstopping\n"
    captured_union_test.go:118: counts: allocations 21, frees 21, in regions 0
=== RUN   TestCapturedUnionAgrees/class.ts/javascript
=== NAME  TestCapturedUnionAgrees/nested.ts/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 0.76
=== NAME  TestCapturedUnionAgrees/released.ts/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 0.77
=== NAME  TestCapturedUnionAgrees/class.ts/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 0.73
=== CONT  TestCapturedUnionAgrees/released.a
=== CONT  TestCapturedUnionAgrees/cycle.ts
=== NAME  TestCapturedUnionAgrees/by_value.ts/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 0.94
=== CONT  TestCapturedUnionAgrees/cycle.a
=== CONT  TestCapturedUnionAgrees/graph.a
=== RUN   TestCapturedUnionAgrees/released.a/native
=== RUN   TestCapturedUnionAgrees/cycle.ts/native
=== RUN   TestCapturedUnionAgrees/cycle.a/native
=== RUN   TestCapturedUnionAgrees/graph.a/native
=== RUN   TestCapturedUnionAgrees/graph.ts/sanitized
=== RUN   TestCapturedUnionAgrees/cycle.ts/sanitized
=== RUN   TestCapturedUnionAgrees/released.a/sanitized
=== RUN   TestCapturedUnionAgrees/cycle.a/sanitized
=== RUN   TestCapturedUnionAgrees/graph.a/sanitized
=== NAME  TestCapturedUnionAgrees/cycle.ts/sanitized
    captured_union_test.go:93: full got: exit 0, stdout "undefined, session first 0, stopping, session second 0, undefined, session last 0\nundefined, session first 1, stopping, session second 1, undefined, session last 1\n", stderr ""; want stdout "undefined, session first 0, stopping, session second 0, undefined, session last 0\nundefined, session first 1, stopping, session second 1, undefined, session last 1\n"
=== RUN   TestCapturedUnionAgrees/cycle.ts/counted
=== NAME  TestCapturedUnionAgrees/released.a/sanitized
    captured_union_test.go:93: full got: exit 1, stdout "first 0 of 3, stopping, undefined, second 0 of 3, third 0 of 3, stopping, last 0 of 3\nfirst 1 of 3, stopping, undefined, second 1 of 3, third 1 of 3, stopping, last 1 of 3\nfirst 2 of 3, stopping, undefined, second 2 of 3, third 2 of 3, stopping, last 2 of 3\n", stderr "\n=================================================================\n==212484==ERROR: LeakSanitizer: detected memory leaks\n\nDirect leak of 252 byte(s) in 3 object(s) allocated from:\n    #0 0x55c3dec1e3a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3\n    #1 0x55c3dec8b3ff in allocate_storage /adamic-runtime/heap.c:253:10\n    #2 0x55c3dec8b3ff in adamic_allocate /adamic-runtime/heap.c:267:9\n    #3 0x55c3dec628cc in adamic_function_released_a_cycle_write_anonymous_1_f66d096b0481ccf1bd7e006bc2a986dd /adamic-build/main.c:197:85\n    #4 0x55c3dec61d4d in adamic_function_released_a_cycle_bed10a1973e0de82b2d589920ae3ed09 /adamic-build/main.c:80:6\n    #5 0x55c3dec61d4d in adamic_initialize_released_a_0d43d726d956cc76c866f06cfbcdbdac /adamic-build/main.c:466:94\n    #6 0x55c3dec61a6d in adamic_initialize_released_a_3493427bc7441308aba02ae6046df03f /adamic-build/main.c:484:2\n\nDirect leak of 252 byte(s) in 3 object(s) allocated from:\n    #0 0x55c3dec1e3a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3\n    #1 0x55c3dec8b3ff in allocate_storage /adamic-runtime/heap.c:253:10\n    #2 0x55c3dec8b3ff in adamic_allocate /adamic-runtime/heap.c:267:9\n    #3 0x55c3dec626a6 in adamic_function_released_a_cycle_write_anonymous_1_f66d096b0481ccf1bd7e006bc2a986dd /adamic-build/main.c:178:84\n    #4 0x55c3dec61d4d in adamic_function_released_a_cycle_bed10a1973e0de82b2d589920ae3ed09 /adamic-build/main.c:80:6\n    #5 0x55c3dec61d4d in adamic_initialize_released_a_0d43d726d956cc76c866f06cfbcdbdac /adamic-build/main.c:466:94\n    #6 0x55c3dec61a6d in adamic_initialize_released_a_3493427bc7441308aba02ae6046df03f /adamic-build/main.c:484:2\n\nDirect leak of 252 byte(s) in 3 object(s) allocated from:\n    #0 0x55c3dec1e3a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3\n    #1 0x55c3dec8b3ff in allocate_storage /adamic-runtime/heap.c:253:10\n    #2 0x55c3dec8b3ff in adamic_allocate /adamic-runtime/heap.c:267:9\n    #3 0x55c3dec62431 in adamic_function_released_a_cycle_write_anonymous_1_f66d096b0481ccf1bd7e006bc2a986dd /adamic-build/main.c:145:80\n    #4 0x55c3dec61d4d in adamic_function_released_a_cycle_bed10a1973e0de82b2d589920ae3ed09 /adamic-build/main.c:80:6\n    #5 0x55c3dec61d4d in adamic_initialize_released_a_0d43d726d956cc76c866f06cfbcdbdac /adamic-build/main.c:466:94\n    #6 0x55c3dec61a6d in adamic_initialize_released_a_3493427bc7441308aba02ae6046df03f /adamic-build/main.c:484:2\n\nDirect leak of 216 byte(s) in 3 object(s) allocated from:\n    #0 0x55c3dec1e3a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3\n    #1 0x55c3dec8b3ff in allocate_storage /adamic-runtime/heap.c:253:10\n    #2 0x55c3dec8b3ff in adamic_allocate /adamic-runtime/heap.c:267:9\n\nIndirect leak of 1218 byte(s) in 18 object(s) allocated from:\n    #0 0x55c3dec1e3a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3\n    #1 0x55c3dec8b3ff in allocate_storage /adamic-runtime/heap.c:253:10\n    #2 0x55c3dec8b3ff in adamic_allocate /adamic-runtime/heap.c:267:9\n\nIndirect leak of 216 byte(s) in 9 object(s) allocated from:\n    #0 0x55c3dec1e3a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3\n    #1 0x55c3dec66daf in adamic_array_make /adamic-runtime/array.c:22:21\n\nSUMMARY: AddressSanitizer: 2406 byte(s) leaked in 39 allocation(s).\n"; want stdout "first 0 of 3, stopping, undefined, second 0 of 3, third 0 of 3, stopping, last 0 of 3\nfirst 1 of 3, stopping, undefined, second 1 of 3, third 1 of 3, stopping, last 1 of 3\nfirst 2 of 3, stopping, undefined, second 2 of 3, third 2 of 3, stopping, last 2 of 3\n"
    captured_union_test.go:95: leaks: LeakSanitizer ended the run, so its output wasn't compared: exit 1
        
        =================================================================
        ==212484==ERROR: LeakSanitizer: detected memory leaks
        
        Direct leak of 252 byte(s) in 3 object(s) allocated from:
            #0 0x55c3dec1e3a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x55c3dec8b3ff in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x55c3dec8b3ff in adamic_allocate /adamic-runtime/heap.c:267:9
            #3 0x55c3dec628cc in adamic_function_released_a_cycle_write_anonymous_1_f66d096b0481ccf1bd7e006bc2a986dd /adamic-build/main.c:197:85
            #4 0x55c3dec61d4d in adamic_function_released_a_cycle_bed10a1973e0de82b2d589920ae3ed09 /adamic-build/main.c:80:6
            #5 0x55c3dec61d4d in adamic_initialize_released_a_0d43d726d956cc76c866f06cfbcdbdac /adamic-build/main.c:466:94
            #6 0x55c3dec61a6d in adamic_initialize_released_a_3493427bc7441308aba02ae6046df03f /adamic-build/main.c:484:2
        
        Direct leak of 252 byte(s) in 3 object(s) allocated from:
            #0 0x55c3dec1e3a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x55c3dec8b3ff in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x55c3dec8b3ff in adamic_allocate /adamic-runtime/heap.c:267:9
            #3 0x55c3dec626a6 in adamic_function_released_a_cycle_write_anonymous_1_f66d096b0481ccf1bd7e006bc2a986dd /adamic-build/main.c:178:84
            #4 0x55c3dec61d4d in adamic_function_released_a_cycle_bed10a1973e0de82b2d589920ae3ed09 /adamic-build/main.c:80:6
            #5 0x55c3dec61d4d in adamic_initialize_released_a_0d43d726d956cc76c866f06cfbcdbdac /adamic-build/main.c:466:94
            #6 0x55c3dec61a6d in adamic_initialize_released_a_3493427bc7441308aba02ae6046df03f /adamic-build/main.c:484:2
        
        Direct leak of 252 byte(s) in 3 object(s) allocated from:
            #0 0x55c3dec1e3a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x55c3dec8b3ff in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x55c3dec8b3ff in adamic_allocate /adamic-runtime/heap.c:267:9
            #3 0x55c3dec62431 in adamic_function_released_a_cycle_write_anonymous_1_f66d096b0481ccf1bd7e006bc2a986dd /adamic-build/main.c:145:80
            #4 0x55c3dec61d4d in adamic_function_released_a_cycle_bed10a1973e0de82b2d589920ae3ed09 /adamic-build/main.c:80:6
            #5 0x55c3dec61d4d in adamic_initialize_released_a_0d43d726d956cc76c866f06cfbcdbdac /adamic-build/main.c:466:94
            #6 0x55c3dec61a6d in adamic_initialize_released_a_3493427bc7441308aba02ae6046df03f /adamic-build/main.c:484:2
        
        Direct leak of 216 byte(s) in 3 object(s) allocated from:
            #0 0x55c3dec1e3a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x55c3dec8b3ff in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x55c3dec8b3ff in adamic_allocate /adamic-runtime/heap.c:267:9
        
        Indirect leak of 1218 byte(s) in 18 object(s) allocated from:
            #0 0x55c3dec1e3a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x55c3dec8b3ff in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x55c3dec8b3ff in adamic_allocate /adamic-runtime/heap.c:267:9
        
        Indirect leak of 216 byte(s) in 9 object(s) allocated from:
            #0 0x55c3dec1e3a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x55c3dec66daf in adamic_array_make /adamic-runtime/array.c:22:21
        
        SUMMARY: AddressSanitizer: 2406 byte(s) leaked in 39 allocation(s).
=== RUN   TestCapturedUnionAgrees/released.a/counted
=== NAME  TestCapturedUnionAgrees/cycle.a/sanitized
    captured_union_test.go:93: full got: exit 0, stdout "undefined, session first 0, stopping, session second 0, undefined, session last 0\nundefined, session first 1, stopping, session second 1, undefined, session last 1\n", stderr ""; want stdout "undefined, session first 0, stopping, session second 0, undefined, session last 0\nundefined, session first 1, stopping, session second 1, undefined, session last 1\n"
=== RUN   TestCapturedUnionAgrees/cycle.a/counted
=== NAME  TestCapturedUnionAgrees/cycle.ts/counted
    captured_union_test.go:110: full got: exit 0, stdout "undefined, session first 0, stopping, session second 0, undefined, session last 0\nundefined, session first 1, stopping, session second 1, undefined, session last 1\n", stderr "adamic: graph region: live 5 bytes 356 reachable 3 bytes 188 unreachable 2 bytes 168 metadata 224\nadamic: graph region: live 5 bytes 356 reachable 3 bytes 188 unreachable 2 bytes 168 metadata 224\nadamic: graph counts: regions 2 merges 8\nadamic: counts: allocations 32 frees 32 retains 82 releases 96 peak 13 regions 0\n"; want stdout "undefined, session first 0, stopping, session second 0, undefined, session last 0\nundefined, session first 1, stopping, session second 1, undefined, session last 1\n"
    captured_union_test.go:118: counts: allocations 32, frees 32, in regions 0
=== RUN   TestCapturedUnionAgrees/cycle.ts/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 0.07
=== NAME  TestCapturedUnionAgrees/released.a/counted
    captured_union_test.go:110: full got: exit 0, stdout "first 0 of 3, stopping, undefined, second 0 of 3, third 0 of 3, stopping, last 0 of 3\nfirst 1 of 3, stopping, undefined, second 1 of 3, third 1 of 3, stopping, last 1 of 3\nfirst 2 of 3, stopping, undefined, second 2 of 3, third 2 of 3, stopping, last 2 of 3\n", stderr "adamic: counts: allocations 96 frees 66 retains 198 releases 216 peak 43 regions 0\n"; want stdout "first 0 of 3, stopping, undefined, second 0 of 3, third 0 of 3, stopping, last 0 of 3\nfirst 1 of 3, stopping, undefined, second 1 of 3, third 1 of 3, stopping, last 1 of 3\nfirst 2 of 3, stopping, undefined, second 2 of 3, third 2 of 3, stopping, last 2 of 3\n"
    captured_union_test.go:112: counts: heap values leaked: 30 (allocations 96, frees 66, in regions 0)
    captured_union_test.go:118: counts: allocations 96, frees 66, in regions 0
=== RUN   TestCapturedUnionAgrees/released.a/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 0.04
=== NAME  TestCapturedUnionAgrees/cycle.a/counted
    captured_union_test.go:110: full got: exit 0, stdout "undefined, session first 0, stopping, session second 0, undefined, session last 0\nundefined, session first 1, stopping, session second 1, undefined, session last 1\n", stderr "adamic: graph region: live 5 bytes 356 reachable 3 bytes 188 unreachable 2 bytes 168 metadata 224\nadamic: graph region: live 5 bytes 356 reachable 3 bytes 188 unreachable 2 bytes 168 metadata 224\nadamic: graph counts: regions 2 merges 8\nadamic: counts: allocations 32 frees 32 retains 82 releases 96 peak 13 regions 0\n"; want stdout "undefined, session first 0, stopping, session second 0, undefined, session last 0\nundefined, session first 1, stopping, session second 1, undefined, session last 1\n"
    captured_union_test.go:118: counts: allocations 32, frees 32, in regions 0
=== RUN   TestCapturedUnionAgrees/cycle.a/javascript
=== CONT  TestCapturedUnionAgrees/object.ts
=== NAME  TestCapturedUnionAgrees/cycle.a/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 0.03
=== CONT  TestCapturedUnionAgrees/object.a
=== CONT  TestCapturedUnionAgrees/alternating.ts
=== RUN   TestCapturedUnionAgrees/object.ts/native
=== RUN   TestCapturedUnionAgrees/object.a/native
=== RUN   TestCapturedUnionAgrees/alternating.ts/native
=== RUN   TestCapturedUnionAgrees/object.a/sanitized
=== RUN   TestCapturedUnionAgrees/alternating.ts/sanitized
=== RUN   TestCapturedUnionAgrees/object.ts/sanitized
=== NAME  TestCapturedUnionAgrees/alternating.ts/sanitized
    captured_union_test.go:93: full got: exit 1, stdout "session 0.0 with 2, stopping, undefined, session 0.3 with 2, stopping, undefined, session 0.6 with 2, stopping\nsession 1.0 with 2, stopping, undefined, session 1.3 with 2, stopping, undefined, session 1.6 with 2, stopping\n", stderr "\n=================================================================\n==212628==ERROR: LeakSanitizer: detected memory leaks\n\nDirect leak of 252 byte(s) in 3 object(s) allocated from:\n    #0 0x564d99a673a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3\n    #1 0x564d99ad3dcf in allocate_storage /adamic-runtime/heap.c:253:10\n    #2 0x564d99ad3dcf in adamic_allocate /adamic-runtime/heap.c:267:9\n    #3 0x564d99aaae1d in adamic_function_alternating_ts_alternate_565c9a204ed396dab50febfce9b6094f /adamic-build/main.c:75:6\n    #4 0x564d99aaaac0 in adamic_initialize_alternating_ts_798a1070fca02ff27cf4b4600fa7ac9f /adamic-build/main.c:404:2\n\nDirect leak of 252 byte(s) in 3 object(s) allocated from:\n    #0 0x564d99a673a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3\n    #1 0x564d99ad3dcf in allocate_storage /adamic-runtime/heap.c:253:10\n    #2 0x564d99ad3dcf in adamic_allocate /adamic-runtime/heap.c:267:9\n    #3 0x564d99aaae1d in adamic_function_alternating_ts_alternate_565c9a204ed396dab50febfce9b6094f /adamic-build/main.c:75:6\n    #4 0x564d99aaaa84 in adamic_initialize_alternating_ts_798a1070fca02ff27cf4b4600fa7ac9f /adamic-build/main.c:400:2\n\nIndirect leak of 834 byte(s) in 12 object(s) allocated from:\n    #0 0x564d99a673a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3\n    #1 0x564d99ad3dcf in allocate_storage /adamic-runtime/heap.c:253:10\n    #2 0x564d99ad3dcf in adamic_allocate /adamic-runtime/heap.c:267:9\n\nIndirect leak of 96 byte(s) in 6 object(s) allocated from:\n    #0 0x564d99a673a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3\n    #1 0x564d99aaf77f in adamic_array_make /adamic-runtime/array.c:22:21\n\nSUMMARY: AddressSanitizer: 1434 byte(s) leaked in 24 allocation(s).\n"; want stdout "session 0.0 with 2, stopping, undefined, session 0.3 with 2, stopping, undefined, session 0.6 with 2, stopping\nsession 1.0 with 2, stopping, undefined, session 1.3 with 2, stopping, undefined, session 1.6 with 2, stopping\n"
    captured_union_test.go:95: leaks: LeakSanitizer ended the run, so its output wasn't compared: exit 1
        
        =================================================================
        ==212628==ERROR: LeakSanitizer: detected memory leaks
        
        Direct leak of 252 byte(s) in 3 object(s) allocated from:
            #0 0x564d99a673a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x564d99ad3dcf in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x564d99ad3dcf in adamic_allocate /adamic-runtime/heap.c:267:9
            #3 0x564d99aaae1d in adamic_function_alternating_ts_alternate_565c9a204ed396dab50febfce9b6094f /adamic-build/main.c:75:6
            #4 0x564d99aaaac0 in adamic_initialize_alternating_ts_798a1070fca02ff27cf4b4600fa7ac9f /adamic-build/main.c:404:2
        
        Direct leak of 252 byte(s) in 3 object(s) allocated from:
            #0 0x564d99a673a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x564d99ad3dcf in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x564d99ad3dcf in adamic_allocate /adamic-runtime/heap.c:267:9
            #3 0x564d99aaae1d in adamic_function_alternating_ts_alternate_565c9a204ed396dab50febfce9b6094f /adamic-build/main.c:75:6
            #4 0x564d99aaaa84 in adamic_initialize_alternating_ts_798a1070fca02ff27cf4b4600fa7ac9f /adamic-build/main.c:400:2
        
        Indirect leak of 834 byte(s) in 12 object(s) allocated from:
            #0 0x564d99a673a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x564d99ad3dcf in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x564d99ad3dcf in adamic_allocate /adamic-runtime/heap.c:267:9
        
        Indirect leak of 96 byte(s) in 6 object(s) allocated from:
            #0 0x564d99a673a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x564d99aaf77f in adamic_array_make /adamic-runtime/array.c:22:21
        
        SUMMARY: AddressSanitizer: 1434 byte(s) leaked in 24 allocation(s).
=== RUN   TestCapturedUnionAgrees/alternating.ts/counted
=== NAME  TestCapturedUnionAgrees/object.ts/sanitized
    captured_union_test.go:93: full got: exit 0, stdout "start undefined false\nstep 0 session first true object\nstep 1 stopping true string\nstep 2 undefined false undefined\nenabled\nenable while profiled\ndisabled profiled during stopping\nnothing to disable\nenable while stopping\nenabled\nafter a write the cell says string\nstopping\n", stderr ""; want stdout "start undefined false\nstep 0 session first true object\nstep 1 stopping true string\nstep 2 undefined false undefined\nenabled\nenable while profiled\ndisabled profiled during stopping\nnothing to disable\nenable while stopping\nenabled\nafter a write the cell says string\nstopping\n"
=== RUN   TestCapturedUnionAgrees/object.ts/counted
=== NAME  TestCapturedUnionAgrees/object.a/sanitized
    captured_union_test.go:93: full got: exit 0, stdout "start undefined false\nstep 0 session first true object\nstep 1 stopping true string\nstep 2 undefined false undefined\nenabled\nenable while profiled\ndisabled profiled during stopping\nnothing to disable\nenable while stopping\nenabled\nafter a write the cell says string\nstopping\n", stderr ""; want stdout "start undefined false\nstep 0 session first true object\nstep 1 stopping true string\nstep 2 undefined false undefined\nenabled\nenable while profiled\ndisabled profiled during stopping\nnothing to disable\nenable while stopping\nenabled\nafter a write the cell says string\nstopping\n"
=== RUN   TestCapturedUnionAgrees/object.a/counted
=== NAME  TestCapturedUnionAgrees/alternating.ts/counted
    captured_union_test.go:110: full got: exit 0, stdout "session 0.0 with 2, stopping, undefined, session 0.3 with 2, stopping, undefined, session 0.6 with 2, stopping\nsession 1.0 with 2, stopping, undefined, session 1.3 with 2, stopping, undefined, session 1.6 with 2, stopping\n", stderr "adamic: counts: allocations 54 frees 36 retains 120 releases 128 peak 27 regions 0\n"; want stdout "session 0.0 with 2, stopping, undefined, session 0.3 with 2, stopping, undefined, session 0.6 with 2, stopping\nsession 1.0 with 2, stopping, undefined, session 1.3 with 2, stopping, undefined, session 1.6 with 2, stopping\n"
    captured_union_test.go:112: counts: heap values leaked: 18 (allocations 54, frees 36, in regions 0)
    captured_union_test.go:118: counts: allocations 54, frees 36, in regions 0
=== RUN   TestCapturedUnionAgrees/alternating.ts/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 0.03
=== CONT  TestCapturedUnionAgrees/alternating.a
=== NAME  TestCapturedUnionAgrees/object.a/counted
    captured_union_test.go:110: full got: exit 0, stdout "start undefined false\nstep 0 session first true object\nstep 1 stopping true string\nstep 2 undefined false undefined\nenabled\nenable while profiled\ndisabled profiled during stopping\nnothing to disable\nenable while stopping\nenabled\nafter a write the cell says string\nstopping\n", stderr "adamic: graph region: live 7 bytes 432 reachable 1 bytes 32 unreachable 6 bytes 400 metadata 256\nadamic: graph counts: regions 1 merges 6\nadamic: counts: allocations 21 frees 21 retains 63 releases 80 peak 9 regions 0\n"; want stdout "start undefined false\nstep 0 session first true object\nstep 1 stopping true string\nstep 2 undefined false undefined\nenabled\nenable while profiled\ndisabled profiled during stopping\nnothing to disable\nenable while stopping\nenabled\nafter a write the cell says string\nstopping\n"
    captured_union_test.go:118: counts: allocations 21, frees 21, in regions 0
=== RUN   TestCapturedUnionAgrees/object.a/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 0.02
=== CONT  TestCapturedUnionAgrees/by_value.a
=== NAME  TestCapturedUnionAgrees/object.ts/counted
    captured_union_test.go:110: full got: exit 0, stdout "start undefined false\nstep 0 session first true object\nstep 1 stopping true string\nstep 2 undefined false undefined\nenabled\nenable while profiled\ndisabled profiled during stopping\nnothing to disable\nenable while stopping\nenabled\nafter a write the cell says string\nstopping\n", stderr "adamic: graph region: live 7 bytes 432 reachable 1 bytes 32 unreachable 6 bytes 400 metadata 256\nadamic: graph counts: regions 1 merges 6\nadamic: counts: allocations 21 frees 21 retains 63 releases 80 peak 9 regions 0\n"; want stdout "start undefined false\nstep 0 session first true object\nstep 1 stopping true string\nstep 2 undefined false undefined\nenabled\nenable while profiled\ndisabled profiled during stopping\nnothing to disable\nenable while stopping\nenabled\nafter a write the cell says string\nstopping\n"
    captured_union_test.go:118: counts: allocations 21, frees 21, in regions 0
=== RUN   TestCapturedUnionAgrees/object.ts/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 0.03
=== CONT  TestCapturedUnionAgrees/nested.a
=== RUN   TestCapturedUnionAgrees/alternating.a/native
=== RUN   TestCapturedUnionAgrees/by_value.a/native
=== RUN   TestCapturedUnionAgrees/nested.a/native
=== NAME  TestCapturedUnionAgrees/graph.a/sanitized
    captured_union_test.go:93: full got: exit 0, stdout "held undefined, held first, held stopping, held undefined, held again 3, held again 4, held stopping\nheld undefined, held first, held stopping, held undefined, held again 3, held again 4, held stopping\n", stderr ""; want stdout "held undefined, held first, held stopping, held undefined, held again 3, held again 4, held stopping\nheld undefined, held first, held stopping, held undefined, held again 3, held again 4, held stopping\n"
=== RUN   TestCapturedUnionAgrees/graph.a/counted
=== NAME  TestCapturedUnionAgrees/graph.ts/sanitized
    captured_union_test.go:93: full got: exit 0, stdout "held undefined, held first, held stopping, held undefined, held again 3, held again 4, held stopping\nheld undefined, held first, held stopping, held undefined, held again 3, held again 4, held stopping\n", stderr ""; want stdout "held undefined, held first, held stopping, held undefined, held again 3, held again 4, held stopping\nheld undefined, held first, held stopping, held undefined, held again 3, held again 4, held stopping\n"
=== RUN   TestCapturedUnionAgrees/graph.ts/counted
=== RUN   TestCapturedUnionAgrees/alternating.a/sanitized
=== RUN   TestCapturedUnionAgrees/by_value.a/sanitized
=== RUN   TestCapturedUnionAgrees/nested.a/sanitized
=== NAME  TestCapturedUnionAgrees/alternating.a/sanitized
    captured_union_test.go:93: full got: exit 1, stdout "session 0.0 with 2, stopping, undefined, session 0.3 with 2, stopping, undefined, session 0.6 with 2, stopping\nsession 1.0 with 2, stopping, undefined, session 1.3 with 2, stopping, undefined, session 1.6 with 2, stopping\n", stderr "\n=================================================================\n==212769==ERROR: LeakSanitizer: detected memory leaks\n\nDirect leak of 252 byte(s) in 3 object(s) allocated from:\n    #0 0x5608286b03a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3\n    #1 0x56082871cdcf in allocate_storage /adamic-runtime/heap.c:253:10\n    #2 0x56082871cdcf in adamic_allocate /adamic-runtime/heap.c:267:9\n    #3 0x5608286f3e1d in adamic_function_alternating_a_alternate_01e607951fdf653584ac75f3f81a1146 /adamic-build/main.c:75:6\n    #4 0x5608286f3ac0 in adamic_initialize_alternating_a_8aabf12e8e980896d6070e8d5697d4ba /adamic-build/main.c:404:2\n\nDirect leak of 252 byte(s) in 3 object(s) allocated from:\n    #0 0x5608286b03a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3\n    #1 0x56082871cdcf in allocate_storage /adamic-runtime/heap.c:253:10\n    #2 0x56082871cdcf in adamic_allocate /adamic-runtime/heap.c:267:9\n    #3 0x5608286f3e1d in adamic_function_alternating_a_alternate_01e607951fdf653584ac75f3f81a1146 /adamic-build/main.c:75:6\n    #4 0x5608286f3a84 in adamic_initialize_alternating_a_8aabf12e8e980896d6070e8d5697d4ba /adamic-build/main.c:400:2\n\nIndirect leak of 834 byte(s) in 12 object(s) allocated from:\n    #0 0x5608286b03a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3\n    #1 0x56082871cdcf in allocate_storage /adamic-runtime/heap.c:253:10\n    #2 0x56082871cdcf in adamic_allocate /adamic-runtime/heap.c:267:9\n\nIndirect leak of 96 byte(s) in 6 object(s) allocated from:\n    #0 0x5608286b03a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3\n    #1 0x5608286f877f in adamic_array_make /adamic-runtime/array.c:22:21\n\nSUMMARY: AddressSanitizer: 1434 byte(s) leaked in 24 allocation(s).\n"; want stdout "session 0.0 with 2, stopping, undefined, session 0.3 with 2, stopping, undefined, session 0.6 with 2, stopping\nsession 1.0 with 2, stopping, undefined, session 1.3 with 2, stopping, undefined, session 1.6 with 2, stopping\n"
    captured_union_test.go:95: leaks: LeakSanitizer ended the run, so its output wasn't compared: exit 1
        
        =================================================================
        ==212769==ERROR: LeakSanitizer: detected memory leaks
        
        Direct leak of 252 byte(s) in 3 object(s) allocated from:
            #0 0x5608286b03a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x56082871cdcf in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x56082871cdcf in adamic_allocate /adamic-runtime/heap.c:267:9
            #3 0x5608286f3e1d in adamic_function_alternating_a_alternate_01e607951fdf653584ac75f3f81a1146 /adamic-build/main.c:75:6
            #4 0x5608286f3ac0 in adamic_initialize_alternating_a_8aabf12e8e980896d6070e8d5697d4ba /adamic-build/main.c:404:2
        
        Direct leak of 252 byte(s) in 3 object(s) allocated from:
            #0 0x5608286b03a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x56082871cdcf in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x56082871cdcf in adamic_allocate /adamic-runtime/heap.c:267:9
            #3 0x5608286f3e1d in adamic_function_alternating_a_alternate_01e607951fdf653584ac75f3f81a1146 /adamic-build/main.c:75:6
            #4 0x5608286f3a84 in adamic_initialize_alternating_a_8aabf12e8e980896d6070e8d5697d4ba /adamic-build/main.c:400:2
        
        Indirect leak of 834 byte(s) in 12 object(s) allocated from:
            #0 0x5608286b03a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x56082871cdcf in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x56082871cdcf in adamic_allocate /adamic-runtime/heap.c:267:9
        
        Indirect leak of 96 byte(s) in 6 object(s) allocated from:
            #0 0x5608286b03a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x5608286f877f in adamic_array_make /adamic-runtime/array.c:22:21
        
        SUMMARY: AddressSanitizer: 1434 byte(s) leaked in 24 allocation(s).
=== RUN   TestCapturedUnionAgrees/alternating.a/counted
=== NAME  TestCapturedUnionAgrees/by_value.a/sanitized
    captured_union_test.go:93: full got: exit 1, stdout "step 0 false false false false true\nstep 1 true true true true false\nstep 2 false false false false true\nstep 3 true true true true false\nthe cell holds the word\nthe cell holds something else\nthe cell holds the word\n", stderr "\n=================================================================\n==212777==ERROR: LeakSanitizer: detected memory leaks\n\nDirect leak of 72 byte(s) in 1 object(s) allocated from:\n    #0 0x5618acdd93a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3\n    #1 0x5618ace45d2f in allocate_storage /adamic-runtime/heap.c:253:10\n    #2 0x5618ace45d2f in adamic_allocate /adamic-runtime/heap.c:267:9\n\nDirect leak of 66 byte(s) in 1 object(s) allocated from:\n    #0 0x5618acdd93a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3\n    #1 0x5618ace45d2f in allocate_storage /adamic-runtime/heap.c:253:10\n    #2 0x5618ace45d2f in adamic_allocate /adamic-runtime/heap.c:267:9\n    #3 0x5618ace1d02d in adamic_function_by_value_a_compare_1954618ce2d9e07b1980a9ac4e4160ae /adamic-build/main.c:136:2\n    #4 0x5618ace1d02d in adamic_initialize_by_value_a_4c44bd7515c4580a5a4f6d72acde192b /adamic-build/main.c:367:2\n    #5 0x5618ace1ca6d in adamic_initialize_by_value_a_e96b858367cf1e592bc28bb5116317f3 /adamic-build/main.c:377:2\n\nDirect leak of 66 byte(s) in 1 object(s) allocated from:\n    #0 0x5618acdd93a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3\n    #1 0x5618ace45d2f in allocate_storage /adamic-runtime/heap.c:253:10\n    #2 0x5618ace45d2f in adamic_allocate /adamic-runtime/heap.c:267:9\n    #3 0x5618ace1cdef in adamic_function_by_value_a_compare_1954618ce2d9e07b1980a9ac4e4160ae /adamic-build/main.c:91:6\n    #4 0x5618ace1cdef in adamic_initialize_by_value_a_4c44bd7515c4580a5a4f6d72acde192b /adamic-build/main.c:367:2\n    #5 0x5618ace1ca6d in adamic_initialize_by_value_a_e96b858367cf1e592bc28bb5116317f3 /adamic-build/main.c:377:2\n\nSUMMARY: AddressSanitizer: 204 byte(s) leaked in 3 allocation(s).\n"; want stdout "step 0 false false false false true\nstep 1 true true true true false\nstep 2 false false false false true\nstep 3 true true true true false\nthe cell holds the word\nthe cell holds something else\nthe cell holds the word\n"
    captured_union_test.go:95: leaks: LeakSanitizer ended the run, so its output wasn't compared: exit 1
        
        =================================================================
        ==212777==ERROR: LeakSanitizer: detected memory leaks
        
        Direct leak of 72 byte(s) in 1 object(s) allocated from:
            #0 0x5618acdd93a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x5618ace45d2f in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x5618ace45d2f in adamic_allocate /adamic-runtime/heap.c:267:9
        
        Direct leak of 66 byte(s) in 1 object(s) allocated from:
            #0 0x5618acdd93a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x5618ace45d2f in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x5618ace45d2f in adamic_allocate /adamic-runtime/heap.c:267:9
            #3 0x5618ace1d02d in adamic_function_by_value_a_compare_1954618ce2d9e07b1980a9ac4e4160ae /adamic-build/main.c:136:2
            #4 0x5618ace1d02d in adamic_initialize_by_value_a_4c44bd7515c4580a5a4f6d72acde192b /adamic-build/main.c:367:2
            #5 0x5618ace1ca6d in adamic_initialize_by_value_a_e96b858367cf1e592bc28bb5116317f3 /adamic-build/main.c:377:2
        
        Direct leak of 66 byte(s) in 1 object(s) allocated from:
            #0 0x5618acdd93a4 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
            #1 0x5618ace45d2f in allocate_storage /adamic-runtime/heap.c:253:10
            #2 0x5618ace45d2f in adamic_allocate /adamic-runtime/heap.c:267:9
            #3 0x5618ace1cdef in adamic_function_by_value_a_compare_1954618ce2d9e07b1980a9ac4e4160ae /adamic-build/main.c:91:6
            #4 0x5618ace1cdef in adamic_initialize_by_value_a_4c44bd7515c4580a5a4f6d72acde192b /adamic-build/main.c:367:2
            #5 0x5618ace1ca6d in adamic_initialize_by_value_a_e96b858367cf1e592bc28bb5116317f3 /adamic-build/main.c:377:2
        
        SUMMARY: AddressSanitizer: 204 byte(s) leaked in 3 allocation(s).
=== RUN   TestCapturedUnionAgrees/by_value.a/counted
=== NAME  TestCapturedUnionAgrees/nested.a/sanitized
    captured_union_test.go:93: full got: exit 0, stdout "no base\none base ./base.json\n3 bases ./a.json+./b.json+./c.json\n2 bases ./a.json+./b.json\none base ./base.json\nno base\nno base\n", stderr ""; want stdout "no base\none base ./base.json\n3 bases ./a.json+./b.json+./c.json\n2 bases ./a.json+./b.json\none base ./base.json\nno base\nno base\n"
=== RUN   TestCapturedUnionAgrees/nested.a/counted
=== NAME  TestCapturedUnionAgrees/alternating.a/counted
    captured_union_test.go:110: full got: exit 0, stdout "session 0.0 with 2, stopping, undefined, session 0.3 with 2, stopping, undefined, session 0.6 with 2, stopping\nsession 1.0 with 2, stopping, undefined, session 1.3 with 2, stopping, undefined, session 1.6 with 2, stopping\n", stderr "adamic: counts: allocations 54 frees 36 retains 120 releases 128 peak 27 regions 0\n"; want stdout "session 0.0 with 2, stopping, undefined, session 0.3 with 2, stopping, undefined, session 0.6 with 2, stopping\nsession 1.0 with 2, stopping, undefined, session 1.3 with 2, stopping, undefined, session 1.6 with 2, stopping\n"
    captured_union_test.go:112: counts: heap values leaked: 18 (allocations 54, frees 36, in regions 0)
    captured_union_test.go:118: counts: allocations 54, frees 36, in regions 0
=== RUN   TestCapturedUnionAgrees/alternating.a/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 0.01
=== NAME  TestCapturedUnionAgrees/nested.a/counted
    captured_union_test.go:110: full got: exit 0, stdout "no base\none base ./base.json\n3 bases ./a.json+./b.json+./c.json\n2 bases ./a.json+./b.json\none base ./base.json\nno base\nno base\n", stderr "adamic: graph region: live 4 bytes 208 reachable 1 bytes 32 unreachable 3 bytes 176 metadata 208\nadamic: graph region: live 7 bytes 360 reachable 1 bytes 32 unreachable 6 bytes 328 metadata 400\nadamic: graph region: live 7 bytes 360 reachable 1 bytes 32 unreachable 6 bytes 328 metadata 400\nadamic: graph region: live 10 bytes 512 reachable 1 bytes 32 unreachable 9 bytes 480 metadata 592\nadamic: graph region: live 10 bytes 512 reachable 1 bytes 32 unreachable 9 bytes 480 metadata 592\nadamic: graph region: live 10 bytes 512 reachable 1 bytes 32 unreachable 9 bytes 480 metadata 592\nadamic: graph region: live 7 bytes 360 reachable 1 bytes 32 unreachable 6 bytes 328 metadata 400\nadamic: graph counts: regions 16 merges 48\nadamic: counts: allocations 97 frees 97 retains 99 releases 151 peak 21 regions 0\n"; want stdout "no base\none base ./base.json\n3 bases ./a.json+./b.json+./c.json\n2 bases ./a.json+./b.json\none base ./base.json\nno base\nno base\n"
    captured_union_test.go:118: counts: allocations 97, frees 97, in regions 0
=== RUN   TestCapturedUnionAgrees/nested.a/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 0.02
=== CONT  TestCapturedUnionAgrees/class.a
=== NAME  TestCapturedUnionAgrees/by_value.a/counted
    captured_union_test.go:110: full got: exit 0, stdout "step 0 false false false false true\nstep 1 true true true true false\nstep 2 false false false false true\nstep 3 true true true true false\nthe cell holds the word\nthe cell holds something else\nthe cell holds the word\n", stderr "adamic: counts: allocations 35 frees 32 retains 65 releases 71 peak 10 regions 0\n"; want stdout "step 0 false false false false true\nstep 1 true true true true false\nstep 2 false false false false true\nstep 3 true true true true false\nthe cell holds the word\nthe cell holds something else\nthe cell holds the word\n"
    captured_union_test.go:112: counts: heap values leaked: 3 (allocations 35, frees 32, in regions 0)
    captured_union_test.go:118: counts: allocations 35, frees 32, in regions 0
=== RUN   TestCapturedUnionAgrees/by_value.a/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 0.02
=== RUN   TestCapturedUnionAgrees/class.a/native
=== RUN   TestCapturedUnionAgrees/class.a/sanitized
    captured_union_test.go:93: full got: exit 0, stdout "start undefined false\nstep 0 session first true object\nstep 1 stopping true string\nstep 2 undefined false undefined\nenabled\nenable while profiled\ndisabled profiled during stopping\nnothing to disable\nenable while stopping\nenabled\nafter a write the cell says string\nstopping\n", stderr ""; want stdout "start undefined false\nstep 0 session first true object\nstep 1 stopping true string\nstep 2 undefined false undefined\nenabled\nenable while profiled\ndisabled profiled during stopping\nnothing to disable\nenable while stopping\nenabled\nafter a write the cell says string\nstopping\n"
=== RUN   TestCapturedUnionAgrees/class.a/counted
    captured_union_test.go:110: full got: exit 0, stdout "start undefined false\nstep 0 session first true object\nstep 1 stopping true string\nstep 2 undefined false undefined\nenabled\nenable while profiled\ndisabled profiled during stopping\nnothing to disable\nenable while stopping\nenabled\nafter a write the cell says string\nstopping\n", stderr "adamic: graph region: live 7 bytes 432 reachable 1 bytes 32 unreachable 6 bytes 400 metadata 256\nadamic: graph counts: regions 1 merges 6\nadamic: counts: allocations 21 frees 21 retains 66 releases 83 peak 9 regions 0\n"; want stdout "start undefined false\nstep 0 session first true object\nstep 1 stopping true string\nstep 2 undefined false undefined\nenabled\nenable while profiled\ndisabled profiled during stopping\nnothing to disable\nenable while stopping\nenabled\nafter a write the cell says string\nstopping\n"
    captured_union_test.go:118: counts: allocations 21, frees 21, in regions 0
=== RUN   TestCapturedUnionAgrees/class.a/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 0.02
=== NAME  TestCapturedUnionAgrees/graph.ts/counted
    captured_union_test.go:110: full got: exit 0, stdout "held undefined, held first, held stopping, held undefined, held again 3, held again 4, held stopping\nheld undefined, held first, held stopping, held undefined, held again 3, held again 4, held stopping\n", stderr "adamic: graph region: live 5 bytes 316 reachable 1 bytes 32 unreachable 4 bytes 284 metadata 368\nadamic: graph region: live 5 bytes 316 reachable 1 bytes 32 unreachable 4 bytes 284 metadata 368\nadamic: graph counts: regions 4 merges 8\nadamic: counts: allocations 42 frees 42 retains 144 releases 164 peak 16 regions 0\n"; want stdout "held undefined, held first, held stopping, held undefined, held again 3, held again 4, held stopping\nheld undefined, held first, held stopping, held undefined, held again 3, held again 4, held stopping\n"
    captured_union_test.go:118: counts: allocations 42, frees 42, in regions 0
=== RUN   TestCapturedUnionAgrees/graph.ts/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 0.01
=== NAME  TestCapturedUnionAgrees/graph.a/counted
    captured_union_test.go:110: full got: exit 0, stdout "held undefined, held first, held stopping, held undefined, held again 3, held again 4, held stopping\nheld undefined, held first, held stopping, held undefined, held again 3, held again 4, held stopping\n", stderr "adamic: graph region: live 5 bytes 316 reachable 1 bytes 32 unreachable 4 bytes 284 metadata 368\nadamic: graph region: live 5 bytes 316 reachable 1 bytes 32 unreachable 4 bytes 284 metadata 368\nadamic: graph counts: regions 4 merges 8\nadamic: counts: allocations 42 frees 42 retains 144 releases 164 peak 16 regions 0\n"; want stdout "held undefined, held first, held stopping, held undefined, held again 3, held again 4, held stopping\nheld undefined, held first, held stopping, held undefined, held again 3, held again 4, held stopping\n"
    captured_union_test.go:118: counts: allocations 42, frees 42, in regions 0
=== RUN   TestCapturedUnionAgrees/graph.a/javascript
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 3d2bb0c39c0a hit 0.01
--- FAIL: TestCapturedUnionAgrees (0.00s)
    --- PASS: TestCapturedUnionAgrees/nested.ts (11.43s)
        --- PASS: TestCapturedUnionAgrees/nested.ts/native (0.45s)
        --- PASS: TestCapturedUnionAgrees/nested.ts/sanitized (1.02s)
        --- PASS: TestCapturedUnionAgrees/nested.ts/counted (0.47s)
        --- PASS: TestCapturedUnionAgrees/nested.ts/javascript (1.26s)
    --- FAIL: TestCapturedUnionAgrees/released.ts (11.48s)
        --- PASS: TestCapturedUnionAgrees/released.ts/native (0.39s)
        --- FAIL: TestCapturedUnionAgrees/released.ts/sanitized (1.05s)
        --- FAIL: TestCapturedUnionAgrees/released.ts/counted (0.47s)
        --- PASS: TestCapturedUnionAgrees/released.ts/javascript (1.30s)
    --- PASS: TestCapturedUnionAgrees/class.ts (11.57s)
        --- PASS: TestCapturedUnionAgrees/class.ts/native (0.64s)
        --- PASS: TestCapturedUnionAgrees/class.ts/sanitized (1.11s)
        --- PASS: TestCapturedUnionAgrees/class.ts/counted (0.59s)
        --- PASS: TestCapturedUnionAgrees/class.ts/javascript (0.90s)
    --- FAIL: TestCapturedUnionAgrees/by_value.ts (11.61s)
        --- PASS: TestCapturedUnionAgrees/by_value.ts/native (0.47s)
        --- FAIL: TestCapturedUnionAgrees/by_value.ts/sanitized (0.96s)
        --- FAIL: TestCapturedUnionAgrees/by_value.ts/counted (0.46s)
        --- PASS: TestCapturedUnionAgrees/by_value.ts/javascript (1.38s)
    --- PASS: TestCapturedUnionAgrees/cycle.ts (2.17s)
        --- PASS: TestCapturedUnionAgrees/cycle.ts/native (0.50s)
        --- PASS: TestCapturedUnionAgrees/cycle.ts/sanitized (0.64s)
        --- PASS: TestCapturedUnionAgrees/cycle.ts/counted (0.45s)
        --- PASS: TestCapturedUnionAgrees/cycle.ts/javascript (0.26s)
    --- PASS: TestCapturedUnionAgrees/cycle.a (2.23s)
        --- PASS: TestCapturedUnionAgrees/cycle.a/native (0.50s)
        --- PASS: TestCapturedUnionAgrees/cycle.a/sanitized (0.77s)
        --- PASS: TestCapturedUnionAgrees/cycle.a/counted (0.39s)
        --- PASS: TestCapturedUnionAgrees/cycle.a/javascript (0.15s)
    --- FAIL: TestCapturedUnionAgrees/released.a (2.42s)
        --- PASS: TestCapturedUnionAgrees/released.a/native (0.63s)
        --- FAIL: TestCapturedUnionAgrees/released.a/sanitized (0.74s)
        --- FAIL: TestCapturedUnionAgrees/released.a/counted (0.40s)
        --- PASS: TestCapturedUnionAgrees/released.a/javascript (0.30s)
    --- FAIL: TestCapturedUnionAgrees/alternating.ts (2.66s)
        --- PASS: TestCapturedUnionAgrees/alternating.ts/native (0.39s)
        --- FAIL: TestCapturedUnionAgrees/alternating.ts/sanitized (1.39s)
        --- FAIL: TestCapturedUnionAgrees/alternating.ts/counted (0.41s)
        --- PASS: TestCapturedUnionAgrees/alternating.ts/javascript (0.15s)
    --- PASS: TestCapturedUnionAgrees/object.a (2.89s)
        --- PASS: TestCapturedUnionAgrees/object.a/native (0.43s)
        --- PASS: TestCapturedUnionAgrees/object.a/sanitized (1.62s)
        --- PASS: TestCapturedUnionAgrees/object.a/counted (0.42s)
        --- PASS: TestCapturedUnionAgrees/object.a/javascript (0.13s)
    --- PASS: TestCapturedUnionAgrees/object.ts (3.22s)
        --- PASS: TestCapturedUnionAgrees/object.ts/native (0.58s)
        --- PASS: TestCapturedUnionAgrees/object.ts/sanitized (1.46s)
        --- PASS: TestCapturedUnionAgrees/object.ts/counted (0.60s)
        --- PASS: TestCapturedUnionAgrees/object.ts/javascript (0.17s)
    --- FAIL: TestCapturedUnionAgrees/alternating.a (2.14s)
        --- PASS: TestCapturedUnionAgrees/alternating.a/native (0.46s)
        --- FAIL: TestCapturedUnionAgrees/alternating.a/sanitized (0.82s)
        --- FAIL: TestCapturedUnionAgrees/alternating.a/counted (0.38s)
        --- PASS: TestCapturedUnionAgrees/alternating.a/javascript (0.11s)
    --- PASS: TestCapturedUnionAgrees/nested.a (1.86s)
        --- PASS: TestCapturedUnionAgrees/nested.a/native (0.46s)
        --- PASS: TestCapturedUnionAgrees/nested.a/sanitized (0.68s)
        --- PASS: TestCapturedUnionAgrees/nested.a/counted (0.28s)
        --- PASS: TestCapturedUnionAgrees/nested.a/javascript (0.17s)
    --- FAIL: TestCapturedUnionAgrees/by_value.a (2.10s)
        --- PASS: TestCapturedUnionAgrees/by_value.a/native (0.48s)
        --- FAIL: TestCapturedUnionAgrees/by_value.a/sanitized (0.73s)
        --- FAIL: TestCapturedUnionAgrees/by_value.a/counted (0.43s)
        --- PASS: TestCapturedUnionAgrees/by_value.a/javascript (0.13s)
    --- PASS: TestCapturedUnionAgrees/class.a (2.09s)
        --- PASS: TestCapturedUnionAgrees/class.a/native (0.47s)
        --- PASS: TestCapturedUnionAgrees/class.a/sanitized (0.68s)
        --- PASS: TestCapturedUnionAgrees/class.a/counted (0.50s)
        --- PASS: TestCapturedUnionAgrees/class.a/javascript (0.11s)
    --- PASS: TestCapturedUnionAgrees/graph.ts (20.80s)
        --- PASS: TestCapturedUnionAgrees/graph.ts/native (3.90s)
        --- PASS: TestCapturedUnionAgrees/graph.ts/sanitized (5.04s)
        --- PASS: TestCapturedUnionAgrees/graph.ts/counted (3.45s)
        --- PASS: TestCapturedUnionAgrees/graph.ts/javascript (0.07s)
    --- PASS: TestCapturedUnionAgrees/graph.a (9.25s)
        --- PASS: TestCapturedUnionAgrees/graph.a/native (0.50s)
        --- PASS: TestCapturedUnionAgrees/graph.a/sanitized (4.66s)
        --- PASS: TestCapturedUnionAgrees/graph.a/counted (3.53s)
        --- PASS: TestCapturedUnionAgrees/graph.a/javascript (0.11s)
FAIL
gate cache: native hits=0 misses=32
gate cache: node hits=0 misses=32
gate cache: probe hits=0 misses=0
FAIL	github.com/system-inc/adamic/internal/oracle	20.936s
FAIL
```


## Repository state and delivery

All temporary compiler, mutant, diagnostic-test, baseline-test, fixture, and dependency-link edits were restored. git diff --exit-code passed before report creation. The report branch is proof/lower-n7s-3-report, based directly on the requested HEAD; the only committed path is proof/lower-n7s-3/report.md. No compiler branch or main is pushed.

## Evidence hashes

- `M1-lower.log` SHA-256 `26ade5abe48b1fab0c3fe68919bad569ea6d8e4afc1700535b301d14d32e5a99`
- `M1-oracle.log` SHA-256 `e8e3d27280e670b5a831523043789d3a4159fbdc4807ed3f766b1e55c6a86dab`
- `M1-restore.log` SHA-256 `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855`
- `M2-lower.log` SHA-256 `302e6d94101115c66a437179397ae583107f789cc72a5710241442ceb04a83a5`
- `M2-oracle.log` SHA-256 `0e45e0fec6768e023bb434f0ad598aef0f752cc32c6bde2c1a64ac5b3bd1e0a1`
- `M2-restore.log` SHA-256 `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855`
- `M3-diagnostic.log` SHA-256 `41a31e3213ab9a445dfd11031dc323e8b7f2d8311d32a6ef1d49d51fa0115555`
- `M3-lower.log` SHA-256 `dee9edd5967dcd68aed4427174d3dc2d73f6fe6f5d6a3f33340d17d3e297b095`
- `M3-oracle.log` SHA-256 `d3a135b40a730b372d70d3a5676b5b4696fcee1b1395979d39ba4c11a233ff04`
- `M3-restore.log` SHA-256 `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855`
- `gate-attempt1-build.log` SHA-256 `e02ca47049870f312b4c55590bed8afe0c4d3c426c76ac36a940e501be10fbd1`
- `gate-build.log` SHA-256 `c7838f7670d0a4f9fa5a369169304765a2817adf1e3c1853d5f7e923ff8de088`
- `gofmt.log` SHA-256 `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855`
- `head-assignment-provisioned.log` SHA-256 `76aab303b25a16e76e5f2bcb8c48f184e08b41c6cac7f84d9b8e12b652c24168`
- `head-isolated-CountsAreRecorded.log` SHA-256 `61fe31c2d9e24611920d141cbfcb2e704f1bba2a7e67d42fa6bfd93eb4a95139`
- `head-isolated-Stage0EmitsDeterministically.log` SHA-256 `2f78b26f2884163d08d676331b9d57b5a2f8658275e35f81258e66b28a113622`
- `head-lower.log` SHA-256 `f8b006b810acde59511daf3b7fac1fd70f9138aef78fb6bdd4d5a0ce168673d6`
- `head-oracle.log` SHA-256 `6912e98b22c53bdc3828677cdd918f6b28d64c263bc57d05d9148298c772d050`
- `ir.log` SHA-256 `9f7df2fad77dd7ebf4ab6712971667e8564762f773d5517ca0cd70905b77bf20`
- `main-assignment-provisioned.log` SHA-256 `b73c843a24197b34b81960b3c049e5f1ce8703a1ca2bb58f99d54a4530222d38`
- `main-lower-0-transplanted.log` SHA-256 `44b150d6ed2e6fa449187dc001eae6b74b4955a8e8f3c2832a3e5df9e33e5e21`
- `main-lower-0.log` SHA-256 `735c8e373d4f332416ccbc0e37e28cfb92bd4988e9f32399bd189013f4b0bb9a`
- `main-oracle-0.log` SHA-256 `a1c18496aa959231824a797d5db9d4d1b52027427a25839bf2eb5eb316bf5a51`
- `main-oracle-1.log` SHA-256 `87e2a5b2fd23894e59e596af5e723aad02543d7b20046ef50a56286684527086`
- `main-oracle-10.log` SHA-256 `ad7f843b9f1839326bf0dbf85af04466aa25c6919a5e4f008303723188fd0fe1`
- `main-oracle-11.log` SHA-256 `3def42f7660beb8ee0282581cff32a497901dae987cd9d020b87a58d787db020`
- `main-oracle-12.log` SHA-256 `4f158fadefcb33a03d1d4f6ea781f797d0d252fc0e5a1d2dddee109df1d86e68`
- `main-oracle-13.log` SHA-256 `bb31b4afbdf994d594409a47bcf25af87249030dcb502081627397cbcd0e79d7`
- `main-oracle-14.log` SHA-256 `889b00d63abf2c7f5ce871329c834a105d6acc106f86b5640eb12f5f7fe4c3a2`
- `main-oracle-2.log` SHA-256 `f5c626091fe6ca8422293ac7e730744d27c963cf6ebbe6e0b6fd1e1154a62375`
- `main-oracle-3.log` SHA-256 `85cf0f970758384f18204f9e31cc6c1033192ef62ed7d476fe1a52ff33abdd99`
- `main-oracle-4.log` SHA-256 `298c4df1c782ccf388829c4a531df272b1267466d5f68d93d2987d458a89fead`
- `main-oracle-5.log` SHA-256 `8d8de43e0589e57d1b2ebf7962c33d3d5a6b3cf514536f022852c6d52354f28c`
- `main-oracle-6.log` SHA-256 `0b02660fcf038a4bfcc8eddc397065151698b5a086b6f63ae618df93e6e0d85e`
- `main-oracle-7.log` SHA-256 `38e59509feff05cae825ddc42a190c6b739cb03a8e020b839062fa64f4a023fe`
- `main-oracle-8.log` SHA-256 `c615fedeb9da469b4525c6e8286da2932a1cd8f83e28d4f66f0a53f4f1228c26`
- `main-oracle-9.log` SHA-256 `ace48280be24eb4a87cbdaf45f0a36b75cc5739cbc274e1e133a70019b3fccc8`
- `timeout-build.log` SHA-256 `a13682ceed8b8052d8f1a77ee83979285e38310ac666ee1ca9964d9484e8629c`
- `vet.log` SHA-256 `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855`
- `/tmp/lower.log` SHA-256 `d9a134974cfb47c8c7263c25e8bef8acf10a1ab48826dd67d6a75fc005b193a2`
- `/tmp/oracle.log` SHA-256 `eb19aa7e5f2d25f33bb155c4b45a98383dd95d593a9b06529c93c94e294b24da`
