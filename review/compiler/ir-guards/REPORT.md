Built IR-level guards for mixed optional fixed slots and unknown-closure throwing effects.
Branch compiler/ir-guards, base origin/main f91994f019703ba25d2918cf529c0e0b0c05d93c; test-only changes.
Both restored leaves passed below 0.01s; focused IR checks and vet passed.
Exact M10 fails on [1] versus [7]; exact M07 fails on lost throwing effect.
No production change, full-package run, new oracle fixture or counts row.

Task #j5j3vy6. Audit: test-audit/internal-ir d563034a, review/test-audit/internal-ir/M10.diff, M07.diff and survivor-witness.go. Only the two exact diffs are saved here, as .diff. No compilable Go evidence was added.

M10 first:
- The affected layout is ClosureArgumentLayout's fixed parameter slots, not object-field storage. M10 flips of.IsMaybe() to !of.IsMaybe() in internal/ir/argument_slots.go.
- With M10 applied, TestNoReaderCallingConvention passed. The uncached TestNativeAgreesWithNode leaves for arguments_length_no_reader.a, closure_convention_plain.a, census_optional_values_padding.a and census_optional_values_receivers.a also passed, with source Node, JavaScript backend, release native, ASan/UBSan and leak checks. Package binary elapsed 10.115s. No outside catcher was observed in these selected tests; the entire native or oracle package was not run.
- TestMixedOptionalFixedLayoutPreservesMaybeNumber constructs the audit's Number/MaybeNumber callees sharing a function-type target set. It requires Fixed == []ir.Type{ir.MaybeNumber} in both target orders. Applying the exact M10 diff yields [1] for both orders, expected [7]. The assertion, not compilation, fails. After restoring M10, it passes.

M07 second:
- M07 flips function.MayThrow to !function.MayThrow in ClosureMayThrow's unknown-target scan in internal/ir/call_targets.go.
- TestUnknownClosurePreservesThrowingEffect is next to the existing CallMayThrow guards in call_targets_effects_test.go. Its unknown closure read has no constant target proof, and its sole candidate has an unconditional Throw body and MayThrow=true. It first requires unknown targets, then requires ClosureMayThrow=true. Exact M07 fails only the throwing-effect assertion. Restored production passes.

Every go test command used -count=1 -v -timeout 90s and an outer timeout 100:
1. ./internal/native -run '^TestNoReaderCallingConvention$' under M10 (M10-existing-native.log).
2. ./internal/oracle -run '^TestNativeAgreesWithNode$/internal/oracle/testdata/(arguments_length_no_reader|closure_convention_plain|census_optional_values_padding|census_optional_values_receivers)\.a$' under M10, ADAMIC_GATE_UNCACHED=1 (M10-existing-oracles.log).
3. ./internal/ir -run '^TestMixedOptionalFixedLayoutPreservesMaybeNumber$' under M10: exit 1, binary 0.007s (M10-new-test.log); restored: exit 0, binary 0.008s (M10-restored.log).
4. ./internal/ir -run '^TestUnknownClosurePreservesThrowingEffect$' under M07: exit 1, binary 0.007s (M07-new-test.log).
5. Restored ./internal/ir -run '^Test(UnknownClosurePreservesThrowingEffect|MixedOptionalFixedLayoutPreservesMaybeNumber|CallMayThrowUsesReachableTargets|DirectClosureTargetsUseEncodedIndex|ArgumentLayouts)$': exit 0, binary 0.008s (restored-focused.log). Both new leaves report 0.00s; each is below 0.01s and calls t.Parallel first.
6. timeout 90 go vet ./internal/ir: exit 0, empty output (vet.log).
7. Integration lane checks: output in lane-checks.log.

The new tests pin the IR contracts requested by the ruling: MaybeNumber preserves undefined, and unknown targets retain throwing effects. They do not claim new end-to-end Node coverage. Counts were not regenerated because no registered oracle fixture, allocation or runtime behavior changed.

Setup: GOPROXY='https://proxy.golang.org|direct'; cloud/setup.sh passed. Go ready 0.021s, Node 0.024s, markdown dependencies 0.075s, submodules 0.081s, clang 0.177s, go build 36.133s, total 36.404s. nproc 5, CPU quota 4. Full output: setup.log.
