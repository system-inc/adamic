Re-cut slice 5 placeholders, parser step 24 and the negative-witness ledger onto the pinned main, preserving concurrent runtime infrastructure.
Delivery is one commit on dbdf977f5b8b0979d414ebe6f20e7aadd312e309; source is 05eb9043dc8b01b32402ce33b91f97aadf524305; the handoff gives the delivery SHA.
Core results: 38 focused commands, 55 final uncached differential fixtures, guards, call-target readers and changed-package vet pass; counts add 41 rows and change 21.
All 27 recorded independent mutants and two added metadata-publication mutants are caught, along with the source embedded mutants and eight witness repairs.
Not cleared for a heavy block: CSS took 63.12s, typeaware hit its 90s timeout during archive setup, and other consumer smoke tests have product-owner prerequisite failures; the gate was not run.

The cherry-pick was --no-commit of the source commit itself, never a main-to-source diff. HEAD remained the exact pinned base until the single delivery commit. Conservative assumptions: the explicit pinned base overrides the current-main boilerplate; main owns packed caches, static owner tracking, atomic publication and async readiness rules; source Program-region lifetime remains pending. No cohere code was copied. Source .ts compatibility fixtures retain their dialect; no new .ts was authored. No ahra command, PR or Loom submission was made.

Conflicts and main changes

All 11 marked hunks across nine files are listed below; conflict-hunks.json retains their exact main/source/resolved text (counts is restored then regenerated). Main f93a65fb introduced concurrent lowering, fcf5c2ad concurrent runtime infrastructure, and e3282288 traced admission compiler products and the ledger schema.

- `cloud/admission-corpus/negative-witnesses.json`, hunk 1: Retain main schema/comment and install the source eight exact-hash ruled witnesses.
- `internal/javascript/javascript.go`, hunk 1: Preserve placeholder global declaration check and main async readiness exclusion.
- `internal/native/emit_objects.go`, hunk 1: Keep main slot-derived metadata index and source distinction between unset payload and uninitialized storage.
- `internal/native/reuse.go`, hunk 1: Keep main slot-derived metadata index and source distinction between unset payload and uninitialized storage.
- `internal/native/runtime/adamic.h`, hunk 1: Use one main packed atomic cache snapshot, adding source per-object presence check on its local slot.
- `internal/native/runtime/class_static.c`, hunk 1: Publish the field while retaining main slot-derived readiness bookkeeping.
- `internal/native/runtime/object.c`, hunk 1: Skip absent construction fields and retain main zero initializer for packed cache and owner-aware copy below.
- `internal/native/runtime/object.c`, hunk 2: Keep atomic packed lookup and add per-object presence without rereading shared cache.
- `internal/native/runtime/object.c`, hunk 3: Initialize and publish the own slot, preserving main rule not to initialize a static parent.
- `internal/native/runtime/union.c`, hunk 1: Retain source construction scalar tagging and main acquire publication of immutable shape metadata.
- `internal/oracle/counts.md`, hunk entire table: Restored the pinned main file verbatim, then regenerated once on Linux with clang 20.1.8 after compilation; no hand merge. 41 added and 21 changed rows, all source-attributed, no removals.

Compatible-merge adaptations also replace library_object.c publication cache.index with an index derived from the returned slot. Reserved object-view writes find storage through immutable shape names, avoiding a second shared-cache read. Placeholder views use the actual static owner: main's owner-aware helper is exposed as adamic_object_field_owned and metadata indices derive from its returned slot.

Runtime clearance

Runtime edit lines differ from the source in exactly five files: adamic.h, class_static.c, library_object.c, object.c and share.c. Every differing added/deleted line is in runtime-line-differences.md. runtime-comparison.json retains both edit sequences; source-runtime-patch.json and rebased-runtime-patch.json contain complete exact patches as JSON strings. The other nine runtime files retain identical edit lines. union.c retains main's acquire load of shape_types in its surrounding context.

The additional share.c change visits reference children of NodeArray interior metadata while leaving its zero-count interior header unshared. Optional metadata admits strings; omitting this traversal would leave reachable child reference counts unprepared for workers. TestParserNodeArrayMetadataSharing passes ASan/UBSan and leak checks in 1.02s. Its valid-C skip-metadata-sharing mutant fails with metadata child not shared. A second valid-C retain-interior-metadata mutant fails with interior metadata retained, proving the zero-count-header assertion too. This is an internal graph-publication invariant, separate from the Node differential fixtures. The first harness attempt failed -Wnewline-eof in 7.91s; its corrected run and mutant are retained. No counts fixture was added by this test. Post-sharing NodeArray counted-row, Node differential construction, async/shared-reference and runtime-set regressions pass.

