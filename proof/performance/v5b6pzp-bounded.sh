# #v5b6pzp re-cut on main 932a6344 with TestPromiseIsNotYetWhereAnObjectIsHeld's lowerings bounded at three:
# its peak at the unit's 8-way parallelism, then the packages it reaches whole. Run after environment.sh,
# sealed-job.sh, sealed_git and sealed_products ./internal/lower ./internal/oracle ./stage3/fixtures. Never merged.
set -u
bad=0
echo "on $(hostname), $(nproc) cores, $(awk '/MemTotal/ {print int($2/1024) " MB"}' /proc/meminfo), load $(cut -d' ' -f1-3 /proc/loadavg)"
step() { local name=$1; shift; local start=$(date +%s); "$@" > "$TMPDIR/s-$name.log" 2>&1; local code=$?; echo "$([ $code = 0 ] && echo PASS || echo FAIL) $name exit $code, $(( $(date +%s) - start )) s"; [ $code = 0 ] || { grep -E '^(--- FAIL|FAIL|panic|.*_test.go:[0-9]+:)' "$TMPDIR/s-$name.log" | head -14 | cut -c1-300; bad=1; }; return $code; }
peak() {
	local name=$1; shift
	local file="$TMPDIR/mem-$name.peak"
	echo 0 > "$file"
	( max=0; while :; do
		total=$(awk -v root="$$" 'FNR == 1 { pid = 0 } /^Pid:/ { pid = $2 } /^PPid:/ { parent[pid] = $2 } /^VmRSS:/ { rss[pid] = $2 } END { for (p in parent) { q = p; for (i = 0; i < 64 && q != "" && q != 0; i++) { if (q == root) { total += rss[p]; break } q = parent[q] } } print int(total / 1024) }' /proc/[0-9]*/status 2>/dev/null)
		[ "${total:-0}" -gt "$max" ] && { max=$total; echo "$max" > "$file"; }
		sleep 0.2
	done ) &
	local sampler=$!
	local start=$(date +%s)
	"$@" > "$TMPDIR/mem-$name.log" 2>&1
	local code=$?
	kill "$sampler" 2>/dev/null; wait "$sampler" 2>/dev/null
	echo "MEASURE $name: peak $(cat "$file") MB, exit $code, $(( $(date +%s) - start )) s"
	[ $code = 0 ] || { grep -E '^(--- FAIL|FAIL|panic|.*_test.go:[0-9]+:)' "$TMPDIR/mem-$name.log" | head -10 | cut -c1-260; bad=1; }
}
step gofmt bash -c '[ -z "$(gofmt -l internal/oracle/ internal/lower/)" ]'
step vet go vet ./internal/oracle/ ./internal/lower/
GOMAXPROCS=8 go test -c -o "$TMPDIR/oracle.test" ./internal/oracle/ || { echo "VERDICT FAIL build"; exit 1; }
( cd internal/oracle && peak mine_p8 env GOMAXPROCS=8 "$TMPDIR/oracle.test" -test.count=1 -test.run '^TestPromiseIsNotYetWhereAnObjectIsHeld$' -test.parallel 8 )
peak=$(cat "$TMPDIR/mem-mine_p8.peak")
[ "$peak" -lt 9000 ] && echo "PASS bound: peak $peak MB under 9000" || { echo "FAIL bound: peak $peak MB"; bad=1; }
step lower go test -count=1 -timeout 40m ./internal/lower/
step fixtures go test -count=1 -timeout 20m ./stage3/fixtures/...
start=$(date +%s)
go test -count=1 -timeout 40m ./internal/oracle/ > "$TMPDIR/o.log" 2>&1
echo "oracle exit $?, $(( $(date +%s) - start )) s"
fails=$(grep -oE '^--- FAIL: [A-Za-z0-9_]+ ' "$TMPDIR/o.log" | cut -d' ' -f3 | sort -u | tr '\n' ' ')
echo "oracle top-level fails: $fails"
other=$(echo "$fails" | tr ' ' '\n' | grep -v '^$' | grep -vE '^(TestAssignmentProof|TestStringLengthJoinMemory$)' | tr '\n' ' ')
[ -z "$other" ] && echo "PASS oracle: only main's sealed set" || { echo "FAIL oracle beyond main's sealed set: $other"; grep -E '^ +[a-z_]+_test.go:[0-9]+:' "$TMPDIR/o.log" | grep -v 'stock TypeScript' | head -12 | cut -c1-260; bad=1; }
echo "VERDICT $([ $bad = 0 ] && echo OK || echo FAIL)"
exit $bad
