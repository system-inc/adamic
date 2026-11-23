set -u
source /opt/adamic-block/environment.sh
source cloud/sandboxed-job.sh
# #jm2ba55: an own field named like an Object.prototype method, read through a { message: string } view and
# called, at this commit, each held to Node on both backends.
echo "box: $(hostname) at $(git rev-parse HEAD 2>/dev/null)"
bash cloud/probe/run.sh scratch-jm/p --logs "$TMPDIR/logs" > "$TMPDIR/table.txt" 2>&1; code=$?
cat "$TMPDIR/table.txt"
for log in "$TMPDIR/logs"/*; do
	case "$log" in *build*|*submodules*) continue ;; esac
	[ -f "$log" ] || continue
	echo "-- $(basename "$log")"
	head -c 600 "$log"
	echo
done
exit $code
