Certified 19 original one-read array-field pairs toward roadmap step 09; retained one compiler frontier.
Commits: base 70522aa1d73a80e295df608dee44b38414301c97; delivery SHA accompanies this report; observed worker tip 908cafeb42ad01dfcfab2250dbb0454de1a791a7.
Checks: 135 original probes and 24 mutants pass in 67.395s; 128 scoped count rows measured in 109.411s; required global counts fail in existing fixtures; vet and formatting pass.
Mutants: all 24 are caught by exact stopping oracles in sanitized native, release native and JavaScript; every finishing mutant is leak-clean.
Not covered: project-reference path admission, nested array replacement, delegated union/tuple/dictionary/mixed-member pairs, later ranks, and production runtime reachability.

This is the first 20-pair delivery batch on codex/views-arrays-b, based exactly on the requested 70522aa1d tip. No compiler or runtime implementation changed. Fixtures and declaration preparation are owned under originalB; the only existing test hook change appends this worker's count rows.

Scheduling assumption: the recorded union-target, tuple, dictionary and mixed-member handoffs retain their owners. This conservatively skips their one-read overlaps, rather than claiming them from lane 2's array adapter. The three already-certified fileInfos joints are skipped too. progress.json retains every original span, type ID, disposition and reason across the scanned tail. Own eligible pairs are processed in reverse candidate_queue order. All selected pairs have one static read. The next unselected own pair is ConfigFileSpecs.validatedFilesSpecBeforeSubstitution. This batch stops at the required delivery size; it has not met the worker's three-read frontier.

The initial report at 70522aa1d records 183 pairs / 2899 static reads certified, 151 / 290 remaining. Before delivery, the worker advanced to 908cafeb and added its callback pair / four reads, giving 184 / 2903 and 150 / 286 remaining. Its full report is preserved in evidence/worker-ranked23-report.md. These are disjoint from this batch. The combined scheduling ledger would therefore be 203 / 2922 certified and 131 / 267 remaining, out of 334 / 3189. These are fixture-held static candidate obligations, not executions of the compiler corpus. Consumer, intrinsic and own-array-field totals receive no additional credit here. No worker branch was merged.

| Original pair | Type ID | Static reads | Result |
| --- | ---: | ---: | --- |
| `IncrementalMultiFileEmitBuildInfo.fileIdsList` | 97922 | 1 | certified |
| `ParsedCommandLine | undefined.projectReferences` | 94919 | 1 | code needed |
| `SortedAndCanonicalizedMutableFileSystemEntries.directories` | 93573 | 1 | certified |
| `ObjectLiteralExpressionBase<ObjectLiteralElement>.properties` | 46301 | 1 | certified |
| `HeritageClause | undefined.types` | 40723 | 1 | certified |
| `ObjectLiteralExpression | undefined.properties` | 38740 | 1 | certified |
| `JSDoc | undefined.tags` | 36241 | 1 | certified |
| `SyntaxList._children` | 11834 | 1 | certified |
| `ParsedCommandLine | undefined.fileNames` | 11560 | 1 | certified |
| `WatchOptions.excludeFiles` | 11164 | 1 | certified |
| `WatchOptions.excludeDirectories` | 11164 | 1 | certified |
| `TsConfigSourceFile | undefined.extendedSourceFiles` | 11111 | 1 | certified |
| `SignatureDeclaration | JSDocSignature | undefined.parameters` | 10862 | 1 | certified |
| `ConditionalRoot.aliasTypeArguments` | 10856 | 1 | certified |
| `TupleType.typeParameters` | 10808 | 1 | certified |
| `InterfaceTypeWithDeclaredMembers.typeParameters` | 10788 | 1 | certified |
| `SetAccessorDeclaration | undefined.parameters` | 10497 | 1 | certified |
| `Signature | undefined.typeParameters` | 10371 | 1 | certified |
| `ConfigFileSpecs | undefined.validatedFilesSpec` | 9854 | 1 | certified |
| `ConfigFileSpecs.validatedIncludeSpecsBeforeSubstitution` | 9853 | 1 | certified |

Preparation uses independent pristine microsoft/TypeScript at 050880ce59e30b356b686bd3144efe24f875ebc8. The existing lane4b declaration adapter is invoked by reference; no cohere file is copied. All 78 complete emitted declarations are retained. The complete original private SortedAndCanonicalizedMutableFileSystemEntries and its Canonicalized dependency are printed from the original AST into an external declaration file, adding only exports and the original SortedArray dependency import. All 79 declaration hashes, original member types, complete receiver/selected-element field sets and exact census spans are checked. The compressed original manifest is preserved under evidence.

Every selected pair has good, stopping, unread malformed storage, unread later element and alias witnesses. Optional fields additionally have absent and explicit undefined controls; optional receivers additionally have undefined controls. Required union-with-undefined fields keep a distinct missing-field exit-70 pin. Source Node runs all 135 fixtures with explicit stdout, empty stderr and exit zero. 128 executable fixtures run on sanitized native, release native and JavaScript; successful programs have native leak checks. The seven remaining sources retain the exact compile refusal below. In the signature/JSDoc optional union, complete original receiver-arm descriptors are asserted; the source controls select the first original signature arm and ParameterDeclaration element. Other original arms remain in the descriptor graph and are not claimed as additional runtime arm coverage.

Observed code requirements:

