Certified the original signature parameter array union across all 15 receiver arms and both complete element contracts; no compiler change.
Commits: follows group 21 5bb11e73 on codex/views-arrays-callables-parser; group commit is recorded in the push output.
Validation: 31 Node, sanitized native, release native and JavaScript probes pass in 36.437s; 21 finishing leak checks; 31 measured counts plus restored witnesses pass in 26.928s; vet passes.
Mutants: native and JavaScript numeric-field bypasses caught for CallSignatureDeclaration and JSDocSignature descendants, four executed guard mismatches.
Limits: callable-array map and replacement frontiers remain pending; tuples and production execution are unmeasured; required global counts still fail in existing fixtures.

SignatureDeclaration | JSDocSignature.parameters accounts for one previously blocked pair and five exact original source reads. The original field is NodeArray<ParameterDeclaration> | readonly JSDocParameterTag[]. All 78 complete declarations come from pristine microsoft/TypeScript 050880ce59e30b356b686bd3144efe24f875ebc8. The adapter validates declared member types, exact source spans and emitted hashes. Fixtures assert full original receiver fields and selected ParameterDeclaration or JSDocParameterTag fields. Both element families hold valid reads, lazy unread arrays/elements, invalid arrays and scalar elements, missing/wrong numeric descendants and map consumers. Every original receiver arm is exercised. Initial expected diagnostics named one arm; corrected expectations preserve the full original union name for array representation and element selection rejections.

Commands, each redirected to retained evidence logs:

```sh
source /workspace/adamic-tools/env.sh
node stage3/interface-downcasts/lane2/original22/generate.cjs /workspace/lane2-original-pin
node stage3/interface-downcasts/lane2/original22/prepare.cjs /workspace/lane2-original-pin /workspace/lane2-original-declarations22
export ADAMIC_ARRAY22_ORIGINAL_DECLS=/workspace/lane2-original-declarations22
go test ./internal/oracle -run '^TestCheckedViewRanked22OriginalArrays$' -count=1 -v -timeout 15m
python3 stage3/interface-downcasts/lane2/original22/run-mutants.py
go test ./internal/oracle -run '^TestCheckedViewRanked22ArrayCounts$|^TestCheckedViewRanked22OriginalArrays$/^signature-(180|324)-wrong-pos$' -count=1 -v -timeout 15m -args -update-counts
go vet ./internal/oracle
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 30m -args -update-counts
```

All four mutants execute descendant reads and fail on stderr mismatch against the required exit-70 rejection. No compilation failure counts. Sources are restored before measuring all 31 rows. Removing these rows leaves the previous table byte for byte. The global refresh fails in 50.341s in existing integration fixtures, with no table mutation. Final group counts take 24.45s and restored witnesses 2.46s. No whole package or full gate was run.

Fixture obligations now hold 179 pairs / 2892 reads of 334 / 3189; remaining 155 / 297. Consumer and intrinsic production totals remain uncredited; own array fields remain 3 / 26. IncrementalBuildInfo.fileNames and InterfaceType.resolvedBaseTypes remain untagged-union frontiers at this certification. The latest instruction assigns lane 2 ownership of the array joints and untagged admission needed by fileNames, so these dependencies are next work, not a reason to pause.

A complete private-declaration probe now exposes FileWatcherWithModifiedTime by appending an export in the original declaration-emission host, preserving all three fields and the imported FileWatcherCallback signature. Direct callback calls pass Node and all backends plus leaks. Map refuses at the reached element with unsupported callable contract. Reference replacement reaches an uncertified source element contract, with different native and JS wording. These observations give no pair credit and will be handled after the newly prioritized array joints. Preparation and initial probes remain working files for group 23. No reduced declaration or cohere code is copied.

Toolchain is reused from setup at the merged tip (43.041s; nproc 5, quota 4): Go 1.27.1, clang 20.1.8, Node 24.19.0. Integration 432d4913 is already merged. Evidence is under original22/evidence.
