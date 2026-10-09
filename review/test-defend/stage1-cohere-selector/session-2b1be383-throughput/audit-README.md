# u134 selector test audit

Base d054e3578d4e53bef8cb8c9ab82d6ace69cdafeb. Evidence scope from discovery.log.

Scope: five discovered top-level tests, no families, no skips with both opt-ins enabled. Starting commit d054e3578d4e53bef8cb8c9ab82d6ace69cdafeb. The Go helper under testdata is an external-oracle subprocess entry, not a test discovered in this package. No other package tests were run for uniqueness. Production matrix runs every package row.

Brief ambiguities and costs
1. The default baseline skips throughput, corpus controls and the upstream nontermination proof, and skips the library subcase. I installed the exact upstream library and PostCSS, then enabled ADAMIC_SELECTOR_LIBRARY and ADAMIC_SELECTOR_BENCH. The first opt-in attempt failed during setup because I had omitted PostCSS, which the local wrapper requires. This was my installation error; it was reported, corrected, and the complete baseline passed before any mutation. Both failed setup and successful baseline logs are retained.
2. TestThePortParsesAsGoCohereDoes combines actual product agreement and three built-in mutation witnesses in one top-level row. It cannot be assigned a witness-only verdict. Production kills establish its sacred verdict; W1 separately proves all three embedded witness comparisons can fail. The empty-port probe leaves all three embedded mutation witnesses passing, so those subcases alone do not establish a working port.
3. TestSelectorThroughput is an opt-in test of aggregate counts, not a full semantic oracle. It misses the line-number and quote-output mutations even with the upstream library enabled. P1 demonstrates that native and Node source can agree on an empty answer. The upstream count check then fails the row. The supplemental P1-no-library run records whether the same row passes with that library oracle absent; it is an environment probe, not a production mutant.
4. The nontermination row observes an external oracle's bounded behavior rather than Adamic parsing. I treat its marker-bearing proving invocation as setup construction: S2 drops the local invocation, leaving both markers, and the test fails. The external parser is untouched. Its setup-check verdict applies to that observation fixture, not to upstream semantic correctness or proof of infinite execution.
5. The corpus row is also setup construction. S1 drops collection from the local glue after PostCSS parsing, not from PostCSS itself. P3 empties only corpus mode, leaving ordinary upstream comparisons unchanged. Neither is counted as a production kill.
6. The gap row tests compiler lowering rather than the selector port. I used measured coverage to inventory 303 reached lowering functions. M4 changes a compiler diagnostic constant; its unique catch proves the expected wording is pinned, not that every unsupported operation is diagnosed correctly. Node validates the fixture's executable meaning, while the refusal labels come from our own GAPS.md.
7. A source-mutant selector would not remove this harness's temporary native rebuilds. I used the allowed four-mutant route: three port mutations and one compiler mutation. All compiler-mutant/probe runs receive unique ADAMIC_BUILD_CACHE_DIR values. Individual standalone sanitized rebuild and vet timings are retained.
8. The compiler's empty answer is nil IR and nil error. Replacing the whole Lower body also requires removing imports used only by that body. The initial draft failed compilation because fmt and filepath became unused. That invalid draft contributes no probe result. After cleanup, the corrected probe was rerun against the package and every row separately, because nil IR panics abort the binary.
9. The nil-IR probe's crashes are distinguishable from semantic mismatches. The gap's negative cases also reject the missing refusal. Only a row's own entry probe determines its vacuity: P1 for port/throughput, P2 for lowering gaps, P3 for corpus construction, P4 for the nontermination proving entry.
10. Numbered wrapper families are absent here. Several rows use table-driven subtests, but they remain one top-level row. The port row's distinct product and witness assertions are recorded as subcase findings instead of inventing extra top-level rows.
11. Four source mutants are a sample. Throughput subsumption rests on its shared root-child kill alone. It is not evidence that throughput adds no performance information, nor deletion advice. Package uniqueness is observed for this sample; repository uniqueness remains for central replay.
12. Raw logs are ignored by the repository, so they must be force-added explicitly. Commands are shell-quoted in runs.json; their environment is recorded beside them. Printed binary ok-line durations, not outer Go command wall time, determine the three-run medians.

Measured rows

| Row | Binary median seconds | Verdict | Kills |
|---|---:|---|---|
| TestEachGapStandsWhereGapsMdSaysItDoes | 0.419 | sacred | M4 |
| TestThePortParsesAsGoCohereDoes | 17.379 | sacred | M1, M2, M3 |
| TestSelectorThroughput | 20.645 | subsumed | M2 |
| TestCorpusKeepsEveryParseableFile | 0.237 | setup-check |  |
| TestTheLibraryDoesNotReturnOnUnconsumedNamespaceBars | 1.053 | setup-check |  |

Mutants

