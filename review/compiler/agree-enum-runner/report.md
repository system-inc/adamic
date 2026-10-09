Moved all eight TestNumericEnumsAreOpen acceptance rows onto lowersAndAgreesWithNode and deleted enum_agree_test.go's duplicate runner.
Started from compiler/agree-modules c9022ec1 while #260 was pending; after #260 merged as 7d544715, rebased this follow-up onto current origin/main 0406ad39.
Focused tests pass with -timeout 90s; delivery TestCallTargetReaders and lane results are recorded separately.
All six enum mutants fail: M01/M02 by shared stdout comparison, M04 by unique native-slot assertion, M14/M15/M16 by named flag-proof assertions.
No source rows, observations, refusal contracts or fixtures removed; no new fixtures, no counts refresh, no full package or full gate run.

The eight rows are singleton, member_tags, coalesce, arithmetic, signed_flags, cast, containers and optional_object. Existing computed observations are preserved, including enum reverse names. No enumLowersAndAgreesWithNode callers remain.

TestReadinessErrorIncludesReceiverExpression is the exempted checked-failure row: agreeEntry(t, sourcePath, program, 0, &nodeObservation{stdout: nil, stderr: expectedPanic, code: 70}, nil) mechanically covers its complete exit-70 contract, but its documented predicate-only checkedFailure contract means that readiness row remains an exemption here, using runAgreementNode; no helper contract was broadened.

Commands run, each with output saved under this review directory:

- `timeout 100 go test ./internal/lower -run '^Test(NumericEnumsAreOpen|StringEnumsStayClosed|NumericEnumNeverProof|NumericEnumLiteralPromises|EnumReverseMappingUsesSingleSlot|EnumFlagProofsWithoutObservableLoweringEffect)$' -count=1 -timeout 90s -v`: PASS before rebase, package 1.240s; TestNumericEnumsAreOpen 1.23s.
- `timeout 700 python3 review/compiler/agree-enum-runner/replay.py`: exit 0; six mutants restored after each replay, each returned exit 1 with a test assertion. Full commands and elapsed seconds in mutants.json.
- M01 increments lowered enum values: singleton's backend output 42 then 1/A differs from source 42 then 0/A.
- M02 empties reverse names: singleton's backend output 42 then 0/ differs from source 42 then 0/A.
- M04 duplicates numeric reverse slots: TestEnumReverseMappingUsesSingleSlot fails on duplicate field slot 1.
- M14 changes literal-equality classification, M15 drops bit 30, M16 changes the AND-domain proof from either side to both: TestEnumFlagProofsWithoutObservableLoweringEffect catches each.

Toolchain reused from the preceding turn: source /workspace/adamic-tools/env.sh, GOPROXY https://proxy.golang.org|direct, nproc 5, four-CPU cgroup quota. No new test leaves; each selected leaf is below 60s.
