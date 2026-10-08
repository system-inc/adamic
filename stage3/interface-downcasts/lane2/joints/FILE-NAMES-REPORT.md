Built lazy common-array union admission for all four original IncrementalBuildInfo.fileNames reads.
Commits: 1fbb892e own admission and certificate; cddc09c1 shared lowering hooks; certification follows separately.
Commands: Node, sanitized native, release native, JavaScript, leaks, mutants, counts and focused regressions pass in 22.345s; vet passes.
Mutants: reached-element membership, array kind, required presence and whole-member omission all fail pinned stopping oracles in three modes and finish with Node output without leaks.
Not covered: production consumer/intrinsic totals, unrelated unsupported union descendants and the remaining ranked pairs; callbacks are next.

Admission checks reverse assignability and writable-slot restrictions for every original object arm. It installs the complete existing view descriptors and preserves aliases. A required common array read may use its complete primitive element contract only when all arms contain the same field contract. The helper does not select a union arm, discard unread fields, or erase an unsupported whole-value obligation. Existing whole-member union checks still reject a missing payload at a whole union read.

file-names/prepare.cjs emits the complete 78 pinned declarations and verifies all four original source spans and their readonly string[] type. The semantic oracle compares both complete original object-arm field sets against the manifest. Eight fixtures cover joins, four repeated reads, alias replacement, early-stopping some, unread malformed storage, a reached numeric element, wrong array kind and missing array storage. Positive controls remain lazy and match Node. The handed-off original19 frontier now also matches Node in all modes and passes leak checks.

Three fixture negatives and an inline whole-union boundary pin exact exit-70 messages in checked_views_file_names_joint_test.go. Their mutants widen the reached element descriptor, change only the demanded physical kind, permit absent storage, or omit required whole-member fields. Compilation failure does not count: all four finish with the source Node output, empty stderr and no native leaks. Initial mutation setup used a toString selector and changed the separator; the corrected mutant retains the original join operation. The complete declaration preparation and final evidence are retained under file-names/evidence.

Commands, output redirected to logs:

```sh
source /workspace/adamic-tools/env.sh
node stage3/interface-downcasts/lane2/joints/file-names/prepare.cjs /workspace/lane2-original-pin /workspace/lane2-joint-original-declarations
export ADAMIC_TUPLE_ORIGINAL_DECLS=/workspace/lane2-joint-original-declarations
export ADAMIC_ARRAY19_ORIGINAL_DECLS=/workspace/lane2-original-declarations19
go test ./internal/oracle -run '^(TestCheckedViewFileNamesJoint|TestCheckedViewFileNamesWholeUnionBoundary|TestCheckedViewFileNamesJointCounts|TestCheckedViewRanked19NextFrontier|TestCheckedViewUntaggedSourceFlows|TestCheckedViewNativeArrays)$' -count=1 -v
go vet ./internal/ir ./internal/lower ./internal/native ./internal/javascript ./internal/oracle
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 30m -args -update-counts
go test ./internal/oracle -run '^TestCheckedViewFileNamesJointCounts$' -count=1 -v -args -update-counts
```

The required global count refresh fails in 52.560s in pre-existing fixtures. The scoped refresh passes in 6.290s and adds eight measured rows. Removing those rows reproduces the previous table byte for byte. No whole package or full gate ran. cddc09c1 changes only the shared cast preflight and lazy read dispatcher; other family files are unchanged in this group.

This adds one lane 2 array-field pair / four original reads. The ledger is now 183 pairs / 2899 reads certified, remaining 151 / 290 of 334 / 3189. Own array fields remain 3 / 26 of 30 / 72. The requested mixed-array and array-to-tuple joints are already pushed; this completes the third requested joint. Ranked callback preparation remains preserved and is next.
