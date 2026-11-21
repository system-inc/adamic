set -u
# Compiler's probe on #s03ndgw at piece 1: does native's plain view read let a class constructor through wrongly?
source /opt/adamic-block/environment.sh
source cloud/sandboxed-job.sh
go build -o "$TMPDIR/adamic" ./cmd/adamic || { echo FAIL build; exit 1; }
for name in widened_write wide_static; do
	echo "--- js"; d0="$TMPDIR/p/$name"; mkdir -p "$d0"; cp "scratch-s03p1/probe2/$name.ts" "$d0/main.ts"; ( cd "$d0" && "$TMPDIR/adamic" js main.ts > main.mjs 2> js.log; echo "js compile exit=$?"; head -3 js.log )
	d="$TMPDIR/p/$name"; mkdir -p "$d"; cp "scratch-s03p1/probe2/$name.ts" "$d/main.ts"
	echo "=================== $name"
	echo "--- node"; ( cd "$d" && node --experimental-strip-types main.ts 2>&1 | head -6; echo "exit=${PIPESTATUS[0]}" )
	for mode in release sanitized; do :;
		flag=""; [ "$mode" = sanitized ] && flag="--sanitize"
		( cd "$d" && "$TMPDIR/adamic" build main.ts -o "bin-$mode" $flag > "build-$mode.log" 2>&1 ) || { echo "--- $mode build: $(head -3 "$d/build-$mode.log")"; continue; }
		echo "--- native $mode"; ( cd "$d" && ASAN_OPTIONS=detect_leaks=0 "./bin-$mode" 2>&1 | head -12; echo "exit=${PIPESTATUS[0]}" )
	done
done
echo probe done
