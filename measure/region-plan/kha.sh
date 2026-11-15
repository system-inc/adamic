#!/bin/bash
# #kha2ea0: graph and copy with their timers, at Optimizer's head and with its emitter change undone
# (head.patch reversed, so the program-wide gate is back), against their _nt twins and Node. Never lands.
set -u
T=$(mktemp -d); D=measure/region-plan
echo "box $(hostname) $(nproc) cpus, $(uname -r), tree at $(git rev-parse --short=10 HEAD 2>/dev/null || echo sealed)"
go build -buildvcs=false -o $T/adamic.head ./cmd/adamic > $T/b.log 2>&1 && echo "PASS head built" || { echo "FAIL head build"; tail -20 $T/b.log; exit 1; }
patch -R -p1 -s < $D/head.patch && echo "PASS head.patch reversed" || { echo "FAIL reverse"; exit 1; }
grep -q 'program.HasPromises() || program.HasGenerators()' internal/native/emit.go && echo "PASS the program-wide gate is back" || echo "FAIL the gate isn't back"
go build -buildvcs=false -o $T/adamic.before ./cmd/adamic > $T/b2.log 2>&1 && echo "PASS before built" || { echo "FAIL before build"; tail -20 $T/b2.log; }
patch -p1 -s < $D/head.patch && echo "PASS head.patch reapplied" || echo "FAIL reapply"
for v in before head; do for p in graph graph_nt copy copy_nt; do
  $T/adamic.$v build $D/$p.ts -o $T/$p.$v > $T/$p.$v.blog 2>&1 || { echo "FAIL build $p at $v"; tail -3 $T/$p.$v.blog; continue; }
  $T/adamic.$v build $D/$p.ts -o $T/$p.$v.count --count > $T/c.log 2>&1 || { echo "FAIL count $p at $v"; continue; }
  $T/$p.$v.count > $T/$p.$v.out 2> $T/$p.$v.err; node $D/$p.ts > $T/$p.node 2>/dev/null
  echo "$p $v: $(grep -h 'adamic: counts' $T/$p.$v.err) | $(cmp -s $T/$p.$v.out $T/$p.node && echo 'matches Node' || echo 'differs from Node')"
done; done
echo "== time: release, best of 5 by user, interleaved (before, head, Node); load before $(cut -d' ' -f1-3 /proc/loadavg)"
for p in graph copy graph_nt copy_nt; do
  for side in before head node; do echo "999 999 999 0 0" > $T/$p.$side.best; done
  for r in 1 2 3 4 5; do for side in before head node; do
    if [ $side = node ]; then cmd="node $D/$p.ts"; else cmd="$T/$p.$side"; fi
    /usr/bin/time -o $T/t -f "%e %U %S %R %M" $cmd > /dev/null 2> /dev/null
    read e u s f m < $T/t; read be bu rest < $T/$p.$side.best
    awk "BEGIN{exit !($u < $bu)}" && echo "$e $u $s $f $m" > $T/$p.$side.best
  done; done
  for side in before head node; do read e u s f m < $T/$p.$side.best; echo "$p $side wall $e user $u sys $s faults $f rss $((m/1024)) MB"; done
done
echo "load after $(cut -d' ' -f1-3 /proc/loadavg)"
echo "== end"
