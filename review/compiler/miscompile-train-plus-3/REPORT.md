Built: restored ordinary-source spread reuse after final readiness removes unnecessary member checks, keeping checked-view checks before copying.
Commits: fix 6a056842ff91f8d260aa086675050a61a5e2f2f4; counts/evidence b0965b7bd57f317e5cb302d938f493a0fd07059c; base 51aa3a9636d141c5dd24394cd4bba4c85204b099; delivery tip is reported in the final response.
Commands/results: native and lower full shards, Node/backend train fixtures, reader guard, counts, vet and committed-tip lane checks pass; counts regenerated once in 54.666s.
Mutants: exact fix revert fails TestInheritanceMemoryPlans; removing member checks fails the 132 spread/operation family and forward-certificate witness. Both caught, no new survivor.
Limits: shared full gate not run; native has 86 existing configured skips, lower has two existing skips; historical equivalent single-filter and creation-masked method-spread survivors remain named and were not part of this turn's two requested mutations.

This lands the requested ownership regression fix on the views miscompile train, retaining 132's checked spread boundary and rule 142's creation admission. The specified base takes precedence over the generic start-from-main instruction. No protected branch, PR, compiler/runtime emitter change, cohere copy or new permanent .ts source was made.

239a3a8f unconditionally wraps every object spread source with checked_view_members, including ordinary typed sources. After readiness clears View/Readiness on its ordinary slot reads, the identity call survives. native.variableRead only accepts ir.Read (optionally ir.Defined), so spreadsOf cannot see the source and planReuse cannot establish the consumed parameter. This is the exact loss in the reported test; NoReuse itself was not newly set for this fixture. The baseline reproduces the reported failure at class_inheritance_test.go:37.

The fix waits until the entire readiness pass completes, including all functions and late view certificates. It recognizes only a one-parameter checked_view_members helper whose body consists of unchecked own-slot reads (including optional representation unwraps), an optional undefined-source guard, and an identity return. A surviving view or initialization check, method read, other statement, nonidentity result, virtual call or expanded argument blocks removal. Replacing the call with its original operand evaluates the source exactly once. Existing NoReuse flags, liveness, borrowing, weak-reference, frozen-object and uniqueness guards remain in charge. Checked-view helpers remain intact and copy only after their checks succeed; no checked-view reuse optimization is introduced.

The inheritance fixture has exactly one source spread: Reader.replace's {...value, text: ...}. Its Item parameter is ordinary, with no view assertion or initialization boundary. Its helper has no remaining check, so the original parameter read is restored and the existing uniqueness-tested spread plan can reuse it. ChildReader.replace has no spread. The strengthened test identifies this one planned spread and still verifies consumed-convention joining across Reader/ChildReader and escaping arguments across Reader/SavingReader. Its final standalone run passes in 0.04s. Runtime aliases can still require copying; a planned reuse is conditional on uniqueness.

New forward-declaration witnesses put the copy function before the view assertion. The invalid count remains a checked failure in JavaScript, release native and sanitized native with a source Node control; its valid version agrees with Node. This protects against the unsound shortcut of deciding whether a receiver is checked during syntax lowering. The two new top-level parallel leaves initially passed in 0.48s and 0.49s, including setup. The final train-fixtures log contains their repeated passing results.

The exact revert mutant substitutes both changed production files with their 51aa3a96 contents, while leaving the strengthened test active. TestInheritanceMemoryPlans fails with the original lost-every-reused-spread message (mutant-revert-fix.log; 2.425s command). The member-removal mutant returns the original value immediately from checkedViewMembers. It fails TestCheckedViewSpread and TestCheckedViewSpreadForwardCertificate, plus TestCheckedViewIn, Keys, KeysAlias, Values and Entries, on missing runtime refusals (mutant-drop-member-check.log; 22.746s command). p01 and its ordinary read siblings remain covered by the same TestCheckedView selection and the restored train suite. Both mutations were restored before any passing package shards, counts or delivery validation. Patches and exact commands are in mutants.json/run-mutants.py.

