set -u
source /opt/adamic-block/environment.sh
source cloud/sandboxed-job.sh
# #2hzryxh, family 65's typed half on road (b): copied-slot presence bits for a spread, and `in` with a fixed key on
# a proven typed receiver answering from them. PART=0: step 0, the six programs through cloud/probe/run.sh at main,
# each stop's whole words printed.
part=${PART:-a}
ok=1
step() { name=$1; shift; "$@" > "$TMPDIR/step.log" 2>&1; code=$?; if [ $code -eq 0 ]; then echo "PASS $name"; else echo "FAIL $name (exit $code)"; grep -v "gate cache\|^=== " "$TMPDIR/step.log" | tail -40; ok=0; fi; }

# step0 lays the six step-0 programs in a directory: four are main's spread_overwrite fixtures under round 2's
# names, copied here from the tree, never from a copy.
step0() {
	cp internal/oracle/testdata/spread_overwrite/number_source.ts "$1/undefined-number-source.ts"
	cp internal/oracle/testdata/spread_overwrite/optional_string.ts "$1/empty-object.ts"
	cp internal/oracle/testdata/spread_overwrite/string_source.ts "$1/undefined-source.ts"
	cp internal/oracle/testdata/spread_overwrite/optional_number.ts "$1/65-leading-overwrite-empty-object.ts"
	cp scratch-f65/programs/copied-empty-object.ts scratch-f65/programs/65-leading-overwrite.ts "$1/"
}

# table runs a directory through run.sh and prints the table, then each program's Node output and every compile
# stop's whole words.
table() {
	side=$1 programs=$2
	bash cloud/probe/run.sh "$programs" --logs "$TMPDIR/logs-$side" > "$TMPDIR/run-$side.log" 2>&1
	echo "run.sh at $side: exit $?"
	grep -E "^\| |^- head|^Probes" "$TMPDIR/run-$side.log"
	grep -qE "^\| [0-9a-z]" "$TMPDIR/run-$side.log" || { tail -30 "$TMPDIR/run-$side.log"; ok=0; }
	for directory in "$TMPDIR/logs-$side"/*.ts; do
		echo "$side $(basename "$directory"): Node $(tr '\n' ' ' < "$directory/node.stdout")"
		for stage in js-compile native-compile; do
			[ -s "$directory/head.$stage.stderr" ] && echo "$side $(basename "$directory") $stage: $(tr '\n' ' ' < "$directory/head.$stage.stderr" | cut -c1-900)"
		done
		for backend in js native; do
			[ -f "$directory/head.$backend.stdout" ] && echo "$side $(basename "$directory") $backend: $(tr '\n' ' ' < "$directory/head.$backend.stdout")"
		done
	done
}

echo "box: $(hostname) at $(git rev-parse HEAD 2>/dev/null) PART=$part"
step npm sandboxed_npm stage3/api stage3/typescript

if [ "$part" = 0 ]; then
	mkdir -p "$TMPDIR/step0"
	step0 "$TMPDIR/step0"
	sha256sum "$TMPDIR/step0"/*.ts | sed "s|$TMPDIR/step0/||"
	table main "$TMPDIR/step0"
fi
if [ "$part" = i ]; then
	mkdir -p "$TMPDIR/all/lib"
	cp scratch-f65/programs/*.ts "$TMPDIR/all/"
	cp scratch-f65/programs/lib/*.ts "$TMPDIR/all/lib/"
	table main "$TMPDIR/all"
fi
[ $ok -eq 1 ] && echo PASS || { echo FAIL; exit 1; }
