New unit for you, from @system_adamic_tests (the keeper of Adamic's test audit, #xphstyt). Your workspace is warm, so skip setup if /workspace/adamic-tools/env.sh works. Check `df -h /tmp /workspace` first, and clear earlier units' scratch and per-mutant caches under /tmp if space is low (never /workspace/adamic or your tools). Write evidence outside the repository (review/ in the tree upsets corpus setup), then copy it in at the end.

## rp:M06p: a program whose output tells u045 M06 apart
M06 removes `held.Captured ||` from internal/native/element_borrow.go:276 (the diff is below). That lets the compiler's element-borrow optimization apply to a local that a closure captured. No test in any covered package fails under it, even run alone at base bc9edc5560379b63d4688a21006efb6032b60434 (test-audit/u045-M06i). The question is whether the condition is needed for correct output.

1. Read element_borrow.go whole: what the optimization does, what chainUnchanged and assigned[root] already exclude, and why a captured local is treated separately. Then read how internal/oracle's fixtures run (testdata/*.a and the test that compares native output with Node).
2. Write the smallest Adamic program you can where a local array or object is captured by a closure, an element of it is read into a variable, the closure then replaces or changes that local, and the program prints the element afterwards. Try the variations the code suggests: replace or push, array or object, closure called directly or passed as a callback.
3. Run each candidate the way the oracle does, at base and with the diff, and compare stdout and exit status with Node. A useful program matches Node at base and differs from Node under the diff, or the oracle's standard native build reports an error under the diff.

If you find one, add it as a fixture-shaped test on test-audit/u045-M06p (in internal/oracle or internal/native, matching that package's tests). Show it fails with the diff and passes without, and give the exact failing line.
If after honest attempts no program's output changes, say why: for example, a captured local always counts as assigned, or a closure call always breaks chainUnchanged. Name the file and line that already covers it. "The condition is redundant, here is the line" is a good answer.
Restore the tree.

Reply in two or three sentences of prose, then this JSON:
{"mutant":"u045 M06","base":"bc9edc55","probe":"path or null","probe_source":"...","base_vs_node":"agree|disagree","mutant_vs_node":"agree|disagree|build_error","failing_line":null,"redundant_because":null,"branch":null}

## The diff
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
