#!/bin/bash
# run.sh <tree> <program> <n>: one timed cold-unit build, runtime archive warm.
source /tmp/bench/scripts/env.sh
tree=$1 program=$2 n=$3
cache=/tmp/bench/cache-$tree
export ADAMIC_BUILD_CACHE_DIR=$cache
rm -rf "${cache:?}/native-units"
name=$program-$tree-$n
export ADAMIC_BUILD_LOG=/tmp/bench/out/build-$name.log
: > "$ADAMIC_BUILD_LOG"
cd /tmp/bench/$tree
echo "before $(cat /proc/loadavg)" > /tmp/bench/out/load-$name.txt
/usr/bin/time -v -o /tmp/bench/out/time-$name.txt ./adamic build stage1/cohere/$program/main.ts -o /tmp/bench/out/bin-$tree-$program > /tmp/bench/out/stdout-$name.txt 2>&1
status=$?
echo "after $(cat /proc/loadavg)" >> /tmp/bench/out/load-$name.txt
echo "$name exit=$status $(grep 'Elapsed' /tmp/bench/out/time-$name.txt)"
