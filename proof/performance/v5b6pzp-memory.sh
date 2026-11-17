# #v5b6pzp's memory: chg_mfqwa06r went red when internal/oracle#52/92 passed its 16384 MB share (18792 MB) on
# 8 cpus. This measures what TestPromiseIsNotYetWhereAnObjectIsHeld adds, at 8-way parallelism like the unit's
# share: the resident memory of every process descending from this script, summed and sampled every 0.2 s, the peak
# kept. Run after environment.sh, sealed-job.sh and sealed_products ./internal/oracle. Never merged.
set -u
echo "on $(hostname), $(nproc) cores, $(awk '/MemTotal/ {print int($2/1024) " MB"}' /proc/meminfo), load $(cut -d' ' -f1-3 /proc/loadavg)"
me=$(id -u)
peak() {
	# Runs the command with a sampler beside it, and prints the summed peak in MB with the command's exit.
	local name=$1; shift
	local file="$TMPDIR/mem-$name.peak"
	echo 0 > "$file"
	( max=0; while :; do
		total=$(awk -v root="$$" 'FNR == 1 { pid = 0 } /^Pid:/ { pid = $2 } /^PPid:/ { parent[pid] = $2 } /^VmRSS:/ { rss[pid] = $2 } END { for (p in parent) { q = p; for (i = 0; i < 64 && q != "" && q != 0; i++) { if (q == root) { total += rss[p]; break } q = parent[q] } } print int(total / 1024) }' /proc/[0-9]*/status 2>/dev/null)
		[ "$total" -gt "$max" ] && { max=$total; echo "$max" > "$file"; }
		sleep 0.2
	done ) &
	local sampler=$!
	local start=$(date +%s)
	"$@" > "$TMPDIR/mem-$name.log" 2>&1
	local code=$?
	kill "$sampler" 2>/dev/null; wait "$sampler" 2>/dev/null
	echo "MEASURE $name: peak $(cat "$file") MB, exit $code, $(( $(date +%s) - start )) s"
	grep -E '^(--- FAIL|FAIL|panic)' "$TMPDIR/mem-$name.log" | head -6 | cut -c1-200
}
export GOMAXPROCS=8
go test -c -o "$TMPDIR/oracle.test" ./internal/oracle/ || { echo "VERDICT FAIL build"; exit 1; }
cd internal/oracle
mine='^TestPromiseIsNotYetWhereAnObjectIsHeld$'
shard='^(TestRecord|TestReflectGet|TestRegExp|TestRestParameterRelation|TestReuse)'
# Warm the result and product caches once, so each measured run reads the same caches.
"$TMPDIR/oracle.test" -test.run "$mine|$shard" -test.parallel 8 > "$TMPDIR/mem-warm.log" 2>&1; echo "warm exit $?"
peak idle sleep 2
peak mine_p8 "$TMPDIR/oracle.test" -test.count=1 -test.run "$mine" -test.parallel 8
peak mine_p1 "$TMPDIR/oracle.test" -test.count=1 -test.run "$mine" -test.parallel 1
peak shard_p8 "$TMPDIR/oracle.test" -test.count=1 -test.run "$shard" -test.parallel 8
peak both_p8 "$TMPDIR/oracle.test" -test.count=1 -test.run "$mine|$shard" -test.parallel 8
peak both_p8_uncached env ADAMIC_GATE_UNCACHED=1 "$TMPDIR/oracle.test" -test.count=1 -test.run "$mine|$shard" -test.parallel 8
peak shard_p8_uncached env ADAMIC_GATE_UNCACHED=1 "$TMPDIR/oracle.test" -test.count=1 -test.run "$shard" -test.parallel 8
# One subtest at a time, each its own process: which probe is heavy, if one is.
for sub in initializer argument union_slot union_then_in literal_field literal_then_in typed_promise union_promise_member awaited unknown_slot promise_all resolve_adopts async_return_adopts union_object_keyword async_arrow_adopts; do
	for extension in a ts; do
		peak "sub_${sub}_$extension" "$TMPDIR/oracle.test" -test.count=1 -test.run "$mine/^${sub}\\.${extension}\$" -test.parallel 1
	done
done
echo "VERDICT OK"
