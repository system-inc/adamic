#!/usr/bin/env bash
set -euo pipefail
source /workspace/adamic-tools/env.sh
timeout 360 go test $(cat review/compiler/loom-implied-target/reader-packages.txt) -run '^TestCycleScalar|^TestImpliedTarget|^TestAMutableLocationSeenWiderIsRefused$|^TestAViewThatCantWriteIsNotRefused$|^TestOptionalWidening|^TestNodeHostOptionsRefusalBoundaries$|^TestNodeFSFileDoesNotAuthorizeMutableWidening$|^TestEnumSlotViews$|^TestEmptyNeverMapCannotGainWritableInhabitants$|^TestProvenRelations|^TestStage3EnumBoundaries$|^TestClosureMergeRefusals$|^TestCheckedViewUntaggedSourceDispatch$|^Test.*Cycle|^TestLiteralMethodCapturesCannotMakeCycles$|^TestReadonlyFieldsAreJudgedByTheirConstructorsWrites$|^TestFreshWriteProbesStayRefused$' -count=1 -v -timeout 90s > review/compiler/loom-implied-target/followup/readers-main.log 2>&1
timeout 150 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/(invariance_readonly|optional_widening_(class|declared|fresh)|node_options_widening)' -count=1 -v -timeout 90s > review/compiler/loom-implied-target/followup/widening-fixtures-main.log 2>&1
timeout 120 go test ./internal/ir -run '^TestCallTargetReaders$' -count=1 -v -timeout 90s > review/compiler/loom-implied-target/followup/reader-guard-main.log 2>&1
timeout 120 go vet ./internal/lower > review/compiler/loom-implied-target/followup/vet-main.log 2>&1