Counts rows

Main's counts.md was restored verbatim before exactly one successful -update-counts invocation on Linux with house clang 20.1.8. It completed in 144.102s with 1,099 measured leaves; the largest count leaf was 14.73s. There are 41 added rows, 21 changed rows, no removals, no unattributed rows and no row differing from the source's measurement. counts-attribution.json records every before/after value and source attribution. The table below lists every affected row.

| Row | Before | After |
|---|---|---|
| `internal/oracle/testdata/4ddd17f_weak_single_narrowed.a` | added | 4/2/4/6/4/0 |
| `internal/oracle/testdata/backend_stop_catches.a` | added | 2/2/6/4/1/0 |
| `internal/oracle/testdata/d96d304_try_stack.a` | added | 0/0/0/0/0/0 |
| `internal/oracle/testdata/non_null_initialized.ts` | 0/0/0/0/0/0 | 55/55/50/109/18/0 |
| `internal/oracle/testdata/non_null_literal_assignment.ts` | 0/0/0/0/0/0 | 1/1/2/2/1/0 |
| `internal/oracle/testdata/non_null_static_initialized.ts` | 1/0/0/1/1/0 | 9/9/12/22/6/0 |
| `internal/oracle/testdata/non_null_static_uninitialized.ts` | 1/0/0/1/1/0 | 1/0/3/2/1/0 |
| `internal/oracle/testdata/non_null_uninitialized_capture.ts` | 0/0/0/0/0/0 | 2/0/2/0/2/0 |
| `internal/oracle/testdata/non_null_uninitialized_const.ts` | 0/0/0/0/0/0 | 0/0/2/1/0/0 |
| `internal/oracle/testdata/non_null_uninitialized_exception.ts` | 0/0/0/0/0/0 | 1/1/3/2/1/0 |
| `internal/oracle/testdata/non_null_uninitialized_field.ts` | 1/0/0/0/1/0 | 1/0/3/2/1/0 |
| `internal/oracle/testdata/non_null_uninitialized_interface.ts` | 1/0/1/0/1/0 | 1/0/2/1/1/0 |
| `internal/oracle/testdata/non_null_uninitialized_iteration.ts` | 1/0/0/1/1/0 | 6/1/10/8/6/0 |
| `internal/oracle/testdata/non_null_uninitialized_loop.ts` | 0/0/0/0/0/0 | 0/0/1/0/0/0 |
| `internal/oracle/testdata/non_null_uninitialized_map_entry.ts` | 0/0/0/0/0/0 | 0/0/1/1/0/0 |
| `internal/oracle/testdata/non_null_uninitialized_optional.ts` | 1/0/0/0/1/0 | 3/3/2/6/3/0 |
| `internal/oracle/testdata/non_null_uninitialized_spread.ts` | 1/0/0/0/1/0 | 2/0/5/5/2/0 |
| `internal/oracle/testdata/placeholder_nonnull_alias_reset.a` | added | 2/1/6/7/2/0 |
| `internal/oracle/testdata/placeholder_nonnull_assignment_result.a` | added | 0/0/2/3/0/0 |
| `internal/oracle/testdata/placeholder_nonnull_before_use.a` | added | 0/0/1/1/0/0 |
| `internal/oracle/testdata/placeholder_nonnull_field.a` | added | 9/9/15/23/5/0 |
| `internal/oracle/testdata/placeholder_nonnull_local.a` | added | 12/12/11/24/4/0 |
| `internal/oracle/testdata/placeholder_nonnull_null_before_use.a` | added | 0/0/2/1/0/0 |
| `internal/oracle/testdata/placeholder_nonnull_null_coalesce.a` | added | 8/8/15/27/4/0 |
| `internal/oracle/testdata/placeholder_nonnull_null_copy.a` | added | 26/26/102/121/7/0 |
| `internal/oracle/testdata/placeholder_nonnull_null_equality.a` | added | 4/4/10/16/2/0 |
| `internal/oracle/testdata/placeholder_nonnull_null_json.a` | added | 5/5/0/6/3/0 |
| `internal/oracle/testdata/placeholder_nonnull_null_loose.a` | added | 10/10/23/35/4/0 |
| `internal/oracle/testdata/placeholder_nonnull_null_optional.a` | added | 6/6/16/25/3/0 |
| `internal/oracle/testdata/placeholder_nonnull_null_saved_leak.a` | added | 1/0/4/4/1/0 |
| `internal/oracle/testdata/placeholder_nonnull_null_truthiness.a` | added | 1/1/7/10/1/0 |
| `internal/oracle/testdata/placeholder_nonnull_null_typeof.a` | added | 3/3/9/11/3/0 |
| `internal/oracle/testdata/placeholder_nonnull_observe.a` | added | 27/27/48/79/9/0 |
| `internal/oracle/testdata/placeholder_nonnull_return_assignment.a` | added | 0/0/2/3/0/0 |
| `internal/oracle/testdata/placeholder_nonnull_saved_leak.a` | added | 1/0/3/4/1/0 |
| `internal/oracle/testdata/placeholder_nonnull_scanner.a` | added | 4/4/9/8/3/0 |
| `stage3/interface-downcasts/default-boxed-write.ts` | 0/0/0/0/0/0 | 5/5/5/12/4/0 |
| `stage3/interface-downcasts/default-read-before-set.ts` | 0/0/0/0/0/0 | 1/0/4/2/1/0 |
| `stage3/interface-downcasts/default-staged.ts` | 0/0/0/0/0/0 | 2/2/6/7/2/0 |
| `stage3/interface-downcasts/readiness-identifier-uninitialized.ts` | 0/0/0/0/0/0 | 1/0/4/3/1/0 |
| `stage3/interface-downcasts/readiness-identifier.ts` | 0/0/0/0/0/0 | 2/2/6/8/2/0 |
| `stage3/interface-downcasts/readiness-number-uninitialized.ts` | 0/0/0/0/0/0 | 1/0/5/3/1/0 |
| `stage3/interface-downcasts/readiness-number.ts` | 0/0/0/0/0/0 | 5/5/5/10/5/0 |
| `stage3/parser-ahead/factories/already-complete.a` | added | 1/1/4/4/1/0 |
| `stage3/parser-ahead/factories/complete.a` | added | 1/1/5/4/1/0 |
| `stage3/parser-ahead/factories/escaped.a` | added | 1/1/3/4/1/0 |
| `stage3/parser-ahead/node-array/fields.a` | added | 12/12/10/23/12/0 |
| `stage3/parser-ahead/rulings/factory-unset.a` | added | 3/3/13/20/3/0 |
| `stage3/parser-ahead/rulings/factory-use.a` | added | 1/0/2/3/1/0 |
| `stage3/parser-ahead/rulings/node-array-json.a` | added | 2/2/5/9/2/0 |
| `stage3/parser-ahead/rulings/node-array-keys.a` | added | 15/15/12/16/9/0 |
| `stage3/parser-ahead/rulings/node-array-length.a` | added | 4/4/4/9/2/0 |
| `stage3/parser-ahead/rulings/node-array-unset.a` | added | 4/4/18/29/4/0 |
| `stage3/parser-ahead/rulings/null-placeholder-pending.a` | added | 2/2/4/8/2/0 |
| `stage3/parser-ahead/rulings/own-key-order.a` | added | 28/28/17/37/10/0 |
| `stage3/parser-next/speculation/arrow-tryparse.a` | added | 3/3/8/11/2/0 |
| `stage3/parser-next/speculation/misfit.a` | added | 1/0/2/1/1/0 |
| `stage3/parser-next/speculation/named-lookahead.a` | added | 5/5/1/7/4/0 |
| `stage3/parser-next/speculation/named-tryparse.a` | added | 3/3/10/16/3/0 |
| `stage3/parser-next/speculation/truthiness-rewind.a` | added | 34/34/46/85/6/0 |
| `stage3/parser-next/stack/benchmark.a` | added | 7/7/0/17200008/4/0 |
| `stage3/parser-next/stack/legal.a` | added | 7/7/0/1728/4/0 |

