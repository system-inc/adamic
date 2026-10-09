Keep all three rows: each catches its own production mutant uniquely in the reached-caller matrix.
TestKnownMembership and the other seven callers pass each target mutant.
Starting main bdb89962b178f618e2e6d34e9a7ea5c395089faf; source/tests/oracles restored and unchanged.

CODE UNDER TEST: Lookup/codePoint and generated scriptNames/stringProperties, ContainsSequence for sequence lookup
ORACLE: Unicode 17 and ECMAScript exact name rules plus self-written metadata invariants and stored census snapshots; independent Node membership comparisons run unchanged.

Audit evidence read first from ea8fffe6df912d7586cd423328db68a3bb8f9ca6: report.md and rows.json, with the fixed menu and caller evidence in those reports. Old audit base was fc31f9f15f95d841d4cd230a37ba55852866f2c8. Current go test -list enumerated 51 top-level functions, unchanged from the audit: no added or vanished names. All three target bodies and TestKnownMembership are in alias_test.go and remain distinct rows.

The clean whole-package baseline cooked at 90.012 binary seconds during canonicalization scans, with no prior test failure. All nine current property-lookup callers passed the bounded baseline (8.316 wall seconds). caller-search.log shows the only direct Lookup and alias/string-table callers in _test.go are alias_test.go and node_test.go. Canonicalization tests use separate fold/class tables and functions, not the mutated lookup or alias/sequence-table entries. We ran every current reached row including both Node comparisons. The 42 excluded canonicalization/version/shard-construction rows were not replayed per mutant; their results remain unknown. No whole-package uniqueness claim rests on an unrun row.

Matrix rows: TestBinaryAliases, TestGeneralCategoryAliases, TestScriptAliasSet, TestRejectedNames, TestStringPropertyCensus, TestSetBoundaries, TestKnownMembership, TestNodeAgrees, TestNodeStringProperties.

Per-test coverage used -coverpkg=github.com/system-inc/adamic/internal/unicodeproperties for each target and TestKnownMembership, individually. Four profiles and commands are saved. TestScriptAliasSet has only a final failed-lookup return exclusive block; its D1 defense instead uses shared code with Adlam metadata beyond the subsumer inputs. TestRejectedNames has exclusive lexical-rejection and invalid-property branches including unicodeproperties.go:150. TestStringPropertyCensus has no exclusive instrumented blocks: its semantic difference is completeness of all seven sequence tables beyond two membership examples. Generated table initializers are not a useful instrumented function-line coverage measure, so input and assertion differences are explicit.

## D1 TestScriptAliasSet
internal/unicodeproperties/tables.go:25019; change a constant.
"Adlam":                  {name: "Script", value: "Adlam", set: 91} -> "Adlam":                  {name: "Script", value: "", set: 91}
Adlam canonical Value must be nonempty. KnownMembership samples different scripts; NodeAgrees only checks ranges, not metadata.
Observed failure: alias_test.go:230: script Adlam: sc true scx true same set false
Passing rows: TestBinaryAliases, TestGeneralCategoryAliases, TestRejectedNames, TestStringPropertyCensus, TestSetBoundaries, TestKnownMembership, TestNodeAgrees, TestNodeStringProperties.
Command: ADAMIC_BUILD_CACHE_DIR=/tmp/defend-unicode-alias/cache/D1 timeout 120 go test -json -count=1 -timeout 90s ./internal/unicodeproperties/ -run '^(TestBinaryAliases|TestGeneralCategoryAliases|TestScriptAliasSet|TestRejectedNames|TestStringPropertyCensus|TestSetBoundaries|TestKnownMembership|TestNodeAgrees|TestNodeStringProperties)$' > D1.log 2>&1

