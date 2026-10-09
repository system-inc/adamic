All five scoped tests exist at the fetched starting commit.
Two bounded sacred rows, one subsumed row, one setup-check and one witness.
Thirteen production mutants: twelve caught, M02 survived.
Full baseline cooked at 90s; narrowed baselines passed; nproc=5.
Evidence saved for central replay, with all production and harness sources restored.

```json
[
  {
    "test": "TestVersionMatchesNode",
    "package": "internal/unicodeproperties",
    "file": "internal/unicodeproperties/node_test.go",
    "seconds": 0.041,
    "oracle": "Node process.versions.unicode compared with NodeUnicodeVersion, plus a self-written Node 24 pin. Version patch metadata is not compared: M02 survived.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M01"
    ],
    "unique_kills": [
      "M01"
    ],
    "last_proven_fail": "M01: node_test.go:150: Node reports Unicode 17.0, tables are 16.0 (17.0.0): rerun go generate",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 15,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestVersionMatchesNode",
      "TestNodeAgrees",
      "TestNodeStringProperties",
      "TestBinaryAliases",
      "TestGeneralCategoryAliases",
      "TestScriptAliasSet",
      "TestRejectedNames",
      "TestStringPropertyCensus",
      "TestSetBoundaries",
      "TestKnownMembership"
    ],
    "evidence": "python3 review/test-audit/internal-unicodeproperties-node/run.py > review/test-audit/internal-unicodeproperties-node/runner.log 2>&1; M01.log: node_test.go:150: Node reports Unicode 17.0, tables are 16.0 (17.0.0): rerun go generate"
  },
  {
    "test": "TestNodeAgrees",
    "package": "internal/unicodeproperties",
    "file": "internal/unicodeproperties/node_test.go",
    "seconds": 11.133,
    "oracle": "Node RegExp u membership over every code point; compares Lookup ranges, not the Contains implementation.",
    "oracle_kind": "external-run",
    "kills": [
      "M03",
      "M04",
      "M06",
      "M07",
      "M08",
      "M09"
    ],
    "unique_kills": [
      "M07"
    ],
    "last_proven_fail": "M07: node_test.go:267: \\p{ASCII}: 1 code points disagree, first U+007F",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 15,
    "probe_kills": [
      "P01"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestVersionMatchesNode",
      "TestNodeAgrees",
      "TestNodeStringProperties",
      "TestBinaryAliases",
      "TestGeneralCategoryAliases",
      "TestScriptAliasSet",
      "TestRejectedNames",
      "TestStringPropertyCensus",
      "TestSetBoundaries",
      "TestKnownMembership"
    ],
    "evidence": "python3 review/test-audit/internal-unicodeproperties-node/run.py > review/test-audit/internal-unicodeproperties-node/runner.log 2>&1; M07.log: node_test.go:267: \\p{ASCII}: 1 code points disagree, first U+007F"
  },
  {
    "test": "TestNodeStringProperties",
    "package": "internal/unicodeproperties",
    "file": "internal/unicodeproperties/node_test.go",
    "seconds": 0.152,
    "oracle": "Node RegExp v accepts every returned sequence. Positive membership only: M12 omitted sequences and this row passed; the census caught it.",
    "oracle_kind": "external-run",
    "kills": [
      "M05",
      "M13"
    ],
    "unique_kills": [],
    "last_proven_fail": "M13: node_test.go:429: \\p{Basic_Emoji} does not match \"A\" (node said 0)",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestStringPropertyCensus"
    ],
    "mutants_in_matrix": 15,
    "probe_kills": [
      "P01"
    ],
    "subsumer_seconds": 0.003,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestVersionMatchesNode",
      "TestNodeAgrees",
      "TestNodeStringProperties",
      "TestBinaryAliases",
      "TestGeneralCategoryAliases",
      "TestScriptAliasSet",
      "TestRejectedNames",
      "TestStringPropertyCensus",
      "TestSetBoundaries",
      "TestKnownMembership"
    ],
    "evidence": "python3 review/test-audit/internal-unicodeproperties-node/run.py > review/test-audit/internal-unicodeproperties-node/runner.log 2>&1; M13.log: node_test.go:429: \\p{Basic_Emoji} does not match \"A\" (node said 0)",
    "subsumption_mutants": 2
  },
  {
    "test": "TestUnicodeNodeShardCoverage",
    "package": "internal/unicodeproperties",
    "file": "internal/unicodeproperties/shards_test.go",
    "seconds": 0.607,
    "oracle": "Self-written ordered-input digest and census; exact shard and top-level ownership checks.",
    "oracle_kind": "self",
    "kills": [
      "S01"
    ],
    "unique_kills": [],
    "last_proven_fail": "S01: shards_test.go:192: shard 185 owned 0 times, want 1",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 15,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestUnicodeNodeShardCoverage"
    ],
    "evidence": "python3 review/test-audit/internal-unicodeproperties-node/run.py > review/test-audit/internal-unicodeproperties-node/runner.log 2>&1; S01.log: shards_test.go:192: shard 185 owned 0 times, want 1"
  },
  {
    "test": "TestUnicodeNodeShardPlantedFailure",
    "package": "internal/unicodeproperties",
    "file": "internal/unicodeproperties/shards_test.go",
    "seconds": 17.444,
    "oracle": "Node differential plus self-written required BAD text, child failure location and neighboring pass controls.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "W01"
    ],
    "unique_kills": [],
    "last_proven_fail": "W01: shards_test.go:256: plant did not fail only its Node shard: <nil> === RUN   TestCanonicalizeUnicodeNodeRange0",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 15,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestUnicodeNodeShardPlantedFailure",
      "TestUnicodeNodeStridePlantedFailure"
    ],
    "evidence": "python3 review/test-audit/internal-unicodeproperties-node/run.py > review/test-audit/internal-unicodeproperties-node/runner.log 2>&1; W01.log: shards_test.go:256: plant did not fail only its Node shard: <nil> === RUN   TestCanonicalizeUnicodeNodeRange0"
  }
]
```

