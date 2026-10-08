Certified original configuration include-spec arrays and compiler root-directory arrays.
Commits: this group adds fixtures, oracle checks, measured counts and reports using existing adapters.
Commands: sixteen fixtures and two mutants pass in 21.561s; scoped count refresh passes in 6.935s; oracle vet passes.
Mutants: two reached element-kind replacements are caught by pinned stopping oracles in sanitized native, release native and JavaScript.
Not covered: SourceFile.amdDependencies intersection, remaining ranked pairs and production consumer/intrinsic totals.

The pairs are ConfigFileSpecs.validatedIncludeSpecs and CompilerOptions.rootDirs, with six original static candidate reads. Preparation validates each original span and full declared member type, emits 78 complete declarations and verifies complete named field sets for both original receiver types. Unread configuration members and the original CompilerOptions index signature are retained. No reduced configuration declaration supplies the contract.

Sixteen fixtures cover joins, lazy unread arrays, wrong array and element kinds, later unread invalid elements, map consumers, source alias updates and undefined arrays. Twelve finishing source controls and two finishing mutants pass native leak checks. Both mutants replace only a reached string element descriptor with number; the typeof negative remains valid C and finishes with Node output and no stderr. The unchanged exit-70 stopping oracle catches each replacement in all three modes. rootDirs has an optional-field diagnostic distinct from the required nullable include-spec field; each observed message is pinned exactly.

```sh
source /workspace/adamic-tools/env.sh
node stage3/interface-downcasts/lane2/original28/prepare.cjs /workspace/lane2-original-pin /workspace/lane2-original-declarations28
ADAMIC_ARRAY28_ORIGINAL_DECLS=/workspace/lane2-original-declarations28 go test ./internal/oracle -run '^(TestCheckedViewRanked28OriginalArrays|TestCheckedViewRanked28ArrayMutants)$' -count=1 -v
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 30m -args -update-counts
ADAMIC_ARRAY28_ORIGINAL_DECLS=/workspace/lane2-original-declarations28 go test ./internal/oracle -run '^TestCheckedViewRanked28ArrayCounts$' -count=1 -v -args -update-counts
go vet ./internal/oracle
```

Logs are retained in original28/evidence. The required global count refresh fails in existing fixtures in 39.443s. The scoped refresh records sixteen rows; removing them reproduces the previous table byte for byte. No whole package test or full gate ran. No production or cross-lane files changed.

This adds two pairs / six original reads. Lane 2 totals are 201 pairs / 2954 reads, remaining 133 / 235 of 334 / 3189. Own fields remain 3 / 26 of 30 / 72. The lane 7 intersection list remains separate while the private directory files array proceeds.
