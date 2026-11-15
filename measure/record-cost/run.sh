#!/bin/bash
# #chncwr5's baseline: Memory's record cost harness (zz_record_cost_probe_test.go) on a quiet box, at
# main plus #1vb0x5s step 1 (5e7964fd) with the probe (70396eb4). Four modes at 100K and 1M against
# Node's two kept shapes, every run interleaved round by round, best of ROUNDS (median make beside it),
# load before and after. Bytes per record are the peak above the same mode's N=0 peak. The counted,
# sanitized build runs once per mode and size, for allocations, frees, regions and merges. Never lands.
set -u
ROUNDS=${RECORD_COST_ROUNDS:-9}
T=$(mktemp -d); O=$T/out; mkdir -p $O
echo "box $(hostname) $(nproc) cpus, $(uname -r), tree at $(git rev-parse --short=10 HEAD 2>/dev/null || echo sealed), node $(node --version), rounds $ROUNDS"
echo "load before $(cut -d' ' -f1-3 /proc/loadavg); busy processes: $(ps -eo pcpu=,comm= | awk '$1 > 5 {printf "%s(%s) ", $2, $1}')"
RECORD_COST_OUT=$T go test ./internal/native -run '^TestRecordCostBuild$' -count=1 -v > $T/build.log 2>&1 && echo "PASS harness built: $(grep -o 'release flags: .*' $T/build.log)" || { echo "FAIL harness build"; tail -30 $T/build.log; exit 1; }
MODES="counted-kept graph-kept counted-dropped graph-dropped"
NODE_MODES="counted-kept graph-kept"
S=internal/native/testdata/record_cost/records.mjs
echo "== counted build, once each (ASan, UBSan, counts)"
for n in 100000 1000000; do for m in $MODES; do
  ASAN_OPTIONS=detect_leaks=0 UBSAN_OPTIONS=halt_on_error=1 $T/counted $m $n > $T/c.out 2> $T/c.err; e=$?
  echo "counted $m $n: exit $e | $(grep -o 'allocations=.*' $T/c.out) $(grep -cE '^(PASS)$' $T/c.out | sed 's/1/PASS/;s/^0$/no PASS/')"
done; done
echo "== timing rounds"
for r in $(seq 1 $ROUNDS); do
  for n in 0 100000 1000000; do
    for m in $MODES; do $T/release $m $n >> $O/adamic.$m.$n 2>/dev/null || echo "FAIL release $m $n round $r"; done
    for m in $NODE_MODES; do node --expose-gc $S $m $n >> $O/node.$m.$n 2>/dev/null || echo "FAIL node $m $n round $r"; done
  done
done
echo "load after $(cut -d' ' -f1-3 /proc/loadavg)"
field() { grep -o "$1=[0-9.]*" "$2" | cut -d= -f2; }
best() { sort -g | head -1; }
median() { sort -g | awk '{a[NR]=$1} END {print a[int((NR+1)/2)]}'; }
echo "== table: best of $ROUNDS (make median), seconds; peak MB; bytes per record above the mode's N=0 peak"
printf '%-7s %-16s %8s %9s %9s %9s %9s %10s\n' side mode n make make_med free peak_MB bytes/rec
for side in adamic node; do
  if [ $side = adamic ]; then ms=$MODES; else ms=$NODE_MODES; fi
  for n in 100000 1000000; do for m in $ms; do
    f=$O/$side.$m.$n; z=$O/$side.$m.0
    [ -s $f ] || { echo "$side $m $n: no runs"; continue; }
    mk=$(field make $f | best); mm=$(field make $f | median); fr=$(field free $f | best); pk=$(field peak_kb $f | best); base=$(field peak_kb $z | best)
    awk -v s=$side -v m=$m -v n=$n -v mk=$mk -v mm=$mm -v fr=$fr -v pk=$pk -v b=$base 'BEGIN {printf "%-7s %-16s %8d %9.4f %9.4f %9.4f %9.1f %10.1f\n", s, m, n, mk, mm, fr, pk/1024, (pk-b)*1024/n}'
  done; done
done
echo "== end"
