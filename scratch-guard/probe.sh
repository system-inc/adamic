set -u
source /opt/adamic-block/environment.sh
source cloud/sandboxed-job.sh
# #8nj5pn4, Emit's head, the probe at main with no guard: the narrowed-arm programs through Node, the JavaScript
# backend and native under ASan and UBSan (cloud/probe/run.sh), then each lie's native stderr whole, so ASan's report
# on out_of_bounds.ts is read, not summarized, and the C of each program's read of a, which is the uniform road.
ok=1
step() { name=$1; shift; "$@" > "$TMPDIR/step.log" 2>&1; code=$?; if [ $code -eq 0 ]; then echo "PASS $name"; else echo "FAIL $name (exit $code)"; tail -40 "$TMPDIR/step.log"; ok=0; fi; }

echo "box: $(hostname) at $(git rev-parse HEAD 2>/dev/null)"
step npm sandboxed_npm stage3/api stage3/typescript
step "mkdir probe" mkdir -p "$TMPDIR/probe" "$TMPDIR/logs" "$TMPDIR/release"
for file in scratch-guard/probe/*.ts.txt; do
	step "copy $(basename "$file" .txt)" cp "$file" "$TMPDIR/probe/$(basename "$file" .txt)"
done
step "build adamic" go build -o "$TMPDIR/adamic" ./cmd/adamic
bash cloud/probe/run.sh "$TMPDIR/probe" --logs "$TMPDIR/logs"
echo "run.sh: exit $?"
for file in "$TMPDIR"/probe/*.ts; do
	name=$(basename "$file")
	echo "== $name"
	echo "-- node stdout: $(tr '\n' ' ' < "$TMPDIR/logs/$name/node.stdout")"
	echo "-- sanitized native stdout: $(tr '\n' ' ' < "$TMPDIR/logs/$name/head.native.stdout" 2>/dev/null)"
	echo "-- sanitized native stderr, its first 60 lines:"
	head -60 "$TMPDIR/logs/$name/head.native.stderr" 2>/dev/null
	if "$TMPDIR/adamic" build "$file" -o "$TMPDIR/release/$name.bin" > "$TMPDIR/release/$name.compile" 2>&1; then
		"$TMPDIR/release/$name.bin" > "$TMPDIR/release/$name.stdout" 2> "$TMPDIR/release/$name.stderr"
		echo "-- release native: exit $?, stdout $(tr '\n' ' ' < "$TMPDIR/release/$name.stdout"), stderr $(head -3 "$TMPDIR/release/$name.stderr" | tr '\n' ' ')"
	else
		echo "-- release native: compile stopped: $(head -3 "$TMPDIR/release/$name.compile")"
	fi
	"$TMPDIR/adamic" c "$file" > "$TMPDIR/release/$name.c" 2>&1
	echo "-- C reads and writes of a: $(grep -c '"a"' "$TMPDIR/release/$name.c")"
	grep -n '"a"' "$TMPDIR/release/$name.c" | cut -c1-400 | head -8
done
[ $ok -eq 1 ] && echo PASS || { echo FAIL; exit 1; }