Validation

Commands and selectors are preserved verbatim in focused-results.json, fixture-results.json, consumer-results.json, extra-consumer-results.json and all-mutants-results.json. Complete go-test JSON output is under logs/. Raw .log files remain local as repository policy requires. test-seconds.json lists each observed test's seconds, including every new observed test. GOMAXPROCS=4 and GOFLAGS appended -p=4 while preserving -trimpath and -buildvcs=false; execution used the printed env.sh. Every go test had a 90s timeout except the counts family (180s); every orchestration/child command had an outer hard limit.

- Focused selection: all 38 groups pass. A three-test parser-count group initially failed against main's untouched table, then passed after regeneration; both logs remain. Program-region lifetime and optional stack guard-cost tests explicitly skip and are not coverage claims.
- All 55 final fixtures pass with ADAMIC_GATE_UNCACHED=1 after the final runtime adaptation. These hold source Node, backend JavaScript, ASan/UBSan native, release native and successful-program leak observations. Initial fixture results are retained separately.
- TestEveryMapAProgramHoldsIsSorted passes; placeholder/factory IR maps remain SortedMap. The plain-map mutant is caught.
- TestRuntimeSetsAreEachUsed passes, including after the sharing adaptation. No new feature set is introduced.
- TestNativeAgreementRowsNameRegisteredPrograms passes. No unregistered agreement row is added.
- TestEveryTestIsParallelOrSaysWhy passes after adding the sharing test; that test calls t.Parallel first. No exemption or foreign-test edit was added.
- All three displayed-path product tests pass and their independent physical-path mutants are caught. No Go package is added, so compiler-dependencies.json needs no declaration.
- TestCallTargetReaders passes with -timeout 90s (39s initially, then 8.10s and finally 2.07s after the test edits).
- Flow corpus remainder, the two changed non-null flow corpus units, freshness corpus remainder and JavaScript mixed-union consumer pass.
- Main's async/shared-reference Node/native tests and TestParallelSlotPairs pass. Eight TestSlice5Witness repairs and stack regressions pass.
- Vet on cmd/adamic, internal/flow, fresh, ir, javascript, lower, native and oracle exits 0 without diagnostics; native vet repeats successfully after the new sharing test.
- Thirteen deliberate .a refusal controls moved unchanged into non_null_refused/. moved-refusals.json lists them. Their refusal tests pass. Accepted default-placeholder controls remain runnable; no foreign tests were changed to exclude a failing input.

