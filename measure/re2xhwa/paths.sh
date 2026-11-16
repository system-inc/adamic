#!/bin/bash
# #re2xhwa: does a binary's path length alone move nbody's time? One nbody binary, copied byte for byte
# to a short path and to J-eewt3vf's two names, run 41 times each, interleaved; user seconds to the millisecond
# (bash's time; /usr/bin/time's 10 ms can't see 3% of nbody's 80 ms, J-4yschec), median and quartiles. Never lands.
set -u
T=$(mktemp -d)
echo "box $(hostname) $(nproc) cpus; load before $(cut -d' ' -f1-3 /proc/loadavg)"
go build -buildvcs=false -o $T/adamic ./cmd/adamic || { echo "FAIL build adamic"; exit 1; }
mkdir -p $T/1 $T/work
$T/adamic build bench/nbody.ts -o $T/1/nbody || { echo "FAIL build nbody"; exit 1; }
cp $T/1/nbody $T/work/nbody.baseline; cp $T/1/nbody $T/work/nbody.candidate
cmp $T/work/nbody.baseline $T/work/nbody.candidate && echo "PASS the three copies are the same bytes"
TIMEFORMAT=%3U
for r in $(seq 1 41); do
  for p in 1/nbody work/nbody.baseline work/nbody.candidate; do
    { time $T/$p > /dev/null; } 2>> "$T/$(echo $p | tr / _).times"
  done
done
for p in 1/nbody work/nbody.baseline work/nbody.candidate; do
  f="$T/$(echo $p | tr / _).times"
  echo "$p (path length $(printf %s "$T/$p" | wc -c)): $(sort -g $f | awk '{a[NR]=$1} END {printf "user median %.3f s, quartiles %.3f to %.3f, min %.3f", a[int((NR+1)/2)], a[int(NR/4)+1], a[int(3*NR/4)+1], a[1]}')"
done
echo "load after $(cut -d' ' -f1-3 /proc/loadavg)"
