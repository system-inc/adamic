Legacy Node row defended by an incomplete-equivalence early return.
Unicode Node family rejects three closed-but-wrong tables, but the input pin catches each too.
All evidence is bounded; restored source passed the bounded baseline in 41.242 seconds.

Starting origin/main: bdb89962b178f618e2e6d34e9a7ea5c395089faf. Audit base: 60397548dd8a9494a7d2aaa874b7648b8e625607. Branch: test-defend/internal-unicodeproperties-canonicalize. Current inventory: 51 top-level tests, including all 34 Unicode range members. Added names: []; vanished names: []. Requested tests remain in canonicalize_test.go. The audit REPORT.md, report.json, code-under-test notes, plans, scope and matrix were read and retained.

CODE UNDER TEST and ORACLES
Legacy: CanonicalizeLegacy and LegacyEquivalents plus legacyFold, legacyClassCanon and legacyClass. The value pass reads legacyFold directly; equivalence scans call LegacyEquivalents -> CanonicalizeLegacy. ORACLE: unmodified Node toUpperCase and case-insensitive RegExp scans, comparing exact members and complete scan counts.
Unicode family: unicodeClassCanon and unicodeClass generated production tables. This family reads the tables directly, derives class/singleton scan inputs and compares exact members against unmodified Node iu/iv RegExp scans. It never calls CanonicalizeUnicode or UnicodeEquivalents. Those functions and unicodeFold are covered by the closure row, our comparator. ORACLE: unmodified Node, not our own snapshot. TestUnicodeNodeShardCoverage separately pins the generated scan inputs. No test, harness, oracle or generator was changed.
Production function inventory reached by these rows and the subsumer: CanonicalizeUnicode, UnicodeEquivalents, CanonicalizeLegacy, LegacyEquivalents. sort.Search is standard library. Six generated arrays provide their data. test-callers.txt retains the package-wide reference search.

BASELINES AND BOUNDS
Whole package timed out at 90.019 binary seconds during Unicode ranges; no ordinary assertion failure or skip was observed. The bounded baseline passed in 40.965 seconds. It includes every one of the 17 current non-range tests and the first complete 16-line shard of Range0. Its selector is matrix-selector.txt. All three direct test callers of the D1-mutated LegacyEquivalents function are included: examples, closure and LegacyNode. No other current test calls that function; the unrun Unicode-family members read Unicode data and cannot reach D1. For table mutations, the remaining 381 family shards have unknown outcomes. No package-wide unique kill is claimed from partial family table runs. Every matrix run completed below 90 seconds. No other packages were tested.

COVERAGE AND SEMANTIC DIFFERENCE
Isolated go test profiles use -coverpkg=./internal/unicodeproperties and are saved with their logs. Legacy and closure both cover 100% of CanonicalizeLegacy and LegacyEquivalents. Legacy has zero exclusive covered production lines. Unicode family shard has 0% production statements because static table reads are in the test's preparation, not executable production functions; it has no exclusive production lines either. Coverage-diff.json and three function summaries retain the exact result.
Shared-line semantic distinction: closure checks idempotence, inclusion of the input, sorted members and that every returned member canonicalizes to the same value. It does not prove that all matching members were returned. The examples assert LegacyEquivalents on uppercase S and unchanged long-s/Kelvin inputs, not lowercase s/a. LegacyNode scans every input against independent Node matches.
For the Unicode family, a coherent wrong partition can satisfy every closure invariant yet fail the external Node comparison. A changed generated input also fails the independent setup snapshot. Static-table changes are not visible as new Go statement coverage.

D1: RETURN EARLY IN LegacyEquivalents
After canon is computed, `if canon != codeUnit { return []uint16{codeUnit} }` returns only the input for a noncanonical member. This is the menu's conditional early return, not an empty-answer probe. Canonical representatives and all generated tables stay intact. D1 is standalone D1.diff, origin/main canonicalize.go:63. It passed go vet and compiled in go test.
Only TestCanonicalizeLegacyNode failed: canonicalize_test.go:196: BAD 0061 EXTRA 41 MISSING. Node found uppercase A while the mutated port returned only lowercase a. The value pass remained unchanged. TestEquivalentsAreClosed passed despite missing equivalences. This is a unique kill in the complete reached-function matrix and defends the row. Excluded family members cannot call the changed function; their timing/outcomes were not observed.
Passing rows in D1 (Range0 means its complete first shard):
- TestBinaryAliases
- TestCanonicalizeExamples
- TestCanonicalizeUnicodeNodeRange0
- TestEquivalentsAreClosed
- TestGeneralCategoryAliases
- TestKnownMembership
- TestNodeAgrees
- TestNodeStringProperties
- TestRejectedNames
- TestScriptAliasSet
- TestSetBoundaries
- TestStringPropertyCensus
- TestUnicodeNodeBatchOrderAndLimit
- TestUnicodeNodeShardCoverage
- TestUnicodeNodeShardPlantedFailure
- TestUnicodeNodeStridePlantedFailure
- TestVersionMatchesNode

