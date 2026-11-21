set -u
# #s03ndgw's scout: where each program stops at main, in what words, and what Node prints.
source /opt/adamic-block/environment.sh
source cloud/sandboxed-job.sh
go build -o "$TMPDIR/adamic" ./cmd/adamic || { echo FAIL build; exit 1; }
for f in stage3/interface-downcasts/fx7/p67.ts \
	internal/oracle/testdata/review/agree/fxspptb_oct9_native_p67_static_class_union_member.ts \
	internal/oracle/testdata/view_array_object_stops/field_empty_interface.ts; do
	echo "=================== $f"
	d="$TMPDIR/w/$(basename "$f" .ts)"; mkdir -p "$d"; cp "$f" "$d/main.ts"
	echo "--- node"; ( cd "$d" && node --experimental-strip-types main.ts 2>&1 | head -8; echo "exit=${PIPESTATUS[0]}" )
	echo "--- native build"; ( cd "$d" && "$TMPDIR/adamic" build main.ts -o bin > build.log 2>&1; echo "exit=$?"; head -6 build.log )
	if [ -x "$d/bin" ]; then echo "--- native run"; ( cd "$d" && ./bin 2>&1 | head -8; echo "exit=${PIPESTATUS[0]}" ); fi
	echo "--- js"; ( cd "$d" && "$TMPDIR/adamic" js main.ts > main.mjs 2> js.log; echo "exit=$?"; head -4 js.log )
	if [ -s "$d/main.mjs" ]; then ( cd "$d" && node main.mjs 2>&1 | head -8; echo "run exit=${PIPESTATUS[0]}" ); fi
	echo "--- the view check in the C"; ( cd "$d" && "$TMPDIR/adamic" c main.ts 2>/dev/null | grep -n -i "view\|field read failed\|is not a" | head -12 )
done
echo scout done
