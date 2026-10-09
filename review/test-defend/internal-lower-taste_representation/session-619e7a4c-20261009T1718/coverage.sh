evidence_dir="$(cd -- "$(dirname -- "$0")" && pwd)"
source /workspace/adamic-tools/env.sh
export TMPDIR=/tmp/defend-taste
for name in TestTasteRepresentationLimitsStayExplicit TestEnumLimitsStayLoud TestTypedArrayGaps TestTypedArraysLower TestViewObjectContractsAreAvailableToEraser TestMixedUnionContractRecursiveMember; do
 timeout 120 go test -json -count=1 -timeout 90s -coverpkg=./internal/lower -coverprofile="$evidence_dir/$name.cover" ./internal/lower/ -run "^$name$" > "$evidence_dir/$name.cover.log" 2>&1
 echo "$name $?"
done
