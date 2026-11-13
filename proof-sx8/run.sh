#!/bin/bash
# Proof for #sx8wdg7: the meter Python suites whole, then each named mutant red.
set -u
cd "$(dirname "$0")/.." || exit 9
status=0
run() { (cd stage3/meter && python3 -B -m unittest "$@" 2>&1); }
echo "== report_test and entry_test whole"
out=$(run report_test entry_test node_readiness_test); rc=$?
echo "$out" | tail -n 6
[ $rc -eq 0 ] && echo "PASS suite" || { echo "FAIL suite rc=$rc"; echo "$out" | grep -E '^(FAIL|ERROR):'; status=1; }
mutate() { # name file before after case...
  local name=$1 file=$2 before=$3 after=$4; shift 4
  cp "$file" "$file.orig"
  python3 -I -c 'import sys;p,b,a=sys.argv[1:];s=open(p).read();assert s.count(b)==1,b;open(p,"w").write(s.replace(b,a))' "$file" "$before" "$after" || { echo "FAIL $name: mutant did not apply"; status=1; return; }
  for case in "$@"; do
    out=$(run "report_test.$case"); rc=$?
    if [ $rc -ne 0 ] && echo "$out" | grep -q "^FAIL: "; then echo "PASS $name red on $case"; else echo "FAIL $name survived on $case rc=$rc"; echo "$out" | tail -n 4; status=1; fi
  done
  mv "$file.orig" "$file"
}
mutate body-count stage3/meter/report.py "row['lowering'] = artifact_lowers_whole(artifact)" "row['lowering'] = artifact_lowers_body(artifact)" RootOrderTests.test_artifact_body_only_does_not_lower_sx8wdg7
mutate old-label stage3/meter/report.py 'f"Lowering, {label}:' 'f"Body lowering, {label}:' ReportTests.test_pair_preserves_area_fields_and_prints_four_numbers_first RootOrderTests.test_daily_table_keeps_artifacts_beside_counts RootOrderTests.test_artifact_body_only_does_not_lower_sx8wdg7
mutate tail-and-measured stage3/meter/report.py "    return (artifact_lowers_body(artifact) and not artifact.get('entry_tail_stops')
            and artifact.get('entry_whole_program_passes_measured', False))" "    return artifact_lowers_body(artifact)" RootOrderTests.test_artifact_tail_stop_withholds_lowering_loom03dg
git diff --quiet || { echo "FAIL tree left modified"; status=1; }
echo "VERDICT $([ $status -eq 0 ] && echo PASS || echo FAIL)"
exit $status
