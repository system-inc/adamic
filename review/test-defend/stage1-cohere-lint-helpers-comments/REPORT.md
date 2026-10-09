# Consumer comment helper defense

Keep TestConsumerCommentHelpers: D1 makes it fail and all ten other current top-level tests pass. Base: 157a43552015f41a79331949c2e82b6f8c7caaab. Warm toolchain, setup skipped; nproc 5; npm ci 412 ms. No tests were removed, weakened or rewritten.

Code under test: the TypeScript comments helper port, its driver and AST adapter. Oracle: live Go cohere, comparing complete output with Node and sanitized native. The consumer corpus supplies real Go AST geometry, not expected comment answers. Read audit README, rows, functions, plan and scope, both test files whole, and all owned TypeScript source files before mutation.

The original audit rested subsumption on one shared sorting mutant. The consumer row supplies JSX AST nodes that the independent-parser corpus does not produce. D1 off-by-ones the interior anchor of an empty JsxExpression from node.pos + 1 to node.pos + 0. In the consumer corpus a comment inside JSX braces then disappears. Semantic lead and example are in semantic-lead.json. This is a real port mutation, not an oracle, fixture or test edit.

The failure is comments_test.go:115: output line 695: got "cached 0 0", Go "cached 1 1". All ten other current tests pass under D1, including every compiled-mutant witness. Exact names and outcomes are in matrix.json. Discovery still has the same 11 top-level names as the audit. No skip was observed.

Coverage: two per-test Go profiles instrument internal/lower, internal/native and internal/load, because Go coverage cannot instrument this TypeScript program or its subprocess oracle. The profiles have zero consumer-exclusive Go blocks. They measure preparation, not port execution. No claim of TypeScript-exclusive coverage is made. The defense is the permitted semantic difference on shared code.

Baseline and budget: the full clean package timed out at 90.015 seconds after all three production rows passed, while building the adapter witness. Remaining clean witness/setup rows passed separately in 50.267 seconds. The mutant matrix is partitioned into the three direct production rows (45.007 seconds) and all remaining rows (50.613 seconds). Together these cover every current top-level test, so the catch is package-unique. Timeouts are not kills. Coverage clean runs took 19.660 and 35.568 binary seconds.

Compilation: the unchanged TestCommentsMatchCohere and TestJsxParserGapIsExplicit built the mutated port with native.Build and Sanitize:true and passed. This proves the standalone source change compiles with the actual toolchain. The consumer failure occurs at its Node comparison; its later native comparison was not reached and is not claimed as a native kill. D1 uses ADAMIC_BUILD_CACHE_DIR=/tmp/defend-comments/cache/D1. The standalone diff applies to the recorded origin/main base and was apply-checked after restoring production source.

Brief costs and ambiguities: Go coverpkg cannot measure a TypeScript port; compiler coverage is supplementary. The 90-second full-package allowance is shorter than this package's sequential native-build work, requiring two matrices. The requested unique catch is still observed across all rows through these partitions. The audit's evidence in the prompt was truncated, but its complete report and logs were accessible through the full refspec. One successful aimed mutant sufficed; the allowance is up to three, not a requirement after defense.

The name promises consumer helper behavior and its full-output assertions check it under the stated Go AST adapter contract. It does not promise an independent JSX parser or completed lint-rule integration. No name/assertion mismatch found for the defended row. Other packages and repo-wide uniqueness were not tested. Production source restored before evidence commit.
