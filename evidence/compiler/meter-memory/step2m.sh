#!/usr/bin/env bash
S=/tmp/m/scratch; R=/tmp/m/run4; E=$S/main-adapted/src/tsc/tsc.ts
MEMO_TOSTRING=1 MEMO_STATS=$R/memo-stats.json HEAP_DIR=$R/heap HEAP_SIDECAR=$R/entry-heap-units.jsonl LATENT_FULL=0 LATENT_ASSERT_NO_OUTPUT=1 /usr/bin/time -v -o $R/2-entry-census-memo.time $S/entry-census-memo $E $R/tsc/latent.jsonl > $R/2.out 2> $R/2.err
echo "2-entry-census-memo exit=$?" >> $R/status
