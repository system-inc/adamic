Two gap rows defended; lexer agreement remains cannot-judge after two shared kills and a cooked third attempt.
All 72 current rows have a passing enabled baseline; D1/D2 each passed the other 71 rows.
Production sources and tests are restored; evidence belongs on test-defend/stage1-cohere-yaml-gaps.

```json
[
  {
    "test": "TestLexerGaps",
    "package": "stage1/cohere/yaml",
    "prior_verdict": "subsumed",
    "subsumed_by": [
      "TestStructuralPositionRefusal"
    ],
    "defense": "defended",
    "unique_mutant": "D1 internal/lower/object.go:1266",
    "attempts": [
      {
        "mutant": "D1",
        "file_line": "internal/lower/object.go:1266",
        "change": "change diagnostic constant",
        "rows_failed": [
          "TestLexerGaps"
        ]
      }
    ],
    "rows_passed": [
      "TestComposeMatchGo",
      "TestComposeMutants",
      "TestCSTMatchesGo",
      "TestCSTMutants",
      "TestFileDriver_Setup",
      "TestFormatterMatchesGo",
      "TestFileDriverUnion",
      "TestBundledParserDifference",
      "TestFileDriver_006",
      "TestFileDriver_007",
      "TestFileDriver_002",
      "TestFileDriver_005",
      "TestFileDriver_004",
      "TestFileDriver_000",
      "TestFileDriver_003",
      "TestFileDriver_001",
      "TestFormatterMutantsPlantedFailure",
      "TestProduct_YAMLFormatterMutant002Sources",
      "TestProduct_YAMLFormatterMutant004Sources",
      "TestProduct_YAMLFormatterMutantsOracle",
      "TestFormatterMutants_002",
      "TestProduct_YAMLFormatterMutant001Native",
      "TestProduct_YAMLFormatterMutant001Lowered",
      "TestProduct_YAMLFormatterMutant001Sources",
      "TestFormatterMutants_004",
      "TestFormatterMutants_003",
      "TestProduct_YAMLFormatterMutant003Sources",
      "TestProduct_YAMLFormatterMutant003Native",
      "TestProduct_YAMLFormatterMutant003Lowered",
      "TestProduct_YAMLFormatterMutant002Native",
      "TestProduct_YAMLFormatterMutant002Lowered",
      "TestProduct_YAMLFormatterMutant005Sources",
      "TestFormatterMutants_001",
      "TestProduct_YAMLFormatterMutant000Sources",
      "TestProduct_YAMLFormatterMutant005Native",
      "TestProduct_YAMLFormatterMutant005Lowered",
      "TestProduct_YAMLFormatterMutant004Lowered",
      "TestProduct_YAMLFormatterMutant004Native",
      "TestFormatterMutants_005",
      "TestProduct_YAMLFormatterMutant000Lowered",
      "TestProduct_YAMLFormatterMutant000Native",
      "TestFormatterMutants_000",
      "TestClosedStringPresenceGap",
      "TestClosedLexerGaps",
      "TestClosedValuePresenceGap",
      "TestLexerMatchesGo",
      "TestLexerMutants",
      "TestPropsMatchGo",
      "TestPropsMutants",
      "TestSharedSliceAppendMatchesNode",
      "TestScalarMutants",
      "TestStructuralPositionRefusal",
      "TestSchemaMatchesGo",
      "TestSchemaMutants",
      "TestSpeedCostProbes",
      "TestWidthsMatchGo",
      "TestProduct_YamlScalarsLowered",
      "TestScalarsMatchGoPlantedFailure",
      "TestScalarsMatchGo_006",
      "TestScalarsMatchGo_005",
      "TestScalarsMatchGo_007",
      "TestProduct_YamlScalarsGo",
      "TestScalarsMatchGo_002",
      "TestProduct_YamlScalarsNative",
      "TestScalarsMatchGoUnion",
      "TestScalarsMatchGo_004",
      "TestScalarsMatchGo_000",
      "TestScalarsMatchGo_003",
      "TestScalarsMatchGo_001",
      "TestUnistMatchesGo",
      "TestUnistMutants"
    ],
    "evidence": "ADAMIC_YAML_LIBRARY=/tmp/defend-yaml-gaps/library ADAMIC_BUILD_CACHE_DIR=/tmp/defend-yaml-gaps/cache/D1 timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/yaml/ -run '^(TestLexerGaps|TestStructuralPositionRefusal|TestClosedStringPresenceGap|TestClosedLexerGaps|TestClosedValuePresenceGap|TestLexerMatchesGo|TestLexerMutants|TestPropsMatchGo|TestPropsMutants|TestSharedSliceAppendMatchesNode|TestScalarMutants)$' > D1-batch3.log 2>&1; gaps_test.go:44: gap changed or closed: /workspace/adamic/stage1/cohere/yaml/gaps/multiplePush.ts:2:1: stage 0 can't lower push with unsupported arity yet; update GAPS.md and remove its workaround"
  },
  {
    "test": "TestStructuralPositionRefusal",
    "package": "stage1/cohere/yaml",
    "prior_verdict": "subsumed",
    "subsumed_by": [
      "TestLexerGaps"
    ],
    "defense": "defended",
    "unique_mutant": "D2 internal/lower/cycles.go:277",
    "attempts": [
      {
        "mutant": "D2",
        "file_line": "internal/lower/cycles.go:277",
        "change": "drop mutable-array Refused return and its now-unused target declaration",
        "rows_failed": [
          "TestStructuralPositionRefusal"
        ]
      }
    ],
    "rows_passed": [
      "TestComposeMatchGo",
      "TestComposeMutants",
      "TestCSTMatchesGo",
      "TestCSTMutants",
      "TestFileDriver_Setup",
      "TestFormatterMatchesGo",
      "TestFileDriverUnion",
      "TestBundledParserDifference",
      "TestFileDriver_006",
      "TestFileDriver_002",
      "TestFileDriver_007",
      "TestFileDriver_003",
      "TestFileDriver_004",
      "TestFileDriver_000",
      "TestFileDriver_001",
      "TestFileDriver_005",
      "TestFormatterMutantsPlantedFailure",
      "TestProduct_YAMLFormatterMutant004Sources",
      "TestProduct_YAMLFormatterMutant002Sources",
      "TestProduct_YAMLFormatterMutant003Sources",
      "TestProduct_YAMLFormatterMutantsOracle",
      "TestFormatterMutants_004",
      "TestFormatterMutants_002",
      "TestFormatterMutants_003",
      "TestProduct_YAMLFormatterMutant005Lowered",
      "TestProduct_YAMLFormatterMutant005Sources",
      "TestProduct_YAMLFormatterMutant004Native",
      "TestProduct_YAMLFormatterMutant004Lowered",
      "TestProduct_YAMLFormatterMutant001Native",
      "TestProduct_YAMLFormatterMutant001Lowered",
      "TestProduct_YAMLFormatterMutant001Sources",
      "TestFormatterMutants_001",
      "TestProduct_YAMLFormatterMutant005Native",
      "TestProduct_YAMLFormatterMutant000Sources",
      "TestProduct_YAMLFormatterMutant002Lowered",
      "TestProduct_YAMLFormatterMutant002Native",
      "TestProduct_YAMLFormatterMutant003Native",
      "TestProduct_YAMLFormatterMutant003Lowered",
      "TestFormatterMutants_005",
      "TestProduct_YAMLFormatterMutant000Lowered",
      "TestProduct_YAMLFormatterMutant000Native",
      "TestFormatterMutants_000",
      "TestClosedStringPresenceGap",
      "TestClosedLexerGaps",
      "TestClosedValuePresenceGap",
      "TestLexerMatchesGo",
      "TestLexerMutants",
      "TestPropsMatchGo",
      "TestPropsMutants",
      "TestSharedSliceAppendMatchesNode",
      "TestScalarMutants",
      "TestLexerGaps",
      "TestSchemaMatchesGo",
      "TestSchemaMutants",
      "TestSpeedCostProbes",
      "TestWidthsMatchGo",
      "TestProduct_YamlScalarsLowered",
      "TestScalarsMatchGoPlantedFailure",
      "TestScalarsMatchGo_002",
      "TestScalarsMatchGo_000",
      "TestScalarsMatchGo_005",
      "TestScalarsMatchGo_001",
      "TestProduct_YamlScalarsGo",
      "TestScalarsMatchGo_007",
      "TestProduct_YamlScalarsNative",
      "TestScalarsMatchGoUnion",
      "TestScalarsMatchGo_006",
      "TestScalarsMatchGo_004",
      "TestScalarsMatchGo_003",
      "TestUnistMatchesGo",
      "TestUnistMutants"
    ],
    "evidence": "ADAMIC_YAML_LIBRARY=/tmp/defend-yaml-gaps/library ADAMIC_BUILD_CACHE_DIR=/tmp/defend-yaml-gaps/cache/D2 timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/yaml/ -run '^(TestLexerGaps|TestStructuralPositionRefusal|TestClosedStringPresenceGap|TestClosedLexerGaps|TestClosedValuePresenceGap|TestLexerMatchesGo|TestLexerMutants|TestPropsMatchGo|TestPropsMutants|TestSharedSliceAppendMatchesNode|TestScalarMutants)$' > D2-batch3.log 2>&1; gaps_test.go:71: refusal changed or closed: <nil>"
  },
  {
    "test": "TestLexerMatchesGo",
    "package": "stage1/cohere/yaml",
    "prior_verdict": "subsumed",
    "subsumed_by": [
      "TestPropsMatchGo"
    ],
    "defense": "cannot-judge",
    "unique_mutant": null,
    "attempts": [
      {
        "mutant": "D3",
        "file_line": "stage1/cohere/yaml/lexer.ts:286",
        "change": "drop empty-token emission statement",
        "rows_failed": [
          "TestCSTMatchesGo",
          "TestLexerMatchesGo",
          "TestPropsMatchGo"
        ]
      },
      {
        "mutant": "D4",
        "file_line": "stage1/cohere/yaml/lexer.ts:365",
        "change": "off-by-one explicit block indentation",
        "rows_failed": [
          "TestCSTMatchesGo",
          "TestLexerMatchesGo",
          "TestPropsMatchGo"
        ]
      },
      {
        "mutant": "D5",
        "file_line": "stage1/cohere/yaml/lexer.ts:160",
        "change": "drop CRLF line-bound adjustment",
        "rows_failed": [],
        "status": "over budget; no completed row results",
        "rows_unknown": [
          "TestLexerGaps",
          "TestStructuralPositionRefusal",
          "TestLexerMatchesGo",
          "TestPropsMatchGo",
          "TestCSTMatchesGo"
        ]
      }
    ],
    "bounded": true,
    "matrix_rows": [
      "TestLexerGaps",
      "TestStructuralPositionRefusal",
      "TestLexerMatchesGo",
      "TestPropsMatchGo",
      "TestCSTMatchesGo"
    ],
    "evidence": "ADAMIC_YAML_LIBRARY=/tmp/defend-yaml-gaps/library ADAMIC_BUILD_CACHE_DIR=/tmp/defend-yaml-gaps/cache/D4 timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/yaml/ -run '^(TestLexerGaps|TestStructuralPositionRefusal|TestLexerMatchesGo|TestPropsMatchGo|TestCSTMatchesGo)$' > D4.log 2>&1; lexer_test.go:210: native ASan/UBSan/LSan: byte 860559: got \"0062000a\\n=000a\\ncase 60\\n=0002\\n=001f\\n=0061\\n=003a\\n=0020\\n=003e0032002d\\n=000a\\n=001f\\n=\\n=00200020\\n=001f\\n=0062\\n=000a\\ncase 61\\n=0002\\n=001f\\n=0061\\n=003a\\n=0020\\n=003e0032002d\", want \"0062000a\\n=000a\\ncase 60\\n=0002\\n=001f\\n=0061\\n=003a\\n=0020\\n=003e0032002d\\n=000a\\n=001f\\n=002000200062000a\\ncase 61\\n=0002\\n=001f\\n=0061\\n=003a\\n=0020\\n=003e0032002d\\n=000a\\n=001f\"",
    "reason": "D3 and D4 are shared semantic kills. D5 exhausted 90 seconds in CST before this row ran, and its narrowed lexer replay also did not complete within the binary budget. D5 row results remain unknown; three completed attempts were not obtained."
  }
]
```

