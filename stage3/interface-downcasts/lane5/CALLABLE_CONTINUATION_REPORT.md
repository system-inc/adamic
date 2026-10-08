Built: six groups certify 17 additional callable pairs, 282 candidate reads and 91 fixtures through original viewed member reads.
Commits: 685888b2, 86832529, baa72933, 69883a0e, c0e85282 and this commit, pushed separately to codex/views-callables.
Checks: original tsc declarations/read spans, Node, release/sanitized native, JavaScript, leaks, measured lane counts, restored later-ranked tests and oracle vet pass; whole-table counts remains blocked outside the lane.
Mutants: all 36 runs caught and restored, 91 pair-level checks; arity, result, parameter and reached payload assertions provide evidence.
Uncovered: 62/2818 pairs and 1875/11063 candidate reads certified; 2756 pairs and 9188 reads remain; original union, method-value and intrinsic families remain uncertified.

Reporting date October 12. Continued from 619b65937ce3104353b0d9da570dc85e8545f524. Original source pin 050880ce59e30b356b686bd3144efe24f875ebc8. Candidate read counts are conservative inventory counts, not exact production execution coverage.

| Group | Commit | Certified fixtures | Mutant runs / pair checks | Restored oracle | Lane counts |
|---|---|---:|---:|---|---|
| helper-hoist-modifiers | 685888b2 | 17 | 7 / 17 | 8.296s | 39.895s |
| scanner-end-full-start | 86832529 | 8 | 4 / 8 | 3.444s | 42.116s |
| export-computed-source-diagnostic | baa72933 | 25 | 7 / 26 | 12.522s | 43.863s |
| source-path-resolver | 69883a0e | 12 | 7 / 12 | 5.664s | 48.553s |
| writer-line-watcher-close | c0e85282 | 8 | 4 / 6 | 3.702s | 45.907s |
| logical-named-string-scanner | this commit | 21 | 7 / 22 | 10.369s | 47.012s |

Certified ranks: 91, 92, 94, 95, 96, 98, 99, 100, 101, 103, 107, 109, 110, 115, 116, 117 and 118. Each group retains complete original declarations and member read expressions, with reduced adjacent structural carriers/helpers. Optional arguments, supplied aggregate arrays, scalar/string results and reached descendant fields are covered where applicable. The returned modifier/source-file carriers and passed statement/export-specifier elements stay lazy.

Set ranks 89/90 are skipped as instructed. They need supported intrinsic collection-to-view conversion and recorded callable signatures, including the add self result. Existing probes observe the native adamic_map*/adamic_object* boundary and JavaScript unknown-signature refusal. The requirements are recorded in HELPER_HOIST_MODIFIERS_REPORT.md. Intrinsic Map/SymbolTable, RegExp and array members encountered between supported object members remain deferred. Existing Map-method refusal evidence is in ORIGINAL_GROUP_REPORT.md; applying the Set boundary to other intrinsic receivers is conservative inference, not certification of their original pairs. No plain-object producer is counted for an intrinsic pair.

Three reduced original-read witnesses are separately pinned against Node and lowering: rank 102 createLiteralTypeNode reaches an unsupported untagged object-union field read; rank 106 preserves !system.getEnvironmentVariable and refuses unbound-method; rank 120 createArrowFunction retains its six parameters and ConciseBody union, and refuses the reached body.value read. The union witnesses need proved union representation and descendant-read contracts; the environment read needs an agreed method-value convention. No narrower alias, synthetic tag, property-style declaration or arrow replacement is used to certify these witnesses. They are excluded from native counts and pair totals.

All exact per-group commands, outcomes, mutant names and direct logs are in the six group reports and logs directories. The final verification also ran:

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewCallableLaterRanked' -count=1 -timeout 5m > /tmp/lane5-logical-all-restored.log 2>&1
go vet ./internal/oracle > /tmp/lane5-logical-vet.log 2>&1
gofmt -l internal/oracle/checked_views_callable_later_ranked_test.go internal/oracle/checked_views_callable_counts_test.go
git diff --check
```

ok  	github.com/system-inc/adamic/internal/oracle	85.820s

The complete later-ranked run includes all previous supported members and retained refusals. Oracle vet has an empty successful log. All 91 new counts rows are appended, with existing rows audited unchanged per group. Each required TestCountsAreRecorded updater was run; each failed on existing outside-lane cases. The lane-specific updater passed after each group. No whole-package tests or full gate were run.

Source verification now records 40 original declarations/read spans covering 754 candidate reads, and verifies all 190 retained original-witness fixtures, including retained refusal witnesses and previously supported members. Static certification remains 4/308 pairs and 34/1503 reads, with 304 pairs and 1469 reads remaining. All production compiler/runtime files are restored; no calling-convention guard changed, and no new integrator decision on the Union callable exception was received.

Setup reused from the preceding session: Node .060s, Go .062s, clang .498s, markdown .885s, submodules 177.790s, Go build 417.304s, cache 417.400s, done 417.425s. nproc 5, quota 4 CPUs; Go 1.27.1, clang 20.1.8, Node 24.19.0. No PR opened and only codex/views-callables was pushed.
