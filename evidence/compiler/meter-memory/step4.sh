#!/usr/bin/env bash
S=/tmp/m/scratch; R=/tmp/m/run3; C=$S/main-adapted/src/compiler
LATENT_TIMES=$R/latent-times.jsonl LATENT_FULL=0 LATENT_ASSERT_NO_OUTPUT=1 /usr/bin/time -v -o $R/4-latent-census.time $S/latent-census $C $R/latent.jsonl > $R/4.out 2> $R/4.err
echo "4-latent-census exit=$?" >> $R/status