| ID | Starting file:line | Change | Failed rows |
|---|---|---|---|
| M01 | internal/unicodeproperties/unicodeproperties.go:34 | `NodeUnicodeVersion = "17.0"` -> `NodeUnicodeVersion = "16.0"` | TestVersionMatchesNode |
| M02 | internal/unicodeproperties/unicodeproperties.go:33 | `Version            = "17.0.0"` -> `Version            = "17.0.1"` |  |
| M03 | internal/unicodeproperties/unicodeproperties.go:160 | `case "Script_Extensions", "scx":
			table = scriptExtensionNames` -> `case "Script_Extensions", "scx":
			table = scriptNames` | TestKnownMembership, TestNodeAgrees, TestScriptAliasSet |
| M04 | internal/unicodeproperties/unicodeproperties.go:169 | `codePoint(generalCategoryNames, expression)` -> `codePoint(binaryNames, expression)` | TestGeneralCategoryAliases, TestKnownMembership, TestNodeAgrees |
| M05 | internal/unicodeproperties/unicodeproperties.go:175 | `if unicodeSets {` -> `if !unicodeSets {` | TestKnownMembership, TestNodeStringProperties, TestRejectedNames, TestStringPropertyCensus |
| M06 | internal/unicodeproperties/unicodeproperties.go:188 | `return Property{Kind: KindCodePoints, Name: e.name` -> `return Property{Kind: KindStrings, Name: e.name` | TestBinaryAliases, TestKnownMembership, TestNodeAgrees, TestRejectedNames |
| M07 | internal/unicodeproperties/tables.go:23 | `{ // 0 ASCII
		Ranges: []Range{
			{0x0000, 0x007F}` -> `{ // 0 ASCII
		Ranges: []Range{
			{0x0000, 0x007E}` | TestNodeAgrees |
| M08 | internal/unicodeproperties/unicodeproperties.go:204 | `if equals > 1 || i == 0` -> `if equals > 0 || i == 0` | TestGeneralCategoryAliases, TestKnownMembership, TestNodeAgrees, TestScriptAliasSet |
| M09 | internal/unicodeproperties/unicodeproperties.go:218 | `return expression[:i], expression[i+1:], true` -> `return expression[:i], expression[i:], true` | TestGeneralCategoryAliases, TestKnownMembership, TestNodeAgrees, TestScriptAliasSet |
| M10 | internal/unicodeproperties/unicodeproperties.go:60 | `if ranges[mid].End < cp {` -> `if ranges[mid].End <= cp {` | TestKnownMembership, TestSetBoundaries |
| M11 | internal/unicodeproperties/unicodeproperties.go:66 | `ranges[lo].Start <= cp` -> `ranges[lo].Start < cp` | TestKnownMembership, TestSetBoundaries |
| M12 | internal/unicodeproperties/unicodeproperties.go:177 | `Sequences: s.sequences}` -> `Sequences: s.sequences[:len(s.sequences)-1]}` | TestStringPropertyCensus |
| M13 | internal/unicodeproperties/tables.go:25712 | `var sequences_Basic_Emoji = []string{
	"©️",` -> `var sequences_Basic_Emoji = []string{
	"A",` | TestKnownMembership, TestNodeStringProperties, TestStringPropertyCensus |
