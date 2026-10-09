# Current cache defense

TestCacheBypass is defended by D01, a unique rejection-policy catch. TestNodeCacheProgram and TestNativeCacheGeneratedC are not defended after three attempts each. Their extra command-execution behavior is not covered by the original sole subsumer, but each new catch was shared elsewhere. These results do not authorize deleting or rewriting tests.

## Start and scope

Started clean at origin/main b8bcadb2c493173855f19d7e5c508b34f5eeb5b6. Fetched the audit with its full refspec; read REPORT.md and rows.json, including the report's oracle, code, timing and prior-blocked-audit discussion. Read CLAUDE.md and required repository documentation, the whole cache_test.go and edit_test.go, cache.go and run.go, plus command boundary and compiler comparison tests. Audit used origin/main 8171b3173bdbfce1f7982d3c4f731279307ece37.

Current list contains 44 top-level tests, seven more than the audit. scope.json records every name and added row. All requested names still exist in cache_test.go. No grouping into families: these tests have distinct assertions. These are not executor twins of a shared port: the Node row executes handwritten JavaScript, the native row executes separately handwritten C, and neither compiles a shared port.

## Code under test and oracle

The code under test is the production test262 runner's cache identity, eligibility, storage and subprocess execution, specifically nodeResultKey, nativeResultKey, cacheKey, resultCache.reuse, resultCache.observe, recordExecution and recordedExecution.execution, runCommand and runCommandWithLimit with limitedBuffer. No changes to tests, fixture programs, Node, clang, comparison expectations or the compiler oracle were made. Broader engine setup reached by the subsumer is unmutated.

Oracle for Node: Node actually executes the stored program, then the test checks literal before/after stdout. Oracle for native: clang builds handwritten C and the test checks the binary's literal before/after stdout. The expected words are self-written rather than copied from an external authority. The bypass oracle is self-written cache-state requirements: fresh bypass output, original stored output afterward, and no disk publication of transient failures.

## Coverage and aimed differences

For each of TestNodeCacheProgram, TestNativeCacheGeneratedC, TestCacheKeyDimensions, TestCacheBypass and TestEditCacheSeparation ran, after sourcing env.sh:

```
ADAMIC_TEST262_MEASURE=1 timeout 120 go test -count=1 -timeout 90s ./cmd/adamic-test262/ -run '^NAME$' -coverpkg=./cmd/adamic-test262 -coverprofile=NAME.cover > NAME-coverage.log 2>&1
```

Raw profiles and coverage-differences.json are retained. The two program rows each cover 11 runner blocks absent from TestCacheKeyDimensions, including real command capture and successful execution. TestCacheBypass covers bypass execution and transient-result early return, absent from TestEditCacheSeparation. These profiles are scoped to runner code, where all five mutants were planted; no claim of exhaustive transitive compiler/native coverage.

D01 changes the failed-start sentinel from -1 to -2. The TimedOut clause remains intact, but a non-timeout failed start can be published. Only TestCacheBypass fails with cache_test.go:130: transient failure cached. matrix.json records all 43 passing rows.

D02 removes the Node source key option; D03 removes the generated-C key option. Both reproduce shared source-invalidation catches by the requested row and TestCacheKeyDimensions. D04 changes the runCommand capture limit option from outputLimit to six bytes. Real before\n output requires seven bytes; synthetic key observations do not execute a command. Both program rows fail, along with three other rows. D05 swaps the stdout/stderr destinations of runCommandWithLimit. Both program rows fail alongside eleven other rows. It mutates the production executor, not the external executed programs or expected output. The separate compiler reference in compiler_test.go uses its own command capture and remains unchanged.

Each real-program row received three attempts: its own key option fault, then D04 and D05. These establish extra behavior beyond the former subsumer, but no unique catch. No single row in this matrix subsumes all three catches of either requested program row: TestCacheKeyDimensions catches the key fault but passes D04 and D05. Preserve that distinction from the audit's smaller mutant set.

## Matrix and validation

plan.json was saved before mutation outcomes. Every standalone diff applies to the stated origin/main and passed go vet ./cmd/adamic-test262/ while applied. Each ran the full 44-row package with ADAMIC_TEST262_MEASURE=1, timeout 120 go test -json -count=1 -timeout 90s, and its own /tmp/defend262/cache/ID directory. matrix.json includes exact commands, failures, passing rows, vet and wall times. All rows completed, no skips, panics, timeouts or unknown results. No narrowed matrix or repository-wide uniqueness claim. Production source was restored after every mutant.

- D01: binary 72.180 s, full command wall 83.555 s, vet 0.574 s.
- D02: binary 56.645 s, full command wall 65.922 s, vet 0.361 s.
- D03: binary 70.278 s, full command wall 78.145 s, vet 0.293 s.
- D04: binary 61.941 s, full command wall 72.765 s, vet 0.458 s.
- D05: binary 47.057 s, full command wall 55.550 s, vet 0.325 s.

## Setup, friction and owner findings

nproc=5; warm /workspace/adamic-tools/env.sh worked, so setup was skipped. npm ci in stage3/api ran before baseline; npm.log retains its timing. Enabled clean package baseline passed in 82.504 binary seconds, with no skipped row. Its proximity to 90 seconds constrained the matrix to five mutants, each serving the specified three-attempt limit per program row. No compiler lowering or port mutations were required.

Initial df showed /tmp total capacity 8.8 GB, so 15 GB free there is impossible. Removed the previous unit's named /tmp/defendcorpus scratch directory; recheck showed /tmp 8.8 GB free and /workspace 16 GB free. Tools and repository were untouched.

The audit's oracle wording mentions cache invocation counts for these program rows, but their current bodies assert only the before/after stdout. They never repeat an unchanged input and never assert a hit or avoided execution. Thus their names describe source-change integration checks, not a proof of cache reuse or cost. Neither name explicitly promises a timing threshold. This is an owner finding even though their output checks demonstrably fail. They must not be called untrue based on these results.

The brief's allowed menu does not separately name removing a key argument; D02/D03 change the selected key-input options, the same faults used in the cited audit. No inserted statement or test edit is used. The twin enum extension is irrelevant to these distinct handwritten programs. The existing remote defense branch will be preserved through a normal merge, with no force push or PR.