# YAML gaps defense notes

Starting commit: 6eef5ae586af0354d3032d36535e54cefb2da43a. Audit base: 3bf0a5d9e74d197a38f61982e43b2d791f17b563. All 72 current top-level functions match audit discovery. The three requested rows remain in gaps_test.go and lexer_test.go.

## Code and oracle

TestLexerGaps calls Adamic Lower on gaps/multiplePush.ts. The code under test is arrayMethodArguments in internal/lower/object.go and its NotYet diagnostic. Node runs the original fixture and must print 2; self-written pins require NotYet and the push-arity diagnostic substring. It does not compare YAML lexer output or parse GAPS.md.

TestStructuralPositionRefusal calls Adamic Lower on gaps/structuralPosition.ts. The code under test is cycleFinder.slotsOf and the structural reachability/fresh-write proof that leads to its Refused return. Node must print 1; self-written pins require Refused and the Span[] array diagnostic substring. Span has start/end fields; Properties has those fields plus errors, so structural compatibility admits a back-reference even though the particular fixture does not construct a runtime cycle.

TestLexerMatchesGo checks the TypeScript Lexer port compiled by Adamic and run under Node, plus Adamic's emitted JavaScript. Live Go cohere and yaml@2.9.0 supply exact token-stream bytes. TestPropsMatchGo uses the same complete/chunked input corpus but compares resolved properties after CST parsing. These are different projections, not Node/native twin rows. Each row already combines several executors.

