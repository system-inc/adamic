Built: #41bkfdw wave 3 moves 65 source rows onto shared Node agreement, including five internal-fact rows.
Commits: unit 81b9df15; evidence 37ec839d; current-main merges a28550fc and d6032098, onto origin/main 7911f07b.
Checks: named-file tests PASS 5.502s; TestCallTargetReaders PASS 0.643s; lane checks PASS 4.9s.
Mutants: arguments-count and finite-field both fail converted tests on stdout, while original tests pass.
Uncovered: 11 acceptance conversions skipped; 3 internal-fact conversions skipped; native-only erasure cost is no longer asserted.

The census covers every source row, rather than call sites: 302 total, 71 accept, 198 refuse, 33 IR-only. Sixty acceptance rows and five internal-fact rows now execute through the JavaScript backend and source Node. Eleven acceptance rows retain their original checks with reasons in rows.md. All refusal rows remain. No new oracle fixtures were added, so no unit counts.md refresh was needed.

| file | accept | refuse | IR-only |
|---|---:|---:|---:|
| array_predicate_test.go | 10 | 11 | 0 |
| proven_relations_test.go | 1 | 15 | 2 |
| arguments_length_test.go | 5 | 25 | 1 |
| interface_cast_test.go | 13 | 3 | 0 |
| definite_assignment_test.go | 4 | 0 | 4 |
| census_small_test.go | 2 | 5 | 0 |
| predicates_proof_test.go | 6 | 17 | 25 |
| lower_test.go | 30 | 116 | 1 |
| census_predicate_interaction_test.go | 0 | 4 | 0 |

The awkward rows are malformed checked-interface reads, which intentionally differ from unchecked source; unused generic/intersection bodies whose original checks permit NotYet; and the function-field method neighbor, whose observed NotYet says the view erases its prototype origin. The console stream/Source probe remains an explicitly skipped internal-fact conversion because the common helper rejects nonempty success stderr. Uninitialized definite-assignment rows remain readiness probes because source computes NaN while checked output panics. Proof summaries and per-call use directions remain internal facts, with comments naming what behavior cannot observe. Relation artifact equality snapshots were removed; these rows now hold operand behavior without claiming runtime cost erasure.

No local oracle/node.mjs runner copies occurred in the original eight named files. The added census_predicate_interaction_test.go had four refusal rows and a local source Node command. That command now uses the same bounded runAgreementNode implementation as the acceptance helper. It cannot use lowersAndAgreesWithNode itself because its required outcome is Refused; that helper would fatal before observing source Node. Its source declaration output, exit, stderr, exact predicate refusal and three escape refusals remain checked. The example files outside the assigned territory were left to their units. No production compiler change remains. No cohere source was copied. The conservative assumption was to preserve partial admission and intentional checked-failure contracts as documented skips, rather than claim source agreement for them.

Mutant evidence:

| one-line lowering mutant | converted catcher | converted result | original result |
|---|---|---|---|
| arguments_length.go returns constant 9 instead of the argument-count read | TestArgumentsLengthReadNeighbors | FAIL 0.135s: backend `9\n`, source `0\n` | PASS 0.164s |
| element_access_fields.go fits constant wrong instead of the computed property read | TestCensusSmallFiniteKeyRead | FAIL 0.210s: backend `wrong\n`, source `file\n` | PASS 0.040s |

The .diff files and all four logs are beside this report. Each compiler file was restored after the mutant. The original tests were temporarily restored from the base version for the counterfactual run. Neither behavior mutant was killed by compilation or an unrelated guard.

A third mutant changes the bodyless-predicate refusal in predicates.go to return nil. TestCensusPredicateInteractionBoundary and all three TestCensusPredicateInteractionEscapes rows fail their pinned diagnostics with got <nil>, in both the shared-runner and original versions. Shared-runner mutant result: FAIL 0.116s; original: FAIL 0.104s. This proves the runner refactor preserves the predicate guards. predicate-bodyless.diff, predicate-mutant.json and both logs preserve the evidence. Lowering was restored.

Exact validation commands are in commands.txt, with the full 35-test selection in test-filter.txt. Test output always went directly to logs. The post-merge commands used the same filter and -count=1 -timeout 90s -v. leaf-seconds.tsv records every passing test and subtest from the post-merge run; the longest recorded test and each leaf are recorded in leaf-seconds.tsv. No new top-level test functions were added. Full packages and the full gate were not run.

Integration lane command:

```sh
source /workspace/adamic-tools/env.sh
export GOPROXY='https://proxy.golang.org|direct'
timeout 60 git fetch -q origin main devtools/fast-gate cloud/merge-tree && git show origin/cloud/merge-tree:cloud/integration/lane-checks.py | timeout 180 python3 -
```

Output: `lane checks 4.9 s: gofmt and tools on 8 Go files, t.Parallel on 1 test packages; vet 1 packages`.
The checkout only tracks main by default, so explicit refspecs were needed once to populate origin/devtools/fast-gate and origin/cloud/merge-tree. No main or area branch was pushed.

Toolchain setup passed. nproc: 5, cgroup quota: 4 CPUs. Node ready 0.025s; Go ready 0.040s; clang ready 0.228s; markdown dependencies ready 1.005s; submodules ready 17.990s; Go build ready 229.060s; cache warm 229.162s; done 229.190s. The emitted environment path was /workspace/adamic-tools/env.sh; /opt/adamic-tools/env.sh was absent. The first gofmt attempt before sourcing that path reported command not found; sourcing it fixed the tool lookup. setup.log preserves all timing lines.
