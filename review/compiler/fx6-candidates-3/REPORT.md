Rebuilt candidate 3 with the exact five-member recipe and constructor own-slot fix for tasks #1fk58py and #aecx10a.
Base 708ec9c9; member tips b1d065ba, 32b8583f, a030317e, d0fa7138; constructor fix da68e17e; delivery SHA is reported with the push.
Full lowering, checked-view oracles, fx6/fx7 fixtures, stage3 assertions, reader guard, counts and integration lane checks are recorded in this directory.
Independent member revert mutants are run by run-mutants.py, with each mutation restored and its behavioral catcher saved separately.
Full repository gate and full native/oracle packages are outside this worker verification; compiler/fx7-wrong-aborts is excluded.

This advances checked member-read routing and candidate integration toward the views work in #1fk58py and #aecx10a. The explicit recipe authorizes these member merges despite the general unit rule against merging unlanded worker branches. No tests or expectations were edited. No cohere code was copied.

| Member | Contribution |
|---|---|
| compiler/fx6-candidates-2-fix, 708ec9c9 | Receiver-keyed field certificates, nullable and generic normalization, and common union receiver boundaries. Keeps candidate-2 valid programs and tuple-object refusal. |
| compiler/fx6-key-read, b1d065ba | Checked finite element reads and destructuring; spread, in, keys including aliases, values and entries share the own-member boundary; p19/p72 witnesses and empty-key/method guards. |
| compiler/fx7-scalar-union-view, 32b8583f | Checked maybe-number and maybe-boolean extraction from scalar union views, with string and transition controls. |
| compiler/fx7-callable-producers, a030317e | Assignable callable producer certificates through closure thunks, native registry filtering, scalar result/ABI bounds, explicit adapter refusals, and lazy unread tagged methods. |
| compiler/fx7-name-family, d0fa7138 | Six regression witnesses prove same-named unrelated undefined/null stores and reads remain ordinary; receiver identities prevent global name contamination. |
| Constructor fix, da68e17e | Excludes checker constructor built-ins from represented own data slots while retaining parent lookup for declared static fields. |

The only merge conflict was internal/lower/view_member_read.go. Retained viewReceiverTypeID and checkedViewReceiverField from candidate-2-fix and appended checkedViewMembers from key-read, closing each function separately. The final file's blob is 7674e6959c240fce0a2d3b2596392b0e6034e88e, exactly the recipe's final blob prefix. Counts merged automatically with the union intact.

Initial main was 955e3eb9, newer than recipe main 92d19601. Its added lint test shards merged cleanly. The later required lane fetch discovered main 38c09330, containing only stage1 JSON test changes; delivery integrates that tip. No added main change conflicts with the compiler stack.

Setup used GOPROXY=https://proxy.golang.org|direct. Initial timing lines: Go 0.094s, Node 0.112s, clang 0.628s, markdown 2.203s, submodules 30.390s. Cache warming failed with missing readViewMember/checkedViewMembers methods while merges were in progress. Inference: its earlier source census omitted the newly merged file. Stable-tree rerun passed: Node 0.044s, Go 0.052s, submodules 0.104s, markdown 0.167s, clang 0.236s, build 133.365s, cache warm 134.252s, total 134.464s. Environment file /workspace/adamic-tools/env.sh; nproc 5; cpu.max 400000 100000. npm ci in stage3/api installed the repository-pinned @types/node 25.3.3 and TypeScript 6.0.3.

Initial baseline failures are retained, not claimed green: the cold full-lowering command exhausted its outer limit; the subsequent lowering run and one checked-view oracle test reported missing Node types. After installing the prerequisite, full lowering passed in 123.421s and TestCheckedViewV2ReadsAfterWrites passed in 30.760s. The other selected checked-view oracle tests passed in the earlier 133.667s package run; its only failure was fs-option-boxing's prerequisite. The initial registered fixture regex selected no tests and was corrected. The actual fx7 native fixture command passed in 115.968s; TestCallTargetReaders passed in 30.154s; stage3 TestFixturesAssertions passed in 25.989s. Exact commands are in commands.txt, and raw outputs are individual logs.

Only four existing counts rows moved relative to the merged members. TestCountsAreRecorded measured all rows, with only these differences; counts.md was updated from those measured lines rather than regenerating unrelated rows. A/F/R/L/P/G means allocations/frees/retains/releases/peak/regions.

