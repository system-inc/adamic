New unit for you, from @system_adamic_tests (the keeper of Adamic's test audit, #xphstyt): rp:M06i, a causality check. Your workspace is warm, so skip setup if /workspace/adamic-tools/env.sh works. Before anything else, run `df -h /tmp /workspace`. If either has under 15 GB free, delete earlier units' scratch and per-mutant cache directories under /tmp (never /workspace/adamic or your tools).

## What this settles
Mutant u045 M06 drops `held.Captured ||` from the condition at internal/native/element_borrow.go:276, so a root a closure captured may now lend borrowed elements. The diff is below. A covered-packages replay at base bc9edc5560379b63d4688a21006efb6032b60434 went red in only three places, all under heavy parallel load:
1. internal/native TestDecodeASCIIUnit26: decode-native "signal: killed".
2. stage1/cohere/json TestPortMatchesGoCohere_123: "exceeded its 90s deadline".
3. stage1/cohere/yaml TestFileDriver_Setup: "panic: test timed out after 1m30s".
Each one passed when run alone at base. The question: does M06 cause any of them?

## Do exactly this
Check out bc9edc5560379b63d4688a21006efb6032b60434 detached, with submodules at that commit's pins. Use a fresh build cache directory for each side.
For each of the three tests, run alone with `-count=1 -timeout 10m`, including the family's _Setup where the test needs it (json: `-run '^TestPortMatchesGoCohere_(Setup|123)$'`):
- base, then with the diff applied, then base again.
Record exit status, wall seconds and peak RSS for each run (`/usr/bin/time -v`), and for a kill, the signal and whether it was the OOM killer (dmesg if you can read it).
- If a test is clean at base twice and fails, is killed, or runs over 3x its base time under the mutant, that is a catch. Give the failing line.
- If every mutant run is clean and within 1.5x of base, M06 survives these three.

Don't touch anything else, and don't push code. Put the logs on test-audit/u045-M06i under review/test-audit/u045-M06i/ and push that branch only. Restore the tree when done.

Reply in plain prose (two or three sentences), then this JSON:
{"mutant":"u045 M06","base":"bc9edc55","runs":[{"test":"...","side":"base|mutant|base2","exit":0,"seconds":0,"peak_rss_mb":0,"signal":null,"line":null}],"caught_by":[],"survived":true}

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
