#!/bin/bash
# warm.sh <tree> <program>: one untimed build to warm the runtime archive, copying units.h out of the build directory.
source /tmp/bench/scripts/env.sh
tree=$1 program=$2
export ADAMIC_BUILD_CACHE_DIR=/tmp/bench/cache-$tree
export ADAMIC_BUILD_LOG=/tmp/bench/out/warm-$tree-$program.log
: > "$ADAMIC_BUILD_LOG"
cd /tmp/bench/$tree
( while true; do for d in /tmp/adamic-gate/adamic-units-*/; do [ -f "$d/units.h" ] && cp "$d/units.h" /tmp/bench/out/units-$tree-$program.h 2>/dev/null; done; sleep 0.02; done ) &
poller=$!
./adamic build stage1/cohere/$program/main.ts -o /tmp/bench/out/bin-$tree-$program > /tmp/bench/out/warm-$tree-$program.out 2>&1
status=$?
kill $poller
echo "$tree $program exit=$status"
