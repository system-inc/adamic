#!/usr/bin/env bash
# Correctness: the harness's own suite, then each wrong version applied in turn, each expected red.
set -u
source /opt/adamic-block/environment.sh && source cloud/sandboxed-job.sh || { echo "FAIL setup"; exit 1; }
f=0
run_suite() { (cd stage3/firstlink && python3 -I -B -m unittest test_report) > "$TMPDIR/suite.log" 2>&1; }
if run_suite; then echo "PASS suite $(tail -1 "$TMPDIR/suite.log")"; else tail -30 "$TMPDIR/suite.log"; echo "FAIL suite"; f=1; fi
if go test -count=1 -run '^TestFirstLinkReportPython$' ./stage3/lane/ > "$TMPDIR/wrap.log" 2>&1; then echo "PASS wrapper"; else tail -20 "$TMPDIR/wrap.log"; echo "FAIL wrapper"; f=1; fi
for p in tmp-audit/wrong-*.patch; do
	n=$(basename "$p" .patch)
	git apply -p1 "$p" || { echo "WRONG $n did not apply"; f=1; continue; }
	if run_suite; then echo "WRONG $n SURVIVES"; f=1; else echo "WRONG $n red: $(grep -E '^(FAIL|ERROR):' "$TMPDIR/suite.log" | tr '\n' ' ' | cut -c1-300) $(grep -oE 'test_report.py\", line [0-9]+' "$TMPDIR/suite.log" | head -3 | tr '\n' ' ')"; fi
	git apply -R -p1 "$p" || { echo "WRONG $n did not reverse"; f=1; }
done
git diff --quiet -- stage3 && echo "PASS tree restored" || { echo "FAIL tree not restored"; f=1; }
[ $f = 0 ] && echo VERDICT-ALL PASS || { echo VERDICT-ALL FAIL; exit 1; }
