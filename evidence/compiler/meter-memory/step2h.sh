#!/usr/bin/env bash
S=/tmp/m/scratch; R=/tmp/m/run3; E=$S/main-adapted/src/tsc/tsc.ts
HEAP_DIR=$R/heap HEAP_SIDECAR=$R/entry-heap-units.jsonl LATENT_FULL=0 LATENT_ASSERT_NO_OUTPUT=1 /usr/bin/time -v -o $R/2-entry-census-heap.time $S/entry-census-heap $E $R/tsc/latent.jsonl > $R/2.out 2> $R/2.err
echo "2-entry-census-heap exit=$?" >> $R/status
