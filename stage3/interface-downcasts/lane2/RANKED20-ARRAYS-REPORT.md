Certified original argument, modifier and type union arrays with complete original declarations; no compiler change.
Commits: this certification follows 6513b728; its commit is recorded by the group push.
Validation: 39 Node, sanitized native, release native and JavaScript probes; 24 finishing leak checks; 39 measured count rows; vet passes.
Mutants: native and JavaScript numeric-field bypasses caught by wrong-pos for arguments, modifiers and types, six executed guard mismatches; restored probes pass.
Limits: lane 4c untagged unions are skipped; tuples and production reachability are unmeasured; required global counts remain red in existing fixtures.

| Original pair | Static reads |
| --- | ---: |
| CallExpression \| NewExpression.arguments | 9 |
| HasDecorators.modifiers | 8 |
| UnionOrIntersectionTypeNode.types | 7 |

All 78 declarations come from pristine microsoft/TypeScript 050880ce59e30b356b686bd3144efe24f875ebc8. The adapter checks original declared member types, all 24 exact source spans and emitted hashes. Original reads can be narrowed by control flow: the adapter verifies their assignability to the complete original declared contract, which is held separately. This corrects an initial preparation comparison that incorrectly required the narrowed access type to equal the declaration. Complete receiver fields and selected Expression, ExportKeyword and TypeNode fields are asserted. Optional absent and explicit undefined arrays, unread invalid arrays, invalid array representations, missing and wrong numeric descendants, unread invalid second elements, wrong element kinds and map consumers are exercised. Every receiver arm has a source witness.

Exact commands, each redirected to evidence logs:

```sh
source /workspace/adamic-tools/env.sh
node stage3/interface-downcasts/lane2/original20/prepare.cjs /workspace/lane2-original-pin /workspace/lane2-original-declarations20
export ADAMIC_ARRAY20_ORIGINAL_DECLS=/workspace/lane2-original-declarations20
go test ./internal/oracle -run '^TestCheckedViewRanked20OriginalArrays$' -count=1 -v -timeout 15m
python3 stage3/interface-downcasts/lane2/original20/run-mutants.py
go test ./internal/oracle -run '^TestCheckedViewRanked20ArrayCounts$' -count=1 -v -timeout 15m -args -update-counts
go test ./internal/oracle -run '^TestCheckedViewRanked20OriginalArrays$/^(arguments-214-wrong-pos|decorators-170-wrong-pos|types-193-wrong-pos)$' -count=1 -v -timeout 10m
go vet ./internal/oracle
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 30m -args -update-counts
```

Initial full oracle passed in 42.969s; counts passed in 27.571s; restored witnesses passed in 2.993s. Final verification after adding complete ExportKeyword field assertions is retained in evidence/final.log. The required global count refresh failed in 39.909s in existing integration fixtures and changed no rows. Removing this group's 39 measured rows reproduces the previous counts table byte for byte. Six mutants execute selected descendant reads, return incorrect output instead of the expected exit-70 field rejection, and fail on stderr mismatch. None depends on a compilation failure. Temporary source edits are restored.

IncrementalBuildInfo.fileNames (4 reads) still refuses its direct untagged union cast, as held by the group 19 frontier test after integration 432d4913. This is assigned to lane 4c and earns no credit. IncrementalBuildInfo.changeFileSet is a related pending candidate, not certified. FileWatcherWithModifiedTime.callbacks depends on callable element work in lane 5. Tuples remain assigned elsewhere. The latest user instruction supersedes the earlier pause: continue with unblocked groups.

Array fixture obligations now hold 174 pairs / 2864 static reads of 334 / 3189; remaining 160 / 325. Consumer 251 / 1602 and intrinsic 179 / 794 production obligations remain uncredited; own array fields hold 3 / 26, remaining 27 / 46. These totals track fixture obligations and census spans, not production execution.

Reused toolchain setup at the merged tip: Node/Go 0.023s, markdown 0.069s, submodules 0.079s, clang 0.221s, build 42.907s, deferred tests 43.012s, cache 43.014s, done 43.041s. nproc 5, quota 4 CPUs; Go 1.27.1, clang 20.1.8, Node 24.19.0. Evidence is under original20/evidence. No full gate or whole package test was run.
