# #k8ed1f2's arrow-at-site rule: internal/oracle whole, its reds read against main's sealed set. Run after
# environment.sh, sandboxed-job.sh, sandboxed_npm stage3/api and sandboxed_products ./internal/oracle. Never merged.
set -u
echo "on $(hostname), $(nproc) cores, load $(cut -d' ' -f1-3 /proc/loadavg)"
start=$(date +%s)
go test -count=1 -timeout 40m ./internal/oracle/ > "$TMPDIR/o.log" 2>&1
echo "oracle exit $?, $(( $(date +%s) - start )) s"
grep -E '^(ok|FAIL|panic)' "$TMPDIR/o.log" | head -5
fails=$(grep -oE '^--- FAIL: [A-Za-z0-9_]+ ' "$TMPDIR/o.log" | cut -d' ' -f3 | sort -u | tr '\n' ' ')
echo "fails: $fails"
other=$(echo "$fails" | tr ' ' '\n' | grep -v '^$' | grep -vE '^(TestAssignmentProof|TestStringLengthJoinMemory$)' | tr '\n' ' ')
[ -z "$other" ] && { echo "PASS oracle: only main's sealed set"; echo "VERDICT OK"; } || { echo "FAIL oracle beyond main's sealed set: $other"; grep -A8 -E "^--- FAIL: $(echo $other | cut -d' ' -f1)" "$TMPDIR/o.log" | head -20 | cut -c1-300; echo "VERDICT FAIL"; exit 1; }