## Coverage

Each requested row and its subsumer was run independently with -count=1, anchored -run, -coverprofile and -coverpkg covering internal/lower and internal/native. The row-exclusive covered Go blocks number 15 for LexerGaps against StructuralPositionRefusal, 760 in the opposite direction, and two for LexerMatchesGo against PropsMatchGo. The push diagnostic block is exclusive in the first pair. Most structural exclusivity reflects LexerGaps stopping before the cycle pass.

Go coverage does not instrument TypeScript or C. Separate NODE_V8_COVERAGE runs provide actual source-port coverage for LexerMatchesGo and PropsMatchGo. Both reach 34 Lexer script functions; neither the range comparison nor the line-start approximation identifies a lexer-only covered range. Raw V8 reports and the extracted lexer scripts are retained. Shared coverage permits semantic defenses; the attempted lexer changes exercise empty raw tokens, explicit block indentation and CRLF token boundaries.

## Baseline and scope

npm ci ran in stage3/api. Tools were already usable; setup was skipped, nproc is 5. External npm oracles are pinned to yaml 2.9.0, prettier 3.9.6 and yaml-unist-parser 3.2.0.

The whole-package baseline exhausted the test binary's 90-second budget. Bounded baseline batches then exposed a missing yaml-unist-parser import in TestUnistMatchesGo. This was a dependency setup failure, not a production mutation. Installing the latest parser initially upgraded yaml to 2.9.1; the port's comments identify parser 3.2.0, so all three packages were then explicitly pinned. The affected baseline was rerun; every current row has passing enabled baseline evidence before any mutant was applied. The original oversized last batch also exhausted 90 seconds and was split into scalar/schema/width, Unist agreement, and Unist witness batches.

