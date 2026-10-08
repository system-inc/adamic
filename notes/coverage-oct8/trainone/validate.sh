#!/usr/bin/env bash
# Run only after every mutation has been restored.
set -u
source /workspace/adamic-tools/env.sh
unit=notes/coverage-oct8/trainone/evidence
ADAMIC_GATE_UNCACHED=1 go test ./internal/native -run 'TestRecordsAgainstNode' -count=1 -timeout 30m > "$unit/validation-native.log" 2>&1
native_status=$?
printf 'native exit=%s\n' "$native_status"
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestSyntaxSubstrMutants|TestEntriesProvenance|TestEntriesRuntimeReadiness|TestNativeAgreesWithNode/internal/oracle/testdata/(optional_indexing|syntax_substr|entries_)' -count=1 -timeout 30m -v > "$unit/validation-oracle.log" 2>&1
oracle_status=$?
printf 'oracle exit=%s\n' "$oracle_status"
go vet ./... > "$unit/vet.log" 2>&1
vet_status=$?
printf 'vet exit=%s\n' "$vet_status"
gofmt -l cmd internal > "$unit/format.log" 2>&1
format_status=$?
printf 'format exit=%s\n' "$format_status"
git diff --check > "$unit/diff-check.log" 2>&1
diff_status=$?
printf 'diff-check exit=%s\n' "$diff_status"
exit "$((native_status || oracle_status || vet_status || format_status || diff_status))"