Every listed native/lower test was assigned to exactly one shard (shards.json). Native: 285 top-level leaves, 199 pass and 86 existing skips, 16 shards; maximum command 71.395s. Lower: 361 leaves, 359 pass and two existing skips, four shards; maximum command 36.098s. Each Go shard has -timeout 80s and a subprocess hard limit of 89s. Native's longest passing leaf is TestDecodeASCIIUnit16 at 43.79s; lower's is TestPredicateOverloadRuntime at 9.57s. Native skips request the opt-in WASI configuration, a built checker archive or opt-in clang measurement; lower's skips are TestOriginalCycleLedger and TestOptionalWideningCensus. package-coverage.json preserves every skip name.

Train fixture command (train-fixtures.log; oracle 54.127s):
`timeout 89 go test ./internal/oracle -run 'TestCheckedView|TestRequiredView|TestNarrowedField|TestViewField|TestDefaultTagged|TestFX[67]|TestScalarUnion|TestCallableProducer|TestCallableUnionResult|TestViewSnapshot|TestReviewPrograms(AgreeWithNode|Refuse)/fxspptb_oct9_(views|native)_' -count=1 -v -timeout 80s`
This covers candidate 3's view/scalar/callable probes, unsupported-view admission, snapshots, depth/failure text and callable union results. Candidate 4's p18/p20/wide, tuple p68/p69, bad-callable before-argument check, acceptance p17/p18/p20 and all cleanup diagnostic fixtures are also exercised by full lower shards and registered/review suites. p26/p33/p34/p35 remain precise creation refusals in .a and ephemeral .ts paths. No admission guard was weakened.

Remaining registered train fixtures and class_inheritance_memory pass TestNativeAgreesWithNode with source Node, JavaScript, release native, sanitized native and successful-program leak checks (registered-fixtures.log; oracle 3.045s). The exact selection is recorded in the log command through this report's commands file below. stage3 TestFixturesAssertions passes in 11.308s. Two explicit supplemental fixtures, review p19's formerly named call gap and p70_copy, agree with Node in both backends; six refusal sources have successful Node controls and matching c/js compile refusals, including p70, p135_object_result, unsupported p37 and all three cleanup diagnostic sources (extra-fixtures.json). All 36 registered changed-count fixtures additionally pass Node/backend agreement; 35 ordinary oracle fixtures take 5.608s and tree takes 0.513s. The two other moved rows are the snapshot fixtures already covered above.

Counts: exactly one TestCountsAreRecorded -update-counts invocation, exit 0, 54.666s. No added or removed rows, 38 changed rows. Every old and new row is reproduced by independent counted CLI executables from the base and fixed code. All have a positive decrease in checked_view_members call sites in emitted C. count-attribution.json records six-column before/after/delta, direct measurements, helper-call removals, new uniqueness guards and cause commit 6a056842. Each corresponding C source and diff is saved under count-attribution/. No table value was picked by hand. undefined-read-write.a remains at four retains: the train's snapshot reference-tag guard is unchanged.

The retain increases are accounted for, rather than assumed to be noise. borrow_loop (+1) now retains the borrowed loop element when handing it to consumed grown; reuse_lent_global (+1) holds the global's alias while emptied consumes a count; throw_global_move (+2) restores the consumed-spread transfer and global hold across the throwing/moving path. class_inheritance_memory (+2) and class_instance_key_consume (+1) regain joined consumed conventions and counted argument handoffs through virtual replacement calls. Shared live aliases force the fallback copy, explaining unchanged allocations in these five rows. Their emitted C exposes the added handoffs/releases and uniqueness guards, and direct base/fixed measurements and Node/backend checks agree. The two snapshots lose three copies, three frees/releases and one peak-live value each because their ordinary source spread is now reusable on each of three calls; the later checked snapshot read remains active.

Every moved row follows. Columns A/F/R/L/P/G mean allocations/frees/retains/releases/peak/regions. I means removal of no-check identity helper temporary ownership; R means the same removal also restores ordinary spread reuse/consumed conventions. All rows are attributed to fix commit 6a056842 and directly measured on both compilers.

