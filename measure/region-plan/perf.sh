#!/bin/bash
# #ahnc64w: graph and copy by function, where perf may count (a rented box). Never lands.
set -u
T=$(mktemp -d); D=measure/region-plan
echo "box $(hostname) $(nproc) cpus, $(uname -r), at $(git rev-parse --short=10 HEAD 2>/dev/null || echo sealed), paranoid $(cat /proc/sys/kernel/perf_event_paranoid 2>/dev/null)"
P=$(command -v perf || ls /usr/lib/linux-tools/*/perf 2>/dev/null | head -1); echo "perf: ${P:-none}  valgrind: $(command -v valgrind || echo none)"
go build -buildvcs=false -o $T/adamic ./cmd/adamic > $T/b.log 2>&1 && echo "PASS adamic built" || { echo "FAIL build"; tail -20 $T/b.log; exit 1; }
for p in graph graph_nt copy copy_nt; do
  $T/adamic build $D/$p.ts -o $T/$p > $T/$p.blog 2>&1 || { echo "FAIL build $p"; tail -3 $T/$p.blog; continue; }
  echo "$p symbols: $(nm $T/$p 2>/dev/null | grep -c ' [tT] ')"
done
echo "load before $(cut -d' ' -f1-3 /proc/loadavg)"
[ -n "$P" ] || { echo "GAP no perf on this box"; echo "== end"; exit 0; }
for p in graph graph_nt copy copy_nt; do
  $P stat -e instructions,cycles,task-clock -x, -o $T/$p.stat $T/$p > /dev/null 2>&1 && echo "$p stat: $(grep -E 'instructions|cycles|task-clock' $T/$p.stat | cut -d, -f1,3 | tr '\n' ' ')" || echo "$p stat refused: $(tail -2 $T/$p.stat 2>/dev/null | tr '\n' ' ')"
  if $P record -F 2999 -g -o $T/$p.data $T/$p > /dev/null 2> $T/$p.rec; then
    echo "== $p by function, self, top 22"
    $P report -i $T/$p.data --no-children --stdio --sort symbol 2>/dev/null | grep -E '^\s+[0-9.]+%' | head -22 | cut -c1-150
  else echo "$p record refused: $(tail -2 $T/$p.rec | tr '\n' ' ')"; fi
done
echo "load after $(cut -d' ' -f1-3 /proc/loadavg)"
echo "== end"