| ID | Base file:line | Change | Failing rows |
|---|---|---|---|
| M1 | stage1/cohere/selector/tokenize.ts:67 | `let line = 1;` to `let line = 2;` | TestThePortParsesAsGoCohereDoes |
| M2 | stage1/cohere/selector/parser.ts:28 | `root.children.push(1);` to `(drop statement)` | TestSelectorThroughput, TestThePortParsesAsGoCohereDoes |
| M3 | stage1/cohere/selector/nodes.ts:42 | `if(code === 34) {` to `if(code === 35) {` | TestThePortParsesAsGoCohereDoes |
| M4 | internal/lower/object.go:1266 | `push with other than one value` to `push with other than two values` | TestEachGapStandsWhereGapsMdSaysItDoes |

All four production mutants were caught. No survivors or equivalent candidates. No bounded production matrix or package timeout. Every source diff applies to base and compiles with its own tool. Go source and test source are restored.

The active throughput row is nonvacuous only because the optional upstream count oracle is enabled: P1-no-library.log shows it passing with the empty port. The port row lists its passing empty-driver witness subcases in rows.json. W1 weakens only their shared difference detector and fails all three.

Compilation and probe reruns

- P2 (corrected probe): exit 1, 17.309s; `timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/selector/ -run .`.
- P2-TestEachGapStandsWhereGapsMdSaysItDoes (corrected probe): exit 1, 2.029s; `timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/selector/ -run '^TestEachGapStandsWhereGapsMdSaysItDoes$'`.
- P2-TestThePortParsesAsGoCohereDoes (corrected probe): exit 1, 8.884s; `timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/selector/ -run '^TestThePortParsesAsGoCohereDoes$'`.
- P2-TestSelectorThroughput (corrected probe): exit 1, 8.749s; `timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/selector/ -run '^TestSelectorThroughput$'`.
- P2-TestCorpusKeepsEveryParseableFile (corrected probe): exit 0, 1.999s; `timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/selector/ -run '^TestCorpusKeepsEveryParseableFile$'`.
- P2-TestTheLibraryDoesNotReturnOnUnconsumedNamespaceBars (corrected probe): exit 0, 2.913s; `timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/selector/ -run '^TestTheLibraryDoesNotReturnOnUnconsumedNamespaceBars$'`.
- M1 (standalone compile): exit 0, 1.568s; `timeout 90 go run ./cmd/adamic build /workspace/adamic/stage1/cohere/selector/main.ts -o /tmp/u134/M1-native --sanitize`.
- M2 (standalone compile): exit 0, 1.708s; `timeout 90 go run ./cmd/adamic build /workspace/adamic/stage1/cohere/selector/main.ts -o /tmp/u134/M2-native --sanitize`.
- M3 (standalone compile): exit 0, 1.617s; `timeout 90 go run ./cmd/adamic build /workspace/adamic/stage1/cohere/selector/main.ts -o /tmp/u134/M3-native --sanitize`.
- M4 (standalone compile): exit 0, 0.478s; `go vet ./internal/lower/`.
- P1 (standalone compile): exit 0, 0.573s; `timeout 90 go run ./cmd/adamic build /workspace/adamic/stage1/cohere/selector/main.ts -o /tmp/u134/P1-native --sanitize`.
- P1-no-library (supplemental environment probe): exit 0, 5.799s; `timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/selector/ -run '^TestSelectorThroughput$'`.
- P2 (standalone compile): exit 0, 0.469s; `go vet ./internal/lower/`.
- W1 (standalone compile): exit 0, 0.153s; `go vet ./stage1/cohere/selector/`.
- S1 (standalone compile): exit 0, 0.027s; `node --check /workspace/adamic/stage1/cohere/selector/testdata/library.mjs`.
- S2 (standalone compile): exit 0, 0.024s; `node --check /workspace/adamic/stage1/cohere/selector/testdata/nontermination.mjs`.
- P3 (standalone compile): exit 0, 0.024s; `node --check /workspace/adamic/stage1/cohere/selector/testdata/library.mjs`.
- P4 (standalone compile): exit 0, 0.025s; `node --check /workspace/adamic/stage1/cohere/selector/testdata/nontermination.mjs`.

Timing summary

```json
{
  "setup_seconds": 0,
  "nproc": 5,
  "npm_api_reported_seconds": 0.442,
  "npm_selector_initial_reported_seconds": 0.72,
  "npm_complete_reported_seconds": 0.608,
  "clean_complete_binary_seconds": 39.599,
  "phase_wall_seconds": {
    "three-run timings": 149.5381924259982,
    "production matrix": 156.07311129699883,
    "probes and controls": 49.872856525000316,
    "invalid probe drafts": 1.4879217079997034
  },
  "validation_wall_seconds": 54.34928020299776,
  "source_restored": true
}
```

The standalone rebuild times include Go CLI startup. Test binary times combine Go oracle generation, native builds, library work and execution; those internal phases were not separately timed. The full enabled clean baseline was 39.599s. The earlier default baseline passed in 14.119s but skipped installable opt-ins; it is not used to justify complete coverage.

Not covered: other packages, repo-wide uniqueness, exhaustive grammar mutations, upstream parser source mutation, infinite-time nontermination, or separate build/run phases inside each test. Throughput subsumption rests on one shared mutant, M2. Raw logs, standalone diffs, environment and cache selectors, timing repetitions, function inventory, coverage, matrix, row JSON and workflow scripts are included.
