Certified original mapper sources, optional mapper targets and flow antecedent arrays using existing checked-read adapters.
Commits: this group contains only fixtures, oracle checks, counts, reports and the lane 7 handoff list; no production adapter changes.
Commands: controls, ten mutants and count verification pass in 65.553s; scoped count refresh passes in 22.584s; oracle vet passes.
Mutants: three array-kind reads, three element-kind reads, two reached flag kinds and two missing-flag readiness reads are caught by pinned stopping oracles in all three modes.
Not covered: SourceFile.amdDependencies intersection, production consumer/intrinsic totals, remaining ranked pairs and whole FlowNode operations.

Preparation emits 78 complete original declarations at pin 050880ce59e30b356b686bd3144efe24f875ebc8 and checks all nine original spans and declared member types. Extract<TypeMapper, {kind: 1}> retains the entire original array variant, including kind, sources and targets. Every reached Type descriptor retains the original field set. FlowReduceLabelData retains target and antecedents; all nine original FlowNode arms retain complete fields.

Twenty-eight fixtures cover successful reads, bad array kinds, bad element kinds, lazy unread arrays and later elements, reached flag checks, missing fields, map consumers, alias field updates and undefined targets. A valid FlowNode element supplies original required id, node and antecedent fields before structural union selection. Initial incomplete FlowNode controls were rejected; they were corrected rather than reducing original contracts. Sixteen finishing source controls and ten finishing mutants pass native leak checks. Flow bad-flag and missing-flag inputs stop at the complete FlowNode selection, with their messages pinned separately.

The ten mutants replace only the reached contract. Wrong-array and wrong-element probes observe typeof so the widened mutants remain valid C and execute. Optional targets replace the complete nullish read metadata along with its scalar contract. Wrong-flag and missing-flag mutants similarly observe typeof and change only the reached flag guard. Earlier mutants caught only by clang are not counted. Each final mutant exits normally with Node stdout and no stderr; the unchanged exit-70 stopping oracle catches every replacement in sanitized native, release native and JavaScript.

```sh
source /workspace/adamic-tools/env.sh
node stage3/interface-downcasts/lane2/original24/prepare.cjs /workspace/lane2-original-pin /workspace/lane2-original-declarations24
ADAMIC_ARRAY24_ORIGINAL_DECLS=/workspace/lane2-original-declarations24 go test ./internal/oracle -run '^(TestCheckedViewRanked24OriginalArrays|TestCheckedViewRanked24ArrayMutants|TestCheckedViewRanked24ArrayCounts)$' -count=1 -v
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 30m -args -update-counts
ADAMIC_ARRAY24_ORIGINAL_DECLS=/workspace/lane2-original-declarations24 go test ./internal/oracle -run '^TestCheckedViewRanked24ArrayCounts$' -count=1 -v -args -update-counts
go vet ./internal/oracle
```

Logs are retained in original24/evidence. The required global count refresh fails in existing fixtures in 55.236s. The scoped refresh records 28 rows; removing those rows reproduces the previous table byte for byte. No whole package test or full gate ran. SourceFile.amdDependencies remains listed for lane 7 in INTERSECTION-HANDOFFS.md and receives no credit.

This adds three pairs / nine original reads. Lane 2 totals are 187 pairs / 2912 reads, remaining 147 / 277 of 334 / 3189. Own fields remain 3 / 26 of 30 / 72. The next group covers accessor, call-signature and construct-signature parameter arrays.
