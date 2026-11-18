# #k8ed1f2: adaptation 99-readonly-family3-views. apply with and without it, a second run that changes nothing, the
# files it changes, stock tsc over tsc's whole program on both trees, and stage3/lane. Run after environment.sh,
# sandboxed-job.sh, sandboxed_typescript_source, sandboxed_npm stage3/api and sandboxed_products ./stage3/lane.
# Never merged.
set -u
bad=0
D=stage3/adapt/99-readonly-family3-views
echo "on $(hostname), $(nproc) cores, load $(cut -d' ' -f1-3 /proc/loadavg)"
step() { local name=$1; shift; local start=$(date +%s); "$@" > "$TMPDIR/p-$name.log" 2>&1; local code=$?; echo "$([ $code = 0 ] && echo PASS || echo FAIL) $name exit $code, $(( $(date +%s) - start )) s"; [ $code = 0 ] || { grep -E 'Error|error|FAIL|panic|_test.go:[0-9]+:' "$TMPDIR/p-$name.log" | head -12 | cut -c1-300; bad=1; }; return $code; }
A="$TMPDIR/with"; B="$TMPDIR/without"
step apply-with bash stage3/apply.sh --build-product "$A"
grep -E 'family3' "$TMPDIR/p-apply-with.log" | head -3 | cut -c1-200
mv "$D" "$TMPDIR/aside"
step apply-without bash stage3/apply.sh --build-product "$B"
mv "$TMPDIR/aside" "$D"
before=$(cd "$A" && find src -type f -name '*.ts' -exec sha256sum {} + | sort | sha256sum)
NODE_PATH="$PWD/stage3/api/node_modules" CENSUS_TYPESCRIPT="$PWD/stage3/api/node_modules/typescript" node "$D/adapt.cjs" "$A" > "$TMPDIR/p-again.log" 2>&1
code=$?
after=$(cd "$A" && find src -type f -name '*.ts' -exec sha256sum {} + | sort | sha256sum)
lines=$(wc -l < "$TMPDIR/p-again.log"); kept=$(grep -c 'already adapted' "$TMPDIR/p-again.log")
[ $code = 0 ] && [ "$before" = "$after" ] && echo "PASS second run: exit 0, no file changed, $kept of $lines lines already adapted" || { echo "FAIL second run: exit $code, changed $([ "$before" = "$after" ] && echo no || echo yes)"; tail -5 "$TMPDIR/p-again.log"; bad=1; }
changed=$(diff -rq "$B/src" "$A/src" | sed -E 's#^Files [^ ]*/src/([^ ]+) and .*#\1#' | sort | tr '\n' ' ')
echo "CHANGED $(echo $changed | wc -w) files: $changed"
node proof/performance/k8ed1f2/tscheck.mjs "$B" | sort > "$TMPDIR/tsc-without"
node proof/performance/k8ed1f2/tscheck.mjs "$A" | sort > "$TMPDIR/tsc-with"
new=$(comm -13 "$TMPDIR/tsc-without" "$TMPDIR/tsc-with" | wc -l); gone=$(comm -23 "$TMPDIR/tsc-without" "$TMPDIR/tsc-with" | wc -l)
echo "TSC without $(wc -l < "$TMPDIR/tsc-without"), with $(wc -l < "$TMPDIR/tsc-with"), new $new, gone $gone"
[ "$new" = 0 ] && echo "PASS stock tsc: 0 new" || { echo "FAIL stock tsc: $new new"; comm -13 "$TMPDIR/tsc-without" "$TMPDIR/tsc-with" | head -10 | cut -c1-300; bad=1; }
step lane go test -count=1 -timeout 30m -skip 'TestShard_|TestProduct_' ./stage3/lane/...
echo "VERDICT $([ $bad = 0 ] && echo OK || echo FAIL)"
exit $bad