- ParsedCommandLine | undefined.projectReferences: the full declaration has ProjectReference.path: string, but its demanded read stops with `Adamic 0.1 refuses checked view read of field path with unsupported intersection contract; prove or implement the intersection contract before reading this field`. All seven demanded source controls pin that message; the unread array control runs. This pair earns no credit. The inference is that array-element demand scoping must avoid importing an unrelated branded path obligation, while preserving every applicable read check. No field or original declaration was reduced to avoid it.
- IncrementalMultiFileEmitBuildInfo.fileIdsList: nested reads and scalar writes through the original inner-array alias pass. Replacing the inner array with another array remains `adamic: panic: element read failed: <array write> expected array, found uncertified source element contract`, exit 70, in all three modes. Its source Node prints 9. The full original replacement witness is retained and counted. Reference-element producer certificates are needed for this write; the read pair's credit does not claim this write is implemented.

mutants.json lists every independent execution, pair, catcher and observed output. Seven string-consumer mutants admit numeric elements through a string | number consumer descriptor and print 42. Eleven numeric descendant mutants replace the read contract and conversion with a string read and print bad. Five required-presence mutants permit absent fields and print absent. The nested numeric-index mutant returns a type-correct maybe-number zero instead of reading the invalid inner element and prints 0; source Node prints bad. Each mutant changes exactly one selected operation, finishes with exit zero and empty stderr, passes ASan/UBSan and native leaks, and disagrees with the original exact exit-70 oracle. The first scratch attempt at the nested default omitted its MaybeNumber representation and caused an emitter panic; that attempt was not counted as a mutant catch. Only the corrected executing mutant is credited. Unsupported compiler/write frontiers are listed as code work, with no completion or mutant credit.

Counts refresh adds exactly 128 own rows. count-rows.json explains each new row: a new fixture measured either at normal completion or at its pinned stopping read. There are no changed pre-existing rows, numerical changes, reordering or header changes: removing the own prefix reproduces the base table byte for byte. The seven compile refusals have no runtime row. TestCountsAreRecorded was run twice and remains red in existing graph-region invalid frees, process.exit lowering, fs operations and other existing fixtures. The final run failed in 54.560s; its complete log is preserved. This is not a successful global refresh. The scoped measured update passes and leaves all other rows untouched.

Exact commands, with every test's output sent directly to its named log:

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/lane2-b-setup.log 2>&1
source /workspace/adamic-tools/env.sh
npm ci --prefix stage3/api > /tmp/lane2-b-api-install.log 2>&1
node stage3/interface-downcasts/lane2/originalB/prepare.cjs /workspace/lane2-b-original-pin /workspace/lane2-b-declarations > /tmp/lane2-b-prepare.log 2>&1
ADAMIC_ARRAYB_ORIGINAL_DECLS=/workspace/lane2-b-declarations ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewArraysBOriginal$|^TestCheckedViewArraysBMutants$' -count=1 -parallel 4 -v -timeout 20m > /tmp/lane2-b-final.log 2>&1
ADAMIC_ARRAYB_ORIGINAL_DECLS=/workspace/lane2-b-declarations ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewArraysBCounts$' -count=1 -timeout 15m -args -update-counts > /tmp/lane2-b-counts-update.log 2>&1
ADAMIC_ARRAYB_ORIGINAL_DECLS=/workspace/lane2-b-declarations ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 20m -args -update-counts > /tmp/lane2-b-global-counts-final.log 2>&1
ADAMIC_ARRAYB_ORIGINAL_DECLS=/workspace/lane2-b-declarations go test ./internal/oracle -run '^TestCheckedViewArraysBOriginal$/^incrementalmultifileemitbuildinfo-fileidslist-good$' -count=1 -v -timeout 10m > /tmp/lane2-b-manifest-check.log 2>&1
go vet ./internal/oracle > /tmp/lane2-b-vet.log 2>&1
gofmt -l internal/oracle/checked_views_arrays_b_test.go internal/oracle/checked_views_array_counts_test.go > /tmp/lane2-b-format.log 2>&1
git diff --check > /tmp/lane2-b-diff-check.log 2>&1
```

The final extra manifest check verifies that every selected pair has its good, bad, lazy, lazy-element and alias witnesses, names are unique, and all written .a files are registered. It passed in 1.501s. This guard was added after the complete oracle run; it does not change backend execution. Scoped vet, formatting and diff-check logs are empty. No whole package or full gate ran. Without opt-in prepared declarations, the original-fixture suites skip and the global counts hook explicitly preserves its recorded rows without claiming remeasurement.

Setup passed: Go ready 0.021s, Node ready 0.023s, submodules ready 0.055s, markdown dependencies ready 0.073s, clang ready 0.166s, go build ready 45.076s, test binaries deferred 45.285s, build cache warm 45.287s, done 45.313s. nproc is 5; cgroup cpu.max is 400000 100000. Go is 1.27.1, clang is 20.1.8, Node is 24.19.0. Environment file: /workspace/adamic-tools/env.sh. Full setup output is preserved in evidence/setup.log.gz. There was no setup failure.

Evidence logs are stored as .log.gz so the repository's ignored .log rule cannot drop them. Original declaration files remain external, with their complete hashes and repeatable preparation path committed. Runtime reachability, unchanged tsc execution and later ranked pairs remain unmeasured.
