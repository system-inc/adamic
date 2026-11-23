set -u
source /opt/adamic-block/environment.sh
source cloud/sandboxed-job.sh
# #hh1cezv: 162 round 6 and 55 round 4 together, the direct road and host-error key parity probes,
# each held to Node on both backends, native sanitized.
echo "box: $(hostname) at $(git rev-parse HEAD 2>/dev/null)"
bash cloud/probe/run.sh scratch-rv2/p --logs "$TMPDIR/logs" > "$TMPDIR/table.txt" 2>&1; code=$?
cat "$TMPDIR/table.txt"
for log in "$TMPDIR/logs"/*; do
	case "$log" in *build*|*submodules*) continue ;; esac
	[ -f "$log" ] || continue
	echo "-- $(basename "$log")"
	head -c 600 "$log"
	echo
done
exit $code
