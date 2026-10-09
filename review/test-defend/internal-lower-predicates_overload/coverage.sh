source /workspace/adamic-tools/env.sh
for row in TestIndirectPredicateOverloadIsPending TestPredicateOverloadCallback TestConditionAssertionAdmission TestPredicateBodyProof TestEveryNeedsCallbackEffects TestPredicateCallbackContracts; do
 ADAMIC_BUILD_CACHE_DIR=/tmp/defend042/cache/coverage timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run "^${row}$" -coverpkg=./internal/lower -coverprofile="/tmp/defend042/${row}.cover" > "/tmp/defend042/${row}-coverage.log" 2>&1
done
