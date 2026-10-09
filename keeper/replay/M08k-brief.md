New unit for you, from @system_adamic_tests (the keeper of Adamic's test audit, #xphstyt). Your workspace is warm, so skip setup if /workspace/adamic-tools/env.sh works. Check `df -h /tmp /workspace` first, and clear earlier units' scratch and per-mutant caches under /tmp if space is low (never /workspace/adamic or your tools). Write evidence outside the repository (review/ in the tree upsets corpus setup), then copy it in at the end.

## rp:M08k: what u045 M08 actually changes
M08 inverts the region plan at internal/native/region.go:86 (diff below). Base: bc9edc5560379b63d4688a21006efb6032b60434, submodules at its pins, a fresh ADAMIC_BUILD_CACHE_DIR per run.

A covered-packages replay went red only on two kinds of test. The first is resource-sensitive tests under heavy load, each green on a clean control. The second is internal/oracle TestCountsAreRecorded, the counts.md golden. Do this:
1. Run internal/oracle TestCountsAreRecorded alone at base, then with the diff. Copy every recorded and measured row it prints under the mutant. Say which columns moved (allocations, frees, retains, releases, peak, in regions) and in which direction, summed over all fixtures.
2. With the diff, run internal/oracle's agreement tests alone with -count=1 -timeout 30m: `-run '^(TestNativeAgreesWithNode|TestInputAgreesWithNode)$'`, or whatever the package's native-against-Node agreement tests are named at this base (name them). Do any fail, or any ASan or UBSan report? Give the line.
3. With the diff, run each of these alone with -count=1 -timeout 10m. Each must finish under the mutant, or fail with a real assertion, not a kill or deadline: internal/native TestDecodeASCIIUnit40 to 44, stage1/cohere/css TestCSSPrinterAgreesWithGo_000, stage1/cohere/cssnumbers TestCSSNumbers_350, stage1/cohere/json TestPortMatchesGoCohere_1649 (with TestPortMatchesGoCohereSplit_Setup), stage1/cohere/yaml TestFileDriver_Setup. For any that fails, rerun it alone at base twice.

Push the logs on test-audit/u045-M08k under review/test-audit/u045-M08k/, and push that branch only. Restore the tree.
Reply in two or three sentences of prose, then this JSON:
{"mutant":"u045 M08","base":"bc9edc55","counts_moved":{"fixtures":0,"columns":{"allocations":0,"frees":0,"retains":0,"releases":0,"peak":0,"in_regions":0},"sample_rows":["recorded: ...\nmeasured: ..."]},"agreement_failures":[],"sanitizer_reports":[],"isolated":[{"test":"...","exit":0,"seconds":0,"line":null}],"real_catches":[]}

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
