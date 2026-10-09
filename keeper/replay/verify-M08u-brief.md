New unit for you, from @system_adamic_tests (the keeper of Adamic's test audit, #xphstyt): rp:verify-M08u. Your workspace is warm, so skip setup if /workspace/adamic-tools/env.sh works. Check `df -h /tmp /workspace` first. Write evidence outside the repository, then copy it in.

## What this settles
Your rp:verify-u045 run showed TestStatementRegionsAreUsed passes under u045 M08, because it reads internal/oracle/counts.md rather than measuring. Its design is a pair: TestCountsAreRecorded keeps counts.md equal to what's measured, and TestStatementRegionsAreUsed refuses a counts.md with no statement regions. So the realistic way M08 lands is someone running `-update-counts` under it. Prove that path fails, on main 946a8f095a (submodules at its pins, a fresh ADAMIC_BUILD_CACHE_DIR):

1. Apply the diff below.
2. Run `go test ./internal/oracle -count=1 -run '^TestCountsAreRecorded$' -args -update-counts`. This rewrites counts.md under the mutant. Report how many rows changed (`git diff --stat internal/oracle/counts.md`).
3. Then run `go test ./internal/oracle -count=1 -run '^(TestCountsAreRecorded|TestStatementRegionsAreUsed)$'`. TestCountsAreRecorded should pass now, since it agrees with itself. Does TestStatementRegionsAreUsed fail? Give its failing lines.
4. Restore counts.md and the source (`git checkout -- internal/oracle/counts.md internal/native/region.go` in your own checkout), and run the same two tests at base: both pass.

Don't push code. Put the logs on test-audit/u045-verify under review/test-audit/u045-verify/M08u/, and push that branch only.
Reply in two sentences of prose, then:
{"main":"946a8f095a","counts_rows_changed":0,"after_update":{"TestCountsAreRecorded":0,"TestStatementRegionsAreUsed":1},"failing_lines":["..."],"base_after_restore":{"TestCountsAreRecorded":0,"TestStatementRegionsAreUsed":0}}

## The diff
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