| Fixture | Before A/F/R/L/P/G | After A/F/R/L/P/G | Helper calls removed | New uniqueness guards | Cause |
|---|---|---|---:|---:|---|
| internal/oracle/testdata/route_targets_virtual_fresh.a | 56/52/22/60/10/4 | 54/50/20/64/10/4 | 1 | 1 | R |
| internal/oracle/testdata/route_targets_sort_values.a | 84/84/99/149/17/0 | 84/84/92/142/17/0 | 2 | 0 | I |
| internal/oracle/testdata/borrow_chain_coverage_spread.a | 12/12/4/11/7/0 | 12/12/3/10/7/0 | 1 | 0 | I |
| internal/oracle/testdata/call_targets_reuse.a | 16/13/10/19/7/3 | 15/12/9/21/6/3 | 1 | 1 | R |
| internal/oracle/testdata/call_targets_sort.a | 11/11/10/16/7/0 | 11/11/9/15/7/0 | 1 | 0 | I |
| internal/oracle/testdata/library_object_freeze.a | 16/16/11/28/5/0 | 16/16/10/27/5/0 | 1 | 1 | R |
| internal/load/testdata/0.1/compile/09_tree.ts | 33/33/73/76/16/0 | 19/19/28/69/16/0 | 2 | 2 | R |
| internal/oracle/testdata/maybe_number_slots.a | 484/484/313/806/16/0 | 484/484/304/797/16/0 | 1 | 0 | I |
| internal/oracle/testdata/spread_snapshot.a | 22/22/4/25/11/0 | 22/22/3/24/11/0 | 2 | 0 | I |
| internal/oracle/testdata/reuse.a | 85/85/75/136/16/0 | 76/76/55/120/16/0 | 6 | 4 | R |
| internal/oracle/testdata/reuse_global_sibling.a | 6/6/3/8/4/0 | 6/6/2/7/4/0 | 1 | 1 | R |
| internal/oracle/testdata/reuse_weak_during_spread.a | 10/10/9/17/8/0 | 10/10/8/16/8/0 | 1 | 1 | R |
| internal/oracle/testdata/reuse_weak_after_reuse.a | 9/9/7/14/7/0 | 9/9/6/13/7/0 | 1 | 1 | R |
| internal/oracle/testdata/borrow_element.a | 113/113/69/130/15/0 | 113/113/67/128/15/0 | 3 | 1 | R |
| internal/oracle/testdata/borrow_loop.a | 128/126/109/180/14/2 | 128/126/110/181/14/2 | 1 | 1 | R |
| internal/oracle/testdata/throw_keeps_old_value.a | 18/18/10/22/6/0 | 18/18/9/21/6/0 | 2 | 1 | R |
| internal/oracle/testdata/throw_keeps_old_value_variants.a | 26/26/13/34/6/0 | 26/26/11/32/6/0 | 2 | 0 | I |
| internal/oracle/testdata/throw_in_writes.a | 22/22/10/24/8/0 | 22/22/8/22/8/0 | 2 | 0 | I |
| internal/oracle/testdata/throw_global_move.a | 18/18/10/24/6/0 | 18/18/12/26/6/0 | 2 | 2 | R |
| internal/oracle/testdata/spread_undefined.a | 47/47/34/72/10/0 | 45/45/26/64/10/0 | 5 | 2 | R |
| internal/oracle/testdata/reuse_lent_global.a | 10/10/7/14/7/0 | 10/10/8/15/7/0 | 1 | 1 | R |
| internal/oracle/testdata/reuse_spread_method.a | 10/10/8/13/8/0 | 10/10/7/12/8/0 | 1 | 0 | I |
| internal/oracle/testdata/reuse_spread_method_alias.a | 8/8/8/15/6/0 | 8/8/7/14/6/0 | 1 | 0 | I |
| internal/oracle/testdata/class_features_private.a | 47/47/31/59/15/0 | 47/47/29/57/15/0 | 6 | 0 | I |
| internal/oracle/testdata/class_inheritance_exceptions.a | 30/30/20/40/8/0 | 29/29/19/39/8/0 | 1 | 1 | R |
| internal/oracle/testdata/class_identity.a | 12/12/7/18/7/0 | 11/11/6/17/7/0 | 1 | 1 | R |
| internal/oracle/testdata/class_inheritance_memory.a | 48/42/18/56/11/6 | 48/42/20/58/11/6 | 1 | 1 | R |
| internal/oracle/testdata/class_instance_key_consume.a | 8/7/4/11/6/1 | 8/7/5/12/6/1 | 1 | 1 | R |
| internal/oracle/testdata/entries_checked_fit.a | 54/54/25/47/12/0 | 54/54/20/42/12/0 | 5 | 0 | I |
| internal/oracle/testdata/entries_checked_misfit.a | 7/1/4/2/6/0 | 7/1/3/2/6/0 | 1 | 0 | I |
| internal/oracle/testdata/entries_checked_boolean.a | 18/18/7/20/7/0 | 18/18/5/18/7/0 | 2 | 0 | I |
| internal/oracle/testdata/entries_checked_literal.a | 6/0/6/1/6/0 | 6/0/5/1/6/0 | 1 | 0 | I |
| internal/oracle/testdata/entries_checked_boxed.a | 40/40/16/40/11/0 | 40/40/12/36/11/0 | 4 | 0 | I |
| internal/oracle/testdata/fallthrough_ownership.a | 74/74/27/76/7/0 | 71/71/24/76/7/0 | 2 | 2 | R |
| internal/oracle/testdata/notyet_library_object_entries_const.a | 40/40/38/37/23/0 | 40/40/36/35/23/0 | 2 | 0 | I |
| internal/oracle/testdata/unknown_narrowing.a | 31/31/43/64/3/0 | 31/31/42/63/3/0 | 1 | 0 | I |
| stage3/interface-downcasts/v2/snapshot-maybe-boolean.a | 9/9/9/27/3/0 | 6/6/9/24/2/0 | 1 | 1 | R |
| stage3/interface-downcasts/v2/snapshot-maybe-number.a | 13/13/9/24/5/0 | 10/10/9/21/4/0 | 1 | 1 | R |

