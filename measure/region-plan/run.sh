#!/bin/bash
# #ahnc64w: where graph's and copy's extra user time against Node goes, at this commit. Never lands.
set -u
T=$(mktemp -d); D=measure/region-plan
echo "box $(hostname) $(nproc) cpus, $(uname -r), at $(git rev-parse --short=10 HEAD), paranoid $(cat /proc/sys/kernel/perf_event_paranoid)"
echo "tools: valgrind=$(command -v valgrind || echo none) node=$(node --version 2>&1)"
go build -buildvcs=false -o $T/adamic ./cmd/adamic > $T/b.log 2>&1 && echo "PASS adamic built" || { echo "FAIL build"; tail -20 $T/b.log; exit 1; }
for p in graph copy graph_sum; do
  $T/adamic c $D/$p.ts > $T/$p.c 2> $T/$p.cerr || { echo "FAIL c $p"; tail -5 $T/$p.cerr; }
  echo "$p C: $(wc -l < $T/$p.c) lines; ADAMIC_REGION $(grep -c '= ADAMIC_REGION;' $T/$p.c), region_end $(grep -c 'adamic_region_end(' $T/$p.c), _in calls $(grep -cE '_in\(' $T/$p.c), object_new $(grep -c 'adamic_object_new' $T/$p.c), region allocs $(grep -cE 'adamic_region_(alloc|object|new)' $T/$p.c)"
  $T/adamic build $D/$p.ts -o $T/$p > $T/$p.blog 2>&1 || { echo "FAIL build $p"; tail -5 $T/$p.blog; continue; }
  $T/adamic build $D/$p.ts -o $T/$p.count --count > $T/$p.clog 2>&1 || { echo "FAIL count build $p"; tail -5 $T/$p.clog; continue; }
  $T/$p.count > $T/$p.count.out 2> $T/$p.count.err; node $D/$p.ts > $T/$p.node 2>/dev/null
  echo "$p counts: $(grep -h 'adamic: counts' $T/$p.count.err) | $(head -1 $T/$p.count.out)$(cmp -s $T/$p.count.out $T/$p.node && echo ', matches Node' || echo ', differs from Node')"
done
echo "== the region statement in graph() and copyAll(), emitted"
grep -n -B2 -A8 '= ADAMIC_REGION;' $T/graph.c | cut -c1-220 | head -30
echo "== time: release, best of 5, interleaved with Node; load before $(cut -d' ' -f1-3 /proc/loadavg)"
for p in graph copy graph_sum; do
  for side in native node; do echo 999 > $T/$p.$side.best; done
  for r in 1 2 3 4 5; do
    for side in native node; do
      if [ $side = native ]; then cmd="$T/$p"; else cmd="node $D/$p.ts"; fi
      /usr/bin/time -o $T/t -f "%e %U %S %R %M" $cmd > /dev/null 2> /dev/null
      read e u s f m < $T/t
      if awk "BEGIN{exit !($u < $(cut -d' ' -f2 $T/$p.$side.best 2>/dev/null || echo 999))}"; then echo "$e $u $s $f $m" > $T/$p.$side.best; fi
    done
  done
  for side in native node; do read e u s f m < $T/$p.$side.best; echo "$p $side wall $e user $u sys $s faults $f rss $((m/1024)) MB"; done
done
echo "load after $(cut -d' ' -f1-3 /proc/loadavg)"
if command -v valgrind > /dev/null; then
  for p in graph copy graph_sum; do
    valgrind --tool=callgrind --callgrind-out-file=$T/$p.cg $T/$p > /dev/null 2> $T/$p.vg || { echo "FAIL callgrind $p"; tail -5 $T/$p.vg; continue; }
    echo "== callgrind $p, inclusive and self instructions, top 25 by self"
    callgrind_annotate --threshold=99 $T/$p.cg 2>/dev/null | sed -n '1,60p' | grep -vE '^--|^$|^(Profile|Events|Thresh|Include|User|Auto)' | cut -c1-200 | head -40
  done
else
  echo "GAP no valgrind on this box: no profile by function"
fi
echo "== end"
