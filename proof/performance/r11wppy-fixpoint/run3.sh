#!/usr/bin/env bash
# #r11wppy item 1, third pass: the readonly fixpoint with every Public slot held as tsc wrote it, read in tsc's adapted
# tree at this commit's parent. Run from the tree's root after environment.sh, sealed-job.sh, sealed_typescript_source
# and sealed_npm stage3/api. The rounds leave Public slots alone, then blocking puts back each slot whose readonly
# value reaches a Held or Dropped slot, so what stays readonly is the Local and Internal change, and each Held slot
# shows the listed sites it holds. Never merged.
set -u
: "${TMPDIR:?TMPDIR must name a writable directory}"
stamp() { echo "==== $(date -u +%Y-%m-%dT%H:%M:%SZ) $*"; }

A="$TMPDIR/adapted"
OUT="$TMPDIR/out3"

stamp "apply: the adapted tree into $A"
bash stage3/apply.sh "$A" > "$TMPDIR/apply.log" 2>&1 || { echo "VERDICT FAIL apply"; tail -5 "$TMPDIR/apply.log"; exit 3; }

stamp "fixpoint: node $(node --version), --hold-public"
node proof/performance/r11wppy-fixpoint/fixpoint.mjs "$A" proof/performance/scan03/never-sites.tsv "$OUT" --hold-public
status=$?

stamp "table"
if [ -f "$OUT/table3.md" ]; then
	cat "$OUT/table3.md"
else
	echo "no table3.md under $OUT"
fi

stamp "edits"
if [ -f "$OUT/edits.diff" ]; then
	echo "edits.diff: $(wc -l < "$OUT/edits.diff" | tr -d ' ') lines"
	cat "$OUT/edits.diff"
else
	echo "no edits.diff under $OUT"
fi

stamp "verdict"
case $status in
	0) echo "VERDICT OK" ;;
	4) echo "VERDICT FAIL no fixpoint within 80 rounds" ;;
	5) echo "VERDICT FAIL the script hit errors while classifying, blocking or editing (listed under Script errors)" ;;
	6) echo "VERDICT FAIL blocking still put slots back after 40 rounds" ;;
	7) echo "VERDICT FAIL a slot left readonly is not Local or Internal (see the internal manifest)" ;;
	*) echo "VERDICT FAIL fixpoint.mjs exited $status" ;;
esac
[ "$status" -eq 0 ] || exit 1