Setup succeeded with GOPROXY=https://proxy.golang.org|direct. Timing lines: Node ready 0.024s; Go ready 0.031s; markdown dependency validation 0.007s, ready 0.072s; submodules ready 0.115s; clang ready 0.213s; Go build ready 39.953s; test binaries deferred 40.112s; build cache warm 40.113s; done 40.143s. nproc=5, cgroup quota four CPUs. Every final Go/Node command sources /workspace/adamic-tools/env.sh.

The isolated detached base CLI build hit its 89s bound without compiler diagnostics; a Go overlay of the exact base production files on the current checkout reused dependency caches and built successfully. No baseline source was changed. An initial new console fixture used a numeric console.log argument and was corrected to the supported string argument before any final test runs. An optional extra-fixture audit initially omitted p135's existing producer refusal from its expected refusal wording; its corrected audit passes. An initial changed-fixture command used the wrong shell toolchain, then a whole-path alternation that selected no subtests; the component-wise corrected commands explicitly assert all 35 plus one registered subtest leaves ran and pass. These preliminary outputs are retained and are not counted as passing evidence.

Final reader command: timeout 89 go test ./internal/ir -run '^TestCallTargetReaders$' -count=1 -v -timeout 80s, PASS 11.036s. Standalone inheritance command: timeout 89 go test ./internal/native -run '^TestInheritanceMemoryPlans$' -count=1 -v -timeout 80s, PASS 0.044s package. Vet: timeout 89 go vet ./internal/lower ./internal/native ./internal/oracle ./internal/ir ./internal/javascript, exit 0 with no output. Counts: timeout 89 go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 80s -args -update-counts, PASS. Exact full-shard, changed-fixture and supplemental commands are in shard-results.json, changed-fixtures-command.json, extra-fixtures.json and their runners.

Committed-tip lane command:
`timeout 180 bash -c 'git fetch -q origin main devtools/fast-gate cloud/merge-tree && git show origin/cloud/merge-tree:cloud/integration/lane-checks.py | python3 -'`
Initial committed counts/evidence tip: b0965b7bd57f317e5cb302d938f493a0fd07059c
lane checks 7.7 s: gofmt and tools on 71 Go files, t.Parallel on 4 test packages; a-check 13 .a files; vet 4 packages
Remote lane/tool refs were explicitly refreshed first. The same lane command and vet are rerun on the final evidence commit before the single delivery push. The shared whole gate is integration's and was not run.
