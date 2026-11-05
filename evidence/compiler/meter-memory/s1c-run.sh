#!/usr/bin/env bash
# Sequential, nothing else running: S1c entry census without the memo, then with it. 3 hour cap each.
S=/tmp/m/s1c-scratch; E=$S/adapted/src/tsc/tsc.ts
for mode in 0 1; do
  R=/tmp/m/s1c-run-memo$mode; mkdir -p $R/heap $R/tsc
  /tmp/m/sampler.sh $R/rss.csv & sp=$!
  MEMO_TOSTRING=$mode MEMO_STATS=$R/memo-stats.json HEAP_DIR=$R/heap HEAP_SIDECAR=$R/entry-heap-units.jsonl LATENT_FULL=0 LATENT_ASSERT_NO_OUTPUT=1 \
    /usr/bin/time -v -o $R/entry-census.time timeout --signal=TERM 10800 $S/entry-census-probe $E $R/tsc/latent.jsonl > $R/out.txt 2> $R/err.txt
  echo "memo=$mode exit=$?" >> /tmp/m/s1c-status
  kill $sp
done
echo finished >> /tmp/m/s1c-status
