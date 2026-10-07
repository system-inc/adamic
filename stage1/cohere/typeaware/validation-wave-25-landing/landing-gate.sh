#!/usr/bin/env bash
set -euo pipefail
cd /workspace/adamic
source /workspace/adamic-tools/env.sh
export ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-25-corpus
export ADAMIC_WAVE25_ARTIFACTS=/workspace/wave-25-landing/original
export ADAMIC_WAVE25_NEXT_ARTIFACTS=/workspace/wave-25-landing/next
export ADAMIC_WAVE25_THIRD_ARTIFACTS=/workspace/wave-25-landing/third
export ADAMIC_WAVE25_FOURTH_ARTIFACTS=/workspace/wave-25-landing/fourth
export ADAMIC_WAVE25_FIFTH_ARTIFACTS=/workspace/wave-25-landing/fifth
export ADAMIC_WAVE25_FIFTH_REFUSALS=/workspace/wave-25-landing/refusals
mkdir -p /workspace/wave-25-landing
go test ./stage1/cohere/typeaware -run '^TestWave25(AgreementAndMutants|NextAgreementAndMutants|ThirdAgreementAndMutants|FourthAgreementAndMutants|FifthAgreementAndMutants|FifthBooleanRefusals)$' -count=1 -v -timeout 30m > /workspace/wave-25-validation/landing-rules.log 2>&1
go test ./bridge/tsgo/... -count=1 -v -timeout 30m > /workspace/wave-25-validation/landing-bridge.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestTheOracleCatchesOneByte|TestNativeAgreesWithNode/internal/oracle/testdata/(closures|classes|strings|array_from)\.a$' -count=1 -v -timeout 10m > /workspace/wave-25-validation/landing-node.log 2>&1
go vet ./... > /workspace/wave-25-validation/landing-vet.log 2>&1
