#!/bin/sh
# One hour of fuzzing the allocator, merged into integrate-7, in batches of 100 seeds from 500000.
. /tmp/adamic-gate/env.sh
cd /tmp/adamic-gate/fa
start=$(date +%s)
seed=500000
mkdir -p /tmp/adamic-gate/fuzz-alloc
while [ $(( $(date +%s) - start )) -lt 3300 ]; do
	/tmp/adamic-gate/adamic-fuzz-a -root . -seed $seed -count 100 -parallel 4 -work /tmp/adamic-gate/fuzz-alloc/work -findings /tmp/adamic-gate/fuzz-alloc/findings > /tmp/adamic-gate/fuzz-alloc/batch-$seed.txt 2>&1
	echo "batch $seed exit $? at $(( $(date +%s) - start ))s"
	seed=$((seed + 100))
done
echo "done at $(( $(date +%s) - start ))s, next seed $seed"
