#!/usr/bin/env bash
set -euo pipefail
cd /workspace/adamic
source /workspace/adamic-tools/env.sh
export ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-25-corpus
export ADAMIC_WAVE25_ARTIFACTS=/workspace/wave-25-landing-second/original
export ADAMIC_WAVE25_NEXT_ARTIFACTS=/workspace/wave-25-landing-second/next
export ADAMIC_WAVE25_THIRD_ARTIFACTS=/workspace/wave-25-landing-second/third
export ADAMIC_WAVE25_FOURTH_ARTIFACTS=/workspace/wave-25-landing-second/fourth
export ADAMIC_WAVE25_FIFTH_ARTIFACTS=/workspace/wave-25-landing-second/fifth
export ADAMIC_WAVE25_FIFTH_REFUSALS=/workspace/wave-25-area-next-refusals
mkdir -p /workspace/wave-25-landing-second
go test ./stage1/cohere/typeaware -run '^TestWave25(AgreementAndMutants|NextAgreementAndMutants|ThirdAgreementAndMutants|FourthAgreementAndMutants|FifthAgreementAndMutants|FifthBooleanRefusals)$' -count=1 -v -timeout 30m > /workspace/wave-25-validation/area-next-rules.log 2>&1
bash /workspace/wave-25-validation/area-next-dependency-gate.sh > /workspace/wave-25-validation/area-next-dependency-runner.log 2>&1
go test ./bridge/tsgo/... -count=1 -v -timeout 30m > /workspace/wave-25-validation/area-next-bridge.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestTheOracleCatchesOneByte|TestNativeAgreesWithNode/internal/oracle/testdata/(closures|classes|strings|array_from)\.a$' -count=1 -v -timeout 10m > /workspace/wave-25-validation/area-next-node.log 2>&1
go vet ./... > /workspace/wave-25-validation/area-next-vet.log 2>&1
