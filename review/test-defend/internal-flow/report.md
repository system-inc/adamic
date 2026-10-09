None of the three assigned rows was defended with these seven production mutations.
All 835 current top-level tests were included in each completed whole-package matrix.
No tests were edited; all production mutations were restored.

Base and scope

Base origin/main: b8bcadb2c493173855f19d7e5c508b34f5eeb5b6. Two corpus tests were added since the audit: node_buffer_digest_twice and pow_fractional. None vanished. The full name lists and difference are in scope-changes.json. The audit report, README, grouped rows and results were read and retained. The generated wrappers use the same four checker bodies; the timsort rows each select one of those bodies while other corpus wrappers invoke all four.

Code under test and oracle

Build and its loop/for-of lowering, Construct/ssaBuilder, and InferMutableRanges/buildAliasingGraph are code under test. These flow functions also run during preparation by the compiler. VerifySSA, textbook reachingDefinitions, fmt-based IR-read counting, the trace markers and trace walker are protected oracle/harness code and were not mutated. SSA's oracle is self. Graph paths and ranges use Node execution of instrumented Adamic JavaScript as the observed path/mutation oracle. This is external-run execution, with a compiler-generated input and self-written trace comparison; it is not an independent native execution.

Coverage and semantic attempts

Raw clean per-row profiles and full TestFlowProgram family profile use -coverpkg=./internal/flow. All four runs passed. No assigned row covers a positive production block that the family does not cover; exclusive-coverage.json retains the empty lists. family-functions.txt names the reached functions. Go coverage reflects code blocks, not the multi-million-event input history.

Timsort's semantic leads are nested counted loops and a for-of length traversal, multi-parameter comparators, and repeated mutations of arrays returned by numbers. Seven menu mutations were fixed in plan.json before mutant results. Each row has three selected attempts; some preparation mutations are shared across questions.

SSA: D1 removes counted-loop updates and catches missing IR reads. D3 omits the final parameter's SSA renaming; the assigned row passes, while five other corpus members fail. Timsort's final parameters are not reassigned, so this fault does not violate its tested reaching definitions. D4 decrements unsealed predecessors by two, losing a nested-loop phi operand. D1 and D4 both fail the assigned row and other corpus members. Neither is unique.

Graph: D1 removes update instructions, D2 routes the for-of body's back edge to its exit, and D5 removes the counted-loop exit edge. All three fail the timsort graph row, and all three also fail members of the corpus family.

Ranges: D2 challenges graph preparation for the length traversal, D6 narrows the transitive-mutation end by one, and D7 removes the entire deferred phi-alias assignment loop. The assigned range row passes all three. D2 and D6 fail other corpus members. D6 changes transitive effects, not the ordinary direct-mutation end; it did not demonstrate a fault in the ranges timsort checks. D7 survives the entire package. No changed-output witness was produced for D7, so it is an equivalent candidate, not a demonstrated unguarded defect. These attempts did not find unique range behavior. No new-session claim that the range row is untrue is made, and this is not a recommendation to delete it.

Matrix and validation

Every standalone diff applies unchanged to the base and passed go vet ./internal/flow/. Each whole-package matrix used timeout 120 go test -json -count=1 -timeout 90s ./internal/flow/ -run . and its own fresh ADAMIC_BUILD_CACHE_DIR. All seven completed within the test budget, so no bounded selection or panic rerun was needed. Full failing and passing top-level names and first assertion lines are in matrix.json. rows.json groups corpus members as TestFlowProgram family for readable verdicts. The full logs are retained as gzip files, compressed only after go test wrote to a normal file. No test output was piped.

Timing

Warm tools skipped setup; nproc 5. npm ci was rerun in stage3/api before baseline. Clean whole-package test binary: 55.254 seconds. Coverage command walls: SSA 12.725, ranges 13.985, graph 14.266, family 46.076 seconds. Whole mutant command walls: D1 66.648, D2 59.770, D3 59.375, D4 58.786, D5 72.281, D6 59.982, D7 58.456 seconds, rounded here. Exact measured values are in matrix.json. Compilation and vet were not separately phase-timed. Log parsing/compression is outside those go test wall measurements. No required run exceeded 90 seconds.

Brief friction and owner findings

The audit fetch carried approximately 118 MB of packed objects and took most of the first setup minute; it completed before the budget. The /tmp filesystem is only 8.8 GB total, so 15 GB free is impossible. The earlier deletion-parser scratch/cache was removed; tools and /workspace/adamic were untouched. /tmp had 8.3 GB free after cleanup. The source tree was clean apart from earlier untracked evidence, which was not committed. D1 and D5 produced large repeated assertion logs; archiving retained complete observations while recovering disk space. This processing added time beyond the test-binary costs.

The family has hundreds of generated wrappers. Their repeated bodies were compared through their checker calls, rather than treating wrapper spelling as a behavioral difference. Families are represented as one row in the concise result but all actual test names are retained in the matrix. Shared code coverage does not itself prove subsumption; the verdict here rests on mutation observations.

The SSA name matches the asserted single-assignment/reaching-definition and read-count properties on this lowered fixture. It does not promise that unchanged parameters receive particular fresh identifier numbers. The graph name checks observed Node paths, not exact equality of all possible CFG edges: extra permitted paths may survive. The range name should not be read as promising a range for every mutation. ranges_test.go explicitly skips unset ranges and only requires observed mutations to lie inside ranges that are set. Timsort has no assertion requiring even one range to be set; only the separate mutations.a witness requires a nonzero observed mutation count. The earlier audit's empty-table result is consistent with this source, but an empty-table probe was not rerun here.

The seven-mutant cap limited additional range faults after D6 and D7 failed to challenge this fixture. Broad widening faults from the audit were not repeated merely to reproduce known nonunique catches. Untested fault models, explicit 30-second reference-box cost enforcement, and repository-wide tests remain outside this defense. No deletion, rewrite, main push or pull request was performed.

Publication: the requested remote branch already existed. Full and tip-only fetches reached the 90-second budget and were stopped; a metadata-only blob-filtered tip fetch succeeded. The existing branch history and its complete evidence tree are retained under previous/337b16e6. These fetches added publication time beyond the approximate unit budget.