| S01 | internal/unicodeproperties/shards_test.go:35 | `{"TestCanonicalizeUnicodeNodeRange0", 0, 186}` -> `{"TestCanonicalizeUnicodeNodeRange0", 0, 185}` | TestUnicodeNodeShardCoverage |
| W01 | internal/unicodeproperties/canonicalize_test.go:809 | `for index, problem := range disagreements {` -> `for index, problem := range disagreements[:0] {` | TestUnicodeNodeShardPlantedFailure, TestUnicodeNodeStridePlantedFailure |

Survivor M02: `Version=17.0.0 NodeUnicodeVersion=17.0` became `Version=17.0.1 NodeUnicodeVersion=17.0`, in survivor-clean.log and survivor-M02.log.

Starting commit was fetched origin/main 60397548dd8a9494a7d2aaa874b7648b8e625607, not historical 8de93800f4. All five scoped names exist in their cited files.

The complete package baseline cooked at 90.018 binary seconds (90.219 command), with no earlier assertion failure. Scoped clean baseline passed at 27.102 binary seconds (27.421 command). The ten-row production caller matrix had its own green baseline (13.367 binary seconds). Every production column completed all ten selected rows; none cooked, panicked or skipped. Remaining package rows are unknown, so all uniqueness and subsumption claims are bounded.

The full package's 34 expensive canonicalize range wrappers form a family, but they do not call the mutated Lookup, property sets or version metadata. The saved git-grep function inventory motivates their exclusion. Setup and witness rows use their own submatrices, not production precondition failures. Shard coverage has additional census, digest and construction assertions and is its own setup row. The stride planted witness runs the same plant helper with additional stride-specific BAD count and extra-folded-member assertions; it was tested with W01 as an additional witness, without merging it into the five scoped rows.

The fixed menu has thirteen production faults plus one setup construction fault and one witness harness weakening. M12 makes the implicit sequence upper bound len-1, not a new inserted statement. S01 is the allowed setup construction mutation. W01 skips Go reporting of disagreements; Node scripts, scanner self-checks and oracle answers were unchanged. Both planted witnesses failed when their child tests accepted the still-reported Node BAD answer.

M02 changes the public UCD patch-version metadata while preserving the Node Unicode major/minor pin. Its observed before/after output proves a real metadata change, not a membership or native-execution change. It survived every row in the bounded production matrix. M12 survived the Node string row but was caught by the string census, so it is not a matrix survivor. M10/M11 changed Contains and were caught by boundary/membership tests while the raw-range Node comparison passed.

Only the two Lookup-calling scoped rows received P01 (Lookup returns Property{},false). Both failed it. The metadata row calls no production function to empty; setup/witness rows were judged under their separate rules. Their vacuity is null, not inferred. P01 is not a mutant and contributes no uniqueness or subsumption.

The string row's subsumption rests on two observed mutants (M05 and M13). It is a hint, not a deletion recommendation. The subsumer median is 0.003 seconds. No named row was over 60 seconds.

Test binary seconds are package-line medians of three independent count=1 commands, rather than the parallel parent's often-zero PASS duration. These Node tests have no oracle-result cache. No rows in the scoped baselines or the bounded production matrix skipped. node scripts load only built-in fs, so no additional node_modules install was needed beyond the requested stage3/api npm ci.

Building the switched package was included in the first matrix command, not separately timed. Each standalone production diff was vetted separately as required; a later restored warm binary build was measured. The whole-package timeout and three runs of the planted child compilation were the main costs. No other packages, repo-wide replay, complete gate or exhaustive generated-table mutations were run. Whole-source generated tables were read at their mutated sites; the entire generated data corpus was not manually reviewed. README and doctrine files had no diff from the previously read versions.

Every standalone diff applies at the starting commit and passes go vet ./internal/unicodeproperties/. Production diffs have no selector; P01 retains a probe selector. All production and harness sources were restored. No main push or pull request was made.

Measured totals:
```json
{
  "runs.jsonl": 249.57795269100097,
  "timing-commands.jsonl": 89.69327187799354,
  "vet-times.jsonl": 4.950131771995075,
  "production_matrix_binary_seconds": 183.764,
  "timings_binary_seconds": 86.598,
  "finished_utc": "2026-10-09T11:41:27.007378+00:00"
}
```
Additional setup, baseline and build records are in their timing JSON files.