D2, D3, D4: THREE UNICODE-FAMILY ATTEMPTS
Each is a small coherent generated-data mutation: change the uppercase fold target to itself and remove that uppercase member from the lowercase class. D2 splits B/b at canonicalize_tables.go:22 and :3023; D3 splits C/c at :23 and :3024; D4 splits D/d at :24 and :3025. This combines a constant change with dropping one member constant. The class arrays, ordering and containment remain internally consistent. No data are deleted from the external oracle.
For every attempt, closure and examples pass. The first family shard fails at canonicalize_test.go:327: D2 BAD CLASS iu 62 EXTRA 42 MISSING; D3 BAD CLASS iu 63 EXTRA 43 MISSING; D4 BAD CLASS iu 64 EXTRA 44 MISSING. iv comparisons fail as well. The external comparison proves the partitions wrong, independently of closure.
Every attempt also fails TestUnicodeNodeShardCoverage at shards_test.go:160 because the pre-split input digest changed. TestUnicodeNodeShardPlantedFailure fails too: its neighboring range now contains another real disagreement, breaking its single-failure fixture precondition. That witness side effect is not used as agreement strength or as the subsumer. The genuine input-pin assertion supplies the co-kill.
Thus the family is not uniquely defended after three honest targeted attempts. Its prior subsumer TestEquivalentsAreClosed passed all three. rows.json uses subsumed_by to name the observed current co-catcher, TestUnicodeNodeShardCoverage; the original audit metadata remains in the retained audit report.
No survivor exists across the bounded matrix. All diffs apply separately against the starting commit and all four go vet logs are clean. Production source is restored and the final bounded run passed in 41.242 seconds.

OWNER FINDINGS AND BRIEF COSTS
- Keep TestCanonicalizeLegacyNode: D1 proves an independent completeness check that both closure and examples miss.
- The Unicode family's name can suggest that it tests CanonicalizeUnicode, but it does not call that function. It checks the generated equivalence tables. This is a name/reach distinction, not a missing speed threshold. It compares exact members, not only counts. No cost-row mutation requirement applies to these names or assertions.
- No unique production mutant was found for the Unicode family, but closure demonstrably passed all three wrong-table variants that Node rejected. This evidence is not a deletion recommendation. The input snapshot and external semantic check establish different facts, even though these particular edits trigger both.
- The original audit excluded the input-pin and planted-failure rows. Including all current non-range rows changes the observed catcher. Their witness failures are precondition failures and are not counted as proof that a comparison works.
- A single inconsistent table edit would merely re-trigger closure. Two coordinated small data changes were necessary to test the external semantics against internally consistent but wrong tables. Both changes are in the fixed menu; no supplemental insertion or harness edit was used.
- The whole package is over the binary budget. Coverage for the Unicode family and its matrix are limited to its first complete shard. Exact outcomes for other family shards remain unknown. Static call references justify the complete reached-function set for D1, not execution claims about the rest.
- Before any work, df showed /tmp 2.4 GB free and /workspace 14 GB free. Removing the identified previous unit's /tmp/estree-defend scratch (196 MB) left /tmp 2.6 GB free; /workspace remained 14 GB. /tmp's total capacity is 8.8 GB, so 15 GB free there is impossible. Repository, tools and unidentified directories were not deleted. No test output reports disk exhaustion.
- Warm env.sh worked; setup skipped. npm ci added three packages in 430 ms. nproc=5. No native rebuilds apply to this Go package.
- Matrix command wall times sum to 164.542 seconds. Whole clean binary 90.019 s; bounded clean 40.965 s; restored clean 41.242 s. Vet/Go compilation is included in logs, and matrix timings include Go compilation overhead. Approximate session duration: 14 minutes including audit reading, disk cleanup, baseline, coverage and evidence work.

No tests were deleted, rewritten or weakened. No main push or PR. Standalone diffs, exact full commands, complete pass/fail lists, coverage profiles and old audit notes are preserved under the requested review path.
