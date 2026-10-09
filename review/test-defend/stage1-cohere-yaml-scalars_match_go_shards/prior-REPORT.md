u153: 19 named functions present, grouped into 11 rows at 3bf0a5d9e74d197a38f61982e43b2d791f17b563.
Every grouped row passed three isolated clean runs; no selected skips observed.
Bounded verdicts: one sacred, one overlapping, three subsumed, three setup-checks, three witnesses.
Four production mutants: three caught, one survivor with an initialization-state witness; eight separate empty-answer probes.
Three product readiness rows are vacuous; production source restored; no package-wide or repo-wide uniqueness claim.

The complete requested row JSON is rows.json. matrix.json records grouped results; menu.json fixes production mutations and origin lines. Every standalone diff is switch-free and applies to the starting index (apply-validation.json). Standalone Go variants pass go vet with the warm Go toolchain. Port variants and the runtime header compile through native.Build before their observed runtime results. W/S/P are separate witness, construction and probe evidence, never production kills.

CODE UNDER TEST: the YAML TypeScript port compiled natively, Adamic's native string runtime, lowering for standalone cost fixtures, and construction helpers for readiness rows. ORACLE: unchanged Go cohere and the installed independent YAML libraries for agreement; handwritten checksums additionally constrain cost fixtures; coverage ownership and deliberately corrupted stdout constrain witnesses. functions.txt conservatively lists declarations in the transitive TypeScript module closure, not a dynamic proof that every declaration executes.

Production matrix:
- M1 scalar.ts:221, BLOCK_FOLDED selector constant '>' becomes '|'. Scalar family and unist agreement fail.
- M2 schemaPattern.ts:125, '+' minimum 1 becomes 0. Schema fails; unist passes.
- M3 unistContext.ts:58, initial line 0 becomes 1. Unist agreement passes. In isolated port copies, a fresh unparsed parser produces offset(0) = 0,0,0 before, 1,1,0 after. See M3-witness.ts and original/mutant logs. This witness runs the port in Node; the production matrix also rebuilt the native port successfully. The finding concerns the unparsed initialization state, not normal parsed documents.
- M4 internal/native/runtime/adamic.h:742, ASCII fast-path equality becomes inequality. Cost checksum and width bytes fail. Scalar _000 alone fails at its own 55-second subprocess deadline. Schema passes. Unist remains unknown after its isolated 90-second timeout. The grouped scalar run also cooked; only the completed _000 assertion proves its family kill.

Subsumption is a hint from finite observed mutants. Unist is subsumed by the scalar family on M1 alone; its M4 outcome is unknown. The cost row is subsumed by the scalar family on M4 alone. Width is subsumed by the cost row on M4 alone. Scalar overlaps unist and cost, which collectively cover its two kills. Schema's M2 uniqueness is within this bounded matrix only.

Brief ambiguities, incorrect assumptions and time costs:

1. The brief says 11 rows but lists 19 functions. Its explicit family rule explains the count: union plus eight scalar shards form one family. The union checks ownership construction, while the leaves check runtime bytes. The exception for siblings with extra assertions conflicts with the instruction to group unions. This audit follows the explicit union family rule and records the union separately as a probe-passing construction member.

2. Scope is described using historical commit 8de93800f4, but the start command requires current origin/main. The fetched starting commit is 3bf0a5d9, and every named function is still present in the named files. Every production location and diff refers to that actual starting commit.

3. The requested whole-package command does not enforce 90 seconds over TestMain plus tests. YAML's TestMain runs a separate setup child before m.Run. The whole-package attempt hit the outer 120-second backstop; the selected combined attempt also reached the binary's 90-second limit. Neither completed baseline is classified red. Every selected row subsequently passed isolated three times before production mutations.

4. The shared run helper's childguard default stall window is two minutes, longer than the audit's 90-second binary budget. A hanging native program therefore aborts the Go binary before childguard provides a normal failure. This happened under M4 for unist. The result is unknown, not a kill. Scalar leaves have their own 55-second subprocess deadline; _000 provided a completed failure. This difference in guards affects the evidence that can be collected within the budget.

5. Timed-out binaries left native children orphaned. Only identified orphan processes from these audit runs were stopped, after their parents exited. Active mutant children were left to their test's deadline. The completed scalar signal-killed assertion came from that deadline, not cleanup.

6. Product tests ignore the helper return value. Each passes when its own entry immediately returns an empty directory. Yet invalid construction input/options make each fail. The setup-check verdict and vacuous finding therefore coexist. A build completed and an answer was checked are different claims; the readiness guards prove only the first when the helper executes.

7. A survivor needs an output witness before being called unguarded. M3 changes fresh-parser points but normal parsed inputs reset the line before observing it. The witness is synthetic and its initialization-state scope is explicit. It does not establish that a normal YAML document receives the wrong point.

8. Full byte comparisons are substantially stronger than mere exit-code checks. Cost probes instead compare handwritten checksums, although Node and emitted JavaScript are also executed. Equal sums could hide compensating errors. Readiness rows rely on construction errors and nonempty-directory guards; they do not compare artifact contents.

9. The SDK opt-in text names YAML and Prettier but unist additionally imports yaml-unist-parser. Installing that package initially advanced YAML through its npm range; the required YAML and Prettier versions were pinned before the unist oracle ran. Actual versions are yaml 2.9.0, prettier 3.9.6 and yaml-unist-parser 3.2.1. The test checks the YAML version but does not pin the unist package version, despite a source comment naming 3.2.0. The lock file preserves the versions used here.

10. The warm env must be sourced for each independent command. The default executable named go is another tool and rejects vet. Initial standalone validation attempts hit that problem, then all nine standalone Go variants were revalidated successfully with env.sh sourced. No failed validation is presented as a successful build.

11. A switch-free standalone diff and a scratch runtime selector are different artifacts. Harness runs use a combined Go overlay with selectors, while saved standalone variants remove selectors entirely. Each saved Go variant was checked separately. Original log locations shifted by three inserted probe lines in that overlay; row JSON maps those locations back to origin, while raw logs preserve the observed locations.

12. Four native production mutations limit the strength of a 11-row unit. The brief also asks about three mutants per row, but its stage1 rebuild limit takes precedence. No row deletion is recommended. Mutations were selected from code before observing failures; two initial expression-offset choices were tightened to literal constant changes and a condition flip before planting any mutant.

Timing and limits:
Warm setup skipped, nproc 5. Logged dependency operations total about 3.48 seconds. Isolated timing test-binary lines total 362.634 seconds, wall 432.535 seconds. Witness/construction/product-probe wall totals 84.401 seconds. Production matrix, adaptive reruns and port probes total 460.735 seconds; _000 adds 55.735 binary seconds. Full baseline hit outer 120 seconds; selected baseline hit 90.018 binary seconds. Production first-run wall times: M1 48.316, M2 41.182, M3 36.449, M4 95.009 (cooked). M1 cached product misses separately report Go build 0.50, lowering 2.37, native build 3.73 seconds. Other tests do not log native build durations separately, so their run wall is an upper bound, not a claimed exact rebuild time. Total session exceeded the approximate 30-minute port budget while collecting isolated timings, timeout reruns and evidence.

Not covered: package callers outside this unit; repo-wide uniqueness; a completed unist result for M4; dynamic execution of every declaration in functions.txt; independent exact native-build timings where the tests emit none; performance throughput beyond the checksum checks. These are explicit evidence limits, not inferred verdicts.
