Held FunctionLikeDeclaration.parameters and ClassDeclaration | ClassExpression.members against complete original declarations; no compiler change was needed.
Commits: based on ffe428ab26eefb73154adbf570ad99e9c1b4f872; this report and its fixtures are committed together on codex/views-arrays-callables-parser.
Validation: 81 probes pass Node, sanitized/release native and JavaScript; 36 finishing probes pass leaks; 81 measured count rows; final oracle and counts pass in 140.243s; vet passes.
Mutants: native and JavaScript numeric-field bypasses caught for both union targets by wrong-pos witnesses, four executed failures; restored witnesses pass in 2.432s.
Limits: IncrementalBuildInfo.fileNames remains blocked at union admission; production reachability, tuples and other pending candidates are unmeasured; global counts still fail in existing fixtures.

| Original pair | Type id | Static candidate reads | Declaration |
| --- | ---: | ---: | --- |
| FunctionLikeDeclaration.parameters | 9226 | 22 | NodeArray<ParameterDeclaration> |
| ClassDeclaration \| ClassExpression.members | 7679 | 21 | NodeArray<ClassElement> |

The pinned original microsoft/TypeScript revision is 050880ce59e30b356b686bd3144efe24f875ebc8. The adapter emits all 78 complete declaration files, checks every exact source span and original member type, and pins emitted hashes. It records complete fields for all nine receiver arms and both element interfaces. Each probe checks the complete field set of its receiver arm and every selected element contract. No cohere code or reduced declaration schema is substituted.

Nine probes per arm cover valid selected elements, unread wrong arrays, reached wrong arrays, wrong scalar pos, missing pos readiness, unread malformed second elements, wrong element kinds, map consumers and wrong map descendants. All seven FunctionLikeDeclaration arms and both class arms use direct original union casts. Parameter tags use the original Parameter value 170; ClassElement payloads use PropertyDeclaration 172. Unread fields remain lazy obligations. These probes certify fixture obligations and static candidate spans, not production execution.

Both numeric-field mutants execute through each original target. Native returns the slot without validating its stored kind and prints pointer bits as numeric output with exit zero. JavaScript makes the numeric predicate unconditional and prints bad with exit zero. Each differs from the required named exit-70 rejection. No kill relies on compilation or warnings. Temporary edits are restored in finally. The final restored witnesses and the next-candidate frontier pass, and source files outside the oracle have no retained changes.

Reproduction, with each command's output sent to a log:

```sh
source /workspace/adamic-tools/env.sh
node stage3/interface-downcasts/lane2/original19/prepare.cjs /workspace/lane2-original-pin /workspace/lane2-original-declarations19
export ADAMIC_ARRAY19_ORIGINAL_DECLS=/workspace/lane2-original-declarations19
go test ./internal/oracle -run '^TestCheckedViewRanked19' -count=1 -v -timeout 10m -args -update-counts
python3 stage3/interface-downcasts/lane2/original19/run-mutants.py
go test ./internal/oracle -run '^TestCheckedViewRanked19OriginalArrays$/^(function-263-wrong-pos|class-264-wrong-pos)$|^TestCheckedViewRanked19NextFrontier$' -count=1 -v -timeout 10m
go vet ./internal/oracle
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 30m -args -update-counts
```

The complete initial oracle passed in 106.610s. Final corrected fixture sources and measured counts passed together in 140.243s (the counts portion took 53.25s). Final restored witnesses passed in 2.432s. Vet and git diff --check pass. The global counts attempt failed in 44.715s before the lane registry, including graph_regions_regression_06.a, process_exit.a and existing node_fs fixtures. These are observed failures at the unchanged compiler tip; the previous reports establish baseline status for the first two. This group writes only its 81 measured rows; removing those rows yields the previous counts table byte for byte. Missing external inputs skip declaration-dependent tests and retain recorded counts without claiming remeasurement. No whole-package test or full gate was run.

Next four-read candidate IncrementalBuildInfo.fileNames stops at the original untagged union cast before selecting its array. Node prints a;b; lowering reports adamic/no-unchecked-cast. TestCheckedViewRanked19NextFrontier pins this observed blocker and grants no certification credit. Its full declarations include tuple descendants, which remain unread and outside this worker's scope. origin/codex/views-integration still equals ffe428ab, so there is no newer implementation to merge. The remaining FileWatcherWithModifiedTime.callbacks candidate has callable elements; call signatures belong to lane 5. Tuple candidates remain skipped.

| Family | Candidate pairs / reads | Cumulative fixture obligations held | Remainder |
| --- | ---: | ---: | ---: |
| Array contracts | 334 / 3189 | 171 / 2840 | 163 / 349 |
| Element or consumer reads | 251 / 1602 | 0 / 0 production credit | 251 / 1602 |
| Intrinsic contracts | 179 / 794 | 0 / 0 production credit | 179 / 794 |
| Own array fields | 30 / 72 | 3 / 26 | 27 / 46 |

Toolchain: nproc 5, cgroup quota 4 CPUs, Go 1.27.1, clang 20.1.8, Node 24.19.0. Setup initially failed with undefined internal/load symbols while branch checkout completed; retry at the stable tip succeeded. Retry timing lines: Node/Go 0.018s, submodules 0.044s, markdown 0.070s, clang 0.149s, build 41.090s, tests deferred 41.188s, cache 41.189s, done 41.216s. The environment file is /workspace/adamic-tools/env.sh. Fetch required an explicit branch ref because this clone initially fetched only main. No merge or code conflict occurred.

Evidence is retained under original19/evidence. The prepared group nineteen mentioned in the task was not present at the fetched tip; the newest existing complete-original report was group eighteen. This group nineteen supplies the requested union array certification first.
