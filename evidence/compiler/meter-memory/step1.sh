#!/usr/bin/env bash
S=/tmp/m/scratch; R=${1:-/tmp/m/run1}; mkdir -p $R/tsc
E=$S/main-adapted/src/tsc/tsc.ts; C=$S/main-adapted/src/compiler
t(){ n=$1; shift; /usr/bin/time -v -o $R/$n.time "$@" > $R/$n.out 2> $R/$n.err; echo "$n exit=$?" >> $R/status; }
t 1-census $S/census $E $R/tsc/census.jsonl
LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 t 2-entry-census $S/entry-census $E $R/tsc/latent.jsonl
t 3-census-compiler $S/census $C $R/census.jsonl
LATENT_TIMES=$R/latent-times.jsonl LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 t 4-latent-census $S/latent-census $C $R/latent.jsonl
echo finished >> $R/status