## D2 TestRejectedNames
internal/unicodeproperties/unicodeproperties.go:150; change a constant.
if !expressionOK(expression) {
		return Property{}, false -> if !expressionOK(expression) {
		return Property{}, true
Lexically invalid empty, equals-only and space-suffixed expressions are rejected; KnownMembership only uses valid spellings.
Observed failure: alias_test.go:262: accepted "" without the v flag
Passing rows: TestBinaryAliases, TestGeneralCategoryAliases, TestScriptAliasSet, TestStringPropertyCensus, TestSetBoundaries, TestKnownMembership, TestNodeAgrees, TestNodeStringProperties.
Command: ADAMIC_BUILD_CACHE_DIR=/tmp/defend-unicode-alias/cache/D2 timeout 120 go test -json -count=1 -timeout 90s ./internal/unicodeproperties/ -run '^(TestBinaryAliases|TestGeneralCategoryAliases|TestScriptAliasSet|TestRejectedNames|TestStringPropertyCensus|TestSetBoundaries|TestKnownMembership|TestNodeAgrees|TestNodeStringProperties)$' > D2.log 2>&1

## D3 TestStringPropertyCensus
internal/unicodeproperties/tables.go:33643; off-by-one an implicit slice bound.
name: "RGI_Emoji_Tag_Sequence", sequences: sequences_RGI_Emoji_Tag_Sequence} -> name: "RGI_Emoji_Tag_Sequence", sequences: sequences_RGI_Emoji_Tag_Sequence[1:]}
Drop one valid tag sequence. Census pins completeness while NodeStringProperties only validates sequences still present; KnownMembership samples Basic_Emoji and keycaps.
Observed failure: alias_test.go:312: RGI_Emoji_Tag_Sequence: 2 sequences, sha256 da8091ff33578db62e71203ad1eddc2ae9d78c632ba202b2d8e696c4b263bafd
Passing rows: TestBinaryAliases, TestGeneralCategoryAliases, TestScriptAliasSet, TestRejectedNames, TestSetBoundaries, TestKnownMembership, TestNodeAgrees, TestNodeStringProperties.
Command: ADAMIC_BUILD_CACHE_DIR=/tmp/defend-unicode-alias/cache/D3 timeout 120 go test -json -count=1 -timeout 90s ./internal/unicodeproperties/ -run '^(TestBinaryAliases|TestGeneralCategoryAliases|TestScriptAliasSet|TestRejectedNames|TestStringPropertyCensus|TestSetBoundaries|TestKnownMembership|TestNodeAgrees|TestNodeStringProperties)$' > D3.log 2>&1

D3 is an off-by-one implicit start bound: the complete sequence slice starts at zero; [1:] starts at one and drops the England tag sequence. No new statement, test, harness or oracle edit is part of any mutant. All three standalone diffs apply to this starting main and pass go vet ./internal/unicodeproperties/. Each runtime matrix gets its own ADAMIC_BUILD_CACHE_DIR, although these are pure Go products and no Adamic-native cache is used. No panic or timeout occurred in a mutant matrix. No empty-answer probes or compiler failures are counted as defenses.

Direct production witnesses are in witnesses.json: D1 changes sc=Adlam Value from Adlam to empty while retaining membership/set; D2 changes empty-name ok from false to true; D3 changes tag-sequence count from three to two. The disposable Go caller was removed and its source saved as witness.go.txt. Unicode 17 PropertyValueAliases independently says Adlm=Adlam. Node 24.19.0 Unicode 17 rejected all four D2 malformed strings, and accepted the dropped England tag sequence. Those independent session probes do not change the target rows into external-run tests. The Census oracle remains a self snapshot. NodeStringProperties passing D3 proves its one-way validity check cannot detect omission, whereas Census can.

## Brief friction and costs
- Disk first: /tmp initially had 8.8 GB free and /workspace 20 GB free. Because /tmp was below 15 GB, the earlier /tmp/defend-estree-scalars scratch and cache directory was deleted. /tmp still has 8.8 GB free because that is its total capacity. Repo and tools were untouched. No disk-related baseline failure was counted.
- The full baseline exceeds the 90-second binary budget because of unrelated canonicalization scans. The permitted caller-based narrowing covers all current property callers; results outside it are explicitly unknown.
- Generated-table metadata and omission can differ on shared executable lines. Coverage alone would miss both defenses; the brief explicitly permits semantic differences. D3 uses the brief's implicit-bound rule.
- D1's existing failure message does not print the empty Value that triggers the assertion. The direct production witness supplies that missing detail; no message or test was rewritten.
- Prior subsumption was evidence over the audit's fixed mutations, not a logical implication. Each row has a new unique production catch against its named subsumer.
- No target is a cost/performance row or a separate-executor twin. No cost or compiler-twin attempt was needed.
- All three were defended, so there is no not-defended name/assertion mismatch finding. Their actual assertions cover script alias metadata, exact-name rejection and sequence census, respectively.

## Timing
{
  "setup_seconds": 0,
  "nproc": 5,
  "npm_ci_separately_timed": false,
  "whole_baseline_binary_seconds": 90.012,
  "bounded_baseline_wall_seconds": 8.316222384000866,
  "coverage_wall_seconds": 1.0862853080034256,
  "mutant_vet_wall_seconds": 0.5548987919974024,
  "mutant_matrix_wall_seconds": 25.32097763600177,
  "restored_baseline_wall_seconds": 8.073177562000637
}
Warm env.sh worked; setup skipped. npm ci in stage3/api succeeded but was not separately timed. No package other than internal/unicodeproperties was tested; the disposable direct caller only queried this package. One successful aimed mutant per row sufficed under the up-to-three limit. Final restored bounded baseline passed. No tests deleted/rewritten/weakened, no main push, no PR.