| Row | Before | After | Cause |
|---|---|---|---|
| nbody_static_collision.a | 33/24/15/41/13/0 | 43/43/17/57/19/0 | Constructor Object.keys reaches its expected output and cleanup instead of stopping at the nonexistent built-in slot. |
| field_write_paths.a | 43/38/10/50/8/0 | 46/46/11/60/8/0 | Final inherited-static Object.keys completes and releases its live values. |
| class_features_static.a | 51/41/12/55/14/0 | 110/110/55/162/28/0 | Static constructor enumeration completes; subsequent static blocks, generics, exceptions and aliases execute. |
| class_features_static_private.a | 10/4/15/21/6/0 | 42/42/53/94/12/0 | Constructor enumeration skips built-ins/private storage; following brand-error catches and private accessor operations execute. |

These are restored execution/cleanup counts, not allocation optimizations. Each final row has allocations equal to frees. The four sources were inspected to identify the enumeration site and subsequent execution.

Mutation evidence maps each member obligation to its intended catcher. Every mutation is restored before the next. The historical single-filter direct-producer-certificate revert was masked by the newer untaggedCallableABI filter and passed; direct-producer-certificate-masked.log preserves it. The adapted member revert removes both redundant producer filters and is caught by the certificate assertion. The callable and assignable member certificate obligations overlap on this merged stack; they are separate member runs of the same two-filter revert. The duplicate native-filter obligation is run once. These account for the requested 29 member mutant runs without claiming 29 distinct source edits.

| Mutant | Catcher |
|---|---|
| nullable-receiver | TestCheckedViewOptionalReadBoundary and optional-receiver: optional checked read escaped |
| generic-receiver | TestCheckedViewUntaggedSourceFlows/generic/wrong: exit codes differ |
| union-receiver | Stage3 assertions/19_identifier_kind.a/stage0: gap changed |
| revert-receiver | TestFX6P37: JavaScript backend stdout differs |
| revert-destructure-type-id | TestFX6P53: JavaScript backend stdout differs |
| revert-p70 | TestTupleObjectViewRefused: got nil instead of located refusal |
| element-bypass | TestCheckedViewElementP05: exit 0 instead of ruled exit 70 |
| destructure-bypass | TestCheckedViewDestructuredUnion: exit 0 instead of ruled exit 70 |
| skip-conversion-check | TestScalarUnionViewMisfitNumber/Boolean: native and sanitized exit 0 instead of 70 |
| unfiltered-native | TestCallableProducerRegistryFiltersDirectFunctions: emitted registry contains comparison of distinct pointer types |
| direct-producer-certificate, adapted | TestCallableProducerCertificatesUseClosureThunks: producer certificate includes direct function |
| assignable-direct-certificate | TestCallableProducerCertificatesUseClosureThunks: producer certificate includes direct function |
| exact-identity | TestCallableProducerLiteralReturn: JavaScript backend stdout differs |
| skip-adapter-refusal | FewerParameters/MethodShorthand/ExtraOptional refusal tests: got nil |
| skip-result-registry-bound | TestCallableProducerDiscardedObjectResultRefused: got nil |
| eager-tagged-callable | TestCallableProducerTaggedUnreadMethod: Lower refused acceptance row |
| 147-name-wide-undefined-write | TestFX7P04/P06: native stdout differs |
| 148-name-wide-read | TestFX7P75/P77: backend stdout differs |
| 149-name-wide-null-write | TestFX7P08/P59: native stdout differs |
| final-spread | TestCheckedViewSpread: exit 0 instead of 70 |
| final-in | TestCheckedViewIn: exit 0 instead of 70 |
| final-keys | TestCheckedViewKeys: exit 0 instead of 70 |
| final-keys-alias | TestCheckedViewKeysAlias: exit 0 instead of 70 |
| final-values | TestCheckedViewValues: exit 0 instead of 70, literal member contract bypassed |
| final-entries | TestCheckedViewEntries: exit 0 instead of 70, literal member contract bypassed |
| p19-read | TestCheckedViewElementP19Read: exit 0 instead of 70 |
| p72-input | TestCheckedViewElementP72: compatible string input makes expected misfit stop disappear |
| in-empty-selector | TestViewInEmptyKeyControl: JavaScript stdout differs because an unrelated field is checked |
| spread-method | TestViewSpreadMethodRefused: got nil instead of own-slot refusal |

The native registry mutant is caught by an explicit emitted-C assertion; a compiler/clang build failure is not counted as its kill. Syntax misfit mutants must report runtime exit-code disagreement, with no sanitizer failure. p72-input is an input control substitution, not a missing-check mutation. Original masked evidence is distinguished from successful adapted checks.

Remaining limitations inherited from members: native callable-union numeric-result boxing for the separate compatible p19 invocation probe; checked method enumeration without an own-slot certificate; tuple/destructuring consumers outside the completed object bindings; existing optional-read NotYet boundaries; general callable adapters beyond supported scalar ABI conventions. Full repository gate was not run. No new test leaves were added by this rebuild. One inherited p135 native leaf took 79.12s while cold concurrent baselines were running; a final isolated check records its warm duration.
