Built original mixed-array storage and consumer selection with preserved aliases and reached element checks.
Commits: lane 2 44d64b64; lane 4 adapters b2aa9c9c; lane 7 controls 0acf7c6c; lane 4b controls 998de79f; evidence follows.
Checks: 21 owned probes, six original consumer controls and two original frontier controls pass Node, sanitized native, release native and JavaScript; focused final run 41.614s; regressions 11.228s; 21 counts recorded and verified.
Mutants: nine descriptor omissions run in all three modes, fail the pinned stopping oracle, finish with Node output and pass leak checks.
Limits: tuple storage admission, IncrementalBuildInfo.fileNames and callback-array map/writes remain next; no production tsc reachability or null/object mixed storage credit.

Complete declarations remain the 78 original files at Microsoft TypeScript 050880ce59e30b356b686bd3144efe24f875ebc8. The preparation adapter verifies exact original spans, hashes, declaration field sets and complete consumer signatures. Union display order is ignored when comparing otherwise identical declaration strings; signature truncation is disabled. Three original array fields are exercised: 97922 IncrementalMultiFileEmitBuildInfo.fileInfos, 97934 IncrementalBundleEmitBuildInfo.fileInfos and 97937 IncrementalBuildInfo.fileInfos. Consumers 97923 and 98493 cover the requested lane 7 joints. These five original pairs account for five static reads across overlapping inventories. They do not measure execution of tsc.

The existing heap-reference array representation stores scalar boxes and object references. The capability admits only represented scalar/undefined alternatives and flat structural objects with primitive fields. Tuple, nominal and unsupported descendants remain outside it. A dedicated physical extraction mode retains the existing array and selected object identity. Ordinary arrays use the same per-element metadata as views. The new shared union adapter classifies the actual heap and unpacks optional boolean fields using the existing representation bytes. Complete mixed-array selection checks every field rather than treating a finite boolean as a sufficient selector. Readonly unions of these arrays check storage at the field read and defer joined element contracts to reached consumers. Their existing descriptors and member lists remain complete.

Each bundle, multi and union family has a string control, object control, alias mutation control, early-stopping some control, wrong scalar, wrong version and missing version. Alias controls read second after modifying the original string array. Early-stopping consumers leave a later number unread. Reached numbers, wrong versions and missing required versions exit 70 before callback output. Exact pins use:

```
adamic: panic: field read failed: node.fileInfos[element] matches no member of <original element type>; expected <original element type>, found number
adamic: panic: field read failed: node.fileInfos[element] matches no member of <original element type>; expected <original element type>, found object
```

The element names are IncrementalBundleEmitBuildInfoFileInfo, IncrementalMultiFileEmitBuildInfoFileInfo and string | FileInfo | IncrementalMultiFileEmitBuildInfoBuilderStateFileInfo. The test file pins every complete message. Number mutants add an otherwise excluded numeric member; version mutants omit only the version field from object descriptors. Compilation failure does not count. All nine mutants return the source Node output with empty stderr, so their disagreement with the exit-70 oracle demonstrates the stopping check. Twenty successful source controls and nine finishing mutants pass native leak checks. Existing untagged helper, generic, callback and stored flows also pass.

Commands, all output redirected to retained logs:

```sh
source /workspace/adamic-tools/env.sh
node stage3/interface-downcasts/lane4b/original/prepare.cjs /workspace/lane2-original-pin /workspace/lane2-joint-original-declarations
node stage3/interface-downcasts/lane2/joints/mixed/prepare.cjs /workspace/lane2-original-pin /workspace/lane2-joint-original-declarations
export ADAMIC_INTERSECTION_ORIGINAL_DECLS=/workspace/lane2-joint-original-declarations
export ADAMIC_OBJECT_PRIMITIVE_ORIGINAL_DECLS=/workspace/lane2-joint-original-declarations
go test ./internal/oracle -run '^TestCheckedViewOriginalMixedArrayJoints$|^TestCheckedViewIntersectionOriginalMixedArrays$|^TestCheckedViewObjectPrimitiveRemainingFrontiers$/^incremental-(bundle|union)-each-frontier$|^TestCheckedViewUntaggedSourceFlows$' -count=1 -v -timeout 15m
go test ./internal/oracle -run '^(TestCheckedViewArrays|TestCheckedViewNativeArrays|TestCheckedViewPrimitiveArraySafety|TestCheckedViewUntaggedArrayUnion|TestCheckedViewMixedSelection|TestCheckedViewObjectUnions)$' -count=1 -v -timeout 10m
go test ./internal/ir -run '^TestMixedArrayContractBounds$' -count=1
go vet ./internal/ir ./internal/lower ./internal/native ./internal/javascript ./internal/oracle
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 30m -args -update-counts
go test ./internal/oracle -run '^TestCheckedViewMixedArrayJointCounts$' -count=1 -v -args -update-counts
go test ./internal/oracle -run '^TestCheckedViewOriginalMixedArrayJoints$/^multi-number$|^TestCheckedViewMixedArrayJointCounts$' -count=1
go test ./internal/oracle -run '^TestCheckedViewOriginalMixedArrayJoints$/^(multi|union)-number$' -count=1
```

Scoped counts pass in 17.075s; recorded-row verification plus the original multi-number witness passes in 18.130s. Removing the 21 new rows leaves the previous counts table byte for byte. The required global refresh fails in 38.199s in existing node_fs, graph_regions and fresh_refused fixtures and leaves the table unchanged. No whole package or full gate ran. Initial probes exposed optional boolean handling, finite-selector shortcuts and eager whole-array selection; their failed logs are retained alongside final evidence.

Cross-lane changes are separate: b2aa9c9c changes only the six lane 4 union adapter/runtime files, 0acf7c6c changes only the lane 7 original controls, and 998de79f changes only the lane 4b frontier test. The latest integration tip fetched was 432d4913, already merged. No code merge conflict occurred. Tool setup is reused: 43.041s, nproc 5, CPU quota 4, Go 1.27.1, clang 20.1.8, Node 24.19.0.

Lane 2 gains three array field pairs / three reads: cumulative fixture obligations 182 pairs / 2895 reads, remaining 152 / 294 of 334 / 3189. The two consumer certificates overlap the other lanes' inventories and receive no duplicate array-pair credit. Production consumer and intrinsic totals remain uncredited. Own array fields remain 3 / 26 of 30 / 72. The requested tuple cast frontier and original fileNames cast are next work rather than other-lane waiting points. Group 23 callback preparation is preserved in a named stash until these joints finish.
