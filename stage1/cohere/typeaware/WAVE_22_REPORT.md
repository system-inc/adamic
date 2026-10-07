Built: three wave 22 rules, an isolated checker question, and an independent byte oracle.
Commits: claim 893009ccc279031860181059c71765fa32974996; implementation is the commit containing this report.
Commands and outputs: rule agreement PASS (113.722s), bridge PASS (121.063s), Node oracle PASS (38.603s), vet PASS.
Mutants: three rule mutants, export normalization, entity-kind guard, released registry, seven bridge checks, and Node one-byte check caught.
Not covered: complete repository gate, cohere CLI lint on .a, all option combinations, and full upstream fixture matrix.

## Selection and scope

Base is origin/codex/tsgo-c-library at 0d540f413625f016f20fea39761c7b184f335de6. The explicit unit base takes precedence over the generic main instruction. CLAUDE.md, the typeaware README, VOLUME_REPORT.md and COVERAGE_REPORT.md were read before edits.

Ranking combines the frozen compiler and repository volumes, descending by total and then rule name, excluding the base's 26 ports (including method-signature-style, absent from the checker-dependent table). Positions 64–66 among the remaining 172 checker-dependent rules are:

| Position | Rule | Recorded corpus findings |
| --- | --- | --- |
| 64 | @typescript-eslint/no-non-null-asserted-nullish-coalescing | 0 |
| 65 | @typescript-eslint/no-unnecessary-qualifier | 0 |
| 66 | @typescript-eslint/no-unused-private-class-members | 0 |

All fetched origin branches were searched for these ports and claims before claiming; none were present, so none were skipped. The claim commit was pushed before implementation. No PR was opened.

Each rule has its own .a file. The dedicated wave_22_suite.a runs them together, preserving the existing full diagnostic serialization and ordering. The new scope-export-symbols question lives in scope_export_symbols.go and scope_export_symbols.a, with protocol documentation and direct checker tests in separate files. Its only shared edit is one dispatch line in bridge/tsgo/checker/facts.go. Adamic directly imports the new adapter, so no shared Adamic registration file is necessary. Shared/protected compiler files and existing rule implementations are unchanged.

The Go question returns checker facts: the entity symbol, alias target, and export-normalized identity of a same-name scoped symbol. Adamic decides whether qualification is unnecessary. The Go oracle invokes unchanged production cohere rules, not the new bridge or Adamic decisions. symbol_facts_wave_22.a decodes existing symbol facts without importing the broad Caller helper. Importing that helper initially failed the compiler's existing this-escape refusal in unused.ts; the focused decoder resolved it without changing existing files. Newly written Adamic files, including mutant copies of legacy helpers, use .a.

## Agreement and independent controls

The frozen manifests contain 77 compiler and 287 repository roots. Compiler source is TypeScript commit 050880ce59e30b356b686bd3144efe24f875ebc8. Repository roots retain the base branch's manifest. Cohere is pinned at 715ba94f3608a6500086b1076ce5cb7e51b836db, typescript-go at 8d550c837c90bd1805b047b7eeccc2baac2d5e7a.

| Input | Roots | Findings | Identical bytes, Go/native/native sanitizers |
| --- | ---: | ---: | ---: |
| Production fixture sources plus targeted controls | 206 | 104 | 27,900 |
| Repository | 287 | 0 | 18,485 |
| TypeScript src/compiler | 77 | 0 | 5,857 |

The 206 controls are extracted independently from production rule test case strings, supplemented with 15 cases for Unicode, CRLF, parentheses, alias and namespace shadowing, writes, destructuring, private parameter properties, and nested classes. Go's parser accepts every control. Each rule has positive controls. The full output includes finding rule/message/spans, fixes, suggestions, and edits; comparisons do not reduce to counts. The zero corpus findings match the ranking and are not treated as sufficient evidence alone.

Committed evidence in validation-wave-22 contains exact gzip streams, their byte counts and SHA-256 hashes, all control sources in controls.json, portable corpus manifests, source hashes, test logs, and timing counters. Absolute file headers correspond to /workspace/wave-22-final-a. Executables and generated native build artifacts remain there. The new test regenerates inputs and validates them on another checkout.

## Mutants and refusals

All three rule mutants compile, exit 0, and have empty stderr; only full output comparison catches them.

| Mutant | Catch |
| --- | --- |
| Nullish assertion suggestion edit ends one byte late | Full diagnostic byte mismatch at 327 |
| Unnecessary qualifier fix ends one byte late | Full diagnostic byte mismatch at 17,734 |
| Private-member read predicate inverted | Full diagnostic byte mismatch at 3,472 |
| Scoped candidate identity skips GetExportSymbolOfSymbol | Full diagnostic byte mismatch at 23,847; normal exit and empty stderr |
| New checker question accepts non-entity nodes | TestScopeExportSymbols fails: accepted non-entity node |
| Released checker registry keeps handle live | Required refusal detects mutant exit 0; correct implementation panics with exit 70 and exact invalid-or-released-handle message |
| Existing bridge input length increased by one | ASan heap-buffer-overflow |
| Existing bridge output length increased by one | ASan heap-buffer-overflow |
| Existing bridge released handle retained | Stale-handle assertion |
| Existing bridge type queried at source-file position | Independent byte oracle mismatch at byte 6 |
| Existing bridge explicit link guard removed | Expected compiler refusal fails |
| Existing bridge C output free omitted | LeakSanitizer |
| Existing bridge region entry heap allocated | LeakSanitizer |
| Existing native/Node oracle one-byte output change | TestTheOracleCatchesOneByte |

