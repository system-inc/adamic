set -u
source ~/.adamic-tools/env.sh 2>/dev/null || source ~/adamic-tools/env.sh 2>/dev/null || true
export TMPDIR=$HOME/scratch/system_adamic_runtime_performance/tmp; mkdir -p "$TMPDIR"; T=$(mktemp -d)
git -c protocol.file.allow=always submodule update --init --recursive > $T/sub.log 2>&1 && echo "PASS submodules" || { echo "FAIL submodules"; tail -5 $T/sub.log; }
echo "box $(hostname) at $(git log --format=%h -1)"
go build -buildvcs=false -o $T/adamic ./cmd/adamic > $T/b.log 2>&1 && echo "PASS adamic built" || { echo "FAIL build"; tail -5 $T/b.log; }
# #cgn2cwt (c): region.c's cap at 1 MB (a huge block never fits, so it runs as 291f3ffb) against 2 MB
# (a thread keeps one huge block idle), on trees, graph and copy, beside Node. Never lands.
R=internal/native/runtime/region.c
M=measure/region-cap
echo "box $(hostname) $(nproc) cpus, kernel $(uname -r), uptime $(cut -d' ' -f1 /proc/uptime) s, $(ldd --version | head -1), node $(node --version)"
echo "thp enabled: $(cat /sys/kernel/mm/transparent_hugepage/enabled) defrag: $(cat /sys/kernel/mm/transparent_hugepage/defrag)"
echo "load at start $(cut -d' ' -f1-3 /proc/loadavg)"
cp $R $T/region.c.orig
CAPS="1048576 2097152"
for cap in $CAPS; do
  sed "s/^#define ADAMIC_REGION_BLOCK_CACHE ((size_t)[0-9]*)\$/#define ADAMIC_REGION_BLOCK_CACHE ((size_t)$cap)/" $T/region.c.orig > $R
  [ "$(grep -c "^#define ADAMIC_REGION_BLOCK_CACHE ((size_t)$cap)\$" $R)" = 1 ] && echo "PASS define at $cap" || echo "FAIL define at $cap"
  go build -buildvcs=false -o $T/adamic.$cap ./cmd/adamic > $T/b.$cap.log 2>&1 && echo "PASS adamic built at cap $cap" || { echo "FAIL adamic build at cap $cap"; tail -5 $T/b.$cap.log; }
  echo "== TestRegionBlockCache/idle at cap $cap (its bound is 1 MB a thread plus 2 MB, the 1 MB cap's)"
  ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 10m -v -run '^TestRegionBlockCache$/^idle$' ./internal/native/ > $T/idle.$cap.log 2>&1; echo "exit=$?"
  grep -E -- '--- (PASS|FAIL)|KB above' $T/idle.$cap.log
done
cp $T/region.c.orig $R
git diff --quiet -- $R && echo "PASS region.c restored" || echo "FAIL region.c differs from the commit"
source_of() { case $1 in trees) echo bench/trees.ts ;; *) echo $M/$1.ts ;; esac; }
P="trees graph copy"
for cap in $CAPS; do for p in $P; do
  $T/adamic.$cap build $(source_of $p) -o $T/$p.$cap > $T/c.log 2>&1 && echo "PASS built $p at cap $cap" || { echo "FAIL build $p at cap $cap"; tail -5 $T/c.log; }
done; done
echo "== count lines (cap 1 MB, --count): regions must be most of allocations; ring and array_copy are the brief's first shapes"
for p in $P ring array_copy; do
  $T/adamic.1048576 build $(source_of $p) -o $T/$p.count --count > $T/c.log 2>&1 || { echo "FAIL count build $p"; tail -5 $T/c.log; continue; }
  $T/$p.count > $T/$p.count.out 2> $T/$p.count.err
  node --no-warnings $(source_of $p) > $T/$p.count.node 2> /dev/null
  echo "$p: $(grep -h 'adamic: counts' $T/$p.count.err) | $(head -1 $T/$p.count.out)$(cmp -s $T/$p.count.out $T/$p.count.node && echo ', matches Node' || echo ', DIFFERS from Node')"
done
echo "== outputs against Node"
for p in $P; do
  node --no-warnings $(source_of $p) > $T/$p.node.out 2> /dev/null
  for cap in $CAPS; do $T/$p.$cap > $T/$p.$cap.out 2> /dev/null; cmp -s $T/$p.$cap.out $T/$p.node.out && echo "PASS $p at cap $cap matches Node: $(tail -1 $T/$p.node.out)" || echo "FAIL $p at cap $cap differs from Node"; done
done
echo "== the slot at work: 4 MB huge reservations and 2 MB unmaps, one run each"
if command -v strace > /dev/null; then for p in $P; do for cap in $CAPS; do
  strace -f -e trace=mmap,munmap -o $T/st $T/$p.$cap > /dev/null 2>&1
  echo "$p cap $cap: mmap 4 MB $(grep -c 'mmap(NULL, 4194304' $T/st), munmap 2 MB $(grep -cE 'munmap\(0x[0-9a-f]+, 2097152\)' $T/st), mmap of 4 KB to 1 MB $(grep -cE 'mmap\(NULL, (4096|8192|16384|32768|65536|131072|262144|524288|1048576),' $T/st)"
done; done; else echo "no strace"; fi
echo "load before $(cut -d' ' -f1-3 /proc/loadavg)"
for round in 1 2 3 4 5; do for p in $P; do for v in $CAPS node; do
  if [ $v = node ]; then cmd="node --no-warnings $(source_of $p)"; else cmd="$T/$p.$v"; fi
  /usr/bin/time -o $T/t -f "%e %U %S %R %M" $cmd > /dev/null 2> $T/e
  last=$(awk '/^vmrss last/{print $3}' $T/e); quiet=$(awk '/^vmrss quiet/{print $3}' $T/e)
  echo "$p $v $(tail -1 $T/t) ${last:--} ${quiet:--}" >> $T/rounds
done; done; done
echo "load after $(cut -d' ' -f1-3 /proc/loadavg)"
echo "== every run: program variant wall user sys minor-faults max-rss-kb vmrss-after-last-call-kb vmrss-after-400ms-kb"; cat $T/rounds
echo "== best of 5 by wall (that run's user, sys, faults, max RSS); VmRSS after the last call and after the 400 ms wait, median of 5"
for p in $P; do for v in $CAPS node; do
  best=$(awk -v p=$p -v v=$v '$1==p && $2==v' $T/rounds | sort -k3 -g | head -1)
  last=$(awk -v p=$p -v v=$v '$1==p && $2==v {print $8}' $T/rounds | sort -g | sed -n 3p)
  quiet=$(awk -v p=$p -v v=$v '$1==p && $2==v {print $9}' $T/rounds | sort -g | sed -n 3p)
  echo "$best" | awk -v l=$last -v q=$quiet '{printf "%-6s %-8s wall %s s, user %s, sys %s, faults %s, max RSS %.1f MB, idle last %s kB, quiet %s kB\n", $1, $2, $3, $4, $5, $6, $7/1024, l, q}'
done; done
echo "load end $(cut -d' ' -f1-3 /proc/loadavg)"
echo "== end"