Consumer limitations and observed red

The additional direct-consumer inventory is direct-consumers.json; extra-consumer-selection.json records bounded smoke selectors, not whole suites. Selected CSS TestCompositionMatchesGo_000 passed its assertions but took 63.12s in this four-CPU run. It is unchanged by this commit and was never rerun. This is an observation, not proof of its cold-tree timing or of the cause. Splitting another package's test is outside this unit's territory. The branch therefore has no claim of complete green consumer validation.

Typeaware TestShadowIndexMissingBinding_000 also hit the 90s test timeout while its goroutine waited in harness.archive/buildProduct/typeAwareRunBuildCommand. Its stage0 phase had completed in 4.535561s; this does not establish a compiler semantic failure or its cause. It was not rerun.

The following are the latest failed consumer attempts. They are not mutant kills or compiler semantic evidence. Running the dedicated MarkdownInline and CSSNumbers printer owner tests succeeded (0.758s and 1.63s), but did not make the untraced CSSNumbers reader succeed. No product-owner implementation or unrelated test was edited.

| Package | Selected test(s) | Result |
|---|---|---|
| `./stage1/cohere/cssnumbers` | `^(TestCSSNumbers_000)$` | shards_test.go:467: build cssnumbers Go oracle answers oracle-owner-v1: settle the Go full-answer owner first |
| `./stage1/cohere/cssstrings` | `^(TestCSSStrings_000)$` | port_test.go:303: build go build cohere /workspace/adamic/cohere/cmd/adamic_stage_one/main.go go-printer-oracle-owner-v1: settle the printer owner first |
| `./stage1/cohere/estree` | `^(TestAcceptanceMutants_000)$` | oracles.go:220: build estree-acceptance-mutants-go-oracle oracle-owner-v1: oracle estree-acceptance-mutants-go-oracle oracle-owner-v1 is missing: settle stage1/cohere/estree/oracles owners first |
| `./stage1/cohere/formatfiles` | `^(TestThePortParsesAsGoCohereDoes_000)$` | formatfiles_prepare_test.go:138: build go test -c cohere ./internal/format/formatfiles formatfiles.test-owner-v1: settle the formatfiles Go oracle owner first |
| `./stage1/cohere/gitignore` | `^(TestThePortAnswersAsGoCohereAndGitDo_000)$` | answers_prepared_units_test.go:218: build go test -c cohere ./internal/gitignore oracle-oracle-owner-v1: settle the Go oracle owner first |
| `./stage1/cohere/graphql/printer` | `^(TestPrinterAsGoCohere_000)$` | oracles.go:64: build go test -c cohere ./internal/format/graphql oracle-oracle-owner-v1: settle the Go printer owner first |
| `./stage1/cohere/json` | `^(TestPortMatchesGoCohereClassOrder_000)$` | port_matches_prepared_split_test.go:190: build go build cohere ./command/formatter_comparison go-cohere-oracle-owner-v1: settle json/oracles owners first |
| `./stage1/cohere/lint` | `^(TestDecodedOptionsAndMutant_000)$` | oracles.go:175: build go build cohere ./adamic_lint_oracle oracle-owner-v1: settle the lint Go oracle owner first |
| `./stage1/cohere/lint/helpers` | `^(TestMessageRefusalsMatchGo)$` | helpers_test.go:160: build go build cohere /workspace/adamic/cohere/adamic_helper_oracle.go /workspace/adamic/cohere/adamic_helper_descriptors.go go-oracle-owner-v1: settle the oracle owner first |
| `./stage1/cohere/lint/helpers/comments` | `^(TestCommentMutants_000)$` | comment_mutants_test.go:26: build go build cohere /workspace/adamic/cohere/adamic_comments_oracle.go go-oracle-owner-v1: settle the oracle owner first |
| `./stage1/cohere/markdownblocks` | `^(TestTokenizerEvents_000)$` | pilot_family_products_test.go:102: build markdownblocks-pilot-events-go oracle-owner-v1: oracle markdownblocks-pilot-events-go oracle-owner-v1 is missing: settle the markdownblocks/oracles owners first |
| `./stage1/cohere/typeaware` | `^(TestShadowIndexMissingBinding_000)$` | exit 1; see JSON log |
| `./stage1/cohere/yaml` | `^(TestComposeMatchGo_000)$` | oracles.go:338: build go build cohere ./command/formatter_comparison go-compose-oracle-owner-v1: settle the yaml/oracles owner first |
| `./stage1/typescript/parser` | `^(TestParserChildGuard_000)$` | childguard_test.go:85: build typescript-parser-oracle oracle-owner-v1: settle stage1/typescript/parser/oracles TestProduct_ParserOracle before reading the oracle |
| `./stage1/typescript/scanner` | `^(TestScannerAgreesWithTypescriptGo_000)$` | scanner_products_test.go:308: build scanner-go-oracle oracle-owner-v1: settle stage1/typescript/scanner/oracles TestProduct_ScannerOracle before reading the oracle |

