# u115 comments helper test audit

Starting commit: ce1c5a2fd91e40b05160e587ea5d88ed5bdfedb2. Warm toolchain: Go 1.27.1; nproc 5; setup skipped. npm ci reported 371 ms.

Scope and limits
The starting commit is ce1c5a2fd91e40b05160e587ea5d88ed5bdfedb2. Discovery listed 11 top-level tests. Five numbered runCommentMutant wrappers form one witness family. Union independently checks the shard ID sets, so it is separate under the brief's distinct-assertion exception. No row was skipped. No previous-unit evidence was reused.

The whole clean package timed out at 90.018 binary seconds after its three production checks passed. Separate runs established green baselines for the remaining witness family, union, planted-failure parent and adapter-guard witness. The production matrix is bounded to the three direct production rows. The witness and setup rows receive their own controlled edits. Package-wide and repository-wide production uniqueness outside that bounded set remain unknown.

Brief ambiguities and costs
1. The family rule both includes numbered shards plus their union and permits a member with different assertions to be separate. Here the union proves ID coverage, while the numbered members prove agreement can detect compiled port defects. Keeping the union separate makes its setup-check verdict and cost visible; scope.json records this decision.
2. Every-row source mutants conflict with the witness exception: the five compiled-mutant checks, planted-failure parent, and adapter-guard mutant are witnesses, not ordinary agreement rows. Their production-mutant failures would only show a broken precondition. I used weakened checks and excluded those failures from production kills.
3. The prescribed whole-package matrix is infeasible after the clean package reaches 90 seconds. I used the brief's bounded exception, listing the three direct production callers in every production row. The timeout is not treated as a red baseline or a kill.
4. The three checks share a module driver entry, although they cover different helper behavior. P1 empties that module body once, rather than probing preparation helpers or substituting empty Go oracle answers. The construction row has its own P2 empty menu probe. Witness vacuity stays null.
5. TypeScript module entry cannot return at top level. P1 deletes the complete driver body, leaving imports and its read helper, which produces the empty program answer. It is a probe and contributes no production kill or uniqueness.
6. The existing native build helper creates a new temporary binary for each test. A source selector would not make those products reusable without changing the harness. I used the explicitly allowed four-source-mutant rebuild route. Internal build and execution costs are combined in the test times; standalone sanitized CLI rebuild timings are separately recorded in validation.json.
7. Agreement tests compare Node before building/running native. A Node mismatch ends the row early. The source mutant catches therefore prove the Node-side comparison can fail, not that the later native assertion executed. Each standalone mutant is also compiled with the native sanitized build command.
8. The consumer agreement uses Go's actual AST traversal spans as input. That is the test's declared adapter contract. Its expected comment output is still produced by live Go cohere, not copied into a snapshot. It does not prove Adamic's independent parser handles those consumer inputs.
9. The JSX expectation is partly self-written. Live Go validates the input's diagnostics, but exit 70 and the NotYet label are an Adamic refusal contract. The predicate checks both, and the test also compares Node/native stderr, so a matching exit code alone is insufficient.
10. Four source mutations are a sample, not exhaustive semantic coverage. Any subsumption rests only on the shared observed kills, not on all behavior. No deletion recommendation is justified by that sample.
11. Go JSON wall Elapsed includes rounding beyond the printed ok line. The medians use the printed test binary line as requested. The run command's outer wall time, which includes Go startup/compilation, is kept separately in runs.json.

Measured rows

| Row | Median binary seconds | Verdict | Production kills |
|---|---:|---|---|
| TestCommentMutants family | 27.289 | witness |  |
| TestCommentMutantsUnion | 0.007 | setup-check |  |
| TestCommentMutantsPlantedFailure | 0.041 | witness |  |
| TestCommentsMatchCohere | 19.098 | sacred | M1, M2, M3 |
| TestConsumerCommentHelpers | 36.962 | subsumed | M3 |
| TestJsxParserGapIsExplicit | 21.843 | sacred | M4 |
| TestJsxAdapterGuardMutant | 19.111 | witness |  |

Source mutations

| ID | Origin file:line | Change | Failing rows in bounded matrix |
|---|---|---|---|
| M1 | stage1/cohere/lint/helpers/comments/can_begin_at.ts:14 | `code > 13` to `code > 12` | TestCommentsMatchCohere |
| M2 | stage1/cohere/lint/helpers/comments/all.ts:78 | `point > 65535 ? 2 : 1` to `point > 65535 ? 1 : 1` | TestCommentsMatchCohere |
| M3 | stage1/cohere/lint/helpers/comments/sort_by_position.ts:3 | `let outer = 1;` to `let outer = 2;` | TestCommentsMatchCohere, TestConsumerCommentHelpers |
| M4 | stage1/cohere/lint/helpers/comments/main.ts:28 | `panic('NotYet: stage-1 JSX parser adapter');` to `(drop statement)` | TestJsxParserGapIsExplicit |

Every production mutant was caught. No survivors or equivalent candidates. The consumer subsumption is based on one shared mutant, M3, and is only a hint.

Compilation checks

- M1: exit 0, 18.116 wall seconds; `timeout 90 go run ./cmd/adamic build /workspace/adamic/stage1/cohere/lint/helpers/comments/main.ts -o /tmp/u115/M1-native --sanitize`.
- M2: exit 0, 17.860 wall seconds; `timeout 90 go run ./cmd/adamic build /workspace/adamic/stage1/cohere/lint/helpers/comments/main.ts -o /tmp/u115/M2-native --sanitize`.
- M3: exit 0, 17.848 wall seconds; `timeout 90 go run ./cmd/adamic build /workspace/adamic/stage1/cohere/lint/helpers/comments/main.ts -o /tmp/u115/M3-native --sanitize`.
- M4: exit 0, 18.475 wall seconds; `timeout 90 go run ./cmd/adamic build /workspace/adamic/stage1/cohere/lint/helpers/comments/main.ts -o /tmp/u115/M4-native --sanitize`.
- P1: exit 0, 7.256 wall seconds; `timeout 90 go run ./cmd/adamic build /workspace/adamic/stage1/cohere/lint/helpers/comments/main.ts -o /tmp/u115/P1-native --sanitize`.
- W1: exit 0, 0.167 wall seconds; `go vet ./stage1/cohere/lint/helpers/comments/`.
- W2: exit 0, 0.172 wall seconds; `go vet ./stage1/cohere/lint/helpers/comments/`.
- W3: exit 0, 0.163 wall seconds; `go vet ./stage1/cohere/lint/helpers/comments/`.
- S1: exit 0, 0.151 wall seconds; `go vet ./stage1/cohere/lint/helpers/comments/`.
- P2: exit 0, 5.545 wall seconds; `go vet ./stage1/cohere/lint/helpers/comments/`.

The standalone rebuild commands include Go CLI startup and compilation. The tests do not expose separate native build timings; their binary times include oracle construction, native builds, and execution. Total timed audit commands: 718.857 wall seconds; this excludes initial package baseline, bounded clean discovery runs, npm install, and standalone compilation validations.

Evidence files: rows.json, matrix.json, mutants.json, scope.json, functions.json, plan.json, runs.json, validation.json, all command logs, standalone M1..M4 diffs, P1/P2 probes, W1/W2/W3 weakening diffs and S1 construction diff. The Python files record this session workflow; central replay should use the standalone diffs.

Not covered: other packages; production effects outside the three-row matrix; exhaustive mutations; internal native build versus runtime separation; later native assertions after a Node mismatch. No test skipped. Source restored before commit.