D1 and D2 replay every current top-level function across seven bounded batches, with a distinct ADAMIC_BUILD_CACHE_DIR for each mutant. D3-D5 initially replay five reached rows: LexerGaps, StructuralPositionRefusal, LexerMatchesGo, PropsMatchGo and CSTMatchesGo. Their outside rows are unknown unless separately replayed. A shared catch disproves exclusivity for that attempt without needing an absence claim outside this set. These lexer results are a limited mutation-set defense attempt, not a general redundancy proof.

All mutation files are standalone diffs against the starting commit. D1 changes only the push diagnostic constant. D2 drops the mutable-array Refused return together with the declaration used solely by that return. D3-D5 change lexer.ts itself, not the comparison driver, oracle, harness or test. Every source mutation is reversed after its run. Tests are untouched.

## Brief friction and owner findings

1. The enabled-library skip hints mention only yaml and prettier, while Unist also imports yaml-unist-parser. The missing dependency caused a real baseline failure and required pin discovery from port comments. All failed setup logs are retained; no defense runs took place on that red baseline.
2. A 90-second whole-package limit cannot cover this package's serial agreement and mutant witnesses. Batching preserves the limit and discovers current rows beyond the audit slice. TestMain additionally performs a file-driver setup subprocess before the parent test timer.
3. Go -coverpkg cannot measure the TypeScript port. V8 source coverage supplements it, with an explicit line-range approximation rather than a claim of exact statement coverage.
4. The twin instruction allows defense=twin although the JSON enum omits it. None of the requested rows is an executor twin, so no schema exception is needed here.
5. The raw lexer and property rows share lexical inputs and coverage but assert different answer representations. Comparing covered lines alone cannot settle their relative value. D3 and D4 are shared completed output disagreements. D5 exhausts the binary budget, first in CST and then in a narrowed raw-lexer replay; these are unknown row results, not zero kills. Two native children outlive their timed-out parents and require session-owned process-group cleanup.
6. TestLexerGaps pins a diagnostic, rather than reading GAPS.md. A diagnostic-only exclusive mutant defends that contract, not the semantics of arity rejection. TestStructuralPositionRefusal checks both refusal class and diagnostic; removal of the refusal directly tests its claimed safety guard.
7. TestLexerMatchesGo's name promises agreement with Go, and its assertions check exact output bytes against live Go on complete and chunked inputs. There is no missing performance threshold or other unasserted promise in this row. Failure to find an exclusive mutant in three attempts is not a deletion recommendation.

The third lexer defense remains cannot-judge if its narrowed CRLF replay also exhausts the budget. Two shared catches do not meet the requested three-completed-attempt criterion for not defended. The row stays pending rather than being marked redundant. Its name does not overpromise: exact live-Go token agreement is asserted.

Coverage profiles are gzip-compressed; decompress before using go tool cover. matrix.json lists every observed pass and every unknown row. runs.json files retain commands, wall time and status. The overall work took about 35 minutes, beyond the nominal port budget, because of enabled baseline dependency correction, full-package replays and two cooked CRLF runs. Dependency installations were not separately timed.
