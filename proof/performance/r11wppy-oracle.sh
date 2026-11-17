# #r11wppy item 2's oracle run, sealed, the head's parent at the proof commit. Lists every top-level fail, so
# main's own sealed reds (the assignment-proof family, stock TypeScript's loader) read apart. Never merged.
set -u
echo "on $(hostname), $(nproc) cores, load $(cut -d' ' -f1-3 /proc/loadavg)"
start=$(date +%s)
go test -count=1 -timeout 40m ./internal/oracle/ > "$TMPDIR/o.log" 2>&1
code=$?
echo "oracle exit $code, $(( $(date +%s) - start )) s"
echo "top-level fails: $(grep -oE '^--- FAIL: [A-Za-z0-9_]+ ' "$TMPDIR/o.log" | cut -d' ' -f3 | sort -u | tr '\n' ' ')"
grep -E '^ +[a-z_]+_test.go:[0-9]+:' "$TMPDIR/o.log" | grep -v 'stock TypeScript' | head -12 | cut -c1-260
echo VERDICT DONE
