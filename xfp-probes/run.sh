#!/usr/bin/env bash
# #xfpsqtn scout, scratch: runs the miniatures through Node and both backends at this head (main plus the
# tracing hack), then again with XFP_LIFT=1, which lets a declined never argument through to its next stop.
set -u
# shellcheck disable=SC1091
source cloud/sandboxed-job.sh && echo "PASS sandboxed-job" || { echo "FAIL sandboxed-job"; exit 1; }
scratch="$TMPDIR/xfp"
mkdir -p "$scratch/probes" "$scratch/logs-main" "$scratch/logs-lift" && echo "PASS mkdir $scratch" || { echo "FAIL mkdir $scratch"; exit 1; }
for file in xfp-probes/*.ts.txt; do
	cp "$file" "$scratch/probes/$(basename "$file" .txt)" || { echo "FAIL copy $file"; exit 1; }
done
echo "PASS copy $(ls "$scratch/probes" | wc -l) probes"

show() {
	local logs=$1 build=$2 directory name
	for directory in "$logs"/*/; do
		name=$(basename "$directory")
		echo "=== $name"
		echo "--- node exit stdout, stderr's first 2 lines"
		cat "$directory/node.stdout"
		head -2 "$directory/node.stderr"
		echo "--- js compile stderr"
		cat "$directory/$build.js-compile.stderr"
		if [ -s "$directory/$build.mjs" ]; then
			echo "--- js run stdout, then stderr's first 3 lines"
			cat "$directory/$build.js.stdout" 2>/dev/null
			head -3 "$directory/$build.js.stderr" 2>/dev/null
		fi
		echo "--- native compile stderr, XFP lines dropped (the js compile's are the same)"
		grep -v '^XFP' "$directory/$build.native-compile.stderr"
		if [ -e "$directory/$build.bin" ]; then
			echo "--- native run stdout, then stderr's first 3 lines"
			cat "$directory/$build.native.stdout" 2>/dev/null
			head -3 "$directory/$build.native.stderr" 2>/dev/null
		fi
	done
}

echo "##### main (tracing only)"
bash cloud/probe/run.sh "$scratch/probes" --logs "$scratch/logs-main"
echo "run.sh main exit=$?"
show "$scratch/logs-main" head

echo "##### lifted (XFP_LIFT=1)"
XFP_LIFT=1 bash cloud/probe/run.sh "$scratch/probes" --logs "$scratch/logs-lift"
echo "run.sh lifted exit=$?"
show "$scratch/logs-lift" head
