#!/usr/bin/env bash
set -euo pipefail
cd /workspace/adamic
source /workspace/adamic-tools/env.sh
export ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-25-corpus
export ADAMIC_COVERAGE_ARTIFACTS=/workspace/wave-25-landing-second/dependency
export ADAMIC_COVERAGE_REPOSITORY_MANIFEST=/workspace/wave-25-landing-second/fifth/repository.manifest
export ADAMIC_COVERAGE_COMPILER_MANIFEST=/workspace/wave-25-landing-second/fifth/compiler.manifest
go test ./stage1/cohere/typeaware -run '^TestCoverageAgreementAndMutants$' -count=1 -v -timeout 30m > /workspace/wave-25-validation/landing-third-dependency.log 2>&1
