#!/usr/bin/env bash
# #zkp356b probe, scratch, never to land: which flags-then-cast shapes the fall-through lowers.
set -u
scratch="$TMPDIR/f4probe"
mkdir -p "$scratch/probes" "$scratch/logs" && echo "PASS mkdir" || { echo "FAIL mkdir"; exit 1; }
for file in f4-probes/*.ts.txt; do
	cp "$file" "$scratch/probes/$(basename "$file" .txt)" || { echo "FAIL copy $file"; exit 1; }
done
bash cloud/probe/run.sh "$scratch/probes" --logs "$scratch/logs"
echo "run.sh exit=$?"
for directory in "$scratch/logs"/*/; do
	echo "=== $(basename "$directory")"
	echo "--- node"
	cat "$directory/node.stdout"
	head -2 "$directory/node.stderr"
	for build in "$directory"/*.native-compile.stderr; do
		echo "--- native compile stderr"
		head -5 "$build"
	done
	for run in "$directory"/*.native.stdout; do
		echo "--- native run"
		cat "$run"
	done
	for build in "$directory"/*.js-compile.stderr; do
		echo "--- js compile stderr"
		head -5 "$build"
	done
done
echo VERDICT printed
