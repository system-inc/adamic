#!/bin/bash
# incremental.sh <tree>: a timed cold lint build, then one empty function added to context.ts and a rebuild with the unit cache warm, then the edit reverted.
source /tmp/bench/scripts/env.sh
tree=$1
/tmp/bench/scripts/run.sh $tree lint inc-cold
cd /tmp/bench/$tree
file=stage1/cohere/lint/context.ts
sed -i '66a\
\
export function benchmarkEmptyFunction(): void {}' $file
git diff --stat; git diff > /tmp/bench/out/incremental-$tree.diff
export ADAMIC_BUILD_CACHE_DIR=/tmp/bench/cache-$tree
name=lint-$tree-inc-warm
export ADAMIC_BUILD_LOG=/tmp/bench/out/build-$name.log
: > "$ADAMIC_BUILD_LOG"
echo "before $(cat /proc/loadavg)" > /tmp/bench/out/load-$name.txt
/usr/bin/time -v -o /tmp/bench/out/time-$name.txt ./adamic build stage1/cohere/lint/main.ts -o /tmp/bench/out/bin-$tree-lint > /tmp/bench/out/stdout-$name.txt 2>&1
echo "$name exit=$? $(grep Elapsed /tmp/bench/out/time-$name.txt)"
echo "after $(cat /proc/loadavg)" >> /tmp/bench/out/load-$name.txt
./adamic c stage1/cohere/lint/main.ts 2>/dev/null | grep -c benchmarkEmptyFunction | sed 's/^/emitted mentions: /'
git checkout -- $file
git status --short
