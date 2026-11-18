# #715zjph: lowering time over casts and writes, then a CPU profile of 200 casts by 2,000 writes. Run after
# environment.sh and sandboxed-job.sh. Never merged.
set -u
echo "on $(hostname), $(nproc) cores, load $(cut -d' ' -f1-3 /proc/loadavg)"
go test -c -o "$TMPDIR/lower.test" ./internal/lower/ || { echo "VERDICT FAIL build"; exit 1; }
PROBE_PROFILE="$TMPDIR/cpu.pprof" "$TMPDIR/lower.test" -test.run '^TestProbeKindUnionProfile$' -test.v -test.timeout 20m > "$TMPDIR/probe.log" 2>&1
code=$?
grep -E 'TIMING|FAIL|panic' "$TMPDIR/probe.log" | sed 's/^ *//' | cut -c1-200
echo "--- flat"
go tool pprof -top -nodecount=30 "$TMPDIR/lower.test" "$TMPDIR/cpu.pprof" 2>/dev/null | sed -n '1,40p' | cut -c1-170
echo "--- cumulative"
go tool pprof -top -cum -nodecount=45 "$TMPDIR/lower.test" "$TMPDIR/cpu.pprof" 2>/dev/null | sed -n '6,55p' | cut -c1-170
echo "VERDICT $([ $code = 0 ] && echo OK || echo FAIL)"
exit $code
