New unit for you, from @system_adamic_tests (the keeper of Adamic's test audit, #xphstyt): rp:verify-u045. Your workspace is warm, so skip setup if /workspace/adamic-tools/env.sh works. Check `df -h /tmp /workspace` first and clear earlier units' scratch under /tmp if space is low. Write evidence outside the repository, then copy it in.

## What this settles
Two new tests landed on main to close two mutants the audit found unguarded. Prove each one fails under its mutant and passes without it, on main `946a8f095a7fa419a92117406314b7b3d44630f0` (submodules at its pins, a fresh ADAMIC_BUILD_CACHE_DIR per run).

1. u045 M06 (diff A below). PR #300 added internal/oracle/testdata/borrow_chain_captured_root.a, a closure that reassigns a captured local after an element was read from it. Find the test that runs this fixture against Node (read internal/oracle/oracle_test.go). Run only that fixture's case, alone with -count=1: base, then with diff A, then base again.
2. u045 M08 (diff B below). PR #301 added internal/oracle/statement_regions_test.go. Run TestStatementRegionsAreUsed alone with -count=1: base, then with diff B, then base again.

If a diff doesn't apply cleanly at this main, say so, show the current code at that spot, and apply the same one-token change by hand.
For each run, record the exit status, wall seconds and the first failing line. A mutant run that fails only by kill or deadline doesn't count. Rerun it once alone.

Don't push code. Put the logs on test-audit/u045-verify under review/test-audit/u045-verify/, and push that branch only. Restore the tree.
Reply in two sentences of prose, then:
{"main":"946a8f095a","M06":{"test":"...","base":0,"mutant":1,"base2":0,"failing_line":"...","applied":"clean|by-hand"},"M08":{"test":"TestStatementRegionsAreUsed","base":0,"mutant":1,"base2":0,"failing_line":"...","applied":"clean|by-hand"}}

## Diff A (M06)
```diff
--- a/internal/native/element_borrow.go
+++ b/internal/native/element_borrow.go
@@ -273,7 +273,7 @@
 		return false
 	}
 	held := program.Locals[root]
-	if held.Global || held.Captured || held.Function != function || assigned[root] {
+	if held.Global || held.Function != function || assigned[root] {
 		return false
 	}
 	return chainUnchanged(program, program.Functions[function].Body, names)
```

## Diff B (M08)
```diff
--- a/internal/native/region.go
+++ b/internal/native/region.go
@@ -83,7 +83,7 @@
 		for index := range list {
 			switch statement := list[index].(type) {
 			case ir.Assign, ir.Declare, ir.Evaluate, ir.WriteLine:
-				if plan.feedsRegion(statement) {
+				if !plan.feedsRegion(statement) {
 					plan.statements[&list[index]] = true
 				}
 			}
```
