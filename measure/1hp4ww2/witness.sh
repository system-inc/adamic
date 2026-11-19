#!/usr/bin/env bash
# #1hp4ww2's delete witness, sealed: record delete.a's counted row, prove it, the counts census and the Node pin
# green, then plant M4 (adamic_map_delete's let-go dropped) and show the row red while the Node pin stays green.
# Never lands.
set -uo pipefail
source /opt/adamic-block/environment.sh && source cloud/sandboxed-job.sh || { echo "FAIL sandboxed environment"; exit 1; }
row='^TestCountsAreRecorded_011$/internal/oracle/testdata/records_for_in_cache/delete'
file=internal/oracle/testdata/counts/runtime/internal/oracle/testdata/records_for_in_cache/delete.a.counts
sandboxed_products ./internal/oracle > "$TMPDIR/products.log" 2>&1 && echo "PASS products ./internal/oracle" || { tail -30 "$TMPDIR/products.log"; echo "FAIL products"; exit 1; }
go test -count=1 -timeout 0 -v -run "$row" ./internal/oracle -args -update-counts > "$TMPDIR/record.log" 2>&1; code=$?
grep -E -- '^(=== RUN|--- |ok |FAIL)' "$TMPDIR/record.log" | head -10
[ $code = 0 ] && [ -f "$file" ] || { tail -40 "$TMPDIR/record.log"; echo "FAIL record"; exit 1; }
echo "PASS recorded: $(cat "$file")"
check() { go test -count=1 -timeout 0 -v -run "$2" ./internal/oracle > "$TMPDIR/$1.log" 2>&1; }
for pin in "$row" '^TestCountsFilesCensus$' '^TestRecordForInCachedKeys$'; do
	name=$(echo "$pin" | tr -c 'A-Za-z0-9' _)
	check "green$name" "$pin" && grep -q -- '--- PASS' "$TMPDIR/green$name.log" && echo "PASS $pin on the head" || { grep -E -- '--- (FAIL|PASS)|differ|counts' "$TMPDIR/green$name.log" | head -20 | cut -c1-300; tail -20 "$TMPDIR/green$name.log"; echo "FAIL $pin on the head"; exit 1; }
done
python3 -I - <<'PY' || { echo "FAIL M4 didn't apply"; exit 1; }
path = "internal/native_runtime/c/map.c"
old = "\tadamic_map_forget_key_list(map);\n\tadamic_map_entry *entry = &map->entries[index];"
new = "\tadamic_map_entry *entry = &map->entries[index];"
text = open(path).read()
assert text.count(old) == 1, "the delete's let-go isn't where the head put it"
open(path, "w").write(text.replace(old, new))
PY
echo "M4 applied"
export ADAMIC_GATE_UNCACHED=1
if check mutant "$row"; then grep -E -- '--- ' "$TMPDIR/mutant.log" | head; echo "FAIL M4 survived: delete.a's counted row stayed green"; exit 1; fi
grep -E -- '--- FAIL|counts|want|got|\| internal' "$TMPDIR/mutant.log" | head -12 | cut -c1-400
grep -q -- '--- FAIL: TestCountsAreRecorded_011/internal/oracle/testdata/records_for_in_cache/delete.a' "$TMPDIR/mutant.log" && echo "PASS M4 is caught by delete.a's counted row" || { tail -30 "$TMPDIR/mutant.log"; echo "FAIL M4 red for another reason"; exit 1; }
check mutantnode '^TestRecordForInCachedKeys$' && echo "PASS under M4 TestRecordForInCachedKeys stays green (no output witness, as the mutants file says)" || { grep -E -- '--- FAIL|differ' "$TMPDIR/mutantnode.log" | head | cut -c1-300; echo "NOTE under M4 TestRecordForInCachedKeys went red: the mutants file's no-output claim is wrong"; }
echo "PASS witness"
