#!/bin/sh
# One hour of fuzzing in batches of 100 seeds from 200000, each batch's summary kept.
. /tmp/adamic-gate/env.sh
cd /tmp/adamic-gate/fz
start=$(date +%s)
seed=200000
mkdir -p /tmp/adamic-gate/fuzz-hour
while [ $(( $(date +%s) - start )) -lt 3300 ]; do
	/tmp/adamic-gate/adamic-fuzz-r -root . -seed $seed -count 100 -parallel 4 -work /tmp/adamic-gate/fuzz-hour/work -findings /tmp/adamic-gate/fuzz-hour/findings > /tmp/adamic-gate/fuzz-hour/batch-$seed.txt 2>&1
	echo "batch $seed exit $? at $(( $(date +%s) - start ))s"
	seed=$((seed + 100))
done
echo "done at $(( $(date +%s) - start ))s, next seed $seed"