The 11 additional compiler consumers declared by main's dependency registry were also attempted; declared-consumer-results.json holds commands and logs. Ten pass. stage3/lane TestApplyPython fails because its apply checkout has neither the product manifest nor ADAMIC_TYPESCRIPT_SOURCE pointing to TypeScript v6.0.3 commit 050880ce59e30b356b686bd3144efe24f875ebc8. This is a missing gate-input prerequisite; no foreign test or source pin was changed.

Independent mutants and catchers

| Mutant | Catcher |
|---|---|
| `plain-placeholder-map` | `^TestEveryMapAProgramHoldsIsSorted$ is a plain` |
| `placeholder-physical-path` | `^TestParserPlaceholderProductsUseDisplayedPaths$ parser product changed` |
| `factory-physical-path` | `^TestParserFactoryProductsUseDisplayedPaths$ parser product changed` |
| `speculation-physical-path` | `^TestParserSpeculationProductsUseDisplayedPaths$ parser product changed` |
| `branch-union` | `^TestParserFactory(Completion.* / ReadBeforeCompletion)$` |
| `omit-escape-missing` | `^TestParserFactory(Completion.* / ReadBeforeCompletion)$` |
| `erase-earlier-read` | `^TestParserFactory(Completion.* / ReadBeforeCompletion)$` |
| `trust-assertion` | `^TestParserFactory(Completion.* / ReadBeforeCompletion)$` |
| `slot-pos` | `^TestParserNodeArrayLayout$` |
| `slot-end` | `^TestParserNodeArrayLayout$` |
| `slot-hasTrailingComma` | `^TestParserNodeArrayLayout$` |
| `slot-transformFlags` | `^TestParserNodeArrayLayout$` |
| `drop-unset-use` | `^TestParserConstructionUnsetUse$` |
| `drop-speculation-check` | `^TestParserSpeculationMisfit$` |
| `disable-tiny-stack` | `^TestParserStackTinyLimit$` |
| `drop-flow-check` | `^TestPlaceholderUseCheckSavedLeak$` |
| `retain-alias-proof` | `^TestPlaceholderUseCheckAliasReset$` |
| `nullish-receiver-twice` | `^TestNativeAgreesWithNode_[0-9]+$/internal/oracle/testdata/placeholder_nonnull_null_loose[.]a$` |
| `json-ordinary-assertion` | `^TestNativeAgreesWithNode_[0-9]+$/internal/oracle/testdata/placeholder_nonnull_null_json[.]a$` |
| `admit-weak-slot` | `^TestPlaceholderWeakSlotStaysNotYet$` |
| `omit-refusal-state` | `^TestOptionalWideningSpreadOverwrite$` |
| `drop-native-nullish-tags` | `^TestNativeAgreesWithNode_[0-9]+$/internal/oracle/testdata/placeholder_nonnull_null_copy[.]a$` |
| `drop-javascript-union-tag` | `^TestDefaultTaggedSourceViewsDefaultBoxedWrite$` |
| `unlisted-output` | `unlisted backend divergence: stdout differs` |
| `drop-terminal-guard` | `exit codes differ` |
| `drop-range-error-kind` | `exit codes differ` |
| `drop-overflow-message` | `non-overflow RangeError was terminal: exit codes differ` |
| `skip-metadata-sharing` | `^TestParserNodeArrayMetadataSharing$ TestParserNodeArrayMetadataSharing: metadata child not shared` |
| `retain-interior-metadata` | `^TestParserNodeArrayMetadataSharing$ TestParserNodeArrayMetadataSharing: interior metadata retained` |

