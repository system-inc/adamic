set -u
# A static method named like an optional field: Node, native release and sanitized, and JavaScript, at this tree.
source /opt/adamic-block/environment.sh
source cloud/sandboxed-job.sh
go build -o "$TMPDIR/adamic" ./cmd/adamic || { echo "probe: FAIL build"; exit 1; }
d="$TMPDIR/m"; mkdir -p "$d"; cp scratch-k24/method_probe.ts "$d/main.ts"
echo "--- node"; ( cd "$d" && node --experimental-strip-types main.ts 2>&1 | head -4; echo "exit=${PIPESTATUS[0]}" )
for mode in release sanitized; do
	flag=""; [ "$mode" = sanitized ] && flag="--sanitize"
	( cd "$d" && "$TMPDIR/adamic" build main.ts -o "bin-$mode" $flag > "b-$mode.log" 2>&1 ) || { echo "--- $mode build: $(head -2 "$d/b-$mode.log")"; continue; }
	echo "--- native $mode"; ( cd "$d" && ASAN_OPTIONS=detect_leaks=0 "./bin-$mode" 2>&1 | head -4; echo "exit=${PIPESTATUS[0]}" )
done
echo "--- js"; ( cd "$d" && "$TMPDIR/adamic" js main.ts > main.mjs 2> js.log; echo "js compile exit=$?"; head -2 js.log )
echo PROBE DONE
