source /workspace/adamic-tools/env.sh
export TMPDIR=/tmp/defend-taste
for name in TestTasteRepresentationLimitsStayExplicit TestEnumLimitsStayLoud TestTypedArrayGaps TestTypedArraysLower TestViewObjectContractsAreAvailableToEraser TestMixedUnionContractRecursiveMember; do
 timeout 120 go test -json -count=1 -timeout 90s -coverpkg=./internal/lower -coverprofile=review/test-defend/internal-lower-taste_representation/$name.cover ./internal/lower/ -run "^$name$" > review/test-defend/internal-lower-taste_representation/$name.cover.log 2>&1
 echo "$name $?"
done