All 29 independent mutant catchers exited nonzero through the intended assertion or behavior, not a compiler/build failure. omit-refusal-state specifically checks nil-result initialization via a Go panic. Embedded source mutants include publication/key/JSON, wrong factory payload, rewind, stack guard, flow-check removal, null-tag, spread readiness, wrong-zero/erasure/capture/exception/missing-value/message changes and eight negative-witness repairs. Their focused assertions require the intended disagreement.

Toolchain and recovery

The first conflict resolver's closing-marker regex was greedy and truncated file tails. Inspection caught this; all exact three-way inputs were restored from git objects, the regex was corrected, and every hunk was reapplied. The first setup consequently failed on unexpected EOF in JavaScript/native Go files and is not credited. The repaired setup succeeded with ADAMIC_GOCACHE_OFF=1. setup-observations.json records the complete timing lines.

- setup: node ready (0.020s)
- setup: go ready (0.021s)
- setup: submodules ready (0.058s)
- setup: markdown dependencies skipped (validated lock and installed bytes); step-duration=0.006s
- setup: markdown dependencies ready (0.064s)
- setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0.156s)
- setup: shared cache off (ADAMIC_GOCACHE_OFF=1) (0.158s)
- setup: go build ready (68.639s)
- setup: test binaries deferred (use --warm-tests) (68.780s)
- setup: build cache warm (68.782s)
- setup: build-flags commit=dbdf977f5b8b0979d414ebe6f20e7aadd312e309 nproc=5 cpu.max=400000 100000 go=go version go1.27.1 linux/amd64 clang=clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261) node=v24.19.0 cached=yes warm-tests=false gocacheprog=off load-before=0.42 0.45 0.18 1/144 1854 load-after=8.98 3.10 1.13 2/156 3500
- setup: done on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB (68.808s)
- setup: source /workspace/adamic-tools/env.sh
- setup: logs /tmp/adamic-gate/setup.DwmvrA

nproc=5; cpu.max=400000 100000 (four CPU quota). Go 1.27.1, Node 24.19.0, clang 20.1.8. GOPROXY was exported as https://proxy.golang.org|direct before setup. Subsequent tests unset GOCACHEPROG for the local Go cache.

Not covered: whole-package suites, full gate, ahra/Loom submission, admission census, pending Program-region lifetime, opt-in guard-cost measurement, WASI or non-Linux systems. Consumer prerequisites and the observed CSS budget red remain explicit handoff limitations. Runtime line changes require the user's runtime-clearance route; no implicit clearance is claimed.
