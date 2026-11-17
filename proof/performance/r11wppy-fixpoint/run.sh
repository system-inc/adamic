#!/usr/bin/env bash
# #r11wppy item 1: the readonly fixpoint table for tsc's shared emptyArray and resolvingEmptyArray, read in tsc's
# adapted tree at this commit's parent. Run from the tree's root after environment.sh, sealed-job.sh,
# sealed_typescript_source and sealed_npm stage3/api. The second pass adds shorthand properties, class members that
# miss their base, alias bodies, written slots as unions, Dropped slots, sentinels by slot and unmatched sites
# explained. Never merged.
set -u
: "${TMPDIR:?TMPDIR must name a writable directory}"
stamp() { echo "==== $(date -u +%Y-%m-%dT%H:%M:%SZ) $*"; }

A="$TMPDIR/adapted"
OUT="$TMPDIR/out"

stamp "apply: the adapted tree into $A"
bash stage3/apply.sh "$A" > "$TMPDIR/apply.log" 2>&1 || { echo "VERDICT FAIL apply"; tail -5 "$TMPDIR/apply.log"; exit 3; }

stamp "fixpoint: node $(node --version)"
node proof/performance/r11wppy-fixpoint/fixpoint.mjs "$A" proof/performance/scan03/never-sites.tsv "$OUT"
status=$?

stamp "table"
if [ -f "$OUT/table.md" ]; then
	cat "$OUT/table.md"
else
	echo "no table.md under $OUT"
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
	5) echo "VERDICT FAIL the script hit errors while classifying or editing (listed under Script errors)" ;;
	*) echo "VERDICT FAIL fixpoint.mjs exited $status" ;;
esac
[ "$status" -eq 0 ] || exit 1
