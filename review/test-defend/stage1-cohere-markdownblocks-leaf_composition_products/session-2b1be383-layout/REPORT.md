# Markdown layout defense

Starting commit: 2b1be38362046455e0e5454d8b0e674a8630e95d. Prior audit base ce1c5a2f. Current discovery: 763 top-level tests. All three target members and all 16 list shards remain at their prior file locations. No tests were deleted, rewritten or weakened.

The grouped target is TestMarkdownLayout family: TestMarkdownCodeBlockLayout, TestMarkdownHTMLBlockLayout, TestMarkdownRootLayout. Its subsumer is TestMarkdownListLayout family (16 shards plus union). CODE UNDER TEST and ORACLE, aimed plans and input comparison are in code-and-oracle.md. Source Node execution is the port under test; Go cohere and the retained fork decide the expected semantic bytes.

## Result

Not defended by these three attempts. D1 loses nonempty fenced-code metadata; both families catch it. D2 changes the minimum HTML comment length from 7 to 9; both families pass. A standalone minimal multiline-comment witness also has identical printed bytes, so D2 is an equivalent candidate for these observations, not evidence of unguarded output. D3 skips the first interior character of an ignored source range; both families catch it. These are bounded negative findings, not permission or a recommendation to delete the target.

The target owns nine built-in code/HTML/root mutant assertions; the list family owns three list-specific mutant assertions. That harness-witness distinction is real, but no harness was weakened in this defense and no unique production kill was demonstrated for it. The names promise code/HTML/root layout agreement, and the assertions compare complete outputs against independent answers. No missing-threshold or count-only name/assertion gap was found. A shared-fixture failure prevents the other two wrappers from entering their own built-in-mutant body; the grouped row avoids treating those as three independent catches.

## Coverage and scope limits

The clean whole package timed out at 90.101 seconds without an assertion failure. The narrowed clean target and list families passed, without skips. Go compiler coverage used -coverpkg=./internal/lower: target 67.996 binary seconds, subsumer 42.797. The profiles and block differences are saved. The 2,919 target-only compiler blocks are not port exclusivity: cached list products avoid compiler work. Go coverage cannot instrument this TypeScript port or its generated native products. Exact requested per-port Go line coverage therefore cannot be produced with this tool. No exclusive TypeScript lines are claimed.

Both target and list shards construct inputs with blockCorpus(root, "whitespace"). List shards append three controls and partition that same cumulative corpus; the target runs it as one batch. The batch history is a possible defense lead, but these three component mutations do not alter persistent cross-document state. No input-exclusive component behavior was found. Node/native are executors inside the existing grouped checks, not separately judged twin rows.

135 current tests in layout call-owner files, including separate per-input wrappers, are enumerated in static-reach.json. Expanded replay commands select that complete name set. Timeouts abort late observations; those rows are unknown, never recorded as passing or mutant kills merely because the package timed out. The 19-row target/list matrices completed for every mutant, without panic or skip; their observed failures establish nonuniqueness even when expanded rows are unknown. Other package rows and repo-wide results are unknown.

## Evidence and validation

D1 codeblocks.ts:34 replaces metadata expression with empty string.
D2 htmlblocks.ts:34 changes value.length >= 7 to value.length >= 9.
D3 root.ts:76 changes source.slice(child.endOffset, last.startOffset) to source.slice(child.endOffset + 1, last.startOffset).
All locations refer to the starting main commit. D*.diff are standalone, without a switch, and apply to that base. Each matrix uses ADAMIC_BUILD_CACHE_DIR=/tmp/defend-markdown/cache/<ID>. In the target fixture nativeTask.await succeeds before source Node byte comparison at lists_test.go:342, so D1/D3 reached that failure only after the mutated port's sanitized native build and execution returned. D2's full passing target run also validates native builds and comparisons. Build preparation is included in recorded command times, not falsely reported as a separately measured clang duration.

## Costs and unclear parts

Warm env.sh worked; no tool setup. npm ci in stage3/api reported 485ms; nproc 5. The completed six narrow mutant commands took 418.786s wall total. Two clean coverage commands took 119.685s wall. The first whole baseline cost 90.101s in the binary. Expanded run timings are in expanded-runs.json.

The brief requests Go coverage of a TypeScript code-under-test port; this measurement is unavailable. Compiler profiles answer a different question and cache hits distort their exclusive blocks. The family is supplied as one row, so coverage was collected over the family's members together, not as three separate assertions of exclusivity. Current package discovery is far larger than the historical 35-name audit slice, which made full and expanded baselines expensive. The twin instruction adds a defense value absent from the final JSON enum; no separate twin rows occur here. Two meaningful shared kills and one surviving equivalent candidate do not exhaust all possible compiler, runtime or input-history differences. This defense does not settle whether a different aimed production break could uniquely catch the target. No other package was run.

Expanded replays all timed out without additional semantic observations. Expanded command wall total: 276.739s. Final UTC: 2026-10-09T15:23:49.132511+00:00. All production source restored, standalone application checks passed. Full known/pass/fail/unknown matrix is matrix.json. No other package tests, PR, or main push.
