Three aimed production mutants failed both the Node string row and its census subsumer.
Defense: not defended on these three attempts; no deletion recommendation.
Whole-package baseline timed out; all ten property callers completed the narrowed matrix.

CODE UNDER TEST: Adamic Go Lookup, expressionOK, splitPair, codePoint and generated string-property tables. ORACLE: Node 24 RegExp with the v flag checks every returned sequence. The census checks stored sequences against fixed counts and SHA-256 digests, then checks Lookup length/kind and first-sequence membership through ContainsSequence. Neither oracle, test nor harness was mutated.

Starting origin/main is recorded in origin.txt. The current go test -list inventory has 51 functions, identical to the audited inventory. No new or vanished tests. Warm toolchain, nproc=5, no setup. npm ci in stage3/api ran before baseline. Node scripts use built-in fs only.

Disk check: /tmp initially 8.1GB free, /workspace 17GB. Removed only the prior unit /tmp/defend-yepesta scratch directory. Recheck /tmp 8.6GB free, /workspace 17GB. The /tmp filesystem is only 8.8GB total, so 15GB free is impossible; this did not cause a test failure.

Coverage: independently ran each row with -coverpkg=./internal/unicodeproperties and -coverprofile. Node row 0.210s, 18.3%; census 0.008s, 22.1%. Node had zero exclusive covered blocks; census four, all ContainsSequence. Eighteen production blocks were shared. coverage-diff.json lists exact source spans. Exclusive lines were not assumed necessary for defense. The semantic difference is that Node checks all returned sequences for actual membership, while the census hashes stored tables and samples lookup membership. Each attempt preserves count and attacks a different sequence class instead of repeating the audit's ASCII replacement of the first sequence.

Attempts, standalone diffs against origin/main:
D1 tables.go:25713: interior Basic_Emoji registered-sign selector FE0F becomes FE0E. Node says it does not match; census digest changes. Both fail, eight other callers pass. Binary 7.872s.
D2 tables.go:27119: keycap base 2 becomes A. Node rejects the unchanged-length malformed keycap; census digest changes. Both fail, eight pass. Binary 7.725s.
D3 tables.go:28061: Scotland tag final letter t becomes u. Node rejects the unsupported tag sequence; census digest changes. Both fail, eight pass. Binary time in D3.json.
All changes are constant substitutions from the permitted menu. Diffs passed go vet ./internal/unicodeproperties/ individually, and applied sources compiled and ran. Production sources were restored. Separate ADAMIC_BUILD_CACHE_DIR values supplied for every run, though Go table changes do not build native products. No switches remain in production. Each mutant has complete statuses and failing lines in its JSON and compressed log.

Whole clean package timed out at 90.010 binary seconds during canonicalization corpus, with no earlier assertion failure. It was cooked rather than treated as red. Narrowed clean baseline passed at 7.760s. All ten direct property callers are listed in matrix-rows.json and were run per mutant. Canonicalization uses unicodeFold/unicodeClass/legacyFold tables, not Lookup or these string tables. Its expensive range wrappers and setup/witness tests were excluded after timeout. Their dynamic catches remain unknown; no uniqueness claim is made. No narrowed run timed out, skipped or aborted. Complete package execution and repository-wide replay were not completed.

Owner finding and brief friction
The name TestNodeStringProperties promises a Node membership check, and the assertions perform it. They only require positive membership of returned sequences, not completeness: omissions can pass this row. The census guards completeness against fixed snapshots. The Node row supplies an independent external semantic check of those snapshots, so failure to find a unique mutant is not evidence that this role should be deleted. On the permitted menu, an edited table sequence necessarily changes the census digest, as all three attempts demonstrate. Lookup and data are the shared code; no exclusive Node-only production path was found.
The whole-package 90-second baseline was the largest cost. Go coverage cannot count generated variable-initializer data as distinct covered statements. The requested 15GB free disk threshold cannot be met on /tmp. No other ambiguity blocked the work. No test was deleted, rewritten or weakened; no main push or pull request.
