# V3 array union completion

Base: 8880e6bca1207cda77b00e8710d864cdeaf050d7. Roadmap step 11; completeness #f07qcpn and #tn0p5w0.

Array union membership now checks physical element kinds, scalar selectors, sparse present slots, explicit undefined and tuple identity in native and JavaScript backends. Payload fields stay checked at their reads. Array demand passes through union descriptors. Narrowed arrays use Array.isArray and recheck after calls. Indexed unions whose members use different element storage remain refused: joining logical types must not reinterpret physical storage. Dictionary, nominal and unavailable callable contracts stay refused.

Former skips: TestMixedUnionContractGraph admits its array graph; TestCheckedViewObjectPrimitiveSource runs all three comment cases (wrong flags stop at the read); TestCheckedViewUntaggedArraySource runs all five cases (good/empty match Node, wrong/mixed/nested stop loudly). An additional V3 NotYet assertion, TestCheckedViewV2ArrayArmBoundary, is replaced by TestCheckedViewV2ArrayArmAdmission. No skips naming V2 or V3 remain in Go tests. All 23 TestCheckedViewArrays cases pass, none skipped.

Oracle cases observe original source under Node, native release, native ASAN/UBSAN and JavaScript. Negative native and JavaScript results are pinned to exit 70 and the exact diagnostic. Native sanitizer runs are uncached. All runs use affinity [0,1,2,3] with GOMAXPROCS=4; host nproc is 5.

| Touched top-level test | Wall seconds |
|---|---:|
| TestArrayMembershipUsesUnionMatcher | 0.05 |
| TestMixedArrayElementStorageRemainsRefused | 0.077 |
| TestMixedUnionContractGraph | 0.164 |
| TestCheckedViewUntaggedArraySource | 1.636 |
| TestCheckedViewV2ArrayArmAdmission | 1.636 |
| TestCheckedViewArrayUnionNarrowing | 1.673 |
| TestCheckedViewArrayUnionMembership | 2.469 |
| TestCheckedViewObjectPrimitiveSource | 4.985 |
| TestCheckedViewArrays | 8.066 |
| TestCallTargetReaders | 18.594 |

All touched top-level tests and leaves are under 60 seconds. See results.json for individual leaves.

Commands: go test -json -count=1 -timeout=10m ./internal/lower -run "^(TestMixedUnionContractGraph|TestArrayMembershipUsesUnionMatcher|TestMixedArrayElementStorageRemainsRefused)$"; equivalent named oracle run for the six oracle top-level tests listed above; go test -json -count=1 ./internal/ir -run "^TestCallTargetReaders$". No whole packages run. TestCallTargetReaders verifies the unchanged allowlist after gofmt corrected inherited formatting.

Mutants: run-mutants.py (13), then run-mutants.py joined-storage (1). All 14 omissions are caught by test failures, not build failures. The mutants remove native kind, element, selector, sparse-slot, present-undefined and tuple checks; JavaScript element, selector, sparse-slot and tuple checks; array demand activation; unsupported-element refusal; narrowing recheck; and same-storage refusal. mutants.log and joined-storage-mutant.log name each catching fixture.

A-check: go build -o /tmp/views-v3-completeness-adamic ./cmd/adamic, then adamic c for each of the six changed .a files; zero failures. The unchanged original witness .ts files are not Adamic programs.

Tool setup: node 0.022s, Go 0.024s, markdown step 0.011s/ready 0.071s, clang 0.157s. Setup stopped at the existing symlink cohere worktree; installed /workspace/adamic-tools/env.sh used successfully.

Not covered: differing-storage indexed array union boxing; boolean|undefined array producers (existing frontend refusal). No V4/V5 re-skips needed. No non-view runtime files changed. Evidence is confined to review/compiler/views-v3/. Integration lane check follows the committed implementation.

Counts: go test ./internal/oracle -run "^TestCountsAreRecorded$" -args -update-counts passed (76.400s package); counts.md unchanged, zero moved rows. This existing census test was not added or touched.

Committed implementation b6093b70c: integration lane checks passed after the exact repository-root fetch/show command. Output: `lane checks 10.0 s: gofmt and tools on 56 Go files, t.Parallel on 6 test packages; no t.Parallel analyzer on this tree; vet 6 packages`. Each added or touched test function has t.Parallel first; the analyzer absence is reported rather than treated as validation by an analyzer.
