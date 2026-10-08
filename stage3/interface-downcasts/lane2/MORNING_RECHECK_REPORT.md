Merged integration 432d4913 and rechecked the original union arrays; no array compiler change was needed.
Commits: integration 432d4913d31daaa49d8ca4eb46f30f21f90da5ee merged as cd68da47a912e01f93d18976a1fe8a1cc0028157; this report and evidence follow that merge on codex/views-arrays-callables-parser.
Validation: all 81 behavior probes and 36 finishing leak checks pass, all 81 recorded count rows match, the blocked-candidate pin passes; full focused command PASS 138.349s; vet and diff checks pass.
Mutants: native-number-field-function, native-number-field-class, javascript-number-field-function and javascript-number-field-class each fail the appropriate wrong-pos witness during execution; restored witnesses PASS 2.020s.
Limits: IncrementalBuildInfo.fileNames remains blocked on direct untagged-union admission, assigned to lane 4c in the plan; no new candidate credit, tuple work or production reachability claim.

The branch contained e659d860 before this merge. Explicit fetch found the requested integration SHA 432d4913d31daaa49d8ca4eb46f30f21f90da5ee. The merge had no code-file conflict. Both branches appended to docs/checked-views-plan.md and internal/oracle/counts.md. All plan sections and both disjoint sets of count rows were retained unchanged; the counts conflict had no shared fixture key. The focused measured-count test needs no refresh and makes no table change.

Original declarations remain the complete pinned 78-file graph from microsoft/TypeScript 050880ce59e30b356b686bd3144efe24f875ebc8. All seven FunctionLikeDeclaration arms and both class arms retain their complete receiver and selected element fields. The original union array pairs remain certified for fixture obligations: FunctionLikeDeclaration.parameters accounts for 22 static candidate reads; ClassDeclaration | ClassExpression.members accounts for 21.

The next four-read candidate is still blocked. The original IncrementalBuildInfo is the untagged IncrementalMultiFileEmitBuildInfo | IncrementalBundleEmitBuildInfo union. Node prints a;b. Adamic refuses `(base as IncrementalBuildInfo)` with adamic/no-unchecked-cast before it can read fileNames. This is the existing TestCheckedViewRanked19NextFrontier probe, importing the complete original builder declarations. Replacing the union with IncrementalBuildInfoBase would reduce the target and is not credited. The plan's Lane 4 pure primitive property selectors section explicitly says direct untagged union admission remains lane 4c. The newest integration still equals 432d4913 at the final remote check. Per the user's instruction to stop when blocked on another lane, no later ranked group is started.

Commands, all writing complete output to logs retained under original19/morning-evidence:

```sh
git -c fetch.recurseSubmodules=false fetch origin refs/heads/codex/views-integration:refs/remotes/origin/codex/views-integration
git merge --no-edit origin/codex/views-integration
# Retain both independent Markdown additions, then finish the merge.
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/lane2-morning-setup.log 2>&1
source /workspace/adamic-tools/env.sh
export ADAMIC_ARRAY19_ORIGINAL_DECLS=/workspace/lane2-original-declarations19
go test ./internal/oracle -run '^TestCheckedViewRanked19NextFrontier$|^TestCheckedViewRanked19OriginalArrays$/^(function-263-good|class-264-good|function-263-wrong-pos|class-264-wrong-pos)$' -count=1 -v -timeout 10m > /tmp/lane2-morning-recheck.log 2>&1
go test ./internal/oracle -run '^TestCheckedViewRanked19' -count=1 -v -timeout 10m > /tmp/lane2-morning-group19.log 2>&1
python3 stage3/interface-downcasts/lane2/original19/run-mutants.py > /tmp/lane2-morning-mutants.log 2>&1
go test ./internal/oracle -run '^TestCheckedViewRanked19OriginalArrays$/^(function-263-wrong-pos|class-264-wrong-pos)$' -count=1 -v -timeout 10m > /tmp/lane2-morning-restored.log 2>&1
go vet ./internal/oracle > /tmp/lane2-morning-vet.log 2>&1
git diff --check
```

Initial recheck PASS 19.298s. Complete behavior portion PASS 84.66s, measured-count comparison PASS 53.47s, refusal pin PASS 0.20s; combined command PASS 138.349s. Both native modes and JavaScript match the expected behavior, and all 36 finishing probes remain leak-clean. Native numeric-field mutants print unchecked pointer bits with exit zero; JavaScript numeric-field mutants print bad with exit zero. Each disagrees with the exact required exit-70 refusal. No kill relies on a compile failure. The runner restores each file in finally, final witness tests pass, and the workspace was clean before writing this report. Vet and diff checks produce no errors.

There are no new fixtures, so no global counts update is requested by this checkpoint. The full focused count comparison verifies the existing 81 rows unchanged. The prior global counts failures are not claimed fixed or remeasured. No whole-package test or full gate was run, and no tuple test was added or run.

Setup PASS: Node/Go 0.023s, markdown 0.069s, submodules 0.079s, clang 0.221s, build 42.907s, tests deferred 43.012s, cache 43.014s, done 43.041s. nproc=5, cgroup quota 4 CPUs. Go 1.27.1, clang 20.1.8, Node 24.19.0. Source environment: /workspace/adamic-tools/env.sh.

| Family | Candidate pairs / reads | Fixture obligations held | Remainder |
| --- | ---: | ---: | ---: |
| Array contracts | 334 / 3189 | 171 / 2840 | 163 / 349 |
| Element or consumer reads | 251 / 1602 | 0 / 0 production credit | 251 / 1602 |
| Intrinsic contracts | 179 / 794 | 0 / 0 production credit | 179 / 794 |
| Own array fields | 30 / 72 | 3 / 26 | 27 / 46 |

The ledger is unchanged. FileWatcherWithModifiedTime.callbacks still has callable elements belonging to lane 5. Tuple candidates remain with their worker. This checkpoint does not certify those candidates or any later ranked pairs.
