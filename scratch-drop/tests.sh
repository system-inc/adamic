#!/usr/bin/env bash
# scratch for #6c4swvt, never cut: the drop head's tests and mutant on a keeper box.
source ~/adamic-tools/env.sh; export TMPDIR="${TMPDIR:-/tmp}"
L="$TMPDIR/drop-logs"; mkdir -p "$L"
echo "gofmt: [$(gofmt -l internal/lower internal/oracle | tr '\n' ' ')]"
go vet ./internal/lower ./internal/oracle > "$L/vet" 2>&1; echo "vet exit $?"; tail -n 5 "$L/vet"
go test ./internal/lower > "$L/lower" 2>&1; c=$?; echo "lower exit $c"; [ $c = 0 ] || grep -E '^(--- FAIL|FAIL|panic)|_test.go:[0-9]+:' "$L/lower" | head -n 40 | cut -c1-600
go test ./internal/oracle > "$L/oracle" 2>&1; c=$?; echo "oracle exit $c"; [ $c = 0 ] || grep -E '^(--- FAIL|FAIL|panic)|_test.go:[0-9]+:' "$L/oracle" | head -n 40 | cut -c1-600
cp internal/lower/records_named.go "$L/records_named.go"
sed -i 's/, ir.Array: true, ir.Record: true, ir.Object: true}/, ir.Array: true}/' internal/lower/records_named.go
if cmp -s internal/lower/records_named.go "$L/records_named.go"; then echo "FAIL mutant didn't apply"; else
  go test ./internal/oracle -run 'TestOptionsBaseObjectField' -count=1 > "$L/mutant" 2>&1; c=$?
  n=$(grep -c 'stops not yet\|must lower: ' "$L/mutant"); echo "mutant (drop ir.Record and ir.Object) exit $c, $n subtests stopped at the view"
  grep -E -- '--- FAIL' "$L/mutant" | head -n 12; grep -E '_test.go:[0-9]+:' "$L/mutant" | head -n 2 | cut -c1-400
fi
cp "$L/records_named.go" internal/lower/records_named.go
echo done