The new optimized and sanitizer native binaries agree with Go on all controls and both corpora. The existing full bridge suite additionally compares 1,600 positions across four compiler files, 54,982 bytes under ASan/UBSan/LSan, checks retained C output after release, invalid/zero/released handles, and region allocation behavior. Direct question tests cover value/type namespaces, lexical shadowing, aliases, canonical flag encoding, malformed questions and entity-kind refusal.

## Timing observations

Three alternating Go/native rounds ran after all builds and tests finished. Timings include process startup, loading, parsing, lint traversal, full output serialization, and teardown. Diagnostic bytes were compared on every timed round. These are observations on this worker, not a speed guarantee.

| Corpus | Native median | Go median | Native / Go | Native checker queries |
| --- | ---: | ---: | ---: | ---: |
| Repository, 287 roots | 0.524900s | 0.200730s | 2.615 | 30 |
| Compiler, 77 roots | 2.838664s | 0.479345s | 5.922 | 6,486 |

Native is slower on these workloads. Per-round elapsed times and load/run/query counters are in validation-wave-22/timings.json and corresponding stderr files. Zero findings make findings-per-second inappropriate; these measurements cover whole manifests.

## Commands and outputs

All test stdout/stderr went to files, without piping. Source /workspace/adamic-tools/env.sh before commands (the setup-selected path).

```sh
bash cloud/setup.sh > /workspace/wave-22-setup.log 2>&1
# go 1.27.1; clang 20.1.8; Node 24.19.0
# go ready 0s; clang ready 0s; node ready 0s; submodules ready 0s
# build cache warm 118s; done 118s; nproc 5
# cpu.max 400000 100000; reported memory 17.6 GB

ADAMIC_WAVE22_ARTIFACTS=/workspace/wave-22-final-a \
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-22-typescript-pinned \
go test ./stage1/cohere/typeaware -run '^TestWave22AgreementAndMutants$' \
-count=1 -timeout=30m -v > /workspace/wave-22-final-a.log 2>&1
# PASS 113.722s

TMPDIR=/workspace/wave-22-scratch \
ADAMIC_TSGO_CORPUS=/workspace/wave-22-typescript-pinned \
go test ./bridge/tsgo/... -count=1 -timeout=15m -v > /workspace/wave-22-bridge.log 2>&1
# bridge PASS 121.063s; checker PASS 0.325s

go test ./bridge/tsgo/checker -count=1 -v > /workspace/wave-22-checker-final.log 2>&1
# PASS 0.363s

go test -overlay /workspace/wave-22-scope-kind-overlay.json \
./bridge/tsgo/checker -run '^TestScopeExportSymbols$' -count=1 -v \
> /workspace/wave-22-scope-kind-mutant.log 2>&1
# Expected FAIL, exit 1: accepted non-entity node

go test ./internal/oracle \
-run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(maps_and_text|sorting|string_index|lone_surrogates|functions|closures)\.a$' \
-count=1 -timeout=10m -v > /workspace/wave-22-node-oracle.log 2>&1
# PASS 38.603s; eight fixtures and one-byte mutant

go vet ./... > /workspace/wave-22-vet.log 2>&1
# PASS, empty output

go vet ./stage1/cohere/typeaware ./bridge/tsgo/... > /workspace/wave-22-vet-final.log 2>&1
# PASS after final harness change, empty output

gofmt -l cmd internal bridge/tsgo/checker/scope_export_symbols.go \
bridge/tsgo/checker/scope_export_symbols_test.go stage1/cohere/typeaware/wave_22_test.go \
stage1/cohere/typeaware/testdata/oracle_wave_22.go \
> stage1/cohere/typeaware/validation-wave-22/gofmt.log
# Empty output; shared dispatch line retained as one line per unit instruction

python3 stage1/cohere/typeaware/validation-wave-22/benchmark.py \
/workspace/wave-22-final-a /workspace/adamic /workspace/wave-22-typescript-pinned \
/workspace/wave-22-benchmark > /workspace/wave-22-benchmark.log 2>&1
# Three rounds per implementation/corpus; every stream identical
```

## Limits

The complete repository gate was not run; the touched rule harness, entire bridge package tree, vet, and filtered Node oracle were run instead. The pinned cohere CLI built successfully but rejects explicitly named .a paths before linting, saying nothing to check because the paths are not TypeScript/JavaScript. Its lint gate therefore cannot validate these new .a files without an upstream CLI extension change. This is recorded in cohere-cli.log; no CLI lint success is claimed.

The extracted controls do not represent every production test helper, option combination, JSX mode or source shape. Agreement covers defaults used by the existing corpus runner, serialized repairs and suggestions; applying fixes and re-running every resulting file is not separately covered. The direct checker node-kind mutant was run via a scratch overlay, while rule/export/registry mutants are reproducible in TestWave22AgreementAndMutants. Concurrent future origin claims cannot be excluded by the initial branch scan; the pre-code pushed claim establishes this unit's reservation.
