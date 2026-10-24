#!/bin/bash
# control.sh <tree>: the lint build again, unedited, unit cache warm from the previous build.
source /tmp/bench/scripts/env.sh
tree=$1
cd /tmp/bench/$tree
export ADAMIC_BUILD_CACHE_DIR=/tmp/bench/cache-$tree
for n in 1 2; do
name=lint-$tree-control-$n
export ADAMIC_BUILD_LOG=/tmp/bench/out/build-$name.log
: > "$ADAMIC_BUILD_LOG"
echo "before $(cat /proc/loadavg)" > /tmp/bench/out/load-$name.txt
/usr/bin/time -v -o /tmp/bench/out/time-$name.txt ./adamic build stage1/cohere/lint/main.ts -o /tmp/bench/out/bin-$tree-lint > /tmp/bench/out/stdout-$name.txt 2>&1
echo "$name exit=$? $(grep Elapsed /tmp/bench/out/time-$name.txt)"
echo "after $(cat /proc/loadavg)" >> /tmp/bench/out/load-$name.txt
done
