TestWholeGeneratedAgrees is defended by one package-unique production mutant.
TestProduct_ParserOracle is cannot-judge under the forbidden harness/oracle mutation rule.
All 100 current package rows have definitive results, including enabled performance rows.

[
  {
    "test": "TestProduct_ParserOracle",
    "package": "stage1/typescript/parser",
    "prior_verdict": "untrue",
    "subsumed_by": [],
    "defense": "cannot-judge",
    "unique_mutant": null,
    "attempts": [],
    "evidence": "ADAMIC_BUILD_CACHE_DIR=/tmp/parser-defense/product-coverage/TestProduct_ParserOracle timeout 120 go test -json -count=1 -timeout 90s -coverpkg=github.com/system-inc/adamic/internal/buildcache -coverprofile=review/test-defend/stage1-typescript-parser-setup_deadline_products/TestProduct_ParserOracle.cold.cover ./stage1/typescript/parser/ -run '^TestProduct_ParserOracle$'; passed; 57 covered blocks, zero exclusive versus TestProduct_ExpressionsOracle. No allowed production mutation target: _test.go preparation recipe and Go oracle are prohibited."
  },
  {
    "test": "TestWholeGeneratedAgrees",
    "package": "stage1/typescript/parser",
    "prior_verdict": "subsumed",
    "subsumed_by": [
      "TestEveryTypeNodeKindAgrees"
    ],
    "defense": "defended",
    "unique_mutant": "D1 stage1/typescript/parser/statements.ts:748",
    "attempts": [
      {
        "mutant": "D1",
        "file_line": "stage1/typescript/parser/statements.ts:748",
        "change": "WithStatement constant -> WhileStatement in Statements.statement's WithKeyword branch",
        "rows_failed": [
          "TestWholeGeneratedAgrees"
        ]
      }
    ],
    "evidence": "python3 review/test-defend/stage1-typescript-parser-setup_deadline_products/run-matrix.py D1 (15 bounded timeout/go-test groups); whole_generated_test.go:66: Node: case 12 line 403: port \"1 WhileStatement 0 13 0 0 -1 0 0\\t\\t\\t\\t\", Go \"1 WithStatement 0 13 0 0 -1 0 0\\t\\t\\t\\t\""
  }
]

## Evidence

D1.diff applies cleanly to the recorded origin/main base. It changes only statements.ts:748, WithStatement to WhileStatement. The changed port compiled with the actual sanitized native build pipeline in 22.843 seconds. Native and Node witness logs show the changed kind. Only TestWholeGeneratedAgrees failed; all 99 other package rows passed. The complete passed list is rows-passed.txt, the per-row results are matrix.json, and exact bounded commands are D1-runs.json. No production source change remains and no test or oracle source was edited.

# Constraints, ambiguities, and costs

1. Go coverage cannot instrument the TypeScript port. Host Go profiles and actual Node V8 profiles were both produced, and their distinction is explicit in CODE-AND-ORACLE.md. V8 establishes that the WithKeyword branch is exclusive to the generated row relative to its named subsumer.
2. The earlier audit treated _test.go preparation recipes as construction code under test and mutated them. This defender brief prohibits harness, test, and oracle mutations. For the oracle-product wrapper, that leaves no permitted production target. No prohibited edit was substituted for an honest defense. The result is cannot-judge, not a deletion recommendation.
3. TestProduct_ParserOracle's name suggests an oracle product is available. Its assertion checks only successful recipe completion. It does not check that the returned artifact exists or can run. TestProduct_ExpressionsOracle invokes precisely the same recipe and also discards its path. Cold per-row coverage is identical: 57 blocks each, zero exclusive blocks.
4. The whole clean package exceeded the 90-second test budget in TestJsxScannerMutants. An initial twelve-row clean JSX-shard group also exceeded it. That group was split into groups of four. Every current row then passed on clean source. All timeout logs are preserved; no timeout was counted as a mutant kill.
5. Both performance rows skip by default. They were enabled for baseline and D1 so neither remains an unknown competitor to the unique catch.
6. No listed test names were added or removed between audit base 23a19b8b48662dfb6a19baa00b8efbb52ab5275e and defense base 76c59c81e8617cea1892a01841494895927a712c. TestMain is a hook, not a listed row, and remains present.
7. A workflow restart while removing a redundant whole-package mutant replay interrupted orchestration after D1 was planted. Its native build finished successfully. resume-defense.py verified the exact source against HEAD and continued from that completed build. No test or oracle source was changed.

Warm env.sh worked; setup was skipped. nproc=5. npm ci completed before baseline. The sanitized native compile probe took 22.843 seconds. Clean bounded group shell wall total was 442.872 seconds, excluding initial timeouts, coverage, and enabled performance runs. D1-runs.json and the logs retain individual command wall times and test-binary timings. Work began approximately 14:17 UTC on 2026-10-09.

One aimed production mutant was enough to defend the generated row. No allowed production target exists for the product wrapper's claimed construction behavior, so it received no prohibited construction/oracle mutants. No other repository packages were run. The final per-row matrix accounts for every listed package test, including both enabled performance rows. Subsumption was not inferred from mere absence of exclusive host-Go coverage.

Mutant bounded command wall total: 460.461 seconds. Evidence completed: 2026-10-09T14:46:40.363559+00:00.
