#!/bin/bash
# #ahnc64w: does one timer, or readFileSync, cost graph and copy their regions (emit.go's HasPromises gate)? Never lands.
set -u
T=$(mktemp -d); D=measure/region-plan
echo "box $(hostname) $(nproc) cpus, at $(git rev-parse --short=10 HEAD)"
go build -buildvcs=false -o $T/adamic ./cmd/adamic > $T/b.log 2>&1 && echo "PASS adamic built" || { echo "FAIL build"; tail -20 $T/b.log; exit 1; }
for p in graph graph_nt graph_nofs copy copy_nt copy_nofs; do
  $T/adamic c $D/$p.ts > $T/$p.c 2> $T/$p.cerr || { echo "FAIL c $p: $(head -2 $T/$p.cerr)"; continue; }
  $T/adamic build $D/$p.ts -o $T/$p > $T/$p.blog 2>&1 || { echo "FAIL build $p"; tail -3 $T/$p.blog; continue; }
  $T/adamic build $D/$p.ts -o $T/$p.count --count > $T/$p.clog 2>&1 || { echo "FAIL count $p"; continue; }
  $T/$p.count > $T/$p.out 2> $T/$p.err; node $D/$p.ts > $T/$p.node 2>/dev/null
  echo "$p: ADAMIC_REGION $(grep -c '= ADAMIC_REGION;' $T/$p.c) | $(grep -h 'adamic: counts' $T/$p.err) | $(cmp -s $T/$p.out $T/$p.node && echo 'matches Node' || echo 'differs from Node')"
done
echo "== time: release, best of 5 by user, interleaved with Node; load before $(cut -d' ' -f1-3 /proc/loadavg)"
for p in graph_nt copy_nt graph copy; do
  for side in native node; do echo "999 999 999 0 0" > $T/$p.$side.best; done
  for r in 1 2 3 4 5; do for side in native node; do
    if [ $side = native ]; then cmd="$T/$p"; else cmd="node $D/$p.ts"; fi
    /usr/bin/time -o $T/t -f "%e %U %S %R %M" $cmd > /dev/null 2> /dev/null
    read e u s f m < $T/t; read be bu rest < $T/$p.$side.best
    awk "BEGIN{exit !($u < $bu)}" && echo "$e $u $s $f $m" > $T/$p.$side.best
  done; done
  for side in native node; do read e u s f m < $T/$p.$side.best; echo "$p $side wall $e user $u sys $s faults $f rss $((m/1024)) MB"; done
done
echo "load after $(cut -d' ' -f1-3 /proc/loadavg)"
echo "== end"
